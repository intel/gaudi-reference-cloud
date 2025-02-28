#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
dry_run=${DRYRUN:-0}
DOCKER=${DOCKER:-0}
TESTCASE=${TESTCASE}

if [ "$DOCKER" != "0" ]; then
    LOG_DIR=${WORKSPACE}/bench_lammps_container_logs
    WORK_DIR=${WORKSPACE}/bench_lammps_container_workdir
else
    LOG_DIR=${WORKSPACE}/bench_lammps_logs
    WORK_DIR=${WORKSPACE}/bench_lammps_workdir
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

tiles=1
#Max 1550 device id 0bd5
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
	    tiles=2
fi
device_name=$(echo $(lspci -vmm |grep Max | head -n 1 |awk -F: '{print $2}') )
GPU_MODEL=`clinfo -l | grep "GPU Max" |awk  'NR==1 {print  $NF}'`

CPU_NAME=$( echo $( lscpu|grep "Model name:" |awk -F: '{print $2}') )
LOGIC_CORES=$( lscpu |grep "CPU(s):" |awk -F: '{print$2}' |awk 'NR==1{print$1}'|tr -d ' ' )
PHYSICAL_CORES=$( lscpu |grep "Core(s) per socket:" |awk -F: '{print$2}'|tr -d ' ' )
SOCKETS=$( lscpu |grep "Socket(s):" |awk -F: '{print$2}'|tr -d ' ' )
num_gpu=$( lspci |grep Display |wc -l )
hostname=$(hostname)
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
    testcases=(lc lj eam water tersoff) 
else
    testcases=($TESTCASE)
fi

threshold=0.05


set_test_config(){
   tiles2runs=(1)
    if [ $tiles == 1 ]; then
        reference_data=(0)
    fi
    if [ $tiles == 2 ]; then
   	reference_data=(0)
    fi

    #sys-421
    if [ "$tiles" == "1" ] && [ "$num_gpu" == "8" ] ; then
	if [ ! "$CPU_NAME" == "Intel(R) Xeon(R) Platinum 8480+" ]; then
    	    echo "WARNING: The reference data is collected on $cpu_name. Different type of CPU/system configurations may have significant performance difference."
        fi
     
	tiles2runs=(1 2 4 8)
	reference_data=(0 0 0 0)
	case $testcase in
	"lc")
	    reference_data=(74 119 175 230)
	    ;;
	"lj")
	    reference_data=(460 579 669 790)
	    ;;
        "eam")
	    reference_data=(223 300 350 430)
	    ;;
        "water")
	    reference_data=(140 228 334 446)
	    ;;
        "tersoff")
	    reference_data=(275 405 486 608)
	    ;;
        esac
    fi	
    #sys-821
    if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
        if [ ! "$CPU_NAME" == "Intel(R) Xeon(R) Platinum 8468V" ]; then
    	    echo "WARNING: The reference data is collected on $cpu_name. Different type of CPU/system configurations may have significant performance difference."
        fi    
    
	tiles2runs=(1 2 4 8 16)
	reference_data=(0 0 0 0 0)
	case $testcase in
	"lc")
	    reference_data=(83 123 171 213 227)
	    ;;
	"lj")
	    reference_data=(430 524 668 697 787)
	    ;;
        "eam")
	    reference_data=(228 273 341 406 471)
	    ;;
        "water")
	    reference_data=(160 243 360 448 497)
	    ;;
        "tersoff")
	    reference_data=(293 388 496 591 662)
	    ;;
        esac
    fi	
    #DNP
    if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
	if [ ! "$CPU_NAME" == "Intel(R) Xeon(R) Platinum 8480+" ]; then
    	    echo "WARNING: The reference data is collected on $cpu_name. Different type of CPU/system configurations may have significant performance difference."
        fi     
    
	tiles2runs=(1 2 4 8)
	reference_data=(0 0 0 0)
	case $testcase in
	"lc")
	    reference_data=(82 124 180 227)
	    ;;
	"lj")
	    reference_data=(443 529 715 766)
	    ;;
        "eam")
	    reference_data=(240 284 355 425)
	    ;;
        "water")
	    reference_data=(162 243 365 460)
	    ;;
        "tersoff")
	    reference_data=(301 403 508 612)
	    ;;
        esac
    fi	

  
}

