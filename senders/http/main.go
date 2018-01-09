package main

import (
	"bytes"
	"deklerk-startup-project"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("HTTP_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable HTTP_RECEIVER_IP")
	}

	sendPort := os.Getenv("HTTP_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable HTTP_RECEIVER_PORT")
	}

	hs := httpSender{sendIp, sendPort}

	messages.NewSendHost(&hs, sendIp, sendPort).Start()
}

type httpSender struct {
	sendIp   string
	sendPort string
}

func (hs *httpSender) SendMessage(sendRequest *messages.SendRequest) error {
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
		sendResp, err := http.Post(fmt.Sprintf("http://%s:%s", hs.sendIp, hs.sendPort), "text/plain", b)
		if err != nil {
			return err
		}

		if sendResp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Expected status 200, got %d\n", sendResp.StatusCode))
		}
	}

	return nil
}
