#!/bin/bash

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}

cd $WORKSPACE
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
	echo "Using Python venv to create env"
	#sudo apt install python3-virtualenv
	#virtualenv -p python ${VENV_NAME}
        python3	-m venv ${VENV_NAME}
	source ${VENV_NAME}/bin/activate
  else
	echo "Python3 not found, please install python3 first"
	exit 1
  fi
fi

python3 -m pip install --upgrade pip
echo "Install Pytorch requirements..."
#python -m pip install torch==2.0.1a0 torchvision==0.15.2a0 intel_extension_for_pytorch==2.0.110+xpu oneccl_bind_pt==2.0.100+gpu -f https://developer.intel.com/ipex-whl-stable-xpu
python -m pip install torch==2.1.0a0 torchvision==0.16.0a0 intel-extension-for-pytorch==2.1.10+xpu oneccl_bind_pt==2.1.100 --extra-index-url https://pytorch-extension.intel.com/release-whl/stable/xpu/us/

cd $WORKSPACE
echo "Clone Intel AI models and install requirements for Bert large training..."
git clone https://github.com/IntelAI/models intelai-models
cd intelai-models
git checkout -b v2.12.0 v2.12.0
chmod 755 quickstart/language_modeling/pytorch/bert_large/training/gpu/*.sh
pip install -r models/language_modeling/pytorch/bert_large/training/gpu/requirements.txt
bash ./quickstart/language_modeling/tensorflow/bert_large/training/gpu/setup.sh
if [ ! -f models/language_modeling/pytorch/bert_large/training/gpu/data/vocab.txt ]; then
  wget https://s3.amazonaws.com/models.huggingface.co/bert/bert-base-uncased-vocab.txt -O models/language_modeling/pytorch/bert_large/training/gpu/data/vocab.txt
fi

cd ${WORKSPACE}
echo "Apply patch..."
cp intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/ddp_bf16_training_plain_format.sh intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/ddp_bf16_training_plain_format.sh.bak
cp intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/bf16_training_plain_format.sh intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/bf16_training_plain_format.sh.bak
cp ${WORKSPACE}/ddp_bf16_training_plain_format.sh intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/
cp ${WORKSPACE}/bf16_training_plain_format.sh intelai-models/quickstart/language_modeling/pytorch/bert_large/training/gpu/

echo "Setup done"
