FROM golang:1.8-alpine
RUN apk add --no-cache bash git openssh

ADD . /go/src/app
RUN go get -u cloud.google.com/go/pubsub
RUN go install app

FROM alpine:latest
COPY --from=0 /go/bin/app .
COPY --from=0 /go/src/app/creds.json /tmp/creds.json

ENV GCP_PROJECT_ID deklerk-sandbox
ENV PORT 8080
ENV GOOGLE_APPLICATION_CREDENTIALS "/tmp/creds.json"
CMD ["./app"]
