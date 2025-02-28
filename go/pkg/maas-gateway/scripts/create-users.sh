#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

export TOKEN=$(curl "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin")

INTEL_CLOUDACCOUNTS=(
#    "anatoly.tarnavsky@intel.com"
    "john.doe@intel.com"
    "jane.smith@intel.com"
)

STANDARD_CLOUDACCOUNTS=(
#    "anatoly.tarnavsky@intel.com"
    "john.doe@gmail.com"
    "jane.smith@gmail.com"
)

echo "Adding intel users"
for CLOUDACCOUNTNAME in "${INTEL_CLOUDACCOUNTS[@]}"
do
    curl -k \
    -H 'Content-type: application/json' \
    -H "Origin: http://localhost:3001/" \
    -H "Authorization: Bearer ${TOKEN}" \
    -X POST \
    https://dev.api.cloud.intel.com.kind.local/v1/cloudaccounts \
    -d @- << EOF
    {
        "tid": "$(uuidgen)",
        "oid": "$(uuidgen)",
        "name": "${CLOUDACCOUNTNAME}",
        "owner": "${CLOUDACCOUNTNAME}",
        "type": "ACCOUNT_TYPE_INTEL",
        "billingAccountCreated": true,
        "enrolled": true,
        "paidServicesAllowed": true,
        "personId": "234569"
    }
EOF
    echo "Added $CLOUDACCOUNTNAME"
done


echo "Adding standard users"
for CLOUDACCOUNTNAME in "${STANDARD_CLOUDACCOUNTS[@]}"
do
    curl -k \
    -H 'Content-type: application/json' \
    -H "Origin: http://localhost:3001/" \
    -H "Authorization: Bearer ${TOKEN}" \
    -X POST \
    https://dev.api.cloud.intel.com.kind.local/v1/cloudaccounts \
    -d @- << EOF
    {
        "name": "${CLOUDACCOUNTNAME}",
        "owner": "${CLOUDACCOUNTNAME}",
        "tid": "$(uuidgen)",
        "oid": "$(uuidgen)",
        "type": "ACCOUNT_TYPE_STANDARD",
        "countryCode":"GB"
    }
EOF
    echo "Added $CLOUDACCOUNTNAME"
done