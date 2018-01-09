package messages

import "time"

type SendRequest struct {
	RunId  string `json:"runId"`
	Amount int    `json:"amount"`
}

type Message struct {
	RunId      string    `json:"runId"`
	CreatedAt  time.Time `json:"createdAt"` // Deprecated - use Runs.CreatedAt
	SentAt     time.Time `json:"sentAt"`
	ReceivedAt time.Time `json:"receivedAt"`
}
