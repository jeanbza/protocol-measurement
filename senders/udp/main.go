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
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	sendIp := os.Getenv("UDP_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive an environment variable UDP_RECEIVER_IP")
	}

	sendPort := os.Getenv("UDP_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive an environment variable UDP_RECEIVER_PORT")
	}

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", sendIp, sendPort))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	us := udpSender{conn}

	messages.NewSendHost(&us, sendIp, sendPort).Start()
}

type udpSender struct {
	conn net.Conn
}

func (us *udpSender) SendMessage(sendRequest *messages.SendRequest) error {
	for i := 0; i < sendRequest.Amount; i++ {
		m := messages.Message{
			RunId:  sendRequest.RunId,
			SentAt: time.Now(),
		}

		outBytes, err := json.Marshal(m)
		if err != nil {
			return err
		}

		_, err = us.conn.Write(outBytes)
		if err != nil {
			return err
		}
	}

	return nil
}
