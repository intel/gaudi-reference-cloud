#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Generate a Java Web Token (JWT) that can impersonate any email and groups.
# The output will be in a format can be copy/pasted to another terminal.
# This script can also be executed with "source" to set the TOKEN in the current environment.
# This runs in a function and uses local variables to not pollute the caller's environment.

get_token() {
    # Optional input parameters
    local OIDC_EMAIL=${OIDC_EMAIL:-admin@intel.com}
    local OIDC_GROUPS=${OIDC_GROUPS:-IDC.Admin}
    local OIDC_CURL_OPTS="${OIDC_CURL_OPTS:---silent -k}"

    local IDC_OIDC_FQDN=dev.oidc.cloud.intel.com.kind.local
    local IDC_OIDC_PORT=${IDC_OIDC_PORT:-80}
    local IDC_OIDC_URL_PREFIX=http://${IDC_OIDC_FQDN}:${IDC_OIDC_PORT}

    # Calculate variables
    local OIDC_PARAMS="email=${OIDC_EMAIL}&groups=${OIDC_GROUPS}"
    local OIDC_GET_TOKEN_URL="${IDC_OIDC_URL_PREFIX}/token?${OIDC_PARAMS}"
    echo OIDC_GET_TOKEN_URL: "${OIDC_GET_TOKEN_URL}"

    # Get token
    export TOKEN=$(curl ${OIDC_CURL_OPTS} "${OIDC_GET_TOKEN_URL}")

    # Print token
    echo export TOKEN=\"${TOKEN}\"
}

get_token
