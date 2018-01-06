FROM alpine:latest

RUN apk update
RUN apk --no-cache --update add ca-certificates

ADD static static
ADD dist dist
ADD orchestrator app
ADD creds.json /creds.json

ENV PORT 8080
ENV GOOGLE_APPLICATION_CREDENTIALS /creds.json
ENV GCP_PROJECT_ID deklerk-sandbox

CMD ["./app"]
