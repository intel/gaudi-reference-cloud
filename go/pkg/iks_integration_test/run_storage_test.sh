#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# export KUBECONFIG
export KUBECONFIG=$1

# clone the storage test repo to tmp folder
git clone https://github.com/intel-sandbox/applications.infrastructure.caas.iks-storage-testing.git /tmp/iks_storage_test

# change directory to storage test
pushd /tmp/iks_storage_test

# run storage test script
./storage-test.sh

popd

sudo rm -r /tmp/iks_storage_test
