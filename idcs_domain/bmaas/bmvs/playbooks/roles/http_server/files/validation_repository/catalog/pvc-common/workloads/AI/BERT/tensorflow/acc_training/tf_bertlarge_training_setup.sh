#!/bin/bash
python3 -m pip install --upgrade pip
python3 -m pip install absl-py  
python3 -m pip install --upgrade intel-extension-for-tensorflow[xpu]
python3 -m pip install git+https://github.com/NVIDIA/dllogger
python3 -m pip install requests tqdm sentencepiece tensorflow_hub wget progressbar 
python3 -m pip install tensorflow-addons  # Version details in https://github.com/tensorflow/addons
python3 -m pip install intel-optimization-for-horovod