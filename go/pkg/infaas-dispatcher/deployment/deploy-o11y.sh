#!/usr/bin/env bash

eval "$(teller sh)"

otelNamespace="idcs-otel"
prometheusNamespace="idcs-prometheus"

envName="$1"
if [ -z "$envName" ]; then
  echo "Error: environment name not provided."
  echo "Usage: $0 <env_name>"
  exit 1
fi


promEnvValuesFilePath="./o11y/prom-values-$envName.yaml"
if [ ! -f "$promEnvValuesFilePath" ]; then
  echo "Error: values file not found at $promEnvValuesFilePath for environment $envName."
  exit 1
fi


promEnvCaCertPath="./o11y/certs/prom-cacert-$envName"
if [ ! -f "$promEnvCaCertPath" ]; then
  echo "Error: CA CERT file not found at $promEnvCaCertPath for environment $envName."
  exit 1
fi


otelEnvValuesFilePath="./o11y/otel-values-$envName.yaml"
if [ ! -f "$otelEnvValuesFilePath" ]; then
  echo "Error: values file not found at $otelEnvValuesFilePath for environment $envName."
  exit 1
fi


otelEnvCaCertFilePath="./o11y/certs/otel-cacert-$envName"
if [ ! -f "$otelEnvCaCertFilePath" ]; then
  echo "Error: CA CERT file not found at $otelEnvCaCertFilePath for environment $envName."
  exit 1
fi


echo deploy prometheus stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install prometheus -n $prometheusNamespace \
  prometheus-community/kube-prometheus-stack \
   -f ./o11y/prom-values.yaml \
   -f "$promEnvValuesFilePath"
echo =======================================
echo
echo

echo creating $otelNamespace namespace...
kubectl create namespace $otelNamespace
kubectl label namespace $otelNamespace istio-injection=enabled
echo =======================================
echo
echo

echo deploy otel-agent...
helm upgrade --install otel-agent -n $otelNamespace \
  oci://amr-idc-registry-pre.infra-host.com/idc-observability/opentelemetry-collector-agent \
  -f ./o11y/otel-values.yaml \
  -f "$otelEnvValuesFilePath"
echo =======================================
echo
echo