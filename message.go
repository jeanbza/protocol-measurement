package messages

import "time"

type Message struct {
	RunId      string    `json:"runId"`
	CreatedAt  time.Time `json:"createdAt"`
	SentAt     time.Time `json:"sentAt"`
	ReceivedAt time.Time `json:"receivedAt"`
}
