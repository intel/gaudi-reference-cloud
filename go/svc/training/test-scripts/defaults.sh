#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

REGION=${REGION:-us-dev-1}
AZONE=${AZONE:-${REGION}a}
CLOUDACCOUNT=${CLOUDACCOUNT:-090631835287}
PREFIXLENGTH=${PREFIXLENGTH:-24}
VNETNAME=${VNETNAME:-${AZONE}-default}
KEYNAME=${KEYNAME:-user2@acme.com}
NAME=${NAME:-my-instance-1}
SSHPUBLICKEY=$(cat ~/.ssh/id_rsa.pub)
INSTANCE_TYPE=${INSTANCE_TYPE:-vm-spr-sml}
MACHINE_IMAGE=${MACHINE_IMAGE:-ubuntu-2204-jammy-v20230122}
CURL_OPTS=${CURL_OPTS:--vk}

if [ "${IDC_REGIONAL_URL_PREFIX}" == "" ]; then
    IDC_REGIONAL_FQDN=dev.compute.us-dev-1.api.cloud.intel.com.kind.local
    IDC_REGIONAL_PORT_FILE=${SCRIPT_DIR}/../../../../local/${REGION}_host_port_443
    IDC_REGIONAL_PORT=$(cat "${IDC_REGIONAL_PORT_FILE}")
    IDC_REGIONAL_URL_PREFIX=https://${IDC_REGIONAL_FQDN}:${IDC_REGIONAL_PORT}
fi

