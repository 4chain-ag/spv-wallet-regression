#!/usr/bin/env bash

# Get my public IP
ip=$(curl -s ifconfig.me || exit 1)

# Update microk8s
sudo snap refresh microk8s --classic

sudo microk8s enable community
sudo microk8s enable dns
sudo microk8s enable dashboard
sudo microk8s enable helm
sudo microk8s enable helm3
sudo microk8s enable openebs
sudo microk8s enable cert-manager
echo enabling metallb with ip: "${ip}"
sudo microk8s enable metallb "${ip}/32"
sudo microk8s enable traefik --set="additionalArguments={--serverstransport.insecureskipverify=true,--providers.kubernetesingress.ingressendpoint.publishedservice=traefik/traefik}"
