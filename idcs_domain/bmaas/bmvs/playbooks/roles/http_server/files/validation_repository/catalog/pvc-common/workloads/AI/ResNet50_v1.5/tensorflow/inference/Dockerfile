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

COPY batch_inference.sh /workspace/
COPY fp16_preprocessing.patch /workspace/
COPY tf_resnet50_inference_run.sh /workspace/

RUN git clone -b v2.12.0 https://github.com/IntelAI/models intelai-models
RUN pushd intelai-models && \
    cp ../batch_inference.sh quickstart/image_recognition/tensorflow/resnet50v1_5/inference/gpu/batch_inference.sh && \
    git apply ../fp16_preprocessing.patch && \
    popd

ENV PRETRAINED_MODEL_DIR=tf_resnet_models
RUN mkdir -p ${PRETRAINED_MODEL_DIR}
RUN wget https://storage.googleapis.com/intel-optimized-tensorflow/models/v1_8/resnet50v1_5_int8_pretrained_model.pb -O  ${PRETRAINED_MODEL_DIR}/resnet50v1_5_int8_pretrained_model.pb
RUN wget  --no-check-certificate https://zenodo.org/record/2535873/files/resnet50_v1.pb -O  ${PRETRAINED_MODEL_DIR}/resnet50_v1.pb

RUN apt-get clean && \
    rm -rf  /var/lib/apt/lists/*

