#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# This script adds K8s objects to a kind cluster that are not deployed with the Helm chart.
# Deprecated.

set -ex

script_dir=$(cd "$(dirname "$0")" && pwd)
cd "$script_dir"

export NAMESPACE=dev-namespace-${USER}
./generate_namespace.sh | kubectl apply -f -

./generate_ipaddresses.sh | kubectl apply -f -

FILES="
*vnet.yaml
*sshpublickey_*.yaml
"
for f in $FILES; do
  kubectl apply -n "${NAMESPACE}" -f "$f"
  # Must update status subresource using patch.
  kubectl patch -n "${NAMESPACE}" -f "$f" --subresource=status --type=merge --patch-file "$f" || true
done

./generate_sshpublickeys.sh | kubectl apply -n "${NAMESPACE}" -f -

FILES="
*instance_my-tiny-vm-*.yaml
"
for f in $FILES; do
  kubectl apply -n "${NAMESPACE}" -f "$f"
  # Must update status subresource using patch.
  kubectl patch -n "${NAMESPACE}" -f "$f" --subresource=status --type=merge --patch-file "$f" || true
done
