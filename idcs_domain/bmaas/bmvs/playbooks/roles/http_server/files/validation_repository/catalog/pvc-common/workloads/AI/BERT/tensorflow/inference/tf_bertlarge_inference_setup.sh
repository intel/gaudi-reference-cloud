#!/bin/bash

# Use script dir for workspace by default
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE=${WORKSPACE:-${SCRIPT_DIR}}
echo $WORKSPACE

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
        conda install -c conda-forge -y libstdcxx-ng=12
  elif [  $(which $PYTHON) ]; then
    echo "Using Python venv to create env"
    #sudo apt install python3-virtualenv
    #virtualenv -p python ${VENV_NAME}
    $PYTHON -m venv ${VENV_NAME}
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
git clone https://github.com/IntelAI/models.git
cd models
git checkout -b v2.12.0 v2.12.0
cp quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh.bak
cp ../benchmark.sh quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh
cd ..

