# HPL (High Performance LINKPACK)
HPL is part of Intel® oneAPI Math Kernel Library (oneMKL) since oneAPI 2024.0 release.   
If you have oneAPI basekit installed with default installation path, the Intel(R) Distribution for LINPACK* Benchmark binary for GPU (xhpl_intel64_dynamic_gpu) can be found from /opt/intel/oneapi/mkl/latest/share/mkl/benchmarks/mp_linpack/. The binary can also be downloaded from [oneMKL Benchmarks Suite](https://www.intel.com/content/www/us/en/developer/articles/technical/onemkl-benchmarks-suite.html), including the prior version of oneMKL.  

## Download the HPL benchmark package
Please visit [oneMKL Benchmarks Suite](https://www.intel.com/content/www/us/en/developer/articles/technical/onemkl-benchmarks-suite.html) Website to download the latest oneMKL benchmark package   
```
# For oneAPI 2024.0 version
wget https://internal-placeholder.com/793598/l_onemklbench_p_2024.0.0_49515.tgz
tar xf l_onemklbench_p_2024.0.0_49515.tgz
export PATH=$PATH:`pwd`/benchmarks_2024.0/linux/share/mkl/benchmarks/mp_linpack/

# For oneAPI 2023.2.0 version
wget https://internal-placeholder.com/781888/l_onemklbench_p_2023.2.0_49340.tgz
tar xf l_onemklbench_p_2023.2.0_49340.tgz
export PATH=$PATH:`pwd`/benchmarks_2023.2.0/linux/mkl/benchmarks/mp_linpack/
```
## Benchmark
Go to the corresponding sub folder and launch the run.sh for benchmarking.   
The sripts may be customized and fined-tuned based on different system configuration.
- [max1550x1](./max1550x1/): Run HPL on single Intel Max 1550 GPU.
- [max1550x8](./max1550x8/): Run HPL on system with 8 Intel Max 1550 GPUs with four MPI ranks. Each rank runs with two GPUs(four tiles) in the same NUMA node and bind CPU core accordingly.
- [max1550x16-2N](./max1550x16-2N/): Run HPL on two nodes, each nodes has 8 Intel Max 1550 GPUs. There will be 8 MPI ranks totally, 4 ranks per node. Each rank runs with two GPUs(four tiles) in the same NUMA node and bind CPU core and IB HCA nic accordingly.
  To get a better multiple nodes GPU HPL performance, knowing the hardware topology will be very important. Here is a script showing how to query the CPU, GPU and IB affinities.
  ```bash
  lscpu
  lspci -D | grep Infiniband | awk '{print $1}' | while read i
  do
    dev=$(ls -l /sys/class/infiniband | grep -o ${i}.* | cut -f3 -d'/')
    node=$(cat /sys/class/infiniband/${dev}/device/numa_node)
    echo "$(hostname) $i $dev $node"
  done
  lspci -D | grep Display | awk '{print $1}' | while read i
  do
    dev=$(ls -l /sys/class/drm | grep -o ${i}.*card.* | cut -f3 -d'/')
    node=$(cat /sys/class/drm/${dev}/device/numa_node)
    echo "$(hostname) $i $dev $node"
  done
  ```
  The final topology like:
    - NUMA node0: CPU[0-47],GPU[0-3],IB[mlx5_0,mlx5_1,mlx5_2,mlx5_5]
    - NUMA node1: CPU[48-95],GPU[4-7],IB[mlx5_6,mlx5_7,mlx5_8,mlx5_11]

## Reference
> https://www.intel.com/content/www/us/en/developer/articles/technical/onemkl-benchmarks-suite.html
> https://www.intel.com/content/www/us/en/docs/onemkl/developer-guide-linux/2023-2/intel-distribution-for-linpack-benchmark-and-intel.html
