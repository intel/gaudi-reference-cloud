#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

node_name="$1"

if [[ $* == *--control-plane* ]]; then
    kubectl label nodes $node_name training.cloud.intel.com/role=control-plane
    kubectl label nodes $node_name hub.jupyter.org/node-purpose=core
    kubectl label nodes $node_name node-role.kubernetes.io/control-plane=control-plane
elif [[ $* == *--data-plane* ]]; then
    kubectl label nodes $node_name training.cloud.intel.com/role=data-plane
    kubectl label nodes $node_name hub.jupyter.org/node-purpose=user
    kubectl label nodes $node_name node-role.kubernetes.io/data-plane=data-plane
else
    echo "Must provide either --control-plane or --data-plane"
    return 1
fi
