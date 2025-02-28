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
RUN apt-get update && apt-get install -y pciutils

SHELL ["/bin/bash", "-c"]
RUN pip install scikit-image gin gin-config tfa-nightly tensorflow-addons tensorflow-model-optimization tensorflow-datasets pyyaml
RUN git clone -b v2.8.0 https://github.com/tensorflow/models.git tensorflow-models
RUN mkdir -p hvd_configure
RUN pushd hvd_configure && \
    wget https://raw.githubusercontent.com/intel/intel-extension-for-tensorflow/main/examples/train_resnet50/hvd_configure/hvd_support.patch && \
    wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_bf16_lars.yaml && \
    wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_dummy_bf16_lars.yaml && \
    wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_dummy_fp32_lars.yaml && \
    wget https://github.com/intel/intel-extension-for-tensorflow/raw/r2.14/examples/train_resnet50/hvd_configure/itex_fp32_lars.yaml && \
    popd
RUN pushd tensorflow-models && \
    git apply ../hvd_configure/hvd_support.patch

COPY tf_resnet50_training_run.sh /workspace/

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*

