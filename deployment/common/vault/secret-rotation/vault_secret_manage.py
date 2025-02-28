"""
# load secrets into Vault
"""
#!/usr/bin/env python3

import os
import sys
import argparse
import logging
import shutil
import uuid
import yaml
import hvac
import paramiko
import vault_secret_utils

from pathlib import Path

LOGGER = logging.getLogger()

def get_local_secrets_dir(idc_deploy_env):
    """
    Get local secrets directory
    Args:
        idc_deploy_env (string): IDC deployment infra
    """
    if idc_deploy_env:
        return str(Path(os.getcwd()).parents[2]) + '/local/secrets/' + idc_deploy_env
    else:
        return str(Path(os.getcwd()).parents[2]) + '/local/secrets'

def get_ssh_public_key(key):
    """
    Get public key from an RSAKey
    Args:
        key (paramiko.RSAKey): SSH RSAKey
    """
    return vault_secret_utils.ssh_public_key(key)

def get_ssh_private_key(key):
    """
    Get private key from an RSAKey
    Args:
        key (paramiko.RSAKey): SSH RSAKey
    """
    return vault_secret_utils.ssh_private_key(key)

def get_value_from_file(file):
    """
    Get content of a file
    Args:
        file (string): file to read
    """
    with open(file, 'r', encoding='utf-8') as fh:
        data = fh.read()
    return data

def get_uuid():
    """
    Generate a UUID.
    """
    return str(uuid.uuid1())

def get_approle_role_id(vault_client, role_name):
    """
    Get approle role ID
    Args:
        vault_client (havc.Client): vault client instance
        role_name (string): vault approle name
    """
    response = vault_client.auth.approle.read_role_id(
                            role_name=role_name)
    if response:
        return response['data']['role_id']

def get_approle_secret_id(vault_client, role_name):
    """
    Get approle secret ID
    Args:
        vault_client (havc.Client): vault client instance
        role_name (string): vault approle name
    """
    response = vault_client.auth.approle.generate_secret_id(
                            role_name=role_name)
    if response:
        return response['data']['secret_id']

def write_secrets_to_local_file(kv_secret, region, az, idc_deploy_env, secret_dict,
                                harvester_cluster_id = ""):
    """write secret to local file during initial deployment

    Args:
        kv_secret (dict): local
        region (string): IDC region
        az (string): IDC availability zone
        idc_deploy_env (string): IDC deployment environment
        secret_dict (dict): vault secret dictionary
    """
    # Don't update public certificate files
    if kv_secret['type'] == 'public-certificate':
        return
    for secret in kv_secret['secrets']:
        # local secret file to write the secret
        local_secret_file = secret['local_secret_file'].format(
            local_secrets_base_dir=get_local_secrets_dir(idc_deploy_env),
            az=az, region=region, harvester_cluster_id = harvester_cluster_id)
        # create parent directory if doesn't exist
        Path(Path(local_secret_file).parents[0]).mkdir(parents=True, exist_ok=True)
        # write secret to local file
        LOGGER.info("Writing %s secret to file: %s", kv_secret['name'], local_secret_file)
        with open(local_secret_file, 'w', encoding='utf-8') as secret_fh:
            secret_data = secret_dict.get(secret['name'])
            if secret_data:
                secret_fh.write(secret_data)
    return

