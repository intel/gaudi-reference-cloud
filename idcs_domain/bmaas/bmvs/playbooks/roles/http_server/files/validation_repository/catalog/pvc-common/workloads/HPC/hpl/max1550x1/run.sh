ONEAPI_ROOT=/opt/intel/oneapi
source ${ONEAPI_ROOT}/setvars.sh
export PATH=$PATH:${ONEAPI_ROOT}/mkl/latest/share/mkl/benchmarks/mp_linpack/
export ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE
export ZE_ENABLE_PCI_ID_DEVICE_ORDER=1
export HPL_DRIVER=0
export I_MPI_OFFLOAD=0
HPL_HOST_NODE=0 HPL_DEVICE=:0.0,:0.1 mpirun -np 1 numactl -l xhpl_intel64_dynamic_gpu