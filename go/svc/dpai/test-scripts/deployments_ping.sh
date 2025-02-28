#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

export no_proxy=${no_proxy},.kind.local


# API endpoint to get the token

# TOKEN_URL="https://dev.oidc.cloud.intel.com.kind.local"
# API_URL="https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"
# TOKEN_API="$TOKEN_URL/token?email=admin@intel.com&groups=IDC.Admin"

# # API endpoint to use the token
# CLOUDACCOUNT_ID="033395876667"  #"$1"
# CREATE_KF_API="$API_URL/v1/cloudaccounts/$CLOUDACCOUNT_ID/dpai/deployments"


source ./go/svc/dpai/test-scripts/common.sh

forwardPort us-dev-1-grpc-proxy-external 8443 8443

# Use the token to make a request to the createKF API
MAIN_API_RESPONSE=$(grpcurl -insecure -H "Authorization: Bearer $ACCESS_TOKEN" localhost:8443 proto.DpaiDeploymentService/Ping 2>&1 || true)

# Process the response from the main API as needed
echo "Main API Response: $MAIN_API_RESPONSE"

exit 0