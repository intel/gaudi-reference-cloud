#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
source ${ONEAPI_ROOT}/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

LOG_DIR=${WORKSPACE}/bench_zepeer_logs
WORK_DIR=${WORKSPACE}/bench_zepeer_workdir
mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cp ${SCRIPT_DIR}/../workloads/levelzero/ze_peer ${WORK_DIR}/
export PATH=$PATH:${WORK_DIR}
cd $WORK_DIR
echo "Current PATH environments:"
echo $PATH

TESTSET=${TESTSET:-0}
dry_run=${DRYRUN:-0}

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

# for unidirectional one src card to one dst card single process test
# per port bw (XL4, 4 lanes), magic data based on driver 647.21
zepeer_bwid=("read" "write" "biread" "biwrite")
# EU copy, implicit scaling triggered
zepeer_bwmagic_g0=(15.5 18.5 23.2 31.5)
#BCS0 copy (no implicit scaling)
zepeer_bwmagic_g1=(15.5 15.5 23.5 23.5)
ratio=0.02

# For parallel multiple card test, read/write only
zepeer_pmt_bwid=("read" "write")
# magic number reset for each configuration in below.
# EU
zepeer_pmt_bwmagic_g0=( 0 0 )
# BCS0
zepeer_pmt_bwmagic_g1=( 0 0 )
# BCS1-5 for 1100 and BCS1-7 for 1550
zepeer_pmt_bwmagic_g2=( 0 0 )



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
		engine_g2=2
		zepeer_pmt_bwmagic_g0=( 71 94 )
		zepeer_pmt_bwmagic_g1=( 71 71 )
		zepeer_pmt_bwmagic_g2=( 18 40 )

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
		engine_g2="2,4,6"
        zepeer_pmt_bwmagic_g0=( 28 36 )
        zepeer_pmt_bwmagic_g1=( 29 29 )
        zepeer_pmt_bwmagic_g2=( 52 85 )

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
		engine_g2="2,4,6"
        zepeer_pmt_bwmagic_g0=( 28 36 )
        zepeer_pmt_bwmagic_g1=( 29 29 )
        zepeer_pmt_bwmagic_g2=( 50 93 )
		
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
		engine_g2="2,3,4,5,6,7"
	    zepeer_pmt_bwmagic_g0=( 29 37 )
        zepeer_pmt_bwmagic_g1=( 14 15 )
        zepeer_pmt_bwmagic_g2=( 48 28 )
		return
	
	fi

	echo "GPU topology not supported by this script"
	exit -1

}
dry_run=${DRYRUN:-0}

