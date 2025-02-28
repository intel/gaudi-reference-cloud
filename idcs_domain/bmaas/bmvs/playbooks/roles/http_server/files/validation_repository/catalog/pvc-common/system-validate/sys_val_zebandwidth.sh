#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
source ${ONEAPI_ROOT}/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

LOG_DIR=${WORKSPACE}/bench_zebw_logs
WORK_DIR=${WORKSPACE}/bench_zebw_workdir
mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cp ${SCRIPT_DIR}/../workloads/levelzero/ze_bandwidth ${WORK_DIR}/
cd ${WORK_DIR}

export PATH=$PATH:${SCRIPT_DIR}/../utils:${WORK_DIR}
echo "Current PATH environments:"
echo $PATH
which ze_bandwidth

dry_run=${DRYRUN:-0}
TESTSET=${TESTSET:-1}

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
else
   tiles=1
fi

device_name=$(lspci -vmm |grep Max | head -n 1)

XPUM=${XPUM:-xpu-smi}
if [ ! -e "`which ${XPUM}`" ]; then
    echo "${XPUM} not found, switch to xpumcli"
    XPUM=xpumcli
    if [ ! -e "`which ${XPUM}`" ]; then
	echo "Can't find XPUManager installed! Please install xpumanager or xpu-smi."
	exit -1
    fi
fi

#
zebw_bwid=("h2d" "d2h" "bidir")
# EU copy, through group 0, implicit scaling disabled
zebw_bwmagic_g0=(37 52 43)
#BCS0 copy through group 1 (no implicit scaling)
zebw_bwmagic_g1=(28 45 34)
#BCS1 copy through group 2
zebw_bwmagic_g2=(27 39 32)
threshold=0.05

#data for x8 system (smc sys-821)
zebw_bwmagic_max1550_x8_g0=(33 42 37)
zebw_bwmagic_max1550_x8_g1=(40 42 48)
zebw_bwmagic_max1550_x8_g2=(22 42 29)

#data for x8 system (smc sys-421)
zebw_bwmagic_max1100_x8_g0=(37 52 43)
zebw_bwmagic_max1100_x8_g1=(28 45 34)
zebw_bwmagic_max1100_x8_g2=(27 39 32)


#data for x4 system (dnp x4)
zebw_bwmagic_max1550_x4_g0=(44 51 47)
zebw_bwmagic_max1550_x4_g1=(47 51 60)
zebw_bwmagic_max1550_x4_g2=(26 42 32)


if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
    zebw_bwmagic_g0=( ${zebw_bwmagic_max1550_x8_g0[@]} )
    zebw_bwmagic_g1=( ${zebw_bwmagic_max1550_x8_g1[@]} )
    zebw_bwmagic_g2=( ${zebw_bwmagic_max1550_x8_g2[@]} )
fi

if [ "$tiles" == "1" ] && [ "$num_gpu" == "8" ] ; then
    zebw_bwmagic_g0=( ${zebw_bwmagic_max1100_x8_g0[@]} )
    zebw_bwmagic_g1=( ${zebw_bwmagic_max1100_x8_g1[@]} )
    zebw_bwmagic_g2=( ${zebw_bwmagic_max1100_x8_g2[@]} )
fi

if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
    zebw_bwmagic_g0=( ${zebw_bwmagic_max1550_x4_g0[@]} )
    zebw_bwmagic_g1=( ${zebw_bwmagic_max1550_x4_g1[@]} )
    zebw_bwmagic_g2=( ${zebw_bwmagic_max1550_x4_g2[@]} )
fi




