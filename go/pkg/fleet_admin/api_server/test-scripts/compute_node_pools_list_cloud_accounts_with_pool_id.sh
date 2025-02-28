#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# ====================================================
# DEPRECATED: This script is no longer maintained.
# Please use scripts from go/pkg/fleet_admin_ui_server/api_server/test-scripts/ instead.
# ====================================================
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

: "${COMPUTE_NODE_POOL_ID:?environment variable is required}"

curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X GET \
${IDC_REGIONAL_URL_PREFIX}/v1/fleetadmin/computenodepools/${COMPUTE_NODE_POOL_ID}/cloudaccounts \
| jq .
