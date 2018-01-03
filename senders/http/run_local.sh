#!/usr/bin/env bash
export GCP_PROJECT_ID="deklerk-sandbox"
export GOOGLE_APPLICATION_CREDENTIALS="/Users/deklerk/workspace/go/src/deklerk-startup-project/pusher/creds.json"

gcloud container clusters get-credentials b-node-cluster --zone us-central1-a --project deklerk-sandbox

export HTTP_RECEIVER_IP=`kubectl get services | grep http-receiver | awk '{print $4}'`
export HTTP_RECEIVER_IP=127.0.0.1
export HTTP_RECEIVER_PORT=8081

go run main.go
