#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SECRETS_DIR=${SECRETS_DIR:-${SCRIPT_DIR}/../../../../local/secrets}
OPERATOR_DIR=${OPERATOR_DIR:-${SECRETS_DIR}/vm-cluster-operator}

REGION=${REGION:-us-dev-1}
AZONE=${AZONE:-${REGION}a}
CLOUDACCOUNT=${CLOUDACCOUNT:-090631835287}
# The reserved subnet will have a prefix length with this value or less.
# Use PREFIXLENGTH of 24 to select one of the 192.168.{201,202}.0/24 test data subnets
PREFIXLENGTH=${PREFIXLENGTH:-24}

# Instance Defaults
VNETNAME=${VNETNAME:-${AZONE}-default}
KEYNAME=${KEYNAME:-user2@acme.com}
NAME=${NAME:-my-instance-1}
INSTANCE_GROUP=${INSTANCE_GROUP:-my-group}
INSTANCE_GROUP_SIZE=${INSTANCE_GROUP_SIZE:-4}
SSHPUBLICKEY=${SSHPUBLICKEY:-$(cat ~/.ssh/id_rsa.pub)}
INSTANCE_TYPE=${INSTANCE_TYPE:-vm-spr-sml}
MACHINE_IMAGE=${MACHINE_IMAGE:-ubuntu-2204-jammy-v20230122}

# Load Balancer Defaults
# LB_MONITOR=${LB_MONITOR:TCP}
# LB_TYPE=${LB_TYPE:EXTERNAL}
# LB_PORT=${LB_PORT:8080}

CURL_OPTS=${CURL_OPTS:--vk}

if [ "${IDC_REGIONAL_URL_PREFIX}" == "" ]; then
    IDC_REGIONAL_FQDN=dev.compute.${REGION}.api.cloud.intel.com.kind.local
    IDC_REGIONAL_PORT_FILE=${SCRIPT_DIR}/../../../../local/${REGION}_host_port_443
    IDC_REGIONAL_PORT=$(cat "${IDC_REGIONAL_PORT_FILE}")
    IDC_REGIONAL_URL_PREFIX=https://${IDC_REGIONAL_FQDN}:${IDC_REGIONAL_PORT}
fi

CLUSTER_NAME=$(cat ${OPERATOR_DIR}/cluster.json | jq -r '.metadata.name')
VAULT_ADDR=http://$(ip route get 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p'):30990/