validate_zepeer_bw_result(){
	local testid=$1
	local testid_str=${zepeer_bwid[$testid]}
	local log_file=$2
	if [ ! -f $log_file ]; then
	  echo "$log_file not found"
	  return
	fi


	local xlport=$3
	local gpu_list=$4

	local test_engine="EU"
	zepeer_bwmagic=( ${zepeer_bwmagic_g0[@]} )
	if [ "$engine" == "1" ]; then
		test_engine="BCS0"
		zepeer_bwmagic=( ${zepeer_bwmagic_g1[@]} )
	fi
	local bwval_a=( `cat $log_file | grep GBPS |awk '{print $5}' |sort` )
	local bwval_num=${#bwval_a[*]}
	local bwmin=${bwval_a[0]}
	local bwmax=${bwval_a[$bwval_num-1]}
	local bwmed=${bwval_a[($bwval_num-1)/2]}

	#for UE engine, two tiles are triggered for bench
	#for BCS0, 1 BCS is triggered in ze_peer only even for max1550
	local ref_x=$tiles
	if [ "$engine" == "1" ]; then
		ref_x=1
	fi
	local bwref=`bc -l <<< "${zepeer_bwmagic[$testid]} * $xlport  * $ref_x"`
	local bwval=`bc -l <<< "$bwmin > $bwref*(1-$ratio)"`
	local bwmin_ratio=`bc -l <<< "$bwmin / $bwref * 100"`
	      bwmin_ratio=`printf '%.2f' $bwmin_ratio`
	local bwmax_ratio=`bc -l <<< "$bwmax / $bwref * 100"`
	      bwmax_ratio=`printf '%.2f' $bwmax_ratio`
	local bwmed_ratio=`bc -l <<< "$bwmed / $bwref * 100"`
	      bwmed_ratio=`printf '%.2f' $bwmed_ratio`

    if (( $bwval ));then
		result=PASSED
	else
		result=WARNING
	fi
	echo "DATAVALIDATION ze_peer bandwidth - ${zepeer_bwid[$testid]} through $test_engine across GPU $gpu_list : $result : $bwval_num xelink tested, GBPS(min,med,max)=($bwmin,$bwmed,$bwmax), ratio=($bwmin_ratio%,$bwmed_ratio%,$bwmax_ratio%), ref=$bwref"
	# report to a csv file
	csv_file="zepeer_single_target_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, Kernel, Engine, GPUs, Result, Perf min(GBPS), Perf med, Perf max, Reference, Ratio min(%), Ratio med(%), Ratio max(%), CMD" > $csv_file
	fi
	echo "ze_peer,${zepeer_bwid[$testid]},\"${test_engine}\",\"${gpu_list}\",${result},${bwmin},${bwmed},${bwmax},${bwref},${bwmin_ratio},${bwmed_ratio},${bwmax_ratio},\"${cmd}\"" >> $csv_file

}

#single target test
validate_zepeer_bw(){
    zepeer_bin=ze_peer
    zepeer_opt="-t transfer_bw -z 268435456 -x src -i 10"	  
	for gpu_list in ${gpu_lists[@]}; do
    	print_prefix "Bench ze_peer bandwidth for GPU $gpu_list with DATAVALIDATION"
		for engine in 0 1; do
	  		echo "Test engine - $engine"
 	  		for test_item in ${zepeer_bwid[@]}; do
				if [ "$test_item" == "read" ] ; then
				zepeer_opt2="-o read"
				test_id=0
				elif [ "$test_item" == "write" ]; then
				zepeer_opt2="-o write"
				test_id=1
				elif [ "$test_item" == "biread" ]; then
				zepeer_opt2="-o read -b"
				test_id=2
				elif [ "$test_item" == "biwrite" ]; then
				zepeer_opt2="-o write -b"
				test_id=3
				else
				echo "unknow test item"
				exit -1
				fi

				log_file=$LOG_DIR/zepeer_${test_item}_u${engine}_${gpu_list}.log
				cmd="$zepeer_bin $zepeer_opt -s $gpu_list -d $gpu_list $zepeer_opt2 -u $engine"
				if (( $dry_run )); then
					echo $cmd
				else
					eval "$cmd  2>&1 >$log_file"
				fi
	    		validate_zepeer_bw_result $test_id $log_file $xlport $gpu_list
        	done
	  		print_aftfix
		done
    done
}

validate_zepeer_bw_pmt_result(){
	local testid=$1
	local testid_str=${zepeer_bwid[$testid]}
	local xlport=$2
	local gpu_list=$3
	local test_engine="EU"
	#magic number and retio to be fine tuned
	zepeer_bwmagic=( 0 0 0 0 )
	ratio=0.1

	if [ "$engine" == "0" ]; then
		test_engine="EU"
		zepeer_bwmagic=(${zepeer_pmt_bwmagic_g0[@]})
	elif [ "$engine" == "1" ]; then
		test_engine="BCS0"
		zepeer_bwmagic=(${zepeer_pmt_bwmagic_g1[@]})
	else
		test_engine="BCS{$engine}"
		zepeer_bwmagic=(${zepeer_pmt_bwmagic_g2[@]})
	fi
	for logfile in ${log_files[@]}; do
		if [ ! -f $logfile ]; then
		  echo "$logfile not found"
		  continue
		fi

	 	bwvalue=`cat ${logfile} | grep GBPS |grep -v Device | awk '{print $5}' `
		bwref=${zepeer_bwmagic[$testid]}
		bwval=`bc -l <<< "$bwvalue > $bwref*(1-$ratio)"`
		if [ "$bwref" != "0" ];then
		  bwratio=`bc -l <<< "$bwvalue / $bwref * 100"`
		  bwratio=`printf '%.2f' $bwratio`
	  	else
		  bwratio="n/a"
		  bwval=0
		fi
		src=`echo $logfile | grep -o -E _s.*_ |tr -d _s`
		dst=`echo $logfile | grep -o -E _d.* |tr -d _d.log`
		if (( $bwval )); then
			result=PASSED
		else
			result=WARNING
		fi
		echo "DATAVALIDATION ze_peer PMT - ${zepeer_pmt_bwid[$testid]} - src: $src, dst: $dst, engine: $test_engine : $result : ${bwvalue} GBPS, ref: ${bwref}, ratio: ${bwratio}%"
		# report to a csv file
		csv_file="zepeer_parallel_multi_target_test.csv"
		if ! test -f $csv_file; then
			echo "Workload, Kernel, Engine, GPU src, GPU dst, Result, Perf.(GBPS), Reference, Ratio%, CMD" > $csv_file
		fi
		echo "ze_peer,${zepeer_pmt_bwid[$testid]},\"${test_engine}\",\"${src}\",\"${dst}\",${result},${bwvalue},${bwref},${bwratio},\"${cmd}\"" >> $csv_file
	done
}



#parallel multi target test
validate_zepeer_pmt_bw(){
    zepeer_bin=ze_peer
    zepeer_opt="-t transfer_bw -z 268435456 -x src --parallel_multiple_targets -i 500"

    for gpu_list in ${gpu_lists[@]}; do	    
    	print_prefix "Bench ze_peer bandwidth parallel multiple targets for GPU $gpu_list with DATAVALIDATION"
		for engine in 0 1 $engine_g2; do
	  		echo "Test engine - $engine"
 	  		for test_item in ${zepeer_pmt_bwid[@]}; do
				if [ "$test_item" == "read" ] ; then
				zepeer_opt2="-o read"
				test_id=0
				elif [ "$test_item" == "write" ]; then
				zepeer_opt2="-o write"
				test_id=1
				elif [ "$test_item" == "biread" ]; then
				zepeer_opt2="-o read -b"
				test_id=2
				elif [ "$test_item" == "biwrite" ]; then
				zepeer_opt2="-o write -b"
				test_id=3
				else
				echo "unknow test item"
				exit -1
				fi
	    
				cmd=""
				log_files=()
					gpus=( `echo $gpu_list | sed 's/,/ /g'` )
				for gpu in ${gpus[@]}; do
					gpu_s=$gpu
					gpu_d=$( eval "echo $gpu_list | sed 's/${gpu_s},//' |sed 's/,${gpu_s}//'" )
					log_file=$LOG_DIR/zepeer_${test_item}_pmt_u${engine}_s${gpu_s}_d${gpu_d}.log
					cmd="$cmd $zepeer_bin $zepeer_opt -s $gpu_s -d $gpu_d $zepeer_opt2 -u $engine 2>&1 >$log_file & "
					log_files+=($log_file)
				done

				if (( $dry_run )); then
					echo $cmd
				else
					eval "$cmd"
					wait
				fi

	    		validate_zepeer_bw_pmt_result $test_id  $xlport $gpu_list
         	done
	 		print_aftfix
		done
    done
}


validate_sys_topology

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
  validate_zepeer_bw
fi

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "2" ]; then
  validate_zepeer_pmt_bw
fi

