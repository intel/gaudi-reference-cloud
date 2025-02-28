#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
#source ${ONEAPI_ROOT}/setvars.sh
source ${ONEAPI_ROOT}/compiler/latest/env/vars.sh
source ${ONEAPI_ROOT}/mkl/latest/env/vars.sh
source ${ONEAPI_ROOT}/dnnl/latest/env/vars.sh
source ${ONEAPI_ROOT}/mpi/latest/env/vars.sh

export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

LOG_DIR=${WORKSPACE}/bench_hpl_ai_logs
WORK_DIR=${WORKSPACE}/bench_hpl_ai_workdir
mkdir -p ${LOG_DIR}
mkdir -p ${WORK_DIR}
cd ${WORK_DIR}
export PATH=$PATH:${ONEAPI_ROOT}/mkl/latest/share/mkl/benchmarks/mp_linpack/:${WORK_DIR}/benchmarks_2024.0/linux/mkl/benchmarks/mp_linpack/
echo "Current PATH Environments:"
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
   gpu_memory=64
else
   tiles=1
   gpu_memory=48
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

hpl_magic_1100=120000
hpl_magic_1550=215000

# order with GPU number 1 2 3 4 5 6 7 8, data with 1 2 4 8
hpl_scale_magic_1100_x8=(100000 0 0 0 0 0 0 750000)
hpl_scale_magic_1550_x4=(215000 0 0 0 0 0 0 0)
hpl_scale_magic_1550_x8=(215000 0 0 0 0 0 0 1420000)


threshold=0.05


get_hpl_benchmarks(){
    print_prefix "Check HPL binary"
    hplbench_bin="xhpl-ai_intel64_dynamic_gpu"
    if [ ! -e "`which $hplbench_bin`" ]; then
	    mkdir -p $WORK_DIR; cd $WORK_DIR
	    echo "Can't find xhpl-ai_intel64_dynamic_gpu binary for execution"
	    #echo "trying to download: https://internal-placeholder.com/781888/l_onemklbench_p_2023.2.0_49340.tgz"
	    #wget https://internal-placeholder.com/781888/l_onemklbench_p_2023.2.0_49340.tgz
	    #tar xf l_onemklbench_p_2023.2.0_49340.tgz
	    #export PATH=$PATH:${WORK_DIR}/benchmarks_2023.2.0/linux/mkl/benchmarks/mp_linpack/
		
		echo "trying to download: https://internal-placeholder.com/793598/l_onemklbench_p_2024.0.0_49515.tgz"
		wget https://internal-placeholder.com/793598/l_onemklbench_p_2024.0.0_49515.tgz
		tar xf l_onemklbench_p_2024.0.0_49515.tgz
		export PATH=$PATH:${WORK_DIR}/benchmarks_2024.0/linux/share/mkl/benchmarks/mp_linpack/
	    cd ..
    fi

    if [ ! -e "`which $hplbench_bin`" ]; then
	    echo "ERROR: Can't find hpl binary to run"
	    exit -1
    fi
    echo "HPL binary found:"    
    echo $( which $hplbench_bin)
    cp $( which $hplbench_bin) $LOG_DIR/
    print_aftfix

}

create_numa_gpu_map(){
   print_prefix "Check GPU topology and NUMA"
   numa_gpu_map=()
   gpu_numa_map=()

#   for gpu in $(seq 0 $(($num_gpu -1))); do
#       bdf=$(${XPUM} discovery -d ${gpu} |grep BDF | awk -F'0000:' '{print $2}' |awk -F' ' '{print $1}')
#       numanode=$(lspci -s $bdf -vv |grep NUMA |awk -F: '{print $2}')
#       numanode=$(echo $numanode)
#       numa_gpu_map[$numanode]="${numa_gpu_map[$numanode]} $gpu"
#       gpu_numa_map[$gpu]=$numanode
#       echo "GPU $gpu connected with NUMA node ${numanode}"
#   done
   gpu=0
   while read line; do
       bdf=$( echo $line |awk '{print $1}' )
       numanode=$(lspci -s $bdf -vv |grep NUMA |awk -F: '{print $2}')
       numanode=$(echo $numanode)
       numa_gpu_map[$numanode]="${numa_gpu_map[$numanode]} $gpu"
       gpu_numa_map[$gpu]=$numanode
       echo "GPU $gpu connected with NUMA node ${numanode}"
       gpu=$(( gpu+1 ))
   done <<< $( lspci |grep Display )

   print_aftfix
}

