# Protocol measurement

This small project is meant as a visualization of the time it takes to send large amounts of messages over several
protocols. This is _not_ intended to be a definitive measurement of the included protocols.

## Running locally

1. Install [docker](https://www.docker.com/get-docker)
1. Install [protobuf](https://github.com/golang/protobuf)
1.

    ```
    go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
    go get github.com/gorilla/websocket
    go get golang.org/x/net/context
    go get google.golang.org/grpc
    go get -t -u github.com/lucas-clemente/quic-go/...
    ```

1. `protoc *.proto --go_out=plugins=grpc:.`
1. `docker-compose up # note - protobuf generation must have already occurred (previous step)` 

## Deploying and updating deployments

1. Install [npm (comes with nodejs)](https://nodejs.org/en/download/)
1. Install [docker](https://www.docker.com/get-docker)
1. Install [gcloud](https://cloud.google.com/sdk/gcloud/)
1. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
1. Create a [google cloud project](console.cloud.google.com)
1. Create two GKE clusters
1. Update `deploy.sh` and `update.sh` with the names and regions of your clusters (look for the 
`gcloud container clusters ...` statements)
1. [Generate credentials](https://cloud.google.com/docs/authentication/getting-started) and copy the file into
every subfolder (`/`, `/orchestrator`, `/receivers/http`, etc.). This will be cleaned up at some point, but for now
that's the simplest way to deploy :)

- `./deploy.sh <your-google-cloud-project-name>`
- `./update.sh <your-google-cloud-project-name>`

## Debugging

- See an app's IP addresses `kubectl get services`
- See an app's environment variables
    1. `kubectl get pods`
    1. `kubectl describe pods <some pod>`