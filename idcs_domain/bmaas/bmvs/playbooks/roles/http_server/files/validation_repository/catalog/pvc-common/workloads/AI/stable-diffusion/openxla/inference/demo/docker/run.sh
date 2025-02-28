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

#web UI IP and port
HOST_IP=$(hostname -I |awk '{print $1}')
HOST_PORT=8081

WORKDIR=/workspace
HOST_HF_HOME=~/.cache/huggingface
CONTAINER_HF_HOME=${WORKDIR}/huggingface

CONDA_ENVNAME="openxla-sd-demo"
docker run -it --rm \
	--device /dev/dri/card5:/dev/dri/card5 \
	--device /dev/dri/renderD132:/dev/dri/renderD132 \
	--device /dev/dri/card6:/dev/dri/card6 \
	--device /dev/dri/renderD133:/dev/dri/renderD133 \
	-v /dev/dri/by-path:/dev/dri/by-path \
	--ipc=host --net=host 	\
	--workdir ${WORKDIR} \
	--env http_proxy=${http_proxy} \
	--env https_proxy=${https_proxy} \
	--env no_proxy=localhost,127.0.0.1,${HOST_IP} \
	--env CONDA_ENVNAME=${CONDA_ENVNAME} \
	--env HF_HOME=${CONTAINER_HF_HOME} \
	--env HOST_IP=${HOST_IP} \
	--env HOST_PORT=${HOST_PORT} \
	--env SHARE=false \
	--volume ${HOST_HF_HOME}:${CONTAINER_HF_HOME} \
	openxla-sd-demo \
	/bin/bash -c "export PATH=/root/miniconda3/bin:$PATH && \
	source activate ${CONDA_ENVNAME} && \
	pip list  && \
	cd stable-diffusion && \
	python app.py"
	#bash"

# option to pass one specific GPU, e.g.
#--device /dev/dri/card1:/dev/dri/card1 \
#--device /dev/dri/renderD128:/dev/dri/renderD128 \


