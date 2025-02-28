## Tensorflow Resnet50 v1.5 inference on Intel Data Center GPU Max Series

### Prepare the Datasets
Please follow the [Datasets instructions](https://github.com/IntelAI/models/blob/master/datasets/imagenet/README.md) to download and prepare the imagenet datasets.    
After running the conversion script you should have a directory with the ImageNet dataset in the TF records format.   
Set the DATASET_DIR to point to the TF records directory when running ResNet50 v1.5.  
Notes: The inference can use the synthetic dataset. By default, the synthetic dataset is used for benchmarking which the real dataset is not needed.      
In order to use the real imagenet datasets, please set the environment DATASET_DUMMY to 0   
Notes: The inference with real datasets includes the data load time so the performance of inference benchmarking can be significantly impacted by the system CPUs.   
In this case, the numactl can have significant impact on the performance data. It requires fine-tune for the numa binding in order to get a good performance.   
```
# set DATASET_DIR
export DATASET_DIR=/dataset/imagenet_data/tf_records
```

### Run Tensorflow Resnet50 Inference in Bare Metal
```
# For first run, please call the scripts tf_resnet50_inference_setup.sh to setup the environment
# 1. Create a conda or python venv environment
## If you have conda/miniconda installed,e.g.
conda create -n itex python=3.10
conda activate itex

## or use python venv to create a virtual environment
python3 -m venv itex
source itex/bin/activate

# 2. Install requirements and setup the tensorflow inference environment
./tf_resnet50_inference_setup.sh

# By default, the script uses current env for setup. It can accept the VENV_NAME environment to create and setup the env automatically
VENV_NAME=itex ./tf_resnet50_inference_setup.sh

# 3. Clean the output folder and Run the resnet50 inference
rm output/* -f
./tf_resnet50_inference_run.sh

# 4. Customize the environment for inference
## The run script accept the following envs to customize the inference workload
VENV_NAME=<venv name, use current venv if not set>
DATASET_DIR=<path to dataset>
PRECISION=<precision, default int8>
STEPS=<inference steps, default 25>
BATCH_SIZE=<batch size to run, default 1024>
NUMBER_OF_GPU=<# of GPU card to run inference in parallel, default 1. The inference process on each GPU equals the number of Tiles/Stacks per GPU.>
DEVICEID=<the device ID to run inference, default 0.>
DATASET_DUMMY=< if to use dummy dataset, default 1. Change to 0 to use the real imangenet dataset for inference>

# For example
# Run inference on 1 Intel Data Center GPU Max card
./tf_resnet50_inference_run.sh

# Run inference on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 ./tf_resnet50_inference_run.sh

# Run fp16 inference with batch size 256 on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 PRECISION=fp16 BATCH_SIZE=256 ./tf_resnet50_inference_run.sh

```


### Run Tensorflow Resnet50 inference in Container 
```
# Setup the Docker Image
## Run the scripts to pull the docker image and setup the inference workload 
./tf_resnet50_inference_container_setup.sh


# Run the inference in Container
## The script will use the docker image created during setup and run the workload
./tf_resnet50_inference_container_run.sh

# Cutomize the configs for inference
## The scirpt accepts the same environments as script for bare metal, e.g.
## Run inference in parallel with data type fp32 and batch size 1 on 2 Intel Data Center GPU Max card from container
NUMBER_OF_GPU=2 PRECISION=fp32 BATCH_SIZE=1 ./tf_resnet50_inference_container_run.sh

```


