#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

CONTAINER_NAME=${CONTAINER_NAME:-ptbertlargeinference}
IMAGE_NAME=intel/intel-extension-for-pytorch:${CONTAINER_NAME}
DOCKER_ARGS=${DOCKER_ARGS:--it --rm --name ${CONTAINER_NAME}}

# Use below environment to customize
export OUTPUT_DIR=${OUTPUT_DIR:-${WORKSPACE}/output/}
export DATASET_DIR=${DATASET_DIR:-${WORKSPACE}/SQuAD1.0/}
export BERT_WEIGHT=${BERT_WEIGHT:-${WORKSPACE}/bert_squad_model/}
export PRECISION=${PRECISION:-fp16}
export BATCH_SIZE=${BATCH_SIZE:-64}
export WARMUP_STEPS=${WARMUP_STEPS:-5}
export STEPS=${STEPS:-25}
export NUMBER_OF_GPU=${NUMBER_OF_GPU:-1}
export DEVICEID=${DEVICEID:-0}

#Max 1550 device id 0bd5, Max 1100 device id 0dba
device_type=$( lspci -nn |grep -i Display |grep -i 0bd5)
if [ ! -z "$device_type" ]; then
    export Tile=${Tile:-2}
else
    export Tile=${Tile:-1}
fi

cd $WORKSPACE
mkdir -p $DATASET_DIR
cd $DATASET_DIR
if [ ! -f train-v1.1.json ]; then
    echo "Download dataset, SQuAD1.0 for Bert large inference..."
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/train-v1.1.json
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v1.1.json
    wget https://github.com/allenai/bi-att-flow/blob/master/squad/evaluate-v1.1.py
fi

cd $WORKSPACE
echo "clean up the folder $OUTPUT_DIR"
if [ ! -d "$OUTPUT_DIR" ]; then
    mkdir -p $OUTPUT_DIR
else
    rm -rf $OUTPUT_DIR
    mkdir -p $OUTPUT_DIR
fi

cd $WORKSPACE
mkdir -p $BERT_WEIGHT
cd $BERT_WEIGHT
if [ ! -f config.json ]; then
    echo "Donwload pre-trained models in container"
    wget -c https://huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad/resolve/main/config.json
    wget -c https://huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad/resolve/main/pytorch_model.bin
    wget -c https://huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad/resolve/main/tokenizer.json
    wget -c https://huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad/resolve/main/tokenizer_config.json
    wget -c https://huggingface.co/bert-large-uncased-whole-word-masking-finetuned-squad/resolve/main/vocab.txt
fi

cd $WORKSPACE

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"

docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --privileged \
  --device=/dev/dri \
  --ipc=host \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --env DATASET_DIR=${DATASET_DIR} \
  --env OUTPUT_DIR=${OUTPUT_DIR} \
  --env BERT_WEIGHT=${BERT_WEIGHT} \
  --env Tile=${Tile} \
  --env PRECISION=${PRECISION} \
  --env BATCH_SIZE=${BATCH_SIZE} \
  --env WARMUP_STEPS=${WARMUP_STEPS} \
  --env STEPS=${STEPS} \
  --env NUMBER_OF_GPU=${NUMBER_OF_GPU} \
  --env DEVICEID=${DEVICEID} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume ${BERT_WEIGHT}:${BERT_WEIGHT} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  ${IMAGE_NAME} \
  /bin/bash /workspace/pt_bertlarge_inference_run.sh
