FROM alpine:latest

ADD http app
ADD creds.json /creds.json

ENV PORT 8080
ENV GOOGLE_APPLICATION_CREDENTIALS /creds.json

CMD ["./app"]
