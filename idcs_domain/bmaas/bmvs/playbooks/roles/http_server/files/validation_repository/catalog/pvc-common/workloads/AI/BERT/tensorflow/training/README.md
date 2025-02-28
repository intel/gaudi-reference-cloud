# Tensorflow Bert_Large training on Intel Data Center GPU Max Series
The [BERT Large Training Example](https://github.com/IntelAI/models/tree/v3.0.0/quickstart/language_modeling/tensorflow/bert_large/training/gpu) provides the overall instructions to run Bert Large training on Intel Data Center GPU.

The instructions here provide further detailed steps with automated scripts to run the training benchmarks with various configurations. It is based on the [Intel Extension for Tensorflow](https://github.com/intel/intel-extension-for-tensorflow) [v2.14.0.1](https://github.com/intel/intel-extension-for-tensorflow/releases/tag/v2.14.0.1) release.
HVD is enabled by defaut

## Prepare Pretrained model
Download and extract the bert large uncased (whole word masking) pre-trained model checkpoints from the [google bert repo](https://github.com/google-research/bert#pre-trained-models). The extracted directory should be set to the BERT_LARGE_DIR environment variable when running the scripts. 
A dummy dataset will be auto generated and used for training scripts.

```
# This scripts will download the pretrained models and unzip it automatically
./download_pretrained_models.sh
```

## Run Tensorflow Bert_Large Training in Bare Metal
### For first run, please call the scripts tf_bertlarge_training_setup.sh to setup the environment
### Create a conda or python venv environment
* If you have conda/miniconda installed
```
conda create -y -n tf_bert python=3.10
conda activate tf_bert
```

* or use python venv to create a virtual environment
```
sudo apt install python3-virtualenv
python3 -m venv tf_bert
source tf_bert/bin/activate
```

### Install requirements and setup the tensorflow training environment
```
./tf_bertlarge_training_setup.sh
```

### Run the Bert_Large Tensorlow bf16 training
```
./tf_bertlarge_training_run.sh
```

### Customize the configs for training
#### The run script accept the following envs to customize the training workload
```
VENV_NAME=<venv name, use current venv if not set>
PRECISION=<precision, default bfloat16>
BERT_LARGE_DIR=<Pretrained model location, try to search in current directory if not set>
OUTPUT_DIR=<Generated dummy dataset and train checkpoint outputs, use current directory if not set>
BATCH_SIZE=<batch size to run, default 32>
TILE=<Tile num, default 2 for Max1550 and 1 for Max1100>
NUMBER_OF_PROCESS=<# of process to run, default 2 for Max1550 and 1 for Max1100. Change to the nubmer of PVC stacks to run on multiple cards>
PROCESS_PER_NODE=<# of process per node, default 2 for Max1550 and 1 for Max1100. For single node system test, please set it same as NUMBER_OF_PROCESS>

* For example
* Run training on 1 Intel Data Center GPU Max 1550 card(1Card2Tile)
./tf_bertlarge_training_run.sh

* Run training on 2 Intel Data Center GPU Max 1550 card(1Card2Tile)
NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./tf_bertlarge_training_run.sh

* Run training on 8 Intel Data Center GPU Max 1550 card(1Card2Tile) with 480 train steps
NUM_TRAIN_STEPS=480 NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_bertlarge_training_run.sh

* Run training on 1 Intel Data Center GPU Max 1100 card(1Card1Tile)
NUMBER_OF_PROCESS=1 PROCESS_PER_NODE=1 ./tf_bertlarge_training_run.sh

* Run training on 8 Intel Data Center GPU Max 1100 card(1Card1Tile) with 240 train steps and batch_size 16
NUM_TRAIN_STEPS=240 BATCH_SIZE=16 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./tf_bertlarge_training_run.sh
```

## Run Tensorflow Bert_Large training in Container 
### Setup the Docker Image
#### Run the scripts to pull the docker image and setup the training workload 
```
./tf_bertlarge_training_container_setup.sh
```

### Run the training in Container
#### The script will use the docker image created during setup and run the workload
```
./tf_bertlarge_training_container_run.sh
```

### Cutomize the configs for training
#### The scirpt accepts the same environments as script for bare metal, e.g.
```
* Run training on 2 Intel Data Center GPU Max 1550 card(1Card2Tile)
NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./tf_bertlarge_training_container_run.sh

* Run training on 8 Intel Data Center GPU Max 1550 card(1Card2Tile) with 480 train steps
NUM_TRAIN_STEPS=480 NUMBER_OF_PROCESS=16 PROCESS_PER_NODE=16 ./tf_bertlarge_training_container_run.sh

* Run training on 1 Intel Data Center GPU Max 1100 card(1Card1Tile)
NUMBER_OF_PROCESS=1 PROCESS_PER_NODE=1 ./tf_bertlarge_training_container_run.sh

* Run training on 8 Intel Data Center GPU Max 1100 card(1Card1Tile) with 240 train steps and batch_size 16
NUM_TRAIN_STEPS=240 BATCH_SIZE=16 NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./tf_bertlarge_training_container_run.sh

```
