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
CONDA_ENVNAME="bigdl-llm-demo"
docker run -it --rm \
	--device /dev/dri:/dev/dri \
	-v /dev/dri/by-path:/dev/dri/by-path \
	--ipc=host --net=host 	\
	--workdir ${WORKDIR} \
	--env http_proxy=${http_proxy} \
	--env https_proxy=${https_proxy} \
	--env no_proxy=localhost,127.0.0.1,${HOST_IP} \
	--env CONDA_ENVNAME=${CONDA_ENVNAME} \
	bigdl-llm-fastchat-demo \
	/bin/bash -c "export PATH=/root/miniconda3/bin:$PATH && \
	source activate ${CONDA_ENVNAME} && \
	pip list  && \
	python3 -m fastchat.serve.controller \
	"

