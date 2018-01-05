package main

import (
	"cloud.google.com/go/pubsub"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"cloud.google.com/go/spanner"
)

const (
	topicName          = "send_queue"
	routines           = 10
	database           = "projects/deklerk-sandbox/instances/protocol-measurement/databases/protocol-measurement"
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

	spannerClient, err := spanner.NewClient(ctx, database)
	if err != nil {
		panic(err)
	}
	defer spannerClient.Close()

	sm := setManager{
		spannerClient: spannerClient,
		topic: t,
		ctx:   ctx,
	}

	r := mux.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("static"))) // hacky - be sure to run go run *.go in this folder
	r.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))
	r.HandleFunc("/sets", sm.getSetsHandler).Methods("GET")
	r.HandleFunc("/sets", sm.createSetHandler).Methods("POST")

	fmt.Println("Serving")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		panic(err)
	}
}
