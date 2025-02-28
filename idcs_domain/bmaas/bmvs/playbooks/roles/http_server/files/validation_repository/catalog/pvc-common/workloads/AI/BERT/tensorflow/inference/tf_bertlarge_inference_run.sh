#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

# Use environment in below env to customize
VENV_NAME=${VENV_NAME}

cd ${WORKSPACE}
if [ "${VENV_NAME}x" != "x" ]; then
    if [ $(which conda) ]; then
        echo "Using conda venv ${VENV_NAME}"
        eval "$(conda shell.bash hook)"
        conda activate ${VENV_NAME}
    elif [  $(which $PYTHON) ]; then
        echo "Using Python venv ${VENV_NAME}"
        source ${VENV_NAME}/bin/activate
    else
        echo "Python3 not found, please install python3 and run setup first"
        exit 1
    fi
fi

export PRECISION=${PRECISION:-fp16}
export BATCH_SIZE=${BATCH_SIZE:-64}
export PRETRAINED_DIR=${PRETRAINED_DIR:-${WORKSPACE}/wwm_uncased_L-24_H-1024_A-16}
export TENSORFLOW_MODEL_PATH=${WORKSPACE}/models
export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
export FROZEN_GRAPH=${FROZEN_GRAPH:-${WORKSPACE}/fp32_bert_squad.pb}
export SQUAD_DIR=${SQUAD_DIR:-${WORKSPACE}/SQuAD1.0}
export NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
export DEVICEID=${DEVICEID:-0}

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

export PYTHONPATH=${TENSORFLOW_MODEL_PATH}

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

export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

test_log=${WORKSPACE}/test.log
cd models
# Run quickstart script:
./quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh 2>&1 |tee ${test_log}

echo "---------Summary--------"
echo "Throughput per process:"
grep "Total throughput (examples/sec):" ${test_log}

total_throughput=$( grep "Total throughput (examples/sec):" ${test_log} | awk -F':' '{sum+=$2;} END{print sum} ' )
echo "Total Throughput: ${total_throughput} sentences/s"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Tile: ${Tile}"
echo "Number of GPU: ${NUMBER_OF_GPU}"
