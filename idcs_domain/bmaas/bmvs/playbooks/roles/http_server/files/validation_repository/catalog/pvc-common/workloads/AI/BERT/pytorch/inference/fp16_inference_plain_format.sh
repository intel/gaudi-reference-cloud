#!/usr/bin/env bash
#
# Copyright (c) 2021-2023 Intel Corporation
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
#

MODEL_DIR=${MODEL_DIR-$PWD}
BATCH_SIZE=${BATCH_SIZE-64}

#source ${MODEL_DIR}/quickstart/setvars.sh

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    Tile=${Tile:-2}
else
    Tile=${Tile:-1}
fi


if [[ -z "${DATASET_DIR}" ]]; then
    echo "The required environment variable DATASET_DIR has not been set"
    exit 1
fi

if [[ ! -d "${DATASET_DIR}" ]]; then
    echo "The DATASET_DIR '${DATASET_DIR}' does not exist"
    exit 1
fi

if [[ -z $OUTPUT_DIR ]]; then
    echo "The required environment variable OUTPUT_DIR has not been set"
    exit 1
fi

# Create the output directory, if it doesn't already exist
mkdir -p $OUTPUT_DIR

bertsquad_log_analysis() {
    # $1 : src raw log
    # $2 : dst format log
    # $3 : inference or training
    # $4 : bs

    if [ -f $2 ]; then
        rm $2
    fi

    bs=$4

    if [ "inference" == "$3" ]; then
        echo -e 'Batch Size: ' $bs >$2
        cat $1 | grep latency | tail -n6 | head -n4 |
            awk -v bs=${bs} -F ' ' '{sum+=$8} END{printf "Performance Benchmark Time: %.3f sec, Throughput: %.2f seq/sec\n", sum/4, bs*4/sum}' >>$2
        grep "\"f1\": " $1 | awk -F ' ' '{printf "Accuracy: f1 %.4f\n", $NF}' >>$2
    elif [ "training" == "$3" ]; then
        # only for fine tune (accuracy only)
        echo -e 'Batch Size: ' $bs >$2
        echo -e 'Performance Benchmark Time: N/A' >>$2
        grep "\"f1\": " $1 | awk -F ' ' '{printf "Accuracy: f1 %.4f\n", $NF}' >>$2
    else
        echo -e 'Invalid input! Only inference or training are supported.'
        exit 0
    fi
}

NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DEVICEID=${DEVICEID:-0}

echo "bertsquad fp16 inference plain nchw for warm-up to generate the needed trace model/cached dataset"
cd ${MODEL_DIR}/models/language_modeling/pytorch/bert_large/inference/gpu/
bash cmd_infer.sh \
    -m bert_large \
    -d xpu \
    -b $BATCH_SIZE \
    -t FP16 \
    -o None

cd ${MODEL_DIR}
source "quickstart/common/utils.sh"

for p in $(seq ${DEVICEID} $(( DEVICEID + NUMBER_OF_GPU - 1 )) ); do
    if [[ ${Tile} == "1" ]]; then
        echo "bertsquad fp16 inference plain nchw"
        dev_id=`expr $p + 1`
        dev=${dev_id}p
        mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
        echo $mac
        node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
        echo $node

        cd ${MODEL_DIR}/models/language_modeling/pytorch/bert_large/inference/gpu/
        ZE_AFFINITY_MASK=${p} _command numactl -N $node -l bash cmd_infer.sh \
            -m bert_large \
            -d xpu \
            -b $BATCH_SIZE \
            -t FP16 \
            -o None 2>&1 | tee ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0_raw.log &
    elif [[ ${Tile} == "2" ]]; then
        echo "bertsquad fp16 inference plain nchw 2 tile"
        dev_id=`expr $p + 1`
        dev=${dev_id}p
        mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
        echo $mac
        node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
        echo $node

        cd ${MODEL_DIR}/models/language_modeling/pytorch/bert_large/inference/gpu/
        ZE_AFFINITY_MASK=${p}.0 _command numactl -N $node -l bash cmd_infer.sh \
            -m bert_large \
            -d xpu \
            -b $BATCH_SIZE \
            -t FP16 \
            -o None 2>&1 | tee ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0_raw.log &
        ZE_AFFINITY_MASK=${p}.1 _command numactl -N $node -l bash cmd_infer.sh \
            -m bert_large \
            -d xpu \
            -b $BATCH_SIZE \
            -t FP16 \
            -o None 2>&1 | tee ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t1_raw.log &
    else
        echo "The specified Tile '${Tile}' is unsupported."
        echo "Supported tile number are: 1 and 2"
        exit 1
    fi
done

wait

for p in $(seq ${DEVICEID} $(( DEVICEID + NUMBER_OF_GPU - 1 )) ); do
    if [[ ${Tile} == "1" ]]; then
        bertsquad_log_analysis ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0_raw.log ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0.log inference ${BATCH_SIZE}
    elif [[ ${Tile} == "2" ]]; then
        bertsquad_log_analysis ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0_raw.log ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t0.log inference ${BATCH_SIZE}
        bertsquad_log_analysis ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t1_raw.log ${OUTPUT_DIR}/bertsquad_fp16_inf_plain_nchw_d${p}_t1.log inference ${BATCH_SIZE}
    fi
done

cd -
