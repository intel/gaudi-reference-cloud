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
                       wget git ca-certificates pkg-config build-essential gpg vim

RUN wget -O- https://internal-placeholder.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB \
| gpg --dearmor | tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null
RUN echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://internal-placeholder.com/oneapi all main" | tee /etc/apt/sources.list.d/oneAPI.list

RUN wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | \
    gpg --dearmor --output /usr/share/keyrings/intel-graphics.gpg
RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy/production/2328 unified" | \
    tee /etc/apt/sources.list.d/intel-gpu-jammy.list

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates cmake curl unzip python3-pip python3-mako vim clinfo intel-opencl-icd intel-level-zero-gpu level-zero

#RUN apt-get update && apt-get install -y intel-basekit-2023.2.0
#RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2023.2.0-49016

# for oneAPI 2023.2 runtime only
RUN apt-get update && apt install -y intel-oneapi-common-vars=2023.2.0-49462 intel-oneapi-common-licensing=2023.2.0-49462\
    intel-oneapi-runtime-openmp=2023.2.2-47 intel-oneapi-runtime-dpcpp-cpp=2023.2.2-47 intel-oneapi-runtime-mkl=2023.2.0-49495 intel-oneapi-runtime-dnnl=2023.2.0-49516\
    intel-oneapi-runtime-mpi=2021.10.0-49371 intel-oneapi-runtime-tbb=2021.10.0-49541\
    intel-oneapi-runtime-ccl=2021.10.0-49084

RUN wget \
    https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh \
    && mkdir /root/.conda \
    && bash Miniconda3-latest-Linux-x86_64.sh -b \
    && rm -f Miniconda3-latest-Linux-x86_64.sh

RUN export PATH=/root/miniconda3/bin:$PATH && conda --version

SHELL ["/bin/bash", "-c"]

RUN export PATH=/root/miniconda3/bin:$PATH && \
	conda create -n bigdl-llm-demo python=3.9 && \
	source activate bigdl-llm-demo && \
	conda install libstdcxx-ng==12.2.0 -c conda-forge -y

RUN export PATH=/root/miniconda3/bin:$PATH && \        
        source activate bigdl-llm-demo && \ 
		pip install --pre --upgrade bigdl-llm[xpu,serving] --extra-index-url https://pytorch-extension.intel.com/release-whl/stable/xpu/us/ && \
		pip install gradio==3.23.0


RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*

