# Tensorflow Resnet50 v1.5 training on Intel Data Center GPU Max Series
The [Resnet50 Training Example](https://github.com/intel/intel-extension-for-tensorflow/tree/v2.14.0.1/examples/train_resnet50) provides the overall instructions to run Resnet50 training on Intel Data Center GPU.   
The instructions here provide further detailed steps with automated scripts to run the training benchmarks with various configurations. It is based on the [Intel Extension for Tensorflow](https://github.com/intel/intel-extension-for-tensorflow) [v2.14.0.1](https://github.com/intel/intel-extension-for-tensorflow/releases/tag/v2.14.0.1) release.   

## Prepare the Datasets
For Resnet50 training benchmarks with Intel Extension for Tensorflow (ITEX), both synthetic dataset and real dataset are supported.   
In case you want to use the real ImageNet dataset for training, [Intel® AI Reference Models](https://github.com/IntelAI/models) provides the scripts to process the ImageNet Dataset for Tensorflow training. Please follow the [ImageNet Datasets instructions](https://github.com/IntelAI/models/blob/master/datasets/imagenet/README.md) to download and prepare the imagenet datasets.    
The instructions for preparing the dataset shown as below. This is required when you use real ImageNet dataset for training.

### Download the ImageNet Dataset
Go to the [ImageNet webpage] (https://image-net.org) with your account. Select "2012" from the list of available datasets, download the following tar files and save to the system.
- Training images (Task 1 & 2). 138GB. MD5: 1d675b47d978889d74fa0da5fadfb00e
- Validation images (all tasks). 6.3GB. MD5: 29b22e2961454d5413ddabcf34fc5622
```
# With a successful download, save the tar file in the system folder, for example, in the IMAGENET_RAW_DATA
export IMAGENET_RAW_DATA=<path to the imagenet raw dataset>
ls ${IMAGENET_RAW_DATA}
    ILSVRC2012_img_train.tar
    ILSVRC2012_img_val.tar
```
### Pre-Process the Dataset
Use Python venv or conda environment to pre-process the dataset. Here we use conda environment for example:
```
# Create and prepare the environment
conda create -n tf_env -y
conda activate tf_env
pip install intel-tensorflow
pip install -I urllib3
pip install wget

# Download the scripts for pre-processing
wget https://raw.githubusercontent.com/IntelAI/models/master/datasets/imagenet/imagenet_to_tfrecords.sh

# pre-process the entire dataset for training
bash imagenet_to_tfrecords.sh ${IMAGENET_RAW_DATA} training

# In above step, in case you see the error something like "Check failed: ret == 0 (11 vs. 0)Thread tf_ForEach creation via pthread_create() failed.",
# use numactl to limit the processing thread
numactl -N 0 bash imagenet_to_tfrecords.sh ${IMAGENET_RAW_DATA} training

# After the pre-process done, the tf_records folder can be found in the ${IMAGENET_RAW_DATA} folder.
# The folder should contains 1024 training files and 128 validation files.
# Set the DATASET_DIR for later training steps
export DATASET_DIR=${IMAGENET_RAW_DATA}/tf_records

```

## Run Tensorflow Resnet50 Training
The [Resnet50 Training Example](https://github.com/intel/intel-extension-for-tensorflow/tree/v2.14.0.1/examples/train_resnet50) from ITEX github repo provides the instructions to run Resnet50 training. Based on these instructions, we created the scripts for the out of box benchmarks for Resnet50 training on Intel Data Center GPU.
The scripts follows the instructions from [Resnet50 Training Example](https://github.com/intel/intel-extension-for-tensorflow/tree/v2.14.0.1/examples/train_resnet50) but customized for the convinent of the benchmarks with different configurations.
By default, the synthetic dataset is used for benchmarking. Follow the instructions to change the configs if needed.

### Run in Bare Metal
```
# 1. Create a conda or python venv environment
## If you have conda/miniconda installed
conda create -n itex python=3.10
conda activate itex

## or use python venv to create a virtual environment
sudo apt install python3-virtualenv
python3 -m venv itex
source itex/bin/activate

# 2. Install requirements and setup the tensorflow training environment.
./tf_resnet50_training_setup.sh

# By default, the script uses current env for setup. It can accept the VENV_NAME environment to create and setup the env automatically, e.g.
VENV_NAME=itex ./tf_resnet50_training_setup.sh

# 3. Run the resnet50 training, by default, it will use 1 GPU for the training with BF16 and run with 1 Epoch only
./tf_resnet50_training_run.sh

# 4. Customize the configs for training
## The run script accept the following envs to customize the training workload. It will update the training configs in yaml file and run the training.
VENV_NAME=<venv name, use current venv if not set>
DATASET_DIR=<path to dataset, the tf_records folder>
PRECISION=<precision, default bfloat16>
EPOCHS=<epochs to run, default 1>
BATCH_SIZE=<batch size to run, default 256>
NUMBER_OF_PROCESS=<# of process to run, default 2. Change to the nubmer of PVC stacks to run on multiple cards>
PROCESS_PER_NODE=<# of process per node, default 2. For single node system test, please set it same as NUMBER_OF_PROCESS>
DATASET_DUMMY=< if to use synthetic dataset, default 1. Change to 0 to use the real imangenet dataset for training>

# Examples
# Run training on 1 Intel Data Center GPU Max 1550. Default data type bfloat16, 1 epoch, synthetic dataset.
./tf_resnet50_training_run.sh

# Run training on 2 Intel Data Center GPU Max 1550 
NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./tf_resnet50_training_run.sh

# Run training on 4 Intel Data Center GPU Max 1550 with 4 EPOCHS
EPOCHS=4 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./tf_resnet50_training_run.sh

# Run training on 8 Intel Data Center GPU Max 1550 with 4 EPOCHS
EPOCHS=4 NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_resnet50_training_run.sh

# Run training on 8 Intel Data Center GPU Max 1550 with 4 EPOCHS and real Imagenet dataset
EPOCHS=4 DATASET_DUMMY=0 DATASET_DIR=<path to tf_records folder> NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_resnet50_training_run.sh

# Run training on 4 Intel Data Center GPU Max 1550 with 42 EPOCHS and real Imagenet dataset, reach to > 75.9% accuracy
EPOCHS=42 BATCH_SIZE=512 DATASET_DUMMY=0 DATASET_DIR=<path to tf_records folder> NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./tf_resnet50_training_run.sh

# Run training on 8 Intel Data Center GPU Max 1550 with 42 EPOCHS and real Imagenet dataset, reach to > 75.9% accuracy
EPOCHS=42 BATCH_SIZE=256 DATASET_DUMMY=0 DATASET_DIR=<path to tf_records folder> NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_resnet50_training_run.sh

```

### Run in Container
```
# 1. Build and setup the Docker Image
# Run the scripts to pull the docker image and setup the training working environment 
./tf_resnet50_training_container_setup.sh

# 2. Run the training in Container
./tf_resnet50_training_container_run.sh

# Customize the configs for training
# The scirpt accepts the same environments as script for bare metal, for example:

# Run training on 1 Intel Data Center GPU Max 1550 
NUMBER_OF_PROCESS=2 PROCESS_PER_NODE=2 ./tf_resnet50_training_container_run.sh

# Run training on 2 Intel Data Center GPU Max 1550 
NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./tf_resnet50_training_container_run.sh

# Run training on 4 Intel Data Center GPU Max 1550 with 4 EPOCHS
EPOCHS=4 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./tf_resnet50_training_container_run.sh

# Run training on 8 Intel Data Center GPU Max 1550 with 4 EPOCHS
EPOCHS=4 NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_resnet50_training_container_run.sh

# Run training on 8 Intel Data Center GPU Max 1550 with 4 EPOCHS and real Imagenet dataset
EPOCHS=4 DATASET_DUMMY=0 DATASET_DIR=<path to tf_records folder> NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_resnet50_training_container_run.sh
```


