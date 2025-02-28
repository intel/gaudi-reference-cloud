#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -x
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

# labels
labels=""
if [[ -n "$APPNAME" ]]; then
  labels='"labels": {"Application": "'$APPNAME'"},'
fi

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/network/subnets/id/${SUBNETID} --data-binary @- \
| jq .
{
  "metadata": {
    "name": "${NAME}",
    ${labels}
    "vpcId": "${VPCID}",
    "resourceVersion": "${RESOURCEVERSION}",
    "availabilityZoneId": "${AZID}"
  },
  "spec": {
    "cidrBlock": "10.0.0.0/24"
  }
}
EOF
