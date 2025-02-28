# ⏱ Benchmarking ⏱
We created a simple python benchmark script, using a simple http package (httpx) and python time package.
We also create a synthetic user-question-dataset with 2000 questions varying between 8-51 tokens (data.yaml).

## ⤳ Test Flow
1. Create the test specified number of threads.
2. For each thread:<br>
&ensp;2.1 Create an http client<br>
&ensp;2.2 Sample random queries from data file according to the specified number of tests.<br>
&ensp;2.3 Run over the queries with `generate_stream` call.<br>
&ensp;2.4 Ror each query:<br>
    &emsp;&emsp;2.4.1 Measure time to 1st token (ttft).<br>
    &emsp;&emsp;2.4.2 Measure total generation time.<br>
    &emsp;&emsp;2.4.3 When streaming is done - wait specified sleep time between calls and go to next query.<br>
&ensp;2.5 Aggregate result in a thread-name:`BenchResults` dictionary.
3. Create a csv file with all test results - save locally on the container and keep the container open.
4. Optional: using `bench_runner.sh` and run the benchmark for multiple number of threads.
5. Optional: vusialize, create ttft and total generation time (over number of thread) graphs with `visualize.ipynb`.

## 📦 Build benchmark container
1. Clone the repo.

2. In the new local repo folder, run
```bash
    cd idcs_domain/infaas/inference
```

3. If you have Intel VPN on, make sure to export the following:
```bash
    export HTTP_PROXY=http://internal-placeholder.com:912
    export HTTPS_PROXY=http://internal-placeholder.com:912
    export NO_PROXY=intel.com,.intel.com,10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16,192.168.150.0/24,.kind.local
```

4. Build the tester.Dockerfile:
* If you don't have Intel VPM, ignore the `build-arg`
```bash
    docker build --platform=linux/amd64 --no-cache \
    -t <image_name> \
    -f inference_engine/tester.Dockerfile \
    --build-arg http_proxy=$HTTP_PROXY \
    --build-arg https_proxy=$HTTPS_PROXY \
    --build-arg no_proxy=$NO_PROXY .
```

5. In order to pull image in kubernetes cluster, push image to docker-hub (or any Intel internal registry):
```bash
    docker tag <image_name>:latest <dockerhub_user>/<image_name>:latest
    docker push <dockerhub_user>/<image_name>:latest
```

## 🏃 Run benchmark with kubernetes (IKS cluster)

1. Use the following pod file:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: tgi-llama31-70b-benchmark
  labels:
    name: tgi-llama31-70b-benchmark
spec:
  runtimeClassName: habana
  containers:
  - name: tgi-llama31-70b  
    image: ghcr.io/huggingface/tgi-gaudi:2.0.5
    securityContext:
      capabilities:
        add: ["SYS_NICE"]   
    env:
      - name: HABANA_VISIBLE_DEVICES
        value: "all"
      - name: OMPI_MCA_btl_vader_single_copy_mechanism
        value: "none"
      - name: PT_HPU_ENABLE_LAZY_COLLECTIVES
        value: "true"
      - name: MAX_TOTAL_TOKENS
        value: "2048"
      - name: BATCH_BUCKET_SIZE
        value: "256"
      - name: PREFILL_BATCH_BUCKET_SIZE
        value: "4"
      - name: PAD_SEQUENCE_TO_MULTIPLE_OF
        value: "64"
      - name: USE_FLASH_ATTENTION
        value: "true"
      - name: FLASH_ATTENTION_RECOMPUTE
        value: "true"
      - name: MODEL_ID
        value: meta-llama/Meta-Llama-3.1-70B-Instruct
      - name: PORT 
        value: "8080"
      - name: HUGGINGFACE_HUB_CACHE 
        value: /models-cache
      - name: TGI_PROFILER_ENABLED 
        value: "true"    
      - name: NUM_SHARD 
        value: "8"
      - name: SHARDED 
        value: "true"    
      - name: HUGGING_FACE_HUB_TOKEN 
        value: <YOUR_HF_API_TOKEN>
    resources:
      limits:
        habana.ai/gaudi: "8"
    volumeMounts:
        - name: models-cache
          mountPath: models-cache
    args:
      - --max-input-length
      - "1024"
      - --max-batch-prefill-tokens 
      - "4096"
      - --max-total-tokens 
      - "2048"
      - --max-batch-total-tokens 
      - "524288"
      - --max-waiting-tokens 
      - "5"
      - --waiting-served-ratio 
      - "1.1"
      - --max-concurrent-requests 
      - "512"
      - "--hostname"
      - "127.0.0.1"
      - "--port"
      - "8080"
  - name: tgi-tester
    image: <benchmark_image_address>
    imagePullPolicy: Always 
    env:
      - name: hf_api_key 
        value: <YOUR_HF_API_TOKEN>
    resources:
      requests:
        # cpu can also be 32 or 16, less than 16 might do problems
        cpu: "64"
  # imagePullSecrets - if needed
  imagePullSecrets:
    - name: <YOUR_IMAGE_PULL_SECRET>
  volumes:
  - name: models-cache
    hostPath:
     path: /data
     type: Directory    
```

2. Create the cluster, this will also start the tests after the model warmup:
```bash
    kubectl apply -f /pod/file/path/
```
* See `bench_runner.sh` file for full test arguments.

3. When the test is done, `bench_results` folder will have the tests vsc results.<br>
You can copy the directory to your local machine:
```bash
    kubectl cp <namespace>/tgi-llama31-70b-benchmark:/app/bench_results/ /local/results/path/ -c tgi-tester
```
* NOTE:<br>
`tgi-llama31-70b-benchmark` is the pod name, so if you changed it, change in the cp command as well.
