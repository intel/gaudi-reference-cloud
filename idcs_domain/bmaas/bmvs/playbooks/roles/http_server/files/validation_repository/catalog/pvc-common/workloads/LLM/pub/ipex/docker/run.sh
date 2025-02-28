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


WORKDIR=/workspace
HOST_HF_HOME=~/.cache/huggingface
CONTAINER_HF_HOME=${WORKDIR}/huggingface
DOCKER_IMG_NAME=ipex-llm:2.1.10

DOCKER_ARGS="--privileged -it --rm \
	--device /dev/dri:/dev/dri \
	-v /dev/dri/by-path:/dev/dri/by-path \
	--ipc=host --net=host 	\
	--cap-add=ALL \
	-v /lib/modules:/lib/modules \
	--workdir ${WORKDIR} \
	--env http_proxy=${http_proxy} \
	--env https_proxy=${https_proxy} \
	--env no_proxy=localhost,127.0.0.1,${HOST_IP} \
	--env HF_HOME=${CONTAINER_HF_HOME} \
	--volume ${HOST_HF_HOME}:${CONTAINER_HF_HOME} \
	--volume `pwd`/llm_inference_test:${WORKDIR}/llm_inference_test"

docker run  ${DOCKER_ARGS} \
	${DOCKER_IMG_NAME}	 \
	/bin/bash 	\
	-c "
	sudo chmod g+w llm_inference_test -R &&\
	sudo chmod g+w ${CONTAINER_HF_HOME} -R &&\
	export PATH=~/miniconda3/bin:$PATH && \
	source activate py310 &&\
	export LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libstdc++.so.6.0.30 &&\
	cd llm_inference_test		   &&\
	./test.sh
	"


