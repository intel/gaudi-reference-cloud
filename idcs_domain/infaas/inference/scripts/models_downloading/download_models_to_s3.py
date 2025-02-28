from huggingface_hub import HfApi, hf_hub_download
from botocore.client import BaseClient
from argparse import ArgumentParser, Namespace
from typing import Optional, Tuple, BinaryIO, List
from tqdm import tqdm
import tempfile
import logging
import boto3
import time
import sys
import os


def set_logger() -> None:
    logging.basicConfig(
        format="%(asctime)s %(levelname)s:%(filename)s: %(message)s",
        datefmt="%d-%m-%Y %H:%M:%S",
        level=logging.INFO,
        stream=sys.stdout
    )

def get_existing_s3_files(
    s3_client: BaseClient,
    bucket_name: str,
    folder_name: str
) -> List[str]:

    response = s3_client.list_objects_v2(
        Bucket=bucket_name, Prefix=folder_name
    )
    response_content = response.get("Contents", [])
    existing_files = [
        file["Key"].split("/")[-1]
        for file in response_content
    ]
    return existing_files

def is_valid_file_name(file_name: str) -> bool:

    invalid_posfixes = [".md", ".bin", ".txt"]
    for invalid_posfix in invalid_posfixes:
        if file_name.lower().endswith(invalid_posfix):
            return False

    invalid_prefixes = [".", "pytorch_model"]
    for invalid_prefix in invalid_prefixes:
        if file_name.lower().startswith(invalid_prefix):
            return False

    if "/" in file_name:
        return False
    if file_name.upper() == "LICENSE":
        return False

    return True

def get_cli_args() -> Namespace:
    parser = ArgumentParser("")
    parser.add_argument(
        "--hf_token",
        help="HuggingFace api token",
        type=str,
        default=os.environ.get("HF_TOKEN")
    )
    parser.add_argument(
        "--model_id",
        help="HuggingFace model repo to download",
        type=str
    )
    parser.add_argument(
        "--s3_bucket",
        help="S3 bucket name",
        type=str,
        default="cnvrg-test-maas"
    )
    parser.add_argument(
        "--base_s3_path",
        help="Base path at S3 bucket",
        type=str,
        default="models"
    )
    parser.add_argument(
        "--aws_access_key",
        help="Access key for aws",
        type=str,
        default=os.environ.get("AWS_ACCESS_KEY_ID")
    )
    parser.add_argument(
        "--aws_secret_access_key",
        help="Secret access key for aws",
        type=str,
        default=os.environ.get("AWS_SECRET_ACCESS_KEY")
    )
    parser.add_argument(
        "--tmp_dir",
        help="Local dir for tmp files creation",
        type=str,
        default=None
    )

    return parser.parse_args()

def get_streaming_clients(
        hf_token: str,
        aws_access_key: str,
        aws_secret_access_key: str
) -> Tuple[HfApi, BaseClient]:

    hf_api = HfApi(token=hf_token)

    s3_client = boto3.client(
        "s3",
        aws_access_key_id=aws_access_key,
        aws_secret_access_key=aws_secret_access_key
    )

    return hf_api, s3_client

def get_file_size(file_path: str) -> int:
    return os.path.getsize(file_path)

def try_delete_tmp_folder(folder_path: str) -> None:
    time.sleep(5)
    if os.path.exists(folder_path):
        os.remove(folder_path)

def stream_to_s3(
    file_handle: BinaryIO,
    s3_client: BaseClient,
    bucket: str,
    key: str,
    file_size: int
) -> None:

    progress = tqdm(total=file_size, unit='iB', unit_scale=True)

    s3_client.upload_fileobj(
        Fileobj=file_handle,
        Bucket=bucket,
        Key=key,
        Callback=lambda bytes_transferred: progress.update(bytes_transferred)
    )
    progress.close()

def stream_model_to_s3(
    model_id: str,
    s3_bucket: str,
    aws_access_key: str,
    tmp_dir: str,
    aws_secret_access_key: str,
    base_s3_path: Optional[str] = None,
    hf_token: Optional[str] = None
) -> None:

    folder_path = f"{base_s3_path.rstrip('/')}/{model_id.rstrip('/')}"
    hf_api, s3_client = get_streaming_clients(
        hf_token, aws_access_key, aws_secret_access_key
    )
    
    try:
        model_files = hf_api.list_repo_files(model_id)
        logging.info(f"Found {len(model_files)} files in model repository")
        s3_existing_files = get_existing_s3_files(
            s3_client, s3_bucket, folder_path
        )
        logging.info(f"Found {len(s3_existing_files)} existing files to skip")

        for file_info in model_files:
            try:
                with tempfile.TemporaryDirectory(dir=tmp_dir) as temp_dir:
                    if (
                        not is_valid_file_name(file_info) or
                        file_info in s3_existing_files
                    ):
                        logging.info(f"Skipping {file_info}")
                        continue

                    logging.info(f"Processing {file_info}")
                    temp_file = os.path.join(temp_dir, os.path.basename(file_info))
                    local_path = hf_hub_download(
                        repo_id=model_id,
                        filename=file_info,
                        token=hf_token,
                        local_dir=temp_dir,
                        local_dir_use_symlinks=False
                    )

                    file_size = get_file_size(local_path)
                    if (
                        file_info.endswith('.safetensors') and 
                        file_size < 1_000_000_000 # 1GB
                    ):
                        raise ValueError(
                            f"Safetensors file {file_info} seems incomplete:" \
                                + f" {file_size} bytes"
                        )

                    s3_key = f"{folder_path}/{file_info}"
                    logging.info(
                        f"Streaming {file_info} to s3://{s3_bucket}/{s3_key}"
                    )

                    with open(local_path, "rb") as file_handle:
                        stream_to_s3(
                            file_handle=file_handle,
                            s3_client=s3_client,
                            bucket=s3_bucket,
                            key=s3_key,
                            file_size=file_size
                        )

                    logging.info(f"Successfully uploaded {file_info}")
            
            except Exception as e:
                logging.error(f"Error: {str(e)}")
                try_delete_tmp_folder(temp_file)

            # double check for 'with' scope cleanup
            try_delete_tmp_folder(temp_file)
            
    except Exception as e:
        logging.error(f"Error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":

    set_logger()
    args = get_cli_args()

    stream_model_to_s3(
        model_id=args.model_id,
        s3_bucket=args.s3_bucket,
        aws_access_key=args.aws_access_key,
        tmp_dir=args.tmp_dir,
        aws_secret_access_key=args.aws_secret_access_key,
        base_s3_path=args.base_s3_path,
        hf_token=args.hf_token
    )