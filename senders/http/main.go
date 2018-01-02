package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"deklerk-startup-project"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
	"os"
)

const (
	topicName = "send_queue"
)

var (
	id = fmt.Sprintf("http-receiver-%s", uuid.NewV4().String())
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

	fmt.Println("Creating subscription", id)
	s, err := client.CreateSubscription(ctx, id, pubsub.SubscriptionConfig{Topic: t})
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for messages")
	err = s.Receive(ctx, func(c context.Context, msg *pubsub.Message) {
		fmt.Println("Got message!")

		var i = new(messages.Message)
		json.Unmarshal(msg.Data, &i)

		fmt.Println(i)

		err := sendMessage(sendIp, sendPort, msg.Data)
		if err != nil {
			panic(err)
		}
		fmt.Println("Done receiving")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Done listening")
}

func sendMessage(sendIp, sendPort string, msg []byte) error {
	b := bytes.NewBuffer(msg)
	sendResp, err := http.Post(fmt.Sprintf("http://%s:%s", sendIp, sendPort), "text/plain", b)
	if err != nil {
		return err
	}

	if sendResp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Expected status 200, got %d\n", sendResp.StatusCode))
	}

	return nil
}

func cleanup(s *pubsub.Subscription) {
	fmt.Println("Deleting subscription")
	if s == nil {
		return
	}

	exists, err := s.Exists(context.Background())
	if err != nil {
		panic(err)
	}

	if exists {
		err := s.Delete(context.Background())
		if err != nil {
			panic(err)
		}
	}
}
