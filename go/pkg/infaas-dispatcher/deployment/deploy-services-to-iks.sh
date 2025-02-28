#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

regionName="$1"
if [ -z "$regionName" ]; then
  echo "Error: region name not provided."
  echo "Usage: $0 <region_name>"
  exit 1
fi

echo This script is designed to run from the root of IDC...
pushd go/pkg/infaas-dispatcher/deployment


echo reinstall dispatcher...
#helm uninstall infaas-dispatcher -n idcs-system
helm upgrade --install infaas-dispatcher ./infaas-dispatcher -n idcs-system --set region=$regionName,image.pullPolicy=Always
echo =======================================
echo
echo

echo creating inference services...

echo creating infaas-inference-llama-3-1-70b service...
helm uninstall infaas-inference-llama-3-1-70b -n idcs-system
helm install infaas-inference-llama-3-1-70b ./infaas-inference -n idcs-system --set region=$regionName,deployedModel="meta-llama/Meta-Llama-3.1-70B-Instruct" --set image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating infaas-inference-qwen-2-5-32b service...
helm uninstall infaas-inference-qwen-2-5-32b -n idcs-system
helm install infaas-inference-qwen-2-5-32b ./infaas-inference -n idcs-system --set region=$regionName,deployedModel="Qwen/Qwen2.5-32B-Instruct" --set replicaCount=2 --set image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating infaas-inference-qwen-2-5-coder-32b service...
helm uninstall infaas-inference-qwen-2-5-coder-32b -n idcs-system
helm install infaas-inference-qwen-2-5-coder-32b ./infaas-inference -n idcs-system --set region=$regionName,deployedModel="Qwen/Qwen2.5-Coder-32B-Instruct" --set replicaCount=2 --set image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating infaas-inference-llama-3-1-8b service...
helm uninstall infaas-inference-llama-3-1-8b -n idcs-system
helm install infaas-inference-llama-3-1-8b ./infaas-inference -n idcs-system --set region=$regionName,deployedModel="meta-llama/Meta-Llama-3.1-8B-Instruct" --set replicaCount=2 --set image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating infaas-inference-mistral-7b service...
helm uninstall infaas-inference-mistral-7b -n idcs-system
helm install infaas-inference-mistral-7b ./infaas-inference -n idcs-system --set region=$regionName,deployedModel="mistralai/Mistral-7B-Instruct-v0.1" --set replicaCount=2 --set image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating safeguard service...
helm uninstall infaas-safeguard -n idcs-system
helm install infaas-safeguard ./infaas-safeguard -n idcs-system --set replicaCount=4 --set precision=bf16
echo =======================================
echo
echo

