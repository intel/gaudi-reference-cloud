#!/bin/bash

# The scripts for NAMD test purpose.
# The performance data will write to namd_test.csv file
# Example to control the namd_test.sh
# For single process namd run:
# WORKLOAD=stmv NUM_DEVICES=2 NUM_CORES=48 ITERATION=1 bash namd_test.sh
# For multi process namd run:
# MPIRUN=1 WORKLOAD=stmv NUM_DEVICES=2 NUM_CORES=24 ITERATION=1 bash namd_test.sh

# setup oneAPI environment
source /opt/intel/oneapi/setvars.sh
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1

# system GPU info
TOTGPUs=`clinfo -l | grep "GPU Max" |awk  '{print  $NF}'`
GPU_MODEL=`clinfo -l | grep "GPU Max" |awk  'NR==1 {print  $NF}'`
TOTGPU_NUM=`clinfo -l | grep "GPU Max" |wc -l`
TILES=2
if [ "$GPU_MODEL" == "1100" ];then
	TILES=1
fi

# system CPU info
LOGIC_CORES=`lscpu |grep "CPU(s):" |awk -F: '{print$2}' |awk 'NR==1{print$1}'|tr -d ' '`
PHYSICAL_CORES=`lscpu |grep "Core(s) per socket:" |awk -F: '{print$2}'|tr -d ' '`
SOCKETS=`lscpu |grep "Socket(s):" |awk -F: '{print$2}'|tr -d ' '`

# Use ENVs to control the test
WORKLOAD=${WORKLOAD:-apoa1}  	# the workload, apoa1 or stmv
FFTWONGPU=${FFTWONGPU:-1}    	# whether the FFTW offload to GPU, valid for single process run
NUM_DEVICES=${NUM_DEVICES:-2}	# the number of devices, each tile as one device
NUM_CORES=${NUM_CORES:-${PHYSICAL_CORES}}  # the number of cores for each process
ITERATION=${ITERATION:=10}		# the number of iterations to run
DRYRUN=${DRYRUN:-0}				# dryrun to echo command and report result only
MPIRUN=${MPIRUN:-0}				# for MPI run with multi-processing namd


if [ "${MPIRUN}" == "0" ]; then
	# Use splittiles to compatible with oneAPI 2023 release
	export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
else
	#for MPIRUN, use FLAT (the default since 476.25 driver)
	export ZE_FLAT_DEVICE_HIERARCHY=FLAT
fi

if [ "${ZE_FLAT_DEVICE_HIERARCHY}" == "COMPOSITE" ]; then
	SPLITTILES="+splittiles"
fi

# update config file based on FFTW on GPU or CPU
# PMEOffload [on|off] # On for offloading Jim’s GPU implementation (only offloads charge spreading and interpolation)
# usePMECUDA [on|off] # on for offloading Antti-pekka’s GPU implementation (offloads the entire FFT operations)
# Use "off" option in both parameters to use the CPU implementation
# BondedCUDA 0 #Disable Bonded-Force offload
config_file=${WORKLOAD}/${WORKLOAD}_nve_cuda.namd
if [ "$MPIRUN" == "0" ]; then
	if [ "${FFTWONGPU}" == "1" ]; then
		# Support up to 6 devices/tiles only
		BondedCUDA=0
		PMEOffload=off
		usePMECUDA=on
	else
		BondedCUDA=0
		PMEOffload=off
		usePMECUDA=off
	fi
else
		# for MPI run
		BondedCUDA=0
		PMEOffload=on
		usePMECUDA=off
fi
bondedcuda=$( grep BondedCUDA $config_file )
if [ "$bondedcuda" == "" ]; then
	echo "BondedCUDA $BondedCUDA" >> $config_file
else
	sed -i -E "s/(BondedCUDA ).*/\1 ${BondedCUDA}/" $config_file
fi
pmeoffload=$( grep PMEOffload $config_file )
if [ "$pmeoffload" == "" ]; then
	echo "PMEOffload $PMEOffload" >> $config_file
else
	sed -i -E "s/(PMEOffload ).*/\1 ${PMEOffload}/" $config_file
fi
usepmecuda=$( grep usePMECUDA $config_file )
if [ "$usepmecuda" == "" ]; then
	echo "usePMECUDA $usePMECUDA" >> $config_file
else
	sed -i -E "s/(usePMECUDA ).*/\1 ${usePMECUDA}/" $config_file
fi

# update devices/gpu tiles to run
NUM_GPUS=$(( $NUM_DEVICES/$TILES ))
DEVICES=$( seq -s, 0 1 $(( NUM_DEVICES-1 )) )
if [ "${MPIRUN}" == "1" ]; then
	PERHOST=${NUM_DEVICES}
fi

# update the PEMAP for cpu affinity
CROSS_SOCKET_PREFERRED=${CROSS_SOCKET_PREFERRED:-1} #distribute the processes to different CPU preferred
if [ "${MPIRUN}" == "0" ]; then
	NUM_CORES_PERCPU=$(( $NUM_CORES/2 ))
	PEMAP_first=$(( $PHYSICAL_CORES - $NUM_CORES_PERCPU ))
	PEMAP_last=$(( $PHYSICAL_CORES + $NUM_CORES_PERCPU -1 ))
	PEMAP="${PEMAP_first}-${PEMAP_last}"
