## Pytorch Bert-Large training on Intel Data Center GPU Max Series

### Prepare the Datasets
Please follow the [Datasets instructions](https://github.com/IntelAI/models/blob/master/quickstart/language_modeling/pytorch/bert_large/training/gpu/DEVCATALOG.md#datasets) to download and prepare for the Bert-Large training for Intel Data Center GPU   
In order to save disks, we can select a subset of datasets for our testing. For example, keep only 10 dataset out of 500 by copying part-0000*-of-00500 files into a folder /dataset/mlcommons_bert/results4.   
```bash
# export the DATASET-DIR env to the folder where the "results4" is located
export DATASET_DIR=/dataset/mlcommons_bert

# export a PROCESSED_DATASET_DIR env to a folder where the processed dataset can be stored, e.g.
export PROCESSED_DATASET_DIR=/dataset/mlcommons_bert_processed/

```
Please make sure you have the datasets prepared. The scripts can download the datasets and process the datasets automatically if the specified DATASET_DIR doesn't contain the right datasets. It will take long time to download depends on your network bandwidth. The processing of the datasets also takes time and large disk space (more than 530GB).   

### Run Pytorch training in Bare Metal
```bash
# Create a venv environment, make sure your python3 version >= 3.7,  e.g. 
python3 -m venv ipex_venv
source ipex_venv/bin/activate

# Setup the environments
./pt_bertlarge_training_setup.sh

# Run the training
./pt_bertlarge_training_run.sh

# Customize the environment for the training
PRECISION=<precision, default bf16>
BATCH_SIZE=<batch size to run, default 16>
NUMBER_OF_PROCESS=<# of process to run, default 2. Change to the nubmer of PVC stacks to run on multiple cards>
PROCESS_PER_NODE=<# of process per node, default 2. For single node system test, please set it same as NUMBER_OF_PROCESS>

# Sample command
# Run distributed training on 1 Max 1550 GPU
./pt_bertlarge_training_run.sh
# Run distributed training on 2 Max 1550 GPU
NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./pt_bertlarge_training_run.sh
# Run distributed training on 4 Max 1100 GPU with precision tf32
PRECISION=tf32 NUMBER_OF_PROCESS=4 PROCESS_PER_NODE=4 ./pt_bertlarge_training_run.sh

```

### Run Pytorch training in container
```bash
# Setup the container
./pt_bertlarge_training_container_setup.sh

# Run the training
./pt_bertlarge_training_container_run.sh

# The scripts accept the same environments as in bare metal.
# for example
# Run distributed training on 4 Max 1550 GPU
NUMBER_OF_PROCESS=8 PROCESS_PER_NODE=8 ./pt_bertlarge_training_container_run.sh

```

