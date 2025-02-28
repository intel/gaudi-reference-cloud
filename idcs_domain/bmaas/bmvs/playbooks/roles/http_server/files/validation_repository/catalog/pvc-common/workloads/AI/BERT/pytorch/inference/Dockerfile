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

COPY pt_bertlarge_inference_run.sh /workspace/
COPY fp16_inference_plain_format.sh /workspace/
COPY pt_bert_infer.patch /workspace/

RUN git clone -b v2.12.0 https://github.com/IntelAI/models.git
RUN pushd models && \
    python -m pip install -r models/language_modeling/pytorch/bert_large/inference/gpu/requirements.txt && \
    cp ../fp16_inference_plain_format.sh quickstart/language_modeling/pytorch/bert_large/inference/gpu/ && \
    git apply ../pt_bert_infer.patch && \
    popd

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
