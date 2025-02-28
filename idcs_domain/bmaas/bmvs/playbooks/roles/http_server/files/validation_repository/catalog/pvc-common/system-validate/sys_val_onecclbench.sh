#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
source ${ONEAPI_ROOT}/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

LOG_DIR=${WORKSPACE}/bench_oneccl_logs
WORK_DIR=${WORKSPACE}/bench_oneccl_workdir

mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cp ${SCRIPT_DIR}/../workloads/oneccl/benchmark.2021.11 $WORK_DIR/benchmark
cd ${WORK_DIR}

export PATH=$PATH:${WORK_DIR}
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

#  magic data based on driver 647.21, oneapi 2023.2.0, oneccl 2021.10
oneccl_benchid=(allreduce  allgatherv  alltoall alltoallv bcast reduce reduce_scatter )
oneccl_benchmagic=(0 0 0 0 0 0 0 )
ratio=0.05

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
		xlport=6
		gpu_lists=("0,1")
		oneccl_benchmagic=(7900 7270 1910 1900 1500 4920 3980)
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
	        xlport=2
	        gpu_lists=("0,1,2,3" "4,5,6,7")	
		oneccl_benchmagic=(3100 5200 5700 5700 4400 2800 1500)
		return
	fi
	if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 24 (8x3) XL8 connection across GPU 0-3"
		num_xl=$( cat $gpu_topo_file | grep -oh XL8 | wc -l )
		if [ "$num_xl" != "24" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
		xlport=2
		gpu_lists=("0,1,2,3")
		oneccl_benchmagic=(2100 8560 11410 11410 4810 1640 1500)
		return
	
	fi
	if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
		echo "DATAVALIDATION: Checking GPU topology, expect 112 (16x7) XL4 connection across GPU 0-7"
		num_xl=$( cat $gpu_topo_file | grep -oh XL4 | wc -l )
		if [ "$num_xl" != "112" ];then
			report_xelink_topo_error
		fi
		report_xelink_topo_pass
		xlport=1
		gpu_lists=("0,1,2,3,4,5,6,7")
		oneccl_benchmagic=(2600 19500 22600 22600 9100 1850 1350)
		return
	
	fi

	echo "GPU topology not supported by this script"
	exit -1

}
dry_run=${DRYRUN:-0}

validate_onecclbench_result(){
	local testid=$1
	local testid_str=${oneccl_benchid[$testid]}
	local log_file=$2
	local xlport=$3
	local gpu_list=$4

	local benchval_a=( `cat $log_file |grep t_avg -A1 |tail -1 |awk '{print $5}'` )
	local benchref=${oneccl_benchmagic[$testid]}
	local val=`bc -l <<< "$benchval_a < $benchref*(1+$ratio)"`
	local bench_ratio="N/A"
	if (( $benchref )); then
	    #usec, lower is better
  	    bench_ratio=`bc -l <<< "$benchref / $benchval_a  * 100"`
	    bench_ratio=`printf '%.2f' $bench_ratio`
	fi

    if (( $val ));then
		val_str="PASSED"
	else
		val_str="WARNING"
	fi
	echo "DATAVALIDATION oneccl benchmark on GPU $gpu_list: $testid_str : $val_str (usec)[current=$benchval_a, ratio=$bench_ratio%, ref=$benchref]"

	csv_file="oneccl_benchmark_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, Kernel, GPU, Result, Performance (usec), Reference, Ratio(%), CMD" > $csv_file
	fi
	echo "oneccl benchmark,${testid_str},\"${gpu_list}\",${val_str},${benchval_a},${benchref},${bench_ratio},\"${cmd}\"" >> $csv_file

}

validate_onecclbench(){
    onecclbench_bin=benchmark
    onecclbench_opt="-w 16 -i 1000 -c last -b sycl -t 33554432 -f 33554432 -j off "

    for gpu_list in ${gpu_lists[@]}; do
    	print_prefix "Bench oneccl for GPU $gpu_list with DATAVALIDATION"
	num_gpu=$( echo $gpu_list | awk -F, '{print NF}' )
	num_process=$(( $num_gpu * $tiles ))
	test_id=0
	for test_item in ${oneccl_benchid[@]}; do
	    log_file=$LOG_DIR/onecclbench_${test_item}_${gpu_list}.log
	    cmd="ZE_AFFINITY_MASK=$gpu_list mpiexec -np $num_process $onecclbench_bin $onecclbench_opt  --coll $test_item"
	    if (( $dry_run )); then
        	echo $cmd
    	    else
		eval "$cmd  2>&1 >$log_file"
	    fi
	    validate_onecclbench_result $test_id $log_file $xlport $gpu_list
	    test_id=$(( $test_id +1 ))
        done
    done
}

validate_sys_topology

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
  validate_onecclbench
fi

