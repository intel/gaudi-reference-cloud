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

