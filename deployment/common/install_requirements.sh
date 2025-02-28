#!/usr/bin/env bash 
set -e


HELM_VERSION=v3.9.4
GO_VERSION=1.23.2
KIND_VERSION=v0.17.0


install_docker() {
    if [ -x "$(command -v docker)" ]; then
        echo "docker is already installed; skip docker installation"
    else
        echo "Install docker"
        sudo apt update
        sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker ${USER}
        # setup docker proxy
        sudo mkdir -p /etc/systemd/system/docker.service.d
        sudo cp ./deployment/kind/http-proxy.conf /etc/systemd/system/docker.service.d/http-proxy.conf
        sudo systemctl daemon-reload
        sudo systemctl restart docker
        sudo gpasswd -a ${USER} docker
    fi
}

install_kind() {
    pushd /tmp
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
    popd
}

install_go() {
    pushd /tmp
    curl -LO "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    popd
}

install_helm() {
    pushd /tmp
    curl -LO "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz"
    tar -xzvf helm-${HELM_VERSION}-linux-amd64.tar.gz
    sudo mv linux-amd64/helm /usr/local/bin/helm
    popd
}

install_kubectl() {
    pushd /tmp
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    popd
}

install_docker
install_go
install_kind
install_kubectl
install_helm
