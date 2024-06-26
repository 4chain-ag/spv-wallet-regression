#!/usr/bin/env bash

#!/usr/bin/env bash

# Adding color variables for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Install required packages
echo -e "${GREEN}[+] Installing necessary packages...${NC}"
sudo apt install -y open-iscsi
sudo systemctl enable iscsid

# Check if snap is installed, install if not
if ! sudo which snap > /dev/null; then
    echo -e "${YELLOW}[!] snap not found, installing snapd...${NC}"
    sudo apt update && sudo apt install -y snapd
fi

# Install microk8s
echo -e "${GREEN}[+] Installing MicroK8s...${NC}"
sudo snap install microk8s --classic --channel=latest/stable

# Execute update_microk8s.sh to enable required addons
echo -e "${GREEN}[+] Running update_microk8s.sh...${NC}"
chmod +x ./update_microk8s.sh
sudo ./update_microk8s.sh

# Generating SSH key for Argo CD
if [ ! -f ~/.ssh/argo_github_ssh_key ]; then
    echo -e "${YELLOW}SSH key not found, generating new key...${NC}"
    ssh-keygen -t ed25519 -f ~/.ssh/argo_github_ssh_key -N ""
fi

# Setting up GHCR token
echo -e "${GREEN}[+] Setting up GHCR token to get app packages...${NC}"
./set-ghcr-token.sh

# Show public key and ask user to add to GitHub
echo -e "${YELLOW}Please add the following public key to your GitHub deploy keys:${NC}"
cat ~/.ssh/argo_github_ssh_key.pub
echo -e "${YELLOW}After adding the key, type YES to continue...${NC}"
read -r user_input
if [[ $user_input == "YES" || $user_input == "yes" ]]; then
    sudo microk8s helm repo add argo https://argoproj.github.io/argo-helm
    sudo microk8s helm repo update

    sudo microk8s helm dependency build charts/argo-cd
    echo -e "\r${GREEN}[+] Dependencies built successfully.${NC}"

    echo -e "${GREEN}[+] Installing Argo CD ...${NC}"
    argokey="$(cat ~/.ssh/argo_github_ssh_key)"
    sudo microk8s helm upgrade --install --create-namespace argocd -n argocd --set "argo-cd.configs.repositories.private-repo.sshPrivateKey=$argokey" ./charts/argo-cd

    echo -e "${GREEN}[+] Applying app of apps...${NC}"
    sudo microk8s helm upgrade argo-apps -n argocd --install charts/argo-apps
else
    echo -e "${RED}[-] Operation aborted by the user.${NC}"
    exit 1
fi

argocdpswd=$(sudo microk8s kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
printf "\n${GREEN}Argo CD username: admin\npassword: ${NC}%s\n\n" "$argocdpswd"
echo -e "${GREEN}[+] Completed!${NC}"

