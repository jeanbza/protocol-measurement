package main

import (
	"bytes"
	"deklerk-startup-project"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devsisters/goquic"
	"net/http"
	"os"
	"time"
)

var (
	c *http.Client
)

func main() {
	sendIp := os.Getenv("QUIC_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable QUIC_RECEIVER_IP")
	}

	sendPort := os.Getenv("QUIC_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable QUIC_RECEIVER_PORT")
	}

	c = &http.Client{
		Transport: goquic.NewRoundTripper(false),
	}

	qs := quicSender{sendIp, sendPort}

	messages.NewSendHost(&qs, sendIp, sendPort).Start()
}

type quicSender struct {
	sendIp   string
	sendPort string
}

func (qs *quicSender) SendMessage(sendRequest *messages.SendRequest) error {
	for i := 0; i < sendRequest.Amount; i++ {
		m := messages.Message{
			RunId:  sendRequest.RunId,
			SentAt: time.Now(),
		}

		o, err := json.Marshal(m)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer(o)
		sendResp, err := c.Post(fmt.Sprintf("https://%s:%s", qs.sendIp, qs.sendPort), "text/plain", b)
		if err != nil {
			return err
		}

		if sendResp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Expected status 200, got %d\n", sendResp.StatusCode))
		}
	}

	return nil
}
