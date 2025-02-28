# Instructions for oneCCL benchmark

## Build oneCCL benchamrk
```
# Get the oneCCL example code from oneAPI installation folder
cp /opt/intel/oneapi/ccl/latest/share/doc/ccl/examples ./ -r
cd examples && mkdir build && cd build
source /opt/intel/oneapi/setvars.sh
cmake .. -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DCOMPUTE_BACKEND=dpcpp
make -j && make install
benchmark/benchmark -h


# Alternatively we can get the oneCCL from github repo and build everything
git clone https://github.com/oneapi-src/oneCCL
cd oneCCL
git checkout 2021.10
mkdir build && cd build
source /opt/intel/oneapi/compiler/latest/env/vars.sh
cmake .. -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DCOMPUTE_BACKEND=dpcpp
make -j && make install
source _install/env/setvar.sh
_install/examples/benchmark/benchmark -h

```

## Run oneCCL benchmark
```
# For example, for allreduce collective
mpiexec -np 4 benchmark -w 16 -i 1000 -c last -b sycl -t 33554432 -f 33554432 -j off --coll allreduce

```

