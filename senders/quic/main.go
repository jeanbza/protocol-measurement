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
