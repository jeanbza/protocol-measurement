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
	"cloud.google.com/go/spanner"
	"context"
	xcontext "golang.org/x/net/context"
	"deklerk-startup-project"
	"fmt"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"time"
)

const (
	database = "projects/deklerk-sandbox/instances/protocol-measurement/databases/protocol-measurement"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		panic("Expected to receive an environment variable PORT")
	}

	ctx := context.Background()
	insertQueue := make(chan (*spanner.Mutation), 16384)

	client, err := spanner.NewClient(ctx, database)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go messages.NewSpannerSaver(client, insertQueue).RepeatedlySaveToSpanner(ctx)

	fmt.Println("Listening!")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	r := unaryGrpcServerReplier{q: insertQueue}
	messages.RegisterGrpcUnaryInputterServiceServer(s, r)

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		panic(err)
	}

	fmt.Println("We're done here!")
}

type unaryGrpcServerReplier struct {
	q chan (*spanner.Mutation)
}

func (r unaryGrpcServerReplier) MakeRequest(ctx xcontext.Context, in *messages.ProtoMessage) (*messages.Empty, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	r.q <- spanner.Insert(
		"results",
		[]string{"id", "runId", "protocol", "createdAt", "sentAt", "receivedAt"},
		[]interface{}{id.String(), in.RunId, "grpc-unary", time.Unix(in.CreatedAt, 0), time.Unix(in.SentAt, 0), time.Now()},
	)

	return &messages.Empty{}, nil
}