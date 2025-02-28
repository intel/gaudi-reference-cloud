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
 
 
FROM nvcr.io/nvidia/tritonserver:23.12-pyt-python-py3
 
ENV LANG=C.UTF-8
 
ARG DEBIAN_FRONTEND=noninteractive
 
RUN apt-get update && \
    apt-get install -y --no-install-recommends --fix-missing \
    apt-utils \
    build-essential \
    ca-certificates \
    clinfo \
    curl \
    git \
    gnupg2 \
    gpg-agent \
    rsync \
    sudo \
    unzip \
    wget && \
    apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
 
RUN wget -O- https://internal-placeholder.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB \
| gpg --dearmor | tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null
RUN echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://internal-placeholder.com/oneapi all main" | tee /etc/apt/sources.list.d/oneAPI.list

RUN wget -qO - https://repositories.intel.com/gpu/intel-graphics.key | \
    gpg --dearmor --output /usr/share/keyrings/intel-graphics.gpg

RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/intel-graphics.gpg] https://repositories.intel.com/gpu/ubuntu jammy unified" | \
    tee /etc/apt/sources.list.d/intel-gpu-jammy-rolling-stable.list

RUN apt-get update && apt-get install -y intel-oneapi-runtime-libs=2024.0.0-49055

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates cmake curl unzip python3-pip python3-mako vim clinfo intel-opencl-icd intel-level-zero-gpu level-zero
 
SHELL ["/bin/bash", "-c"]

RUN pip install psutil  && \
    pip install sentencepiece   && \ 
    pip install numpy && \
    pip install "transformers>=4.33.1"  && \ 
    pip install --pre --upgrade bigdl-llm[xpu_2.1] -f https://developer.intel.com/ipex-whl-stable-xpu && \
    pip install fastapi  && \
    pip install "uvicorn[standard]" && \
    pip install "pydantic<2"  && \
    pip install py-cpuinfo

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