else
	PPN=$((NUM_CORES -1))
	if [ "${CROSS_SOCKET_PREFERRED}" == "1" ]; then
		COMMAP_first_s0=$(( PHYSICAL_CORES - (NUM_CORES*PERHOST/2) ))
		COMMAP_last_s0=$((PHYSICAL_CORES-1))
		COMMAP_first_s1=$((PHYSICAL_CORES))
		COMMAP_last_s1=$((PHYSICAL_CORES*2-1))
		COMMAP_stride=$(( PPN + 1 ))

		PEMAP_first_s0=$(( PHYSICAL_CORES - (NUM_CORES*PERHOST/2) +1 ))
		PEMAP_last_s0=$((PHYSICAL_CORES-1))
		PEMAP_first_s1=$((PHYSICAL_CORES+1))
		PEMAP_last_s1=$((PHYSICAL_CORES*2-1))
		PEMAP_stride=$(( PPN + 1 ))
		PEMAP_run=${PPN}

	else		
		COMMAP_first_s0=0
		COMMAP_last_s0=$((PHYSICAL_CORES-1))
		COMMAP_first_s1=$((PHYSICAL_CORES))
		COMMAP_last_s1=$((PHYSICAL_CORES*2-1))
		COMMAP_stride=$(( PPN + 1 ))

		PEMAP_first_s0=1
		PEMAP_last_s0=$((PHYSICAL_CORES-1))
		PEMAP_first_s1=$((PHYSICAL_CORES+1))
		PEMAP_last_s1=$((PHYSICAL_CORES*2-1))
		PEMAP_stride=$(( PPN + 1 ))
		PEMAP_run=${PPN}
	fi
	
	COMMAP="${COMMAP_first_s0}-${COMMAP_last_s0}:${COMMAP_stride},${COMMAP_first_s1}-${COMMAP_last_s1}:${COMMAP_stride}"
	PEMAP="${PEMAP_first_s0}-${PEMAP_last_s0}:${PEMAP_stride}.${PEMAP_run},${PEMAP_first_s1}-${PEMAP_last_s1}:${PEMAP_stride}.${PEMAP_run}"
fi

# generate log prefix
if [ "${MPIRUN}" == "0" ]; then
	mkdir -p logs/${WORKLOAD}
	log_prefix="logs/${WORKLOAD}/${WORKLOAD}_+p-${NUM_CORES}_+pemap-${PEMAP}_+device-${DEVICES}_FFTWONGPU-${FFTWONGPU}_${SPLITTILES}"
	# generate the command to run
	cmd="./namd2 +p ${NUM_CORES} +pemap ${PEMAP} ${SPLITTILES} +devices ${DEVICES} +platform 'Intel(R) Level-Zero' ${WORKLOAD}/${WORKLOAD}_nve_cuda.namd"
else
	mkdir -p logs/mpirun-${WORKLOAD}
	log_prefix="logs/mpirun-${WORKLOAD}/${WORKLOAD}_perhost-${PERHOST}_+ppn-${NUM_CORES}_+pemap-${PEMAP}_+commap-${COMMAP}_+device-${DEVICES}"
	# generate the command to run
	cmd="mpirun -perhost ${PERHOST} ./namd2 +ppn ${PPN} +pemap ${PEMAP} +commap ${COMMAP} +devices ${DEVICES} +platform 'Intel(R) Level-Zero' ${WORKLOAD}/${WORKLOAD}_nve_cuda.namd"
fi

# run
log_files=()
echo "# Start to run" > $log_prefix.cmd
for n in $(seq ${ITERATION}); do
	echo "# RUN: $n " |tee -a $log_prefix.cmd
	log_file="${log_prefix}_run-${n}.log"
	echo "# log file: $log_file"
	echo "$cmd 2>&1 |tee $log_file" |tee -a $log_prefix.cmd	
	if [ "${DRYRUN}" == "0" ]; then
		eval "$cmd" 2>&1 |tee $log_file
		sleep 5
	fi
	log_files+=($log_file)
done

# parse the logs to report performance
perfs=()
for log in ${log_files[@]}; do
	echo "===== Parse log ====="
        echo "$log"
	python3 ns_per_day.py $log 2>&1 |tee ${log}.perf
	perf=$( cat ${log}.perf |grep "Nanoseconds per day" |awk -F: '{print $2}' )
	perf=$( echo $perf )
	perfs+=($perf)
	echo ""
done

# choose the best performance from multiple runs
IFS=$'\n' perfs_sorted=($(sort -n <<< "${perfs[*]}")); unset IFS
perfs_num=${#perfs_sorted[*]}
perf_best=${perfs_sorted[$perfs_num-1]}
perf_avg=$( echo ${perfs_sorted[@]} | awk '{s=0; for (i=1;i<=NF;i++)s+=$i; print s/NF;}' )
echo "==========================="
echo "Best Nanoseconds per day from $ITERATION runs: $perf_best"
echo "Average Nanoseconds per day from $ITERATION runs: $perf_avg"

# report best perf to csv file
csv_file="namd_perf.csv"
if ! test -f $csv_file; then
	echo "Workload, Best Performance (ns/day), Avg Performance (ns/day), #Tiles, #GPU, FFTWONGPU, MPIRUN, #CORES PER PROCESS, Pemap, Commap, CMD" > $csv_file
fi
echo "${WORKLOAD},${perf_best},${perf_avg},${NUM_DEVICES},${NUM_GPUS},${FFTWONGPU},${MPIRUN},${NUM_CORES},\" ${PEMAP}\",\" ${COMMAP}\",\"$cmd\"" >> $csv_file



