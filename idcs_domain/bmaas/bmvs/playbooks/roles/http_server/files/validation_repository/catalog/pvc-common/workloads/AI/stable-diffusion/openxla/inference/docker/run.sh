#!/bin/bash
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

echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

WORKDIR=/workspace
HOST_HF_HOME=~/.cache/huggingface
CONTAINER_HF_HOME=${WORKDIR}/huggingface

CONDA_ENVNAME="openxla"
docker run -it --rm --device /dev/dri:/dev/dri \
	--ipc=host -v /dev/dri/by-path:/dev/dri/by-path \
	--workdir ${WORKDIR} \
	--env http_proxy=${http_proxy} \
	--env https_proxy=${https_proxy} \
	--env CONDA_ENVNAME=${CONDA_ENVNAME} \
	--env HF_HOME=${CONTAINER_HF_HOME} \
	--volume ${HOST_HF_HOME}:${CONTAINER_HF_HOME} \
	openxla_sd \
	/bin/bash -c "export PATH=/root/miniconda3/bin:$PATH && \
	source activate ${CONDA_ENVNAME} && \
	pip list  && \
	ZE_AFFINITY_MASK=0 python openxla/intel-extension-for-openxla/example/stable_diffusion/jax_stable.py --num-inference-steps 50"
	#bash"



