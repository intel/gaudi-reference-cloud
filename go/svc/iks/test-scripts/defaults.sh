#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
CLOUDACCOUNTNAME=${CLOUDACCOUNTNAME:-user1@acme.com}
CURL_OPTS=${CURL_OPTS:--vk}

if [ "${IDC_GLOBAL_URL_PREFIX}" == "" ]; then
    IDC_GLOBAL_FQDN=dev.api.cloud.intel.com.kind.local
    IDC_GLOBAL_PORT_FILE=${SCRIPT_DIR}/../../../../local/global_host_port_443
    IDC_GLOBAL_PORT=$(cat "${IDC_GLOBAL_PORT_FILE}")
    IDC_GLOBAL_URL_PREFIX=https://${IDC_GLOBAL_FQDN}:${IDC_GLOBAL_PORT}
fi
