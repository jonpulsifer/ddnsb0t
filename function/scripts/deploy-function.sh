#!/usr/bin/env bash
function deploy(){
  gcloud functions deploy ddns \
    --trigger-http \
    --region=northamerica-northeast1 \
    --runtime=go116 \
    --entry-point DDNSCloudEventReceiver \
    --project=homelab-ng \
    --service-account=ddns-function@homelab-ng.iam.gserviceaccount.com \
    --memory=128Mi \
    --max-instances=2 \
    --set-env-vars DDNS_API_TOKEN=${1}
}

deploy "${1}"
