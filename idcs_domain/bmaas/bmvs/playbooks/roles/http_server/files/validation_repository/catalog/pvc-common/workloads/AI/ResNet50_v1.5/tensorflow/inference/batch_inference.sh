#!/usr/bin/env bash
#
# Copyright (c) 2022 Intel Corporation
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

echo 'MODEL_DIR='$MODEL_DIR
echo 'PRECISION='$PRECISION
echo 'OUTPUT_DIR='$OUTPUT_DIR

cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor >/tmp/scaling_governor.txt 
while read line 
do
   if [ "$line" != "performance" ]; then
    if [ "${user}" == "root" ]; then
        echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
    else
        echo "Please set the cpu scaling governor to 'performance' for benchmarking. Try:"
        echo "echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor"
        rm /tmp/scaling_governor.txt
        exit 1
    fi
   fi
done < /tmp/scaling_governor.txt
rm /tmp/scaling_governor.txt

export TF_NUM_INTEROP_THREADS=1

# Create an array of input directories that are expected and then verify that they exist
declare -A input_envs
input_envs[PRECISION]=${PRECISION}
input_envs[OUTPUT_DIR]=${OUTPUT_DIR}
input_envs[GPU_TYPE]=${GPU_TYPE}

for i in "${!input_envs[@]}"; do
  var_name=$i
  env_param=${input_envs[$i]}
 
  if [[ -z $env_param ]]; then
    echo "The required environment variable $var_name is not set" >&2
    exit 1
  fi
done

# Create the output directory in case it doesn't already exist
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# If batch size env is not mentioned, then the workload will run with the default batch size.
if [ -z "${BATCH_SIZE}" ]; then
  BATCH_SIZE="1024"
  echo "Running with default batch size of ${BATCH_SIZE}"
fi
WARMUP_STEPS=${WARMUP_STEPS:-5}
STEPS=${STEPS:-25}
# Check for GPU type
if [[ $GPU_TYPE == "flex_series" ]]; then
  export OverrideDefaultFP64Settings=1 
  export IGC_EnableDPEmulation=1 
  if [[ $PRECISION == "int8" ]]; then
    echo "Precision is $PRECISION"
    if [[ ! -f "${FROZEN_GRAPH}" ]]; then
      pretrained_model=/workspace/tf-flex-series-resnet50v1-5-inference/pretrained_models/resnet50v1_5-frozen_graph-${PRECISION}-gpu.pb
    else
      pretrained_model=${FROZEN_GRAPH}
    fi
    WARMUP="-- warmup_steps=${WARMUP_STEPS} steps=${STEPS}"
  else 
    echo "FLEX SERIES GPU SUPPORTS ONLY INT8 PRECISION"
    exit 1
  fi
elif [[ $GPU_TYPE == "max_series" ]]; then
  if [[ $PRECISION == "int8" || $PRECISION == "fp16" || $PRECISION == "fp32" ]]; then
    echo "Precision is $PRECISION"
    if [[ ! -f "${FROZEN_GRAPH}" ]]; then
      pretrained_model=/workspace/tf-max-series-resnet50v1-5-inference/pretrained_models/resnet50v1_5-frozen_graph-${PRECISION}-gpu.pb
    else
      pretrained_model=${FROZEN_GRAPH}
    fi
    WARMUP="-- warmup_steps=${WARMUP_STEPS} steps=${STEPS} disable-tcmalloc=True"
  else 
    echo "MAX SERIES GPU SUPPORTS ONLY INT8, FP32 AND FP16 PRECISION"
    exit 1
  fi
fi

if [[ $PRECISION == "int8" ]]; then
   BENCHMARK="--benchmark "
else
   BENCHMARK=" "
fi

if [[ $PRECISION == "fp16" ]]; then
  export ITEX_AUTO_MIXED_PRECISION=1
  export ITEX_AUTO_MIXED_PRECISION_DATA_TYPE="FLOAT16"
fi

NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
DEVICEID=${DEVICEID:-0}

if [[ -z "${Tile}" ]]; then
    Tile=${Tile:-1}
