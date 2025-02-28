# Pytorch Resnet50 v1.5 Inference on Intel Data Center GPU Max Series

This document provides instructions for running Resnet50 v1.5 inference using Intel-optimized Pytorch with Intel® Data Center GPU.

The instructions here provide further detailed steps with automated scripts to run the training benchmarks with various configurations. It is based on the [Intel Extension for Pytorch](https://intel.github.io/intel-extension-for-pytorch/xpu/2.1.10+xpu/) [v2.1.10+xpu] release.

## Prepare Host

Host server should be install ***Intel GPU driver*** and ***docker engine*** or ***oneAPI kits***. Please following official guide:
1. [Intel GPU driver](https://dgpu-docs.intel.com/driver/installation.html)
2. [Install Docker Engine](https://docs.docker.com/engine/install/) for test in container
3. [Install Intel oneAPI Toolkits](https://intel.github.io/intel-extension-for-pytorch/index.html#installation?platform=gpu&version=v2.1.10%2Bxpu) for test on Bare Metal. For example on Ubuntu:
```bash
sudo apt install -y intel-oneapi-dpcpp-cpp-2024.0 intel-oneapi-mkl-devel=2024.0.0-49656
```

## Prepare the Datasets
For Resnet50 inference benchmarks with Intel Extension for Pytorch(IPEX), both dummy synthetic dataset and real dataset are supported.
DATASET_DUMMY is enabled by default and dummy synthetic dataset will be used
If real dataset is planned, pls set "DATASET_DUMMY=0"

Here is the steps to download imagenet dataset and process the dataset:
1. Download the ImageNet 2012 dataset from http://www.image-net.org/
2. Download the extraction script
```bash
wget -q https://raw.githubusercontent.com/pytorch/examples/main/imagenet/extract_ILSVRC.sh
```
Your folder should have the following structure:
```
<current directory>
 ├── ILSVRC2012_img_train.tar
 ├── ILSVRC2012_img_val.tar
 ├── extract_ILSVRC.sh
```
3. Prepare the dataset: Execute the `extract_ILSVRC.sh` script to move and extract the training and validation images to labeled subfolders.
```bash
export PARENT_DIR=$(pwd)
bash extract_ILSVRC.sh
```
4. The ***imagenet*** folder that contains the ***val*** and ***train*** directories should be set as the DATASET_DIR environment variable before running the quickstart scripts. 
```
#Example
export DATASET_DIR=$PARENT_DIR/imagenet/
```
## Run Pytorch Resnet50 Infernce in Bare Metal
### Create a conda or python venv environment
* If you have conda/miniconda installed
```bash
conda create -y -n pt_resnet python=3.10
conda activate pt_resnet
```

* or use python venv to create a virtual environment
```bash
sudo apt install python3-virtualenv
python3 -m venv pt_resnet
source pt_resnet/bin/activate
```
### Install requirements and setup the pytorch inference environment in one time
```bash
./pt_resnet50_inference_setup.sh
```

### Run the Resnet50 pytorch inference
```bash
./pt_resnet50_inference_run.sh

```
### Customize the configs for inference
#### The run script accept the following envs to customize the inference workload
```bash
VENV_NAME=<venv name>
PRECISION=<precision, default bfloat16>
BATCH_SIZE=<batch size to run, default 256>
NUMBER_OF_GPU=<# of GPU to run, default 1. Change to the numbber of GPU card to run on multiple cards>
DATASET_DIR=<dataset dir>
OUTPUT_DIR=<Output dir, default to output under current dir>
DATASET_DUMMY=<whether to use dummy dataset, default 1, dummy dataset will be used>
ITEM_NUM=<inference num, default 100>

Pls note:
DATASET_DIR will be ignored if DATASET_DUMMY is set, then dummy dataset will be used

* For examplea
* Run Inference on 1x Intel Data Center Max1550 with int8/batch_size 1024 and dummy dataset, total 2 inference processes running
NUMBER_OF_GPU=1 DATASET_DUMMY=1 BATCH_SIZE=1024 PRECISION=int8 ./pt_resnet50_inference_run.sh

* Run Inference on 8x Intel Data Center Max1550 with fp16/batch_size 256 and dummy dataset, total 16 inference processes running 
NUMBER_OF_GPU=8 DATASET_DUMMY=1 BATCH_SIZE=256 PRECISION=fp16 ./pt_resnet50_inference_run.sh

* Run Inference on 1 GPU(from device index 0) 2 stacks(tiles) - Intel Data Center Max1550 with int8/batch_size 1024 and imagenet dataset 
VENV_NAME=pt_venv NUMBER_OF_GPU=1 DATASET_DUMMY=0 BATCH_SIZE=1024 PRECISION=int8 DEVICEID=0 DATASET_DIR=$PARENT_DIR/imagenet/ ./pt_resnet50_inference_run.sh
```

## Run Pytorch Resnet50 inference in Container
### Setup the Docker Image
#### Run the scripts to pull the docker image and setup the inference workload
```bash
./pt_resnet50_inference_container_setup.sh
```

### Run the inference in Container
#### The script will use the docker image created during setup and run the workload
```bash
./pt_resnet50_inference_container_run.sh
```

### Cutomize the configs for inference
### The scirpt accepts the same environments as script for bare metal, e.g.
```bash
* Run inference on 2 Intel Data Center GPU Max 1550 card from container, total 4 inference process running
NUMBER_OF_GPU=2 ./pt_resnet50_inference_container_run.sh
```

## References
1. Pytorch [example](https://github.com/pytorch/examples/tree/main/imagenet#requirements) for building Imagenet Dataset.
2. Imagenet extraction [shell script](https://github.com/pytorch/examples/blob/main/imagenet/extract_ILSVRC.sh)
3. Intel ModelZoo [release](https://github.com/IntelAI/models/tree/r2.11/quickstart/image_recognition/pytorch/resnet50v1_5/training/gpu)
