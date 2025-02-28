## Pytorch Bert-Large inference on Intel Data Center GPU Max Series

## Run Pytorch inference model in Bare Metal

### Default Env and Configuration
```
# Run the setup script 
./pt_bertlarge_inference_setup.sh

# Excute the run script
./pt_bertlarge_inference_run.sh

```
### Customize the environment for inference
#### The run script accept the following envs to customize the inference workload
```
VENV_NAME=<venv name, use current env if not set>
DATASET_DIR=<path to dataset>
PRECISION=<precision, default fp16>
BATCH_SIZE=<batch size to run, default 64>
NUMBER_OF_GPU=<# of GPU card to run inference in parallel, default 1. The inference process on each GPU equals the number of Tiles/Stacks per GPU.>
DEVICEID=<the device ID to run inference, default 0.>

# For example
# Run inference on 1 Intel Data Center GPU Max card
./pt_bertlarge_inference_run.sh

# Run inference on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_run.sh

# Run fp16 inference with batch size 32 on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 PRECISION=fp16 BATCH_SIZE=32 ./pt_bertlarge_inference_run.sh
```

## Run Pytorch inference model in container
### Default Env and Configuration
```
# Run the setup script
./pt_bertlarge_inference_container_setup.sh

# Run the inference
./pt_bertlarge_inference_container_run.sh
```
### Customize the environment for inference
## The scirpt accepts the same environments as script for bare metal
```
# For example
# Run inference on 1 Intel Data Center GPU Max card
./pt_bertlarge_inference_container_run.sh

# Run inference on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_container_run.sh

# Run fp16 inference with batch size 32 on 4 Intel Data Center GPU Max cards
NUMBER_OF_GPU=4 PRECISION=fp16 BATCH_SIZE=32 ./pt_bertlarge_inference_container_run.sh

```
## Quick Start Scripts
### Quick Start at Bare-metal Env
```
#Setup Env and run only once
./pt_bertlarge_inference_setup.sh

#Max_1550
Tile=1 NUMBER_OF_GPU=1 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_1C1T_Log.txt
NUMBER_OF_GPU=1 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_1C2T_Log.txt
NUMBER_OF_GPU=2 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_2C4T_Log.txt
NUMBER_OF_GPU=3 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_3C6T_Log.txt
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_4C8T_Log.txt
#NUMBER_OF_GPU=8 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1550_8C16T_Log.txt

#Max_1100
NUMBER_OF_GPU=1 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1100_1C1T_Log.txt
NUMBER_OF_GPU=2 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1100_2C2T_Log.txt
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1100_4C4T_Log.txt
NUMBER_OF_GPU=6 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1100_6C6T_Log.txt
NUMBER_OF_GPU=8 ./pt_bertlarge_inference_run.sh 2>&1 | tee -a Max1100_8C8T_Log.txt
```

### Quick Start at Container Env
```
#Setup Env and run only once
./pt_bertlarge_inference_container_setup.sh

#Max_1550
Tile=1 NUMBER_OF_GPU=1 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_1C1T_Log.txt
NUMBER_OF_GPU=1 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_1C2T_Log.txt
NUMBER_OF_GPU=2 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_2C4T_Log.txt
NUMBER_OF_GPU=3 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_3C6T_Log.txt
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_4C8T_Log.txt
#NUMBER_OF_GPU=8 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1550_8C16T_Log.txt

#Max_1100
NUMBER_OF_GPU=1 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1100_1C1T_Log.txt
NUMBER_OF_GPU=2 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1100_2C2T_Log.txt
NUMBER_OF_GPU=4 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1100_4C4T_Log.txt
NUMBER_OF_GPU=6 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1100_6C6T_Log.txt
NUMBER_OF_GPU=8 ./pt_bertlarge_inference_container_run.sh 2>&1 | tee -a Max1100_8C8T_Log.txt

```

## If you'd like to mannually set up the requied and environments and model, see this section.

### Pre-requisite
Please set the environment variables as below before running the setup script(`pt_bertlarge_inference_setup.sh`).


For more information on datasets for Bert-Large inference, see the [Datasets instructions](https://github.com/IntelAI/models/tree/master/quickstart/language_modeling/pytorch/bert_large/inference/gpu).
```
# export a DATASET-DIR env to the folder where the "dev-v1.1.json", "evaluate-v1.1.py" and "train-v1.1.json" are located, e.g.
export DATASET_DIR=$PWD/dataset

# export a VENV_NAME env for virtual env name, e.g.
export VENV_NAME=ipex_py310

# export the BERT_WEIGHT environment variable, e.g.
export BERT_WEIGHT=/home/user1/benchmarks/bert_squad_model/

# export the number of tile, e.g.
export Tile=1

# export a OUTPUT_DIR env to a folder where the logs can be stored, e.g.
export OUTPUT_DIR=$PWD/logs
### Create a virtual environment for Python
Install the following pre-requisites
* Create and activate virtual environment
  ```bash
  conda create -n ipex_py310 python=3.10
  conda activate ipex_py310
  ```
* Install required libs
  ```bash
  pip install torch==1.13.0a0+git6c9b55e torchvision==0.14.1a0 intel_extension_for_pytorch==1.13.120+xpu oneccl_bind_pt==1.13.200+gpu -f https://developer.intel.com/ipex-whl-stable-xpu
  ```

### Pre-trained Model
Download the `config.json` and fine tuned model from huggingface and set the `BERT_WEIGHT` environment variable to point to the directory that has both files:
  ```
  mkdir bert_squad_model
  wget https://s3.amazonaws.com/models.huggingface.co/bert/bert-large-uncased-whole-word-masking-finetuned-squad-config.json -O bert_squad_model/config.json
  wget https://cdn.huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad-pytorch_model.bin  -O bert_squad_model/pytorch_model.bin
  BERT_WEIGHT=<workspace>/bert_squad_model
  ```

### Download Model Zoon and configure
* Clone the Model Zoo repository
  ```bash
  git clone https://github.com/IntelAI/models.git
  ```
* Navigate models directory and install model specific dependencies for the workload:
  ```bash
  # Navigate to the model zoo repo
  cd models
  # Install model specific dependencies
  python -m pip install -r models/language_modeling/pytorch/bert_large/inference/gpu/requirements.txt
  ```
