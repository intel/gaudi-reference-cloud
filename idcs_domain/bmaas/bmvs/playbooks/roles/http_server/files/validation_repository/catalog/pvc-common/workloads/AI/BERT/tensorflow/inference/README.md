# Tensorflow Bert_Large inference on Intel Data Center GPU Max Series
This document has instructions for running BERT Large inference using Intel-optimized TensorFlow with Intel® Data Center GPU.

The instructions here provide further detailed steps with automated scripts to run the inference benchmarks with various configurations. It is based on the [Intel Extension for Tensorflow](https://github.com/intel/intel-extension-for-tensorflow) [v2.14.0.1](https://github.com/intel/intel-extension-for-tensorflow/releases/tag/v2.14.0.1) release.

## Prepare Pretrained model and Dataset

Download and unzip the BERT Large uncased (whole word masking) model from the [google bert repo](https://github.com/google-research/bert#pre-trained-models).
Then, download the Stanford Question Answering Dataset (SQuAD) dataset

* Download the frozen graph model file, and set the FROZEN_GRAPH environment variable to point to where it was saved:
  ```bash
  wget https://storage.googleapis.com/intel-optimized-tensorflow/models/v2_7_0/fp32_bert_squad.pb
  ```

* Download the pretrained model directory and set the PRETRAINED_DIR environment variable to point where it was saved:
  ```bash
  wget  https://storage.googleapis.com/bert_models/2019_05_30/wwm_uncased_L-24_H-1024_A-16.zip
  unzip wwm_uncased_L-24_H-1024_A-16.zip
  ```

* Download the SQUAD directory and set the SQUAD_DIR environment variable to point where it was saved:
  ```bash
  wget https://rajpurkar.github.io/SQuAD-explorer/dataset/train-v1.1.json
  wget https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v1.1.json
  wget https://raw.githubusercontent.com/allenai/bi-att-flow/master/squad/evaluate-v1.1.py
  ```

* Download all of the required pretrained models/dataset/frozen graph in one time:
  ```
  ./download.sh
  ```

## Run Tensorflow Bert_Large Training in Bare Metal
### For first run, please call the scripts tf_bertlarge_inference_setup.sh to setup the environment
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

### Install requirements and setup the tensorflow inference environment in one time
```
./tf_bertlarge_inference_setup.sh
```

### Run the Bert_Large Tensorlow fp16 inference
```
./tf_bertlarge_inference_run.sh
```
### Customize the configs for inference
#### The run script accept the following envs to customize the inference workload
```
VENV_NAME=<venv name, default tf_bert>
PRECISION=<precision, default fp16>
BATCH_SIZE=<batch size to run, default 64>
NUMBER_OF_GPU=<# of GPU to run, default 1. Change to the numbber of GPU card to run on multiple cards>
PRETRAINED_DIR=<Pretrained model dir, default to wwm_uncased_L-24_H-1024_A-16 under current dir>
OUTPUT_DIR=<Output dir, default to output under current dir>
FROZEN_GRAPH=<frozen graph pb file, default to fp32_bert_squad.pb under current dir>
SQUAD_DIR=<Dataset dir, default to SQuAD1.0 under current dir>

* For examplea
* Run inference on 1 Intel Data Center GPU Max 1550 card, total 2 inference process running
./tf_bertlarge_inference_run.sh

* Run inference on 8 Intel Data Center GPU Max 1550 card, total 16 inference process running, batch_size 64 with specific SQUAD_DIR
BATCH_SIZE=64 SQUAD_DIR=<path to SQuAD1.0 folder> NUMBER_OF_GPU=8 ./tf_bertlarge_inference_run.sh

* Run inference on 4 Intel Data Center GPU Max 1100 card, total 4 inference process running
NUMBER_OF_GPU=4 ./tf_bertlarge_inference_run.sh
```

## Run Tensorflow Bert_Large inference in Container 
### Setup the Docker Image
#### Run the scripts to pull the docker image and setup the inference workload 
```
./tf_bertlarge_inference_container_setup.sh
```

### Run the inference in Container
#### The script will use the docker image created during setup and run the workload
```
./tf_bertlarge_inference_container_run.sh
```

### Cutomize the configs for inference
### The scirpt accepts the same environments as script for bare metal, e.g.
```
* Run inference on 2 Intel Data Center GPU Max 1550 card from container, total 4 inference process running
NUMBER_OF_GPU=2 ./tf_bertlarge_inference_container_run.sh
```
