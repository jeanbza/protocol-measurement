package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
)

func main() {
	sendIp := os.Getenv("HTTP_RECEIVER_IP")
	if sendIp == "" {
		panic("Expected to receive a port with environment variable HTTP_RECEIVER_IP")
	}

	sendPort := os.Getenv("HTTP_RECEIVER_PORT")
	if sendPort == "" {
		panic("Expected to receive a port with environment variable HTTP_RECEIVER_PORT")
	}

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		b := bytes.NewBuffer([]byte("10"))
		sendResp, err := http.Post(fmt.Sprintf("http://%s:%s", sendIp, sendPort), "text/plain", b)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
			return
		}

		if sendResp.StatusCode != 200 {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(fmt.Sprintf("Expected 200, got %d", sendResp.StatusCode)))
			return
		}

		bodyBytes, err := ioutil.ReadAll(sendResp.Body)
		if err != nil {
			panic(err)
		}

		resp.Write([]byte(fmt.Sprintf("Wrote 10 messages! Resp: %s", string(bodyBytes))))
	})

	http.ListenAndServe(":8080", nil)
}
