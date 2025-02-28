#!/bin/bash

# Set global flag for exit code
is_test_failed=0

(( EUID != 0 )) && exec sudo -E -- "$0" "$@"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
bkcconfig=${BKCCONFIG:-}
bkcconfig_file=$SCRIPT_DIR/config/${bkcconfig}.cfg
if [ -e "$bkcconfig_file" ]; then
	echo "Use BKC Config file $bkcconfig_file for version check"
elif [ ! "$bkcconfig" == "" ]; then
	echo "ERROR: BKC config file $bkcconfig_file not exist"
	exit -1
fi

VERBOSE=${VERBOSE:-0}
if (( ! $VERBOSE ));then
  verbose=" >/dev/null"
fi
STAGE=${STAGE:-0}

#Device ID #Max 1450 ID to be confirmed
device_id_intel=(56c1 56c0 0bda  0bd5)
device_name_intel=("Flex 140" "Flex 170" "Max 1100" "Max 1550")

# Running log path
LOG_PATH="${SCRIPT_DIR}/sanitycheck_logs"
#_$(date '+%Y%m%d_%H%M%S')"
mkdir -p $LOG_PATH
echo "Log files will be saved in folder $LOG_PATH"

ZE_PEAK_DIR=${ZE_PEAK:-${SCRIPT_DIR}/../workloads/levelzero/ze_peak}
export PATH=$PATH:${ZE_PEAK_DIR}

# The tool requires XPUManager or xpu-smi installed
XPUM=${XPUM:-xpu-smi}
if [ ! -e "`which ${XPUM}`" ]; then
	echo "${XPUM} not found, switch to xpumcli"
	XPUM=xpumcli
	if [ ! -e "`which ${XPUM}`" ]; then
		echo "Can't find XPUManager installed! Please install xpumanager or xpu-smi."
		exit -1
	fi
fi

#tick-tock count
ttc=${TTC:-120}
zepeakloop=${ZEPEAKITERATION:-1000}
gpu_maxtemp=${MAXTEMP:-80}

print_prefix(){
	echo "--------------------------------------------------------------------------------"
	echo "# $1"
	echo "--------------------------------------------------------------------------------"
}

print_affix(){
	echo "--------------------------------------------------------------------------------"
	echo ""
}

