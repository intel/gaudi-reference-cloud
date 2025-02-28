#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# API endpoint to get the token

TOKEN_URL="https://dev.oidc.cloud.intel.com.kind.local"
API_URL="https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"
TOKEN_API="$TOKEN_URL/token?email=admin@intel.com&groups=IDC.Admin"

# API endpoint to use the token
CLOUDACCOUNT_ID="$1"
CREATE_KF_API="$API_URL/v1/cloudaccounts/$CLOUDACCOUNT_ID/kfaas/deployments"




# Get the token
TOKEN_RESPONSE=$(curl -v -s -X GET \
  "$TOKEN_API")

  

# Extract the token from the response
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE)
echo $ACCESS_TOKEN

# Check if the token is obtained successfully
if [ -z "$ACCESS_TOKEN" ]; then
  echo "Failed to obtain access token."
  exit 1
fi

# Use the token to make a request to the createKF API
MAIN_API_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $ACCESS_TOKEN" --data '{
    "deploymentName": "demo-2",
    "kfVersion":"2",
    "k8sClusterID":"k8sClusterID-test",
    "k8sClusterName":"k8sClusterNametest",
    "storageClassName":"storageClassNameTest",
    "createdDate":"createdDateTest",
    "status":"create"
}
' "$CREATE_KF_API")

# Process the response from the main API as needed
echo "Main API Response: $MAIN_API_RESPONSE"

deploymentID=$(echo "$MAIN_API_RESPONSE" | jq -r '.deploymentID')
echo "Deployment ID: $deploymentID"

# # Use the token to make a request to the createKF API
# MAIN_API_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" --data '{
#     "deploymentName": "demo-2",
#     "kfVersion":"1",
#     "k8sClusterID":"k8sClusterID-test",
#     "k8sClusterName":"k8sClusterNametest",
#     "storageClassName":"storageClassNameTest",
#     "createdDate":"createdDateTest",
#     "status":"create"
# }
# ' "$CREATE_KF_API="https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local/v1/cloudaccounts/$CLOUDACCOUNT_ID/kfaas/deployments/$deploymentID/deploy"")
