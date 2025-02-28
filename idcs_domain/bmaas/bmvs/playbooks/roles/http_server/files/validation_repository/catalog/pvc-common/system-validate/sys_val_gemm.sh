#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
#source ${ONEAPI_ROOT}/setvars.sh
source ${ONEAPI_ROOT}/compiler/latest/env/vars.sh
source ${ONEAPI_ROOT}/mkl/latest/env/vars.sh
source ${ONEAPI_ROOT}/dnnl/latest/env/vars.sh
source ${ONEAPI_ROOT}/tbb/latest/env/vars.sh

LOG_DIR=${WORKSPACE}/bench_gemm_logs
WORK_DIR=${WORKSPACE}/bench_gemm_workdir
ITERATIONS=${ITERATIONS:-5}
GEMM=${GEMM:-"fp64 fp32 f16 bf16 s8"} #default for all using fp64 fp32 f16 bf16 s8
GPU=${GPU} # GPU ID for single gpu stress
#export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cp ${SCRIPT_DIR}/../workloads/GEMM/* ${WORK_DIR}/
cd ${WORK_DIR}
export PATH=$PATH:${SCRIPT_DIR}/../utils:${WORK_DIR}
echo "Current PATH environments:"
echo $PATH

dry_run=${DRYRUN:-0}
TESTSET=${TESTSET:-0}

print_prefix (){
    echo "-----------------------------------------------------------------"
    echo "# $1"
    echo "-----------------------------------------------------------------"
}

print_aftfix(){
    echo "-----------------------------------------------------------------"
    echo ""
}


num_gpu=$( lspci |grep Display |wc -l )

tiles=1
#Max 1550 device id 0bd5
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
		tiles=2
fi
device_name=$(lspci -vmm |grep Max | head -n 1)

validate_gemm(){
	print_prefix "GEMM benchmark evaluation - Number of GPU: $num_gpu, GPU $device_name "
	if [ $tiles == 1 ]; then
		gemm_type=(fp64 fp32 f16 bf16 s8)
		reference_data=(14000 21000 270000 260000 456000)
		threshold=(0.05 0.05 0.05 0.05 0.05)
	fi
	if [ $tiles == 2 ]; then
		gemm_type=(fp64 fp32 f16 bf16 s8)		
		reference_data=(15500 23000 360000 350000 650000)
		threshold=(0.05 0.05 0.05 0.05 0.05)
	fi
	
	if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
	  print_prefix "Run GEMM benchmark on each GPU one by one"
	  for i in ${!gemm_type[@]}; do
	  	if [ ! "$( echo $GEMM |grep ${gemm_type[i]} )" == "" ]; then
			for n in $( seq 0 $((num_gpu-1))); do
				if [ ! "$GPU" == "" ] && [ ! "$n" == "$GPU" ]; then continue; fi
		  		log_file=${LOG_DIR}/valiate_gemm_test1_${gemm_type[$i]}_d${n}.log
		  		if (( $dry_run )); then
		    		echo "parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data[$i]} THRESHOLD=${threshold[$i]} bench_gemm.sh -t ${gemm_type[$i]} -i ${ITERATIONS}\" -d $n -p $tiles"
		    		cat $log_file |grep DATAVALIDATION
		  		else
		    		eval "parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data[$i]} THRESHOLD=${threshold[$i]} bench_gemm.sh -t ${gemm_type[$i]} -i ${ITERATIONS}\" -d $n -p $tiles" 2>&1 |tee $log_file |grep DATAVALIDATION
		  		fi
			done
		fi
	  done
	  print_aftfix
	fi

	if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "2" ]; then
	  print_prefix "Run GEMM benchmark on all GPUs in parallel"
	  for i in ${!gemm_type[@]}; do
	  	if [ ! "$( echo $GEMM |grep ${gemm_type[i]} )" == "" ]; then
			num_process=$( bc -l <<< "$num_gpu * $tiles" )
			log_file=${LOG_DIR}/valiate_gemm_testp${num_process}_${gemm_type[$i]}.log
			#_$(date '+%Y%m%d_%H%M%S').log
			if (( $dry_run )); then
				echo "parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data[$i]} THRESHOLD=${threshold[$i]} bench_gemm.sh -t ${gemm_type[$i]} -i ${ITERATIONS}\"  -p $num_process"
				cat $log_file |grep DATAVALIDATION
			else
				eval "parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data[$i]} THRESHOLD=${threshold[$i]} bench_gemm.sh -t ${gemm_type[$i]} -i ${ITERATIONS}\"  -p $num_process" 2>&1 |tee $log_file |grep DATAVALIDATION
			fi
		fi
	  done
	  print_aftfix
	fi
}

validate_gemm

