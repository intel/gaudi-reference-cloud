#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

# Use environment in below env to customize
VENV_NAME=${VENV_NAME}
PRECISION=${PRECISION:-bfloat16}
EPOCHS=${EPOCHS:-1}
NUM_TRAIN_STEPS=${NUM_TRAIN_STEPS:-720}
BATCH_SIZE=${BATCH_SIZE:-32}
BERT_LARGE_DIR=${BERT_LARGE_DIR:-${WORKSPACE}/wwm_uncased_L-24_H-1024_A-16}

# tf training runtime preparation
TENSORFLOW_MODEL_PATH=${WORKSPACE}/models
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}

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
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-2}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-2}
    TILE=${TILE:-2}
    BATCH_SIZE=${BATCH_SIZE:-32}
else
    NUMBER_OF_PROCESS=${NUMBER_OF_PROCESS:-1}
    PROCESS_PER_NODE=${PROCESS_PER_NODE:-1}
    TILE=${TILE:-1}
    BATCH_SIZE=${BATCH_SIZE:-16}
fi

# Create Dummy dataset from bert large dataset
export DUMMY_DATA=${OUTPUT_DIR}/tf-examples-512.tfrecord
echo 'DUMMY_DATA='$DUMMY_DATA

echo "----------- Start: DUMMY DATA generation --------------"
# Assumption is BERT_LARGE_DIR, DUMMY_DATA and MODEL_DIR is set.
SOURCE_DIR=${TENSORFLOW_MODEL_PATH}/models/language_modeling/tensorflow/bert_large/training/fp32

pushd $SOURCE_DIR

python create_pretraining_data.py \
        --input_file=./sample_text.txt \
        --output_file=$DUMMY_DATA \
        --vocab_file=$BERT_LARGE_DIR/vocab.txt \
        --do_lower_case=False \
        --max_seq_length=512 \
        --max_predictions_per_seq=76 \
        --masked_lm_prob=0.15 \
        --random_seed=12345 \
        --dupe_factor=5

popd

echo "----------- End: DUMMY DATA generation --------------"

export TF_NUM_INTEROP_THREADS=1

export BERT_LARGE_DIR=${BERT_LARGE_DIR}
export OUTPUT_DIR=${OUTPUT_DIR}
export Tile=${TILE}
export PRECISION=${PRECISION}

source "${TENSORFLOW_MODEL_PATH}/quickstart/common/utils.sh"
_command mpirun -np $NUMBER_OF_PROCESS -ppn $PROCESS_PER_NODE --prepend-rank \
  python ${TENSORFLOW_MODEL_PATH}/models/language_modeling/tensorflow/bert_large/training/bfloat16/run_pretraining.py \
  --input_file=$DUMMY_DATA \
  --output_dir=${OUTPUT_DIR} \
  --precision=${PRECISION} \
  --do_train=True \
  --do_eval=False \
  --bert_config_file=${BERT_LARGE_DIR}/bert_config.json \
  --train_batch_size=${BATCH_SIZE} \
  --max_seq_length=512 \
  --max_predictions_per_seq=76 \
  --num_train_steps=${NUM_TRAIN_STEPS} \
  --num_warmup_steps=6 \
  --accum_steps=1 \
  --learning_rate=2e-5 \
  --do_lower_case=False \
  --mpi_workers_sync_gradients=False \
  --use_tpu=False \
  --experimental_gelu=True \
  --optimized_softmax=True \
  --inter_op_parallelism_threads=1 \
  --intra_op_parallelism_threads=1 2>&1 | tee test.log

echo "---------Summary--------"
echo "Throughput per process:"
grep "Throughput" test.log

total_throughput=$( grep "Throughput" test.log | awk -F' ' '{sum+=$4;} END{print sum} ' )
echo "Total Throughput: ${total_throughput} sentences/s"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Tile: ${TILE}"
echo "Number of Process: ${NUMBER_OF_PROCESS}"
