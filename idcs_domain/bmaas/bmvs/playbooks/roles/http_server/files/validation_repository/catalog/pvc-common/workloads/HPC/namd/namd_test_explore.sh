#!/bin/bash

# Scripts to run multiple namd tests based on different configs
# Update based on demand

<<'COMMENTS'
explore_num_cores=(48 64 68 72)
explore_fftw=(0 1)
explore_workload=(apoa1 stmv)
explore_devices=(2 4 6)

for workload in ${explore_workload[@]}; do
	for fftw in ${explore_fftw[@]}; do
		for devices in ${explore_devices[@]}; do
			for cores in ${explore_num_cores[@]}; do
				cmd="DRYRUN=1 NUM_CORES=$cores NUM_DEVICES=$devices FFTWONGPU=$fftw WORKLOAD=$workload bash namd_test.sh"
				echo "TORUN: $cmd"
				eval "$cmd"
			done
		done
	done
done

COMMENTS


explore_workload=(apoa1 stmv)
explore_devices=(2 4 6 8 16)

for workload in ${explore_workload[@]}; do
		for devices in ${explore_devices[@]}; do
			case $devices in
			2)
				explore_num_cores=(20 24 32 48)
				;;
			4)
				explore_num_cores=(12 16 20 24)
				;;
			6)
				explore_num_cores=(10 12 14 16)
				;;
			8)
				explore_num_cores=(6 8 10 12)
				;;
			16)
				explore_num_cores=(4 5 6)
				;;
			*)
				;;
			esac

			for cores in ${explore_num_cores[@]}; do
				cmd="DRYRUN=0 WORKLOAD=$workload MPIRUN=1 NUM_CORES=$cores NUM_DEVICES=$devices bash namd_test.sh"
				echo "TORUN: $cmd"
				#eval "$cmd"
			done
		done
done