else
    Tile=${Tile}
fi

source "${MODEL_DIR}/quickstart/common/utils.sh"

for p in $(seq ${DEVICEID} $(( DEVICEID + NUMBER_OF_GPU - 1 )) ); do
if [[ ${Tile} == "1" ]]; then
    echo "resnet50 v1.5 ${PRECISION} inference"
    #mac=`lspci | grep Dis| head -n 1| awk '{print $1}'`
     dev_id=`expr $p + 1`
     dev=${dev_id}p
     mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
     echo $mac
     node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
     ZE_AFFINITY_MASK=${p} numactl -N $node -l python -u models/image_recognition/tensorflow/resnet50v1_5/inference/gpu/${PRECISION}/eval_image_classifier_inference.py \
         --input-graph=${pretrained_model} \
         --warmup-steps=${WARMUP_STEPS} \
         --steps=${STEPS} \
         --batch-size=${BATCH_SIZE} \
         ${BENCHMARK} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}_raw.log &

         #ZE_AFFINITY_MASK=${p} python benchmarks/launch_benchmark.py \
         #--model-name=resnet50v1_5 \
         #--precision=${PRECISION} \
         #--mode=inference \
         #--framework tensorflow \
         #--in-graph ${pretrained_model} \
         #--output-dir ${OUTPUT_DIR} \
         #--batch-size=${BATCH_SIZE} \
         #--benchmark-only \
         #--gpu \
         #$@ \
         #${WARMUP} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}_raw.log &

elif [[ ${Tile} == "2" ]]; then
    echo "resnet50 v1.5 ${PRECISION} two-tile inference"
    dev_id=`expr $p + 1`
    dev=${dev_id}p
    mac=`lspci | grep Dis| sed -n $dev| awk '{print $1}'`
    echo $mac
    node=`lspci -s $mac -v | grep NUMA | awk -F, '{print $5}' | awk '{print $3}'`
    ZE_AFFINITY_MASK=${p}.0 numactl -N $node -l python -u models/image_recognition/tensorflow/resnet50v1_5/inference/gpu/${PRECISION}/eval_image_classifier_inference.py \
         --input-graph=${pretrained_model} \
         --warmup-steps=${WARMUP_STEPS} \
         --steps=${STEPS} \
         --batch-size=${BATCH_SIZE} \
         ${BENCHMARK} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}t0_raw.log &

    ZE_AFFINITY_MASK=${p}.1 numactl -N $node -l python -u models/image_recognition/tensorflow/resnet50v1_5/inference/gpu/${PRECISION}/eval_image_classifier_inference.py \
         --input-graph=${pretrained_model} \
         --warmup-steps=${WARMUP_STEPS} \
         --steps=${STEPS} \
         --batch-size=${BATCH_SIZE} \
         ${BENCHMARK} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}t1_raw.log &

        #ZE_AFFINITY_MASK=${p}.0 python benchmarks/launch_benchmark.py \
        # --model-name=resnet50v1_5 \
        # --precision=${PRECISION} \
        # --mode=inference \
        # --framework tensorflow \
        # --in-graph ${pretrained_model} \
        # --output-dir ${OUTPUT_DIR} \
        # --batch-size=${BATCH_SIZE} \
        # --benchmark-only \
        # --gpu \
        # $@ \
        # ${WARMUP} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}t0_raw.log &
        # ZE_AFFINITY_MASK=${p}.1 python benchmarks/launch_benchmark.py \
        # --model-name=resnet50v1_5 \
        # --precision=${PRECISION} \
        # --mode=inference \
        # --framework tensorflow \
        # --in-graph ${pretrained_model} \
        # --output-dir ${OUTPUT_DIR} \
        # --batch-size=${BATCH_SIZE} \
        # --benchmark-only \
        # --gpu \
        # $@ \
        # ${WARMUP} 2>&1 | tee ${OUTPUT_DIR}//resnet50_${PRECISION}_inf_d${p}t1_raw.log &
else
    echo"Only Tiles 1 and 2 supported."
    exit 1
fi
done

wait

