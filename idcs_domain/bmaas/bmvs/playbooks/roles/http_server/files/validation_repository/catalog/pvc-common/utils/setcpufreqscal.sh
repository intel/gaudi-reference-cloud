#!/bin/bash
# Use script to change the cpu scaling governor

action=$1
if [ -z $action ]; then
	echo "cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor "
	cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
	echo "use \"$0 performance\" to set cpu scaling governor to performance"
	echo "use \"$0 powersave\" to set cpu scaling governor to powersave"
	
	exit 0
fi

if [ "$action" == "performance" ]; then
	echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
elif [ "$action" == "powersave" ]; then
	echo "powersave" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
else
	echo "unknow action"
	exit 1
fi

cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor 


