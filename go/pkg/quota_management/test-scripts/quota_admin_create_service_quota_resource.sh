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
${IDC_REGIONAL_URL_PREFIX}/v1/quota/service/${SERVICE_ID}/create --data-binary @- \
| jq .
{
   "serviceQuotaResource": {
    "resourceType": "${RESOURCE_TYPE}",
    "quotaConfig": {
      "limits": 3,
      "quotaUnit": "COUNT"
    },
    "scope": {
      "scopeType": "QUOTA_ACCOUNT_TYPE",
      "scopeValue": "PREMIUM"
    },
    "reason": "create quota for number of filesystems for STANDARD accounts"
  }
}
EOF
