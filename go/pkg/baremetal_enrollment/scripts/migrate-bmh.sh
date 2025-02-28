#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# assign Instance Type to existing bmh
#
# This script does the following:
# - fetches the bmh from all the namespaces.
# - fetches the GPU Count for all bmh in namespaces.
# - assigns instance type label on the basis of GPU count.
# - is only valid for bm-spr, bm-spr-pvc-1100-4, bm-icp-gaudi2 hosts.
#
# Run following commands to run script:
# cd frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/scripts
# chmod +x migrate-bmh.sh (Give the file 'execute' permission)
# ./migrate-bmh.sh
#
# Make sure to have available hosts before running the script.
# No external dependencies needed.
# This overwrites the existing instancetype labels, if any.
# Already labeled bmh are not re-labeled using this script.
#

set -e

# Constants for labels and instance types
LABEL_HOST_GPU_COUNT="cloud.intel.com/host-gpu-count"
LABEL_INSTANCE_TYPE="instance-type.cloud.intel.com"

# Instance type names
BM_SPR="bm-spr"
BM_SPR_PVC="bm-spr-pvc-1100-4"
BM_ICP_GAUDI2="bm-icp-gaudi2"

assign_instance_type_label() {
  device_name=$1
  namespace=$2
  gpu_count=$3

  case $gpu_count in
    0)
      instance_type_label="$LABEL_INSTANCE_TYPE/$BM_SPR=true"
      ;;
    4)
      instance_type_label="$LABEL_INSTANCE_TYPE/$BM_SPR_PVC=true"
      ;;
    8)
      instance_type_label="$LABEL_INSTANCE_TYPE/$BM_ICP_GAUDI2=true"
      ;;
  esac

  kubectl label --overwrite --namespace="$namespace" bmh "$device_name" "$instance_type_label"
}

# Retrieve the list of hosts
hosts=$(kubectl get bmh --all-namespaces --selector "$LABEL_HOST_GPU_COUNT" -o json)

# Iterate through the hosts and call the function
echo "$hosts" | jq -c '.items[] | {device: .metadata.name, namespace: .metadata.namespace, gpuCount: .metadata.labels."cloud.intel.com/host-gpu-count" | tonumber}' | while read -r host; do
  device_name=$(echo "$host" | jq -r '.device')
  namespace=$(echo "$host" | jq -r '.namespace')
  gpu_count=$(echo "$host" | jq -r '.gpuCount')
  assign_instance_type_label "$device_name" "$namespace" "$gpu_count"
done