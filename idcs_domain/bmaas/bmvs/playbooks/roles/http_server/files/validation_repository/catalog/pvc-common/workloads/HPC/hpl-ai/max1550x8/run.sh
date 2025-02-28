#!/bin/bash
ONEAPI_ROOT=/opt/intel/oneapi
source ${ONEAPI_ROOT}/setvars.sh
export PATH=$PATH:${ONEAPI_ROOT}/mkl/latest/share/mkl/benchmarks/mp_linpack/
mpirun -np 2 bash ./hpl_wrapper.sh :0,:1,:2,:3,:4,:5,:6,:7 :8,:9,:10,:11,:12,:13,:14,:15

#export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
#mpirun -np 2 bash ./hpl_wrapper.sh :0.0,:0.1,:1.0,:1.1,:2.0,:2.1,:3.0,:3.1 :4.0,:4.1,:5.0,:5.1,:6.0,:6.1,:7.0,:7.1

