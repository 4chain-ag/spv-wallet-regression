#!/usr/bin/env bash

read -sp "Enter your cloudflare dns edit access token: " CLOUDFLARE_TOKEN

sudo microk8s kubectl create secret generic cloudflare-api-token-secret \
  --namespace cert-manager \
  --from-literal=api-token="${CLOUDFLARE_TOKEN}"

