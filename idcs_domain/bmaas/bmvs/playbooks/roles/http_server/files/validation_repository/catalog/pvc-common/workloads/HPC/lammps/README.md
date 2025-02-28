# LAMMPS
[LAMMPS Molecular Dynamics Simulator](https://www.lammps.org)
Use the instructions below to build and run LAMMPS on Intel Data Center GPU.

## Prerequisites
On the host system, please make sure you have Intel oneAPI Basekit installed.
Install the following packages, e.g. on Ubuntu host system:
```
sudo apt-get install -y ca-certificates curl unzip python3-pip python3-mako clinfo numactl intel-opencl-icd
python3 -m pip install pyyaml
 
```

## Build LAMMPS
Follow the instructions to build LAMMPS binary. Please note that the build option may be customized or further tuned for performance.   
Refer to [LAMMPS manual](https://docs.lammps.org/Manual.html) for more detailed instructions.   
```
WORKDIR=`pwd`

# Clone the source code
git clone https://github.com/intel/compute-aggregation-layer.git cal
git clone https://www.github.com/lammps/lammps lammps -b develop --depth 1

# Build cal
cd cal
mkdir build; cd build
cmake ../
make -j
cd $WORKDIR

# Build LAMMPS
export PATH=$WORKDIR/cal/build:$PATH
cd lammps/src
make yes-asphere yes-kspace yes-manybody yes-misc
make yes-molecule yes-rigid yes-dpd-basic yes-gpu
cd ../lib/gpu
make -f Makefile.oneapi -j
cd ../src
make oneapi -j
cd $WORKDIR

```

## Run LAMMPS TEST
Please note that the runtime options can be tuned for better performance.   
Please refer the [LAMMPS GPU package](https://docs.lammps.org/Speed_gpu.html) for further details.

```
export LMP_ROOT=$WORKDIR/lammps
export CAL_ROOT=$WORKDIR/cal
export PATH=${CAL_ROOT}/build:${LMP_ROOT}/src:$PATH
source /opt/intel/oneapi/setvars.sh

cd $LMP_ROOT/src/INTEL/TEST
# Create restart file, this is onetime setup, e.g.
mpirun --bootstrap ssh -np 54 lmp_oneapi -in in.lc_generate_restart -log none

# Run TEST, e.g. in.intel.lc test
export I_MPI_PIN_DOMAIN=auto:compact
export I_MPI_FABRICS=shm
export KMP_AFFINITY="granularity=core,scatter"
export KMP_BLOCKTIME=1000

export NEOReadDebugKeys=1
export DirectSubmissionRelaxedOrdering=1
export DirectSubmissionRelaxedOrderingForBcs=1

## Run on 1 GPU/Tile with 16 ranks
calrun mpirun -np 16 lmp_oneapi -in in.intel.lc -v N off -pk gpu 1 -sf gpu -log none -screen test.log
## Run on 2 GPUs/Tiles with 16 ranks, 8 ranks for each GPU
calrun mpirun -np 16 lmp_oneapi -in in.intel.lc -v N off -pk gpu 2 -sf gpu -log none -screen test.log
 
```

## Quick Start
The scripts is built for quick start and test. You can use these scripts to build and run LAMMPS test with the predefined options. Please note that the predefined options may not the best options for the best performance from your system.
```
# Setup the LAMMPS in current system
bash lammps_setup.sh

# Run LAMMPS TEST with default in.intel.lc test
bash lammps_run.sh

# RUN LAMMPS eam TEST on 8 GPU Tiles with 8 ranks per GPU
NUMBER_OF_PROCESS=64 TILES=8 TESTCASE=eam bash lammps_run.sh

```
A dockerfile is provided to build and run LAMMPS test. e.g.
```
# Build container for LAMMPS
bash lammps_container_setup.sh

# Run LAMMPS eam TEST on 8 GPU Tiles with 8 ranks per GPU
NUMBER_OF_PROCESS=64 TILES=8 TESTCASE=eam bash lammps_container_run.sh

```

