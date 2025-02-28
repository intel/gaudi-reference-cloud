#!/bin/bash

echo "starting bm-spr-pvc-1100-8 validation"
date
sleep 60
echo "version 0.0.1"
date
# Check xe_link_calibration_date
calibrationCheck=$(for i in {0..9}; do sudo xpu-smi discovery -d $i -j | grep xe_link_calibration_date; done)

# Check if any GPU is "Not Calibrated"
if echo "$calibrationCheck" | grep -q "Not Calibrated"; then
    # Print GPU containing "Not Calibrated"
    echo "Issues found with some GPUs."
    echo "$calibrationCheck" | grep "Not Calibrated"

    # Ensure zip file is present
    if [[ ! -f "/tmp/validation/787990_Intel_Xe_Link_Calibration_Rev1.zip" ]]; then
    echo "Calibration zip file not found"
    exit 1
    fi

    original_dir=$(pwd)
    # Unzip and run calibration steps
    unzip -d /tmp/validation/ /tmp/validation/787990_Intel_Xe_Link_Calibration_Rev1.zip
    # cd inside the calibration folder and perform necessary steps
    cd /tmp/validation/787990_Intel_Xe_Link_Calibration

    sudo python3 ./margin_data.py & 
    # Killing the command process after 2 minutes due to an existing issue
    pid=$! 
    time_duration=120
    sleep $time_duration

    if ps -p $pid > /dev/null; then
    # Kill the background command
    kill $pid
    fi

    # Continue with next steps
    sudo python3 ./txcal_blob.py margin_summary.csv
    sudo python3 ./update_blob.py txcal.txbin --keep-going

    cd $original_dir
    calibrationOutput=$(for i in {0..9}; do sudo xpu-smi discovery -d $i -j | grep xe_link_calibration_date; done)

    # Check if any GPU is "Not Calibrated"
    if echo "$calibrationOutput" | grep -q "Not Calibrated"; then
        # Print GPU containing "Not Calibrated"
        echo "Issues found. Calibration not successful."
        echo "$calibrationOutput" | grep "Not Calibrated"
        exit 1
    fi
else
    echo "Skipping Calibration as GPUs have right calibration date"
fi

# Attempt to install jq if it does not exist.
if command -v apt-get; then
    sudo apt-get install jq -y
elif command -v zypper; then
    sudo zypper -n install jq
    sudo zypper -n install bc
else
    echo "Could not determine a package manager to install pre-requirements"
    exit 1
fi

echo "executing the validation steps after checking/installing missing packages"

cpuCount=$(xpu-smi discovery -j | jq '.[] | length')

# Check if CPU count is 8
if [ "$cpuCount" -eq 8 ]; then
    echo "CPU count is 8"
else
    echo "CPU count from xpu-smi command is not 8"
    exit 1
fi

echo "starting system validate workload for bm-spr-pvc-1100-8"

# Invoke the system sanity check
/tmp/validation/system-validate/sys_sanity_check.sh
if [ $? -ne 0 ]; then
    echo "Sanity check test failed with exit code 1. Exiting start.sh"
    exit 1
fi

# Skip gemm tests for opensuse
if command -v apt-get; then
  # Validate GPU System with Micro Workload Benchmarks
  # Invoke GEMM for GPU Computation
  /tmp/validation/system-validate/sys_val_gemm.sh
  if [ $? -ne 0 ]; then
    echo "GEMM benchmark test failed with exit code 1. Exiting start.sh"
    exit 1
  fi
fi

echo "completed system validate workload for bm-spr-pvc-1100-8"
echo "completed bm-spr-pvc-1100-8 validation"
