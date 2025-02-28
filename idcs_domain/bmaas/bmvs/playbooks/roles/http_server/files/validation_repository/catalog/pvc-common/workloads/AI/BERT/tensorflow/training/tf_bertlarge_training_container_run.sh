#!/bin/bash

# script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

CONTAINER_NAME=${CONTAINER_NAME:-tfbertlargetraining}
IMAGE_NAME=intel/intel-extension-for-tensorflow:${CONTAINER_NAME}
DOCKER_ARGS=${DOCKER_ARGS:--it --rm }

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

# Use below environment to customize
TF_HVD_ENABLED=${TF_HVD_ENABLED:-1}
PRECISION=${PRECISION:-bfloat16}
EPOCHS=${EPOCHS:-1}
NUM_TRAIN_STEPS=${NUM_TRAIN_STEPS:-720}
BATCH_SIZE=${BATCH_SIZE:-32}
DATASET_DUMMY=${DATASET_DUMMY:-1}
BERT_LARGE_DIR=${BERT_LARGE_DIR:-${WORKSPACE}/wwm_uncased_L-24_H-1024_A-16}
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
    TILE=${TILE:-2}
    BATCH_SIZE=${BATCH_SIZE:-32}
else
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-1}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-1}
    TILE=${TILE:-1}
    BATCH_SIZE=${BATCH_SIZE:-16}
fi

# if set ZE_AFFINITY_MASK, add to docker env
if [ "${ZE_AFFINITY_MASK}" != "" ]; then
    DOCKER_ARGS="${DOCKER_ARGS} --env ZE_AFFINITY_MASK=${ZE_AFFINITY_MASK}"
fi

echo "Extra docker options: ${DOCKER_ARGS}"

docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --privileged \
  --ipc=host \
  --env BERT_LARGE_DIR=${BERT_LARGE_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env Tile=${TILE} \
  --env PRECISION=${PRECISION} \
  --env DATASET_DUMMY=${DATASET_DUMMY} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS} \
  --env PROCESS_PER_NODE=${PROCESS_PER_NODE} \
  --env EPOCHS=${EPOCHS} \
  --env NUM_TRAIN_STEPS=${NUM_TRAIN_STEPS} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume ${BERT_LARGE_DIR}:${BERT_LARGE_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  ${IMAGE_NAME} \
  /bin/bash /workspace/tf_bertlarge_training_run.sh
