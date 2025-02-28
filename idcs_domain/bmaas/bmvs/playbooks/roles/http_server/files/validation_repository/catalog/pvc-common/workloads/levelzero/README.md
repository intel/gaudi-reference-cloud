# Instructions to build oneAPI Level Zero Performance Test

```
# Install the system package requirements
sudo apt install level-zero level-zero-dev libpng-dev libboost-all-dev swig cmake python3

# Download the source and build
git clone https://github.com/oneapi-src/level-zero-tests
cd level-zero-tests

# The version below used for prebuilt binaries
#git checkout 2d8e8e14a0dbf5afa816af9a5982a743bb638c68

mkdir build & cd build
cmake -D CMAKE_INSTALL_PREFIX=$PWD/../out -DBUILD_ZE_PERF_TESTS_ONLY=1 -DENABLE_ZESYSMAN=yes ..
cmake --build . --config Release --target install
cd ../out

# Get the binary in the level-zero-tests/out folder, including ze_peak, ze_peer, ze_bandwidth etc
# zesysman as a command-line interface to oneAPI Level Zero system resource management services is also built in the same folder
# e.g. run ze_bandwidth benchmark
./ze_bandwidth

```

# Reference
> https://github.com/oneapi-src/level-zero-tests

