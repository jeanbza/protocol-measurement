# deklerk Startup Project

## Running locally

1. Install docker
1. `docker-compose up`

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