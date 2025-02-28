## Triton inference server with bigdl-llm vllm to deploy Llama2
This Triton inference server via bigdl-llm vllm Example is based on trtion server vllm backend on Intel Data Center GPU, [bigdl-llm vllm backend](https://github.com/intel-analytics/BigDL/tree/main/python/llm/example/GPU/vLLM-Serving) is used
This examples is following [Llama2 deployment via triton server](https://github.com/marvik-ai/triton-llama2-adapter.git) and provides overall intructions to make it run on Intel Data Center GPU

NOTE: bigdl-llm and this workload are still in development and currently it has [known limitations](#limitations).

### Build Triton Server with bigdl-llm vllm Container Env
Open a new terminal and build a new docker container image derived from tritonserver:23.12-pyt-python-py3

```
docker build $(env | grep -E '(_proxy=|_PROXY)' | sed 's/^/--build-arg /') \
        -t bigdl-llm/triton_vllm:latest \
        -f Dockerfile \
        .
```
Or use this scripts directly:
```
./docker_server_build.sh
```

### Llama2 model setup via triton bigdl-llm vllm backend
Pls run this scripts to setup Llama2 model which will apply the patch to enable bigdl-llm vllm
```
./setup_llama_bigdl_vllm_backend.sh
```
Pls Note:
pls modify `triton-llama2-adapter/vLLM/model_repository/vllm/1/model.py` to add your own HF token to access llama2 model
```
huggingface_hub.login(token="Your HF token")
```

Here is Llama2 model repo details based on vLLM backend - `triton-llama2-adapter/vLLM/`
```
model_repository/
|-- vllm
    |-- 1
    |   |-- model.py
    |-- config.pbtxt
    |-- vllm_engine_args.json
```

The `vllm_engine_args.json` file should contain the following:
```
{
    "model":"meta-llama/Llama-2-7b-chat-hf",
    "disable_log_requests": "true",
    "device": "xpu"
}
```
The `config.pbtxt` file define llama2 model configuration, including backend, input/output, instance group, etc. Pls refer to [Triton Model Configuration](https://github.com/triton-inference-server/server/blob/main/docs/user_guide/model_configuration.md)
The `model.py` file intends to define the model using the TritonPythonModel class as in the Python Backend approach. Here is the example to show how to set up a model using bigdl-llm vLLM 

### Start llama2 triton inference server
Update the Huggingface Home directory to actual path
```
HOST_HF_HOME=~/.cache/huggingface
docker run -it --rm -p 8001:8001 --net=host --device=/dev/dri --shm-size=1G --ulimit memlock=-1 --ulimit stack=67108864 --env http_proxy=${http_proxy} --env https_proxy=${https_proxy} --env no_proxy=${no_proxy} -v ${PWD}/triton-llama2-adapter/vLLM/model_repository/:/model_repository -v ${HOST_HF_HOME}:/root/.cache/huggingface/ bigdl-llm/triton_vllm:latest tritonserver --model-store /model_repository
```
Or run this script directly:
```
./docker_server_run.sh
```

### Send llama2 inference request to triton server from another Terminal
Open a new terminal and Start the Triton client:
```
./docker_client_run.sh
```
Send inference request to server: 
```
python client.py
```
Send inference request to server with 4 iterations 
```
python client.py --iterations 4
```

### Limitation:
- Bigdl-llm vLLM doesn't support [PagedAttention](https://blog.vllm.ai/2023/06/20/vllm.html) until now, so you may meet out of memory if too many requests are sent simultaneously, You could use "sudo xpu-smi dump -m 0,1,2,3,4,5,18" to monitor the GPU memory usage
- vLLM backend supports multi-gpu to process one big model via tensor parallism parameters [tensor_parallel_size](https://github.com/triton-inference-server/vllm_backend), but this isn't supported by Bigdl-llm vLLM
- Continuous Batching is enabled in vLLM comparing Dynamic Batching in Triton, bigdl-llm vLLM enabled this feature by default
- Does not support providing specific subset of GPUs to be used.
- If you want to run multiple instances of Triton server on Single GPU, you need to modify instance_group count in `triton-llama2-adapter/vLLM/model_repository/vllm/config.pbtxt`, by default, it's 1, you could update it to 2 to run 2 triton server instance on one Single GPU
