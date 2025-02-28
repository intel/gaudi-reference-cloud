#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
cd ${WORKSPACE}
VENV_NAME=${VENV_NAME}
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

ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
echo "setup oneapi environments"
if test -f ${ONEAPI_ROOT}/setvars.sh; then
    #source ${ONEAPI_ROOT}/setvars.sh --force
    source ${ONEAPI_ROOT}/compiler/latest/env/vars.sh
    source ${ONEAPI_ROOT}/mkl/latest/env/vars.sh
    source ${ONEAPI_ROOT}/ccl/latest/env/vars.sh    
    # use shm for single node by default
    export FI_PROVIDER=shm
else
    #setup oneAPI runtime
    export LD_LIBRARY_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib:$LD_LIBRARY_PATH
    export PATH=/opt/intel/oneapi/redist/bin:$PATH
    export I_MPI_ROOT=/opt/intel/oneapi/redist/lib
    export CCL_ROOT=/opt/intel/oneapi/redist
    export FI_PROVIDER_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib/prov
fi

export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
export DATASET_DIR=${DATASET_DIR:-/dataset/mlcommons_bert/}
export PROCESSED_DATASET_DIR=${PROCESSED_DATASET_DIR:-/dataset/mlcommons_bert_processed/}
export PRECISION=${PRECISION:-bf16}
export DEVICEID=${DEVICEID:-0}
export BATCH_SIZE=${BATCH_SIZE:-16}
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

if [[ ("${DISTRIBUTED_TRAINING}" == "0") ]]; then
	SCRIPT=quickstart/language_modeling/pytorch/bert_large/training/gpu/bf16_training_plain_format.sh
else
	SCRIPT=quickstart/language_modeling/pytorch/bert_large/training/gpu/ddp_bf16_training_plain_format.sh
fi

cd ${WORKSPACE}/intelai-models
test_log=${WORKSPACE}/test.log
eval  $SCRIPT 2>&1 |tee ${test_log}
echo "---------Summary--------"
echo "Throughput per process:"
grep "bert_train throughput" ${test_log}

total_throughput=$( grep "bert_train throughput" ${test_log} | awk -F' ' '{sum+=$3;} END{print sum} ' )
echo "Total Throughput: ${total_throughput} sentences/s"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Number of Process: ${NUMBER_OF_PROCESS}"
