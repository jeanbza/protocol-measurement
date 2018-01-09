package main

import (
	"context"
	"deklerk-startup-project"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"time"
	"io"
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

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", sendIp, sendPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	grpcClient := messages.NewGrpcStreamingInputterServiceClient(conn)
	stream, err := grpcClient.MakeRequest(context.Background())
	if err != nil {
		panic(err)
	}

	sgs := streamingGrpcSender{stream}

	messages.NewSendHost(&sgs, sendIp, sendPort).Start()
}

type streamingGrpcSender struct {
	stream messages.GrpcStreamingInputterService_MakeRequestClient
}

func (s *streamingGrpcSender) SendMessage(sendRequest *messages.SendRequest) error {
	for i := 0; i < sendRequest.Amount; i++ {
		input := messages.ProtoMessage{
			RunId:  sendRequest.RunId,
			SentAt: time.Now().Unix(),
		}

		err := s.stream.Send(&input)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Got an EOF")
				i -= 1
				continue
			}
			return err
		}
	}

	return nil
}
