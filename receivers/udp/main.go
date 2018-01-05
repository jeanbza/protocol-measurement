package main

import (
	"cloud.google.com/go/spanner"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"net"
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

	go repeatedlySaveToSpanner(ctx, client, insertQueue)

	fmt.Println("Listening!")

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	buf := make([]byte, 65536)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		message := buf[0:n]

		var i = new(messages.Message)
		err = json.Unmarshal(message, i)
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
			[]interface{}{id.String(), i.RunId, "udp", i.CreatedAt, i.SentAt, i.ReceivedAt},
		)

	}

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}

func repeatedlySaveToSpanner(ctx context.Context, client *spanner.Client, insertQueue <-chan (*spanner.Mutation)) {
	ticker := time.NewTicker(time.Second)
	toBeSent := []*spanner.Mutation{}

	for {
		select {
		case <-ticker.C:
			if len(toBeSent) == 0 {
				break
			}
			fmt.Println("Saving", len(toBeSent))
			_, err := client.Apply(ctx, toBeSent)
			if err != nil {
				panic(err)
			}
			toBeSent = []*spanner.Mutation{}
		case i := <-insertQueue:
			toBeSent = append(toBeSent, i)
		}
	}
}