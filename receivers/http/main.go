package main

import "net/http"

func main() {
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello world!"))
	})

	http.ListenAndServe(":8080", nil)
}
