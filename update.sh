#!/usr/bin/env bash
set -e

VERSION=`date +%s`
PROJECT_ID="$(gcloud config get-value project -q)"
IP="pending"
MAX_IP_ATTEMPTS=90

function update {
    NAME=$1
    GOOS=linux go build .
    docker build --no-cache -t gcr.io/${PROJECT_ID}/$NAME:$VERSION .
    gcloud docker -- push gcr.io/${PROJECT_ID}/$NAME:$VERSION
    kubectl set image deployment/$NAME $NAME=gcr.io/${PROJECT_ID}/$NAME:$VERSION

    for attempt in $( eval echo {0..$MAX_IP_ATTEMPTS} ); do
        echo "Fetch IP attempt $attempt / $MAX_IP_ATTEMPTS"
        IP=`kubectl get service $NAME --no-headers=true | awk '{print $4}'`

        if [[ $IP != *"pending"* ]]; then
          break
        fi

        sleep 1
    done

    if [[ $IP == *"pending"* ]]; then
      echo "Never got the IP!"
      exit 1
    fi
}

gcloud container clusters get-credentials b-node-cluster --zone us-central1-a --project deklerk-sandbox
pushd orchestrator
    update orchestrator
    rm orchestrator
popd

pushd receivers/http
    update http-receiver
    HTTP_RECEIVER_IP=$IP
    rm http
popd

pushd receivers/udp
    update udp-receiver
    UDP_RECEIVER_IP=$IP
    rm udp
popd

pushd receivers/unary_grpc
    update unary-grpc-receiver
    UNARY_GRPC_RECEIVER_IP=$IP
    rm unary_grpc
popd

pushd receivers/streaming_grpc
    update streaming-grpc-receiver
    STREAMING_GRPC_RECEIVER_IP=$IP
    rm streaming_grpc
popd

pushd receivers/streaming_websocket
    update streaming-websocket-receiver
    STREAMING_WEBSOCKET_RECEIVER_IP=$IP
    rm streaming_websocket
popd

gcloud container clusters get-credentials c-node-cluster --zone asia-south1-c --project deklerk-sandbox
pushd senders/http
    update http-sender
    rm http
popd

pushd senders/udp
    update udp-sender
    rm udp
popd

pushd senders/unary_grpc
    update unary-grpc-sender
    rm unary_grpc
popd

pushd senders/streaming_grpc
    update streaming-grpc-sender
    rm streaming_grpc
popd

pushd senders/streaming_websocket
    update streaming-websocket-sender
    rm streaming_websocket
popd

echo "Updated to version $VERSION"