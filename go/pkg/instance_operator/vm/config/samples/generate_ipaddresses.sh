#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Deprecated. See deployment/charts/compute-resources/templates/private.cloud_v1alpha1_ipaddress.yaml.
# Used to deploy in kind and development environments.

for octet2 in {16..16}; do
  for octet3 in {0..3}; do
    for octet4 in {5..254}; do
      echo "apiVersion: private.cloud.intel.com/v1alpha1
kind: IpAddress
metadata:
  name: 172.${octet2}.${octet3}.${octet4}
spec:
  subnet: 172.16.0.0/16
---"
    done
  done
done