report_perf(){
    current_perf=$( grep Performance $log_file | awk '{c=NF-1; print $c}' )
    echo "Performance $current_perf Matom-step/s"    
    csv_file=lammps_perf.csv
    if [ ! -e "$csv_file" ]; then
        echo "System, Device, Workload, Perf (Matom-step/s), # of GPU, # of Tiles, # of Ranks" >$csv_file
    else
	echo "$hostname, Max $GPU_MODEL, $testcase, $current_perf, $card_num, $tiles2run, $rank_num " >>$csv_file
    fi
}

validate_perf_result(){
    currentdata=${ranks2explore_perfsorted[0]}
    if [ "$currentdata" == "" ]; then
	result="FAILED"
    	echo "DATAVALIDATION: Lammps $testcase on ${card_num} GPU with ${tiles2run} Tiles: Runtime failed."
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
        echo "DATAVALIDATION Lammps $testcase on ${card_num} GPU with ${tiles2run} Tiles $result: $currentdata (Matom-step/sec), ref=$__ref, ratio=$ratio%"
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

get_exploration_ranks(){
    case $tiles2run in
    "1")
	    ranks2explore="1 2 4 8 16 32 64"
	    ;;
    "2")
	    ranks2explore="2 4 8 16 32 64"	    
	    ;;
    "4")
	    ranks2explore="4 8 16 32 64"
	    ;;
    "8")
	    ranks2explore="8 16 32 64"
	    ;;
    "16")
	    ranks2explore="16 32 64"
    esac

}

evaluate_lammps(){
    #running in benchmark folder
    cp -r ${SCRIPT_DIR}/../../workloads/HPC/lammps/* ${WORK_DIR}
    bench_script_dir=${WORK_DIR}
    if [ "$DOCKER" != "0" ]; then
        script2run=./lammps_container_run.sh
        log_file_prefix=${LOG_DIR}/lammps_container_run
        script2setup=./lammps_container_setup.sh
        log_file_setup=${LOG_DIR}/lammps_container_setup.log
    else
        script2run=./lammps_run.sh
        log_file_prefix=${LOG_DIR}/lammps_run
        script2setup=./lammps_setup.sh
        log_file_setup=${LOG_DIR}/lammps_setup.log
    fi

    cd $bench_script_dir
    print_prefix "### Set up LAMMPS working environment ..."
    if (( ! $dry_run ));then
        eval "$script2setup 2>&1 |tee $log_file_setup"
    else
        echo "$script2setup"
    fi

    for testcase in ${testcases[@]}; do
        print_prefix "LAMMPS benchmark evaluation - $testcase on $device_name with DATAVALIDATION"
	set_test_config
        if [ ! "${TILES2RUN}" == "" ];then
	    tiles2runs=($TILES2RUN)
	fi
	for tiles2run in ${tiles2runs[@]}; do
	    get_tiles2run_id
    	    echo "----Run test case $testcase on $tiles2run GPU tiles"
    	    card_num=$((tiles2run / tiles))
    	    if [ "$card_num" == "0" ]; then 
    		card_num=1
	    fi
	    
	    get_exploration_ranks
	    ranks2explore_perf=()
	    for rank_num in $ranks2explore; do
 	      cmd="TESTCASE=$testcase NUMBER_OF_PROCESS=$rank_num TILES=$tiles2run $script2run"
	      log_file=${log_file_prefix}_${testcase}_${card_num}C${tiles2run}T.MPI${rank_num}.log
	      if (( $dry_run )); then
		echo $cmd
    	      else
    		echo $cmd 2>&1 |tee $log_file
    		echo "------" 2>&1 >> $log_file
		eval $cmd 2>&1 >> $log_file
	      fi
	      report_perf
	      ranks2explore_perf+="$current_perf "
    	      print_aftfix
	    done
	    ranks2explore_perfsorted=( $( for x in ${ranks2explore_perf[@]}; do echo $x; done | sort -n -r ) )
	    validate_perf_result
	    print_aftfix
	done
    done

}

evaluate_lammps

