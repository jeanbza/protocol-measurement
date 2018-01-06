#!/usr/bin/env bash
set -e

VERSION=`date +%s`
PROJECT_ID="$(gcloud config get-value project -q)"
IP="pending"
MAX_IP_ATTEMPTS=90

ORCHESTRATOR_PORT=8080

HTTP_RECEIVER_PORT=8081
UDP_RECEIVER_PORT=8082

HTTP_SENDER_PORT=8071
UDP_SENDER_PORT=8072

function deploy {
    NAME=$1
    PORT=$2

    GOOS=linux go build .

    kubectl delete deployment $NAME | true
    kubectl delete service $NAME | true
    docker build --no-cache -t gcr.io/${PROJECT_ID}/$NAME:$VERSION .
    gcloud docker -- push gcr.io/${PROJECT_ID}/$NAME:$VERSION
    kubectl run $NAME --image=gcr.io/${PROJECT_ID}/$NAME:$VERSION --port 8080
    kubectl expose deployment $NAME --type=LoadBalancer --port $PORT --target-port 8080

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
    npm install
    ./node_modules/.bin/webpack --progress
    deploy orchestrator $ORCHESTRATOR_PORT
    rm orchestrator
popd

pushd receivers/http
    deploy http-receiver $HTTP_RECEIVER_PORT
    HTTP_RECEIVER_IP=$IP
    rm http
popd

pushd receivers/udp
    deploy udp-receiver $UDP_RECEIVER_PORT
    UDP_RECEIVER_IP=$IP
    rm udp
popd

gcloud container clusters get-credentials a-node-cluster --zone us-central1-a --project deklerk-sandbox
pushd senders/http
    deploy http-sender $HTTP_SENDER_PORT
    kubectl set env deployments/http-sender HTTP_RECEIVER_IP=$HTTP_RECEIVER_IP
    kubectl set env deployments/http-sender HTTP_RECEIVER_PORT=$HTTP_RECEIVER_PORT
    rm http
popd

pushd senders/udp
    deploy udp-sender $UDP_SENDER_PORT
    kubectl set env deployments/udp-sender UDP_RECEIVER_IP=$UDP_RECEIVER_IP
    kubectl set env deployments/udp-sender UDP_RECEIVER_PORT=$UDP_RECEIVER_PORT
    rm udp
popd

echo "Deployed version $VERSION"