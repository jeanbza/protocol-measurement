# deklerk Startup Project

## Running locally

1. Install docker
1. Install [protobufs](https://github.com/golang/protobuf)
1. `go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`
1. `go get github.com/gorilla/websocket`
1. `go get golang.org/x/net/context`
1. `go get google.golang.org/grpc`
1. `go get -t -u github.com/lucas-clemente/quic-go/...`
1. `protoc *.proto --go_out=plugins=grpc:.`
1. `docker-compose up # note - protobuf generation must have already occurred (previous step)` 

## Deploying and updating deployments

1. Install `gcloud` and `kubectl`
1. [Generate credentials](https://cloud.google.com/docs/authentication/getting-started) and place json file at `pusher/creds.json`
1. Update `pusher/Dockerfile` `GCP_PROJECT_ID`

- `./deploy.sh`
- `./update.sh`

## Debugging

1. See an app's environment variables
    1. `kubectl get pods`
    1. `kubectl describe pods <some pod>`