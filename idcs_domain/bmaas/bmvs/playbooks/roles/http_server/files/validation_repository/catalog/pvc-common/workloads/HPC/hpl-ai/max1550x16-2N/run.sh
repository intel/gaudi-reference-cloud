#!/bin/bash
ONEAPI_ROOT=/opt/intel/oneapi
source ${ONEAPI_ROOT}/setvars.sh
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
export ZE_ENABLE_PCI_ID_DEVICE_ORDER=1
export I_MPI_OFFLOAD=0
export I_MPI_PIN_RESPECT_HCA=0
export I_MPI_PIN_DOMAIN=auto 
export I_MPI_PIN_ORDER=bunch
export I_MPI_PIN_CELL=core
export I_MPI_HYDRA_BOOTSTRAP=ssh
export I_MPI_DEBUG=7
export I_MPI_FABRICS=ofi
export I_MPI_OFI_PROVIDER=mlx
export HPL_DRIVER=0
hosts=(cn01 cn02)
mpirun -genvall \
   -hosts $(echo ${hosts[*]}|tr ' ' ',') \
   -perhost 4 -np 8 \
   bash hpl_wrapper.sh \
   ${ONEAPI_ROOT}/mkl/latest/share/mkl/benchmarks/mp_linpack/xhpl-ai_intel64_dynamic_gpu
