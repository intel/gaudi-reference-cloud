# Pytorch Resnet50 v1.5 Training on Intel Data Center GPU Max Series

This guide will guide the user to build and run Resnet50 v1.5 training using Intel GPU Max.

## Prepare System

The host server should have ***Intel GPU driver*** and ***docker engine*** or ***oneAPI kits*** installed. Please follow [System Setup](../../../../../system-setup/README.md) guide or below official guide:
1. [Intel GPU driver](https://dgpu-docs.intel.com/driver/installation.html)
2. [Install Docker Engine](https://docs.docker.com/engine/install/) for test in container
3. [Install Intel oneAPI Toolkits](https://www.intel.com/content/www/us/en/docs/oneapi/installation-guide-linux/2023-2/overview.html) for test on Bare Metal
```
# Make the .sh scripts exectuable
chmod +x *.sh

# If you want to run with docker container, make sure the current user can run docker container. 
# You may need to add the current user to 'docker' group and relogin to system if needed.
sudo usermod -aG docker $USER

# Set cpu scaling governor to 'performance' for better performance
sudo echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

```

## Prepare the Datasets

Imagenet 2012 Dataset must be downloaded by the user. By default without the dataset, the scripts will use the Synthetic/Dummy ImageNet2012 dataset for test.

For detailed instructions please refer to PyTorch [example](https://github.com/pytorch/examples/tree/main/imagenet#requirements)

1. Download ILSVRC2012_img_train.tar & ILSVRC2012_img_val.tar from https://image-net.org/index.php. Following command may be working for downloading dataset.
```
wget https://image-net.org/data/ILSVRC/2012/ILSVRC2012_img_train.tar --no-check-certificate
wget https://image-net.org/data/ILSVRC/2012/ILSVRC2012_img_val.tar --no-check-certificate
```
2. Navigate to the directory containing ILSVRC2012_img_train.tar & ILSVRC2012_img_val.tar
3. Download [the following shell script](https://github.com/pytorch/examples/blob/main/imagenet/extract_ILSVRC.sh) to the directory containing ILSVRC2012_img_train.tar & ILSVRC2012_img_val.tar.


```
wget -O extract_ILSVRC.sh https://raw.githubusercontent.com/pytorch/examples/main/imagenet/extract_ILSVRC.sh
ls
#<current directory>
# ├── ILSVRC2012_img_train.tar
# ├── ILSVRC2012_img_val.tar
# ├── extract_ILSVRC.sh
```
4. Extract the training and validation images to labeled subfolders using 
```
bash extract_ILSVRC.sh
#<current directory>
# ├── ILSVRC2012_img_train.tar
# ├── ILSVRC2012_img_val.tar
# ├── extract_ILSVRC.sh
```
5. After extraction, set env variable DATASET_DIR with the path to the folder that contains the ***val*** and ***train*** directories. This env variable must be set before running the quickstart scripts. 
```
ls
# imagenet/train/
#  ├── n01440764
#  │   ├── n01440764_10026.JPEG
#  │   ├── n01440764_10027.JPEG
#  │   ├── ......
#  ├── ......
#  imagenet/val/
#  ├── n01440764
#  │   ├── ILSVRC2012_val_00000293.JPEG
#  │   ├── ILSVRC2012_val_00002138.JPEG
#  │   ├── ......
#  ├── ......
export DATASET_DIR=<path to dataset>/imagenet/
```


## Run the test
### Option 1: Run training using Container


1. Setup Container
```
./pt_resnet50_training_container_setup.sh
```

2. Run the test
Run test with Synthetic/Dummy ImageNet2012 dataset   
```
# Run on 1 Max 1550 GPU, 2 ranks DDP training, default PRECISION=bf16
EPOCHS=1 NUMBER_OF_PROCESS=2 PROCESS_PER_NODE=2 ./pt_resnet50_training_container_run.sh

# Run on 8 Max 1550 GPU, 16 ranks DDP training, default PRECISION=bf16
EPOCHS=2 NUMBER_OF_PROCESS=2 PROCESS_PER_NODE=2 ./pt_resnet50_training_container_run.sh

```

To run on the entire dataset, 
a.  Run on 2 stacks(tiles) with epoch 1, bf16 and ImageNet dataset.
```
DATASET_DIR=<path to dataset>/imagenet/ DATASET_DUMMY=0 EPOCHS=1 NUMBER_OF_PROCESS=2 PROCESS_PER_NODE=2 PRECISION=bf16 ./pt_resnet50_training_container_run.sh
```

b.  Run on 8 stacks(tiles) with epoch 1, bf16 and ImageNet dataset.
```
DATASET_DIR=<path to dataset>/imagenet/ DATASET_DUMMY=0 EPOCHS=1 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 PRECISION=bf16 ./pt_resnet50_training_container_run.sh
```




### Option 2: Run Training in Bare Metal 
1. Setup Python VENV
```
VENV_NAME=pt_venv ./pt_resnet50_training_setup.sh
```

2. Run the test

a. Run on 2 stacks(tiles) with epoch 1, bf16 and imagenet dataset
```
VENV_NAME=pt_venv DATASET_DIR=<path to dataset>/imagenet/ DATASET_DUMMY=0 EPOCHS=1 NUMBER_OF_PROCESS=2 PROCESS_PER_NODE=2 PRECISION=bf16 ./pt_resnet50_training_run.sh
```

b. Run on 8 stacks(tiles) with epoch 1, bf16 and imagenet dataset
```
VENV_NAME=pt_venv DATASET_DIR=<path to dataset>/imagenet/ DATASET_DUMMY=0 EPOCHS=1 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 PRECISION=bf16 ./pt_resnet50_training_run.sh 
```


## Multiple Nodes Training

Here is an example of multiple nodes training through SLURM workload manager. In test cluster, every nodes contains 4 Max1550 GPU.

sbatch_job.sh and work.sh are created for launching training on multiple nodes.

### sbatch_job.sh

Below is a sample command line to launch training on 4 nodes which are sorted. If the nodes aren't in order, there could be some unexpected error . 

```
./sbatch_job.sh compute11,compute12,compute13,compute14
```

In sbatch_job.sh, we will dump the nodes list to hostfile which will be passed to MPI and PyTorch inferface. This script will count the nodes number for allocating resource. 

```
#!/bin/bash

nodes=$1
jobname=$(basename $PWD)

echo "generate hostfile from $1"
echo $1 | tr "," "\n" > hostfile

nodes_num=$(wc -l < hostfile)
echo "Required Nodes: ${nodes_num}"

sbatch \
  -N ${nodes_num} \
  -w ${nodes} \
  --job-name=${jobname} \
  --output=$(date +"%Y%m%d")-${nodes}-${jobname}-$(date +"%s").txt \
  --export=node=${nodes} \
  work.sh
```

### work.sh
This script is for launch training script. It derives total required process number, NUMBER_OF_PROCESS, through nodes number and launch ./pt_resnet50_training_run.sh. 

The same as luanch pt_resnet50_training_run.sh directly, we are required to set other parameters for runnning, such as EPOCHS, DATASET_DIR, etc.

```
#!/bin/bash

module load  oneapi/2023.2

[ ! -z ${SLURM_SUBMIT_DIR} ] && cd ${SLURM_SUBMIT_DIR}

echo "BEGIN: `date`"
t0=$(date +%s)

MASTER_ADDR=$(head -n 1 hostfile)
echo "node: $node"
echo "master node: $MASTER_ADDR"

nodes=$(wc -l < hostfile)
echo "total nodes: $nodes, $((nodes * 8))"

PROCESS_PER_NODE=8
NUMBER_OF_PROCESS=$((nodes * PROCESS_PER_NODE))
echo "total process: $NUMBER_OF_PROCESS"

MASTER_ADDR=$MASTER_ADDR \
HOSTFILE=hostfile \
VENV_NAME=pt_venv EPOCHS=2 \
DATASET_DIR=/home/zhixuexi/workspace/imagenet DATASET_DUMMY=0 \
NUMBER_OF_PROCESS=$NUMBER_OF_PROCESS PROCESS_PER_NODE=$PROCESS_PER_NODE \
PRECISION=bf16 \
./pt_resnet50_training_run.sh

t1=$(date +%s)
echo "INFO: Cost $[t1-t0] secs"
echo "END: `date`"
```


