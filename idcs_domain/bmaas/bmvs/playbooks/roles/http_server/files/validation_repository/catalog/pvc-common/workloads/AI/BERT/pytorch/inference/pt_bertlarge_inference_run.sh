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
else
    #setup oneAPI runtime
    export LD_LIBRARY_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib:$LD_LIBRARY_PATH
    export PATH=/opt/intel/oneapi/redist/bin:$PATH
    export I_MPI_ROOT=/opt/intel/oneapi/redist/lib
    export CCL_ROOT=/opt/intel/oneapi/redist
    export FI_PROVIDER_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib/prov
fi

export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output/}
export DATASET_DIR=${DATASET_DIR:-${WORKSPACE}/SQuAD1.0/}
export BERT_WEIGHT=${BERT_WEIGHT:-${WORKSPACE}/bert_squad_model/}
export PRECISION=${PRECISION:-fp16}
export BATCH_SIZE=${BATCH_SIZE:-64}
export WARMUP_STEPS=${WARMUP_STEPS:-5}
export STEPS=${STEPS:-25}
export NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
export DEVICEID=${DEVICEID:-0}

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    export Tile=${Tile:-2}
else
    export Tile=${Tile:-1}
fi

cd ${WORKSPACE}/models
# Run the quickstart script
test_log=${WORKSPACE}/test.log
./quickstart/language_modeling/pytorch/bert_large/inference/gpu/fp16_inference_plain_format.sh 2>&1 |tee ${test_log}

echo "---------Summary--------"
echo "Throughput per process:"
grep "bert_inf throughput" ${test_log} | sed -n '1d;p'

total_throughput=$( grep "bert_inf throughput" ${test_log} | sed -n '1d;p' | awk -F' ' '{sum+=$3;} END{print sum} ' )
echo "Total Throughput: ${total_throughput} sentences/s"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Tile: ${Tile}"
echo "Number of GPU: ${NUMBER_OF_GPU}"
