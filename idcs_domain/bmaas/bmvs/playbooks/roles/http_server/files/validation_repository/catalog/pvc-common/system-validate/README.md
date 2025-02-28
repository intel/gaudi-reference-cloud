<!-- TOC -->

- [System Validate Workload for Intel Data Center GPU Max Series PVC](#system-validate-workload-for-intel-data-center-gpu-max-series-pvc)
    - [Requirements](#requirements)
    - [System Sanity Check](#system-sanity-check)
    - [Validate GPU System with Micro Workload Benchmarks](#validate-gpu-system-with-micro-workload-benchmarks)
        - [GEMM for GPU Computation](#gemm-for-gpu-computation)
        - [STREAM for GPU Memory](#stream-for-gpu-memory)
        - [ZE_PEER for Card to Card Bandwidth through XE Link](#ze_peer-for-card-to-card-bandwidth-through-xe-link)
        - [ZE_BANDWIDTH for Bandwidth between Host and GPU card](#ze_bandwidth-for-bandwidth-between-host-and-gpu-card)
        - [ONECCL Benchmark for Cross Cards Communication](#oneccl-benchmark-for-cross-cards-communication)
    - [Validate GPU System with HPC Workload Benchmarks](#validate-gpu-system-with-hpc-workload-benchmarks)
        - [HPL High Performance LINKPACK Benchmark](#hpl-high-performance-linkpack-benchmark)
        - [HPL-AI Benchmark](#hpl-ai-benchmark)
    - [Validate GPU System with AI Workload Benchmarks](#validate-gpu-system-with-ai-workload-benchmarks)

<!-- /TOC -->

# System Validate Workload for Intel Data Center GPU Max Series (PVC)

## Requirements
```bash
# Make sure you have Intel GPU drivers, XPU Manager or xpu-smi,  and Intel oneAPI toolkits Installed by following the instructions in system-setup folder.
# For Intel GPU user space drivers, make sure the following packages are installed:
sudo apt-get install -y xpu-smi \
  intel-opencl-icd intel-level-zero-gpu level-zero \
  libigc-dev intel-igc-cm libigdfcl-dev level-zero-dev

# The scripts uses default path /opt/intel/oneapi for oneAPI. Use environment ONEAPI_ROOT to overwrite.

# Install the required system packages. 
# On Ubuntu system:
sudo apt install pciutils numactl dkms parallel  python3 python3-venv

# The docker container is optional. 
# Install the docker and make sure the current user is added into Docker group if you want to run workload from docker container
# Recommend to use the official docker installation from https://docs.docker.com/engine/install/ubuntu/

```

All the test cases are done with CPU frequency scaling governor as 'performance'. Please make sure to set it before you run the test.
```bash
# Set CPU frequency scaling governor to 'performance'
echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
```

## System Sanity Check
Run `sys_sanity_check.sh` to check the system readiness for PVC in following stages:
1. Detect system information including Host OS, CPU, Memory, Kernel etc.
2. Scan and check GPUs on System
3. Check GPU driver, firmware and consistency
4. GPU topologies and XE Link connection check
5. Check and set the host settings for performance
6. GPU working environment/temperature test by running ze_peak workload

```bash
# Run the sanity check
./sys_sanity_check.sh

# Use STAGE to select the stage to start, e.g.
STAGE=4 ./sys_sanity_check.sh

# Use VERBOSE=1 for futher detailed logs which also stored in sanity_check_logs folder, e.g.
VERBOSE=1 ./sys_sanity_check.sh

```

## Validate GPU System with Micro Workload Benchmarks
### GEMM for GPU Computation
Run `sys_val_gemm.sh` to verify the system GPU by running the GEMM workloads including DGEMM, SGEMM, BGEMM IGEMM etc.
The scripts will perform the following two test sets for each GEMM.
1. Run GEMM on each GPU one by one and verify the output data.
2. Run GEMM on all GPUs in parallel and verify the output data.
```bash
# To run GEMM test on current system:
./sys_val_gemm.sh

# Dry-run GEMM test which print out the test sets
DRYRUN=1 ./sys_val_gemm.sh

# Customize the test
# TESTSET=<default 0 for all, 1 for each GPU one by one, 2 for parallel run on all GPU>
# GPU=<default none, set the GPU number to run on dedicated GPU>
# GEMM=<default "fp64 fp32 f16 bf16 s8" for all test type, use fp64 for example to run the DGEMM only>
# ITERATIONS=<default 5 to run each GEMM for 5 iterations>
# Example:
# Run DGEMM on GPU 2 only with 10 iterations
TESTSET=1 GEMM=fp64 GPU=2 ITERATIONS=10 ./sys_val_gemm.sh
# Run DGEMM on all GPU in parallel with 10 iterations
TESTSET=2 GEMM=fp64 ITERATIONS=10 ./sys_val_gemm.sh

```

### STREAM for GPU Memory
Run `sys_val_stream.sh` to verify the system GPU by running the BabelStream workload including kernel Copy, Mul, Add, Triad and Dot.   
The scripts will run the STREAM in following two test sets.    
1. Run STREAM on each GPU one by one and verify the output.
2. Run STREAM on all GPUs in parallel and verify the output.
```bash
# To Run STREAM test on current system:
./sys_val_stream.sh

# Dry-run and print out the test sets
DRYRUN=1 ./sys_val_stream.sh

```

### ZE_PEER for Card to Card Bandwidth through XE Link
Run `sys_val_zepeer.sh` to check the card to card communication bandwidth through XE Link by running ze_peer workload from Level Zero Test.   
The bandwidth test kernel includes read, write, biread and biwrite. The kernel will run through different hardware engines in this test sets.   
The scripts will run ZE_PEER test with following test sets:   
1. Run test kernel across GPUs which have XE Link connected. The test is done one by one for each XE Link. The bandwidth data is sorted and verified.
2. Run parallel multiple targets tests, multiple ze_peer tests are invoked in parallel with the data verification.
```bash
# To Run ZE_PEER test on current system:
./sys_val_zepeer.sh

# Dry-run and print out the test command sets
DRYRUN=1 ./sys_val_zepeer.sh

```

### ZE_BANDWIDTH for Bandwidth between Host and GPU card
Run `sys_val_zebandwidth.sh` to test the data transmission bandwidth between Host and GPU card. The test uses ze_bandwidth workload from Level Zero Test.   
The test kernel includes H2D, D2H and BIDIR. The data copy engine 0(EU), 1(BCS0), 2(BCS1) are tested.   
The scripts mainly runs following test items:
1. For each copy engine, and for each test kernel, run the test on each GPU and verify the data.
2. For each copy engine and for each test kernel, run the test on all GPUs in parallel.
```bash
# To Run ZE_BANDWIDTH test on current system:
./sys_val_zebandwidth.sh

# Dry-run and print out the test command sets:
DRYRUN=1 ./sys_val_zebandwidth.sh

```

### ONECCL Benchmark for Cross Cards Communication
Run `sys_val_onecclbench.sh` to test the oneCCL performance for collectives including allreduce, allgatherv, alltoall, alltoallv, bcast, reduce, reduce_scatter.   
The benchmark workload is from oneAPI 2024.0 release.   
The following test is performed in this scripts:
1. For each groups of GPUs, run oneCCL benchmark test for each collective and verify the data.
```bash
# To Run oneCCL benchmark test on current system:
./sys_val_onecclbench.sh

# Dry-run and print out the test command sets:
DRYRUN=1 ./sys_val_zebandwidth.sh

```

## Validate GPU System with HPC Workload Benchmarks
### HPL (High Performance LINKPACK) Benchmark
Run `sys_val_hpl.sh` to test the HPL performance on current system. The HPL binaries uses the Intel Optimized HPL from Intel oneAPI Basekit.   
The scripts will try to find the HPL binaries from oneAPI installation folder. Download from Intel online website if not found.    
The scripts will create the HPL.dat file and test wrapper based on the GPU type and the number of GPUs to be benchmarked to achieve the best performance.   
We can find the HPL.dat file and test wrapper file in the bench_hpl_logs folder after the test.   
The HPL benchmark is done with following items:   
1. Run HPL benchmark on each GPU one by one and verify the data.
2. Run HPL benchmark on multiple GPUs (1, 2, 4 8) based on available GPUs and verify the data and calculate the scalability.
Refer to the [HPL instructions](../workloads/HPC/hpl/) for futher details.

```bash
# To Run HPL benchmark on current system:
./sys_val_hpl.sh

# Dry-run and print out the test items:
DRYRUN=1 ./sys_val_hpl.sh

# Run HPL on each GPU one by one
TESTSET=1 ./sys_val_hpl.sh

# Run with 8 GPU test only, useful for system stress test:
TESTSET=2 GPURUN=8 ./sys_val_hpl.sh
```

### HPL-AI Benchmark
HPL-AI benchmark uses the Intel® Optimized HPL-AI Benchmark from Intel oneAPI Base Toolkit. It is heavily modified based on the High-Performance LINPACK (HPL) Benchmark.   
The usage of the Intel® Optimized HPL-AI Benchmark is very similar to that of the Intel® Distribution for LINPACK Benchmark.   
Refer to the [HPL-AI instructions ](../workloads/HPC/hpl-ai/) for further details.
```bash
# Run predefine HPL-AI benchmark:
./sys_val_hpl_ai.sh

# Run HPL-AI on each GPU one by one
TESTSET=1 ./sys_val_hpl_ai.sh

# Run HPL-AI on 8 GPUs
TESTSET=2 GPUNUM=8 ./sys_val_hpl_ai.sh
```

## Validate GPU System with AI Workload Benchmarks
Please refer to [Instruction of AI](AI/README.md)
