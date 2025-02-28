#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/loadbalancers --data-binary @- \
| jq .
{
  "spec": {
    "monitor": "${LB_MONITOR}",
    "pool": {
      "port": "9090",
      "instanceSelectors": {
          "type": "external",
          "lb": "true"
      }
    },
    "type": "${LB_TYPE}",
    "port": "${LB_PORT}"
  }
}
EOF
