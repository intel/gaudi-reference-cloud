#!/bin/bash
ONEAPI_ROOT=/opt/intel/oneapi
source ${ONEAPI_ROOT}/setvars.sh
export PATH=$PATH:${ONEAPI_ROOT}/mkl/latest/share/mkl/benchmarks/mp_linpack/
HPL_HOST_NODE=0 HPL_DEVICE=:0,:1  numactl -N 0 -m 0 xhpl-ai_intel64_dynamic_gpu

# for oneAPI 2023 version, make sure the oneMKL benchmark suite is download and the xhpl-ai_intel64_dynamic_gpu is added into the PATH
# HPL_HOST_NODE=0 HPL_DEVICE=:0.0,:0.1  numactl -N 0 -m 0 xhpl-ai_intel64_dynamic_gpu
