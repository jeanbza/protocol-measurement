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