create_hpl_dat_file(){
cat > HPL.dat <<- EOF
HPLinpack benchmark input file
Innovative Computing Laboratory, University of Tennessee
HPL.out      output file name (if any)
6            device out (6=stdout,7=stderr,file)
1            # of problems sizes (N)
${N} 192000  Ns
1            # of NBs
${NB} 192 384 576 640 768       NBs
1            PMAP process mapping (0=Row-,1=Column-major)
1            # of process grids (P x Q)
${P}         Ps
${Q}         Qs
16.0         threshold
1            # of panel fact
2 1 0        PFACTs (0=left, 1=Crout, 2=Right)
1            # of recursive stopping criterium
2            NBMINs (>= 1)
1            # of panels in recursion
2            NDIVs
1            # of recursive panel fact.
1 0 2        RFACTs (0=left, 1=Crout, 2=Right)
1            # of broadcast
0            BCASTs (0=1rg,1=1rM,2=2rg,3=2rM,4=Lng,5=LnM)
1            # of lookahead depth
0            DEPTHs (>=0)
0            SWAP (0=bin-exch,1=long,2=mix)
1            swapping threshold
1            L1 in (0=transposed,1=no-transposed) form
1            U  in (0=transposed,1=no-transposed) form
0            Equilibration (0=no,1=yes)
8            memory alignment in double (> 0)
EOF

}

validate_hpl_bench_result(){
	hpl_magic=0
	if [ "$tiles" == 1 ]; then
	    if [ "$num_gpu" == "8" ]; then
		hpl_magic=${hpl_scale_magic_1100_x8[0]}
	    else
		echo "System config not validated"
		hpl_magic=0
	    fi
	elif [ "$tiles" == 2 ]; then
	    if [ "$num_gpu" == "4" ]; then
		hpl_magic=${hpl_scale_magic_1550_x4[0]}
	    elif [ "$num_gpu" == "8" ]; then
		hpl_magic=${hpl_scale_magic_1550_x8[0]}
	    else
		hpl_magic=0
	    fi
	fi
	local val_func=$( grep PASSED $log_file )
	if [ "${val_func}x" == "x"  ];then
		val_result="FAILED"
	else
	  local benchval=$(grep WC $log_file|awk '{print int($NF)}')
	  local benchref=$hpl_magic
	  local val=`bc -l <<< "$benchval >= $benchref*(1-$threshold)"`
	  local bench_ratio="N/A"
	  if [ "$benchref" != "0" ]; then
  	    bench_ratio=`bc -l <<< "$benchval / $benchref * 100"`
	  fi
          if (( $val )); then
	    val_result="PASSED"
          else
	    val_result="WARNING"
	  fi
	fi
	printf "DATAVALIDATION HPL-AI benchmark on GPU %d: %-10s: current=%s(GFLOPS), ref=%s, ratio=%.2f%%\n" $gpu $val_result $benchval $benchref $bench_ratio
	# report to a csv file
	csv_file="hpl_ai_test.csv"
	if ! test -f $csv_file; then
		echo "Workload, GPU, Result, Performance(GFLOPS), Reference, Ratio" > $csv_file
	fi
	echo "HPL-AI,${gpu},${val_result},${benchval},${benchref},${bench_ratio}" >> $csv_file


}

validate_hpl_bench(){
    print_prefix "### Running HPL-AI benchmarking on single GPU in order"
    NNODES=1
    NGPUS=1
    NTILES_PER_GPU=$tiles
    GPUMEM_GB=$gpu_memory
    NB=1152
    P=1
    Q=1
    N=$(echo "(sqrt($NNODES*$NGPUS*$NTILES_PER_GPU*($GPUMEM_GB-6)*1024*1024*1024/8)/($NB*$P*$Q))*($NB*$P*$Q)*2"|bc)

    create_hpl_dat_file

   for numa in ${!numa_gpu_map[@]}; do
       for gpu in ${numa_gpu_map[$numa]}; do
         #echo "Bench HPL on GPU $gpu NUMA node $numa DATAVALIDATION"
         log_file=$LOG_DIR/hpl_gpu_$gpu.log
	 if [ "$tiles" == "1" ]; then
	     hpl_device=":${gpu}.0"
	 else
	     hpl_device=":${gpu}.0,:${gpu}.1"
	 fi
         cmd="HPL_HOST_NODE=$numa HPL_DEVICE=$hpl_device mpirun -np 1 numactl -N $numa -m $numa ${hplbench_bin}"
	 if (( $dry_run )); then
             echo $cmd
	 else
	     cp HPL.dat ${log_file}.HPL.dat
	     echo $cmd >$log_file
             eval "$cmd" 2>&1 >>$log_file
	 fi
	 validate_hpl_bench_result
	done
   done 
   print_aftfix    
}

