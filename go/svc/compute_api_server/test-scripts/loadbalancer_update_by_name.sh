#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/loadbalancers/name/${NAME} --data-binary @- \
| jq .
{
 "spec": {
    "listeners": [{
      "port": 80,
      "pool": {
        "port": "9090",
        "monitor": "tcp",
        "loadbalancingmode": "roundRobin",
        "instanceSelectors": {
          "key1": "value1"
        }
      }
    }],
    "security": { 
      "sourceips": [
        "174.203.100.19",
        "174.203.100.18"
      ]
    },
    "port": "${LB_PORT}"
    }
}
EOF
