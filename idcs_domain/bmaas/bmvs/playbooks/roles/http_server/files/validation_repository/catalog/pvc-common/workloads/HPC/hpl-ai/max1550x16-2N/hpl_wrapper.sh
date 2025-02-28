#!/bin/bash
ulimit -s unlimited
[ -n "${OMPI_COMM_WORLD_RANK}" ] && PMI_RANK=${OMPI_COMM_WORLD_RANK} && MPI_LOCALRANKID=${OMPI_COMM_WORLD_LOCAL_RANK}

#Topology assuming:
#NUMA node0: CPU[0-47],GPU[0-3],IB[mlx5_0,mlx5_1,mlx5_2,mlx5_5]
#NUMA node1: CPU[48-95],GPU[4-7],IB[mlx5_6,mlx5_7,mlx5_8,mlx5_11]

i=$[MPI_LOCALRANKID%4]
c=$[i*24]
export HPL_DEVICE=:$[i*2].0,:$[i*2].1,:$[i*2+1].0,:$[i*2+1].1
export HPL_HOST_CORE=$c-$[c+23]
case $i in
  0)
    [ "X${I_MPI_OFI_PROVIDER}X" == "Xpsm3X" ] && export PSM3_NIC=mlx5_0,mlx5_1
    [ "X${I_MPI_OFI_PROVIDER}X" == "XmlxX" ] && export UCX_NET_DEVICES="mlx5_0:1,mlx5_1:1" && export UCX_MAX_RNDV_RAILS=2
    ;;
  1)
    [ "X${I_MPI_OFI_PROVIDER}X" == "Xpsm3X" ] && export PSM3_NIC=mlx5_2,mlx5_5
    [ "X${I_MPI_OFI_PROVIDER}X" == "XmlxX" ] && export UCX_NET_DEVICES="mlx5_2:1,mlx5_5:1" && export UCX_MAX_RNDV_RAILS=2
    ;;
  2)
    [ "X${I_MPI_OFI_PROVIDER}X" == "Xpsm3X" ] && export PSM3_NIC=mlx5_6,mlx5_7
    [ "X${I_MPI_OFI_PROVIDER}X" == "XmlxX" ] && export UCX_NET_DEVICES="mlx5_6:1,mlx5_7:1" && export UCX_MAX_RNDV_RAILS=2
    ;;
  3)
    [ "X${I_MPI_OFI_PROVIDER}X" == "Xpsm3X" ] && export PSM3_NIC=mlx5_8,mlx5_11
    [ "X${I_MPI_OFI_PROVIDER}X" == "XmlxX" ] && export UCX_NET_DEVICES="mlx5_8:1,mlx5_11:1" && export UCX_MAX_RNDV_RAILS=2
    ;;
esac
[ "X${I_MPI_OFI_PROVIDER}X" == "Xpsm3X" ] && NIC=$PSM3_NIC
[ "X${I_MPI_OFI_PROVIDER}X" == "XmlxX" ] && NIC=$UCX_NET_DEVICES

cmd="numactl -l $@"
echo "HOST=$(hostname), RANK=${PMI_RANK}, LOCALRANK=${MPI_LOCALRANKID}, HPL_DEVICE=${HPL_DEVICE}, HPL_HOST_CORE=${HPL_HOST_CORE}, NIC=${NIC}, CMD=${cmd}, PID=$$ AFF=$(taskset -pc $$)"
eval ${cmd}
