#!/bin/bash

# Script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

CONTAINER_NAME=${CONTAINER_NAME:-ptresnetinference}
IMAGE_NAME=intel/intel-extension-for-pytorch:${CONTAINER_NAME}
DOCKER_ARGS=${DOCKER_ARGS:--it --rm }

# Use below environment to customize
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet/}
PRECISION=${PRECISION:-bf16}
DATASET_DUMMY=${DATASET_DUMMY:-1}
NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DEVICEID=${DEVICEID:-0}
EPOCHS=${EPOCHS:-1}
BATCH_SIZE=${BATCH_SIZE:-256}
ITEM_NUM=${ITEM_NUM:-100}
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}

if [[ ! -d "${DATASET_DIR}" ]]; then
  echo "The DATASET_DIR '${DATASET_DIR}' does not exist"
  #exit 1
fi

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
	export Tile=${Tile:-2}
else
	export Tile=${Tile:-1}
fi
echo "Detected ${Tile} stacks in a GPU"

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
  --env Tile=${Tile} \
  --env ITEM_NUM=${ITEM_NUM} \
  --env DATASET_DIR=${DATASET_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env PRECISION=${PRECISION} \
  --env DATASET_DUMMY=${DATASET_DUMMY} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env NUMBER_OF_GPU=${NUMBER_OF_GPU} \
  --env DEVICEID=${DEVICEID} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash pt_resnet50_inference_run.sh

