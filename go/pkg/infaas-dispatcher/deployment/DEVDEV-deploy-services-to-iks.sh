#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

regionName="$1"
if [ -z "$regionName" ]; then
  regionName="dev"
  echo "Error: region name not provided; defaulting to $regionName"
fi

echo This script is designed to run from the root of IDC...
pushd go/pkg/infaas-dispatcher/deployment

echo reinstall dispatcher...
helm upgrade --install infaas-dispatcher ./infaas-dispatcher -n idcs-system --set region=$regionName,image.agent.pullPolicy=Always
echo =======================================
echo
echo

echo creating inference services...

echo creating infaas-inference-mock service...
helm upgrade --install infaas-inference-mock ./infaas-inference -n idcs-system --set region=$regionName,mockInference=true,deployedModel="meta-llama/Meta-Llama-3.1-8B-Instruct" --set image.agent.pullPolicy=Always
echo =======================================
echo
echo
