#!/bin/bash -e
set -x
echo "starting bm-virtual-sc cluster validation"
#Insert bm-virtual specific validation commands here
# uncomment this to simulate a failure on device-2
# echo "failedNodes=device-2" >> /tmp/validation_result.meta
cat /tmp/validation_result.meta
printenv
sleep 60
echo "end of bm-virtual-sc cluster validation"
