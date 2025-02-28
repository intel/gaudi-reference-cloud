#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

eval "$(teller sh)"

envName="$1"
if [ -z "$envName" ]; then
  echo "Error: environment name not provided."
  echo "Usage: $0 <env_name> <region>"
  exit 1
fi

regionName="$2"
if [ -z "$regionName" ]; then
  regionName="dev"
  echo "Error: region name not provided; defaulting to $regionName"
fi


echo installing vault-secrets-operator...
helm install vault-secrets-operator hashicorp/vault-secrets-operator -n vault-secrets-operator-system \
  --create-namespace --values vault/vault-operator-values.yaml --values vault/vault-operator-values-$envName.yaml \
  --set controller.manager.clientCache.storageEncryption.kubernetes.role=${regionName}-maas-vault-operator-role
echo =======================================
echo
echo

echo installing infaas-resources...
helm upgrade --install infaas-resources ./infaas-resources --set region=$regionName
echo =======================================
echo
echo


echo installing nginx ingress...
ingressNginxOverrideArgs="-f ingress/ingress-nginx-values.yaml"
if [ "$regionName" = "dev" ]; then
  ingressNginxOverrideArgs="--set controller.hostNetwork=true"
fi


echo installing istio service mesh...
istioctl install -y -f ingress/istio-operator-values.yaml
echo =======================================
echo
echo


helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx $ingressNginxOverrideArgs
echo =======================================
echo
echo

echo installing o11y stack...
./deploy-o11y.sh $envName
echo =======================================
echo
echo

echo enforce mTLS for prometheus
kubectl apply -f ./o11y/prom-strict-mtls.yaml
echo =======================================
echo
echo