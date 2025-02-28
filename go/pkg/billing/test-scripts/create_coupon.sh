#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Before running, export TOKEN to the value from https://admin.staging.console.idcservice.net/profile/apikeys.
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

: "${TOKEN:?environment variable is required}"

source "${SCRIPT_DIR}/defaults.sh"

curl -k --silent -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}" -d \
"{\"amount\": \"10000\",\"creator\": \"${USER}@intel.com\",\"start\": \"2024-03-20T16:04:20-06:00\",\"expires\": \"2025-12-31T16:04:00-06:00\",\"numUses\": \"1\",\"isStandard\": false}" \
-X POST ${IDC_GLOBAL_URL_PREFIX}/v1/billing/coupons | jq
