!/bin/bash

# script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

IMAGE_NAME=intel/intel-extension-for-tensorflow:tfresnetinference
DOCKER_ARGS="--rm -it"

OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet_data_2012/tf_records}
PRECISION=${PRECISION:-int8}
DEVICEID=${DEVICEID:-0}
WARMUP_STEPS=${WARMUP_STEPS:-5}
STEPS=${STEPS:-25}
BATCH_SIZE=${BATCH_SIZE:-1024}
NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DATASET_DUMMY=${DATASET_DUMMY:-1}
# The benchmark script will check if dataset in DATASET_DIR exist. If exist, benchmark with real date otherwise with dummy data
if [[ $DATASET_DUMMY == "1" ]]; then
	mkdir -p ${WORKSPACE}/__empty_dataset_dir
    export DATASET_LOCATION="${WORKSPACE}/__empty_dataset_dir"
fi

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
	Tile=${Tile:-2}
else
	Tile=${Tile:-1}
fi

#rm -f ${OUTPUT_DIR}/*.log
#if [ ! -z "$(ls -A $OUTPUT_DIR)" ]; then
#	echo "The folder \"$OUTPUT_DIR\" is not empty"
#	echo "Please clean the folder \"$OUTPUT_DIR\" and restart"
#	exit 1
#fi

GPU_TYPE=max_series

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --ipc=host \
  --privileged \
  --env DEVICEID=${DEVICEID} \
  --env PRECISION=${PRECISION} \
  --env GPU_TYPE=${GPU_TYPE} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env DATASET_LOCATION=${DATASET_DIR} \
  --env Tile=${Tile} \
  --env NUMBER_OF_GPU=${NUMBER_OF_GPU} \
  --env WARMUP_STEPS=${WARMUP_STEPS} \
  --env STEPS=${STEPS} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash /workspace/tf_resnet50_inference_run.sh

