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
                       wget git ca-certificates pkg-config build-essential gpg

RUN wget -O- https://internal-placeholder.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB \
| gpg --dearmor | tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null
RUN echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://internal-placeholder.com/oneapi all main" | tee /etc/apt/sources.list.d/oneAPI.list

RUN wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | \
    gpg --dearmor --output /usr/share/keyrings/intel-graphics.gpg
#RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy/production/2328 unified" | \
#    tee /etc/apt/sources.list.d/intel-gpu-jammy.list
RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy unified" | \
    tee /etc/apt/sources.list.d/intel-gpu-jammy-rolling-stable.list

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates cmake curl unzip python3-pip python3-mako vim clinfo intel-opencl-icd intel-level-zero-gpu level-zero

#RUN apt-get update && apt-get install -y intel-basekit-2023.2.0
#RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2023.2.0-49016
RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2024.0.0-49055

RUN wget \
    https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh \
    && mkdir /root/.conda \
    && bash Miniconda3-latest-Linux-x86_64.sh -b \
    && rm -f Miniconda3-latest-Linux-x86_64.sh

RUN export PATH=/root/miniconda3/bin:$PATH && conda --version


SHELL ["/bin/bash", "-c"]


RUN export PATH=/root/miniconda3/bin:$PATH && \
	conda create -n openxla-sd-demo python=3.9 && \
	source activate openxla-sd-demo && \
	conda install libstdcxx-ng==12.2.0 -c conda-forge -y

RUN export PATH=/root/miniconda3/bin:$PATH && \
    source activate openxla-sd-demo && \ 
    pip install jax==0.4.20 jaxlib==0.4.20 flax==0.7.0 && \
    pip install --upgrade intel-extension-for-openxla && \
    pip install transformers==4.32 diffusers==0.16.1 datasets==2.12.0 msgpack==1.0.7 gradio==3.14.0
	
	
RUN apt-get install -y libgl1
#RUN export PATH=/root/miniconda3/bin:$PATH && source activate openxla-sd-demo && \
#	pip install gradio==3.14.0 accelerate


COPY stable-diffusion-2.1.diff /workspace
RUN git clone https://huggingface.co/spaces/stabilityai/stable-diffusion && \
	pushd stable-diffusion && \
	git checkout 7da827c2024da3963cbaca23d3fcc3d0370c56db && \
	git apply ../stable-diffusion-2.1.diff

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*

