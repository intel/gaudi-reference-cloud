#!/bin/bash
logfile=01_test_sanity.log
bash ./sys_sanity_check.sh 2>&1 |tee $logfile

testwarning=$( grep WARNING $logfile )
testerror=$( grep ERROR $logfile )

if [ ! "$testerror" == "" ]; then	
	echo "$testerror"
	echo "Sanity Check Failed with ERROR"
	exit -2
fi


if [ ! "$testwarning" == "" ]; then	
	echo "$testwarning"
	echo "Sanity Check Failed with WARNING"
	exit -1
fi

echo "Sanity Check PASSED"