detect_system_info(){	
	LOGS="${LOG_PATH}/1_system_info.log"
	OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"

	print_prefix "Stage 1: Detecting System Info" 2>&1 |tee $LOGS
	eval   "echo \"##1. Host CPU\" $OUTPUT_C"
	eval   "lscpu $OUTPUT 
		print_affix $OUTPUT 
		numactl -H $OUTPUT 
		print_affix $OUTPUT 
		numastat -m $OUTPUT"
	cpu_type=$(grep "Model name" -A5 $LOGS)
	numa_node=$(grep "NUMA node(s):" -A2 $LOGS)
	memory_info=$(grep "MemTotal" -B2 $LOGS)
	eval   "echo '$cpu_type' $OUTPUT_C"
	eval   "echo '$numa_node' $OUTPUT_C"
	eval   "echo '$memory_nfo' $OUTPUT_C"
	eval   "print_affix $OUTPUT_C "
	
	eval   "echo \"##2. Host OS\" $OUTPUT_C"
	eval   "cat /etc/*-release $OUTPUT 
		uname -a $OUTPUT 
		cat /proc/cmdline $OUTPUT "
	os_version=$(grep PRETTY_NAME -A2 $LOGS)
	kernel_version=$(uname -a)
	boot_image=$(grep BOOT_IMAGE $LOGS)
	eval   "echo '$os_version' $OUTPUT_C"
	eval   "echo '$kernel_version' $OUTPUT_C"
	eval   "echo '$boot_image'"
	eval   "print_affix $OUTPUT_C"

}
scan_gpu_device(){
	LOGS="${LOG_PATH}/2_gpu_device.log"
        OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"

	print_prefix "Stage 2: Detecting System GPU Device Info" 2>&1 |tee $LOGS
	
	eval "echo '##1. Intel GPU Device' $OUTPUT_C"
	gpu_num=$(lspci | grep -ic display)
	if [ $gpu_num -gt 0 ]; then
		eval "echo 'Intel GPU device detected: $gpu_num' $OUTPUT_C"
		eval "lspci -nn | grep -i display $OUTPUT"
	else
		eval "echo 'ERROR: No Intel GPU device is detected' $OUTPUT_C"
		exit -1
	fi
	
	for i in ${!device_id_intel[@]}; do
		if [[ $(lspci -d:${device_id_intel[$i]}) ]]; then
			device_id=${device_id_intel[$i]}
			device_name=${device_name_intel[$i]}
			break
		fi
	done
	if [ "${device_id}x" == "x" ]; then
		eval "echo 'ERROR: GPU type not recognized.' $OUTPUT_C"
		exit -1
	fi
	eval   "echo \"GPU Device Type - Intel Data Center GPU $device_name\" $OUTPUT_C"
	eval "print_affix $OUTPUT_C"
	
	eval "echo '##2. GPU DKMS Driver version' $OUTPUT_C"
	eval   "dkms status |grep intel-i915-dkms $OUTPUT_C"
	if [[ ! $(dkms status |grep intel-i915-dkms) ]]; then
		eval "echo WARNING: intel-i915-dkms driver not found..."
	fi
	eval   "print_affix $OUTPUT_C"


	eval "echo '##3. Check Intel GPU system node'"
	nodes=$(ls  /dev/dri/* | grep -c renderD)
	if [ $nodes -gt 0 ]; then
		eval "echo 'Intel GPU nodes is detected: $nodes' $OUTPUT_C"
		eval "ls -l /dev/dri/by-path $OUTPUT"
		eval "ls -l /dev/dri/renderD* $OUTPUT"
	else
		eval "echo 'ERROR: No Intel GPU node detected. Intel GPU Driver may not installed properly.' $OUTPUT_C"
		exit -1
	fi
		
	if [ $nodes == $gpu_num ]; then
		eval "echo 'All GPU nodes are detected successfully' $OUTPUT_C"
	else
		eval "echo 'ERROR: GPU device num ($gpu_num) and node num ($nodes) not matching. $OUTPUT_C"
		eval "echo 'ERROR: The system GPU drivers or GPU card may have problems' $OUTPUT_C"
		exit -1
	fi
	eval "print_affix $OUTPUT_C"

}

check_device_fw(){
	LOGS="${LOG_PATH}/3_device_firmware.log"
        OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"

	print_prefix "Stage 3: Discover GPU Runtime Info" 2>&1 |tee $LOGS	
	eval "echo '##1. Detect GPU Device' $OUTPUT_C"
	dev_num=$(($(${XPUM} discovery --dump 1,2  | wc -l) - 1))
	eval "echo 'Detected $dev_num level zero device' $OUTPUT_C"
	eval "${XPUM} discovery --dump 1,2 $OUTPUT_C"
	eval "print_affix $OUTPUT_C"
	
	
	eval "echo '##2. Discover GPU Device Details' $OUTPUT_C"
	discovery0=$( ${XPUM} discovery -d 0 )
	device_name0=$(echo $( echo "$discovery0" |grep "Device Name" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	driver_version0=$(echo $(echo "$discovery0" |grep "Driver Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	gfxfw_version0=$(echo $(echo "$discovery0" |grep "GFX Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	gfxpscfw_version0=$(echo $(echo "$discovery0" |grep "GFX PSC Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))	
	amcfw_version0=$(echo $(echo "$discovery0" |grep "AMC Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	gpu_memory0=$(echo $(echo "$discovery0" |grep "Memory Physical Size" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	ecc_state0=$(echo $(echo "$discovery0" |grep "ECC State" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
	eval "echo '
	Device ID:                 0 
	Device Name:               $device_name0 
	Driver Version:            $driver_version0 
	GFX Firmware Version:      $gfxfw_version0 
	GFX PSC Firmware Version:  $gfxpscfw_version0 
	AMC Firmware Version:      $amcfw_version0 
	Memory Physical Size:      $gpu_memory0 
	ECC State:                 $ecc_state0
	' $OUTPUT_C"
	eval "echo '$discovery0' $OUTPUT"

	if [ ! "$bkcconfig" == "" ]; then
		eval "echo Check GPU firmware version with BKC config - $bkcconfig"
		gfxfw_bkc=$(echo $(cat $bkcconfig_file |grep "GFX Firmware Version" |awk -F: '{print $2}' ) )
		gfxpscfw_bkc=$(echo $(cat $bkcconfig_file |grep "GFX PSC Firmware Version" |awk -F: '{print $2}' ) )
		amcfw_bkc=$(echo $(cat $bkcconfig_file |grep "AMC Firmware Version" |awk -F: '{print $2}' ) )

		gfxfw_bkc_vn=$(echo $gfxfw_bkc |awk -F. '{print $2}')
		gfxpscfw_bkc_vn=$(echo $gfxpscfw_bkc |awk -F'.0x' '{print $2}')
		amcfw_bkc_vn=$(echo $amcfw_bkc | awk -F. '{print $1$2$3$4}')
		echo "BKC firmware version number: $gfxfw_bkc_vn $gfxpscfw_bkc_vn $amcfw_bkc_vn"

		gfxfw_version0_vn=$(echo $gfxfw_version0 |awk -F. '{print $2}')
		gfxpscfw_version0_vn=$(echo $gfxpscfw_version0 |awk -F'.0x' '{print $2}')
		amcfw_version0_vn=$(echo $amcfw_version0 | awk -F. '{print $1$2$3$4}')
		echo "GPU 0 firmware version number: $gfxfw_version0_vn $gfxpscfw_version0_vn $amcfw_version0_vn"


		eval "echo 'BKC Firmware version:
	GFX Firmware Version:      $gfxfw_bkc
	GFX PSC Firmware Version:  $gfxpscfw_bkc
	AMC Firmware Version:      $amcfw_bkc 
				' $OUTPUT_C"
		if [ ! "$gfxfw_version0" == "$gfxfw_bkc" ] && [ $gfxfw_version0_vn -lt $gfxfw_bkc_vn ]; then
			eval "echo 'WARNING: GFX Firmware Version ($gfxfw_version0) of GPU 0 does NOT match and lower than BKC version ($gfxfw_bkc)' $OUTPUT_C"
		else
			eval "echo 'INFO: GFX Firmware Version ($gfxfw_version0) of GPU 0 match or higher than BKC version ($gfxfw_bkc)' $OUTPUT_C"
		fi
		if [ ! "$gfxpscfw_version0" == "$gfxpscfw_bkc" ] && [ $gfxpscfw_version0_vn -lt $gfxpscfw_bkc_vn ]; then
			eval "echo 'WARNING: GFX PSC Firmware Version ($gfxpscfw_version0) of GPU 0 does NOT match and lower than BKC version ($gfxpscfw_bkc)' $OUTPUT_C"
		else
			eval "echo 'INFO: GFX PSC Firmware Version ($gfxpscfw_version0) of GPU 0 match or higher than BKC version ($gfxpscfw_bkc)' $OUTPUT_C"
		fi
		if [ ! "$amcfw_version0" == "$amcfw_bkc" ] && [ $amcfw_version0_vn -lt $amcfw_bkc_vn ]; then
			eval "echo 'WARNING: AMC Version ($amcfw_version0) of GPU 0 does NOT match and lower than BKC version ($amcfw_bkc)' $OUTPUT_C"
		else
			eval "echo 'INFO: AMC Version ($amcfw_version0) of GPU 0 match or higher than BKC version ($amcfw_bkc)' $OUTPUT_C"
		fi

	fi
	
	eval "echo '' $OUTPUT_C"

	for d in $(seq 1 $(( dev_num - 1 )) ); do
		discovery_else=$( ${XPUM} discovery -d ${d} )
		device_name_else=$(echo $( echo "$discovery_else" |grep "Device Name" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		driver_version_else=$(echo $(echo "$discovery_else" |grep "Driver Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		gfxfw_version_else=$(echo $(echo "$discovery_else" |grep "GFX Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		gfxpscfw_version_else=$(echo $(echo "$discovery_else" |grep "GFX PSC Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		amcfw_version_else=$(echo $(echo "$discovery_else" |grep "AMC Firmware Version" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		gpu_memory_else=$(echo $(echo "$discovery_else" |grep "Memory Physical Size" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		ecc_state_else=$(echo $(echo "$discovery_else" |grep "ECC State" |awk -F: '{print $2}'|awk -F'|' '{print $1}'))
		if [[ "$device_name_else" == "$device_name0" && "$driver_version_else" == "$driver_version0" && \
			"$gfxfw_version_else" == "$gfxfw_version0" && "$amcfw_version_else" == "$amcfw_version0" &&\
			"$gfxpscfw_version_else" == "$gfxpscfw_version0" && \
			"$gpu_memory_else" == "$gpu_memory0" && "$ecc_state_else" == "$ecc_state0"  ]]; then
			eval "echo 'Device $d firmware version matches with Device 0' $OUTPUT_C"
		else
			eval "echo 'WARNING: Device $d configs does not match with Device 0' $OUTPUT_C"
			eval "echo '
			Device ID:                 $d 
			Device Name:               $device_name_else
			Driver Version:            $driver_version_else
			GFX Firmware Version:      $gfxfw_version_else
			GFX PSC Firmware Version:  $gfxfw_version_else
			AMC Firmware Version:      $amcfw_version_else
			Memory Physical Size:      $gpu_memory_else
			ECC State:                 $ecc_state_else
			' $OUTPUT_C"
		fi
		eval "echo '$discovery_else' $OUTPUT"
	done

	print_affix
}

check_gpu_topology(){
	LOGS="${LOG_PATH}/4_gpu_topo.log"
        OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"

	print_prefix "Stage 4: Check GPU Topologies" 2>&1 |tee $LOGS

	num_gpu=$( lspci |grep Display |wc -l )
	#0bd5 is Max 1550 which has two tiles
	device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
	if [ ! -z "$device_type" ]; then
	    tiles=2
	else
	    tiles=1
	fi

	eval "echo '## GPU Topologies in System' $OUTPUT_C"
	gpu_topo=$( ${XPUM} topology -m )
	eval "echo '$gpu_topo' $OUTPUT_C"

	xl_str=""
	if [ "$tiles" == "1" ] && [ "$num_gpu" == "2" ] ; then
		xl_str=XL24
		num_xl=$( echo "$gpu_topo" | grep -oh XL24 | wc -l )
		num_xl_expect=2
	fi

	if [ "$tiles" == "1" ] && [ "$num_gpu" == "8" ] ; then
		xl_str=XL8
                num_xl=$( echo "$gpu_topo" | grep -oh XL8 | wc -l )
		num_xl_expect=24
	fi

	if [ "$tiles" == "2" ] && [ "$num_gpu" == "4" ] ; then
		xl_str=XL8
                num_xl=$( echo "$gpu_topo" | grep -oh XL8 | wc -l )
		num_xl_expect=24
	fi

	if [ "$tiles" == "2" ] && [ "$num_gpu" == "8" ] ; then
		xl_str=XL4
                num_xl=$( echo "$gpu_topo" | grep -oh XL4 | wc -l )
		num_xl_expect=112
	fi

	if [ "${xl_str}x" == "x" ]; then
		eval "echo 'WARNING: Unknow system GPU topologies' $OUTPUT_C"
	else
	        if [ "$num_xl" != "$num_xl_expect" ];then
		        eval "echo 'XE Link check FAILED: It seems the PVC GPU are not fully connected, please check the XE Link connections' $OUTPUT_C"
			eval "echo 'XE Link check ERROR: Expect (${num_xl_expect}) (${xl_str}) link but get (${num_xl}) $OUTPUT_C"
			exit -1
		else
			eval "echo '' $OUTPUT_C"
			eval "echo 'XE Link check PASSED: It seems the GPU are fully connected. Found (${num_xl}) (${xl_str}) XE Link connections.' $OUTPUT_C"
		fi										 
	fi	

	# Check if XE Link Calibration is done
	for d in $(seq 0 $(( num_gpu - 1 )) ); do 
		gpu_calib=$( ${XPUM} discovery -d $d |grep "Xe Link Calibration Date")
		gpu_calib_status=$( echo $gpu_calib | grep "Not Calibrated" )
		gpu_calib_date=$(echo $gpu_calib |awk -F':' '{print $2}' |awk -F'|' '{print $1}' )
		if [ "${gpu_calib_status}x" == "x" ]; then
			eval "echo GPU $d XE Link Calibration check PASSED: Calibration Date=$gpu_calib_date"
		else
			eval "echo WARNING: GPU $d Not Calibrated.  Calibration Date=$gpu_calib_date"
		fi
	done
	eval "print_affix $OUTPUT_C"
}

check_performance_settings(){
	LOGS="${LOG_PATH}/5_system_perf_settings.log"
        OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"
		
	print_prefix "Stage 5: Check System Performance Settings" 2>&1 |tee $LOGS

	eval "echo '## CPU Scaling Governor' $OUTPUT_C"
	cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor >/tmp/scaling_governor.txt

	n_core=0
	n_perf_core=0 
	while read line 
	do
		n_core=$((n_core+1))
		if [ "$line" == "performance" ]; then
			n_perf_core=$((n_perf_core+1))			
		fi
	done < /tmp/scaling_governor.txt	
	rm /tmp/scaling_governor.txt
	
	eval "echo '$n_perf_core in $n_core core scaling_governor is set to \"performance\"' $OUTPUT_C"
	if [ $n_perf_core -lt $n_core ]; then
		if [ "${user}" == "root" ]; then
			echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
			eval "echo 'Set all cpu scaling governor to \"performance\" for benchmarking' $OUTPUT_C"
		else
			eval "echo 'WARNING: Please set the cpu scaling governor to \"performance\" for benchmarking. Try:' $OUTPUT_C"
			eval "echo 'echo performance | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor' $OUTPUT_C"
		fi	
	else
		eval "echo 'CPU scaling governor check PASSED: All core scaling_governor is \"performance\" ' $OUTPUT_C"
	fi
	eval "print_affix $OUTPUT_C"

}

check_gpu_work_env(){
	LOGS="${LOG_PATH}/6_check_gpu_work_env.log"
        OUTPUT="2>&1 |tee -a $LOGS $verbose"
	OUTPUT_C="2>&1 |tee -a $LOGS"

	print_prefix "Stage 6: Check GPU working environment" 2>&1 |tee $LOGS
	
	if [ ! -e "`which ze_peak`"  ]; then
		eval "echo 'Can't find ze_peak from PATH environment' $OUTPUT_C"
		exit -1
	fi
	
	dev_num=$(($(${XPUM} discovery --dump 1,2  | wc -l) - 1))
	eval "echo 'Detected $dev_num level zero device' $OUTPUT_C"
	
	eval "echo Launching ${XPUM} process to monitor GPU temperature $OUTPUT_C"
	xpum_pids=()
	for d in $(seq 0 $(( dev_num - 1 )) ); do
		eval "echo 'Launching monitoring process on GPU ${d}' $OUTPUT"
		${XPUM} dump -m 3 -d ${d} -n ${ttc} 2>&1 >"${LOG_PATH}/${XPUM}_dump_dev${d}.log" &
		xpum_pids+=" $!"
	done
	#delay 3 sec and captue GPU initial empty temperature
	sleep 3
	
	eval "echo 'Launching ze_peak process on all GPUs' $OUTPUT_C"
	ze_peak_pids=()
	cd $ZE_PEAK_DIR
	for d in $(seq 0 $(( dev_num - 1 )) ); do			
		eval "echo 'Launching ze_peak process on GPU ${d}' $OUTPUT"
		ze_peak -t int_compute -i ${zepeakloop} -d ${d} 2>&1 >"${LOG_PATH}/ze_peak_dev${d}.log" &
		ze_peak_pids+=" $!"
	done
	eval echo "Waiting monitor process exit in ${ttc} seconds..."
	wait ${xpum_pids[@]}
	kill -9 ${ze_peak_pids[@]}  2>/dev/null	
	
	print_affix
	cd ${SCRIPT_DIR}
	
	for d in $(seq 0 $(( dev_num - 1 )) ); do
		echo "GPU ${d}:"
        temp_isna=0
        temp_isna_gate=30
		index=0
		name="${LOG_PATH}/${XPUM}_dump_dev${d}.log"
		while read line ; do
			IFS=, read -r ts id temp <<< $line
			if [ -z $temp ]; then
				eval "echo ERROR: Cannot monitor GPU Core Temperature through ${XPUM} $OUTPUT_C"
				exit -1
			fi

            temp=$(echo $temp)
            if [ "$temp" == "N/A" ]; then
                temp_isna=$((temp_isna+1))
            fi

			temp_warning=$(bc -l <<< "$temp >= $gpu_maxtemp")
			if (( $temp_warning )); then
				echo "WARNING -- GPU ${d} temperature too high. Temperature captured: $temp, Gating:$gpu_maxtemp"
				echo "This will impact the GPU performance significantly if not resolved"
				is_test_failed=1
			fi
		    MYARRAY[$index]=$temp
		    index=$(($index+1))
		done <<< $(sed 1d $name)

        if [ $temp_isna -gt $temp_isna_gate ] || [ $temp_isna -eq $index ] ; then
            echo "WARNING -- GPU ${d} temperature not captured in $temp_isna samples"
			is_test_failed=1
        fi
		
		#echo "${MYARRAY[@]}"
		result_str=$(
		( IFS=$'\n'; echo "${MYARRAY[*]}" ) | awk '{if(min==""){min=max=$1}; if($1>max) {max=$1}; \
			if($1<min) {min=$1}; total+=$1; count+=1} END \
			{printf "Monitor %d sample points: GPU Core Temperature min. %.2f, max. %.2f\n", count, min, max}'
				)
		eval "echo $result_str $OUTPUT_C"
		eval "print_affix $OUTPUT_C"
	done
}

if [[ ${STAGE} -le 1 ]]; then
	detect_system_info
fi

if [[ ${STAGE} -le 2 ]]; then
	scan_gpu_device
fi

if [[ ${STAGE} -le 3 ]]; then
	check_device_fw
fi

if [[ ${STAGE} -le 4 ]]; then
	check_gpu_topology
fi

if [[ ${STAGE} -le 5 ]]; then
	check_performance_settings
fi

if [[ ${STAGE} -le 6 ]]; then
	check_gpu_work_env
fi


if [[ $is_test_failed -eq 1 ]]; then
    echo "Setting exit code to 1 due to WARNING/FAILURE"
    exit 1
fi
