## BabelStream benchmark implemented with SYCL2020

### Get source code and build

```
# Download source code
git clone https://github.com/UoB-HPC/BabelStream
git checkout v5.0

# Initialize oneAPI environment
source /opt/intel/oneapi/setvars.sh

# Build
cd BabelStream
cmake -Bbuild -H. -DMODEL=sycl2020-acc -DSYCL_COMPILER=ONEAPI-ICPX
cd build
make

# After the build, the sycl2020-stream binary will be generated.
# Use ./sycl2020-acc-stream -h for options

```

### Run BabelStream

```
# Initialize oneAPI environment 
source /opt/intel/oneapi/setvars.sh

# Run through level zero on the first stack of first GPU device
ONEAPI_DEVICE_SELECTOR=level_zero:gpu ZE_AFFINITY_MASK=0 ./sycl2020-acc-stream -s 134217728

# Run on a 2-stack GPU through implicit scaling
ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE ONEAPI_DEVICE_SELECTOR=level_zero:gpu ZE_AFFINITY_MASK=0 EnableImplicitScaling=1 ./sycl2020-acc-stream -s 134217728
```
