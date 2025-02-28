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
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/network/vpcs/id/${RESOURCEID} --data-binary @- \
| jq .
{
  "metadata": {
    "labels": {
      "key1": "value1",
      "key2": "value2"
    }
  }
}
EOF
