#!/bin/bash
# Copyright (c) 2023 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

# Use environment in below env to customize
VENV_NAME=${VENV_NAME}
DATASET_DIR=${DATASET_DIR:-/data0/imagenet_raw_data/tf_records}
PRECISION=${PRECISION:-bfloat16}
EPOCHS=${EPOCHS:-1}
BATCH_SIZE=${BATCH_SIZE:-256}
DATASET_DUMMY=${DATASET_DUMMY:-1}

# tf training runtime preparation
TENSORFLOW_MODEL_PATH=${WORKSPACE}/tensorflow-models
MODEL_OUTPUT_DIR=${MODEL_OUTPUT_DIR:-${WORKSPACE}/output}
if [ "${DATASET_DUMMY}" == "1" ]; then
	echo "Use dummy dataset for training"
    if [ "$PRECISION" == "bfloat16" ]; then
	    CONFIG_FILE=${CONFIG_FILE:-${WORKSPACE}/hvd_configure/itex_dummy_bf16_lars.yaml}
    elif [ "$PRECISION" == "float32" ]; then
        CONFIG_FILE=${CONFIG_FILE:-${WORKSPACE}/hvd_configure/itex_dummy_fp32_lars.yaml}
    else
        echo "PRECISION = ${PRECISION} not supported, use bfloat16 or float32"
        exit
    fi
else
	echo "Use real imagenet dataset for training"
    if [ "$PRECISION" == "bfloat16" ]; then
	    CONFIG_FILE=${CONFIG_FILE:-${WORKSPACE}/hvd_configure/itex_bf16_lars.yaml}
    elif [ "$PRECISION" == "float32" ]; then
        CONFIG_FILE=${CONFIG_FILE:-${WORKSPACE}/hvd_configure/itex_fp32_lars.yaml}
    else
        echo "PRECISION = ${PRECISION} not supported, use bfloat16 or float32"
        exit
    fi
fi

CONFIG_FILE_RUN=${WORKSPACE}/itex_run.yaml
eval "sed 's/ epochs:.*/ epochs: ${EPOCHS}/' < ${CONFIG_FILE} > ${CONFIG_FILE_RUN}"
eval "sed -i 's/ batch_size: .*/ batch_size: ${BATCH_SIZE}/' ${CONFIG_FILE_RUN}"

ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
echo "setup oneapi environments"
if test -f ${ONEAPI_ROOT}/setvars.sh; then
    #source ${ONEAPI_ROOT}/setvars.sh --force
    source ${ONEAPI_ROOT}/compiler/latest/env/vars.sh
    source ${ONEAPI_ROOT}/mkl/latest/env/vars.sh
    source ${ONEAPI_ROOT}/ccl/latest/env/vars.sh
    export FI_PROVIDER=shm
else
    #setup oneAPI runtime
    export LD_LIBRARY_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib:$LD_LIBRARY_PATH
    export PATH=/opt/intel/oneapi/redist/bin:$PATH
    export I_MPI_ROOT=/opt/intel/oneapi/redist/lib
    export CCL_ROOT=/opt/intel/oneapi/redist
    export FI_PROVIDER_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib/prov
fi

export PYTHONPATH=${TENSORFLOW_MODEL_PATH}

cd ${WORKSPACE}
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

echo "clean up the folder $MODEL_OUTPUT_DIR"
if [ ! -d "$MODEL_OUTPUT_DIR" ]; then
    mkdir -p $MODEL_OUTPUT_DIR
else
    rm -rf $MODEL_OUTPUT_DIR
    mkdir -p $MODEL_OUTPUT_DIR
fi

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
    TILE=${TILE:-2}
else
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-1}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-1}
    TILE=${TILE:-1}
fi

mpirun -np $NUMBER_OF_PROCESS -ppn $PROCESS_PER_NODE --prepend-rank \
    python ${PYTHONPATH}/official/vision/image_classification/classifier_trainer.py \
    --mode=train_and_eval \
    --model_type=resnet \
    --dataset=imagenet \
    --model_dir=$MODEL_OUTPUT_DIR \
    --data_dir=$DATASET_DIR \
    --config_file=$CONFIG_FILE_RUN 2>&1 |tee test.log

#report performance summary
echo "==========Summary=========="
echo "Throughput for each rank (using data from steps between 400 and 600):"
grep "steps 400 and 600" test.log
total_throughput=$(grep "steps 400 and 600" test.log |awk -F, '{print $2}'|awk -F ' ' '{sum+=$1}END{print sum}')
echo "Total Throughput= ${total_throughput}"
echo "NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS} PROCESS_PER_NODE=${PROCESS_PER_NODE}"
echo "PRECISION=${PRECISION} BATCH_SIZE=${BATCH_SIZE} EPOCHS=${EPOCHS}"
echo "DATASET_DUMMY=${DATASET_DUMMY} DATASET_DIR=${DATASET_DIR}"

