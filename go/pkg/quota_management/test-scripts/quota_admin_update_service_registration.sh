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
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/quota/service/${SERVICE_ID}/update --data-binary @- \
| jq .
{
  "serviceResources" : [
  {
     "name": "instance_groups",
     "quotaUnit": "COUNT",
     "maxLimit":    10
   },
    {
     "name": "instances",
     "quotaUnit": "COUNT",
     "maxLimit":    12
   }
   ]
}
EOF
