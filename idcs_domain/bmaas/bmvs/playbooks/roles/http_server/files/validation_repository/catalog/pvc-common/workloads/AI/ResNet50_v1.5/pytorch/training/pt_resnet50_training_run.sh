#!/bin/bash

# Script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

VENV_NAME=${VENV_NAME}
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
MASTER_ADDR=${MASTER_ADDR:-127.0.0.1}
ITEM_NUM=${ITEM_NUM:-313}

if [ "${VENV_NAME}x" != "x" ]; then
    if [ $(which conda) ]; then
        echo "Using conda venv ${VENV_NAME}"
        eval "$(conda shell.bash hook)"
        conda activate ${VENV_NAME}
    elif [  $(which python3) ]; then
        echo "Using Python venv ${VENV_NAME}"
        source ${VENV_NAME}/bin/activate
    else
        echo "Python3 not found, please install python3 and run setup first"
        exit 1
    fi
fi

# Clean
if [ -d $OUTPUT_DIR ]; then
	rm $OUTPUT_DIR/*.log
fi

if [ -d tensorboard_log ]; then
	rm -fr tensorboard_log/
fi

# Use below environment to customize
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet/}
PRECISION=${PRECISION:-bfloat16}
DATASET_DUMMY=${DATASET_DUMMY:-1}
NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
EPOCHS=${EPOCHS:-1}
BATCH_SIZE=${BATCH_SIZE:-256}

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

# extra option: precsion and dataset
MAIN_ARGS=""
if [ "${PRECISION}" == "bfloat16" ] || [ "${PRECISION}" == "bf16" ]; then
	echo "Use bf16 for training"
	MAIN_ARGS=${MAIN_ARGS}" --bf16 1"
elif [ "${PRECISION}" == "fp16"  ]; then
	echo "Use fp16 for training"
	MAIN_ARGS=${MAIN_ARGS}" --fp16 1"
elif [ "${PRECISION}" == "tf32" ]; then
	echo "Use tf32 for training"
	MAIN_ARGS=${MAIN_ARGS}" --tf32 1"
else
	MAIN_ARGS=${MAIN_ARGS}" --fp16 1"
fi

if [ "${DATASET_DUMMY}" == "1" ]; then
    echo "Use dummy dataset for training"
    MAIN_ARGS="${MAIN_ARGS} --dummy"
else
    echo "Use ImageNet data in ${DATASET_DIR}"
    MAIN_ARGS="${MAIN_ARGS}  ${DATASET_DIR}"
fi

MAIN_ARGS="${MAIN_ARGS}  --num-iterations ${ITEM_NUM}"

if [ "${NUMBER_OF_PROCESS}" == "16" ]; then
    echo "Add number of data loading workers from default 4 to 8 for x8 Max1550"
    MAIN_ARGS="${MAIN_ARGS} -j 8"
fi

# Create the output directory, if it doesn't already exist
mkdir -p $OUTPUT_DIR

# Set ENV for mpi
if [ "${FI_PROVIDER_PATH}" == "" ]; then
source ${ONEAPI_ROOT}/setvars.sh
fi

# ZE_FLAT_DEVICE_HIERARCHY
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE 

echo "ResNet50 training running on ${NUMBER_OF_PROCESS} stacks, batch size ${BATCH_SIZE}, ${EPOCHS} epochs."
echo "Extra options: ${MAIN_ARGS}"

#python3 main.py -a resnet50 \
#    -b ${BATCH_SIZE} \
#    --xpu 0 \
#    --bucket-cap 200 \
#    --broadcast-buffers False \
#    --epochs ${EPOCHS} \
#    --num-iterations 20

# multiple nodes
if [ "${HOSTFILE}" != "" ]; then
mpirun -np ${NUMBER_OF_PROCESS} --hostfile ${HOSTFILE} -ppn ${PROCESS_PER_NODE} -prepend-rank \
	-genv I_MPI_OFFLOAD_RDMA=1  \
	python3 main.py -a resnet50 \
    -b ${BATCH_SIZE} \
    --xpu 0 \
    --bucket-cap 200 \
    --broadcast-buffers False \
    --epochs ${EPOCHS} \
    --dist-url ${MASTER_ADDR} \
    ${MAIN_ARGS} 2>&1 | tee ${OUTPUT_DIR}/$(hostname)-pt_resnet50_training.log
else
# single node
# use shm for single node by default
export FI_PROVIDER=shm
mpiexec -np ${NUMBER_OF_PROCESS} -ppn ${PROCESS_PER_NODE} -prepend-rank  python3 main.py -a resnet50 \
    -b ${BATCH_SIZE} \
    --xpu 0 \
    --bucket-cap 200 \
    --broadcast-buffers False \
    --epochs ${EPOCHS} \
    --dist-url ${MASTER_ADDR} \
    ${MAIN_ARGS} 2>&1 | tee ${OUTPUT_DIR}/$(hostname)-pt_resnet50_training.log
fi

echo "---------Summary--------"
echo "Throughput per process of ${NUMBER_OF_PROCESS} stacks, ${EPOCHS} epochs:"
grep "Training performance" ${OUTPUT_DIR}/$(hostname)-pt_resnet50_training.log

total_throughput=$( grep "Training performance" ${OUTPUT_DIR}/$(hostname)-pt_resnet50_training.log | \
tail -n ${NUMBER_OF_PROCESS} | awk -F':' '{ print $4}' | awk -F' ' '{ sum+=$1; } END {print sum} ' )
echo "Total Throughput: ${total_throughput} imgs/sec"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Number of PROCESS: ${NUMBER_OF_PROCESS}"

# Clean
rm -f checkpoint.pth.tar
rm -f model_best.pth.tar
