// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
