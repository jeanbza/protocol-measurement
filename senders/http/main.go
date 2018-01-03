package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"deklerk-startup-project"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("HTTP_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable HTTP_RECEIVER_IP")
	}

	sendPort := os.Getenv("HTTP_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable HTTP_RECEIVER_PORT")
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

	fmt.Println("Listening for messages")
	err = s.Receive(ctx, func(c context.Context, msg *pubsub.Message) {
		fmt.Println("About to send")
		msg.Ack()

		var i = new(messages.Message)
		json.Unmarshal(msg.Data, &i)

		i.SentAt = time.Now()

		err := sendMessage(sendIp, sendPort, i)
		if err != nil {
			panic(err)
		}

		fmt.Println("Sent")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}

func sendMessage(sendIp, sendPort string, msg *messages.Message) error {
	o, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	b := bytes.NewBuffer(o)
	sendResp, err := http.Post(fmt.Sprintf("http://%s:%s", sendIp, sendPort), "text/plain", b)
	if err != nil {
		return err
	}

	if sendResp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Expected status 200, got %d\n", sendResp.StatusCode))
	}

	return nil
}
