#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -e

# Set OVN Central IP
OVN_CENTRAL_IP=localhost
docker stop local-ovs >/dev/null 2>&1 || true
docker rm local-ovs -f >/dev/null 2>&1 || true
# Start the ovs-node container
echo "Starting ovs-node container..."
docker run \
  -d --rm --privileged \
  --name ovs-node \
  -v /lib/modules:/lib/modules:ro \
  -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
  -v /etc/ovn:/etc/ovn/ \
  -e OVN_CENTRAL_IP=$OVN_CENTRAL_IP \
  --network=host \
  --user root \
  local-ovs
