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

ORCHESTRATOR_PORT=8080

HTTP_RECEIVER_PORT=8081
UDP_RECEIVER_PORT=8082
UNARY_GRPC_RECEIVER_PORT=8083
STREAMING_GRPC_RECEIVER_PORT=8084
STREAMING_WEBSOCKET_RECEIVER_PORT=8085
QUIC_RECEIVER_PORT=8086

HTTP_SENDER_PORT=8071
UDP_SENDER_PORT=8072
UNARY_GRPC_SENDER_PORT=8073
STREAMING_GRPC_SENDER_PORT=8074
STREAMING_WEBSOCKET_SENDER_PORT=8075
QUIC_SENDER_PORT=8076

PROJECT_NAME=$1

function deploy {
    NAME=$1
    PORT=$2
    PROTOCOL=$3

    GOOS=linux go build .

    kubectl delete deployment $NAME | true
    kubectl delete service $NAME | true
    docker build --no-cache -t gcr.io/${PROJECT_ID}/$NAME:$VERSION .
    gcloud docker -- push gcr.io/${PROJECT_ID}/$NAME:$VERSION
    kubectl run $NAME --image=gcr.io/${PROJECT_ID}/$NAME:$VERSION --port 8080
    kubectl expose deployment $NAME --type=LoadBalancer --port=$PORT --target-port=8080 --protocol=$PROTOCOL
    kubectl set-env deploymenhts/$NAME GCP_PROJECT_ID=$PROJECT_NAME

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
    npm install
    ./node_modules/.bin/webpack --progress
    deploy orchestrator $ORCHESTRATOR_PORT TCP
    rm orchestrator
popd

pushd receivers/http
    deploy http-receiver $HTTP_RECEIVER_PORT TCP
    HTTP_RECEIVER_IP=$IP
    rm http
popd

pushd receivers/udp
    deploy udp-receiver $UDP_RECEIVER_PORT UDP
    UDP_RECEIVER_IP=$IP
    rm udp
popd

pushd receivers/unary_grpc
    deploy unary-grpc-receiver $UNARY_GRPC_RECEIVER_PORT TCP
    UNARY_GRPC_RECEIVER_IP=$IP
    rm unary_grpc
popd

pushd receivers/streaming_grpc
    deploy streaming-grpc-receiver $STREAMING_GRPC_RECEIVER_PORT TCP
    STREAMING_GRPC_RECEIVER_IP=$IP
    rm streaming_grpc
popd

pushd receivers/streaming_websocket
    deploy streaming-websocket-receiver $STREAMING_WEBSOCKET_RECEIVER_PORT TCP
    STREAMING_WEBSOCKET_RECEIVER_IP=$IP
    rm streaming_websocket
popd

# TODO: these don't need to be exposed - maybe add a flag to ignore?
gcloud container clusters get-credentials c-node-cluster --zone asia-south1-c --project $PROJECT_NAME
pushd senders/http
    deploy http-sender $HTTP_SENDER_PORT TCP
    kubectl set env deployments/http-sender HTTP_RECEIVER_IP=$HTTP_RECEIVER_IP
    kubectl set env deployments/http-sender HTTP_RECEIVER_PORT=$HTTP_RECEIVER_PORT
    rm http
popd

pushd senders/udp
    deploy udp-sender $UDP_SENDER_PORT TCP
    kubectl set env deployments/udp-sender UDP_RECEIVER_IP=$UDP_RECEIVER_IP
    kubectl set env deployments/udp-sender UDP_RECEIVER_PORT=$UDP_RECEIVER_PORT
    rm udp
popd

pushd senders/unary_grpc
    deploy unary-grpc-sender $UNARY_GRPC_SENDER_PORT TCP
    kubectl set env deployments/unary-grpc-sender UNARY_GRPC_RECEIVER_IP=$UNARY_GRPC_RECEIVER_IP
    kubectl set env deployments/unary-grpc-sender UNARY_GRPC_RECEIVER_PORT=$UNARY_GRPC_RECEIVER_PORT
    rm unary_grpc
popd

pushd senders/streaming_grpc
    deploy streaming-grpc-sender $STREAMING_GRPC_SENDER_PORT TCP
    kubectl set env deployments/streaming-grpc-sender STREAMING_GRPC_RECEIVER_IP=$STREAMING_GRPC_RECEIVER_IP
    kubectl set env deployments/streaming-grpc-sender STREAMING_GRPC_RECEIVER_PORT=$STREAMING_GRPC_RECEIVER_PORT
    rm streaming_grpc
popd

pushd senders/streaming_websocket
    deploy streaming-websocket-sender $STREAMING_WEBSOCKET_SENDER_PORT TCP
    kubectl set env deployments/streaming-websocket-sender STREAMING_WEBSOCKET_RECEIVER_IP=$STREAMING_WEBSOCKET_RECEIVER_IP
    kubectl set env deployments/streaming-websocket-sender STREAMING_WEBSOCKET_RECEIVER_PORT=$STREAMING_WEBSOCKET_RECEIVER_PORT
    rm streaming_websocket
popd

echo "Deployed version $VERSION"