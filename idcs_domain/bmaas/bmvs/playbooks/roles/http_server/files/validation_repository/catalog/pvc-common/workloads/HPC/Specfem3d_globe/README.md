# Specfem3D Globe
SPECFEM3D_GLOBE simulates global and regional (continental-scale) seismic wave propagation. Go to the official github repo for [Specfem3D Globe](https://github.com/SPECFEM/specfem3d_globe)   
The document provides the quick instructions to build and run Specfem3D for Intel Data Center GPU.   
Make sure you have latest [Intel oneAPI Base Toolkit and HPC Toolkit](https://www.intel.com/content/www/us/en/developer/tools/oneapi/toolkits.html) installed.

## Instructions for Specfem3D on Intel GPU
### 1. Download the source code
```
# Download the source code
git clone --branch devel --recursive https://github.com/SPECFEM/specfem3d_globe.git
```

### 2. Configure with Intel oneAPI tools
```
cd specfem3d_globe
source /opt/intel/oneapi/setvars.sh
./configure FC=ifort CC=icx CXX=icpx MPIFC=mpiifort --with-opencl OCL_LIBS=/opt/intel/oneapi/compiler/latest/linux/lib/libOpenCL.so OCL_INC="/opt/intel/oneapi/compiler/latest/linux/include/sycl/ -DFAST_2D_MEMCPY" OCL_GPU_FLAGS="-cl-std=CL3.0" MPI_INC=/opt/intel/oneapi/mpi/latest/include/
```

### 3. Edit the DATA/Par_file to use OpenCL runtime for Intel GPU
```
#In the file DATA/Par_file, for example for Intel Data Center GPU Max 1550
	# set to true to use GPUs
	GPU_MODE                        = .true.
	# Only used if GPU_MODE = .true. :
	GPU_RUNTIME                     = 2
	# 2 (OpenCL), 1 (Cuda) ou 0 (Compile-time -- does not work if configured with --with-cuda *AND* --with-opencl)
	GPU_PLATFORM                    = Intel(R) OpenCL Graphics
	GPU_DEVICE                      = Intel(R) Data Center GPU Max 1550
    
# Use below bash command to change the default DATA/Par_file automatically
	sed -i -E 's/(GPU_MODE.*=).*/\1 .true./' DATA/Par_file
	sed -i -E 's/(GPU_RUNTIME.*=).*/\1 2/' DATA/Par_file
	sed -i -E 's/(GPU_PLATFORM.*=).*/\1 Intel(R) OpenCL Graphics/' DATA/Par_file
	sed -i -E 's/(GPU_DEVICE.*=).*/\1 Intel(R) Data Center GPU Max 1550/' DATA/Par_file
```

### 4. Change the number of chunks, number of MPI processors
```
# Update the DATA/Par_file, for example:
	NCHUNKS = 1
	NPROC_XI  = 1
	NPROC_ETA = 2
# Use below bash command to change the value if need
	sed -i -E 's/(NCHUNKS.*=)(.*)/\1 1/' DATA/Par_file
	sed -i -E 's/(NPROC_XI.*=)(.*)/\1 1/' DATA/Par_file
	sed -i -E 's/(NPROC_ETA.*=)(.*)/\1 2/' DATA/Par_file
```

### 5. Make the Build
```
make clean
make -j default
```

### 6. Create the mesher 
This needs to be run only once after modifying the DATA/Par_file   
Create mesh using N processes, where N = NPROC_XE * NPROC_ETA
```
# NPROC_XE=1, NPROC_ETA=2, so we have N=2 for example
mpirun -n 2 bin/xmeshfem3D
```

### 7. Run the simulation
To run the main solver, use the N processes where N = NPROC_XE * NPROC_ETA
```
# NPROC_XE=1, NPROC_ETA=2, run with 2 mpi processes
mpirun -n 2 bin/xspecfem3D
```

### 8. Check the output
```
# Check the Device Name in the OUTPUT_FILES/gpu_device_info.txt
$ grep "Device Name" OUTPUT_FILES/gpu_device_info.txt
Device Name = Intel(R) Data Center GPU Max 1550

# Check the time of loop complete
$ grep -A1 "Time-Loop Complete" OUTPUT_FILES/output_solver.txt
Time-Loop Complete. Timing info:
Total elapsed time in seconds =  <...>

```

## Example
Use the EXAMPLES/global_s362ani_shakemovie on Intel Data Center GPU Max 1550
```
# Copy the Par_file from example folder to DATA/ folder
cp -r EXAMPLES/global_s362ani_shakemovie/DATA/* DATA/

# Modify DATA/Par_file, go through the step 3-8 based on above instructions to run on Intel Data Center GPU Max 1550
# Here we only list the updated configs in the Par_file

  # Change device info
	GPU_MODE                        = .true.
	GPU_RUNTIME                     = 2
	GPU_PLATFORM                    = Intel(R) OpenCL Graphics
	GPU_DEVICE                      = Intel(R) Data Center GPU Max 1550

  # Change the STEPS to reduce the time steps to run if needed, e.g.
    RECORD_LENGTH_IN_MINUTES	= 35.0d0

  # Run on 1 Max 1550 GPU
	NCHUNKS	= 1
	NPROC_XI	= 1
	NPROC_ETA	= 2
  # Run on 2 Max 1550 GPU
	NCHUNKS	= 1
	NPROC_XI	= 2
	NPROC_ETA	= 2
  # Run on 4 Max 1550 GPU
	NCHUNKS	= 1
	NPROC_XI	= 2
	NPROC_ETA	= 4
  # Run on 8 Max 1550 GPU
	NCHUNKS	= 1
	NPROC_XI	= 4
	NPROC_ETA	= 4

```
