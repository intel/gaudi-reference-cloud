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


FROM intel/intel-extension-for-pytorch:2.1.10-xpu

WORKDIR /workspace

USER root
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y pciutils numactl

SHELL ["/bin/bash", "-c"]

COPY pt_resnet50_inference_run.sh /workspace/
COPY resnet50_infer.patch /workspace/

RUN wget https://github.com/IntelAI/models/raw/v3.0.0/models/image_recognition/pytorch/resnet50v1_5/inference/gpu/main.py -O main.py
RUN patch -p1 < resnet50_infer.patch

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
