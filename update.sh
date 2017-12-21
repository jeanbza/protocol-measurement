#!/usr/bin/env bash
set -x

VERSION=`date +%s`
PROJECT_ID="$(gcloud config get-value project -q)"
IP="pending"
MAX_IP_ATTEMPTS=90

function update {
    NAME=$1
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
pushd receivers/http
    update http-receiver
    HTTP_RECEIVER_IP=$IP
popd

gcloud container clusters get-credentials a-node-cluster --zone us-central1-a --project deklerk-sandbox
pushd senders/http
    update http-sender
    kubectl set env deployments/http-sender HTTP_RECEIVER_IP=$HTTP_RECEIVER_IP
    kubectl set env deployments/http-sender HTTP_RECEIVER_PORT=8080
popd

echo "Updated to version $VERSION"