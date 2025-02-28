#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

image_tag="$1"

if [ -z $image_tag ]; then
    echo "No tag was provided, using latest"
    image_tag="latest"
fi

docker build --no-cache \
    -t gaudi-base:$image_tag \
    --build-arg https_proxy=http://proxy-dmz.intel.com:912 \
    --build-arg http_proxy=http://proxy-dmz.intel.com:912 \
    --build-arg ARTIFACTORY_URL=vault.habana.ai \
    --build-arg VERSION=1.18.0 \
    --build-arg REVISION=524 \
    .
build_status=$?
if [ $build_status -ne 0 ]; then
    exit $build_status
fi

docker tag gaudi-base:$image_tag amr-fmext-registry.caas.intel.com/idc-training/trainings/gaudi-base:$image_tag
docker push amr-fmext-registry.caas.intel.com/idc-training/trainings/gaudi-base:$image_tag
