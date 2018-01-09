package messages

import (
	"cloud.google.com/go/spanner"
	"context"
	"fmt"
	"time"
)

const maxMutationsPerTx = 1000

type SpannerSaver struct {
	client      *spanner.Client
	insertQueue <-chan (*spanner.Mutation)
}

func NewSpannerSaver(c *spanner.Client, iq <-chan (*spanner.Mutation)) *SpannerSaver {
	return &SpannerSaver{c, iq}
}

func (ss *SpannerSaver) RepeatedlySaveToSpanner(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	toBeSent := []*spanner.Mutation{}

	for {
		select {
		case <-ticker.C:
			if len(toBeSent) == 0 {
				break
			}

			for len(toBeSent) > maxMutationsPerTx {
				buf := toBeSent[:maxMutationsPerTx]
				toBeSent = toBeSent[maxMutationsPerTx:]

				fmt.Println("Saving", len(buf))
				_, err := ss.client.Apply(ctx, buf)
				if err != nil {
					panic(err)
				}
			}

			if len(toBeSent) > 0 {
				fmt.Println("Saving", len(toBeSent))
				_, err := ss.client.Apply(ctx, toBeSent)
				if err != nil {
					panic(err)
				}
			}

			toBeSent = []*spanner.Mutation{}
		case i := <-ss.insertQueue:
			toBeSent = append(toBeSent, i)
		}
	}
}
