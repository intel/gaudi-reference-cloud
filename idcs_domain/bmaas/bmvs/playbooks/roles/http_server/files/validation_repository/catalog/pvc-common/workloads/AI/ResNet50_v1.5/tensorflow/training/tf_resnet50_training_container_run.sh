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

# script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

CONTAINER_NAME=${CONTAINER_NAME:-tfresnettraining}
IMAGE_NAME=intel/intel-extension-for-tensorflow:${CONTAINER_NAME}
DOCKER_ARGS=${DOCKER_ARGS:--it --rm }

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

# Use below environment to customize
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet_data/imagenet/}
PRECISION=${PRECISION:-bfloat16}
DATASET_DUMMY=${DATASET_DUMMY:-1}
EPOCHS=${EPOCHS:-1}
BATCH_SIZE=${BATCH_SIZE:-256}

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
    TILE=${TILE:-2}
else
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-1}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-1}
    TILE=${TILE:-1}
fi

docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --ipc=host \
  --env DATASET_DIR=${DATASET_DIR} \
  --env PRECISION=${PRECISION} \
  --env DATASET_DUMMY=${DATASET_DUMMY} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS} \
  --env PROCESS_PER_NODE=${PROCESS_PER_NODE} \
  --env EPOCHS=${EPOCHS} \
  --env ZE_AFFINITY_MASK=${ZE_AFFINITY_MASK} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash /workspace/tf_resnet50_training_run.sh


