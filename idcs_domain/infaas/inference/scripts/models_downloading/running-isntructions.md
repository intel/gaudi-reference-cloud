# 🏃 Running Local S3 Model Streaming

* NOTE: If you don't have poetry please see main repo README.md for poetrt env setup.

1. Export secrets (get from MaaS team):
```bash
export AWS_ACCESS_KEY_ID=<accessId>
export AWS_SECRET_ACCESS_KEY=<secrete>
export HF_TOKEN=<token>
```

2. Add script to python execution dirs:
```bash
export PYTHONPATH=$(pwd)
```

3. Running script is possible from cli with the following arguments:
```python
--hf_token: HuggingFace api token [optional if you set HF_TOKEN]
--model_id: HuggingFace repo model id
--s3_bucket: The S3 bucket name [defaults to 'cnvrg-test-maas']
--base_s3_path: The S3 bucket path [defaults to 'models'] 
--aws_access_key: Access key for aws [optional if you set AWS_ACCESS_KEY_ID]
--aws_secret_access_key: Secret access key for aws [optional if you set AWS_SECRET_ACCESS_KEY]
--tmp_dir: Local dir for tmp files creation [defaults to None, meaning random dir selected by the OS]
```

4. Run the script:
```bash
poetry run python scripts/models_downloading/download_models_to_s3.py --model_id meta-llama/Meta-Llama-3.1-70B-Instruct --tmp_dir /tmp/dir/on/user/local/machine
```