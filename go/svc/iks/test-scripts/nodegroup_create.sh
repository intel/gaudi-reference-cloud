#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups --data-binary @- \
| jq '.'
{
        "name": "ng-defaults",
        "count": 4,
        "description": "reprehenderit",
        "vnets": [
            {
                "availabilityzonename": "us-dev-1a",
                "networkinterfacevnetname": us-dev-1a-default"
            }
        ],
        "instancetypeid": "vm-spr-tny",
        "sshkeyname": [
            "string12312"
        ],
        "tags": [
            {
                "value": "culpa",
                "key": "proiden"
            }
        ],
        "upgradestrategy": {
            "drainnodes": false,
            "maxunavailablepercentage": 20
        }
}
EOF