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
