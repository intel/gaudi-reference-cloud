#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
TESTCASE_DIR=${TESTCASE_DIR:-${SCRIPT_DIR}/testcases}
TESTCASE=${TESTCASE:-benchMEM}

source ${ONEAPI_ROOT}/setvars.sh
dry_run=${DRYRUN:-0}

APPS_ROOT=${SCRIPT_DIR}/apps
GMX_ROOT=${SCRIPT_DIR}/apps/gromacs
GMX_INSTALL=${APPS_ROOT}/__install
LOG_DIR=${SCRIPT_DIR}/logs
mkdir -p $LOG_DIR

GMX_MPI_RANKS=${NTMPI:-1}
GMX_OMP_THREADS=${NTOMP:-16}


# Query host resources
LOGIC_CORES=`lscpu |grep "CPU(s):" |awk -F: '{print$2}' |awk 'NR==1{print$1}'|tr -d ' '`
PHYSICAL_CORES=`lscpu |grep "Core(s) per socket:" |awk -F: '{print$2}'|tr -d ' '`
SOCKETS=`lscpu |grep "Socket(s):" |awk -F: '{print$2}'|tr -d ' '`

# Query GPU resources
GPUs=`clinfo -l | grep "GPU Max" |awk  '{print  $NF}'`
GPU_MODEL=`clinfo -l | grep "GPU Max" |awk  'NR==1 {print  $NF}'`
GPU_NUM=`clinfo -l | grep "GPU Max" |wc -l`

# Each Tile as one GPU
if [ "$GPU_MODEL" == "1550" ];then
	GPU_NUM=$(( GPU_NUM * 2 ))
fi
echo "GPU_MODEL=$GPU_MODEL GPU_NUM=$GPU_NUM"
fatal_error (){
	echo -e "\033[41;37m Error: $@ \033[0m"
	exit
}

function sanity_check () {
    if [ $GMX_MPI_RANKS -gt 0 ] 2>/dev/null ; then
        echo "mpi ranks number:" $GMX_MPI_RANKS
        if [ $GMX_MPI_RANKS -gt $GPU_NUM  ]; then
            fatal_error "the mpi ranks more than total gpu tiles. MPI ranks=$GMX_MPI_RANKS, GPU num=$GPU_NUM"            
        fi
	total_cores=$(( PHYSICAL_CORES * SOCKETS ))
        if [ $GMX_MPI_RANKS -gt $total_cores ]; then
            fatal_error "the mpi ranks more than physical cores. MPI ranks=$GMX_MPI_RANKS, total cpu cores=$total_cores"
        fi
    else
        fatal_error "invalid mpi ranks from GMX_MPI_RANKS"
    fi

    if [ $GMX_OMP_THREADS -gt 0 ] 2>/dev/null ; then
        echo "openmp threads number:" $GMX_OMP_THREADS
	total_cores=$((PHYSICAL_CORES * SOCKETS))
	total_threads=$((GMX_OMP_THREADS * GMX_MPI_RANKS))
        if [ $total_threads -gt $total_cores ]; then
            fatal_error "The (GMX_OMP_THREADS * GMX_MPI_RANKS) more than physical cores. MPI ranks=$GMX_MPI_RANKS, OMP threads=$GMX_OMP_THREADS, total threads=$total_threads, total cores=$total_cores"
        fi
    else
        fatal_error "Invalid openmp threads from GMX_OMP_THREADS"
    fi
}


# Get every function running status
function res () {
    if [ $? -eq 0 ]
    then
        echo -e "\033[32m $@ sucessed. \033[0m"
    else
        echo -e "\033[41;37m $@ failed. \033[0m"
        exit
    fi
}

# Running workload benchmark
function run_benchmark() {
    if [ "$TESTCASE" == "STMV" ]; then
	TESTCASE=topol
    fi
    workload=$TESTCASE_DIR/$TESTCASE.tpr
    if [ ! -e "$workload" ]; then
	fatal_error "Can't find test case file $workload"
    fi

    #define steps for test case
    if [ "$TESTCASE" == "benchMEM" ]; then
	gmx_args_steps="-nsteps 10000 -resetstep 9000"
    elif [ "$TESTCASE" == "topol" ]; then
	gmx_args_steps="-nsteps 5000 -resetstep 4000"
    elif [ "$TESTCASE" == "benchPEP" ] || [ "$TESTCASE" == "benchPEP-h" ] ; then
	gmx_args_steps="-nsteps 1000 -resetstep 900"
    fi
    if [ $NSTEPS -gt 0 ] 2>/dev/null; then
	gmx_args_steps="-nsteps $NSTEPS"
	if [ $RESETSTEP -gt 0 ] 2>/dev/null; then
		gmx_args_steps="$gmx_args_steps -resetstep $RESETSTEP"
	fi
    fi
    
    GMX_ARGS="$gmx_args_steps -pin on -cpt -1 -resethway -noconfout -nb gpu -pme gpu -notunepme"

    if [ "$TESTCASE" == "PEP-h" ]; then
	GMX_ARGS="$GMX_ARGS -update gpu -bonded gpu"
    fi

    if [ $GMX_MPI_RANKS -gt 1 ] 2>/dev/null ; then
        GMX_ARGS="$GMX_ARGS -npme 1"
    fi
    LOGFILE=${LOG_DIR}/${TESTCASE}.ranks${GMX_MPI_RANKS}.ompt${GMX_OMP_THREADS}.log
    echo "Running Log file $LOGFILE"
    cmd="gmx mdrun $GMX_ARGS -s $workload -ntmpi $GMX_MPI_RANKS -ntomp $GMX_OMP_THREADS ";
    if (( $dry_run )); then
        echo $cmd
    else
	export ONEAPI_DEVICE_SELECTOR=level_zero:gpu
        export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1
        #export SYCL_PI_LEVEL_ZERO_DEVICE_SCOPE_EVENTS=0
        source $GMX_INSTALL/bin/GMXRC
    
        echo "$cmd" 2>&1 |tee $LOGFILE
	echo "---------------------------" 2>&1 |tee -a $LOGFILE
        eval "$cmd" 2>&1 |tee -a $LOGFILE
        echo " $workload benchmark in $GMX_MPI_RANKS ranks with $GMX_OMP_THREADS threads on $GMX_MPI_RANKS gpu tiles finish"
    fi
    
}


sanity_check

run_benchmark

