<!-- TOC -->

- [NAMD](#namd)
    - [Build NAMD](#build-namd)
        - [Download NAMD source](#download-namd-source)
        - [Install NAMD dependencies](#install-namd-dependencies)
        - [Build NAMD for single-process benchmarking with Intel oneAPI](#build-namd-for-single-process-benchmarking-with-intel-oneapi)
        - [Build NAMD for multi-process benchmarking with Intel oneAPI](#build-namd-for-multi-process-benchmarking-with-intel-oneapi)
    - [Benchmark NAMD](#benchmark-namd)
        - [ApoA1 benchmark 92,224 atoms, periodic, PME](#apoa1-benchmark-92224-atoms-periodic-pme)
            - [Download benchmarks](#download-benchmarks)
            - [NAMD single-process benchmarking](#namd-single-process-benchmarking)
            - [NAMD multi-process benchmarking](#namd-multi-process-benchmarking)
        - [STMV benchmark 1,066,628 atoms, periodic, PME](#stmv-benchmark-1066628-atoms-periodic-pme)
            - [Download benchmarks](#download-benchmarks)
            - [NAMD single-process benchmarking](#namd-single-process-benchmarking)
            - [NAMD multi-process benchmarking](#namd-multi-process-benchmarking)
    - [Reference](#reference)

<!-- /TOC -->
# NAMD
[NAMD](https://www.ks.uiuc.edu/Research/namd/) is a parallel molecular dynamics code designed for high-performance simulation of large biomolecular systems.
The guide provides the instructions to build and run NAMD on Intel Data Center GPU Max series.   
Please make sure you have [Intel Data Center GPU Drivers](https://dgpu-docs.intel.com/driver/installation.html) rolling stable version [20231031](https://dgpu-docs.intel.com/releases/stable_736_25_20231031.html) installed. 
Intel oneAPI Base Toolkit and HPC Toolkit [2024.0 release](https://www.intel.com/content/www/us/en/developer/tools/oneapi/toolkits.html) are required. The instructions use default installation path from /opt/intel/oneapi. 

## Build NAMD
The NAMD source code is published in gitlab. You need to send request to UIUC for access.   
Please visit [NAMD gitlab repo](https://gitlab.com/tcbgUIUC/namd) for instructions to apply the access.   

### Download NAMD source
```
git clone https://gitlab.com/tcbgUIUC/namd  
cd namd
git checkout oneapi-forces

# The commit we use for test
# git checkout b985d116804744b6742081641cd3011ff9a2e5a5
```

### Install NAMD dependencies
Download the NAMD dependencies in the NAMD directory
```
cd namd

# fftw
wget http://www.ks.uiuc.edu/Research/namd/libraries/fftw-linux-x86_64.tar.gz
tar xzf fftw-linux-x86_64.tar.gz
mv linux-x86_64 fftw

# tcl
wget http://www.ks.uiuc.edu/Research/namd/libraries/tcl8.5.9-linux-x86_64.tar.gz
wget http://www.ks.uiuc.edu/Research/namd/libraries/tcl8.5.9-linux-x86_64-threaded.tar.gz
tar xzf tcl8.5.9-linux-x86_64.tar.gz
tar xzf tcl8.5.9-linux-x86_64-threaded.tar.gz
mv tcl8.5.9-linux-x86_64 tcl
mv tcl8.5.9-linux-x86_64-threaded tcl-threaded 

# charm++
git clone https://github.com/UIUC-PPL/charm.git
cd charm
git checkout 9f17915271523de2ba36efd363ad1961b16765f9
cd ..

```

### Build NAMD for single-process benchmarking with Intel oneAPI
The build and run procedure for single-process executables is restricted to a single node and a single NAMD communication thread.   
Use instructions for multi-process build if you want to run multiple NAMD processes on single node or multiple nodes.
```
# In NAMD directory
cd namd

# setup the oneAPI environment. Make sure you have both Intel oneAPI base toolkit and HPC toolkit installed
source /opt/intel/oneapi/setvars.sh

# Build single-process Charm++ library
cd charm
./build charm++ multicore-linux-x86_64 icx -j --with-production
cd ..

# Build single-process NAMD executable.
./config Linux-x86_64-dpcpp-AOT --charm-arch multicore-linux-x86_64-icx
# Add options below to config in case the dependencies are downloaded in the other directory.
# --charm-base /path/to/charm --fftw-prefix /path/to/fftw --tcl-prefix /path/to/tcl-threaded
cd Linux-x86_64-dpcpp-AOT/
export  SYCL_CACHE_PERSISTENT=1
make -j

# with a successful build, the namd2 binary can be find in the current folder
# Sanity Check, in namd/Linux-x86_64-dpcpp-AOT/ folder
./namd2 +p 4 +devices 0 +platform "Intel(R) Level-Zero" ../src/alanin

```

### Build NAMD for multi-process benchmarking with Intel oneAPI
Multi-process instructions enable the build and run with multiple NAMD process on single node or multiple nodes.
```
# In NAMD directory
cd namd

# setup the oneAPI environment. Make sure you have both Intel oneAPI base toolkit and HPC toolkit installed
source /opt/intel/oneapi/setvars.sh
CC=icx; CXX=icpx; F90=ifort; F77=ifort; MPICXX=mpiicpc; MPI_CXX=mpiicpc
I_MPI_CC=icx;I_MPI_CXX=icpx;I_MPI_F90=ifort;I_MPI_F77=ifort
export I_MPI_CC I_MPI_CXX I_MPI_F90 I_MPI_F77 CC CXX F90 F77 MPICXX MPI_CXX

# Build multi-process Charm++ library
cd charm
./build charm++ mpi-linux-x86_64 smp mpicxx --with-production
cd ..

# Build multi-process NAMD executable.
./config Linux-x86_64-dpcpp-AOT --charm-arch mpi-linux-x86_64-smp-mpicxx
# Add options below to config in case the dependencies are downloaded in the other directory.
# --charm-base /path/to/charm --fftw-prefix /path/to/fftw --tcl-prefix /path/to/tcl-threaded
cd Linux-x86_64-dpcpp-AOT/
export  SYCL_CACHE_PERSISTENT=1
make -j

# with a successful build, the namd2 binary can be find in the current folder
# Sanity Check, in namd/Linux-x86_64-dpcpp-AOT/ folder
mpirun -perhost 2 ./namd2 +ppn 4 +pemap 1-9:5.2 +commap 0-9:5 +devices 0,1 +platform 'Intel(R) Level-Zero' ../src/alanin

```

## Benchmark NAMD
The benchmarks can be downloaded into a customized folder. Here we use the namd2 binary folder by default.   
The benchmark output can be saved into a log file. Use the python script [ns_per_day.py](https://www.ks.uiuc.edu/~dhardy/scripts/ns_per_day.py) to process the log file and report the performance.
```
cd namd/Linux-x86_64-dpcpp-AOT/
wget https://www.ks.uiuc.edu/~dhardy/scripts/ns_per_day.py
python3 ns_per_day.py <path to namd log file>
```

### [ApoA1 benchmark](https://www.ks.uiuc.edu/Research/namd/utilities/) (92,224 atoms, periodic, PME)

#### Download benchmarks
```
wget https://www.ks.uiuc.edu/Research/namd/utilities/apoa1.tar.gz
tar -xvf apoa1.tar.gz
cd apoa1
wget https://www.ks.uiuc.edu/Research/namd/2.13/benchmarks/apoa1_nve_cuda.namd
cd ..
```

#### NAMD single-process benchmarking
```
# Update apoa1/apoa1_nve_cuda.namd to enable the offloading to GPU for better performance
# Add following 3 lines in the apoa1_nve_cuda.namd file
		BondedCUDA=0
		PMEOffload=off
		usePMECUDA=on

# In order to get the best performance, the number of CPU cores, the CPU affinity etc should be fine-tuned based on the system configurations and topologies.
# Below we use a two sockets server with 8 Intel Max 1550 GPU, two Intel(R) Xeon(R) Platinum 8468V CPU and 48 physical cores per CPU for example:

# Setup environments
source /opt/intel/oneapi/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1

# Run benchmarks on 1 Max 1550 GPU (total 2 tiles/devices) with 48 CPU cores for example
./namd2 +p 48 +pemap 24-71 +splittiles +devices 0,1 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd 2>&1 |tee apoa1.log
python3 ns_per_day.py apoa1.log

# Run benchmarks on two Max 1550 GPU devices (total 4 tiles/devices) with 48 CPU cores for example
./namd2 +p 48 +pemap 24-71 +splittiles +devices 0,1,2,3 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd 2>&1 |tee apoa1.log
python3 ns_per_day.py apoa1.log

```

#### NAMD multi-process benchmarking
```
# Update apoa1/apoa1_nve_cuda.namd to enable the offloading to GPU for better performance
# Add following 3 lines in the apoa1_nve_cuda.namd file
		BondedCUDA=0
		PMEOffload=on
		usePMECUDA=off

# In order to get the best performance, the options should be fine-tuned based on the system configurations and topologies.
# Below we use a two sockets server with 8 Intel Max 1550 GPU, two Intel(R) Xeon(R) Platinum 8468V CPU and 48 physical cores per CPU for example:

# Setup environments
source /opt/intel/oneapi/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=FLAT
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1

# Run benchmarks on 1 Max 1550 GPU (total 2 tiles/devices), two MPI processes, each process uses 24 cores including one core for communication as example
mpirun -perhost 2 ./namd2 +ppn 23 +pemap 25-47:24.23,49-95:24.23 +commap 24-47:24,48-95:24 +devices 0,1 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd 

# Run benchmarks on 2 Max 1550 GPU, 4 processes and each uses 16 CPU cores for example
mpirun -perhost 4 ./namd2 +ppn 15 +pemap 17-47:16.15,49-95:16.15 +commap 16-47:16,48-95:16 +devices 0,1,2,3 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd

# Run benchmarks on 4 Max 1550 GPU, 8 processes and each uses 10 CPU cores for example
mpirun -perhost 8 ./namd2 +ppn 9 +pemap 9-47:10.9,49-95:10.9 +commap 8-47:10,48-95:10 +devices 0,1,2,3,4,5,6,7 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd

# Run benchmarks on 8 Max 1550 GPU, 16 processes and each uses 6 CPU cores for example
mpirun -perhost 16 ./namd2 +ppn 5 +pemap 1-47:6.5,49-95:6.5 +commap 0-47:6,48-95:6 +devices 0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15 +platform 'Intel(R) Level-Zero' apoa1/apoa1_nve_cuda.namd

```

### [STMV benchmark](https://www.ks.uiuc.edu/Research/namd/utilities/) (1,066,628 atoms, periodic, PME)

#### Download benchmarks
```
wget https://www.ks.uiuc.edu/Research/namd/utilities/stmv.tar.gz
tar xzf stmv.tar.gz
cd stmv
wget https://www.ks.uiuc.edu/Research/namd/2.13/benchmarks/stmv_nve_cuda.namd
cd ..
```

#### NAMD single-process benchmarking
```
# Update stmv/stmv_nve_cuda.namd to enable the offloading to GPU for better performance
# Add following 3 lines in the stmv_nve_cuda.namd file
		BondedCUDA=0
		PMEOffload=off
		usePMECUDA=on

# In order to get the best performance, the number of CPU cores, the CPU affinity etc should be fine-tuned based on the system configurations and topologies.
# Below we use a two sockets server with 8 Intel Max 1550 GPU, two Intel(R) Xeon(R) Platinum 8468V CPU and 48 physical cores per CPU for example:

# Setup environments
source /opt/intel/oneapi/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1

# Run benchmarks on 1 Max 1550 GPU (total 2 tiles/devices) with 68 CPU cores for example
./namd2 +p 68 +pemap 14-81 +splittiles +devices 0,1 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd 2>&1 |tee stmv.log
python3 ns_per_day.py stmv.log

# Run benchmarks on 2 Max 1550 GPU devices (total 4 tiles/devices) with 48 CPU cores for example
./namd2 +p 68 +pemap 14-81 +splittiles +devices 0,1,2,3 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd 2>&1 |tee stmv.log
python3 ns_per_day.py stmv.log

```

#### NAMD multi-process benchmarking
```
# Update stmv/stmv_nve_cuda.namd to enable the offloading to GPU for better performance
# Add following 3 lines in the apoa1_nve_cuda.namd file
		BondedCUDA=0
		PMEOffload=on
		usePMECUDA=off

# In order to get the best performance, the options should be fine-tuned based on the system configurations and topologies.
# Below we use a two sockets server with 8 Intel Max 1550 GPU, two Intel(R) Xeon(R) Platinum 8468V CPU and 48 physical cores per CPU for example:

# Setup environments
source /opt/intel/oneapi/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=FLAT
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1

# Run benchmarks on 1 Max 1550 GPU (total 2 tiles/devices), two MPI process, each process uses 32 cores including one core for communication as example
mpirun -perhost 2 ./namd2 +ppn 31 +pemap 17-47:32.31,49-95:32.31 +commap 16-47:32,48-95:32 +devices 0,1 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd

# Run benchmarks on 2 Max 1550 GPU, 4 processes and each uses 20 CPU cores for example
mpirun -perhost 4 ./namd2 +ppn 19 +pemap 9-47:20.19,49-95:20.19 +commap 8-47:20,48-95:20 +devices 0,1,2,3 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd

# Run benchmarks on 4 Max 1550 GPU, 8 processes and each uses 10 CPU cores for example
mpirun -perhost 8 ./namd2 +ppn 9 +pemap 9-47:10.9,49-95:10.9 +commap 8-47:10,48-95:10 +devices 0,1,2,3,4,5,6,7 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd

# Run benchmarks on 8 Max 1550 GPU, 6 processes and each uses 6 CPU cores for example
mpirun -perhost 16 ./namd2 +ppn 5 +pemap 1-47:6.5,49-95:6.5 +commap 0-47:6,48-95:6 +devices 0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15 +platform 'Intel(R) Level-Zero' stmv/stmv_nve_cuda.namd

```

## Reference
- https://gitlab.com/tcbgUIUC/namd
- https://www.ks.uiuc.edu/Research/namd/ 
- https://www.ks.uiuc.edu/Research/namd/benchmarks/ 
- https://www.ks.uiuc.edu/Research/namd/utilities/ 
