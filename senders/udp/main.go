package main

import (
	"cloud.google.com/go/pubsub"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"net"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("UDP_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable UDP_RECEIVER_IP")
	}

	sendPort := os.Getenv("UDP_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable UDP_RECEIVER_PORT")
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

	subscriptionName := fmt.Sprintf("udp-sender-%s", subscriptionId.String())

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
	outBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", sendIp, sendPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(outBytes)
	return err
}