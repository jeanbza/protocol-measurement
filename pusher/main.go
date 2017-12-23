package main

import (
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"os"
)

const (
	topicName = "send_queue"
)

func main() {
	fmt.Println("Starting up")

	ctx := context.Background()

	projectId := os.Getenv("GCP_PROJECT_ID")
	if projectId == "" {
		panic("Expected to receive an environment variable GCP_PROJECT_ID")
	}

	fmt.Println("Getting client")

	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Getting topic")

	t := client.Topic(topicName)
	if t == nil {
		panic("Expected topic not to be nil")
	}

	fmt.Println("Publishing")

	res := t.Publish(ctx, &pubsub.Message{
		Data:       []byte("Hello world"),
		Attributes: map[string]string{"foo": "bar"},
	})

	s, err := res.Get(context.Background())
	fmt.Println(s, err)

	fmt.Println("Done")
}
