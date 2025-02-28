# DPEcho
General Relativity in SYCL for 2020 compatible compilers and beyond.

DPEcho uses a SYCL+MPI porting of the General-Relativity-Magneto-Hydrodynamic (GR-MHD) OpenMP+MPI code Echo, used to model instabilities, turbulence, propagation of waves, stellar winds and magnetospheres, and astrophysical processes around Black Holes.

DPEcho uses exclusively SYCL structures for memory and data management, and the flow control revolves entirely around critical device-code blocks, for which the key physics kernels were re-designed: most data reside almost permanently on the device, maximizing computational times. As a result, on the core physics elements ported so far, the measured performance gain is above 4x on HPC CPU hardware, and of order 7x on commercial GPUs.
# Prerequisites
Intel oneAPI DPC++ Compiler

Intel oneAPI toolkit (version >= 2023.0) targeting Intel CPUs and Intel GPUs.

CMake (>= 3.13)
# Build Steps
git clone https://github.com/LRZ-BADW/DPEcho.git

cd DPEcho

mkdir -p build && cd build

cmake .. && make

//cmake command for single GPU on single node and MPI OFF

cmake .. -DSYCL_DEVICE=GPU

//cmake command for Multiple GPUs and MPI ON

cmake .. -DSYCL_DEVICE=GPU -DENABLE_MPI=1

make

# Input
DPEcho expects a parameter file called echo.par in its working directory. 
The path to an alternative file may also be passed as a command line argument with option -f
sample echo.par file is given in this repository.
# Sample Output
The workload for DPEcho considered here is Alfvén wave for Grid sizes: 36³, 48³, 72³, 96³, 132³, 192³, 264³, 390³, 516³ cells

user1@7cc25526b7a4:~/anita/benchmarks$ echo_gpu

I:     Raising Hx to allowed min 3

I:     Raising Hy to allowed min 3

I:     Raising Hz to allowed min 3

[0][DeviceConfig::deviceWith] Looking at DPEcho default device.

[0][DeviceConfig::printTargetInfo]

        Hardware Intel(R) Data Center GPU Max 1100GPU
        Max Compute Units  : 448
        
[0]
        Max Work Group Size: 1024
        Global Memory / GB : 45.5852
        Local  Memory / kB : 128

[0][Domain::cartInfo] CartCoords:    (0 0 0) CartPeriodic:  (1 1 1) 

[0][Domain::locInfo] Local size (1 1 1) From (-0.5 -0.5 -0.5) To (0.5 0.5 0.5)

[0][Grid::print]  halos_(3 3 3)  cellsWH(70 70 70)  cellsNH(64 64 64)  Dxyz   (0.015625 0.015625 0.015625)

[0][main] Step 1 out 0 characteristic: 168.696 in 1.16498 s

[0][main] Step 2 out 0 characteristic: 168.69 in 0.00709701 s

[0][main] Step 3 out 0 characteristic: 168.696 in 0.00688291 s

[0][main] Step 4 out 0 characteristic: 168.691 in 0.00696802 s

[0][main] Step 5 out 0 characteristic: 168.696 in 0.00685906 s

[0][main] Step 6 out 0 characteristic: 168.692 in 0.00677109 s

[0][main] Step 7 out 0 characteristic: 168.695 in 0.00686884 s

[0][main] Step 8 out 0 characteristic: 168.693 in 0.00680017 s

[0][main] Step 9 out 0 characteristic: 168.695 in 0.00683188 s

[0][main] Step 10 out 0 characteristic: 168.694 in 0.00686097 s

[0][main] Step 11 out 0 characteristic: 168.694 in 0.00685596 s

[0][main] Step 12 out 0 characteristic: 168.695 in 0.00677419 s

[0][main] Step 13 out 0 characteristic: 168.693 in 0.00678587 s

[0][main] Step 14 out 0 characteristic: 168.695 in 0.00683713 s

[0][main] Step 15 out 0 characteristic: 168.692 in 0.00674391 s

[0][main] Step 16 out 0 characteristic: 168.696 in 0.00674009 s

[0][main] Step 17 out 0 characteristic: 168.691 in 0.00679278 s

[0][main] Step 18 out 0 characteristic: 168.696 in 0.00673819 s

[0][main] Step 19 out 0 characteristic: 168.69 in 0.00675297 s

[0][main] Step 20 out 0 characteristic: 168.696 in 0.00682592 s

[0][main] Step 21 out 0 characteristic: 168.69 in 0.00687814 s
