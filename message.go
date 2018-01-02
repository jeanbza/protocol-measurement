package messages

import "time"

type Message struct {
	Set        string    `json:"set"`
	CreatedAt  time.Time `json:"createdAt"`
	SentAt     time.Time `json:"sentAt"`
	ReceivedAt time.Time `json:"receivedAt"`
}
