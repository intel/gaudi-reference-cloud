#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

export no_proxy=${no_proxy},.kind.local

#### Access token

# URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local
# export ACCESS_TOKEN=$(curl "${URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
# echo ${ACCESS_TOKEN}

# # Check if the token is obtained successfully
# # if [ -z "$ACCESS_TOKEN" ]; then
# #   echo "Failed to obtain access token."
# #   exit 1
# # fi

# # personal cloud account id
# CLOUDACCOUNT_ID="033395876667"
# #HOST_URL="https://dev.api.cloud.intel.com.kind.local"

# HOST_URL="https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"


forwardPort() {
    svc=$1
    lport=$2
    rport=$3

    kubectl port-forward -n idcs-system svc/$svc $lport:$rport >/dev/null 2>&1 &
    forwardPids="$forwardPids $!"

    try=0
    while ! nc -vz localhost $lport > /dev/null 2>&1 ; do
        sleep 0.1
        try=$((try + 1))
        if [ $try -gt 20 ]; then
            fatal "timed out waiting for port-forward %svc"
        fi
    done
}
