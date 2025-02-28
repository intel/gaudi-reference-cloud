#export HF_HOME=~/.cache/huggingface/
export HF_DATASETS_OFFLINE=1
export TRANSFORMERS_OFFLINE=1
export HF_EVALUATE_OFFLINE=1

num_iter=10
num_warmup=5
datatype=float16
device=xpu
#xpu_optimize="--disable-xpu-optimize"
beam=(1 4)
inputs=( 1024 )
outputs=(128 )

num_tiles_set=(1)
models=(EleutherAI/gpt-j-6B)

if [ "${TESTSET}" == "1" ]; then
num_tiles_set=(1 2)
models=(EleutherAI/gpt-j-6B facebook/opt-6.7b meta-llama/Llama-2-7b-hf meta-llama/Llama-2-13b-hf bigscience/bloom-7b1)
fi

if [ "${TESTSET}" == "2" ]; then
num_tiles_set=(2 4)
models=(facebook/opt-30b)
fi

if [ "${TESTSET}" == "3" ]; then
num_tiles_set=(4)
models=(meta-llama/Llama-2-70b-hf)
fi

#for max1550 only
if [ "${TESTSET}" == "31" ]; then
num_tiles_set=(8)
models=(meta-llama/Llama-2-70b-hf)
fi


if [ "${TESTSET}" == "4" ]; then
num_tiles_set=(8)
models=(bigscience/bloom)
fi

#not functional yet
#if [ "${TESTSET}" == "41" ]; then
#num_tiles_set=(16)
#models=(bigscience/bloom)
#fi