report_xelink_topo_error(){
	echo "DATAVALIDATION: XE Link check FAILED: It seems the GPU are not fully connected, please check the XE Link connections to get better performance"
	exit -1	
}
report_xelink_topo_pass(){
	echo "DATAVALIDATION: XE Link check PASSED: It seems the GPU are fully connected"
}
validate_sys_topology(){
	print_prefix "Find $num_gpu GPU devices: $device_name, Topology:"
	gpu_topo_file=${LOG_DIR}/gpu_topo.log
	eval " $XPUM topology -m " 2>&1 |tee ${gpu_topo_file}
	print_aftfix
	val=`bc -l <<< "$num_gpu <=1"`
	if (( val )); then
	    echo "GPU number less than 1, ze_peer test not needed"
	    exit 0

	fi

	if [ "$tiles" == "1" ] && [ "$num_gpu" == "2" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 2 (2x1) XL24 connection between GPU 0 and 1"
		num_xl=$( cat $gpu_topo_file | grep -oh XL24 | wc -l )
		if [ "$num_xl" != "2" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
		return
	
	fi


	if [ "$tiles" == "1" ] && [ "$num_gpu" == "4" ] ; then
		echo "1100x4 validation, this is TBD once we have such configuration"
		return
	fi

	if [ "$tiles" == "1" ] && [ "$num_gpu" == "8" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 24 (4x3x2) XL8 connection across GPU 0-3 and GPU 4-7"
		num_xl=$( cat $gpu_topo_file | grep -oh XL8 | wc -l )
		if [ "$num_xl" != "24" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
		return
	fi
	if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 24 (8x3) XL8 connection across GPU 0-3"
		num_xl=$( cat $gpu_topo_file | grep -oh XL8 | wc -l )
		if [ "$num_xl" != "24" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
	
		return
	
	fi
	if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 112 (16x7) XL4 connection across GPU 0-7"
		num_xl=$( cat $gpu_topo_file | grep -oh XL4 | wc -l )
		if [ "$num_xl" != "112" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
	
		return
	
	fi

	echo "GPU topology not supported by this script"
	exit -1

}
dry_run=${DRYRUN:-0}

validate_zebw_bw_result(){
        if [ ! -f $log_file ]; then
	        echo "$log_file not found"
	        return
	fi

	local test_engine="EU"
	if [ "$engine" == "0" ]; then
		test_engine="EU"
		zebw_bwmagic=( ${zebw_bwmagic_g0[@]} )
	elif [ "$engine" == "1" ]; then
		test_engine="BCS0"
		zebw_bwmagic=( ${zebw_bwmagic_g1[@]} )
	elif [ "$engine" == "2" ]; then
		test_engine="BCS1"
		zebw_bwmagic=( ${zebw_bwmagic_g2[@]} )
	fi

	local bwval=( `cat $log_file | grep BW |awk '{print $5}'` )
	local benchref=${zebw_bwmagic[$test_id]}
	local val=`bc -l <<< "$bwval >= $benchref*(1-$threshold)"`
	local bench_ratio=`bc -l <<< "scale=2;$bwval / $benchref * 100"`
    if (( $val ));then
		result="PASSED"
	else
		result="WARNING"
	fi
	printf "DATAVALIDATION ze_bandwidth - %-5s through %-5s for GPU %d Tile %d: %-10s: %-5.2f(GBPS) ref: %-5.2f ratio: %.2f%%\n" ${test_item} $test_engine $gpu $t $result $bwval $benchref $bench_ratio
	csv_file="zebw_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, Kernel, Engine, GPU, Tile, Result, Performance(GBPS), Reference, Ratio (%), CMD" > $csv_file
	fi
	echo "ze_bandwidth,${test_item},${test_engine},${gpu},${t},${result},${bwval},${benchref},${bench_ratio},\"$cmd\"" >> $csv_file

}

#single target test
validate_zebw_bw(){
    zebw_bin=ze_bandwidth
    zebw_opt="-n 0 -s 268435456 -i 100 --immediate"

    for gpu in $( seq 0 $((num_gpu-1))); do
        bdf=$(${XPUM} discovery -d ${gpu} |grep BDF | awk -F'0000:' '{print $2}' |awk -F' ' '{print $1}')
		numanode=$(lspci -s $bdf -vv |grep NUMA |awk -F: '{print $2}')
		numacmd="numactl -N $numanode -m $numanode "

    	print_prefix "Bench ze_bandwidth for GPU $gpu with DATAVALIDATION"
		for engine in 0 1 2; do
	  		test_id=0
 	  		for test_item in ${zebw_bwid[@]}; do
	   			for t in $( seq 0 $((tiles-1))); do
	      			log_file=$LOG_DIR/zebw_${test_item}_u${engine}_d${gpu}_t${t}.log
	      			cmd="ZE_AFFINITY_MASK=${gpu}.${t} $numacmd $zebw_bin $zebw_opt -g $engine -t $test_item"
	      			if (( $dry_run )); then
        				echo $cmd
    	      		else
						eval "$cmd  2>&1 >$log_file"
	      			fi
	      			validate_zebw_bw_result
	    		done
          		test_id=$(( $test_id +1))
          	done
		done
    done
}


#parallel run test
validate_zebw_bw_parallel(){
    print_prefix "Bench ze_bandwidth in parallel on all $num_gpu GPU"
    zebw_bin=ze_bandwidth
    zebw_opt="-n 0 -s 268435456 -i 100 --immediate"
    
    for test_item in ${zebw_bwid[@]}; do
	echo "Test Item - $test_item, engine 0 1 2"
        for engine in 0 1 2; do
	    num_process=$(( num_gpu * tiles ))
	    log_file=$LOG_DIR/zebw_${test_item}_u${engine}_parallel.log
	    cmd="$zebw_bin $zebw_opt -g $engine -t $test_item"
	    cmd_p="parallel_run.sh -c \"$cmd\" -p $num_process -n"
	    if (( $dry_run )); then
	        echo $cmd_p
	    else
		echo $cmd_p
	        eval "$cmd_p" 2>&1 >$log_file
	    fi
	    grep GBPS $log_file
	done
	print_aftfix
    done
}




#validate_sys_topology

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
  validate_zebw_bw
fi

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "2" ]; then
  validate_zebw_bw_parallel
fi


