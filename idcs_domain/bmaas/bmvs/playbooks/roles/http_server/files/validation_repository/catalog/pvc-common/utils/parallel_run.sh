#!/bin/bash

help(){
	echo "The script to trigger parallel execution for same command running on different GPU"
	echo "options: -c | --cmd <command to run in parallel>"
	echo "       : -d | --device <device id to start run, default 0>"
	echo "       : -p | --parallel <number of instances running in parallel, default 1>"
	echo "       : -n | --numa <bind numa node close to GPU device, default 0>"
	echo "       : -h | --help"
	echo "Example: $0 -c GEMM/dgemm.mkl -d 0 -p 2"
}
#export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
DEVICE_HIERARCHY=${ZE_FLAT_DEVICE_HIERARCHY}

XPUM=${XPUM:-xpu-smi}
if [ ! -e "`which ${XPUM}`" ]; then
	echo "${XPUM} not found, switch to xpumcli"
	XPUM=xpumcli
	if [ ! -e "`which ${XPUM}`" ]; then
		echo "Can't find XPUManager installed! Please install xpumanager or xpu-smi."
		exit -1
	fi
fi

cmd="echo 'use -c option to specify the command and -h for help'"
device=0
parallel=1
numa=0

options=c:,d:,p:,n,h
optionl=cmd:,device:,parallel:,numa,help
OPTS=$(getopt -a -n $0 --options $options --longoptions $optionl -- "$@")
eval set -- "$OPTS"
while :
do
  case "$1" in
      -c | --cmd )
         cmd="$2"
         shift 2
         ;;
      -d | --device )
	 device="$2"
	 shift 2
	 ;;
      -p | --parallel )
	 parallel="$2"
	 shift 2
	 ;;
      -n | --numa )
	 numa=1
	 shift 1
	 ;;
      -h | --help)
         help
         exit 0
         ;;
      --)
         shift;
         break
         ;;
      *)
         echo "Unexpected option: $1"
         ;;
  esac
done

tiles=1
#Max 1550 device id 0bd5
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
	tiles=2
fi


#ONEAPI_DEVICE_SELECTOR=level_zero:x.y
#ZE_AFFINITY_MASK=x.y
IMPLICIT_SCALING=${IMPLICIT_SCALING:-0}
cmds=()
if [ "$tiles" == "1" ] || [ "${IMPLICIT_SCALING}" == "1" ]; then
	for i in $(seq ${device} $((${device}+${parallel}-1)) ); do
		if [ "$numa" == "1" ]; then
		    bdf=$( lspci |grep Display |head -$(( i+1 )) |tail -1 |awk '{print $1}' )
		    numanode=$(lspci -s $bdf -vv |grep NUMA |awk -F: '{print $2}')
		    numacmd="numactl -N $numanode -m $numanode "
		fi

		cmds+="\"ZE_AFFINITY_MASK=${i} ${numacmd} ${cmd}\" "
	done
elif [ "$tiles" == "2" ]; then
	if [ ${DEVICE_HIERARCHY} == "COMPOSITE" ]; then
		nt=0
		nd=${device}
		for i in $(seq ${device} $((${device}+${parallel}-1)) ); do
			if [ "${nt}" == "0" ]; then
				d=${nd}
				t=0
				nt=1
			else
				d=${nd}
				t=1
				nt=0
				nd=$(( $nd +1))
			fi
			if [ "$numa" == "1" ]; then
				bdf=$( lspci |grep Display |head -$(( d+1 )) |tail -1 |awk '{print $1}' )
				numanode=$(lspci -s $bdf -vv |grep NUMA |awk -F: '{print $2}')
				numacmd="numactl -N $numanode -m $numanode "
			fi
			cmds+="\"ZE_AFFINITY_MASK=${d}.${t} ${numacmd} ${cmd}\" "
		done
	else
		for i in $(seq $(( device*2 )) $(( device*2 + parallel -1 )) ); do
			cmds+="\"ZE_AFFINITY_MASK=${i} ${numacmd} ${cmd}\" "
		done
	fi
fi

#echo ${cmds[@]}

torun="parallel --lb -d, --tagstring \"[{#}]\" ::: ${cmds[@]}"
echo $torun
eval " $torun"


