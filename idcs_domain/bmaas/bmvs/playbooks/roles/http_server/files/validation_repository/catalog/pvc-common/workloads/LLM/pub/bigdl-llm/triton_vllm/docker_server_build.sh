docker build $(env | grep -E '(_proxy=|_PROXY)' | sed 's/^/--build-arg /') \
        -t bigdl-llm/triton_vllm:latest \
        -f Dockerfile \
        .			 
