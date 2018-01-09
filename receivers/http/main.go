package main

import (
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
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
			[]interface{}{id.String(), i.RunId, "http", i.CreatedAt, i.SentAt, i.ReceivedAt},
		)
	})

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}