#!/bin/bash


# Use script dir for workspace by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
LOG_DIR=${WORKSPACE}/bench_stream_logs

help(){
        echo "options:"
        echo "       : -i | --iter <number of iterations to run, default 1>"
        echo "       : -h | --help "
        exit	
}

iter=1
streambin=sycl2020-acc-stream
streamopt_prefix="ONEAPI_DEVICE_SELECTOR=level_zero:gpu EnableImplicitScaling=1"
streamopt_aftfix="-s 134217728"
test_id="STREAM"

options=i:,h
optionl=iter:,help
OPTS=$(getopt -a -n $0 --options $options --longoptions $optionl -- "$@")
eval set -- "$OPTS"
while :
do
	case "$1" in
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

if [ ! -e ${SCRIPT_DIR}/${streambin} ]; then
	echo "${streambin} not found in ${SCRIPT_DIR}"
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
	log_files[$i]=${LOG_DIR}/${test_id}_d${device_id}_${i}_${date4log}.log
	eval "${streamopt_prefix} ${SCRIPT_DIR}/${streambin} ${streamopt_aftfix}" 2>&1 |tee ${log_files[$i]} 
	print_aftfix
done

#use env to set the refernce data and threshold
REFERENCE_DATA=${REFERENCE_DATA:-0:0:0:0:0}
IFS=':' read -r -a reference_data <<< ${REFERENCE_DATA}
THRESHOLD=${THRESHOLD:-0.1}
perf_avg=()

validate_with_reference(){
	val=1
	threshold=$( bc -l <<< "$THRESHOLD * 100")
	threshold=$( printf '%3.2f' $threshold)
	retio=()
	
	for t in {0..4}; do
		reference_diff=$( bc -l <<<  "${reference_data[$t]} * ${THRESHOLD} " )
		reference=$( bc -l <<< "${reference_data[$t]} - $reference_diff" )
		_val=$(echo "${perf_avg[$t]} >= $reference" |bc -l)	
		val=$( echo " $_val && $val" |bc -l)
		ratio[$t]=$( bc -l <<< "${perf_avg[$t]} / ${reference_data[$t]} * 100")
		ratio[$t]=$( printf '%.2f' ${ratio[$t]} )
	done

	if (( $val )); then
		echo "DATAVALIDATION $test_id on Device ${device_id}: PASSED --" 
		result="PASSED"
	else
		echo "DATAVALIDATION $test_id on Device ${device_id}: WARNING --"
		result="WARNING"
	fi

	# report to a csv file
	csv_file="stream_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, Device ID, Kernel, Performance(MBytes/sec), Reference, Ratio(%), Result " > $csv_file
	fi

	for t in {0..4}; do
		printf '  --DATAVALIDATION %-6s - %-6s - (MBytes/sec)[current data: %-10.2f reference data: %.2f ratio: %.2f%%] \n' $test_id  ${testop[$t]} ${perf_avg[$t]} ${reference_data[$t]} ${ratio[$t]} 
		echo "${test_id},${device_id},${testop[$t]},${perf_avg[$t]},${reference_data[$t]},${ratio[$t]},${result}" >> $csv_file
	done

}

report_stream_summary(){
	print_prefix "Summary"
	testop=( Copy Mul Add Triad Dot )
	perf=()
	for t in {0..4} ; do
		perf[$t]=$(grep -h ${testop[$t]} ${log_files[@]})
		#echo "${perf[$t]}"
		#num_runs=$( echo "$perf" | wc -l )
		perf_avg[$t]=$( grep -h ${testop[$t]} ${log_files[@]} | awk -F' ' '{print $2}' | awk -F' ' '{x+=$1}END{printf "%.2f", x/NR}' )
	done


	echo "Average performance $test_id on Device ${device_id}:"
	for t in {0..4}; do
		printf '%-10s	%-10.2f	MBytes/sec\n' ${testop[$t]} ${perf_avg[$t]}
	done
	echo ""
	if (( ${reference_data[0]} != 0 )); then
		validate_with_reference
	fi
	print_aftfix
}

report_stream_summary




