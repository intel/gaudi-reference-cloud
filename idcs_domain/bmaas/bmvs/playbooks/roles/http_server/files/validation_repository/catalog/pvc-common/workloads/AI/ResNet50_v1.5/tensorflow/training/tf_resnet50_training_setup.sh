#!/bin/bash
# Copyright (c) 2023 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

# Use script dir for workspace by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
cd ${WORKSPACE}

# Set VENV_NAME environment to create and setup in the specified VENV
# if find conda, use conda create venv, else use python venv
# if VENV_NAME not set, use current environment
VENV_NAME=${VENV_NAME}
PYTHON=${PYTHON:-python3}
if [ "${VENV_NAME}x" != "x" ]; then
  echo "Create the new VENV: ${VENV_NAME}"
  if [ $(which conda) ]; then
	conda create -y -n ${VENV_NAME} python=3.10
	eval "$(conda shell.bash hook)"
	conda activate ${VENV_NAME}
  elif [  $(which $PYTHON) ]; then
	echo "Using Python venv to create env $VENV_NAME"
	#sudo apt install python3-virtualenv 
	#virtualenv -p python ${VENV_NAME}
        $PYTHON	-m venv ${VENV_NAME}
	source ${VENV_NAME}/bin/activate
  else
	echo "$PYTHON not found, please install $PYTHON first"
	exit 1
  fi
fi

#check python version
python_ver_major=$($PYTHON -c"import sys; print(sys.version_info.major)")
python_ver_minor=$($PYTHON -c"import sys; print(sys.version_info.minor)")

if [ "$python_ver_major" -lt "3" ] || [ "$python_ver_minor" -lt "8" ] ; then
	echo "Python version must greater than 3.8"
	echo "Current $PYTHON version $python_ver_major.$python_ver_minor"
	echo "Upgrade $PYTHON or use PYTHON env to specify a newer python"
	exit
fi

# pip install the requirments
$PYTHON -m pip install --upgrade pip
$PYTHON -m pip install scikit-image
$PYTHON -m pip install tensorflow==2.14.0
$PYTHON -m pip install --upgrade intel-extension-for-tensorflow[xpu]
$PYTHON -m pip install intel-optimization-for-horovod
$PYTHON -m pip install gin gin-config tfa-nightly tensorflow-addons tensorflow-model-optimization tensorflow-datasets pyyaml

# check out the tensorflow models and apply the patch
cd ${WORKSPACE}
git clone -b v2.8.0 https://github.com/tensorflow/models.git tensorflow-models

mkdir -p hvd_configure
cd hvd_configure
#wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/hvd_support.patch
wget https://raw.githubusercontent.com/intel/intel-extension-for-tensorflow/main/examples/train_resnet50/hvd_configure/hvd_support.patch

wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_bf16_lars.yaml
wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_dummy_bf16_lars.yaml
wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_dummy_fp32_lars.yaml
wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_fp32_lars.yaml
cd ..

# Apply hvd_support patch
cd tensorflow-models
git apply ../hvd_configure/hvd_support.patch
