#!/usr/bin/env bash

set -e

script_dir=$(cd "$(dirname "$0")" && pwd)

NAME=netbox
NAMESPACE=idcs-enrollment

values_file="$script_dir/values.yaml"

if [ "$KIND_CLUSTER" = true ]; then
  values_file="$script_dir/kind-values.yaml"
fi

install_netbox() {
  helm repo add bootc https://charts.boo.tc
  helm upgrade $NAME bootc/netbox \
    --install \
    --version 4.1.1 \
    --create-namespace \
    --namespace $NAMESPACE \
    --values "$values_file"

  kubectl apply -f "$script_dir/rbac.yaml"
}

wait_netbox() {
  kubectl wait pod \
    --namespace $NAMESPACE  \
    --selector app.kubernetes.io/instance=$NAME \
    --for=condition=Ready \
    --timeout=5m
}

print_login() {
  echo "---- Default Login ----"
  echo "  username: admin"
  echo "  password: $(kubectl get secret -n $NAMESPACE $NAME -o go-template='{{.data.superuser_password | base64decode}}')"
}

install_netbox
wait_netbox
print_login
