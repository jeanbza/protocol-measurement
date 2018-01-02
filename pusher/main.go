package main

import (
	"cloud.google.com/go/pubsub"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	topicName          = "send_queue"
	routines           = 2
	messagesPerRoutine = 1
)

var (
	set         = uuid.NewV4().String()
	sent uint64 = 0
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
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	fmt.Println("Getting topic")

	t := client.Topic(topicName)
	if t == nil {
		panic("Expected topic not to be nil")
	}

	fmt.Println("Publishing")

	wg := &sync.WaitGroup{}
	for j := 0; j < routines; j++ {
		wg.Add(1)
		go startAdding(t, ctx, wg)
	}

	go watch()

	wg.Wait()
	printProgress()
	fmt.Println("Done")
}

func startAdding(t *pubsub.Topic, ctx context.Context, wg *sync.WaitGroup) {
	for i := 0; i < messagesPerRoutine; i++ {
		m := messages.Message{
			Set:       set,
			CreatedAt: time.Now(),
		}
		j, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}

		res := t.Publish(ctx, &pubsub.Message{
			Data: j,
		})
		_, err = res.Get(context.Background())
		if err != nil {
			panic(err)
		}

		atomic.AddUint64(&sent, 1)
	}
	wg.Done()
}

func watch() {
	t := time.NewTicker(time.Second)

	for {
		<-t.C
		printProgress()
	}
}

func printProgress() {
	fmt.Printf("%d / %d\n", sent, messagesPerRoutine*routines)
}
