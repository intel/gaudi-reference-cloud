#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

RANCHER_URL=${RANCHER_URL:-https://$(ip route get 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p'):8443}

# TODO This doesn't appear to work if machine is still in process of joining cluster
# Remove instance from cluster first
namespace=$(cat ${OPERATOR_DIR}/cluster.json | jq -r '.metadata.namespace')
machine=$(curl -k --netrc-file ${OPERATOR_DIR}/netrc -X GET ${RANCHER_URL}/v1/cluster.x-k8s.io.machines/${namespace} | jq -r '.data[] | select(.status.nodeRef.name=="'${NAME}'")')
name=$(echo ${machine} | jq -r '.metadata.name')
if [[ ${name} != "" ]]; then
    curl -k --netrc-file ${OPERATOR_DIR}/netrc -X DELETE $(echo ${machine} | jq -r '.links.remove')
    while [[ $(curl -k --netrc-file ${OPERATOR_DIR}/netrc -X GET ${RANCHER_URL}/v1/cluster.x-k8s.io.machines/${namespace}/${name} | jq -r '.status') != "404" ]]; do
        sleep 2
    done
fi

curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X DELETE \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances/name/${NAME} \
| jq .
