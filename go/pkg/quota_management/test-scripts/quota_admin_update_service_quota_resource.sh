#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
#${IDC_REGIONAL_URL_PREFIX}/v1/quota/service/${SERVICE_ID}/resource/${RESOURCE_TYPE} --data-binary '{}' \
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/quota/service/${SERVICE_ID}/resource/${RESOURCE_TYPE}/update --data-binary @- \
| jq .
{
    "ruleId": "${RULE_ID}",
    "quotaConfig": {
      "limits": 3,
      "quotaUnit": "COUNT"
    },
    "reason": "testing independent service"
}
EOF
