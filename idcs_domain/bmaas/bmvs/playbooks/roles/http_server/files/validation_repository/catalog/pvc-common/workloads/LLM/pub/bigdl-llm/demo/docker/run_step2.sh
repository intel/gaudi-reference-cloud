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


#web UI IP and port
HOST_IP=$(hostname -I |awk '{print $1}')
HOST_PORT=8080

WORKDIR=/workspace
HOST_HF_HOME=~/.cache/huggingface
CONTAINER_HF_HOME=${WORKDIR}/huggingface
LLAMA2_7B=hub/models--meta-llama--Llama-2-7b-chat-hf/snapshots/94b07a6e30c3292b8265ed32ffdeccfdadf434a8
LLAMA2_13B=hub/models--meta-llama--Llama-2-13b-chat-hf/snapshots/13f8d72c0456c17e41b3d8b4327259125cd0defa
LLAMA2_70B=hub/models--meta-llama--Llama-2-70b-chat-hf/snapshots/cfe96d938c52db7c6d936f99370c0801b24233c4

#model path in container. change the path if needed
LLAMA2_7B_CDIR=${CONTAINER_HF_HOME}/${LLAMA2_7B}
LLAMA2_13B_CDIR=${CONTAINER_HF_HOME}/${LLAMA2_13B} 
LLAMA2_70B_CDIR=${CONTAINER_HF_HOME}/${LLAMA2_70B} 

CONDA_ENVNAME="bigdl-llm-demo"
docker run -it --rm \
	--device /dev/dri/card1:/dev/dri/card1 \
	--device /dev/dri/card2:/dev/dri/card2 \
	--device /dev/dri/renderD128:/dev/dri/renderD128 \
	--device /dev/dri/renderD129:/dev/dri/renderD129 \
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
	--env LLAMA2_7B_CDIR=${LLAMA2_7B_CDIR} \
	--env LLAMA2_13B_CDIR=${LLAMA2_13B_CDIR} \
	--env LLAMA2_70B_CDIR=${LLAMA2_70B_CDIR} \
	--volume ${HOST_HF_HOME}:${CONTAINER_HF_HOME} \
	--volume `pwd`/scripts:${WORKDIR}/bigdl-llm-fastchat-demo \
	bigdl-llm-fastchat-demo \
	/bin/bash -c " pushd /workspace/bigdl-llm-fastchat-demo && \
					./02_start_models.sh && \	
					ps -ef | grep 'bigdl.llm.serving.model_worker' | grep -v grep | awk '{print $2}' | xargs -r kill -9	
				"


