#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
echo $WORKSPACE

cd $WORKSPACE
DATASET_DIR=${DATASET_DIR:-${WORKSPACE}/SQuAD1.0}
BERT_WEIGHT=${BERT_WEIGHT:-${WORKSPACE}/bert_squad_model}

# Set VENV_NAME environment to create and setup in the specified VENV
# if find conda, use conda create venv, else use python venv
# if VENV_NAME not set, use current environment

VENV_NAME=${VENV_NAME}

if [ "${VENV_NAME}x" != "x" ]; then
    echo "Create the new VENV: ${VENV_NAME}"
    if [ $(which conda) ]; then
        conda create -y -n ${VENV_NAME} python=3.10
        eval "$(conda shell.bash hook)"
        conda activate ${VENV_NAME}
        conda install -c conda-forge -y libstdcxx-ng=12
    elif [  $(which python3) ]; then
        python3 -m venv ${VENV_NAME}
        echo "Using Python venv ${VENV_NAME}"
        source ${VENV_NAME}/bin/activate
    else
        echo "Python3 not found, please install python3 and run setup first"
        exit 1
    fi
fi

python3 -m pip install --upgrade pip
echo "Install Pytorch requirements..."
#python -m pip install torch==2.0.1a0 torchvision==0.15.2a0 intel_extension_for_pytorch==2.0.110+xpu oneccl_bind_pt==2.0.100+gpu -f https://developer.intel.com/ipex-whl-stable-xpu
python -m pip install torch==2.1.0a0 torchvision==0.16.0a0 intel-extension-for-pytorch==2.1.10+xpu oneccl_bind_pt==2.1.100 --extra-index-url https://pytorch-extension.intel.com/release-whl/stable/xpu/us/

cd $WORKSPACE
mkdir -p $DATASET_DIR
echo "Download dataset, SQuAD1.0 for Bert large inference..."
cd $DATASET_DIR
if [ ! -f train-v1.1.json ]; then
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/train-v1.1.json
    wget https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v1.1.json
    wget https://github.com/allenai/bi-att-flow/blob/master/squad/evaluate-v1.1.py
fi

cd $WORKSPACE
echo "Clone the Model Zoo repository for Bert large inference..."
if [ ! -f models/README.md ]; then
    git clone https://github.com/IntelAI/models.git
fi

# Install model specific dependencies
cd models
git checkout v2.12.0
python -m pip install -r models/language_modeling/pytorch/bert_large/inference/gpu/requirements.txt
cp quickstart/language_modeling/pytorch/bert_large/inference/gpu/fp16_inference_plain_format.sh quickstart/language_modeling/pytorch/bert_large/inference/gpu/fp16_inference_plain_format.sh.bak
cp ${WORKSPACE}/fp16_inference_plain_format.sh quickstart/language_modeling/pytorch/bert_large/inference/gpu/
git apply ../pt_bert_infer.patch

echo "Donwload pre-trained models"
if [ ! -f $BERT_WEIGHT/pytorch_model.bin ]; then
    cd models/language_modeling/pytorch/bert_large/inference/gpu
    ./download_squad_large_fine_tuned_model.sh
    cd $WORKSPACE
    mv models/models/language_modeling/pytorch/bert_large/inference/gpu/squad_large_finetuned_checkpoint $BERT_WEIGHT
else
    echo "Find Bert-large model weight in folder $BERT_WEIGHT, skip downloading"
fi

echo "Setup done"
