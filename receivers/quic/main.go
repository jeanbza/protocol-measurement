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
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/devsisters/goquic"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	database = "projects/deklerk-sandbox/instances/protocol-measurement/databases/protocol-measurement"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		panic("Expected to receive an environment variable PORT")
	}

	certFile := os.Getenv("CERT_FILE")
	if port == "" {
		panic("Expected to receive an environment variable CERT_FILE")
	}

	privateKeyFile := os.Getenv("PRIVATE_KEY_FILE")
	if port == "" {
		panic("Expected to receive an environment variable PRIVATE_KEY_FILE")
	}

	ctx := context.Background()
	insertQueue := make(chan (*spanner.Mutation), 16384)

	client, err := spanner.NewClient(ctx, database)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go messages.NewSpannerSaver(client, insertQueue).RepeatedlySaveToSpanner(ctx)

	fmt.Println("Listening!")

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Println("Received")

		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		var i = new(messages.Message)
		err = json.Unmarshal(bodyBytes, i)
		if err != nil {
			panic(err)
		}

		i.ReceivedAt = time.Now()
		id, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}

		insertQueue <- spanner.Insert(
			"results",
			[]string{"id", "runId", "protocol", "createdAt", "sentAt", "receivedAt"},
			[]interface{}{id.String(), i.RunId, "quic", i.CreatedAt, i.SentAt, i.ReceivedAt},
		)
	})

	err = goquic.ListenAndServe(fmt.Sprintf(":%s", port), certFile, privateKeyFile, 1, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}
