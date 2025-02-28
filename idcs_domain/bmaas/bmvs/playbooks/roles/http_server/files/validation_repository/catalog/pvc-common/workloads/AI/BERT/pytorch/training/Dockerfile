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
RUN apt-get update && apt-get install -y pciutils

SHELL ["/bin/bash", "-c"]

COPY ddp_bf16_training_plain_format.sh /workspace/
COPY bf16_training_plain_format.sh /workspace/
COPY pt_bertlarge_training_run.sh /workspace/

RUN git clone -b v2.12.0 https://github.com/IntelAI/models.git intelai-models
RUN pushd intelai-models && \
    pip install -r models/language_modeling/pytorch/bert_large/training/gpu/requirements.txt && \
    chmod 755 quickstart/language_modeling/pytorch/bert_large/training/gpu/*.sh && \
    bash ./quickstart/language_modeling/tensorflow/bert_large/training/gpu/setup.sh && \
    cp ../ddp_bf16_training_plain_format.sh quickstart/language_modeling/pytorch/bert_large/training/gpu/ && \
    cp ../bf16_training_plain_format.sh quickstart/language_modeling/pytorch/bert_large/training/gpu/ && \
    wget https://s3.amazonaws.com/models.huggingface.co/bert/bert-base-uncased-vocab.txt -O models/language_modeling/pytorch/bert_large/training/gpu/data/vocab.txt && \
    popd

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
