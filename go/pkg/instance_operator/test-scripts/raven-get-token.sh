#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -e
URL_PREFIX=${URL_PREFIX:-https://raven-devcloud.app.intel.com/}
: "${RAVEN_USER:?environment variable is required}"
: "${RAVEN_PASSWORD:?environment variable is required}"

cat <<EOF | \
curl -k --silent \
-H 'Content-type: application/json' \
-X POST \
${URL_PREFIX}/login --data-binary @- \
| jq -r .token
{
  "username": "${RAVEN_USER}",
  "password": "${RAVEN_PASSWORD}"
}
EOF
