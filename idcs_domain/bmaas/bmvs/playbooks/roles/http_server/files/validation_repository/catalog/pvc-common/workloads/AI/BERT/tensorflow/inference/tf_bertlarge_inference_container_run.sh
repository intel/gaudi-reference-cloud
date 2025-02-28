#!/bin/bash

# script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

CONTAINER_NAME=${CONTAINER_NAME:-tfbertlargeinference}
IMAGE_NAME=intel/intel-extension-for-tensorflow:${CONTAINER_NAME}
DOCKER_ARGS=${DOCKER_ARGS:--it --rm }

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

# Use below environment to customize
export PRECISION=${PRECISION:-fp16}
export BATCH_SIZE=${BATCH_SIZE:-64}
export PRETRAINED_DIR=${PRETRAINED_DIR:-${WORKSPACE}/wwm_uncased_L-24_H-1024_A-16}
export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
export FROZEN_GRAPH=${FROZEN_GRAPH:-${WORKSPACE}/fp32_bert_squad.pb}
export SQUAD_DIR=${SQUAD_DIR:-${WORKSPACE}/SQuAD1.0}
export WARMUP_STEPS=${WARMUP_STEPS:-5}
export STEPS=${STEPS:-25}
export NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
export DEVICEID=${DEVICEID:-0}

echo "clean up the folder $OUTPUT_DIR"
if [ ! -d "$OUTPUT_DIR" ]; then
    mkdir -p $OUTPUT_DIR
else
    rm -rf $OUTPUT_DIR
    mkdir -p $OUTPUT_DIR
fi

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    export Tile=${Tile:-2}
else
    export Tile=${Tile:-1}
fi

echo "Extra docker options: ${DOCKER_ARGS}"

test_log=${WORKSPACE}/test.log
docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --privileged \
  --ipc=host \
  --env PRECISION=${PRECISION} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env PRETRAINED_DIR=${PRETRAINED_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env FROZEN_GRAPH=${FROZEN_GRAPH} \
  --env SQUAD_DIR=${SQUAD_DIR} \
  --env WARMUP_STEPS=${WARMUP_STEPS} \
  --env STEPS=${STEPS} \
  --env NUMBER_OF_GPU=${NUMBER_OF_GPU} \
  --env DEVICEID=${DEVICEID:-0} \
  --env Tile=${Tile} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${PRETRAINED_DIR}:${PRETRAINED_DIR} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume ${FROZEN_GRAPH}:${FROZEN_GRAPH} \
  --volume ${SQUAD_DIR}:${SQUAD_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  ${IMAGE_NAME} \
  /bin/bash /workspace/tf_bertlarge_inference_run.sh

