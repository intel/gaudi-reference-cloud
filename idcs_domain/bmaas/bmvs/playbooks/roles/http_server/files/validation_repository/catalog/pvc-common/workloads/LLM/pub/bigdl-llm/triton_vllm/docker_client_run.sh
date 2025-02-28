docker run -it --net=host --env http_proxy=${http_proxy} --env https_proxy=${https_proxy} --env no_proxy=${no_proxy} -v ${PWD}/triton-llama2-adapter/vLLM/:/workspace/ nvcr.io/nvidia/tritonserver:23.12-py3-sdk /bin/bash

##After entering docker:
#python client.py
