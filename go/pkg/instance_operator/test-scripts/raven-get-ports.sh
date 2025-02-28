#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
URL_PREFIX=${URL_PREFIX:-https://raven-devcloud.app.intel.com/}
RAVEN_TOKEN=$(${SCRIPT_DIR}/raven-get-token.sh)

curl -vk \
-H 'Content-type: application/json' \
-H "Authorization: Bearer ${RAVEN_TOKEN}" \
${URL_PREFIX}devcloud/v2/staging/list/ports?switch_fqdn=internal-placeholder.com \
| jq .
