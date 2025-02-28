#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"
NODE_ID=${NODE_ID:--1}

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X GET \
${IDC_REGIONAL_URL_PREFIX}/v1/fleetadmin/nodes/${NODE_ID}/instancetypestats \
| jq .
EOF
