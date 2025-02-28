#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
#
# This script adds the required label to all bmh resources on the cluster.

set -e

# Add the label to the resource if not exists
add_label() {
    local RESOURCE_TYPE=$1
    local NAMESPACE=$2
    local NAME=$3
    local LABEL_KEY=$4
    local LABEL_VALUE=$5

    if kubectl get "$RESOURCE_TYPE" "$NAME" -n "$NAMESPACE" -o jsonpath="{.metadata.labels}" | grep -q "$LABEL_KEY"; then
        echo "$RESOURCE_TYPE/$NAME already has the label $LABEL_KEY"
    else
        kubectl label --overwrite "$RESOURCE_TYPE" "$NAME" -n "$NAMESPACE" "$LABEL_KEY=$LABEL_VALUE"
    fi
}

# Label all BMHs with network mode
kubectl get baremetalhost.metal3.io --all-namespaces -o json | jq -r '.items[] | "\(.metadata.namespace) \(.metadata.name)"' | while read -r NAMESPACE NAME; do
    add_label baremetalhost.metal3.io "$NAMESPACE" "$NAME" "cloud.intel.com/network-mode" ""
done
