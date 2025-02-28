# easyWave-sycl

[easyWave](https://git.gfz-potsdam.de/id2/geoperil/easyWave) is an application that is used to simulate tsunami generation and propagation in the context of early warning. It makes use of GPU acceleration to speed up the calculations. 

[SYCL-focused fork](https://github.com/christgau/easywave-sycl) of the tsunami simulation easyWave is described in the guide. Intel aslo adapted another [repo](https://github.com/oneapi-src/Velocity-Bench.git) for easyWave.


## Prepare the Source Codes and Build

Below is the command line to clone the source codes and build out easywve-sycl binary.

```
git clone https://github.com/christgau/easywave-sycl.git

cd easywave-sycl/

cp make.inc.oneapi make.inc

source /opt/intel/oneapi/setvars.sh 

make
```

## Run with Sample Data

We can follow command parameters and data folder in orignal [easyWave](https://git.gfz-potsdam.de/id2/geoperil/easyWave) repo to run test sample.

1. List available GPU

```
# Load env if haven't
source /opt/intel/oneapi/setvars.sh

sycl-ls
```

The output should look like:
```
[opencl:acc:0] Intel(R) FPGA Emulation Platform for OpenCL(TM), Intel(R) FPGA Emulation Device 1.2 [2023.16.7.0.21_160000]
[opencl:cpu:1] Intel(R) OpenCL, Intel(R) Xeon(R) Platinum 8470Q 3.0 [2023.16.7.0.21_160000]
[opencl:gpu:2] Intel(R) OpenCL Graphics, Intel(R) Data Center GPU Max 1550 3.0 [23.22.26516.25]
[opencl:gpu:3] Intel(R) OpenCL Graphics, Intel(R) Data Center GPU Max 1550 3.0 [23.22.26516.25]
[opencl:gpu:4] Intel(R) OpenCL Graphics, Intel(R) Data Center GPU Max 1550 3.0 [23.22.26516.25]
[ext_oneapi_level_zero:gpu:0] Intel(R) Level-Zero, Intel(R) Data Center GPU Max 1550 1.3 [1.3.26516]
[ext_oneapi_level_zero:gpu:1] Intel(R) Level-Zero, Intel(R) Data Center GPU Max 1550 1.3 [1.3.26516]
[ext_oneapi_level_zero:gpu:2] Intel(R) Level-Zero, Intel(R) Data Center GPU Max 1550 1.3 [1.3.26516]
```

2. Run easyWave-sycl on single stack.

Run esaywave-sycl on the first stack of first GPU.
```
# Load env if haven't
source /opt/intel/oneapi/setvars.sh

ZE_AFFINITY_MASK=0.0 ./easywave-sycl -gpu -verbose -propagate 2880 -step 5 -grid "../data/grids/e2r4Pacific.grd" -source "../data/faults/uz.Tohoku11.grd" -time 1440
```

The output:
```
easyWave ver.2023-02-24/ZIB
Selected device: Intel(R) Data Center GPU Max 1550
Profiling supported: 1
Maximum Work group size: 1024
USM explicit allocations supported: 1
Built-in kernels: None.
per-kernel profiling activated
Model time = 00:00:00,   elapsed: 156 msec      domain (1410, 456)-(1513, 536)
...
Model time = 24:00:00,   elapsed: 3308 msec     domain (2, 2)-(1800, 2850)
runtime kernel 0 (wave_update): 1039.122 ms (0.443)
runtime kernel 1 (wave_boundary): 104.658 ms (0.045)
runtime kernel 2 (flux_update): 919.683 ms (0.392)
runtime kernel 3 (flux_boundary): 154.039 ms (0.066)
runtime kernel 4 (grid_extend): 74.476 ms (0.032)
runtime kernel 5 (memset_zero): 24.126 ms (0.010)
runtime kernel 6 (memcpy_extent): 31.119 ms (0.013)
kernels total: 2347.224
Runtime: 3.308 s, final domain: (2, 2)-(1800, 2850), size: 1799 x 2849 
```

3. Run sample test on single GPU. Implicit scaling isn't suggested at this time. When rnning on 2 stacks GPU(like Max 1550), the workload will be implicitly scaled to 2 stack and show lower performance than signle stack.

```
# Load env if haven't
source /opt/intel/oneapi/setvars.sh

./easywave-sycl -gpu -verbose -propagate 2880 -step 5 -grid "../data/grids/e2r4Pacific.grd" -source "../data/faults/uz.Tohoku11.grd" -time 1440
```
