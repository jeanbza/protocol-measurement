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
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"time"
)

const (
	topicName       = "send_queue"
	routines        = 10 // must be power of 10 for division to work out neatly
	messagesPerSend = 1000
	database        = "projects/deklerk-sandbox/instances/protocol-measurement/databases/protocol-measurement"
)

func main() {
	fmt.Println("Starting up")

	ctx := context.Background()

	projectId := os.Getenv("GCP_PROJECT_ID")
	if projectId == "" {
		panic("Expected to receive an environment variable GCP_PROJECT_ID")
	}

	port := os.Getenv("PORT")
	if projectId == "" {
		panic("Expected to receive an environment variable PORT")
	}

	fmt.Println("Getting client")
	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}
	defer func() {
		err := pubsubClient.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Getting topic")
	t := pubsubClient.Topic(topicName)
	if t == nil {
		panic("Expected topic not to be nil")
	}
	t.PublishSettings.CountThreshold = 10000
	t.PublishSettings.DelayThreshold = 100 * time.Millisecond
	t.PublishSettings.ByteThreshold = 1e9

	spannerClient, err := spanner.NewClient(ctx, database)
	if err != nil {
		panic(err)
	}
	defer spannerClient.Close()

	sm := runManager{
		spannerClient: spannerClient,
		topic:         t,
		ctx:           ctx,
	}

	r := mux.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("static"))) // hacky - be sure to run go run *.go in this folder
	r.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))

	r.HandleFunc("/runs/{runId}/results", sm.getRunResultsHandler).Methods("GET")
	r.HandleFunc("/runs/{runId}", sm.getRunHandler).Methods("GET")
	r.HandleFunc("/runs", sm.getRunsHandler).Methods("GET")
	r.HandleFunc("/runs", sm.createRunHandler).Methods("POST")

	fmt.Println("Serving")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		panic(err)
	}
}