create_wrapper_for_mp(){
cat > hpl_wrapper.sh <<- 'EOF'
#!/bin/bash
ulimit -s unlimited
[ -n "${OMPI_COMM_WORLD_RANK}" ] && PMI_RANK=${OMPI_COMM_WORLD_RANK} && MPI_LOCALRANKID=${OMPI_COMM_WORLD_LOCAL_RANK}
hplbench_bin="xhpl-ai_intel64_dynamic_gpu"

if [ $[PMI_RANK%2] -eq 0 ]; then
  export HPL_DEVICE=$1
  export HPL_HOST_NODE=0
fi
if [ $[PMI_RANK%2] -eq 1 ]; then
  export HPL_DEVICE=$2
  export HPL_HOST_NODE=1
fi
cmd="numactl -l $hplbench_bin"
echo "HOST=$(hostname), RANK=${PMI_RANK}, LOCALRANK=${MPI_LOCALRANKID}, HPL_DEVICE=${HPL_DEVICE}, HPL_HOST_NODE=${HPL_HOST_NODE}, CMD=${cmd}, PID=$$"
eval ${cmd}
EOF
}


validate_hpl_scale_bench(){
   print_prefix "Running HPL-AI scalability benchmarking"
   #for HPL-AI, run only for 1 GPU and all GPUs as CPU cores impacts perf a lot
   scale_gpu_num=(1 $num_gpu)
   for i in ${scale_gpu_num[@]}; do
     if [ ! "${GPUNUM}" == "" ] && [ ! "$i" == "${GPUNUM}" ]; then continue; fi
   	 if (( $( echo "$num_gpu >= $i" | bc -l ) )); then
		echo "### Evaluating HPL-AI performance on $i GPU"
		numa_gpu_map0=(${numa_gpu_map[0]})
		if [ "${#numa_gpu_map0[@]}" -ge "$i" ]; then
	    numa=${gpu_numa_map[0]}

    	NNODES=1
	    NGPUS=$i
	    NTILES_PER_GPU=$tiles
	    GPUMEM_GB=$gpu_memory
	    NB=1152
	    P=1
	    Q=1
	    N=$(echo "(sqrt($NNODES*$NGPUS*$NTILES_PER_GPU*($GPUMEM_GB-6)*1024*1024*1024/8)/($NB*$P*$Q))*($NB*$P*$Q)*2"|bc)

	    create_hpl_dat_file

	    log_file=$LOG_DIR/hpl_gpu_scale_$i.log
	    hpl_device=""
	    for gpu in $(seq 0 $(( $i-1 ))); do
                if [ "$tiles" == "1" ]; then
		    if [ "$hpl_device" == "" ]; then
                    	hpl_device=":${gpu}.0"
		    else
			hpl_device="${hpl_device},:${gpu}.0"
		    fi
	        else
		    if [ "$hpl_device" == "" ]; then
			hpl_device=":${gpu}.0,:${gpu}.1"
		    else
	                hpl_device="${hpl_device},:${gpu}.0,:${gpu}.1"
		    fi
                fi
	    done
	    cmd="HPL_HOST_NODE=$numa HPL_DEVICE=$hpl_device mpirun -np 1 numactl -N $numa -m $numa ${hplbench_bin}"

		else

        NNODES=1
	    NGPUS=$i
        NTILES_PER_GPU=$tiles
        GPUMEM_GB=$gpu_memory
	    NB=1152
	    P=2
	    Q=1
	    N=$(echo "(sqrt($NNODES*$NGPUS*$NTILES_PER_GPU*($GPUMEM_GB-6)*1024*1024*1024/8)/($NB*$P*$Q))*($NB*$P*$Q)*2"|bc)

	    create_hpl_dat_file

	    log_file=$LOG_DIR/hpl_gpu_scale_$i.log
	    hpl_device_rank=()
	    for gpu in $(seq 0 $(( $i-1 ))); do
			numa_node=${gpu_numa_map[$gpu]}
            if [ "$tiles" == "1" ]; then
		    	if [ "${hpl_device_rank[$numa_node]}x" == "x" ]; then
                   	hpl_device_rank[$numa_node]=":${gpu}.0"
		    	else
					hpl_device_rank[$numa_node]="${hpl_device_rank[$numa_node]},:${gpu}.0"
		    	fi
	        else
		    	if [ "${hpl_device_rank[$numa_node]}x" == "x" ]; then
					hpl_device_rank[$numa_node]=":${gpu}.0,:${gpu}.1"
		    	else
	               	hpl_device_rank[$numa_node]="${hpl_device_rank[$numa_node]},:${gpu}.0,:${gpu}.1"
		    	fi
            fi
	    done

	    create_wrapper_for_mp 
	    cmd="mpirun -np 2 bash ./hpl_wrapper.sh ${hpl_device_rank[@]}"
	    cp hpl_wrapper.sh ${log_file}.hpl_wrapper

		fi
	
		if (( $dry_run )); then
	    echo $cmd
		else
    	cp HPL.dat ${log_file}.HPL.dat
	    echo $cmd >$log_file
	    eval "$cmd" 2>&1 >>$log_file
		fi
    	val_func=$( grep PASSED $log_file )
    	if [ "${val_func}x" == "x" ]; then
        echo "HPL-AI runtime failed: FUNCFAILED"
    	else
	  	perf=$(grep WC $log_file|awk '{print int($NF)}')
	   	perf_val[$( bc -l <<< "$i -1")]=$perf
	   	echo "HPL-AI performance evaluated: $perf GFLOPS"
		fi
     fi
   done
   print_aftfix

   print_prefix "HPL-AI scalability benchmarking with DATAVALIDATION"
   
   hpl_scale_magic=()
   if [ "$tiles" == 1 ]; then
       if [ "$num_gpu" == "4" ]; then
        	hpl_scale_magic=(${hpl_scale_magic_1100_x4[@]})
       elif [ "$num_gpu" == "8" ]; then
	   		hpl_scale_magic=(${hpl_scale_magic_1100_x8[@]})
       else
            hpl_scale_magic=(0 0 0 0 0 0 0 0)
       fi
   elif [ "$tiles" == 2 ]; then
       if [ "$num_gpu" == "4" ]; then
        	hpl_scale_magic=(${hpl_scale_magic_1550_x4[@]})
       elif [ "$num_gpu" == "8" ]; then
	   		hpl_scale_magic=(${hpl_scale_magic_1550_x8[@]})
       else
        	hpl_scale_magic=(0 0 0 0 0 0 0 0)
       fi
   fi

   val_base=${perf_val[0]}
   for i in $( seq 0 7); do
       val=${perf_val[$i]}
       if [ -z $val ];then
           val=0
       fi
       vali=$(bc -l <<< "$val != 0")
       if (( $vali )); then
            perf_data=$(printf '%.1f' ${perf_val[$i]})
	   		if [ "${val_base}x" == "x" ]; then
	       		echo "Baseline data on 1 GPU empty, the scalability will not be calculated"
	   		else
	       scale=$(printf '%.2f' $(bc -l <<< "$perf_data / $val_base / ($i+1) * 100"))
	   		fi
	   		if [ "${hpl_scale_magic[$i]}" == "0" ];then
	       ratio="N/A"
	   		else
           reference_diff=$( bc -l <<<  "${hpl_scale_magic[$i]} *  $threshold" )
           reference=$( bc -l <<< "${hpl_scale_magic[$i]} - $reference_diff" )
           result=$(echo "$perf_data >= $reference" |bc -l)
           ratio=$( bc -l <<< "$perf_data / ${hpl_scale_magic[$i]} * 100")
           ratio=$( printf '%.2f' $ratio )
           threshold_percent=$( bc -l <<< "$threshold * 100")
	   		fi
       		if (( $result )); then
           resultstr=PASSED
       		else
           resultstr=WARNING
       		fi
	  		echo "DATAVALIDATION HPL-AI benchmarking on $((i+1)) GPU (GFLOPS) - $resultstr : current $perf_data, reference ${hpl_scale_magic[$i]}, ratio $ratio%, scalability $scale% "

			# report to a csv file
			csv_file="hpl_ai_test_scale.csv"
			if ! test -f $csv_file; then
				echo "Workload, #GPU, Result, Performance(GFLOPS), Reference, Ratio(%), Scale(%)" > $csv_file
			fi
			echo "HPL-AI,$((i+1)),${resultstr},${perf_data},${hpl_scale_magic[$i]},${ratio},${scale}" >> $csv_file

       fi
   done

   print_aftfix
}

get_hpl_benchmarks
create_numa_gpu_map

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "1" ]; then
  validate_hpl_bench
fi

if [ "$TESTSET" == "0" ] || [ "$TESTSET" == "2" ]; then
  validate_hpl_scale_bench
fi

