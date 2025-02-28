#!/bin/bash

VENV_NAME=${VENV_NAME}
# Python
PYTHON=${PYTHON:-python3}
if [ "${VENV_NAME}x" != "x" ]; then
   echo "Create the new VENV: ${VENV_NAME}"
   if [ $(which conda) ]; then
      conda create -y -n ${VENV_NAME} python=3.10
      eval "$(conda shell.bash hook)"
      conda activate ${VENV_NAME}
      conda install -c conda-forge -y libstdcxx-ng=12
   elif [  $(which $PYTHON) ]; then
      echo "Using Python venv to create env $VENV_NAME"
      #sudo apt install python3-virtualenv
      #virtualenv -p python ${VENV_NAME}
      $PYTHON -m venv ${VENV_NAME}
      source ${VENV_NAME}/bin/activate
    else
      echo "$PYTHON not found, please install $PYTHON first"
      exit 1
    fi
fi

echo "Python version: $(${PYTHON} --version)"
IFS=. read -r maj_v_1 maj_v_2 min_v <<< $(${PYTHON} --version | cut -d" " -f2)
echo "Python major verion: ${maj_v_1}.${maj_v_2}"
if [ $maj_v_2 -lt 8 ] || [ $maj_v_2 -gt 11 ]; then
	echo "python3.8 - python3.11 is required"
	echo "Please assign python with env: PYTHON=python3.9 or PYTHON=python3.10, etc."
	exit 1
fi

# use scripts folder as WORKSPACE by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Prepare the source codes
wget https://github.com/IntelAI/models/raw/v2.12.0/models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py -O main.py
git apply FixedFileExistsError.patch
if [ $? -eq 0 ]; then
	echo "Successfully download code"
else
	echo "Failed to prepare codes"
  exit 1
fi

# Install PyTorch software stack
python -m pip install torch==2.1.0a0 torchvision==0.16.0a0 torchaudio==2.1.0a0 intel-extension-for-pytorch==2.1.10+xpu oneccl_bind_pt==2.1.100 --extra-index-url https://pytorch-extension.intel.com/release-whl/stable/xpu/us/
python -m pip install tensorboard

# Active venv
ONEAPI_ROOT=${ONEAPI_ROOT:-/opt/intel/oneapi}
source ${ONEAPI_ROOT}/setvars.sh
python -c "import torchvision; torchvision.models.resnet50(pretrained=True)"
