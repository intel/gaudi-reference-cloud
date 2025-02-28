#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
dry_run=${DRYRUN:-0}
DOCKER=${DOCKER:-0}
TESTCASE=${TESTCASE}

if [ "$DOCKER" != "0" ]; then
    LOG_DIR=${WORKSPACE}/bench_gromacs_container_logs
    WORK_DIR=${WORKSPACE}/bench_gromacs_container_workdir
else
    LOG_DIR=${WORKSPACE}/bench_gromacs_logs
    WORK_DIR=${WORKSPACE}/bench_gromacs_workdir
fi

mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}

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
device_name=$(echo $(lspci -vmm |grep Max | head -n 1 |awk -F: '{print $2}') )

CPU_NAME=$( echo $( lscpu|grep "Model name:" |awk -F: '{print $2}') )
LOGIC_CORES=$( lscpu |grep "CPU(s):" |awk -F: '{print$2}' |awk 'NR==1{print$1}'|tr -d ' ' )
PHYSICAL_CORES=$( lscpu |grep "Core(s) per socket:" |awk -F: '{print$2}'|tr -d ' ' )
SOCKETS=$( lscpu |grep "Socket(s):" |awk -F: '{print$2}'|tr -d ' ' )
echo "System Info: 
	CPU name: $CPU_NAME
       	Logic cores: $LOGIC_CORES
	Physical cores: $PHYSICAL_CORES
	Sockets: $SOCKETS
       	GPU Card number: $num_gpu
       	Tiles per card: $tiles
       	Device name = $device_name
	"

if [ "${TESTCASE}" == "" ]; then
    testcases=(benchMEM benchPEP benchPEP-h STMV)
else
    testcases=($TESTCASE)
fi

threshold=0.05

set_test_config(){
    if [ ! "$CPU_NAME" == "Intel(R) Xeon(R) Platinum 8480+" ]; then
	echo "WARNING: The reference data is collected on $cpu_name. Different type of CPU/system configurations may have significant performance difference."
    fi
    tiles2runs=(1)
    if [ $tiles == 1 ]; then
        reference_data=(207)
    fi
    if [ $tiles == 2 ]; then
	reference_data=(251)
    fi

    #DNP
    if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
	tiles2runs=(1 2 4 8)
	if   [ "$testcase" == "benchMEM" ];then
  	    reference_data=(251.875 209.541 215.964 149.82)
	elif [ "$testcase" == "benchPEP" ]; then
	    reference_data=(2.127 1.439 2.511 0)
	elif [ "$testcase" == "benchPEP-h" ]; then
	    reference_data=(2.106 1.566 2.887 3.268)
	elif [ "$testcase" == "STMV" ]; then
	    reference_data=(19.234 14.963 29.381 0)
	fi
    fi

    #SYS-421
    if [ "$tiles" == "1" ] && [ "$num_gpu" == "8" ] ; then
	tiles2runs=(1 2 4)
	if   [ "$testcase" == "benchMEM" ];then
  	    reference_data=(207.48 191.65 227.83 0)
	elif [ "$testcase" == "benchPEP" ]; then
	    reference_data=(1.698 1.254 2.181 0)
	elif [ "$testcase" == "benchPEP-h" ]; then
	    reference_data=(1.686 1.28 2.375 0)
	elif [ "$testcase" == "STMV" ]; then
	    reference_data=(15.198 12.564 24.886 0)
	fi

    fi
    
    #SYS-821
    if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
	tiles2runs=(1 2 4 8 16)
	if   [ "$testcase" == "benchMEM" ];then
  	    reference_data=(199 0 0 0 0)
	elif [ "$testcase" == "benchPEP" ]; then
	    reference_data=(0 0 0 0 0)
	elif [ "$testcase" == "benchPEP-h" ]; then
	    reference_data=(0 0 0 0 0)
	elif [ "$testcase" == "STMV" ]; then
	    reference_data=(0 0 0 0 0)
	fi
    fi   
    
}

report_perf(){
    currentdata=$(grep -h Performance: $log_file | awk -F' ' '{print $2}')
    if [ "$currentdata" == "" ]; then
	result="FAILED"
	echo "DATAVALIDATION: Gromacs $testcase on ${card_num} GPU with ${tiles2run} Tiles, $tiles2run ranks, $ntomp threads per rank: Runtime failed. Check $log_file for details."
    else
	__ref=${reference_data[$tiles2run_id]}
        if [[ $(echo "${currentdata} >= $__ref * (1-$threshold)"|bc) -eq 1 ]]; then
	    val=1
	else
	    val=0
	fi
	ratio="n/a"
	if [[ $(echo "$__ref > 0" |bc) -eq 1 ]]; then
	   ratio=$( bc -l <<< "scale=2; ${currentdata} / $__ref * 100")
	fi
	if (( $val )); then
	   result="PASSED"
        else
	   result="FAILED"
	fi
        echo "DATAVALIDATION Gromacs $testcase on ${card_num} GPU with ${tiles2run} Tiles, $tiles2run ranks, $ntomp threads per rank $result: $currentdata (ns/day), ref=$__ref, ratio=$ratio%"
    fi
    
}
get_tiles2run_id(){
    for i in "${!tiles2runs[@]}"; do
	if [ "$tiles2run" == "${tiles2runs[$i]}" ]; then
	   tiles2run_id=$i
	   return
	fi
    done
}

evaluate_gromacs(){
    #running in benchmark folder
    cp -r ${SCRIPT_DIR}/../../workloads/HPC/gromacs/* ${WORK_DIR}
    bench_script_dir=${WORK_DIR}
    if [ "$DOCKER" != "0" ]; then
        script2run=./gromacs_container_run.sh
        log_file_prefix=${LOG_DIR}/gromacs_container_run
        script2setup=./gromacs_container_setup.sh
        log_file_setup=${LOG_DIR}/gromacs_container_setup.log
    else
        script2run=./gromacs_run.sh
        log_file_prefix=${LOG_DIR}/gromacs_run
        script2setup=./gromacs_setup.sh
        log_file_setup=${LOG_DIR}/gromacs_setup.log
    fi

    cd $bench_script_dir
    print_prefix "### Set up GROMACS working environment ..."
    if (( ! $dry_run ));then
        eval "$script2setup 2>&1 |tee $log_file_setup"
    else
        echo "$script2setup"
    fi

    for testcase in ${testcases[@]}; do
        print_prefix "Gromacs benchmark evaluation - $testcase on $device_name with DATAVALIDATION"
	set_test_config
	if [ ! "${TILES2RUN}" == "" ];then
	    tiles2runs=($TILES2RUN)
	fi
	for tiles2run in ${tiles2runs[@]}; do
	    get_tiles2run_id
	    echo "----Run test case $testcase on $tiles2run GPU tiles"
	    ntomp=$(( PHYSICAL_CORES / tiles2run ))
	    card_num=$((tiles2run / tiles))
	    if [ "$card_num" == "0" ]; then 
		card_num=1 
	    fi
	    cmd="TESTCASE=$testcase NTMPI=$tiles2run NTOMP=$ntomp $script2run"
	    log_file=${log_file_prefix}_${testcase}_MPI${tiles2run}_OMP${ntomp}_${card_num}C${tiles2run}T.log
	    if (( $dry_run )); then
		echo $cmd
	    else
		echo $cmd 2>&1 |tee $log_file
		echo "------" 2>&1 >> $log_file
		eval $cmd 2>&1 >> $log_file
	    fi
	    report_perf
    	    print_aftfix
	done
    done
}

evaluate_gromacs
