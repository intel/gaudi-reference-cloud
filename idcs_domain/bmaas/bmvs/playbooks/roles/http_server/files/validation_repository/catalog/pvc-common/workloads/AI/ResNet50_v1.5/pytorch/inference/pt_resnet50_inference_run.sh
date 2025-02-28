#!/bin/bash

# Script dir as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

cd ${WORKSPACE}

# Use environment in below env to customize
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

# Use below environment to customize
DATASET_DIR=${DATASET_DIR:-/dataset/imagenet/}
OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output}
PRECISION=${PRECISION:-bfloat16}
DATASET_DUMMY=${DATASET_DUMMY:-1}
NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DEVICEID=${DEVICEID:-0}
BATCH_SIZE=${BATCH_SIZE:-256}
ITEM_NUM=${ITEM_NUM:-100}

# extra option: precsion and dataset
MAIN_ARGS=""
if [ "${PRECISION}" == "bfloat16" ] || [ "${PRECISION}" == "bf16" ]; then
    echo "Use bf16 for inference"
    MAIN_ARGS=${MAIN_ARGS}" --bf16 1"
elif [ "${PRECISION}" == "fp16"  ]; then
    echo "Use fp16 for inference"
    MAIN_ARGS=${MAIN_ARGS}" --fp16 1"
elif [ "${PRECISION}" == "tf32" ]; then
    echo "Use tf32 for inference"
    MAIN_ARGS=${MAIN_ARGS}" --tf32 1"
else
    MAIN_ARGS=${MAIN_ARGS}" --int8 1 --benchmark 1"
fi

if [ "${DATASET_DUMMY}" == "1" ]; then
    echo "Use dummy dataset for training"
    MAIN_ARGS="${MAIN_ARGS} --dummy"
else
    echo "Use ImageNet data in ${DATASET_DIR}"
    MAIN_ARGS="${MAIN_ARGS}  ${DATASET_DIR}"
fi

if [ "${Tile}x" == "x" ]; then
    #Max 1550 device id 0bd5, Max 1100 device id 0dba
    device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
    if [ ! -z "$device_type" ]; then
        export Tile=${Tile:-2}
    else
        export Tile=${Tile:-1}
    fi
    echo "Detected ${Tile} stacks in a GPU"
fi

# Create the output directory, if it doesn't already exist
mkdir -p $OUTPUT_DIR

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

echo "ResNet50 inference running on ${NUMBER_OF_PROCESS} stacks, batch size ${BATCH_SIZE}, ${EPOCHS} epochs."
echo "Extra options: ${MAIN_ARGS}"

export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

# Generate jit trace first
IPEX_XPU_ONEDNN_LAYOUT=1 python3 main.py -a resnet50 \
    -b ${BATCH_SIZE} \
    --xpu 0 --pretrained --evaluate  \
    --num-iterations 10 --jit-trace  \
    ${MAIN_ARGS}

for p in $(seq ${DEVICEID} $(( DEVICEID + NUMBER_OF_GPU - 1 )) ); do
    if [[ ${Tile} == "1" ]]; then
        dev_id=`expr $p + 1`
        dev=${dev_id}p
        mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
        echo $mac
        node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
        echo $node

        ZE_AFFINITY_MASK=${p} IPEX_XPU_ONEDNN_LAYOUT=1 numactl -N $node -l python3 main.py -a resnet50 \
        -b ${BATCH_SIZE} \
        --xpu 0 --pretrained --evaluate  \
        --num-iterations ${ITEM_NUM}  --jit-trace  \
        ${MAIN_ARGS} 2>&1 | tee ${OUTPUT_DIR}/pt_resnet50_${PRECISION}_inf_d${p}_raw.log &
    
    elif [[ ${Tile} == "2" ]]; then
        dev_id=`expr $p + 1`
        dev=${dev_id}p
        mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
        echo $mac
        node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
        echo $node

        ZE_AFFINITY_MASK=${p}.0 IPEX_XPU_ONEDNN_LAYOUT=1 numactl -N $node -l python3 main.py -a resnet50 \
        -b ${BATCH_SIZE} \
        --xpu 0 --pretrained --evaluate  \
        --num-iterations ${ITEM_NUM} --jit-trace  \
        ${MAIN_ARGS} 2>&1 | tee ${OUTPUT_DIR}/pt_resnet50_${PRECISION}_inf_d${p}t0_raw.log &
        ZE_AFFINITY_MASK=${p}.1 IPEX_XPU_ONEDNN_LAYOUT=1 numactl -N $node -l python3 main.py -a resnet50 \
        -b ${BATCH_SIZE} \
        --xpu 0 --pretrained --evaluate  \
        --num-iterations ${ITEM_NUM} --jit-trace  \
        ${MAIN_ARGS} 2>&1 | tee ${OUTPUT_DIR}/pt_resnet50_${PRECISION}_inf_d${p}t1_raw.log &    
    else
        echo "Only Tiles 1 and 2 supported."
        exit 1
    fi
    sleep 1
done

# wait all process done
wait

echo "---------Summary--------"
echo "Throughput per process:"
grep -H "Evalution performance" ${OUTPUT_DIR}/pt_resnet50_*_raw.log

total_throughput=$( grep -H "Evalution performance" ${OUTPUT_DIR}/pt_resnet50_*_raw.log | awk -F':' '{ print $5}' | awk -F' ' '{ sum+=$1 } END{print sum} ' )
echo "Total Throughput: ${total_throughput} imgs/sec"
echo "Batch Size: ${BATCH_SIZE}"
echo "Precision: ${PRECISION}"
echo "Tile: ${Tile}"
echo "Number of GPU: ${NUMBER_OF_GPU}"

# Clean
rm -fr $OUTPUT_DIR
