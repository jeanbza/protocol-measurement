package messages

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"os"
)

const (
	topicName = "send_queue"
)

type sender interface {
	SendMessage(sendRequest *SendRequest) error
}

type SendHost struct {
	s            sender
	receiverIp   string
	receiverPort string
}

func NewSendHost(s sender, receiverIp, receiverPort string) *SendHost {
	return &SendHost{s, receiverIp, receiverPort}
}

func (sh *SendHost) Start() {
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

	subscriptionName := fmt.Sprintf("sender-%s", subscriptionId.String())

	fmt.Println("Creating subscription", subscriptionName)
	s, err := client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{Topic: t})
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for messages")
	err = s.Receive(ctx, func(c context.Context, msg *pubsub.Message) {
		fmt.Println("About to send")
		msg.Ack()

		var input = new(SendRequest)
		json.Unmarshal(msg.Data, &input)

		err := sh.s.SendMessage(input)
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
