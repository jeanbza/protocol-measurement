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
