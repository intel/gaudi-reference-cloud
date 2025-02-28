#!/bin/bash

# script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
export DATASET_DIR=${DATASET_DIR:-/dataset/imagenet_data_2012/tf_records}
export DATASET_LOCATION=${DATASET_DIR}
export PRECISION=${PRECISION:-int8}
export DEVICEID=${DEVICEID:-0}
export WARMUP_STEPS=${WARMUP_STEPS:-5}
export STEPS=${STEPS:-25}
export BATCH_SIZE=${BATCH_SIZE:-1024}
export NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DATASET_DUMMY=${DATASET_DUMMY:-1}
# The benchmark script will check if dataset in DATASET_DIR exist. If exist, benchmark with real date otherwise with dummy data
if [[ $DATASET_DUMMY == "1" ]]; then
	mkdir -p ${WORKSPACE}/__empty_dataset_dir
	export DATASET_LOCATION="${WORKSPACE}/__empty_dataset_dir"
fi
#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
	export Tile=${Tile:-2}
else
	export Tile=${Tile:-1}
fi

rm -f ${OUTPUT_DIR}/*.log
#if [ ! -z "$(ls -A $OUTPUT_DIR)" ]; then
#        echo "The folder \"$OUTPUT_DIR\" is not empty"
#        echo "Please clean the folder \"$OUTPUT_DIR\" and restart"
#        exit 1
#fi

cd ${WORKSPACE}
VENV_NAME=${VENV_NAME}
PYTHON=${PYTHON:-python3}
if [ "${VENV_NAME}x" != "x" ]; then
    if [ $(which conda) ]; then
        echo "Using conda venv ${VENV_NAME}"
        eval "$(conda shell.bash hook)"
        conda activate ${VENV_NAME}
    elif [  $(which ${PYTHON}) ]; then
        echo "Using Python venv ${VENV_NAME}"
        source ${VENV_NAME}/bin/activate
    else
        echo "Python3 not found, please install python3 and run setup first"
        exit 1
    fi
fi

PRETRAINED_MODEL_DIR=tf_resnet_models
if [ "${PRECISION}" == "int8" ]; then
	export FROZEN_GRAPH=${WORKSPACE}/${PRETRAINED_MODEL_DIR}/resnet50v1_5_int8_pretrained_model.pb
else
    export FROZEN_GRAPH=${WORKSPACE}/${PRETRAINED_MODEL_DIR}/resnet50_v1.pb
fi

ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
echo "setup oneapi environments"
if test -f ${ONEAPI_ROOT}/setvars.sh; then
    #source ${ONEAPI_ROOT}/setvars.sh --force
    source ${ONEAPI_ROOT}/compiler/latest/env/vars.sh
    source ${ONEAPI_ROOT}/mkl/latest/env/vars.sh
    source ${ONEAPI_ROOT}/ccl/latest/env/vars.sh    
else
    #setup oneAPI runtime
    export LD_LIBRARY_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib:$LD_LIBRARY_PATH
    export PATH=/opt/intel/oneapi/redist/bin:$PATH
    export I_MPI_ROOT=/opt/intel/oneapi/redist/lib
    export CCL_ROOT=/opt/intel/oneapi/redist
    export FI_PROVIDER_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib/prov
fi

cd ${WORKSPACE}/intelai-models
export GPU_TYPE=max_series
SCRIPT=batch_inference.sh
eval ./quickstart/image_recognition/tensorflow/resnet50v1_5/inference/gpu/${SCRIPT}

echo "---------- Summary ----------"

echo "Throughput per process:"
grep Throughput ${OUTPUT_DIR}/resnet*_raw.log

total_throughput=$( grep Throughput ${OUTPUT_DIR}/resnet*_raw.log | awk -F' ' '{sum+=$2;} END{print sum} ' )
echo "Total Throughput: ${total_throughput} images/sec"
echo "Precision: $PRECISION"
echo "Batch Size: ${BATCH_SIZE}"
echo "Number of GPU: ${NUMBER_OF_GPU}"
echo "Tile: ${Tile}"

