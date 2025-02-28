#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

TID=${TID:-$(uuidgen)}
OID=${OID:-$(uuidgen)}

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_GLOBAL_URL_PREFIX}/v1/cloudaccounts --data-binary @- \
| jq .
{
  "name":"${CLOUDACCOUNTNAME}",
  "owner":"${CLOUDACCOUNTNAME}",
  "tid":"${TID}",
  "oid":"${OID}",
  "type":"ACCOUNT_TYPE_INTEL"
}
EOF
