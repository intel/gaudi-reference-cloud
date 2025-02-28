#!/bin/bash

sleep_s="$SLEEP_S"

export hf_api_key="$HF_API_KEY"
export model_id="$MODEL_ID"
export output_dir=/app/bench_results/
export num_tests=100
export warmup="${WARMUP:-true}"

IFS=',' read -r -a threads <<< "$THREADS"

source /app/venv/bin/activate
pip install httpx

for i in "${!threads[@]}"
do
    # Starting with sleep for model warmup
    # keeping it between tests with different loads
    echo "Sleeping for $sleep_s seconds..."
    sleep $sleep_s
    
    thread="${threads[$i]}"
    export num_threads=$thread

    if [ "$i" -ne 0 ]; then
        export warmup="false"
    fi

    echo "Running test with num_threads=$thread"
    python /app/test_runner.py
done

echo "All tests completed. Keeping the container running..."
tail -f /dev/null