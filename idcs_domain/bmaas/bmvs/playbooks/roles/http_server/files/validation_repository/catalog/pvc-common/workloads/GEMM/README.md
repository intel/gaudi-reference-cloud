# GEMM benchmark tools - Based on oneAPI 2024.0

## Use oneMKL example for DGEMM/SGEMM
### Build DGEMM/SGEMM example with oneMKL
```
git clone https://github.com/oneapi-src/oneAPI-samples
git checkout 2024.0.0
cd oneAPI-samples/Libraries/oneMKL/matrix_mul_mkl
source /opt/intel/oneapi/setvars.sh
make

#dgemm.mkl and sgemm.mkl binary will be generated
#Alternatively for simpler steps
mkdir src & cd src
wget https://github.com/oneapi-src/oneAPI-samples/raw/master/Libraries/oneMKL/matrix_mul_mkl/GNUmakefile -O Makefile
wget https://github.com/oneapi-src/oneAPI-samples/raw/master/Libraries/oneMKL/matrix_mul_mkl/matrix_mul_mkl.cpp -O matrix_mul_mkl.cpp
source /opt/intel/oneapi/setvars.sh
make

```
### Run DGEMM/SGEMM
```
#DGEMM
./matrix_mul_mkl double

#SGEMM
./matrix_mul_mkl single

# Use ZE_AFFINITY_MASK=0 to run on device 0, stack 0
ZE_AFFINITY_MASK=0 ./matrix_mul_mkl double 8192

# Output Reference
---------------------------
Device:                  Intel(R) Data Center GPU Max 1550
Core/EU count:           512
Maximum clock frequency: 1600 MHz

Benchmarking (8192 x 8192) x (8192 x 8192) matrix multiplication, double precision
 -> Initializing data...
 -> Warmup...
 -> Timing...

Average performance: 16.6555TF

```


### Use benchdnn for SGEMM/GEMM-FP16/GEMM-BF16/IGEMM
#### Build benchdnn
```
wget https://github.com/oneapi-src/oneDNN/archive/refs/tags/v3.3.1.tar.gz -O oneDNN-3.3.1.tar.gz
tar xf oneDNN-3.3.1.tar.gz
cd oneDNN-3.3.1
mkdir -p build
cd build
source /opt/intel/oneapi/setvars.sh
export CC=icx
export CXX=icpx
cmake ..  -DDNNL_CPU_RUNTIME=SYCL -DDNNL_GPU_RUNTIME=SYCL 
make -j
cd tests/benchdnn
./benchdnn --matmul --help

# Further details please refer online document: https://oneapi-src.github.io/oneDNN/dev_guide_build.html

```

#### Run GEMM with benchdnn
```
ZE_AFFINITY_MASK=0 ./benchdnn --matmul --mode=p --perf-template=%Gflops% --engine=gpu --dt=<data-type> 8192x8192:8192x8192

# Change <data-type> with follow data type:
SGEMM (GEMM float32): f32
GEMM (GEMM float16) : f16
GEMM (GEMM bfloat16): bf16
IGEMM (GEMM int8)   : s8

# More details for benchdnn :  https://github.com/oneapi-src/oneDNN/tree/v3.2/tests/benchdnn

```