def load_kv_secrets(vault_client, region, az, idc_deploy_env, kv_secret):
    """load kv secrets based on config file and write to local file
       if it successfully updated.

    Args:
        vault_client (hvac.Client): vault client instance
        region (string): IDC region
        az (string): IDC availability zone
        idc_deploy_env (string): IDC deployment environment
        kv_secret (dict): secret config information
    """
    LOGGER.info("Creating/updating %s secret", kv_secret['name'])
    dispatcher = {"get_ssh_public_key": get_ssh_public_key,
                  "get_ssh_private_key": get_ssh_private_key,
                  "get_value_from_file": get_value_from_file,
                  "get_uuid": get_uuid,
                  "get_approle_secret_id": get_approle_secret_id,
                  "get_approle_role_id": get_approle_role_id}
    # Load harvester secret type
    if kv_secret['type'] == "harvester":
        # Get the harvester kubeconfig folder
        harvester_kubeconfig_folder = (
            get_local_secrets_dir(idc_deploy_env) + '/harvester-kubeconfig')
        Path(harvester_kubeconfig_folder).mkdir(parents=True, exist_ok=True)
        harvester_file_list = Path(harvester_kubeconfig_folder).rglob('*')
        for file in harvester_file_list:
            secret_dict = {}
            base_name = file.name
            secret_path = kv_secret['path'].format(
                region=region, az=az, harvester_cluster_id=base_name)
            LOGGER.debug("KVv2 secret path: %s", secret_path)
            for secret in kv_secret['secrets']:
                if kv_secret['sub_type'] == "kubeconfig":
                    secret_dict |= {secret['name']: dispatcher[(secret['value'])](file),
                                'harvester_cluster_id': base_name}
                elif kv_secret['sub_type'] == "mmws":
                    mmws_file = secret['local_secret_file'].format(
                        local_secrets_base_dir=get_local_secrets_dir(idc_deploy_env))
                    Path(mmws_file).touch(exist_ok= True)
                    secret_dict |= {secret['name']: dispatcher[(secret['value'])](mmws_file)}
            LOGGER.debug("KVv2 secrets data: %s. name: %s, path: %s",
                        secret_dict, kv_secret['name'], secret_path)
            # create or update vault secret
            vault_secret_utils.create_or_update_vault_secret(
                vault_client, secret_path, kv_secret['mount_point'], secret_dict, LOGGER)
            # update local file with the secret
            write_secrets_to_local_file(kv_secret, region, az, idc_deploy_env,
                secret_dict, harvester_cluster_id = base_name)
    # Load all types except harvester
    else:
        secret_dict = {}
        secret_path = kv_secret['path'].format(region=region, az=az)
        LOGGER.debug("KVv2 secret path: %s", secret_path)
        if kv_secret['type'] == "ssh":
            key = paramiko.RSAKey.generate(4096)
        for secret in kv_secret['secrets']:
            if kv_secret['type'] == "static":
                secret_dict |= {secret['name']: secret['value']}
            elif kv_secret['type'] == "environment":
                secret_dict |= {secret['name']: vault_secret_utils.get_environment_var(secret['value'])}
            elif kv_secret['type'] == "ssh":
                secret_dict |= {secret['name']: dispatcher[(secret['value'])](key)}
            elif kv_secret['type'] == "public-certificate":
                file = secret['local_secret_file'].format(pwd=os.getcwd())
                secret_dict |= {secret['name']: dispatcher[(secret['value'])](file)}
            elif kv_secret['type'] == "database":
                secret_dict |= {secret['name']: dispatcher[(secret['value'])]()}
            elif kv_secret['type'] == "approle":
                secret_dict |= {secret['name']: dispatcher[(secret['value'])](
                                            vault_client, kv_secret['role_name'])}
            elif kv_secret['type'] == "kubeconfig":
                file = secret['local_secret_file'].format(
                local_secrets_base_dir=get_local_secrets_dir(idc_deploy_env), az=az)
                secret_dict |= {secret['name']: dispatcher[(secret['value'])](file)}
        LOGGER.debug("KVv2 secrets data: %s. name: %s, path: %s",
                    secret_dict, kv_secret['name'], secret_path)
        # create or update vault secret
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, secret_path, kv_secret['mount_point'], secret_dict, LOGGER)
        # update local file with the secret
        write_secrets_to_local_file(kv_secret, region, az, idc_deploy_env, secret_dict)

def update_kv_secrets_from_local_secrets(vault_client, region, az, idc_deploy_env, kv_secret):
    """reload kv secrets from local secret files

    Args:
        vault_client (hvac.Client): vault client instance
        region (string): IDC region
        az (string): IDC availability zone
        idc_deploy_env (string): IDC deployment environment
        kv_secret (dict): secret configuration data
    """
    # Load harvester secret type
    if kv_secret['type'] == "harvester":
        # Get the harvester kubeconfig folder
        harvester_kubeconfig_folder = (
            get_local_secrets_dir(idc_deploy_env) + '/harvester-kubeconfig')
        Path(harvester_kubeconfig_folder).mkdir(parents=True, exist_ok=True)
        harvester_file_list = Path(harvester_kubeconfig_folder).rglob('*')
        for file in harvester_file_list:
            secret_dict = {}
            base_name = file.name
            secret_path = kv_secret['path'].format(
                region=region, az=az, harvester_cluster_id=base_name)
            LOGGER.debug("KVv2 secret path: %s", secret_path)
            for secret in kv_secret['secrets']:
                local_secret_file = secret['local_secret_file'].format(
                    local_secrets_base_dir=get_local_secrets_dir(idc_deploy_env),
                    az=az, region=region, harvester_cluster_id=base_name)
                secret_dict |= {secret['name']: get_value_from_file(local_secret_file)}
            LOGGER.debug("KVv2 secrets data: %s. name: %s, path: %s",
                    secret_dict, kv_secret['name'], secret_path)
            # create or update vault secret
            vault_secret_utils.create_or_update_vault_secret(
                vault_client, secret_path, kv_secret['mount_point'], secret_dict, LOGGER)
    else:
        secret_dict = {}
        secret_path = kv_secret['path'].format(region=region, az=az)
        for secret in kv_secret['secrets']:
            if kv_secret['type'] == "public-certificate":
                local_secret_file = secret['local_secret_file'].format(pwd=os.getcwd())
            else:
                local_secret_file = secret['local_secret_file'].format(
                    local_secrets_base_dir=get_local_secrets_dir(idc_deploy_env),
                    az=az, region=region)
            secret_dict |= {secret['name']: get_value_from_file(local_secret_file)}
        LOGGER.debug("KVv2 secrets data: %s. name: %s, path: %s",
                    secret_dict, kv_secret['name'], secret_path)
        # create or update vault secret
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, secret_path, kv_secret['mount_point'], secret_dict, LOGGER)

