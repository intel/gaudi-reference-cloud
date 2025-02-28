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
${IDC_REGIONAL_URL_PREFIX}/v1/quota/service/register --data-binary @- \
| jq .
{
  "serviceName": "compute-test",
  "region": "us-staging-1",
  "serviceResources": [
        {
            "name": "test",
            "quotaUnit": "COUNT",
            "maxLimit": "1"
        }
    ]
}
EOF
