#/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

WEKA_CREDS=admin:adminPassword STORAGE_CONTROLLER_CONFIG_FILE=$1 timeout 30 $2 2> out.txt & 
server_pid=$!

timeout 5 tail -f out.txt | grep -q "Registered new cluster"

$3 run $4; exit_1=$?
$3 run $5; exit_2=$?

kill $server_pid

! (( $exit_1 || $exit_2 ))
