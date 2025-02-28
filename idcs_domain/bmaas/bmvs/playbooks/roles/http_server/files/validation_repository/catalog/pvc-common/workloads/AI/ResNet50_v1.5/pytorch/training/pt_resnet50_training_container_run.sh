#!/bin/bash

# Script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

# Use below environment to customize
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet/}
PRECISION=${PRECISION:-bf16}
DATASET_DUMMY=${DATASET_DUMMY:-1}
NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
EPOCHS=${EPOCHS:-1}
BATCH_SIZE=${BATCH_SIZE:-256}
ITEM_NUM=${ITEM_NUM:-313}

if [[ ! -d "${DATASET_DIR}" ]]; then
  echo "The DATASET_DIR '${DATASET_DIR}' does not exist"
  #exit 1
fi

## Log
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}

DOCKER_ARGS="--rm --init -it "
IMAGE_NAME=intel/intel-extension-for-pytorch:ptresnettraining

# if set ZE_AFFINITY_MASK, add to docker env
if [ "${ZE_AFFINITY_MASK}" != "" ]; then
	DOCKER_ARGS="${DOCKER_ARGS} --env ZE_AFFINITY_MASK=${ZE_AFFINITY_MASK}"
fi

echo "Extra docker options: ${DOCKER_ARGS}"

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --shm-size=10G \
  --privileged \
  --ipc=host \
  --env DATASET_DIR=${DATASET_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env PRECISION=${PRECISION} \
  --env DATASET_DUMMY=${DATASET_DUMMY} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS} \
  --env PROCESS_PER_NODE=${PROCESS_PER_NODE} \
  --env EPOCHS=${EPOCHS} \
  --env ITEM_NUM=${ITEM_NUM} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash /workspace/pt_resnet50_training_run.sh
