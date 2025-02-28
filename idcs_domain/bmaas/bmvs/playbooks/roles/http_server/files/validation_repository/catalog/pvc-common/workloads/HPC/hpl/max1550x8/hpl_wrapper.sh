#!/bin/bash
ulimit -s unlimited
[ -n "${OMPI_COMM_WORLD_RANK}" ] && PMI_RANK=${OMPI_COMM_WORLD_RANK} && MPI_LOCALRANKID=${OMPI_COMM_WORLD_LOCAL_RANK}

#Topology assuming:
#NUMA node0: CPU[0-47],GPU[0-3]
#NUMA node1: CPU[48-95],GPU[4-7]

#4 Ranks per node, each rank is running with 4 GPU tiles, and 24 CPU cores
i=$[MPI_LOCALRANKID%4]
c=$[i*24]
export HPL_DEVICE=:$[i*2].0,:$[i*2].1,:$[i*2+1].0,:$[i*2+1].1
export HPL_HOST_CORE=$c-$[c+23]

cmd="numactl -l $@"
echo "HOST=$(hostname), RANK=${PMI_RANK}, LOCALRANK=${MPI_LOCALRANKID}, HPL_DEVICE=${HPL_DEVICE}, HPL_HOST_CORE=${HPL_HOST_CORE}, CMD=${cmd}, PID=$$ AFF=$(taskset -pc $$)"
eval ${cmd}