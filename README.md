# deklerk Startup Project

## Deploying and updating deployments

1. Install `gcloud` and `kubectl`
1. [Generate credentials](https://cloud.google.com/docs/authentication/getting-started) and place json file at `pusher/creds.json`
1. Update `pusher/Dockerfile` `GCP_PROJECT_ID`

- `./deploy.sh`
- `./update.sh`