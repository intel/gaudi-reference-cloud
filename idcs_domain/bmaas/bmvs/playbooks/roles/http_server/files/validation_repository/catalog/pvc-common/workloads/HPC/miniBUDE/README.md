## miniBUDE

### How to Build

```
git clone https://github.com/UoB-HPC/miniBUDE
cd miniBUDE/sycl
source /opt/intel/oneapi/setvars.sh
sed -i 's/CMAKE_CXX_STANDARD 11/CMAKE_CXX_STANDARD 17/g' CMakeLists.txt
cmake -Bbuild -DCMAKE_BUILD_TYPE=Release -DSYCL_RUNTIME=DPCPP
cmake --build build --target bude --config Release -j $(nproc)

```

### How to Run

```
Make sure oneAPI environment is loaded:

    source /opt/intel/oneapi/setvars.sh

To see the list of devices:

    build/bude -h

To run on a device that has GPU in its name:

    build/bude -d GPU

To run BIG5 dataset (2672 Ligands, 2672 proteins)

    build/bude -d GPU --deck ../data/bm2
```

### Sample output

```
Available SYCL devices:
  0. Intel(R) FPGA Emulation Device(accelerator)
  1. Intel(R) Xeon(R) Platinum 8480+(cpu)
  2. Intel(R) Data Center GPU Max 1550(gpu)
  3. Intel(R) Data Center GPU Max 1550(gpu)
  4. Intel(R) Data Center GPU Max 1550(gpu)
  5. Intel(R) Data Center GPU Max 1550(gpu)
  6. Intel(R) Data Center GPU Max 1550(gpu)
  7. Intel(R) Data Center GPU Max 1550(gpu)
  8. Intel(R) Data Center GPU Max 1550(gpu)
  9. Intel(R) Data Center GPU Max 1550(gpu)
Unable to parse/select device index `GPU`:stoul
Attempting to match device with substring  `GPU`
Using first device matching substring `GPU`
Device    : Intel(R) Data Center GPU Max 1550
        Type    : gpu
        Profile : FULL_PROFILE
        Version : 3.0
        Vendor  : Intel(R) Corporation
        Driver  : 23.05.25593.18
Poses     : 65536
Iterations: 8
Ligands   : 2672
Proteins  : 2672
Deck      : ../data/bm2
WG        : 4 (use nd_range:true)
Context time:    6.43777 ms
Xfer+Alloc time: 7.06953 ms
Warmup time:     7300.15 ms

Kernel time:    57462.460 ms
Average time:   7182.807 ms
Interactions/s: 65.142 billion
GFLOP/s:        2606.153
GFInst/s:       1628.541
Energies
24293.06
46384.62
22487.10
35002.60
45883.47
47784.92
17952.29
25230.55
34944.67
14615.40
80945.87
32816.04
46287.79
26282.73
36581.76
20462.15
Largest difference was 0.000%.
```

### Notes

```
Note 1.

    oneAPI DPC++ Library (oneDPL) CMake scripts now enforce C++17 as the minimally required language version.
    Because of this, CMAKE_CXX_STANDARD needs to be updated from 11 to 17 in the CMake script for DPCPP runtime.
    The change is done using 'sed' command in the build instructions.
    Other SYCL runtime configurations in the CMakeList.txt are already targeting C++17.
    See https://github.com/oneapi-src/oneDPL/blob/main/documentation/release_notes.rst

Note 2.

    There were a number of 'deprecated' warnings during the build process.

```


