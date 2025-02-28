# Pytorch Resnet50 v1.5 Training on Intel Data Center GPU Max Series(Advanced)

Here provides method of running through python code directly.



## Run in Container

### Requirements

1. The same as quick start [README](./README.md) for preparing host server and dataset.
2. Pull docker image:
```
docker pull intel/image-recognition:pytorch-max-gpu-resnet50v1-5-training
```


### Launch Container
1. Set dataset folder and other env:
```
export DATASET_DIR=/home/intel/workspace/dataset/imagenet/

## docker parameters
DOCKER_ARGS="--rm --init -it"
IMAGE_NAME=intel/image-recognition:pytorch-max-gpu-resnet50v1-5-training

VIDEO=$(getent group video | sed -E 's,^video:[^:]*:([^:]*):.*$,\1,')
RENDER=$(getent group render | sed -E 's,^render:[^:]*:([^:]*):.*$,\1,')
test -z "$RENDER" || RENDER_GROUP="--group-add ${RENDER}"
```

2. Launch container
```
docker run \
  --group-add ${VIDEO} \
  ${RENDER_GROUP} \
  --device=/dev/dri \
  --shm-size=10G \
  --privileged \
  --ipc=host \
  --env WORKSPACE=${WORKSPACE} \
  --env DATASET_DIR=${DATASET_DIR} \
  --env http_proxy=${http_proxy} \
  --env https_proxy=${https_proxy} \
  --env no_proxy=${no_proxy} \
  --volume ${DATASET_DIR}:${DATASET_DIR} \
  --volume ${OUTPUT_DIR}:${OUTPUT_DIR} \
  --volume /dev/dri:/dev/dri \
  ${DOCKER_ARGS} \
  $IMAGE_NAME \
  /bin/bash
```

3. check script options
```
python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py -h
```

4. tensorboard could be required in parts case
```
pip install tensorboard
```

### Training Command-line Samples

1. Training benchmark with dummy dataset
```
python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu  0 --bf16 1 --dummy --benchmark 1 --num-iteration 100 --epochs 1
```

2. Training on ImageNet dataset
```
python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu 0 --bf16 1 --epochs 2 --tensorboard ${DATASET_DIR}
```

3. Enabling 2 stacks on a card for training
```
source /opt/intel/oneapi/setvars.sh 
mpirun -np 2 -ppn 2 --prepend-rank python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu  0 --bf16 1 --epochs 90 --tensorboard   ${DATASET_DIR}
```


### Inference Command-line Samples
1. Inference on single stack
```
python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu 0 --fp16 1 --pretrained --evaluate  --num-iterations 64 --jit-trace ${DATASET_DIR}
```
2. Launch 2 process and run on 2 stacks of a card
```
ZE_AFFINITY_MASK=0.0 ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu  0 --fp16 1 --pretrained --evaluate  --num-iterations 64 --jit-trace ${DATASET_DIR} & ZE_AFFINITY_MASK=0.1 python ./models/image_recognition/pytorch/resnet50v1_5/training/gpu/main.py --arch resnet50 --xpu  0 --fp16 1 --pretrained --evaluate  --num-iterations 64 --jit-trace ${DATASET_DIR}
```


## Run on Host

Here are PyTorch software stacks on Intel GPU. We should follow official guide to install driver, onAPI toolkis, Intel extension and Intel oneCCL for Intel GPU.

### Step 1: Install [Intel Max GPU Driver](https://dgpu-docs.intel.com/driver/installation.html)
1. Make sure prerequisites to add repository access are available
```
sudo apt-get update
sudo apt-get install -y gpg-agent wget
```

2. To add the repository for Intel® Data Center GPU Max Series:
```
wget -qO - https://repositories.intel.com/graphics/intel-graphics.key | \
  sudo gpg --dearmor --output /usr/share/keyrings/intel-graphics.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/graphics/ubuntu jammy max" | \
  sudo tee /etc/apt/sources.list.d/intel-gpu-jammy.list
sudo apt-get update
```

3. To install on a bare metal system, sufficient for hardware management and support of the runtimes in containers and bare metal, the kernel and xpu-smi packages can be installed:
```
sudo apt-get install -y intel-i915-dkms xpu-smi
sudo reboot -h now
```

4. Install dependent packages:
```
sudo apt-get install -y \
  intel-opencl-icd intel-level-zero-gpu level-zero \
  intel-media-va-driver-non-free libmfx1 libmfxgen1 libvpl2 \
  libegl-mesa0 libegl1-mesa libegl1-mesa-dev libgbm1 libgl1-mesa-dev libgl1-mesa-dri \
  libglapi-mesa libgles2-mesa-dev libglx-mesa0 libigdgmm12 libxatracker2 mesa-va-drivers \
  mesa-vdpau-drivers mesa-vulkan-drivers va-driver-all vainfo hwinfo clinfo
```

5. Development packages
```
sudo apt-get install -y \
  libigc-dev intel-igc-cm libigdfcl-dev libigfxcmrt-dev level-zero-dev
```

6. add the user to the render node group:
```
sudo gpasswd -a ${USER} render
newgrp render
```

### Step 2: Install [Intel oneAPI Toolkits](https://www.intel.com/content/www/us/en/docs/oneapi/installation-guide-linux/2023-1/overview.html)
1. Set up oneAPI repository:
```
# download the key to system keyring
wget -O- https://internal-placeholder.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB \
| gpg --dearmor | sudo tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null

# add signed entry to apt sources and configure the APT client to use Intel repository:
echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://internal-placeholder.com/oneapi all main" | sudo tee /etc/apt/sources.list.d/oneAPI.list

sudo apt update
```

2. Install Packages
```
sudo apt install intel-basekit intel-hpckit
```

3. Add [Hotfix to oneAPI kits](https://intel.github.io/intel-extension-for-pytorch/xpu/latest/tutorials/installation.html#install-oneapi-base-toolkit)
```
wget https://registrationcenter-download.intel.com/akdlm/IRC_NAS/89283df8-c667-47b0-b7e1-c4573e37bd3e/2023.1-linux-hotfix.zip

unzip 2023.1-linux-hotfix.zip
cd 2023.1-linux-hotfix
source /opt/intel/oneapi/setvars.sh 
bash installpatch.sh
```

### Step 3: Install PyTorch Software Stack
1. Create Python venv
```
python -m venv pytorch_venv

source /opt/intel/oneapi/setvars.sh 
source pytorch_venv/bin/activate
```
2. Install [Intel Extension for PyTorch](https://intel.github.io/intel-extension-for-pytorch/xpu/latest/index.html) and [Intel oneCCL Bindings for PyTorch](https://github.com/intel/torch-ccl)
```
python -m pip install torch==1.13.0a0+git6c9b55e torchvision==0.14.1a0 intel_extension_for_pytorch==1.13.120+xpu -f https://developer.intel.com/ipex-whl-stable-xpu
python -m pip install oneccl_bind_pt -f https://developer.intel.com/ipex-whl-stable-xpu
```

### Step 4: Prepare Source Codes
```
git clone https://github.com/IntelAI/models.git

cd models/models/image_recognition/pytorch/resnet50v1_5/training/gpu/
python main.py -h
```

### Step 5: Command-line

The command-line is the same as test in container



