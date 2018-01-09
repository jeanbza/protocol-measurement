# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env bash
set -e

VERSION=`date +%s`
PROJECT_ID="$(gcloud config get-value project -q)"
IP="pending"
MAX_IP_ATTEMPTS=90

if [ -z "$1" ]
then
      echo "The first argument is required, and should be the name of your Google Cloud project"
      exit 1
fi

PROJECT_NAME=$1

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

gcloud container clusters get-credentials b-node-cluster --zone us-central1-a --project $PROJECT_NAME
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

gcloud container clusters get-credentials c-node-cluster --zone asia-south1-c --project $PROJECT_NAME
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