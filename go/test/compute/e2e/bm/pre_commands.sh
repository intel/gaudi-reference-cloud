#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# This script is running commands outside the Bazel sandbox.
set -ex

QCOW_FILE="idcs_domain/bmaas/bmvs/playbooks/roles/http_server/files/ubuntu-22.04-server-cloudimg-amd64-latest.qcow2"

# Installing dependencies for the bm stack
make install-requirements
make install-interactive-tools

# Set GUEST_HOST_DEPLOYMENTS to the desired number of nodes
export GUEST_HOST_DEPLOYMENTS=12

# Remove any pre-setup stack
make teardown-bmvs

# Setup the virutal stack
make setup-bmvs

# Installing the OS images for the BM
if [ -e "$QCOW_FILE" ]; then
    echo "OS images already exists. Skipping download."
else
    echo "OS images does not exist! Downloading now..."
    pushd idcs_domain/bmaas/bmvs/playbooks/roles/http_server/files
    wget https://internal-placeholder.com/artifactory/intelcloudservices-or-local/images/ubuntu-22.04-server-cloudimg-amd64-latest.qcow2
    wget https://internal-placeholder.com/artifactory/intelcloudservices-or-local/images/ubuntu-22.04-server-cloudimg-amd64-latest.qcow2.md5sum
    popd

fi 

sudo apt install make gcc

user1="guest-${USER}"
user2="bmo-${USER}"

# Function to check and remove home directory
check_and_remove_user() {
    local username="$1"

    # Check if the user exists
    if id "$username" &>/dev/null; then
        # User exists, run deluser command to remove it
        echo "Removing user and home directory: $username"
        sudo deluser --remove-home "$username"
    else
        echo "User '$username' does not exist"
    fi
}

check_and_remove_user "$user1"
check_and_remove_user "$user2"

export SSH_PROXY_IP=$(hostname -f)
export SSH_USER_PASSWORD=$(uuidgen)
sudo useradd -m -p $SSH_USER_PASSWORD guest-${USER}
sudo -u guest-${USER} mkdir /home/guest-${USER}/.ssh
sudo -u guest-${USER} cp local/secrets/test-e2e-compute-bm/ssh-proxy-operator/id_rsa.pub /home/guest-${USER}/.ssh/authorized_keys
sudo useradd -m -p $SSH_USER_PASSWORD bmo-${USER}
sudo -u bmo-${USER} mkdir /home/bmo-${USER}/.ssh
sudo -u bmo-${USER} cp local/secrets/test-e2e-compute-bm/bm-instance-operator/id_rsa.pub /home/bmo-${USER}/.ssh/authorized_keys

sudo iptables -I INPUT -p tcp -m tcp --dport 6443 -j ACCEPT
sudo iptables -I INPUT -p tcp -m tcp --dport 443 -j ACCEPT

for i in $(seq 1 "$GUEST_HOST_DEPLOYMENTS"); do
  sudo iptables -I INPUT -p tcp -m tcp --dport "$((8000+i))" -j ACCEPT
done

# Export variables
export IDC_GLOBAL_URL_PREFIX='https://dev.api.cloud.intel.com.kind.local'
export IDC_REGIONAL_URL_PREFIX='https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local'
export TOKEN_URL_PREFIX='http://dev.oidc.cloud.intel.com.kind.local:80'