def get_kv_secrets_from_vault(vault_client, region, az, idc_deploy_env, kv_secret):
    """get kv secrets from vault and store into to local secret files

    Args:
        vault_client (hvac.Client): vault client instance
        region (string): IDC region
        az (string): IDC availability zone
        idc_deploy_env (string): IDC deployment environment
        kv_secret (dict): secret configuration data
    """
    if kv_secret['type'] == 'harvester':
        harvester_kubeconfig_folder = (
            get_local_secrets_dir(idc_deploy_env) + '/harvester-kubeconfig')
        Path(harvester_kubeconfig_folder).mkdir(parents=True, exist_ok=True)
        harvester_file_list = Path(harvester_kubeconfig_folder).rglob('*')
        for file in harvester_file_list:
            base_name = file.name
            secret_path = kv_secret['path'].format(region=region, az=az,
                                    harvester_cluster_id=base_name)
            LOGGER.info("Get KVv2 secret from path: %s and mount_point: %s",
                        secret_path,kv_secret['mount_point'])
            # get secret data from vault
            secret_data = vault_secret_utils.get_vault_secret(
                vault_client, secret_path, kv_secret['mount_point'], LOGGER)
            # update local file with the secret
            write_secrets_to_local_file(
                    kv_secret, region, az, idc_deploy_env, secret_data,
                    harvester_cluster_id = base_name)
    else:
        secret_path = kv_secret['path'].format(region=region, az=az)
        LOGGER.info("Get KVv2 secret at path: %s and mount_point: %s",
                    secret_path, kv_secret['mount_point'])
        # get secret data from vault
        secret_data = vault_secret_utils.get_vault_secret(
                vault_client, secret_path, kv_secret['mount_point'], LOGGER)
        # update local file with the secret
        write_secrets_to_local_file(kv_secret, region, az, idc_deploy_env, secret_data)

