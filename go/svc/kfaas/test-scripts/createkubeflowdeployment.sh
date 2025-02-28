#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
# source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/kfaas/deployments --data-binary @- \
| jq '.'
{
    "deploymentName": "test-1",
    "kfVersion":"kf-version-test",
    "k8sClusterID":"k8sClusterID-test",
    "k8sClusterName":"k8sClusterNametest",
    "storageClassName":"storageClassNameTest",
    "createdDate":"createdDateTest",
    "status":"create"
}
EOF

