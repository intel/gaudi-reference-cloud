#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

KIND_GATEWAY=${KIND_GATEWAY:-172.18.0.1}
SECRETS_DIR=${SECRETS_DIR:-${SCRIPT_DIR}/../../../../local/secrets}
OPERATOR_DIR=${OPERATOR_DIR:-${SECRETS_DIR}/vm-cluster-operator}

export COMMON_NAME=${USER}
export REGIONAL_GRPC_API=dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local
grpcurl \
  --cacert local/secrets/pki/${COMMON_NAME}/ca.pem \
  --cert local/secrets/pki/${COMMON_NAME}/cert.pem \
  --key local/secrets/pki/${COMMON_NAME}/cert.key \
  -d '{"vNetReference":{"cloudAccountId":"'"${CLOUDACCOUNT}"'","name":"'"${VNETNAME}"'"},"addressReference":{"addressConsumerId":"mgmt"}}' \
  ${REGIONAL_GRPC_API}:443 \
  proto.VNetPrivateService/ReleaseAddress | jq .

curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X DELETE \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/vnets/name/${VNETNAME} \
| jq .

cluster_name=$(cat ${OPERATOR_DIR}/cluster.json | jq -r '.metadata.name')
kubectl --context ${cluster_name} patch FelixConfiguration default --type='json' -p='[{"op": "remove", "path": "/spec/externalNodesList"}]' || true

kind_bridge=$(ip addr | grep ${KIND_GATEWAY} | awk '{print $7}')
for id in {201..202}; do
  ${SCRIPT_DIR}/vlan.sh down 192.168.${id}.0/24 ${id} ${kind_bridge}
done
