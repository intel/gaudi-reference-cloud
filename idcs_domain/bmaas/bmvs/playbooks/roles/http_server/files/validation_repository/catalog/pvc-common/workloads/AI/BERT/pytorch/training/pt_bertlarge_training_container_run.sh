#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

IMAGE_NAME=intel/intel-extension-for-pytorch:ptbertlargetraining
DOCKER_ARGS=${DOCKER_ARGS:---rm -it}

OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
DATASET_DIR=${DATASET_DIR:-/dataset/mlcommons_bert/}
PROCESSED_DATASET_DIR=${PROCESSED_DATASET_DIR:-/dataset/mlcommons_bert_processed/}
PRECISION=${PRECISION:-bf16}
DEVICEID=${DEVICEID:-0}
BATCH_SIZE=${BATCH_SIZE:-16}
DISTRIBUTED_TRAINING=${DISTRIBUTED_TRAINING:-1}
mkdir -p ${OUTPUT_DIR}
#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    export Tile=${Tile:-2}
    export NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
    export PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
else
    export Tile=${Tile:-1}
    export NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-1}
    export PROCESS_PER_NODE=${PROCESS_PER_NODE:-1}
fi

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

# if set ZE_AFFINITY_MASK, add to docker env
if [ "${ZE_AFFINITY_MASK}" != "" ]; then
    DOCKER_ARGS="${DOCKER_ARGS} --env ZE_AFFINITY_MASK=${ZE_AFFINITY_MASK}"
fi

test_log=${WORKSPACE}/test.log
docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --privileged \
  --ipc host \
  --env DATASET_DIR=${DATASET_DIR} \
  --env PROCESSED_DATASET_DIR=${PROCESSED_DATASET_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env Tile=${Tile} \
  --env PRECISION=${PRECISION} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS} \
  --env PROCESS_PER_NODE=${PROCESS_PER_NODE} \
  --env DISTRIBUTED_TRAINING=${DISTRIBUTED_TRAINING} \
  --env DEVICEID=${DEVICEID} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume ${PROCESSED_DATASET_DIR}:${PROCESSED_DATASET_DIR} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash /workspace/pt_bertlarge_training_run.sh
