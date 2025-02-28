#/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
source ${ONEAPI_ROOT}/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

LOG_DIR=${WORKSPACE}/bench_stream_logs
WORK_DIR=${WORKSPACE}/bench_stream_workdir
mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cp ${SCRIPT_DIR}/../workloads/BabelStream/* ${WORK_DIR}
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

validate_stream(){
	print_prefix "STREAM benchmark evaluation - Number of GPU: $num_gpu, GPU $device_name "
	stream_ops=( Copy Mul Add Triad Dot )

	if [ $tiles == 1 ]; then
		reference_data=850000:830000:800000:800000:660000
		threshold=0.05
	fi
	if [ $tiles == 2 ]; then
		reference_data=2350000:2150000:1950000:2050000:1600000
		threshold=0.05
	fi

	if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
	  print_prefix "Run STREAM benchmark on each GPU one by one with DATAVALIDATION"
	  for n in $( seq 0 $((num_gpu-1))); do
	    log_file=${LOG_DIR}/validate_stream_test1_d${n}.log
	    #_$(date '+%Y%m%d_%H%M%S').log		
	    if (( $dry_run ));then
	        echo "IMPLICIT_SCALING=1 parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data} THRESHOLD=${threshold} bench_stream.sh -i 3\" -d $n -p 1"
	        cat $log_file |grep DATAVALIDATION
	    else
	        eval "IMPLICIT_SCALING=1 parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data} THRESHOLD=${threshold} bench_stream.sh -i 3\" -d $n -p 1" 2>&1 |tee $log_file |grep DATAVALIDATION
	    fi
	  done		
	  print_aftfix
	fi

	if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "2" ]; then
	  print_prefix "Run STREAM benchmark on all GPUs in parallel with DATAVALIDATION"
	  num_process=$num_gpu
	  log_file=${LOG_DIR}/validate_stream_testp${num_process}.log
	  #_$(date '+%Y%m%d_%H%M%S').log
	  if (( $dry_run )); then
	    echo "IMPLICIT_SCALING=1 parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data} THRESHOLD=${threshold} bench_stream.sh  -i 3\"  -p $num_process"
	    cat $log_file |grep DATAVALIDATION
	  else
	    eval "IMPLICIT_SCALING=1 parallel_run.sh -c \"WORKSPACE=${WORKSPACE} REFERENCE_DATA=${reference_data} THRESHOLD=${threshold} bench_stream.sh  -i 3\"  -p $num_process" 2>&1 |tee $log_file |grep DATAVALIDATION
	  fi
	  print_aftfix
	fi

}

validate_stream

