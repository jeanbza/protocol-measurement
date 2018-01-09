// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type runManager struct {
	spannerClient *spanner.Client
	topic         *pubsub.Topic
	ctx           context.Context
}

func (sm *runManager) getRunResultsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runId, ok := vars["runId"]
	if !ok {
		panic("Expected to be provided a runId")
	}

	stmt := spanner.Statement{SQL: `
		SELECT protocol, COUNT(*), AVG(TIMESTAMP_DIFF(receivedAt, sentAt, MILLISECOND)) AS avgTravelTime
		FROM results
		WHERE runId = @runId
		GROUP BY protocol`, Params: map[string]interface{}{"runId": runId}}
	iter := sm.spannerClient.Single().Query(sm.ctx, stmt)

	runs := map[string]map[string]float64{}

	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}

		var protocol string
		var count int64
		var avgTravelTime float64
		if err := row.Columns(&protocol, &count, &avgTravelTime); err != nil {
			panic(err)
		}

		runs[protocol] = map[string]float64{
			"count":         float64(count),
			"avgTravelTime": avgTravelTime,
		}
	}

	outBytes, err := json.Marshal(runs)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		panic(err)
	}
}

func (sm *runManager) getRunHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runId, ok := vars["runId"]
	if !ok {
		panic("Expected to be provided a runId")
	}

	stmt := spanner.Statement{SQL: `
		SELECT id, createdAt, finishedCreating, totalMessages
		FROM runs
		WHERE id = @runId`, Params: map[string]interface{}{"runId": runId}}
	iter := sm.spannerClient.Single().Query(sm.ctx, stmt)

	run := map[string]interface{}{}

	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}

		var id string
		var createdAt time.Time
		var finishedCreating bool
		var totalMessages int64
		if err := row.Columns(&id, &createdAt, &finishedCreating, &totalMessages); err != nil {
			panic(err)
		}

		run["id"] = id
		run["createdAt"] = createdAt
		run["finishedCreating"] = finishedCreating
		run["totalMessages"] = totalMessages
	}

	outBytes, err := json.Marshal(run)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		panic(err)
	}
}

func (sm *runManager) getRunsHandler(w http.ResponseWriter, r *http.Request) {
	stmt := spanner.Statement{SQL: `SELECT id, createdAt, finishedCreating, totalMessages FROM runs`}
	iter := sm.spannerClient.Single().Query(sm.ctx, stmt)

	runs := []map[string]interface{}{}

	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}

		var id string
		var createdAt time.Time
		var finishedCreating bool
		var totalMessages int64
		if err := row.Columns(&id, &createdAt, &finishedCreating, &totalMessages); err != nil {
			panic(err)
		}

		run := map[string]interface{}{}
		run["id"] = id
		run["createdAt"] = createdAt
		run["finishedCreating"] = finishedCreating
		run["totalMessages"] = totalMessages

		runs = append(runs, run)
	}

	outBytes, err := json.Marshal(runs)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		panic(err)
	}
}

func (sm *runManager) createRunHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var values map[string]int
	json.Unmarshal(bodyBytes, &values)

	numMessages, ok := values["numMessages"]
	if !ok {
		panic(fmt.Sprintf("Expected numMessages, got %v", values))
	}

	runId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	fmt.Println(numMessages, messagesPerSend, routines, numMessages/messagesPerSend/routines)

	rc := &runCreator{
		sendsPerRoutine: numMessages / messagesPerSend / routines,
		wg:              &sync.WaitGroup{},
		ctx:             sm.ctx,
		spannerClient:   sm.spannerClient,
		topic:           sm.topic,
		runId:           runId.String(),
	}

	rc.create(int64(numMessages))
	rc.printProgress()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s"}`, rc.runId)))
}

type runCreator struct {
	sendsPerRoutine int
	wg              *sync.WaitGroup
	ctx             context.Context
	spannerClient   *spanner.Client
	topic           *pubsub.Topic
	runId           string
	sent            uint64
}

func (sc *runCreator) create(numMessages int64) {
	_, err := sc.spannerClient.Apply(sc.ctx, []*spanner.Mutation{spanner.Insert(
		"runs",
		[]string{"id", "createdAt", "finishedCreating", "totalMessages"},
		[]interface{}{sc.runId, time.Now(), false, numMessages},
	)})
	if err != nil {
		panic(err)
	}

	for j := 0; j < routines; j++ {
		sc.wg.Add(1)
		go sc.startAdding()
	}

	stopPrinting := make(chan (struct{}))

	go func() {
		t := time.NewTicker(time.Second)

		for {
			select {
			case <-t.C:
				sc.printProgress()
			case <-stopPrinting:
				return
			}
		}
	}()

	sc.wg.Wait()
	stopPrinting <- struct{}{}

	_, err = sc.spannerClient.Apply(sc.ctx, []*spanner.Mutation{spanner.Update(
		"runs",
		[]string{"id", "finishedCreating"},
		[]interface{}{sc.runId, true},
	)})
	if err != nil {
		panic(err)
	}
}

func (sc *runCreator) startAdding() {
	fmt.Println("Heyo", sc.sendsPerRoutine)
	for i := 0; i < sc.sendsPerRoutine; i++ {
		fmt.Println("About to send one")
		m := messages.SendRequest{
			RunId:  sc.runId,
			Amount: messagesPerSend,
		}
		j, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}

		res := sc.topic.Publish(sc.ctx, &pubsub.Message{
			Data: j,
		})
		_, err = res.Get(context.Background())
		if err != nil {
			panic(err)
		}

		atomic.AddUint64(&sc.sent, 1)
		fmt.Println("I sent one")
	}
	sc.wg.Done()
}

func (sc *runCreator) printProgress() {
	fmt.Printf("%s: %d / %d\n", sc.runId, sc.sent, sc.sendsPerRoutine*routines)
}