def main():
    """
    create/update secrets into or get secrets from vault
    """
    parser = argparse.ArgumentParser(description="Load secrets into Vault")
    parser.add_argument("--vault-addr",
        help="Vault server address. Get it from the ENV vars if this flag is not provided")
    parser.add_argument(
        "--vault-token", help="Vault Token. Get it from the ENV vars if this flag is not provided")
    parser.add_argument("--region", default='us-dev-1',help="IDC region")
    parser.add_argument("--availability-zone", default='us-dev-1a',help="IDC availability zone")
    parser.add_argument("--idc-deploy-env", default='', help="IDC deployment environment")
    parser.add_argument('--load-vault-secrets', action=argparse.BooleanOptionalAction,
                        default=False, help=("Use this flag to initially create vault "
                        "secrets based on configuration file."))
    parser.add_argument('--load_vault_secrets_from_local_secret_files',
                        action=argparse.BooleanOptionalAction,
                        default=False, help=("create/update vault secrets from local secret files. "
                        "This option is recommended after populating the local secret files "
                        "using '--load-vault-secrets' or '--get-vault-secrets' flags"))
    parser.add_argument('--get-vault-secrets', action=argparse.BooleanOptionalAction, default=False,
                        help="Get secrets from vault and store it in local secret files")
    parser.add_argument('--update-enrollment-kubeconfig', action=argparse.BooleanOptionalAction,
                        default=True, help=("Copy and update enrollment kubeconfig. "
                                " This option is mainly for Dev kind/rke2 clusters"))
    parser.add_argument('--load-postgres-credentials', action=argparse.BooleanOptionalAction,
                        default=True, help=("Load postgres credentials when set to True. "
                                            "This option is mainly for Dev kind/rke2 clusters"))
    parser.add_argument("--log-level", type=int, default=logging.INFO, help="10=DEBUG,20=INFO")
    if len(sys.argv)==1:
        parser.print_help(sys.stderr)
        sys.exit(1)
    args = parser.parse_args()

    logger = vault_secret_utils.setup_logger(args.log_level)
    globals()["LOGGER"] = logger
    LOGGER.debug("args: %s", str(args))

    # Get vault address and token
    vault_addr =  args.vault_addr if args.vault_addr else vault_secret_utils.get_environment_var('VAULT_ADDR')
    vault_token =  args.vault_token if args.vault_token else vault_secret_utils.get_environment_var('VAULT_TOKEN')
    region = args.region
    az = args.availability_zone
    idc_deploy_env = args.idc_deploy_env
    get_vault_secrets = args.get_vault_secrets
    update_enrollment_kubeconfig = args.update_enrollment_kubeconfig
    load_postgres_credentials = args.load_postgres_credentials
    load_vault_secrets = args.load_vault_secrets
    load_vault_secrets_from_local_secret_files = args.load_vault_secrets_from_local_secret_files
    if not vault_addr:
        LOGGER.error("Failed to get the vault server address. use environmental variable "
                     "VAULT_ADDR or argument --vault-addr to set vault address")
        return 1
    if not vault_token:
        LOGGER.error("Failed to get the vault token address. use environmental variable "
                     "VAULT_TOKEN or argument --vault-token to set vault token")
        return 1
    # get vault config file
    with open('secret-map.yaml', 'r', encoding='utf8') as file:
        vault_config = yaml.safe_load(file)
    LOGGER.debug(vault_config)

    # copy enrollment kubeconfig from ~/.kube/config to secrets folder in development clusters
    if update_enrollment_kubeconfig:
        src = str(Path.home()) + '/.kube/config'
        dst = get_local_secrets_dir(idc_deploy_env) + '/' + az + '/enrollment/kubeconfig'
        Path(Path(dst).parents[0]).mkdir(parents=True, exist_ok=True)
        if Path(src).exists():
            LOGGER.info("Copying kubeconfig from %s to %s", src, dst)
            shutil.copyfile(src, dst)

    # create database admin users and passwords
    if load_postgres_credentials:
        for db in ("authz", "compute", "billing", "metering", "usage", "cloudcredits", "cloudaccount", "productcatalog"):
            # DB admin user
            db_admin_user_file = (
                get_local_secrets_dir(idc_deploy_env) +  f"/{db}_db_admin_username")
            # create parent directory if doesn't exist
            Path(Path(db_admin_user_file).parents[0]).mkdir(parents=True, exist_ok=True)
            with open(db_admin_user_file, 'w', encoding='utf-8') as db_fh:
                db_fh.write('postgres')
            # DB admin password
            if db == "compute":
                db_admin_password_file = (
                    get_local_secrets_dir(idc_deploy_env) + f"/{region}-{db}_db_admin_password")
            else:
                db_admin_password_file = (
                    get_local_secrets_dir(idc_deploy_env) + f"/{db}_db_admin_password")
            db_admin_password = str(uuid.uuid1())
            with open(db_admin_password_file, 'w', encoding='utf-8') as db_fh:
                db_fh.write(db_admin_password)

    try:
        vault_client = vault_secret_utils.get_vault_client_instance(vault_addr, vault_token, logger=LOGGER)
        LOGGER.info("Vault initialize status: %s", vault_client.sys.is_initialized())
        # get secret data from vault and store it to corresponding local secret file
        if get_vault_secrets:
            for kv_secret in vault_config['vault']['secrets']['kv']:
                if not load_postgres_credentials and kv_secret[type] == "database":
                    continue
                get_kv_secrets_from_vault(vault_client, region, az, idc_deploy_env, kv_secret)
        # load all vault secrets from config file
        if load_vault_secrets:
            for kv_secret in vault_config['vault']['secrets']['kv']:
                if not load_postgres_credentials and kv_secret[type] == "database":
                    continue
                load_kv_secrets(vault_client, region, az, idc_deploy_env, kv_secret)
        # reload all vault secrets from local secret files
        if load_vault_secrets_from_local_secret_files:
            for kv_secret in vault_config['vault']['secrets']['kv']:
                if not load_postgres_credentials and kv_secret[type] == "database":
                    continue
                update_kv_secrets_from_local_secrets(
                    vault_client, region, az, idc_deploy_env, kv_secret)
    except hvac.exceptions.VaultError as err:
        LOGGER.error("vault-secret.py failed with error: %s", err)

if __name__ == "__main__":
    sys.exit(main())
