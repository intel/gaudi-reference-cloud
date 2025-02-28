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


FROM intel/intel-extension-for-tensorflow:2.14.0.1-xpu
WORKDIR /workspace
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y pciutils numactl

SHELL ["/bin/bash", "-c"]
RUN pip install scikit-image gin gin-config tfa-nightly tensorflow-addons tensorflow-model-optimization tensorflow-datasets pyyaml

COPY benchmark.sh /workspace/
COPY tf_bertlarge_inference_run.sh /workspace/

RUN git clone -b v2.12.0 https://github.com/IntelAI/models.git
RUN pushd models && \ 
    cp quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh.bak && \
    cp /workspace/benchmark.sh quickstart/language_modeling/tensorflow/bert_large/inference/gpu/benchmark.sh && \
    popd

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*
