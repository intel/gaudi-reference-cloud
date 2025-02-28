# GROMACS
[GROMACS](https://www.gromacs.org/) is a free and open-source software suite for high-performance molecular dynamics and output analysis.   
Follow the instructions below to build and run GROMACS on Intel Data Center GPU.

## Prerequisites
On the host system, please make sure you have Intel oneAPI Basekit installed.   
Install the following packages, e.g. on Ubuntu host system:
```
sudo apt-get install -y ca-certificates wget curl unzip python3-pip python3-mako clinfo numactl intel-opencl-icd
```

## Build GROMACS
Follow the instructions to build GROMACS binary. The build option can be customized or further tuned for performance.
For example, use the option -DGMX_MPI=on to enable the MPI support.   
Refer to [GROMACS installation guide](https://manual.gromacs.org/current/install-guide/index.html) for more options.

```
# Download GROMACS source code
git clone https://gitlab.com/gromacs/gromacs.git

# Build GROMACS
cd gromacs
mkdir build; cd build
GMX_INSTALL=../__install
GMX_BUILD_OPTION="-DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DCMAKE_INSTALL_PREFIX=$GMX_INSTALL -DGMXAPI=OFF -DGMX_GPU=SYCL -DGMX_FFT_LIBRARY=mkl -DGMX_GPU_NB_CLUSTER_SIZE=8 -DGMX_GPU_NB_NUM_CLUSTER_PER_CELL_X=1"
source /opt/intel/oneapi/setvars.sh
cmake $GMX_BUILD_OPTION ..
make -j
make install

```
## Download test cases/benchmarks
We can use public test cases/benchmarks, for instance, use the following input files for Gromacs performance evaluations.
1. Benchmark set from [Max Planck GROMACS Benchmark Suite](https://www.mpinat.mpg.de/grubmueller/bench), including benchMEM (82k atoms), benchPEP (12M atoms), benchPEP-h (12M atoms, h-bouds constrained) etc. Please visit the [bench website](https://www.mpinat.mpg.de/grubmueller/bench) to download the benchmarks and get further details.
2. STMV - Satellite Tobacco Mosaic Virus (STMV) has about 1M atoms and is available from [Supplementary Information for Heterogeneous Parallelization and Acceleration of Molecular Dynamics Simulations in GROMACS](https://zenodo.org/record/3893789).
You may select a different test cases or benchmarks for your own evaluation.

```
mkdir testcases; cd testcases

# benchMEM (82k atoms)
wget -O benchMEM.zip https://www.mpinat.mpg.de/benchMEM
unzip benchMEM.zip

# benchPEP (12M atoms)
wget -O benchPEP.zip https://www.mpinat.mpg.de/benchPEP
unzip benchPEP.zip

# benchPEP-h (12M atoms, h-bounds constrained)
wget -O benchPEP-h.zip https://www.mpinat.mpg.de/benchPEP-h
unzip benchPEP-h.zip

# STMV
wget -O GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz https://zenodo.org/record/3893789/files/GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz
tar xf GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP.tar.gz
ln -s GROMACS_heterogeneous_parallelization_benchmark_info_and_systems_JCP/stmv/topol.tpr

```

## Run GROMACS test
To run GROMACS test, set the GROMACS environment and run with proper options.   
Please notes that the runtime options can be fine-tuned for different type of systems and GPU configurations.   
Please follow the [GROMACS user guide](https://manual.gromacs.org/current/user-guide/mdrun-performance.html) for further details.
```
# Set the environment
source /opt/intel/oneapi/setvars.sh
export ONEAPI_DEVICE_SELECTOR=level_zero:gpu
export SYCL_PI_LEVEL_ZERO_USE_IMMEDIATE_COMMANDLISTS=1
source $GMX_INSTALL/bin/GMXRC

# Run benchMEM for example:
gmx mdrun -nsteps 10000 -resetstep 5000 -pin on -cpt -1 -resethway -noconfout -nb gpu -pme gpu -notunepme -s testcases/benchMEM.tpr -ntmpi 1 -ntomp 16

# Run benchPEP-h for example:
gmx mdrun -nsteps 1000 -resetstep 500 -pin on -cpt -1 -resethway -noconfout -nb gpu -pme gpu -notunepme -s testcases/benchPEP-h.tpr -ntmpi 1 -ntomp 16

```

## Quick Start
The scripts is built for quick start and test. You can use these scripts to build and run GROMACS test with the predefined options. Please note that the predefined options may not the best options for the best performance from your system.
```
# Setup the GROMACS in current system
bash gromacs_setup.sh

# Run GROMACS test, default benchMEM
bash gromacs_run.sh

# Run benchPEP-h test
TESTCASE=benchPEP-h bash gromacs_run.sh

# Run benchPEP-h test with 2 MPI threads and each MPI threads with 16 OpenMP threads
TESTCASE=benchPEP-h NTMPI=2 NTOMP=16 bash gromacs_run.sh

```
A dockfile is provided to build and run GROMACS test, you can build the container and run GROMACS from container with predefined options.

```
# Build container with GROMACS
bash gromacs_container_setup.sh

# Run benchmMEM test
bash gromacs_container_run.sh

# Run benchPEP-h test
TESTCASE=benchPEP-h bash gromacs_container_run.sh

```


