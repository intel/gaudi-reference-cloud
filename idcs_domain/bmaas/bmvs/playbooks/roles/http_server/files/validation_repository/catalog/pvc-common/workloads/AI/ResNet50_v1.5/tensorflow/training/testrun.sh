#!/bin/bash

# This is simple scripts sample for testing purpose after you have the right traning conda env or python venv environment.

CURRENT_DIR=`pwd`
# checkpoint output folder
MODEL_DIR=${CURRENT_DIR}/output
# dataset folder, need update to the dataset location in current system
DATA_DIR=/data0/imagenet_raw_data/tf_records
# yaml config file for training. copy default from hvd_configure folder
# e.g. itex_bf16_lars.yaml is for BF16 training
# update the yaml file if needed. e.g. batch size, epochs etc.
CONFIG_FILE=${CURRENT_DIR}/itex_run.yaml

# clean up the output folder
if [ ! -d "$MODEL_DIR" ]; then
    mkdir -p $MODEL_DIR
else
    rm -rf $MODEL_DIR && mkdir -p $MODEL_DIR
fi

#source the oneapi environment
source /opt/intel/oneapi/setvars.sh
# alternativly use below for minimal environment if needed
#source /opt/intel/oneapi/compiler/latest/env/vars.sh
#source /opt/intel/oneapi/mkl/latest/env/vars.sh
#source /opt/intel/oneapi/ccl/latest/env/vars.sh
 
#set the PYTHONPATH to tensorflow-models
export PYTHONPATH=${CURRENT_DIR}/tensorflow-models

#define the number of process, process per node. Use same number for single node test.
NUMBER_OF_PROCESS=16
PROCESS_PER_NODE=16

# run the training
mpirun -np $NUMBER_OF_PROCESS -ppn $PROCESS_PER_NODE --prepend-rank \
	python ${PYTHONPATH}/official/vision/image_classification/classifier_trainer.py \
	--mode=train_and_eval \
	--model_type=resnet \
	--dataset=imagenet \
	--model_dir=$MODEL_DIR \
	--data_dir=$DATA_DIR \
	--config_file=$CONFIG_FILE 2>&1 | tee mytest.log
