"""
vault functions to manage secrets
"""

import logging
import os
import re
import hvac

from cryptography.hazmat.primitives import serialization as crypto_serialization



LOGGER = logging.getLogger()

def get_vault_client_instance(vault_addr, vault_token, verify=False, logger=LOGGER):
    """create vault client instance

    Args:
        vault_addr (string): vault address
        vault_token (string): vault token
        verify (bool, optional): verify certificate. Defaults to False.
    """
    logger.info("Creating vault client instance")
    logger.info("Vault server: %s" , vault_addr)
    vault_client = hvac.Client(url=vault_addr, token=vault_token,
                               verify=verify, timeout=45)
    return vault_client

def create_or_update_vault_secret(vault_client, secret_path, mount_point, secret_dict,
                                  logger = LOGGER):
    """create or update a secret in vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point
        secret_dict (dict): secret dictionary
    """
    current_secret_version = 0
    updated_secret_version = 0
    try:
        current_secret_metadata = vault_client.secrets.kv.v2.read_secret_metadata(
                path=secret_path, mount_point=mount_point)
        if current_secret_metadata and current_secret_metadata['data']['current_version']:
            current_secret_version = current_secret_metadata['data']['current_version']
    # If secretpath is not present, set current_secret_version to 0
    # This happens when trying to get metadata before the secret path is created
    except hvac.exceptions.InvalidPath:
        current_secret_version = 0
    logger.info("current version of KVv2 secret at path '%s' is: %s",
                 secret_path, current_secret_version)
    response = vault_client.secrets.kv.v2.create_or_update_secret(
                    path=secret_path, secret=secret_dict,
                    mount_point=mount_point)
    if response:
        logger.debug(("KVv2 secret create/update response. "
                      "path: %s, mount_point:%s, response:%s"),
                      secret_path, mount_point,
                      response)
    updated_secret_metadata = vault_client.secrets.kv.v2.read_secret_metadata(
            path=secret_path, mount_point=mount_point)
    # metadata check after the secret is updated.
    if updated_secret_metadata and updated_secret_metadata['data']['current_version']:
        updated_secret_version = updated_secret_metadata['data']['current_version']

    if updated_secret_version == current_secret_version + 1:
        logger.info("updated version of KVv2 secret at path '%s' is: %s",
                 secret_path, updated_secret_version)
    else:
        logger.error(("KVv2 secret at path %s is failed to update to a newer version "
                      "current version: %s. expected version: %s."), secret_path,
                      updated_secret_version, current_secret_version + 1)

def get_vault_secret(vault_client, secret_path, mount_point, logger = LOGGER):
    """get secret from vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point
    """
    # update log-level if required.

    response = vault_client.secrets.kv.v2.read_secret(
            path=secret_path, mount_point = mount_point)
    if response and response['data'] and response['data']['data']:
        logger.debug(("KVv2 secret get response. "
                      "path: %s, mount_point:%s, response:%s"),
                      secret_path, mount_point, response)
        return response['data']['data']


def delete_vault_secret(vault_client, secret_path, mount_point, logger = LOGGER):
    """get secret from vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point
    """
    # update log-level if required.

    response = vault_client.secrets.kv.v2.delete_metadata_and_all_versions(
            path=secret_path, mount_point=mount_point)
    if response.status_code != 204:
        logger.error("failed to delete secret at path: %s and mount_point: %s",
                secret_path, mount_point)
    logger.debug(("KVv2 secret delete response. "
                "path: %s, mount_point:%s, response:%s"),
                secret_path, mount_point, response.status_code)

def delete_latest_version_of_vault_secret(vault_client, secret_path,
                                          mount_point, logger = LOGGER):
    """create or update a secret in vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point
    """
    current_secret_version = 0
    updated_secret_version = 0
    current_secret_metadata = vault_client.secrets.kv.v2.read_secret_metadata(
        path=secret_path, mount_point=mount_point)
    if current_secret_metadata and current_secret_metadata['data']['current_version']:
        current_secret_version = current_secret_metadata['data']['current_version']
    response = vault_client.secrets.kv.v2.delete_latest_version_of_secret(
                    path=secret_path, mount_point=mount_point)
    if response:
        logger.debug(("KVv2 secret delete latest version response. "
                      "path: %s, mount_point:%s, response:%s"),
                      secret_path, mount_point,
                      response)
    updated_secret_metadata = vault_client.secrets.kv.v2.read_secret_metadata(
            path=secret_path, mount_point=mount_point)
    # metadata check after the secret is updated.
    if updated_secret_metadata and updated_secret_metadata['data']['current_version']:
        updated_secret_version = updated_secret_metadata['data']['current_version']

    if updated_secret_version == current_secret_version - 1:
        logger.info("updated version of KVv2 secret at path '%s' is: %s",
                 secret_path, updated_secret_version)
    else:
        logger.error(("KVv2 secret at path %s is failed to update to a newer version "
                      "current version: %s. expected version: %s."), secret_path,
                      updated_secret_version, current_secret_version - 1)

def get_environment_var(key, default_value=None):
    """
    Get environmental variable value
    Args:
        key (string): environmental variable
        default_value (string): default value to use if env var not present
    """
    env_var_value = os.environ.get(key)
    if not env_var_value and default_value:
        return default_value
    else:
        return env_var_value

def ssh_public_key(key):
    """
    Get public key from an RSAKey
    Args:
        key (paramiko.RSAKey): SSH RSAKey
    """
    return key.public_key().public_bytes(crypto_serialization.Encoding.OpenSSH,
                                         crypto_serialization.PublicFormat.OpenSSH).decode("utf-8")

def ssh_private_key(key):
    """
    Get private key from an RSAKey
    Args:
        key (paramiko.RSAKey): SSH RSAKey
    """
    return key.private_bytes(crypto_serialization.Encoding.PEM,
                             crypto_serialization.PrivateFormat.OpenSSH,
                             crypto_serialization.NoEncryption()).decode("utf-8")

def setup_logger(log_level):
    """ set up logger
    """
    logger = logging.getLogger()
    logger.setLevel(log_level)
    # create console handler
    ch = logging.StreamHandler()
    ch.setLevel(log_level)
    # create formatter
    formatter = logging.Formatter(
        "[%(asctime)s:%(filename)s:%(lineno)s: %(funcName)20s()] - %(levelname)s -- %(message)s", "%Y-%m-%d %H:%M:%S")
    # add formatter to ch
    ch.setFormatter(formatter)
    ch.setFormatter(RedactingFormatter(
        ch.formatter, patterns=[
            r'PGPASSWORD=([^\s]+)',
            r'WITH PASSWORD ([^\s]+)',
            r'\'password\': ([^\s]+)',
            r'\'passwords\': ([^\s]+)',
            r'vault_token=([^\s]+)']))
    # add ch to logger
    logger.addHandler(ch)

    return logger

class RedactingFormatter(object):
    """Redact sensitive data

    """
    def __init__(self, orig_formatter, patterns):
        self.orig_formatter = orig_formatter
        self._patterns = patterns

    def format(self, record):
        """format message record to redact sensitive data
        """
        msg = self.orig_formatter.format(record)
        for pattern in self._patterns:
            msg = re.sub(pattern, "********", msg)
        return msg

    def __getattr__(self, attr):
        return getattr(self.orig_formatter, attr)
