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
	"deklerk-startup-project"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"time"
)

const (
	topicName = "send_queue"
)

func main() {
	sendIp := os.Getenv("STREAMING_WEBSOCKET_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable STREAMING_WEBSOCKET_RECEIVER_IP")
	}

	sendPort := os.Getenv("STREAMING_WEBSOCKET_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable STREAMING_WEBSOCKET_RECEIVER_PORT")
	}

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%s", sendIp, sendPort), Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}

	wss := websocketSender{
		q: make(chan (*messages.Message), 16384),
		c: c,
	}

	go wss.asyncWriteToWebsocket()

	messages.NewSendHost(&wss, sendIp, sendPort).Start()
}

type websocketSender struct {
	q chan (*messages.Message)
	c *websocket.Conn
}

func (w *websocketSender) asyncWriteToWebsocket() {
	for {
		msg := <-w.q
		err := w.c.WriteJSON(msg)
		if err != nil {
			panic(err)
		}
	}
}

func (wss *websocketSender) SendMessage(sendRequest *messages.SendRequest) error {
	for i := 0; i < sendRequest.Amount; i++ {
		m := messages.Message{
			RunId:  sendRequest.RunId,
			SentAt: time.Now(),
		}

		wss.q <- &m
	}

	return nil
}
