package main

import (
	"cloud.google.com/go/pubsub"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

const (
	topicName          = "send_queue"
	routines           = 2
	messagesPerRoutine = 1
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
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	fmt.Println("Getting topic")

	t := client.Topic(topicName)
	if t == nil {
		panic("Expected topic not to be nil")
	}

	sm := setManager{
		topic: t,
		ctx:   ctx,
	}

	r := mux.NewRouter()
	r.HandleFunc("/set", sm.createSetHandler).Methods("POST")

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		panic(err)
	}
}
