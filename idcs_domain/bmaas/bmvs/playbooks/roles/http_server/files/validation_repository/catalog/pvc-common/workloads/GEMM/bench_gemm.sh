#!/bin/bash

#run mkl gemm benchmarks with N times and get the average output

# Use script dir for workspace by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
LOG_DIR=${WORKSPACE}/bench_gemm_logs

help(){
        echo "options: -t | --type <fp64, fp32, f32, f16, bf16, s8, default fp64>"
        echo "       : -i | --iter <number of iterations to run, default 5 to get average data>"
        echo "       : -h | --help "
        exit	
}

gemm_type=fp64
iter=5
gemm_opt=

options=t:,i:,h
optionl=type:,iter:,help
OPTS=$(getopt -a -n $0 --options $options --longoptions $optionl -- "$@")
eval set -- "$OPTS"
while :
do
	case "$1" in
	  -t | --type )
	    gemm_type="$2"
	    shift 2
	    ;;
	  -i | --iter )
	    iter="$2"
	    shift 2
	    ;;
	  -h | --help)
	    help
	    exit 0
	    ;;
	  --)
            shift;
	    break
	    ;;
	  *)
            echo "Unexpected option: $1"
	    ;;
	esac
done

case $gemm_type in
	fp64 | FP64)
		gemmbin=matrix_mul_mkl
		gemmopt="double 8192"
		test_id=DGEMM
		;;
	fp32 | FP32)
		gemmbin=matrix_mul_mkl
		gemmopt="single 8192"
		test_id=SGEMM
		;;
	f32 | F32)
		gemmbin=benchdnn
		gemmopt="--matmul --mode=p --perf-template=\"performance : %Gflops% GFLOPS\" --engine=gpu --dt=f32 8192x8192:8192x8192"
		test_id=SGEMM-DNN
		;;
	f16 | F16)
		gemmbin=benchdnn
		gemmopt="--matmul --mode=p --perf-template=\"performance : %Gflops% GFLOPS\" --engine=gpu --dt=f16 8192x8192:8192x8192"
		test_id=FP16GEMM
		;;
	bf16 | BF16)
		gemmbin=benchdnn
		gemmopt="--matmul --mode=p --perf-template=\"performance : %Gflops% GFLOPS\" --engine=gpu --dt=bf16 8192x8192:8192x8192"
		test_id=BF16GEMM
		;;
	s8 | S8)
		gemmbin=benchdnn
		gemmopt="--matmul --mode=p --perf-template=\"performance : %Gflops% GFLOPS\" --engine=gpu --dt=s8 8192x8192:8192x8192"
		test_id=IGEMM
		;;
	*)
		echo "Wrong data type provided"
		help
		;;
esac

if [ ! -e ${SCRIPT_DIR}/${gemmbin} ]; then
	echo "${gemmbin} not found in ${SCRIPT_DIR}"
fi

mkdir -p ${LOG_DIR}

print_prefix (){
	echo "-----------------------------------------------------------------"
	echo "# $1"
	echo "-----------------------------------------------------------------"
}
print_aftfix(){
	echo "-----------------------------------------------------------------"
	echo ""
}

device_id=${ZE_AFFINITY_MASK:-0}
log_files=()
for i in $( seq 0 $((iter-1)) ); do
	print_prefix "${i}: ${gemmbin} ${gemmopt}"
	date4log=$(date '+%Y%m%d_%H%M%S')
	log_files[$i]=${LOG_DIR}/${test_id}_d${device_id}_i${i}.log
	eval "${SCRIPT_DIR}/${gemmbin} ${gemmopt}" 2>&1 |tee ${log_files[$i]} 
	print_aftfix
done

#use env to set the refernce data and threshold
REFERENCE_DATA=${REFERENCE_DATA:-0}
THRESHOLD=${THRESHOLD:-0.1}
validate_with_reference(){
	perf_data=$(printf '%.1f' $1)
	reference_diff=$( bc -l <<<  "$REFERENCE_DATA * $THRESHOLD" )
	reference_data=$( bc -l <<< "$REFERENCE_DATA - $reference_diff" )
	val=$(echo "$perf_data >= $reference_data" |bc -l)	
	ratio=$( bc -l <<< "$perf_data / $REFERENCE_DATA * 100")
	ratio=$( printf '%.2f' $ratio )
	threshold=$( bc -l <<< "$THRESHOLD * 100")
	threshold=$( printf '%3.2f' $threshold)
	if (( $val )); then
		result="PASSED"
	else
		result="WARNING"
	fi
	echo "DATAVALIDATION $test_id on Device ${device_id}: $result - [current data: $perf_data, reference data: $REFERENCE_DATA, ratio: $ratio%]"

	# report to a csv file
	csv_file="gemm_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, Device ID, Performance(GFLOPS), Reference, Ratio(%), Result " > $csv_file
	fi
	echo "${test_id},${device_id},${perf_data},${REFERENCE_DATA},${ratio},${result}" >> $csv_file
}

report_gemm_summary(){
	print_prefix "Summary"
	perf=$(grep -h -E 'performance.*: [0-9].*' ${log_files[@]})
	echo "$perf"
	num_runs=$( echo "$perf" | wc -l )
	perf_avg=$( grep -h -E 'performance.*: [0-9].*' ${log_files[@]} | awk -F: '{print $2}' | awk -F' ' '{x+=$1}END{print x/NR}' )
	unit=$( grep -h -E 'performance.*: [0-9].*' ${log_files[@]} |grep TF)
	if [ ! -z "$unit" ] ; then
		perf_avg=$( bc -l <<< "$perf_avg * 1000" )
	fi 
	echo "Average performance $test_id Device ${device_id} from $num_runs runs: $perf_avg GFLOPS"
	echo ""
	if (( $REFERENCE_DATA != 0 )); then
		validate_with_reference $perf_avg
	fi
	print_aftfix
}

report_gemm_summary




