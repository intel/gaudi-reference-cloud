#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Set kubeconfig location for a scope of the current script
export KUBECONFIG=~/.kube/configvm

# Exit immediately if a command exits with a non-zero status
set -e

# Treat unset variables as an error when substituting
set -u

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

export NAMESPACE=idcs-system

FILES=(
  "product.yaml"
  "vendor.yaml"
)

# Loop through the files
for f in "${FILES[@]}"; do
  full_path="$SCRIPT_DIR/crd-files/$f"
  if [ -f "$full_path" ]; then
    echo "Applying $full_path to namespace ${NAMESPACE}"
    kubectl apply -n "${NAMESPACE}" -f "$full_path"
  else
    echo "Error: File $full_path not found" >&2
    exit 1
  fi
done

echo "All files applied successfully"