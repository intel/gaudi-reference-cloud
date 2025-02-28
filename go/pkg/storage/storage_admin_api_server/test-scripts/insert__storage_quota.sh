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
${IDC_REGIONAL_URL_PREFIX}/v1/storage/admin/storageQuota --data-binary @- \
| jq .
{
    "cloudAccountId": "814170029944",
    "reason": "test",
    "filesizeQuotaInTB": 1115,
    "filevolumesQuota": 1115,
    "bucketsQuota": 115
}
EOF