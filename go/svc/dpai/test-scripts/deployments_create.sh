#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# INTEL CONFIDENTIAL
# Copyright (C) 2024 Intel Corporation
echo $PWD
source ./go/svc/dpai/test-scripts/common.sh


API_URL="$HOST_URL/v1/cloudaccounts/$CLOUDACCOUNT_ID/dpai/deployments"

# Use the token to make a request to the createKF API
MAIN_API_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $ACCESS_TOKEN" --data '{
    "workspaceId": "dummy-workspace-id",
    "serviceId":"dummy-service-id",
    "serviceType":"HMS",
    "changeIndicator":"DPAI_ACCEPTED",
    "createdBy":"venkad",
    "input":"{}"
}
' "$API_URL")

# Process the response from the main API as needed
echo "Main API Response: $MAIN_API_RESPONSE"