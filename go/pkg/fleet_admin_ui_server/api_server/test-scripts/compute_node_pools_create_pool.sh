#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

: "${COMPUTE_NODE_POOL_ID:?environment variable is required}"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/fleetadmin/computenodepools/${COMPUTE_NODE_POOL_ID} --data-binary @- \
| jq .
{
  "poolName": "${COMPUTE_NODE_POOL_ID}",
  "poolAccountManagerAgsRole": "${POOL_ACCOUNT_MANAGER_AGS_ROLE}"
}
EOF
