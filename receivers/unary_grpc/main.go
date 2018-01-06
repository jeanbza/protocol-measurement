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

	go repeatedlySaveToSpanner(ctx, client, insertQueue)

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
		[]interface{}{id.String(), in.RunId, "unary_grpc", time.Unix(in.CreatedAt, 0), time.Unix(in.SentAt, 0), time.Now()},
	)

	return &messages.Empty{}, nil
}

func repeatedlySaveToSpanner(ctx context.Context, client *spanner.Client, insertQueue <-chan (*spanner.Mutation)) {
	ticker := time.NewTicker(time.Second)
	toBeSent := []*spanner.Mutation{}

	for {
		select {
		case <-ticker.C:
			if len(toBeSent) == 0 {
				break
			}
			fmt.Println("Saving", len(toBeSent))
			_, err := client.Apply(ctx, toBeSent)
			if err != nil {
				panic(err)
			}
			toBeSent = []*spanner.Mutation{}
		case i := <-insertQueue:
			toBeSent = append(toBeSent, i)
		}
	}
}
