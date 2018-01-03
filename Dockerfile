FROM golang:1.8-alpine
RUN apk add --no-cache bash git openssh

RUN go get -u cloud.google.com/go/...
RUN go get github.com/satori/go.uuid
RUN go get -u github.com/gorilla/mux

ADD . /go/src/deklerk-startup-project
WORKDIR /go/src/deklerk-startup-project