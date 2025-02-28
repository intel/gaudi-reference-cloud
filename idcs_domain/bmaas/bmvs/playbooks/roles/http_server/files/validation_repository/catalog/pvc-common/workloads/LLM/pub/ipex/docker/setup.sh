#!/bin/bash
# Copyright (c) 2023 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

if [ ! -d "intel-extension-for-pytorch" ]; then
        git clone https://github.com/intel/intel-extension-for-pytorch.git
        cd intel-extension-for-pytorch
        git checkout v2.1.10+xpu
        git submodule sync
        git submodule update --init --recursive
        cd ..
fi

cp intel-extension-for-pytorch/examples/gpu/inference/python/llm/* llm_inference_test/

DOCKER_IMG_NAME=ipex-llm:2.1.10
cd intel-extension-for-pytorch
DOCKER_BUILDKIT=1 docker build -f examples/gpu/inference/python/llm/Dockerfile \
        --build-arg GID_RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,') \
        $(env | grep -E '(_proxy=|_PROXY)' | sed 's/^/--build-arg /') \
        -t ${DOCKER_IMG_NAME} \
        .
