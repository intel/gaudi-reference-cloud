# HPL-AI
HPL-AI is part of Intel® oneAPI Math Kernel Library (oneMKL) since oneAPI 2024.0 release.   
If you have oneAPI basekit installed in default system location, the binary for GPU (xhpl-ai_intel64_dynamic_gpu) can be found from /opt/intel/oneapi/mkl/latest/share/mkl/benchmarks/mp_linpack/. The binary can also be downloaded from [oneMKL Benchmarks Suite](https://www.intel.com/content/www/us/en/developer/articles/technical/onemkl-benchmarks-suite.html), including the prior version of oneMKL.   
Use below instruction to download the oneMKL benchmark suite and setup the PATH for HPL-AI
```
# For oneAPI 2023.2.0 version
wget https://internal-placeholder.com/781888/l_onemklbench_p_2023.2.0_49340.tgz
tar xf l_onemklbench_p_2023.2.0_49340.tgz
export PATH=$PATH:`pwd`/benchmarks_2023.2.0/linux/mkl/benchmarks/mp_linpack/

# For oneAPI 2024.0 version
wget https://internal-placeholder.com/793598/l_onemklbench_p_2024.0.0_49515.tgz
tar xf l_onemklbench_p_2024.0.0_49515.tgz
export PATH=$PATH:`pwd`/benchmarks_2024.0/linux/share/mkl/benchmarks/mp_linpack/
```

# Benchmark
Go to the corresponding sub folder and launch the run.sh for benchmarking.
The sripts may be customized and fined-tuned based on different system configuration.
- [max1550x1](./max1550x1/): Run HPL-AI on single Intel Max 1550 GPU.
- [max1550x8](./max1550x8/): Run HPL-AI on system with 8 Intel Max 1550 GPU with two MPI processes. Each process runs with four GPUs in the same NUMA node.
- [max1550x16-2N](./max1550x16-2N/): Run HPL-AI on two nodes, each nodes has 8 Intel Max 1550 GPUs. There will be 8 MPI ranks totally, 4 ranks per node. Each rank runs with two GPUs(four tiles) in the same NUMA node and bind CPU core and IB HCA nic accordingly.
