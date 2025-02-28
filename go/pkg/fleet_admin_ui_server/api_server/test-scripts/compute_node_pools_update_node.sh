#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

: "${NODE_ID:?environment variable is required}"

# Default values
INSTANCE_TYPE_POLICIES="${INSTANCE_TYPE_POLICIES:-false}"
INSTANCE_TYPE_VALUES="${INSTANCE_TYPE_VALUES:-[]}"

POOL_POLICIES="${POOL_POLICIES:-false}"
POOL_VALUES="${POOL_VALUES:-[]}"

cat <<EOF | \
curl -vk \
-H 'Content-Type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
"${IDC_REGIONAL_URL_PREFIX}/v1/fleetadmin/nodes/${NODE_ID}" \
--data-binary @- | jq .
{
  "instanceTypesOverride": {
    "overridePolicies": ${INSTANCE_TYPE_POLICIES},
    "overrideValues": ${INSTANCE_TYPE_VALUES}
  },
  "computeNodePoolsOverride": {
    "overridePolicies": ${POOL_POLICIES},
    "overrideValues": ${POOL_VALUES}
  },
  "nodeId": "${NODE_ID}",
  "region": "${REGION}",
  "availabilityZone": "${AVAILABILITY_ZONE}"
}
EOF
