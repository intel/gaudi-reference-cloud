#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Deprecated.

set -ex

script_dir=$(cd "$(dirname "$0")" && pwd)

API_SERVER=${API_SERVER:-'http://localhost:8001'}
INTEL_CLOUD_API='private.cloud.intel.com/v1alpha1'
NAMESPACE='my-project-123456'
HEADER='Content-type: application/json'

command -v yq  >/dev/null 2>&1 || (echo '"yq" not found'; exit 1)

create() {
  curl -H "${HEADER}" -X POST "$1" -d "$(yq "$2" -o json)"
}

get() {
  curl -H "${HEADER}" -X GET "$1"
}

list () {
  get "$1"
}

create "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instancetypes" "${script_dir}/_v1alpha1_instancetype_m6i.metal.yaml"
create "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instancetypes" "${script_dir}/_v1alpha1_instancetype_dev.vm.yaml"
create "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instances" "${script_dir}/_v1alpha1_instance_my-virtual-machine-1.yaml"
create "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instances" "${script_dir}/_v1alpha1_instance_my-bare-metal-1.yaml"

get "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instancetypes/m6i.metal"
get "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instancetypes/dev.vm"
get "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instances/my-virtual-machine-1"
get "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instances/my-bare-metal-1"

list "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instancetypes"
list "${API_SERVER}/apis/${INTEL_CLOUD_API}/namespaces/${NAMESPACE}/instances"
