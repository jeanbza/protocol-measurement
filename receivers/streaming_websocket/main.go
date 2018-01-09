package main

import (
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
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

	upgrader := websocket.Upgrader{}

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		defer c.Close()

		for {
			var i = new(messages.Message)
			err := c.ReadJSON(&i)
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
				[]interface{}{id.String(), i.RunId, "websocket", i.CreatedAt, i.SentAt, i.ReceivedAt},
			)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%s", port), m)

	fmt.Println("We're done here!")
}