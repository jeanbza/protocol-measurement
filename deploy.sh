#!/usr/bin/env bash
set -e

VERSION=`date +%s`
PROJECT_ID="$(gcloud config get-value project -q)"
IP="pending"
MAX_IP_ATTEMPTS=90

HTTP_RECEIVER_PORT=8081
HTTP_SENDER_PORT=8082

function deploy {
    NAME=$1
    PORT=$2

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
pushd receivers/http
    deploy http-receiver $HTTP_RECEIVER_PORT
    HTTP_RECEIVER_IP=$IP
popd

gcloud container clusters get-credentials a-node-cluster --zone us-central1-a --project deklerk-sandbox
pushd senders/http
    deploy http-sender $HTTP_SENDER_PORT
    kubectl set env deployments/http-sender HTTP_RECEIVER_IP=$HTTP_RECEIVER_IP
    kubectl set env deployments/http-sender HTTP_RECEIVER_PORT=$HTTP_RECEIVER_PORT
popd

echo "Deployed version $VERSION"