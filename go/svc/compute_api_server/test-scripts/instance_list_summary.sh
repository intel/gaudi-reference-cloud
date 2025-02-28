#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Delete all instances in a CloudAccount.
set -e
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

curl --silent -k \
-H 'Content-type: application/json' \
-H "Authorization: Bearer ${TOKEN}" \
-X GET \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances?metadata.instanceGroupFilter=Any \
| jq -r '.items[] | .metadata.cloudAccountId + " " + .metadata.resourceId + " " + .metadata.name + " " + .status.phase + " " + .status.interfaces[0].addresses[0] + " " + .spec.instanceType'
