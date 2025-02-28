#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
LMP_ROOT=${SCRIPT_DIR}/apps/lammps
LOG_DIR=${SCRIPT_DIR}/logs

mkdir -p ${LOG_DIR}
cd ${SCRIPT_DIR}/apps/TEST
export PATH=${SCRIPT_DIR}/apps/cal/build:${LMP_ROOT}/src:$PATH
source ${ONEAPI_ROOT}/setvars.sh
dry_run=${DRYRUN:-0}
LMP_MPI_RANKS=${NUMBER_OF_PROCESS:-16}
LMP_MPI_PERNODE=${PROCESS_PER_NODE:-${LMP_MPI_RANKS}}
LMP_OMP_THREADS=${OMP_NUM_THREADS:-1}
LMP_TILES=${TILES:-1}
LMP_GPUS=${LMP_TILES}
TESTCASE=${TESTCASE:-lc}

# Query host resources
LOGIC_CORES=`lscpu |grep "CPU(s):" |awk -F: '{print$2}' |awk 'NR==1{print$1}'|tr -d ' '`
PHYSICAL_CORES=`lscpu |grep "Core(s) per socket:" |awk -F: '{print$2}'|tr -d ' '`
SOCKETS=`lscpu |grep "Socket(s):" |awk -F: '{print$2}'|tr -d ' '`

GPUs=`clinfo -l | grep "GPU Max" |awk  '{print  $NF}'`
GPU_MODEL=`clinfo -l | grep "GPU Max" |awk  'NR==1 {print  $NF}'`
GPU_NUM=`clinfo -l | grep "GPU Max" |wc -l`

# Each Tile as one GPU
if [ "$GPU_MODEL" == "1550" ];then
    GPU_NUM=$(( GPU_NUM * 2 ))
    LMP_GPUS=$(( LMP_TILES / 2 ))
    if [ "$LMP_GPUS" == "0" ]; then
	LMP_GPUS=1
    fi
fi
echo "GPU_MODEL=$GPU_MODEL GPU_NUM=$GPU_NUM"
echo "LMP_GPUS=$LMP_GPUS LMP_TILES=$LMP_TILES"
echo "LMP_MPI_RANKS=$LMP_MPI_RANKS LMP_MPI_PERNODES=$LMP_MPI_RANKS"
echo "TESTCASE=$TESTCASE"
echo "Current folder: `pwd`"
echo "PATH=$PATH"
echo "LMP_ROOT=$LMP_ROOT"

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

# Restart file generation ,one-time requirement
function create_restart_file() {
    echo "Creating restart file...."
    export UCX_TLS=ud,sm,self
    RESTART_FILE=./restart.lc
    if [ -f "${RESTART_FILE}" ]; then
        echo "Restart file ${RESTART_FILE} exist"
        return 0
    else
        mpirun --bootstrap ssh -np $PHYSICAL_CORES lmp_oneapi -in in.lc_generate_restart -log none 1>/dev/null
    fi
    res "Creating restart file"
}

function print_results() {
    #echo "Results: Matom-step/s"
    echo "---------------------------------------------------------"
    #grep Perf $LOGFILE | awk 'n%2==1{c=NF-1; print $1,$c," Matom-step/s"}{n++}'
    grep Perf $LOGFILE | awk 'n%2==1{print $0}{n++}'
    echo "---------------------------------------------------------"
}


# Running workload benchmark
function run_benchmark() {
    LMP_ARGS="-v N off -pk gpu ${LMP_TILES} -sf gpu -log none"

    export I_MPI_PIN_DOMAIN=auto:compact
    export I_MPI_FABRICS=shm
    export KMP_AFFINITY="granularity=core,scatter"
    export KMP_BLOCKTIME=1000

    export NEOReadDebugKeys=1
    export DirectSubmissionRelaxedOrdering=1
    export DirectSubmissionRelaxedOrderingForBcs=1
    export LMP_ROOT=$LMP_ROOT

    for workload in ${TESTCASE}; do
      if [ $workload = "rhodo" ]; then
          LMP_ARGS="${LMP_ARGS} -v d 0"
      fi
      LOGFILE=$LOG_DIR/$workload.${LMP_GPUS}C${LMP_TILES}T.np${LMP_MPI_RANKS}.ppn${LMP_MPI_PERNODE}.omp${LMP_OMP_THREADS}.log
      cmd="calrun mpirun -np $LMP_MPI_RANKS lmp_oneapi -in in.intel.$workload $LMP_ARGS -screen $LOGFILE"
      if (( $dry_run )); then
	  echo $cmd
      else
	  echo $cmd
	  eval "$cmd"
      fi
      echo "in.intel.$workload in $LMP_MPI_RANKS ranks with $OMP_NUM_THREADS threads on $LMP_TILES gpu tiles benchmark finish"
      print_results
    done
}

create_restart_file

run_benchmark

