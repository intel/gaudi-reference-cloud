#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

: "${CLOUDACCOUNT:?environment variable is required}"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PATCH \
${URL_PREFIX}/v1/cloudaccounts/id/${CLOUDACCOUNT} --data-binary @- \
| jq .
{
  "type": "ACCOUNT_TYPE_ENTERPRISE"
}
EOF
