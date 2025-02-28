#!/bin/bash

# script dir as WORKSPACE by default
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
  elif [  $(which ${PYTHON}) ]; then
	echo "Using Python venv to create env"
	#sudo apt install python3-virtualenv 
	#virtualenv -p python ${VENV_NAME}
	${PYTHON} -m venv ${VENV_NAME}
	source ${VENV_NAME}/bin/activate
  else
 	echo "${PYTHON} not found, please install ${PYTHON} first"
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
${PYTHON} -m pip install --upgrade pip
${PYTHON} -m pip install tensorflow==2.14.0
${PYTHON} -m pip install --upgrade intel-extension-for-tensorflow[xpu]
${PYTHON} -m pip install intel-optimization-for-horovod
${PYTHON} -m pip install gin gin-config tensorflow-addons tensorflow-model-optimization tensorflow-datasets

echo "Checkout Intel AI Model zoo..."
cd ${WORKSPACE}
git clone https://github.com/IntelAI/models intelai-models
cd intelai-models
git checkout v2.12.0
echo ${PWD}
cp quickstart/image_recognition/tensorflow/resnet50v1_5/inference/gpu/batch_inference.sh quickstart/image_recognition/tensorflow/resnet50v1_5/inference/gpu/batch_inference.sh.bak
cp ../batch_inference.sh quickstart/image_recognition/tensorflow/resnet50v1_5/inference/gpu/batch_inference.sh
git apply ../fp16_preprocessing.patch

echo "Download model file..."
cd ${WORKSPACE}
PRETRAINED_MODEL_DIR=tf_resnet_models
mkdir -p ${PRETRAINED_MODEL_DIR}
#https://www.intel.com/content/www/us/en/developer/articles/machine-learning-model/resnet50v1-5-int8-inference-tensorflow-model.html
# Use the model mentioned here:
#https://github.com/IntelAI/models/tree/v2.11.0/benchmarks/image_recognition/tensorflow/resnet50v1_5/inference
echo "Download int8 models from https://storage.googleapis.com/intel-optimized-tensorflow/models/v1_8/resnet50v1_5_int8_pretrained_model.pb"
if [ ! -f ${PRETRAINED_MODEL_DIR}/resnet50v1_5_int8_pretrained_model.pb ]; then
    wget https://storage.googleapis.com/intel-optimized-tensorflow/models/v1_8/resnet50v1_5_int8_pretrained_model.pb -O  ${PRETRAINED_MODEL_DIR}/resnet50v1_5_int8_pretrained_model.pb
fi

echo "Download Resnet50 modles from https://zenodo.org/record/2535873/files/resnet50_v1.pb"
if [ ! -f ${PRETRAINED_MODEL_DIR}/resnet50_v1.pb ]; then
    wget  --no-check-certificate https://zenodo.org/record/2535873/files/resnet50_v1.pb -O  ${PRETRAINED_MODEL_DIR}/resnet50_v1.pb
fi

