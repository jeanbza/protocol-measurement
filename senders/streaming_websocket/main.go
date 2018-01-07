package main

import (
	"cloud.google.com/go/pubsub"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"net/url"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("STREAMING_WEBSOCKET_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable STREAMING_WEBSOCKET_RECEIVER_IP")
	}

	sendPort := os.Getenv("STREAMING_WEBSOCKET_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable STREAMING_WEBSOCKET_RECEIVER_PORT")
	}

	projectId := os.Getenv("GCP_PROJECT_ID")
	if projectId == "" {
		panic("Expected to receive an environment variable GCP_PROJECT_ID")
	}

	ctx := context.Background()

	fmt.Println("Getting client")
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v\n", err))
	}

	fmt.Println("Getting topic")
	t := client.Topic(topicName)
	if t == nil {
		panic("Expected topic not to be nil")
	}

	subscriptionId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	subscriptionName := fmt.Sprintf("http-sender-%s", subscriptionId.String())

	fmt.Println("Creating subscription", subscriptionName)
	s, err := client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{Topic: t})
	if err != nil {
		panic(err)
	}

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%s", sendIp, sendPort), Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}

	w := websocketWriter{
		q: make(chan (*messages.Message), 16384),
		c: c,
	}

	go w.asyncWriteToWebsocket()

	fmt.Println("Listening for messages")
	err = s.Receive(ctx, func(ctx2 context.Context, msg *pubsub.Message) {
		fmt.Println("About to send")
		msg.Ack()

		var i = new(messages.Message)
		json.Unmarshal(msg.Data, &i)

		i.SentAt = time.Now()

		w.q <- i

		fmt.Println("Sent")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}

type websocketWriter struct {
	q chan (*messages.Message)
	c *websocket.Conn
}

func (w *websocketWriter) asyncWriteToWebsocket() {
	for {
		msg := <-w.q
		err := w.c.WriteJSON(msg)
		if err != nil {
			panic(err)
		}
	}
}
