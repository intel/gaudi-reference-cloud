#!/bin/bash
export PATH=/root/miniconda3/bin:$PATH
source activate ${CONDA_ENVNAME}

ln -s ${LLAMA2_7B_CDIR} bigdl-llama2-7b 
ln -s ${LLAMA2_13B_CDIR} bigdl-llama2-13b 
ln -s ${LLAMA2_70B_CDIR} bigdl-llama2-70b 
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/intel/oneapi/lib/intel64/cpu_dpcpp_gpu_dpcpp 
ZE_AFFINITY_MASK=0.0 python3 -m bigdl.llm.serving.model_worker --model-path ./bigdl-llama2-7b --device xpu --limit-worker-concurrency 1 --port 21002 --worker-address 'http://localhost:21002'  &
ZE_AFFINITY_MASK=0.1 python3 -m bigdl.llm.serving.model_worker --model-path ./bigdl-llama2-13b --device xpu --limit-worker-concurrency 1 --port 21003 --worker-address 'http://localhost:21003' &
ZE_AFFINITY_MASK=1.0 python3 -m bigdl.llm.serving.model_worker --model-path ./bigdl-llama2-70b --device xpu --limit-worker-concurrency 1 --port 21004 --worker-address 'http://localhost:21004' &

wait

ps -ef | grep 'bigdl.llm.serving.model_worker' | grep -v grep | awk '{print $2}' | xargs -r kill -9
