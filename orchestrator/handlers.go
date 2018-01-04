package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"google.golang.org/api/iterator"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type setManager struct {
	spannerClient *spanner.Client
	topic         *pubsub.Topic
	ctx           context.Context
}

func (sm *setManager) createSetHandler(w http.ResponseWriter, r *http.Request) {
	setName, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	sc := &setCreator{
		wg:    &sync.WaitGroup{},
		ctx:   sm.ctx,
		topic: sm.topic,
		name:  setName.String(),
	}

	sc.create()
	sc.printProgress()
}

func (sm *setManager) getSetsHandler(w http.ResponseWriter, r *http.Request) {
	stmt := spanner.Statement{SQL: `SELECT resultSet FROM result2 GROUP BY protocol, resultSet`}
	iter := sm.spannerClient.Single().Query(sm.ctx, stmt)

	sets := []string{}

	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		var set string
		if err := row.Columns(&set); err != nil {
			panic(err)
		}

		sets = append(sets, set)
	}

	outBytes, err := json.Marshal(sets)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		panic(err)
	}
}

type setCreator struct {
	wg    *sync.WaitGroup
	ctx   context.Context
	topic *pubsub.Topic
	name  string
	sent  uint64
}

func (sc *setCreator) create() {
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
}

func (sc *setCreator) startAdding() {
	for i := 0; i < messagesPerRoutine; i++ {
		m := messages.Message{
			Set:       sc.name,
			CreatedAt: time.Now(),
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
	}
	sc.wg.Done()
}

func (sc *setCreator) printProgress() {
	fmt.Printf("%s: %d / %d\n", sc.name, sc.sent, messagesPerRoutine*routines)
}
