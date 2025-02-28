# Instructions for running FP8 quantization:

## About:
In this instruction we'll use the following image as our quantization environment: `galsarid/intel:quant-env`.<br>
The image has all the required installations (`pytorch`, `optimum-habana`, `DeepSpeed`) to run over a Gaudi cluster with **SynapseAI v1.18.0**.<br>
Any other SynapseAI version might cause errors.
This insturctions also assume running on a functional MaaS cluster - pod file reads MaaS env secretes.<br>
If running from a different cluster - override the `HF_TOKEN` at the `quantization.yaml` with your HuggingFace token.

## Checking SynapseAI Version:

1. Connect to the target device using `ssh`.
2. Run: `hl-smi` command.
3. Expected result for v1.18.0:
```txt
+-----------------------------------------------------------------------------+
| HL-SMI Version:                              hl-1.18.0-fw-53.1.1.1          |
| Driver Version:                                     1.18.0-ee698fb          |
|-------------------------------+----------------------+----------------------+
| AIP  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |
| Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | AIP-Util  Compute M. |
|===============================+======================+======================|
```

## Running Quantization:

1. Use the repo `quantization.yaml` to create a pod:
```bash
kubectl apply -f <path-to-local-repo>/frameworks.cloud.devcloud.services.idc/idcs_domain/infaas/inference/quantization/quantization.yaml
```

2. Step into the pod:
```bash
kubectl exec -it quantization-pod -- /bin/bash
```

3. Now you can run the quantization script:
    <br>3.1 For single device:
    ```bash
    QUANT_CONFIG=/app/optimum-habana/examples/text-generation/quantization_config/maxabs_measure.json python /app/optimum-habana/examples/text-generation/run_generation.py --model_name_or_path <model-name> --use_hpu_graphs --limit_hpu_graphs --use_kv_cache --bucket_size=128 --use_flash_attention --flash_attention_recompute --bf16 --batch_size 1
    ```
    3.2 For multi device:

    ```bash
    QUANT_CONFIG=/app/optimum-habana/examples/text-generation/quantization_config/maxabs_measure.json python /app/optimum-habana/examples/gaudi_spawn.py --use_deepspeed --world_size <num-devices> /app/optimum-habana/examples/text-generation/run_lm_eval.py -o acc_<model-name>_measure.txt --model_name_or_path <model-name>  --attn_softmax_bf16 --use_hpu_graphs --trim_logits --use_kv_cache --bucket_size=128 --bucket_internal --use_flash_attention --flash_attention_recompute --bf16 --batch_size 1
    ```

NOTE: If you work on multi device - make sure that `<num-devices>` is less than or equal to the pod file's number of gaudis at `habana.ai/gaudi:`.

4. Copy result files from pod to local machine:
   4.1 Stop the kubectl exec: type `exit` in the container's terminal.
   4.2 From user's terminal:
```bash
kubectl cp quantization-pod:/app/quantization_config/maxabs_quant.json /<target-local-path>/maxabs_quant.json
kubectl cp quantization-pod:/app/hqt_output /<target-local-path>/hqt_output
```

NOTE: The `maxabs_quant.json` is the default config name. There are models with a different config name, for example: `maxabs_quant_phi.json`, `maxabs_quant_mixtral.json` and `maxabs_quant_gemma.json`. Check on `optimum-habana` git repo if your target model usases the default config.
