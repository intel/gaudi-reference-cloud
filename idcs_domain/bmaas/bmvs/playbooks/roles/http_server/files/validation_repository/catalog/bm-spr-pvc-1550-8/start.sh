#!/bin/bash -e

echo "starting bm-spr-pvc-1550-8 validation"
echo "version 0.0.1"
sleep 60
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

echo "starting system validate workload for bm-spr-pvc-1550-8"

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

echo "completed system validate workload for bm-spr-pvc-1550-8"
echo "completed bm-spr-pvc-1550-8 validation"

