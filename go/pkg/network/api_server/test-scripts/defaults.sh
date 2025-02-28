#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

REGION=${REGION:-us-dev-1}
CLOUDACCOUNT=${CLOUDACCOUNT:-090631835287}
CURL_OPTS=${CURL_OPTS:--vk}
CREATEADMIN=${CREATEADMIN:-idcadmin@intel.com}

if [ "${IDC_REGIONAL_URL_PREFIX}" == "" ]; then
    IDC_REGIONAL_FQDN=dev.compute.${REGION}.api.cloud.intel.com.kind.local
    IDC_REGIONAL_PORT_FILE=${SCRIPT_DIR}/../../../../../local/${REGION}_host_port_443
    IDC_REGIONAL_PORT=$(cat "${IDC_REGIONAL_PORT_FILE}")
    IDC_REGIONAL_URL_PREFIX=https://${IDC_REGIONAL_FQDN}:${IDC_REGIONAL_PORT}
fi
