#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

export URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local
export TOKEN=$(curl "${URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
echo ${TOKEN}