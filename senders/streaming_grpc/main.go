package main

import (
	"cloud.google.com/go/pubsub"
	"deklerk-startup-project"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("STREAMING_GRPC_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable STREAMING_GRPC_RECEIVER_IP")
	}

	sendPort := os.Getenv("STREAMING_GRPC_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable STREAMING_GRPC_RECEIVER_PORT")
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

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", sendIp, sendPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	grpcClient := messages.NewGrpcStreamingInputterServiceClient(conn)
	stream, err := grpcClient.MakeRequest(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for messages")
	err = s.Receive(ctx, func(c context.Context, msg *pubsub.Message) {
		fmt.Println("About to send")
		msg.Ack()

		var i = new(messages.ProtoMessage)
		json.Unmarshal(msg.Data, &i)

		i.SentAt = time.Now().Unix()

		stream.Send(i)

		fmt.Println("Sent")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}
