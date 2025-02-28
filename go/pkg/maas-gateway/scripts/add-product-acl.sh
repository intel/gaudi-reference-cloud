#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

export TOKEN=$(curl "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin")

CLOUDACCOUNTID="513861623936"

echo "Adding product acl"
    curl -k \
    -H 'Content-type: application/json' \
    -H "Origin: http://localhost:3001/" \
    -H "Authorization: Bearer ${TOKEN}" \
    -X POST \
    https://dev.api.cloud.intel.com.kind.local/v1/products/acl/add \
    -d @- << EOF
    {
      "adminName": "admin@intel.com",
      "cloudaccountId": "${CLOUDACCOUNTID}",
      "created": "2024-06-10T05:25:32.095Z",
      "productId": "ba5d2874-dc83-425e-af98-c810f11dad79",
      "vendorId": "4015bb99-0522-4387-b47e-c821596dc735",
      "familyId": "4004b07c-14b8-446a-8e63-ed7f1508ee1b"
    }
EOF
    echo "Added access to product"