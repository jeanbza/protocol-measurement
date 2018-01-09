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

	us := udpSender{sendIp, sendPort}

	messages.NewSendHost(&us, sendIp, sendPort).Start()
}

type udpSender struct {
	sendIp   string
	sendPort string
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

		conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", us.sendIp, us.sendPort))
		if err != nil {
			return err
		}
		defer conn.Close()

		_, err = conn.Write(outBytes)
		if err != nil {
			return err
		}
	}

	return nil
}
