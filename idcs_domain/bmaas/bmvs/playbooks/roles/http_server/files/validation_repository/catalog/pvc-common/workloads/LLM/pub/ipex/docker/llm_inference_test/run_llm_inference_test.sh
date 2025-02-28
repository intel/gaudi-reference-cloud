#!/bin/bash

ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
if test -f ${ONEAPI_ROOT}/setvars.sh ; then
	source ${ONEAPI_ROOT}/setvars.sh
else
    export LD_LIBRARY_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib:$LD_LIBRARY_PATH
    export PATH=/opt/intel/oneapi/redist/bin:$PATH
    export I_MPI_ROOT=/opt/intel/oneapi/redist/lib
    export CCL_ROOT=/opt/intel/oneapi/redist
    export FI_PROVIDER_PATH=/opt/intel/oneapi/redist/opt/mpi/libfabric/lib/prov
fi

ulimit -n 8192
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=2
export ENABLE_SDP_FUSION=1
export TORCH_LLM_ALLREDUCE=1

mkdir -p logs
dry_run=${DRYRUN:-0}

device_name=$(python -c "import torch; import intel_extension_for_pytorch as ipex; print(torch.xpu.get_device_properties(0).name);")
tiles=1
if [ ! "$(echo $device_name |grep 1550)" == "" ]; then 	tiles=2; fi
echo "Device Name: $device_name, Tiles per GPU: $tiles"


report_to_csv(){
  local model=$1
  local log_file=$2
  if ! test -f $log_file; then
	  echo "--------$log_file not exist"
	  return
  fi

 
  local latency=0
  local first=0
  local next=0
  if [ $num_tiles -gt 1 ]; then
      latency=$( grep "Inference latency" ${log_file} | awk '{print $4}' )
      first=$( grep "First token average latency" ${log_file} | awk '{print $6}' )
      next=$( grep "Average 2..." ${log_file} | awk '{print $5}' )
  else
      latency=$( grep "Inference latency" ${log_file} | awk '{print $3}' )
      first=$( grep "First token average latency" ${log_file} | awk '{print $5}' )
      next=$( grep "Average 2..." ${log_file} | awk '{print $4}' )
  fi
#  local p90=$( grep "P90 2..." ${log_file} | awk '{print $4}' )
#  local p99=$( grep "P99 2..." ${log_file} | awk '{print $4}' )
  local tokensps=0
  local csv_file="llm_inference_bench.csv"

  if [ ! -z "$next" ]; then
    #tokensps=`bc -l <<< " 1 / $next "`
    tokensps=$( python -c "print(\"%.1f\" % (1/float($next)) )" )
  fi

  if ! test -f $csv_file; then
	  echo "Device, Model, Input Tokens, Max new Tokens, Beam Width, Data Type, Infer Latency, First Token Latency, Next Token Latency, Next Tokens/Second, # of GPUs, # of Tiles" > $csv_file
  fi
  echo "----------Benchmark $model - Inference Latency: $latency sec, First Token: $first sec, Next Token: $next sec, Tokens/Second: $tokensps, # of GPUs: $num_gpus, Tiles: $num_tiles"
  echo "$device_name, $model, $input_tokens, $max_new_tokens, $beam_width, $datatype, $latency, $first, $next, $tokensps, $num_gpus, $num_tiles" >> $csv_file

}

num_iter=10
num_warmup=5
datatype=float16
device=xpu
batch_size=1
num_tiles=1

models=(EleutherAI/gpt-j-6B decapoda-research/llama-7b-hf bigscience/bloom-7b1 bigscience/bloom-3b bigscience/bloom-1b7 bigscience/bloom-560m facebook/opt-6.7b facebook/opt-1.3b  t5-3b)
beam=(1 4)
inputs=( 32 64 128 256 512 1024 )
outputs=( 32 128 )

#Overwrite the configs by config2run.sh
config2run=${CONFIG2RUN:-config2run.sh}
if test -f $config2run; then
    source $config2run
else
    echo "$config2run file not found"
    exit
fi

for model in "${models[@]}"; do
  model_name=$( echo $model | awk -F/ '{print $NF}' )
  if [ -z ${model_name} ]; then
    model_name=$model
  fi
  echo "=============Benchmark $model_name============"
  for num_tiles in "${num_tiles_set[@]}"; do
   for input_tokens in "${inputs[@]}"; do
    for max_new_tokens in "${outputs[@]}"; do
      for beam_width in "${beam[@]}"; do
	 num_gpus=$((num_tiles / tiles))
         if [ "$num_gpus" == "0" ]; then
            num_gpus=1
         fi 
	log_file=logs/${model_name}_$device_$datatype_${input_tokens}_${max_new_tokens}_beam${beam_width}_${num_gpus}C${num_tiles}T.log
	echo "---------- Start new test, Log file: $log_file ----------"
	if [ $num_tiles -gt 1 ] 2>/dev/null; then
	  #--sub-model-name $model_name 
	  torun="mpirun -np $num_tiles --prepend-rank python -u run_generation_with_deepspeed.py  --benchmark -m ${model}  --num-beam ${beam_width} --num-iter ${num_iter} --batch-size ${batch_size} --input-tokens ${input_tokens} --max-new-tokens ${max_new_tokens} --dtype $datatype --device xpu --ipex --token-latency 2>&1 |tee -a ${log_file}"
	else 
  	  torun="python run_generation.py --benchmark --ipex --device $device -m $model  --dtype $datatype --input-tokens ${input_tokens} --max-new-tokens ${max_new_tokens} ${xpu_optimize} --token-latency --num-iter $num_iter --num-warmup $num_warmup --num-beam ${beam_width} 2>&1 |tee -a ${log_file}"
	fi

	if (( $dry_run )); then
   	  echo $torun
	else
	  echo $torun 2>&1 |tee  ${log_file}
	  echo = >>${log_file}
	  eval $torun
	  echo "sleep 10 second to start next run..."
	  sleep 10
	fi
	report_to_csv $model_name $log_file
      done
    done
  done
 done
done
