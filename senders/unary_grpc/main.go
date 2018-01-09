package main

import (
	"deklerk-startup-project"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"os"
	"time"
)

func main() {
	sendIp := os.Getenv("UNARY_GRPC_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable UNARY_GRPC_RECEIVER_IP")
	}

	sendPort := os.Getenv("UNARY_GRPC_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable UNARY_GRPC_RECEIVER_PORT")
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", sendIp, sendPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	grpcClient := messages.NewGrpcUnaryInputterServiceClient(conn)

	us := unaryGrpcSender{grpcClient}

	messages.NewSendHost(&us, sendIp, sendPort).Start()
}

type unaryGrpcSender struct {
	grpcClient messages.GrpcUnaryInputterServiceClient
}

func (us *unaryGrpcSender) SendMessage(sendRequest *messages.SendRequest) error {
	for i := 0; i < sendRequest.Amount; i++ {
		i := messages.ProtoMessage{
			RunId:  sendRequest.RunId,
			SentAt: time.Now().Unix(),
		}

		_, err := us.grpcClient.MakeRequest(context.Background(), &i)
		if err != nil {
			return err
		}
	}

	return nil
}
