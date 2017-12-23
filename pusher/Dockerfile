FROM golang:1.8-alpine
RUN apk add --no-cache bash git openssh

ADD . /go/src/pusher
RUN go get -u cloud.google.com/go/pubsub
RUN go install pusher

FROM alpine:latest
RUN apk --no-cache --update add ca-certificates
COPY --from=0 /go/bin/pusher .
COPY --from=0 /go/src/pusher/creds.json /tmp/creds.json

ENV GCP_PROJECT_ID deklerk-sandbox
ENV PORT 8080
ENV GOOGLE_APPLICATION_CREDENTIALS "/tmp/creds.json"
CMD ["./pusher"]
