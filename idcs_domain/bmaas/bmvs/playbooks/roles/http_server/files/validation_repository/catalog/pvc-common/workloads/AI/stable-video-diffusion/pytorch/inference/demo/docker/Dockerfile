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


FROM ubuntu:22.04

WORKDIR /workspace

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y software-properties-common curl gnupg && \
    apt-get install -y --no-install-recommends --fix-missing numactl cmake autoconf unzip \
                       wget git ca-certificates pkg-config build-essential gpg git-lfs ffmpeg mediainfo

RUN wget -O- https://internal-placeholder.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB \
| gpg --dearmor | tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null
RUN echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://internal-placeholder.com/oneapi all main" | tee /etc/apt/sources.list.d/oneAPI.list

RUN wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | \
    gpg --dearmor --output /usr/share/keyrings/intel-graphics.gpg
#RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy/production/2328 unified" | \
#    tee /etc/apt/sources.list.d/intel-gpu-jammy.list

RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy unified" | \
    tee /etc/apt/sources.list.d/intel-gpu-jammy-rolling-stable.list


#RUN apt-get update && apt-get install -y intel-basekit-2024.0
RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2024.0.0-49055
#RUN apt-get update && apt-get install -y intel-basekit-2023.2.0
#RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2023.2.0-49016

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates cmake curl unzip python3-pip python3-mako vim clinfo intel-opencl-icd intel-level-zero-gpu level-zero

RUN wget \
    https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh \
    && mkdir /root/.conda \
    && bash Miniconda3-latest-Linux-x86_64.sh -b \
    && rm -f Miniconda3-latest-Linux-x86_64.sh

RUN export PATH=/root/miniconda3/bin:$PATH && conda --version

COPY stable-video-diffusion-demo.diff /workspace
COPY inputs /workspace/inputs

SHELL ["/bin/bash", "-c"]

RUN export PATH=/root/miniconda3/bin:$PATH && \
	conda create -n svd_demo python=3.9 && \
	source activate svd_demo && \
	conda install libstdcxx-ng==12.2.0 -c conda-forge -y

RUN export PATH=/root/miniconda3/bin:$PATH && \
    source activate svd_demo && \ 
	pip install torch==2.1.0a0 torchvision==0.16.0a0 torchaudio==2.1.0a0 intel-extension-for-pytorch==2.1.10+xpu --extra-index-url https://pytorch-extension.intel.com/release-whl/stable/xpu/us/ && \
    pip install -q -U diffusers transformers accelerate && \
    pip install opencv-python && \
    ulimit -n 8192

RUN export PATH=/root/miniconda3/bin:$PATH && source activate svd_demo && \
    git lfs install && \
    git clone https://huggingface.co/spaces/multimodalart/stable-video-diffusion stable-video-diffusion-demo && \
	pushd stable-video-diffusion-demo && \
	git checkout 5f826658cef33c2db2c0edd7f823cf18265d3634 && \
	git apply ../stable-video-diffusion-demo.diff  && \
	mv /workspace/inputs .  && \
	popd

RUN apt-get install -y libgl1
RUN export PATH=/root/miniconda3/bin:$PATH && source activate svd_demo && \
	pushd stable-video-diffusion-demo && \
	pip install -r requirements.txt && \
	popd


RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*

