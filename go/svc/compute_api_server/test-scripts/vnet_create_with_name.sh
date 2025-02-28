#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/vnets --data-binary @- \
| jq .
{
  "metadata": {
    "name": "${VNETNAME}"
  },
  "spec": {
    "region": "${REGION}",
    "availabilityZone": "${AZONE}",
    "prefixLength": ${PREFIXLENGTH}
  }
}
EOF
