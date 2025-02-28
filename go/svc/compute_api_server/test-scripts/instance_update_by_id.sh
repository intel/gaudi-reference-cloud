#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

: "${RESOURCEID:?environment variable is required}"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances/id/${RESOURCEID} --data-binary @- \
| jq .
{
  "spec": {
    "runStrategy": "Halted",
    "sshPublicKeyNames": [
      "${KEYNAME}"
    ]
  }
}
EOF
