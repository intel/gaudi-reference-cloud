"""
secret rotator
"""

import argparse
import logging
import sys
import uuid
import time
import os
import re
import datetime
import socket
import secrets
import string
import requests
import paramiko
import yaml
import pynetbox
import vault_secret_utils

from io import StringIO
from enum import Enum
from kubernetes import client
from kubernetes import config
from kubernetes.stream import stream
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.backends import default_backend as crypto_default_backend

LOGGER = logging.getLogger()

class ComputeAPIDBUsers(Enum):
    """compute API server DB roles
    """
    USER1 = 'dbuser1'
    USER2 = 'dbuser2'

class BillingDBUsers(Enum):
    """billing DB roles
    """
    USER1 = 'billing_user1'
    USER2 = 'billing_user2'

class MeteringDBUsers(Enum):
    """metering DB roles
    """
    USER1 = 'metering_user1'
    USER2 = 'metering_user2'

class CloudAccountDBUsers(Enum):
    """metering DB roles
    """
    USER1 = 'cloudacct_user1'
    USER2 = 'cloudacct_user2'

class DDIUsers(Enum):
    """DDI roles
    """
    USER1 = 'dns_idcddidev1'
    USER2 = 'dns_idcddidev2'

class SecretRotatorError(Exception):
    """Exception raised for secret rotator errors.

    Attributes:
        message -- error message
    """
    def __init__(self, message="secret rotator error"):
        self.message = message
        super().__init__(self.message)

def update_k8s_client_with_proxy():
    """update k8s client with https proxy
    """
    if os.getenv('HTTPS_PROXY', ""):
        proxy_url = os.getenv('HTTPS_PROXY')
    elif os.getenv('https_proxy', ""):
        proxy_url = os.getenv('https_proxy')
    else:
        LOGGER.error("environment variables 'HTTPS_PROXY' or 'https_proxy' not present. "
                     "will try to access the cluster without setting proxy in k8s client")
        proxy_url = ""
    # update client configuration with proxy if present
    if proxy_url:
        client.Configuration._default.proxy = proxy_url

def get_kube_config_from_file(kubeconfig_file):
    """get kubeconfig from file

    Args:
        kubeconfig_file (string): kubeconfig file

    Returns:
        kubernetes.api_client.ApiClient: generic k8s api client
    """
    with open(kubeconfig_file, encoding='utf-8') as file_h:
        kube_config = yaml.safe_load(file_h)
        # cannot create an api client behind proxy because of the
        # following issue when connecting through proxy
        # https://github.com/kubernetes-client/python/issues/1967
        # api_client = config.new_client_from_config_dict(kube_config)
        config.load_kube_config_from_dict(config_dict=kube_config)

def get_kube_config_from_vault(vault_client, cluster_id):
    """load kubeconfig from vault

    Args:
        vault_client (hvac.Client): vault client instance
        cluster_id (string): cluster secret_path in vault

    Returns:
        kubernetes.api_client.ApiClient: generic k8s api client
    """
    kubeconfig_mount_point = "kubeconfigs"
    cluster_info = vault_secret_utils.get_vault_secret(
        vault_client, cluster_id, kubeconfig_mount_point, logger=LOGGER)
    if cluster_info['kubeconfig']:
        kube_config = yaml.safe_load(cluster_info['kubeconfig'])
    else:
        err_msg = ("failed to get compute DB cluster kubeconfig from vault. "
                   f"mount_point: {kubeconfig_mount_point}, secret_path: {cluster_id}")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)
    # cannot create an api client behind proxy because of the
    # following issue when connecting through proxy
    # https://github.com/kubernetes-client/python/issues/1967
    # api_client = config.new_client_from_config_dict(kube_config)
    config.load_kube_config_from_dict(config_dict=kube_config)

def get_kube_config(vault_client, rotation_args, cluster_kubeconfig_path):
    """get k8s client

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
        cluster_kubeconfig_path (string): vault secret path to get cluster kubeconfig
                                          when fetching the kubeconfig from vault/
                                          kubeconfig file location when using a local file
    """
    if  rotation_args.fetch_kubeconfigs_from_vault:
        get_kube_config_from_vault(vault_client, cluster_kubeconfig_path)
    else:
        get_kube_config_from_file(cluster_kubeconfig_path)

    if rotation_args.connect_to_k8s_cluster_via_proxy:
        update_k8s_client_with_proxy()

def update_role_password(roles_data, role, ddi=False):
    """update DB role password

    Args:
        roles_data (dict): DB auth data
        role (string): role name to change the password

    Returns:
        dict: DB auth data with updated password
        string: new password of the  role
    """
    available_roles = roles_data['usernames'].split(',')
    available_passwords = roles_data['passwords'].split(',')
    role_index = available_roles.index(role)
    # update role password
    # ddi(m&m) supports max upto 20 characters, so not using uuid4
    if ddi:
        new_role_password = ddi_password()
    else:
        new_role_password = str(uuid.uuid4())
    available_passwords[role_index] = new_role_password
    roles_data['passwords'] = (",".join(available_passwords))
    return (roles_data, new_role_password)

def get_inactive_role(active_user, user_type):
    """get inactive user
    Args:
        active_db_user (string): current active DB user
        db_user_type (string): DB user type

    Returns:
        string: inactive DB username
    """
    if user_type == 'compute_api_server':
        Users = ComputeAPIDBUsers
    elif user_type == 'billing':
        Users = BillingDBUsers
    elif user_type == 'metering':
        Users = MeteringDBUsers
    elif user_type == 'cloudaccount':
        Users = CloudAccountDBUsers
    elif user_type == 'ddi':
        Users = DDIUsers
    else:
        Users = None
    if active_user[-1] == '1':
        LOGGER.info('current inactive user %(user)s', {'user': Users.USER2.value})
        return Users.USER2.value

    LOGGER.info('current inactive user %(user)s', {'user': Users.USER1.value})
    return Users.USER1.value

def ddi_password(pwd_length=20):
    """generate ddi password

    Returns:
        string: random password string
    """
    letters = string.ascii_letters
    digits = string.digits
    special_chars = '!#$%&()*+-:;<=>?@[]^_{|}~'
    # generate a password string
    pwd = ''
    for _ in range(pwd_length):
        pwd += ''.join(secrets.choice(letters + digits + special_chars))

    return pwd

def get_active_role(vault_client, secret_path, mount_point):
    """get active user/role of an application from vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        string: active user
    """

    current_user_data = vault_secret_utils.get_vault_secret(
        vault_client, secret_path, mount_point, logger=LOGGER)

    if 'username' in current_user_data and current_user_data['username']:
        LOGGER.info('current active user: %(user)s', {'user': current_user_data['username']})
    else:
        err_msg = 'failed to get the current DB user'
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)

    return current_user_data['username']

def get_available_roles(vault_client, secret_path, mount_point):
    """get list of available users and passwords from vault

    Args:
        vault_client (hvac.Client): vault client instance
        secret_path (string): vault secret path
        mount_point (string): vault mount point

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        dict: available usernames and passwords
    """

    auth_data = vault_secret_utils.get_vault_secret(
        vault_client, secret_path, mount_point, logger=LOGGER)

    if ('usernames' in auth_data and auth_data['usernames'] and
        'passwords' in auth_data and auth_data['passwords'] and
        len(auth_data['usernames'].split(',')) == len(auth_data['passwords'].split(','))):

        LOGGER.info('available users: %(users)s', {'users': auth_data['usernames']})
        return auth_data
    else:
        err_msg = ("failed to get available usernames and passwords from vault "
                   f"secret path: {secret_path}, mount point: {mount_point}")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)

def get_pod_name(api_client, namespace, label):
    """get pod name using label selector

    Args:
        api_client : k8s api client
        namespace (string): pod namespace
        label (string): labels to select pod

    Returns:
        string: pod name
    """

    pod_list = api_client.list_namespaced_pod(
        namespace, label_selector=label, watch=False)
    pod_name = ""
    for pod in pod_list.items:
        pod_name = pod.metadata.name

    return pod_name

def get_pod_status(api_client, namespace, label):
    """get pod status

    Args:
        api_client : k8s api client
        namespace (string): pod namespace
        label (string): labels to select pod

    Returns:
        string: pod name
    """

    pod_list = api_client.list_namespaced_pod(
        namespace, label_selector=label, watch=False)
    pod_status = ""
    for pod in pod_list.items:
        pod_status = pod.status.phase

    return pod_status

def create_postgres_client_pod(rotation_args):
    """create postgresql client pod if it doesn't exist

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:

        pod_namespace = "idcs-system"
        pod_name = "postgres-client-secret-rotator"
        pod_labels = {"app": "secret-rotator"}
        container_name = "postgres-client"
        container_image = "postgres:15"
        container_command = ["sleep", "infinity"]
        label_selector = 'app=secret-rotator'

        get_kube_config_from_file(rotation_args.eks_kubeconfig)
        if rotation_args.connect_to_k8s_cluster_via_proxy:
            update_k8s_client_with_proxy()
        core_v1_api = client.CoreV1Api()
        # check if pod is present
        postgres_pod_name = get_pod_name(core_v1_api, pod_namespace, label_selector)
        postgres_pod_status = get_pod_status(core_v1_api, pod_namespace, label_selector)
        logging.info("postgres client pod status: %(status)s",
                     {'status': postgres_pod_status})
        if postgres_pod_name and postgres_pod_status == "Running":
            LOGGER.info("postgres client pod '%(pod)s' is at running status in namespace "
                        "%(ns)s", {'pod': pod_name, 'ns': pod_namespace})
            return
        else:
            core_v1_api.delete_namespaced_pod(postgres_pod_name, pod_namespace,
                                              grace_period_seconds = 0)
            time.sleep(2)
        container = client.V1Container(
            name=container_name, image=container_image, command=container_command)
        pod_spec = client.V1PodSpec(containers=[container])
        pod = client.V1Pod(api_version="v1", kind="Pod", metadata=client.V1ObjectMeta(
                name=pod_name, namespace=pod_namespace, labels=pod_labels),
                spec=pod_spec)
        resp = core_v1_api.create_namespaced_pod(namespace=pod_namespace, body=pod)
        LOGGER.info("created postgres client pod %(pod)s at namespace %(ns)s",
                    {'pod': pod_name, 'ns': pod_namespace})
        LOGGER.debug("postgres client pod creation response: %(resp)s", {'resp': resp})
        for _ in range(10):
            pod_status = get_pod_status(core_v1_api, pod_namespace, label_selector)
            if pod_status == "Running":
                logging.info("postgres client pod status: %(status)s",
                     {'status': pod_status})
                return
            else:
                logging.error("postgres client pod not in running status. "
                              "current status: %(status)s", {'status': pod_status})
                time.sleep(10)
        err_msg = ("postgres client pod not in running state")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)

    except client.rest.ApiException as err:
        LOGGER.error("postgres client pod creation failed with error: %s", err)
        raise SecretRotatorError(err) from err

def ssh_operator_args_check(args):
    """check if required arguments to rotate ssh key pair is present

    Args:
        args (argparse.Namespace): secret rotator input arguments
    """
    if args.rotate_ssh_proxy_operator_keys and len(args.ssh_proxy_addresses) == 0:
        LOGGER.error("ssh proxy server address list is empty. use argument "
        "--ssh-proxy-addresses to set the list of ssh proxy server addresses")
        sys.exit(1)
    if args.rotate_bm_ssh_proxy_operator_keys and len(args.bm_ssh_proxy_addresses) == 0:
        LOGGER.error("bm ssh proxy server address list is empty. use argument "
        "--bm-ssh-proxy-addresses to set the list of ssh proxy server addresses")
        sys.exit(1)

def global_config_args_check(args):
    """check if required mandatory parameters are part of global
       cluster input arguments

    Args:
        args (argparse.Namespace): secret rotator input arguments
    """
    if not args.rds_superuser:
        LOGGER.error("Failed to get the rds superuser to perform DB role operators "
                    "at the global cluster. use environmental variable "
                    "RDS_SUPERUSER or argument --rds-superuser to set rds superuser")
        sys.exit(1)
    if not args.rds_superuser_password:
        LOGGER.error("Failed to get the rds superuser password to perform DB role "
                    "operators at the global cluster. use environmental variable "
                    "RDS_SUPERUSER_PASSWORD or argument --rds-superuser-password "
                    "to set rds superuser")
        sys.exit(1)
    if not args.rds_host:
        LOGGER.error("Failed to get the rds host to perform DB role "
                    "operators at the global cluster. use environmental variable "
                    "RDS_HOST or argument --rds-host to set rds host")
        sys.exit(1)

def netbox_token_args_check(args):
    """check for the required netbox token arguments

    Args:
        args (argparse.Namespace): secret rotator input arguments
    """

    if not args.netbox_address:
        LOGGER.error("Failed to get the netbox address. use environmental variable "
                    "NETBOX_ADDR or argument --netbox-address to set netbox address")
        sys.exit(1)
    if not args.fetch_netbox_credentials_from_vault and not args.netbox_username:
        LOGGER.error("Failed to get the netbox admin user. use environmental variable "
                    "NETBOX_USER or argument --netbox-username to set netbox admin user")
        sys.exit(1)
    if not args.fetch_netbox_credentials_from_vault and not args.netbox_password:
        LOGGER.error("Failed to get the netbox admin password. use environmental variable "
                    "NETBOX_PASSWORD or argument --netbox-password to set netbox admin password")
        sys.exit(1)


def ddi_args_check(args):
    """check for the required ddi arguments

    Args:
        args (argparse.Namespace): secret rotator input arguments
    """

    if not args.ddi_url:
        LOGGER.error("Failed to get the ddi url. use environmental variable "
                    "DDI_URL or argument --ddi-url to set DDI URL")
        sys.exit(1)
    if not args.ddi_server_address:
        LOGGER.error("Failed to get the ddi address. use environmental variable "
                    "DDI_URL or argument --ddi-server-address to set "
                    "DDI server address")
        sys.exit(1)
    if (not args.fetch_ddi_admin_credentials_from_vault and
                not args.ddi_admin_username):
        LOGGER.error("Failed to get the ddi admin user. use environmental variable "
                    "DDI_ADMIN_USER or argument --ddi-admin-username to set "
                    "DDI admin user")
        sys.exit(1)
    if (not args.fetch_ddi_admin_credentials_from_vault and
                not args.ddi_admin_password):
        LOGGER.error("Failed to get the ddi admin password. use environmental variable "
                    "DDI_ADMIN_PASSWORD or argument --ddi-admin-password to set "
                    "DDI admin password")
        sys.exit(1)

def get_netbox_credentials_from_vault(vault_client, rotation_args):
    """get netbox credentials from vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        (string, string): netbox username, netbox password
    """

    netbox_credentials_secret_path = f"{rotation_args.region_name}/netbox"
    netbox_credentials_mount_point = "controlplane"
    LOGGER.info("getting netbox secret data from secret path: %(path)s, mount_point: %(mp)s "
                "and region %(region)s", {'path': netbox_credentials_secret_path,
                'mp': netbox_credentials_mount_point, 'region': rotation_args.region_name})
    netbox_secrets = vault_secret_utils.get_vault_secret(
        vault_client, netbox_credentials_secret_path,
        netbox_credentials_mount_point, logger=LOGGER)
    nb_user = netbox_secrets.get('netbox_admin_user', None)
    nb_password = netbox_secrets.get('netbox_admin_password', None)
    if not nb_user or not nb_password:
        err_msg = ("failed to get netbox user or password from vault."
                   f"username: {nb_user}, password: {nb_password}")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)
    return nb_user, nb_password

def get_netbox_token_from_vault(vault_client, rotation_args):
    """get current netbox token from vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        string: netbox token from vault
    """

    netbox_token_secret_path = f"{rotation_args.region_name}/baremetal/enrollment/netbox"
    netbox_token_mount_point = "controlplane"
    LOGGER.info("getting netbox token from secret path: %(path)s, mount_point: %(mp)s "
                "and region %(region)s", {'path': netbox_token_secret_path,
                'mp': netbox_token_mount_point, 'region': rotation_args.region_name})
    netbox_token_data = vault_secret_utils.get_vault_secret(
        vault_client, netbox_token_secret_path, netbox_token_mount_point, logger=LOGGER)
    nb_token = netbox_token_data.get('token', None)
    if not nb_token:
        err_msg = ("failed to get netbox token from vault."
                   f"token: {nb_token}")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)
    return nb_token

def create_netbox_token(rotation_args, nb_username, nb_password):
    """create a new netbox token

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        nb_username (string): netbox admin user
        nb_password (string): netbox admin password

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        string: netbox token
    """
    try:
        nb_conn = pynetbox.api(url=rotation_args.netbox_address)
        LOGGER.info("creating new token in netbox %(nb)s with %(user)s", {
                'nb': rotation_args.netbox_address, 'user': nb_username})
        nb_token = nb_conn.create_token(nb_username, nb_password)
        return str(nb_token)

    except Exception as err:
        LOGGER.error("create_netbox_token failed with error: %(err)s", {'err': err})
        raise SecretRotatorError(err) from err

def verify_netbox_token(rotation_args, nb_token):
    """verify new netbox token

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        nb_token (string): netbox token

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        nb_conn = pynetbox.api(url=rotation_args.netbox_address, token=nb_token)
        nb_conn.dcim.devices.get(name="device-1")

    except pynetbox.core.query.RequestError as err:
        LOGGER.error("verify_netbox_token failed with error: %(err)s", {'err': err})
        raise SecretRotatorError(err) from err


def delete_netbox_token(rotation_args, current_token, old_token):
    """ delete netbox token

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        current_token (string): current netbox token
        old_token (string): old netbox token to delete

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        nb_conn = pynetbox.api(url=rotation_args.netbox_address, token=current_token)
        all_tokens = nb_conn.users.tokens.all()
        for token in all_tokens:
            if old_token == str(token):
                LOGGER.info("deleting old netbox token %(token)s", {'token': old_token})
                token.delete()

    except pynetbox.core.query.RequestError as err:
        LOGGER.error("verify_netbox_token failed with error: %(err)s", {'err': err})
        raise SecretRotatorError(err) from err

def wait_for_deployment_complete(k8s_client, deployment_name, namespace, timeout=360):
    try:
        start = time.time()
        while time.time() - start < timeout:
            time.sleep(10)
            response = k8s_client.read_namespaced_deployment_status(deployment_name, namespace)
            s = response.status
            if (s.updated_replicas == response.spec.replicas and
                    s.replicas == response.spec.replicas and
                    s.available_replicas == response.spec.replicas and
                    s.observed_generation >= response.metadata.generation):
                return True
            else:
                LOGGER.info("rollout status of deployment: %s", deployment_name)
                LOGGER.info("updated_replicas: %(update_replicas)s, replicas: %(replicas)s, "
                "available_replicas: %(available_replicas)s , "
                "observed_generation: %(observed_generation)s. will recheck the rollout status in 10 seconds..",
                {'update_replicas': s.updated_replicas, 'replicas': s.replicas,
                'available_replicas': s.available_replicas,
                'observed_generation': s.observed_generation})

        raise SecretRotatorError(f'Waiting timeout for deployment {deployment_name} in namespace {namespace}')

    except client.rest.ApiException as err:
        err_msg = f"failed to verify deployment status after rollout: {deployment_name} at namespace {namespace}"
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg) from err

def rollout_deployment(k8s_client, deployment_name, namespace):
    """rollout k8s deployment

    Args:
        k8s_client (client.AppsV1Api): k8s client
        deployment_name (string): k8s depoyment name
        namespace (string): k8s namespace

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        now = datetime.datetime.utcnow()
        now = str(now.isoformat("T") + "Z")
        body = {
            'spec': {
                'template':{
                    'metadata': {
                        'annotations': {
                            'kubectl.kubernetes.io/restartedAt': now
                        }
                    }
                }
            }
        }
        LOGGER.info("setting annotation to rollout deployment %(dep)s in namespace %(ns)s", {
            'dep': deployment_name, 'ns': namespace})
        deployment_patch_op = k8s_client.patch_namespaced_deployment(
            deployment_name, namespace, body, pretty='true')
        LOGGER.debug(deployment_patch_op)
    except client.rest.ApiException as err:
        err_msg = f"failed to rollout deployment: {deployment_name} at namespace {namespace}"
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg) from err

def verify_pod_running_status(core_v1_api, pod_label, namespace):
    """verify pod status is set to Running

    Args:
        core_v1_api (client.CoreV1Api): k8s client
        pod_label (string): pod label selector
        namespace (string): k8s namespace

    Raises:
        SecretRotatorError: secret rotation error
    """

    monitor_end = time.time() + 360
    pod_running =  True

    try:
        while time.time() < monitor_end:
            pod_list = core_v1_api.list_namespaced_pod(
                namespace, label_selector=pod_label, watch=False)
            for pod in pod_list.items:
                if pod.status.phase == "Running":
                    LOGGER.info("current status of pod %(pod)s in namespace %(ns)s is Running",
                        {'pod': pod.metadata.name, 'ns': namespace })
                    pod_running =  True
                else:
                    LOGGER.info("current status of pod %(pod)s in namespace %(ns)s is %(status)s. "
                                "will re-check the status in 10 seconds", {
                        'pod': pod.metadata.name, 'ns': namespace, 'status': pod.status.phase})
                    pod_running =  False

            if pod_running:
                break
            else:
                time.sleep(10)
        else:
            err_msg = f"failed to verify status of pods in namespace {namespace}"
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

    except client.rest.ApiException as err:
        err_msg = f"failed to verify status of pods in namespace {namespace}"
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg) from err

def restart_bm_enrollment_pod(vault_client, rotation_args):
    """restart netbox pods after updating the token in vault.

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """

    if rotation_args.fetch_kubeconfigs_from_vault:
        cluster_kubeconfig = rotation_args.compute_cluster_id
    else:
        cluster_kubeconfig = rotation_args.compute_cluster_kubeconfig
    get_kube_config(vault_client, rotation_args, cluster_kubeconfig)

    v1_apps = client.AppsV1Api()
    core_v1_api = client.CoreV1Api()
    bm_enrollment_deployment_name = f"{rotation_args.region_name}-baremetal-enrollment-api"
    bm_enrollment_namespace = "idcs-enrollment"
    bm_enrollment_label = f"app={rotation_args.region_name}-baremetal-enrollment-api"
    rollout_deployment(v1_apps, bm_enrollment_deployment_name, bm_enrollment_namespace)
    verify_pod_running_status(core_v1_api, bm_enrollment_label, bm_enrollment_namespace)


def rotate_netbox_token(vault_client, rotation_args):
    """rotate netbox token

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error
    """
    # get netbox credentials from vault
    # get current netbox token
    # create a new token in netbox
    # verify new token
    # update vault with new token
    # restart baremetal enrollment api server
    # delete old token from netbox
    try:
        # get netbox credentials from vault
        if rotation_args.fetch_netbox_credentials_from_vault:
            nb_username, nb_password =  get_netbox_credentials_from_vault(
                vault_client, rotation_args)
        else:
            nb_username = rotation_args.netbox_username
            nb_password = rotation_args.netbox_password

        # get current netbox token that baremetal enrollment uses
        nb_current_token = get_netbox_token_from_vault(vault_client, rotation_args)

        # create a new netbox token
        nb_new_token = create_netbox_token(rotation_args, nb_username, nb_password)

        # verify_new_token
        verify_netbox_token(rotation_args, nb_new_token)

        # update baremetal enrollment api secret path in vault with new token
        new_token_data = {'token': nb_new_token}
        netbox_token_secret_path = f"{rotation_args.region_name}/baremetal/enrollment/netbox"
        netbox_token_mount_point = "controlplane"
        LOGGER.info("updating the netbox token at secret path: %(path)s, "
                    "mount point: %(mp)s and region %(region)s",
                    {'path': netbox_token_secret_path, 'mp': netbox_token_mount_point,
                    'region': rotation_args.region_name})
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, netbox_token_secret_path, netbox_token_mount_point,
            new_token_data, logger = LOGGER)

        # restart baremetal enrollment api pods
        restart_bm_enrollment_pod(vault_client, rotation_args)

        # delete previous netbox token from netbox
        delete_netbox_token(rotation_args, nb_new_token, nb_current_token)

    except Exception as err:
        LOGGER.error("rotate_netbox_token failed with error: %(err)s", {'err': err})
        raise SecretRotatorError(err) from err

def get_ssh_keys_from_vault(vault_client, rotation_args, secret_path, mount_point):
    """get ssh keys from vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
        secret_path (string): vault secret path
        mount_point (string): vault mount point

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        (string, string): ssh private key, public key
    """

    LOGGER.info("getting ssh key pair from secret path: %(path)s, mount_point: %(mp)s "
                "and AZ: %(az)s", {'path': secret_path, 'mp': mount_point,
                'az': rotation_args.az_name})
    ssh_key_data = vault_secret_utils.get_vault_secret(
        vault_client, secret_path, mount_point, logger=LOGGER)
    private_key = ssh_key_data.get('privatekey', None)
    public_key = ssh_key_data.get('publickey', None)
    if not private_key or not private_key:
        err_msg = "failed to get ssh key pair from vault."
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)

    return private_key, public_key

def update_proxy_auth_key(ssh_client, ssh_proxy, old_public_key, new_public_key, delete_old_key):
    """update authorized key file of SSH proxy server

    Args:
        ssh_client (paramiko.client.SSHClient): ssh client
        ssh_proxy (string): ssh proxy server address
        old_public_key (str): old public key.
        new_public_key (str): new public key to update in authorized key file.
        delete_old_key (bool): (delete old key from authorized key file when set to true.
                                    Defaults to False.)

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        LOGGER.info("updating authorized keys file in SSH proxy server %(srv)s", {'srv': ssh_proxy})
        LOGGER.debug("old_public_key: %s", old_public_key)
        LOGGER.debug("new_public_key: %s", new_public_key)
        if delete_old_key:
            cmd = f"sed -i '\,{old_public_key.strip()},d' .ssh/authorized_keys"
        else:
            # insert new public key before old public key in authorized_keys file
            cmd = f"sed -i '\,{old_public_key.strip()},i {new_public_key}\n' .ssh/authorized_keys"
        LOGGER.debug("cmd: %s", cmd)
        _, _stdout,_stderr = ssh_client.exec_command(cmd)
        stdout =  _stdout.read().decode()
        ret_code = _stdout.channel.recv_exit_status()
        stderr = _stderr.read().decode()
        LOGGER.debug("stdout: %s", stdout)
        LOGGER.debug("stderr: %s", stderr)

        # raise if failed to update the ssh key
        if stderr or ret_code != 0:
            err_msg = (f"failed to update the key in ssh proxy server {ssh_proxy}. "
                       f" stderr : {stderr}, return code: {ret_code}")
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

    except (paramiko.BadHostKeyException, paramiko.AuthenticationException, paramiko.SSHException,
            paramiko.ssh_exception.NoValidConnectionsError, socket.error) as err:
        LOGGER.error("update_proxy_auth_key failed with error: %(err)s  ", {'err': err})
        raise SecretRotatorError(err) from err

def ssh_client_to_proxy(rotation_args, private_key, application, old_public_key='',
                        new_public_key='', update_key=False, delete_old_key = False):
    """ssh client to connect, authenticate and run_commands to the ssh proxy server/s

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        private_key (string): private key to connect to the ssh proxy server
        application (string): application type
        old_public_key (str, optional): old public key. Defaults to ''.
        new_public_key (str, optional): new public key to update in authorized key file. Defaults to ''.
        update_key (bool, optional): (update authorized key file when set to true.
                                    Defaults to False.)
        delete_old_key (bool, optional): (delete old key from authorized key file when set to true.
                                    Defaults to False.)

    Returns:
        bool: ssh connection status
    """
    if application == "ssh_proxy_operator":
        ssh_proxy_servers = rotation_args.ssh_proxy_addresses
        ssh_user = rotation_args.ssh_proxy_username
    else:
        ssh_proxy_servers = rotation_args.bm_ssh_proxy_addresses
        ssh_user = rotation_args.bm_ssh_proxy_username
    try:
        for ssh_proxy in ssh_proxy_servers:
            # connect to ssh proxy server
            private_key_obj = StringIO(private_key)
            pkey = paramiko.RSAKey.from_private_key(private_key_obj)
            ssh_client = paramiko.SSHClient()
            ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            LOGGER.info("connecting to ssh proxy server %(srv)s", {'srv': ssh_proxy})
            ssh_client.connect(ssh_proxy, username=ssh_user, pkey=pkey)
            if update_key:
                update_proxy_auth_key(
                    ssh_client, ssh_proxy, old_public_key, new_public_key, delete_old_key)
            ssh_client.close()
        return True

    except (paramiko.BadHostKeyException, paramiko.AuthenticationException,
            paramiko.SSHException, paramiko.ssh_exception.NoValidConnectionsError,
            socket.error) as err:
        LOGGER.error("ssh_client_to_proxy failed with error: %(err)s  ", {'err': err})
        return False

def rotate_ssh_keys(vault_client, rotation_args, application):
    """rotate ssh private and public keys

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
        application (string): application type

    Raises:
        SecretRotatorError: secret rotation error
    """
    # get current ssh key pair
    # create new ssh key pair
    # update ssh proxy with new public key
    # update vault with new ssh key pair
    # restart ssh proxy operator/bm-instance operator
    # verify ssh access to the proxy server with new keys
    # if failed to verify ssh access with new keys-
    ## (ssh verification can fail if an instance creation/deletion occurs after
    ## updating the ssh proxy authorized key with new public key)
    ##   - ssh to the server with old keys
    ##   - update ssh proxy with new public key
    # verify operator pod is running

    try:
        mount_point = "controlplane"
        operator_namespace = "idcs-system"
        if application == "ssh_proxy_operator":
            secret_path = f"{rotation_args.az_name}-ssh-proxy-operator/ssh"
            deployment_name = f"{rotation_args.az_name}-ssh-proxy-operator"
        else:
            secret_path = f"{rotation_args.az_name}-bm-instance-operator/ssh"
            deployment_name = f"{rotation_args.az_name}-bm-instance-operator"

        # get current ssh key-pair
        LOGGER.info("getting current ssh key pair from vault")
        current_private_key, current_public_key = get_ssh_keys_from_vault(
            vault_client, rotation_args, secret_path, mount_point)

        # generate new key pair
        key = rsa.generate_private_key(backend=crypto_default_backend(), public_exponent=65537, key_size=4096)
        new_private_key = vault_secret_utils.ssh_private_key(key)
        public_key = vault_secret_utils.ssh_public_key(key)
        fmt = "%Y-%m-%dT%H-%M-%S%z"
        date_time = datetime.datetime.now().strftime(fmt)
        new_public_key = public_key + f" idc-{date_time} \n"

        # update ssh proxy with the new public key
        LOGGER.info("updating SSH proxy server/s with the new public key")
        conn_status = ssh_client_to_proxy(rotation_args, current_private_key, application,
                            old_public_key=current_public_key,
                            new_public_key=new_public_key, update_key= True)
        if not conn_status:
            err_msg = "failed to communicate to ssh proxy server"
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)
        # update new key_pair in vault
        new_ssh_key_pair = {'publickey': new_public_key, 'privatekey': new_private_key}
        LOGGER.info("updating vault with new ssh keys. secret path: %(path)s, mount point: %(mp)s "
                    "and AZ: %(region)s", {'path': secret_path, 'mp': mount_point,
                    'region': rotation_args.az_name})
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, secret_path, mount_point, new_ssh_key_pair, logger = LOGGER)

        if rotation_args.fetch_kubeconfigs_from_vault:
            cluster_kubeconfig = rotation_args.az_cluster_id
        else:
            cluster_kubeconfig = rotation_args.az_cluster_kubeconfig
        get_kube_config(vault_client, rotation_args, cluster_kubeconfig)

        v1_apps = client.AppsV1Api()
        # restart ssh operator
        LOGGER.info("restart %s pod to use the new key pair from vault", application)
        rollout_deployment(v1_apps, deployment_name, operator_namespace)
        wait_for_deployment_complete(v1_apps, deployment_name, operator_namespace)
        LOGGER.info("verifying access to the ssh proxy server using new key pair")
        ssh_proxy_connection_status = ssh_client_to_proxy(rotation_args, new_private_key, application)
        if ssh_proxy_connection_status:
            LOGGER.info("updated ssh proxy authorized keys file with new key %s", new_public_key)
        else:
            # - update ssh key in case the instance creation/deletion removes
            # the above update of the authorized key file
            # - since proxy operator pod is restarted as part of rollout, proxy operator
            # pod cannot connect to the proxy server via ssh/scp with the new key pair.
            # updating the authorized key file using the old private key
            LOGGER.info("retry: updating SSH proxy server/s with the new public key")
            conn_status = ssh_client_to_proxy(rotation_args, current_private_key, application,
                old_public_key=current_public_key, new_public_key=new_public_key, update_key= True)
            # check connectivity to the ssh proxy servers with the new key
            LOGGER.info("retry: verifying access to the ssh proxy server using new key pair")
            ssh_proxy_connection_status_update = ssh_client_to_proxy(rotation_args, new_private_key, application)
            # if ssh_proxy_connection_status_update is true, wait for the pod to return to active status
            # else:
            # revert the vault secret to the previous version and restart the pod
            if not conn_status or not ssh_proxy_connection_status_update:
                LOGGER.info("reverting SSH key pair in vault as the ssh key rotation is failed")
                ssh_key_pair = {'publickey': current_public_key, 'privatekey': current_private_key}
                vault_secret_utils.create_or_update_vault_secret(
                    vault_client, secret_path, mount_point, ssh_key_pair, logger = LOGGER)
            rollout_deployment(v1_apps, deployment_name, operator_namespace)
            wait_for_deployment_complete(v1_apps, deployment_name, operator_namespace)
        # SSH proxy operator automatically removes the old key from authorized key
        # after the update. need to remove it for BM SSH proxy server
        if application == "bm_ssh_proxy_operator":
            LOGGER.info("removing old key for BM SSH proxy server")
            conn_status = ssh_client_to_proxy(rotation_args, new_private_key, application,
                            old_public_key=current_public_key, update_key= True, delete_old_key= True)

    except Exception as err:
        LOGGER.error("rotate_ssh_keys failed with error: %(err)s  ", {'err': err})
        raise SecretRotatorError(err) from err

def get_mm_credentials_from_vault(vault_client, rotation_args):
    """get ddi admin credentials from vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error

    Returns:
        (string, string): username, password
    """
    mmws_credentials_secret_path = f"{rotation_args.region_name}/mmws/admin"
    mmws_credentials_mount_point = "controlplane"
    LOGGER.info("getting ddi admin credentials from vault. secret path: %(path)s, "
                "mount_point: %(mp)s, region: %(region)s and AZ: %(az)s",
                {'path': mmws_credentials_secret_path, 'mp': mmws_credentials_mount_point,
                 'region': rotation_args.region_name, 'az': rotation_args.az_name})
    mmws_secrets = vault_secret_utils.get_vault_secret(
        vault_client, mmws_credentials_secret_path, mmws_credentials_mount_point, logger=LOGGER)
    mmws_user = mmws_secrets.get('username', None)
    mmws_password = mmws_secrets.get('password', None)
    if not mmws_user or not mmws_password:
        err_msg = ("failed to get ddi admin user or password from vault."
                   f"username: {mmws_user}, password: {mmws_password}")
        LOGGER.error(err_msg)
        raise SecretRotatorError(err_msg)
    return mmws_user, mmws_password

def update_ddi_user_password(rotation_args, mm_user, mm_user_password, mm_admin_user,
                             mm_admin_password):
    """update DDI user password

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        mm_user (string): ddi user
        mm_user_password (string): updated password of mm_user
        mm_admin_user (string): ddi admin user
        mm_admin_password (string): ddi admin password

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        data = {"deleteUnspecified": False,"properties": {"password": mm_user_password }}
        params = {"server": rotation_args.ddi_server_address}
        url = f"https://{rotation_args.ddi_url}/mmws/api/users/{mm_user}"
        resp = requests.put(url, auth=(mm_admin_user, mm_admin_password),
                            json=data, params=params, timeout=60, verify= False)
        if resp.status_code != 204:
            raise SecretRotatorError(f"failed to change password ddi user {mm_user}")
    except Exception as err:
        LOGGER.error("update_ddi_user_password failed with error: %(err)s  ", {'err': err})
        raise SecretRotatorError(err) from err

def rotate_ddi_user(vault_client, rotation_args):
    """rotate ddi user and update passwords

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    LOGGER.info(rotation_args)
    try:
        if rotation_args.fetch_kubeconfigs_from_vault:
            cluster_kubeconfig = rotation_args.az_cluster_id
        else:
            cluster_kubeconfig = rotation_args.az_cluster_kubeconfig
        get_kube_config(vault_client, rotation_args, cluster_kubeconfig)
        v1_apps = client.AppsV1Api()
        mm_user_secret_path = f"{rotation_args.region_name}/baremetal/enrollment/menandmice"
        mm_users_secret_path = f"{rotation_args.region_name}/mmws/customer"
        mm_mount_point = "controlplane"
        user_type = "ddi"
        # get ddi admin credentials from vault
        if rotation_args.fetch_ddi_admin_credentials_from_vault:
            LOGGER.info("getting ddi admin credentials from vault")
            mm_admin_user, mm_admin_password =  get_mm_credentials_from_vault(
                vault_client, rotation_args)
        else:
            mm_admin_user = rotation_args.ddi_admin_username
            mm_admin_password = rotation_args.ddi_admin_password

        # get active mm user
        LOGGER.info("getting current ddi user credentials from path: %(path)s, "
                    "mount_point: %(mp)s", {'path': mm_user_secret_path,
                    'mp': mm_mount_point})
        mm_active_user = get_active_role(
            vault_client, mm_user_secret_path, mm_mount_point)
        mm_inactive_user = get_inactive_role(mm_active_user, user_type)
        # Getting list of current available users and passwords from vault.
        LOGGER.info("getting available ddi usernames from path: %(path)s, "
                    "mount point: %(mp)s", {'path': mm_users_secret_path,
                    'mp': mm_mount_point})
        auth_data = get_available_roles(vault_client, mm_users_secret_path,
                                        mm_mount_point)
        # create a new password for previous active ddi role
        LOGGER.info("generating new password for ddi user %(user)s. ",
                        {'user': mm_inactive_user})
        updated_auth_data, new_inactive_mm_user_password = update_role_password(
                        auth_data, mm_inactive_user, ddi=True)

        # update ddi inactive user password with the new password
        LOGGER.info("updating password of mmws inactive user %(user)s. "
                    "mmws url: %(url)s address: %(address)s", {'user': mm_inactive_user,
                    'url': rotation_args.ddi_url,
                    'address': rotation_args.ddi_server_address})
        update_ddi_user_password(rotation_args, mm_inactive_user, new_inactive_mm_user_password,
                                mm_admin_user, mm_admin_password)

        # update password of inactive ddi user in vault
        LOGGER.info("updating the password of ddi user %(user)s in vault."
                    "path: %(path)s, mount_point: %(mp)s.",
                    {'path': mm_users_secret_path, 'mp': mm_mount_point,
                    'user': mm_inactive_user})
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, mm_users_secret_path, mm_mount_point,
                            updated_auth_data, logger = LOGGER)
        # change username and password of applications to current inactive user(new user)
        # and password in vault
        app_auth_data = {'username': mm_inactive_user, 'password': new_inactive_mm_user_password}

        # update ddi user and password for baremetal enrollment
        LOGGER.info("updating ddi username and password for application "
                    "bm-enrollment in vault")
        vault_secret_utils.create_or_update_vault_secret(vault_client, mm_user_secret_path,
                                mm_mount_point,app_auth_data, logger = LOGGER)

        # update ddi user and password for vm instance operator
        if rotation_args.harvester_cluster_names:
            for harvester in rotation_args.harvester_cluster_names:
                secret_path = f"{rotation_args.az_name}-vm-instance-operator-{harvester}/mmws"
                LOGGER.info("updating ddi username and password for application "
                            "vm-instance-operator/%(name)s in vault",
                            {'name': harvester})
                vault_secret_utils.create_or_update_vault_secret(vault_client,
                        secret_path, mm_mount_point, app_auth_data, logger = LOGGER)

            # -rollout vm-instance-operator to apply the change
            # -baremetal enrollment tasks use vault go client to get vault secrets
            deployment_namespace = "idcs-system"
            for harvester in rotation_args.harvester_cluster_names:
                deployment_name = f"{rotation_args.az_name}-vm-instance-operator-{harvester}"
                rollout_deployment(v1_apps, deployment_name, deployment_namespace)
                wait_for_deployment_complete(v1_apps, deployment_name, deployment_namespace)
                verify_new_role_and_password_from_app_pod(vault_client, rotation_args,
                        mm_inactive_user, new_inactive_mm_user_password, deployment_name)

        # create a new password for previous active ddi role
        LOGGER.info("generating new password for ddi user %(user)s. ", {'user': mm_active_user})
        new_updated_auth_data, old_active_mm_user_password = update_role_password(
                updated_auth_data, mm_active_user, ddi= True)

        # update ddi previous active user password with the new password
        LOGGER.info("updating password of mmws old active user %(user)s. "
                    "mmws url: %(url)s address: %(address)s", {'user': mm_active_user,
                    'url': rotation_args.ddi_url,
                    'address': rotation_args.ddi_server_address})
        update_ddi_user_password(rotation_args, mm_active_user, old_active_mm_user_password,
                                mm_admin_user, mm_admin_password)

        # update password of inactive ddi user in vault
        LOGGER.info("updating the password of ddi user %(user)s in vault."
                    "path: %(path)s, mount_point: %(mp)s.",
                    {'path': mm_users_secret_path, 'mp': mm_mount_point,
                                'user': mm_active_user})
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, mm_users_secret_path, mm_mount_point,
            new_updated_auth_data, logger = LOGGER)

    except Exception as err:
        LOGGER.error("rotate_ddi_user failed with error: %(err)s  ", {'err': err})
        raise SecretRotatorError(err) from err

def monitor_db_activity_stats(vault_client, rotation_args, db_role, application):
    """monitor database activity stats for active connections from a db role.

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
        application (string): IDC application
        db_role (string): DB role to monitor
    """

    try:
        monitor_start = time.time()
        core_v1_api = client.CoreV1Api()
        # load_kube_config
        if application == "compute_api_server" or application == "netbox":
            if rotation_args.fetch_kubeconfigs_from_vault:
                cluster_kubeconfig = rotation_args.compute_cluster_id
            else:
                cluster_kubeconfig = rotation_args.compute_cluster_kubeconfig
            get_kube_config(vault_client, rotation_args, cluster_kubeconfig)

            #db_namespace = "default"
            db_namespace = "idcs-system"
            db_container = "postgresql"
            db_mount_point = "secret"
            if application == "compute_api_server":
                db_pool_label = 'app.kubernetes.io/instance=psql-compute, app.kubernetes.io/component=postgresql'
                #db_pool_label = 'app.kubernetes.io/instance=us-dev-1-compute-db, app.kubernetes.io/component=primary'
                db_customer_secret_path = "dbaas/psql-compute/customer"
                db_name = "main"
            else:
                db_pool_label = 'app.kubernetes.io/instance=psql-netbox, app.kubernetes.io/component=postgresql'
                db_customer_secret_path = "dbaas/psql-netbox/customer"
                db_name = "main"

            # get DB admin user from vault
            db_auth_data = get_available_roles(
                vault_client, db_customer_secret_path, db_mount_point)
            db_admin_user = db_auth_data['usernames'].split(',')[0]
            db_admin_password = db_auth_data['passwords'].split(',')[0]

        else:
            cluster_kubeconfig = rotation_args.eks_kubeconfig
            get_kube_config(vault_client, rotation_args, cluster_kubeconfig)

            db_namespace = "idcs-system"
            db_container = "postgres-client-secret-rotator"
            db_mount_point = "controlplane"
            db_pool_label = 'app=secret-rotator'
            db_customer_secret_path = f"{application}/customer"
            db_name = f"{application}"
            # ToDo: get db admin user and password from vault once it is
            # added to f"{application}/customer" path
            db_admin_user = rotation_args.rds_superuser
            db_admin_password = rotation_args.rds_superuser_password
            db_host = rotation_args.rds_host


        # get postgresql instance pod list
        pod_list = core_v1_api.list_namespaced_pod(db_namespace, label_selector=db_pool_label, watch=False)
        # check stat_activity_in_each_pods
        count = 0
        monitor_end = monitor_start + rotation_args.db_stat_monitor_duration
        while time.time() < monitor_end:
            for pod in pod_list.items:
                active_connections = True
                pod_name = pod.metadata.name
                if application == "compute_api_server" or application == "netbox":
                    monitor_cmd = (
                        f"PGHOST={pod_name}  PGPASSWORD={db_admin_password} "
                        f"psql -U {db_admin_user} -d {db_name} "
                        f"-c \"SELECT count(*) FROM pg_stat_activity WHERE usename='{db_role}' and state='active'\"")
                else:
                    monitor_cmd = (
                        f"PGHOST={db_host}  PGPASSWORD={db_admin_password} "
                        f"psql -U {db_admin_user} -d {db_name} "
                        f"-c \"SELECT count(*) FROM pg_stat_activity WHERE usename='{db_role}' and state='active'\"")
                LOGGER.info(monitor_cmd)
                monitor_cmd_exec_command = ['/bin/sh', '-c', monitor_cmd]
                monitor_cmd_output = stream(core_v1_api.connect_get_namespaced_pod_exec,
                            pod_name, db_namespace, command=monitor_cmd_exec_command, stderr=True,
                            stdin=False, stdout=True, tty=False, container=db_container)
                LOGGER.debug("monitor command output: %(op)s", {'op': monitor_cmd_output})
                active_connections = re.findall(r'\s+(\d)', monitor_cmd_output, re.MULTILINE)
                LOGGER.debug("active connections: %(ac)s", {'ac': active_connections})
                if (active_connections and len(active_connections) == 1
                    and active_connections[0] == '0'):
                    LOGGER.info("active connections for role %(role)s to DB %(db)s at pod "
                                "%(pod)s is 0", {'db': db_name, 'pod': pod_name, 'role': db_role})
                    active_connections = False
                elif (active_connections and len(active_connections) == 1
                      and active_connections[0] != '0'):
                    LOGGER.info("active connections for role %(role)s to DB %(db)s at pod "
                                "%(pod)s is %(conn)s", {'db': db_name, 'pod': pod_name,
                                'conns': active_connections[0], 'role': db_role})
                    active_connections = True
                else:
                    active_connections = False

            # reset count to 0 if there are active connections to DB
            if active_connections:
                count = 0
            else:
                count += 1

            # stop monitor if 5 consecutive stat check doesn't show any active connections
            if count == 5:
                LOGGER.info("stopping monitoring as the last 5 stat queries doesn't show any active "
                            "connections to DB %(db)s for role %(role)s", {'db': db_name, 'role': db_role})
                break
            time.sleep(20)

    except Exception as err:
        elapsed_time = time.time() - monitor_start
        sleep_time = rotation_args.db_stat_monitor_duration - elapsed_time
        logging.warning("failed to monitor DB %(app)s pg_stat_activity with %(err)s "
            "after role and password update. Password of the previous active role %(role)s "
            "will be updated in %(timeout)s seconds", {'app': application, 'role': db_role,
            'timeout': sleep_time, 'err': err})
        time.sleep(sleep_time)

def update_global_db_role_password(rotation_args, application, db_role, db_role_password):
    """update DB role password for global applications in postgresql.
       (billing, metering, cloud account).
       password is updated from postgresql client pod

    Args:
        rotation_args (argparse.Namespace): secret rotator input arguments
        application (string): global application
        db_role (string): DB role that requires password update
        db_role_password (string): DB role password

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:
        pod_namespace = "idcs-system"
        postgres_client_pod_name = "postgres-client-secret-rotator"
        postgres_client_container_name = "postgres-client"

        get_kube_config_from_file(rotation_args.eks_kubeconfig)
        if rotation_args.connect_to_k8s_cluster_via_proxy:
            update_k8s_client_with_proxy()
        core_v1_api = client.CoreV1Api()
        # Alter role with login
        LOGGER.info("enabling login for DB role %(role)s. database: %(db)s",
                    {'role': db_role, 'db': application})
        login_cmd = (
            f"PGHOST={rotation_args.rds_host} PGPASSWORD={rotation_args.rds_superuser_password} "
            f"psql -U {rotation_args.rds_superuser} -d {application} "
            f"-c 'ALTER ROLE {db_role} WITH LOGIN'")

        LOGGER.info(login_cmd)
        login_exec_command = ['/bin/sh', '-c', login_cmd]
        login_exec_command_output = stream(core_v1_api.connect_get_namespaced_pod_exec,
                                    postgres_client_pod_name, pod_namespace,
                                    command=login_exec_command, stderr=True, stdin=False,
                                    stdout=True, tty=False,
                                    container=postgres_client_container_name)
        if login_exec_command_output.strip() != "ALTER ROLE":
            err_msg = f"failed to enable login for role {db_role}. database: {application}"
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)
        # set role with new password
        LOGGER.info("setting up new password for DB role %(role)s. database: %(db)s",
                    {'role': db_role, 'db': application})
        login_pw_cmd = (
            f"PGHOST={rotation_args.rds_host}  PGPASSWORD={rotation_args.rds_superuser_password} "
            f"psql -U {rotation_args.rds_superuser} -d {application} "
            f"-c \"ALTER ROLE {db_role} WITH PASSWORD '{db_role_password}'\"")
        LOGGER.info(login_pw_cmd)
        login_pw_exec_command = ['/bin/sh', '-c', login_pw_cmd]
        login_pw_exec_command_output = stream(core_v1_api.connect_get_namespaced_pod_exec,
                                    postgres_client_pod_name, pod_namespace,
                                    command=login_pw_exec_command, stderr=True,
                                    stdin=False, stdout=True, tty=False,
                                    container=postgres_client_container_name)
        if login_pw_exec_command_output.strip() != "ALTER ROLE":
            err_msg = f"failed to setup password for role {db_role}. database: {application}"
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

    except Exception as err:
        LOGGER.error("setting up global DB role password failed. database: %(db)s, "
                     "error: %(err)s", {'db': application, 'err': err})
        raise SecretRotatorError(err) from err

def verify_dbaas_db_password(
    vault_client, rotation_args, db_role, db_role_password, dbaas_app):
    """verify compute DB password of role `db_role` is updated
    to match the change in vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (dict): secret rotator input arguments
        db_role (string): compute DB role with new password
        db_role_password (string): compute DB password to verify
        dbaas_app (string): DBaaS application

    Raises:
        SecretRotatorError: secret rotation error
    """
    # get kubeconfig from vault if `fetch_kubeconfigs_from_vault` is set to True,
    # else use kubeconfig file location
    if rotation_args.fetch_kubeconfigs_from_vault:
        compute_db_cluster_kubeconfig = rotation_args.compute_db_cluster_id
    else:
        compute_db_cluster_kubeconfig = rotation_args.compute_db_cluster_kubeconfig
    get_kube_config(vault_client, rotation_args, compute_db_cluster_kubeconfig)

    # Get DBaaS DB pgpool pod
    core_v1_api = client.CoreV1Api()
    dbaas_db_namespace = "default"
    dbaas_db_pgpool_label = f'app.kubernetes.io/component=pgpool, app.kubernetes.io/instance={dbaas_app}'
    pgpool_pod_name = get_pod_name(core_v1_api, dbaas_db_namespace, dbaas_db_pgpool_label)
    if not pgpool_pod_name:
        LOGGER.error("failed to get compute pgpool pod name from the regional DBaaS DB cluster")

    # get available compute DB users
    db_roles_command = [
        '/bin/sh', '-c', 'cat vault/secrets/customer | grep POSTGRES_CUSTOM_USERS | cut -d "=" -f 2']
    db_roles = stream(core_v1_api.connect_get_namespaced_pod_exec, pgpool_pod_name,
                      dbaas_db_namespace, command=db_roles_command, stderr=True,
                      stdin=False, stdout=True, tty=False, container="vault-agent")

    # get available compute DB passwords
    db_passwords_command = [
        '/bin/sh', '-c', 'cat vault/secrets/customer | grep POSTGRES_CUSTOM_PASSWORDS | cut -d "=" -f 2']
    db_passwords = stream(core_v1_api.connect_get_namespaced_pod_exec, pgpool_pod_name,
                          dbaas_db_namespace, command=db_passwords_command, stderr=True,
                          stdin=False, stdout=True, tty=False, container="vault-agent")
    if not db_roles:
        LOGGER.error("failed to get available DB users from the regional DB cluster pgpool pod")
    if not db_passwords:
        LOGGER.error("failed to get available DB passwords from the regional DB cluster pgpool pod")

    # check if role password matches expected password
    available_roles = db_roles.strip().split(',')
    available_passwords = db_passwords.strip().split(',')
    LOGGER.info("available roles in regional DBaaS DB cluster %(app)s "
                "pgpool pod: %(roles)s", {'roles': available_roles, 'app':dbaas_app})
    role_index = available_roles.index(db_role)
    role_password = available_passwords[role_index]

    if role_password == db_role_password:
        LOGGER.info("password of role %(role)s in regional DBaaS DB cluster "
                    " %(app)s pgpool pod matches expected password "
                    "after update", {'role': db_role, 'app': dbaas_app})
    else:
        error_message = ("password of role %(role)s in regional compute DBaaS DB "
                         "cluster %(app)s pgpool pod doesn't match expected password "
                         "after update", {'role': db_role, 'app': dbaas_app})
        LOGGER.error(error_message)
        raise SecretRotatorError(error_message)

def update_dbaas_db_role_password_in_vault_and_verify(
    vault_client, rotation_args, auth_data, db_role, db_role_password, application):
    """
    Update vault with new password and verify if the password is updated
    from regional DBaaS DB cluster

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (dict): secret rotator input arguments
        auth_data(dict): dictionary containing existing DB usernames and passwords
        db_role (string): compute DB role whose password is updated
        db_role_password (string): updated db_role password
        application (string): IDC application
    """
    if application == "compute_api_server":
        dbaas_app = "psql-compute"
        dbaas_db_customer_secret_path = "dbaas/psql-compute/customer"
    elif application == "netbox":
        dbaas_app = "psql-netbox"
        dbaas_db_customer_secret_path = "dbaas/psql-netbox/customer"

    dbaas_db_mount_point = "secret"

    LOGGER.info("updating the password of DB user: %(user)s at path: %(path)s, "
                "mount_point: %(mp)s and region: %(region)s", {
                'path': dbaas_db_customer_secret_path,
                'mp':dbaas_db_mount_point, 'region': rotation_args.region_name,
                'user': db_role})
    vault_secret_utils.create_or_update_vault_secret(
        vault_client, dbaas_db_customer_secret_path, dbaas_db_mount_point,
        auth_data, logger = LOGGER)

    # Waiting 140 seconds as the template render interval is set to 120 seconds
    # in compute DB pg pool pod
    LOGGER.info("waiting 140 seconds for the vault agent injector to update the "
                "DB role: %(user)s password in regional DbaaS DB cluster "
                "%(app)s pgpool pod", {'user': db_role, 'app': dbaas_app})
    time.sleep(140)

    # Verify the password is updated from compute DB PG pool pod
    verify_dbaas_db_password(
        vault_client, rotation_args, db_role, db_role_password, dbaas_app)

def verify_new_role_and_password_from_app_pod(
            vault_client, rotation_args, db_role, db_role_password, application):
    """verify if application username and password in app pod are updated
       to match the new role and password in vault

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (dict): secret rotator input arguments
        db_role (string): expected role from pod
        db_role_password (string): expected password from pod
        application (string): IDC application

    Raises:
        SecretRotatorError: secret rotation error
    """
    # regional kubeconfig to verify compute api server DB role changes
    if application == "compute_api_server":
        if rotation_args.fetch_kubeconfigs_from_vault:
            cluster_kubeconfig = rotation_args.compute_cluster_id
        else:
            cluster_kubeconfig = rotation_args.compute_cluster_kubeconfig
        get_kube_config(vault_client, rotation_args, cluster_kubeconfig)
    elif application == "vm_instance_operator":
        if rotation_args.fetch_kubeconfigs_from_vault:
            cluster_kubeconfig = rotation_args.az_cluster_id
        else:
            cluster_kubeconfig = rotation_args.az_cluster_kubeconfig
        get_kube_config(vault_client, rotation_args, cluster_kubeconfig)
    # EKS kubeconfig to verify DB role changes
    else:
        cluster_kubeconfig = rotation_args.eks_kubeconfig
        get_kube_config(vault_client, rotation_args, cluster_kubeconfig)

    app_namespace = "idcs-system"

    if application == "compute_api_server":
        app_label = 'app.kubernetes.io/name=compute-api-server'
    elif application == "billing":
        app_label = 'app.kubernetes.io/name=billing'
    elif application == "metering":
        app_label = 'app.kubernetes.io/name=metering'
    elif application == "cloudaccount":
        app_label = 'app.kubernetes.io/name=cloudaccount'
    elif 'vm-instance-operator' in application :
        app_label = f"app.kubernetes.io/instance: {application}"

    core_v1_api = client.CoreV1Api()

    # get app pod name
    app_pod_list = core_v1_api.list_namespaced_pod(
        app_namespace, label_selector=app_label, watch=False)

    for pod in app_pod_list.items:
        app_pod_name = pod.metadata.name

        # get username and password from app pod
        if application == 'vm_instance_operator':
            pod_db_username_command = ['/bin/sh', '-c', 'cat  vault/secrets/mmws_username']
            pod_db_password_command = ['/bin/sh', '-c', 'cat  vault/secrets/mmws_password']
        else:
            pod_db_username_command = ['/bin/sh', '-c', 'cat  vault/secrets/db_username']
            pod_db_password_command = ['/bin/sh', '-c', 'cat  vault/secrets/db_password']

        pod_db_username = stream(core_v1_api.connect_get_namespaced_pod_exec, app_pod_name,
                        app_namespace, command=pod_db_username_command, stderr=True,
                        stdin=False, stdout=True, tty=False, container="vault-agent")
        pod_db_password = stream(core_v1_api.connect_get_namespaced_pod_exec, app_pod_name,
                        app_namespace, command=pod_db_password_command, stderr=True,
                        stdin=False, stdout=True, tty=False, container="vault-agent")
        if not pod_db_username:
            LOGGER.error("failed to get username from %(app)s pod", {'app': application})
        if not pod_db_password:
            LOGGER.error("failed to get password from %(app)s pod", {'app': application})

        # check if pod DB role matches expected role
        pod_db_username = pod_db_username.strip()
        if pod_db_username == db_role:
            LOGGER.info("role name in %(app)s pod matches expected role name"
                        "after update. role name: %(role)s",
                        {'role': pod_db_username, 'app': application})
        else:
            err_msg = (f" {application} role name doesn't match expected role name "
                    f"after update. current role name: { pod_db_username}, "
                    f"expected role name: {db_role}")
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

        # check if pod DB role password matches expected password
        pod_db_password = pod_db_password.strip()
        if pod_db_password == db_role_password:
            LOGGER.info("role password in %(app)s pod matches expected role password"
                        "after update. role name: %(role)s",
                        {'role': pod_db_username, 'app': application})
        else:
            err_msg = (f" {application} role password doesn't match expected password  "
                    f"after update. role name: {pod_db_username}")
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

def rotate_db_role(vault_client, rotation_args, application):
    """rotate global app DB role

    # steps
    # get_inactive_db_user from vault(application vault path)
    # update_password_of_inactive_db_user from postgres-client pod(aws)/dbaas pgpool secret path
    # verify_if_inactive_db_password_is_updated
    # update_current_db_username_to_inactive_db_user from the steps above
    # verify_if_db_user is updated from application(billing,metering,etc)
    # update_password_of_previous_active_user
    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
        application: idcs global application

    Raises:
        SecretRotatorError: secret rotation error
    """
    try:

        if application == "compute_api_server":
            db_role_secret_path = f"{rotation_args.region_name}-compute-api-server/database"
            db_roles_secret_path = "dbaas/psql-compute/customer"
            db_role_mount_point = "controlplane"
            db_roles_mount_point = "secret"
        elif application == "netbox":
            #ToDo:
            # Assuming netbox k8s secrets will be deployed using external secrets operator
            # generate new passwords for netbox k8s secrets
            # update dbaas psql-netbox path in vault with new password
            # verify that the password is updated from psql netbox pgpool pod
            # update netbox secrets in vault
            # restart netbox deployment
            # verify new credentials
            db_role_secret_path = f"{rotation_args.region_name}/netbox/database"
            db_roles_secret_path = "dbaas/psql-netbox/customer"
            db_role_mount_point = "controlplane"
            db_roles_mount_point = "secret"
        elif application == 'billing' or application == 'metering' or application == 'cloudaccount':
            #db_role_secret_path = f"{application}/aws-database"
            # Note: To test in the kind dev cluster
            db_role_secret_path = f"{application}/database"
            db_roles_secret_path = f"{application}/customer"
            db_role_mount_point = "controlplane"
            db_roles_mount_point = "controlplane"
        else:
            err_msg = f"DB secret rotation is not available for {application}"
            LOGGER.error(err_msg)
            raise SecretRotatorError(err_msg)

        db_user_type = application
        # get active DB user
        LOGGER.info("getting current active DB role from path: %(path)s, mount_point: %(mp)s",
                    {'path': db_role_secret_path, 'mp': db_role_mount_point})
        active_db_role = get_active_role(
            vault_client, db_role_secret_path, db_role_mount_point)
        # get inactive DB user
        inactive_db_role = get_inactive_role(active_db_role, db_user_type)

        if application == 'billing' or application == 'metering' or application == 'cloudaccount':
            # create postgres-client pod
            create_postgres_client_pod(rotation_args)

        # get list of current available roles and passwords
        LOGGER.info("getting available DB users from path: %(path)s, mount point: %(mp)s",
                    {'path': db_roles_secret_path, 'mp': db_roles_mount_point})

        current_db_auth_info = get_available_roles(
            vault_client, db_roles_secret_path, db_roles_mount_point)

        # create a new password for inactive DB role
        updated_db_auth_info = update_role_password(current_db_auth_info, inactive_db_role)
        new_inactive_db_role_auth_data = updated_db_auth_info[0]
        new_inactive_db_role_password = updated_db_auth_info[1]

        if application == "compute_api_server":
            update_dbaas_db_role_password_in_vault_and_verify(vault_client, rotation_args,
            new_inactive_db_role_auth_data, inactive_db_role, new_inactive_db_role_password,
            application)
        else:
            # update password of inactive postgresql DB role
            update_global_db_role_password(
                rotation_args, application, inactive_db_role, new_inactive_db_role_password)

            # update password of inactive DB role in vault
            LOGGER.info("updating the password of DB role: %(user)s at path: %(path)s and "
                    "mount_point: %(mp)s.", {'path': db_roles_secret_path,
                    'mp': db_roles_mount_point, 'user': inactive_db_role})
            vault_secret_utils.create_or_update_vault_secret(
                vault_client, db_roles_secret_path, db_roles_mount_point,
                new_inactive_db_role_auth_data, logger = LOGGER)

        # update DB username and password for the application path in vault to
        # inactive role and updated password
        updated_app_db_auth_info = {'username': inactive_db_role,
                                 'password': new_inactive_db_role_password}
        LOGGER.info("updating the DB username and password of %(app)s at path: %(path)s "
                    "and mount point: %(mp)s. current role: %(current_user)s, new role: "
                    "%(new_user)s",{'path': db_role_secret_path, 'mp': db_role_mount_point,
                    'new_user': inactive_db_role, 'current_user': active_db_role,
                    'app': application})
        vault_secret_utils.create_or_update_vault_secret(
            vault_client, db_role_secret_path, db_role_mount_point,
            updated_app_db_auth_info, logger = LOGGER)

        # Waiting 140 seconds as the template render interval is set to 120 seconds
        LOGGER.info("waiting 140 seconds for the vault agent injector to update the "
                    "DB username to %(user)s and password in %(app)s pod",
                    {'user': inactive_db_role, 'app': application})
        time.sleep(140)
        # verify if DB username and password are updated in app pod
        verify_new_role_and_password_from_app_pod(vault_client, rotation_args, inactive_db_role,
                            new_inactive_db_role_password, application)
        # monitor DB pg_stat_activity_to check the connection status of previous active role
        monitor_db_activity_stats(vault_client, rotation_args, active_db_role, application)
        # create a new password for previous active DB role
        updated_db_auth_info = update_role_password(
            new_inactive_db_role_auth_data, active_db_role)
        previous_active_db_role_auth_data = updated_db_auth_info[0]
        previous_active_db_role_password = updated_db_auth_info[1]

        # update password of previous active postgresql DB role
        if application == "compute_api_server":
            update_dbaas_db_role_password_in_vault_and_verify(vault_client, rotation_args,
            new_inactive_db_role_auth_data, active_db_role, new_inactive_db_role_password,
            application)
        else:
            update_global_db_role_password(rotation_args, application, active_db_role,
                previous_active_db_role_password)

            # update password of previous DB role in vault
            LOGGER.info("updating the password of DB role: %(user)s at path: %(path)s and "
                    "mount_point: %(mp)s.", {'path': db_roles_secret_path,
                    'mp': db_roles_mount_point, 'user': active_db_role})
            vault_secret_utils.create_or_update_vault_secret(
                vault_client, db_roles_secret_path, db_roles_mount_point,
                previous_active_db_role_auth_data, logger = LOGGER)

        LOGGER.info("rotated %(app)s DB role and password", {'app': application})

    except Exception as err:
        LOGGER.error("rotate_%(app)s_db_role failed with error: %(err)s",
            {'app': application, 'err': err})
        raise SecretRotatorError(err) from err

def rotate_billing_db_role(vault_client, rotation_args):
    """rotate billing DB role
    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    """
    rotate_db_role(vault_client, rotation_args, "billing")

def rotate_metering_db_role(vault_client, rotation_args):
    """rotate metering DB role
    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    """
    rotate_db_role(vault_client, rotation_args, "metering")

def rotate_cloudaccount_db_role(vault_client, rotation_args):
    """rotate cloud account DB role
    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    """
    rotate_db_role(vault_client, rotation_args, "cloudaccount")

def rotate_netbox_secrets(vault_client, rotation_args):
    """rotate compute DB user used by compute api server

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error
    """
    rotate_db_role(vault_client, rotation_args, "netbox")

def rotate_compute_api_db_role(vault_client, rotation_args):
    """rotate compute DB user used by compute api server

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments

    Raises:
        SecretRotatorError: secret rotation error
    """
    rotate_db_role(vault_client, rotation_args, "compute_api_server")


def rotate_regional_secrets(vault_client, rotation_args):
    """rotate regional secrets

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    LOGGER.debug(rotation_args)
    rotate_compute_api_db_role(vault_client, rotation_args)
    # rotate_netbox_secrets(vault_client, rotation_args)
    rotate_netbox_token(vault_client, rotation_args)

def rotate_global_secrets(vault_client, rotation_args):
    """rotate global secrets

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    LOGGER.debug(rotation_args)
    rotate_billing_db_role(vault_client, rotation_args)
    rotate_metering_db_role(vault_client, rotation_args)
    rotate_cloudaccount_db_role(vault_client, rotation_args)

def rotate_az_secrets(vault_client, rotation_args):
    """rotate AZ secrets

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    LOGGER.debug(rotation_args)
    rotate_ssh_proxy_operator_keys(vault_client, rotation_args)
    rotate_bm_ssh_proxy_operator_keys(vault_client, rotation_args)

def rotate_ssh_proxy_operator_keys(vault_client, rotation_args):
    """rotate ssh proxy operator secrets

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    rotate_ssh_keys(vault_client, rotation_args, "ssh_proxy_operator")

def rotate_bm_ssh_proxy_operator_keys(vault_client, rotation_args):
    """rotate ssh proxy operator secrets

    Args:
        vault_client (hvac.Client): vault client instance
        rotation_args (argparse.Namespace): secret rotator input arguments
    """
    rotate_ssh_keys(vault_client, rotation_args, "bm_ssh_proxy_operator")

def main():
    """
    Rotate IDC secrets
    """
    parser = argparse.ArgumentParser(description="secret rotator")
    parser.add_argument("--vault-addr",
        help="Vault server address. Get it from the ENV var VAULT_ADDR if this flag is not provided")
    parser.add_argument("--vault-token",
        help="Vault Token. Get it from the ENV var VAULT_TOKEN if this flag is not provided")
    parser.add_argument("--log-level", type=int, default=logging.INFO, help="10=DEBUG,20=INFO")
    parser.add_argument("--db-stat-monitor-duration", type=int, default=600,
                        help="DB stat monitor duration")
    parser.add_argument("--connect-to-k8s-cluster-via-proxy", action=argparse.BooleanOptionalAction,
                        default=True, help="rotate all regional secrets")
    parser.add_argument("--fetch-kubeconfigs-from-vault", action=argparse.BooleanOptionalAction,
                        default=True, help="fetch cluster kubeconfig from vault")
    subparser = parser.add_subparsers()
    ## region specific parameters
    region = subparser.add_parser("region", help= "regional secret rotation configuration")
    region.add_argument("--region-name", default='us-dev-1',help="name of IDC region", required=True)
    region.add_argument("--rotate-all-regional-secrets", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate all regional secrets")
    region.add_argument("--rotate-compute-api-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate compute api DB role and password")
    region.add_argument("--rotate-netbox-k8s-secrets", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate netbox k8s secrets")

    # provide location of the regional kubeconfig files
    region.add_argument(
        "--compute-cluster-kubeconfig", default='', help="IDC regional cluster kubeconfig")
    region.add_argument(
        "--compute-db-cluster-kubeconfig", default='', help="IDC regional DB cluster kubeconfig")
    region.add_argument(
        "--compute-cluster-id", default='rgcp',help=("compute cluster identifier in vault. "
        "This option is required only if fetch-kubeconfigs-from-vault is set to true"))
    region.add_argument(
        "--compute-db-cluster-id", default='rgdb',help=("compute DB cluster identifier in vault. "
        "This option is required only if fetch-kubeconfigs-from-vault is set to true"))
    # netbox arguments
    region.add_argument("--rotate-netbox-token", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate netbox secrets")
    region.add_argument("--fetch-netbox-credentials-from-vault", action=argparse.BooleanOptionalAction,
                        default=True, help="fetch netbox credentials from vault")
    region.add_argument("--netbox-address",  default='', help=("netbox hostname or ip address. "
                        "Get it from the ENV var NETBOX_ADDR if this flag is not provided"))
    region.add_argument("--netbox-username",  default='', help=("netbox username to change the token. "
                        "Get it from the ENV var NETBOX_USER if this flag is not provided. "
                        "This option is required only if fetch-netbox-credentials-from-vault is set to false"))
    region.add_argument("--netbox-password",  default='', help=("netbox password to change the token. "
                        "Get it from the ENV var NETBOX_PASSWORD if this flag is not provided. "
                        "This option is required only if fetch-netbox-credentials-from-vault is set to false"))
    ## global specific parameters
    global_config = subparser.add_parser("global", help= "global secret rotation configuration")
    global_config.add_argument(
        "--eks-kubeconfig", default='', help="global eks cluster kubeconfig")
    global_config.add_argument(
        "--rds-superuser", default='', help="global RDS super user to manage the DB roles. "
                        "Get it from the ENV var RDS_SUPERUSER if this flag is not provided")
    global_config.add_argument(
        "--rds-superuser-password", default='', help="global RDS super user password to manage the DB roles. "
                        "Get it from the ENV var RDS_SUPERUSER_PASSWORD if this flag is not provided")
    global_config.add_argument(
        "--rds-host", default='', help="global RDS host FQDN or IP. "
                        "Get it from the ENV var RDS_HOST if this flag is not provided")
    global_config.add_argument("--rotate-all-global-secrets", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate all global secrets")
    global_config.add_argument("--rotate-authz-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate authz DB role and password")
    global_config.add_argument("--rotate-billing-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate billing DB role and password")
    global_config.add_argument("--rotate-notification-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate notification DB role and password")
    global_config.add_argument("--rotate-metering-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate metering DB role and password")
    global_config.add_argument("--rotate-cloudaccount-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate cloudaccount DB role and password")
    global_config.add_argument("--rotate-productcatalog-db-role", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate productcatalog DB role and password")

    ## region specific parameters
    az = subparser.add_parser("availability-zone", help= "AZ secret rotation configuration")
    az.add_argument("--az-name", default='us-dev-1a',help="name of the IDC availability zone", required=True)
    az.add_argument("--region-name", default='us-dev-1',help="name of the IDC region", required=True)
    # provide location of the az kubeconfig files
    az.add_argument(
        "--az-cluster-kubeconfig", default='', help="IDC AZ cluster kubeconfig")
    az.add_argument(
        "--az-cluster-id", default='azcp',help=("AZ compute cluster identifier in vault. "
        "This option is required only if fetch-kubeconfigs-from-vault is set to true"))
    # AZ rotate secrets
    az.add_argument("--rotate-all-az-secrets", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate all az secrets")
    az.add_argument("--rotate-ssh-proxy-operator-keys", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate SSH proxy operator SSH key pair")
    az.add_argument("--ssh-proxy-username", default='guest', help="SSH proxy username")
    az.add_argument("--ssh-proxy-addresses", nargs='*',
                        help="list of SSH proxy addresses")
    az.add_argument("--rotate-bm-ssh-proxy-operator-keys", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate baremetal operator SSH key pair")
    az.add_argument("--bm-ssh-proxy-username", default='bmo', help="BM ssh proxy username")
    az.add_argument("--bm-ssh-proxy-addresses", nargs='*',
                    help="list of BM SSH proxy addresses")
    # ddi arguments
    az.add_argument("--rotate-ddi-user", action=argparse.BooleanOptionalAction,
                        default=False, help="rotate ddi secrets")
    az.add_argument("--fetch-ddi-admin-credentials-from-vault", action=argparse.BooleanOptionalAction,
                        default=True, help="fetch ddi admin credentials from vault")
    az.add_argument("--ddi-url",  default='', help=("ddi URL fqdn or ip address. "
                        "Get it from the ENV var DDI_URL if this flag is not provided"))
    az.add_argument("--ddi-server-address",  default='', help=("ddi server fqdn or ip address. "
                        "Get it from the ENV var DDI_SERVER_ADDR if this flag is not provided"))
    az.add_argument("--ddi-admin-username",  default='',
                        help=("ddi username to change the credentials. "
                        "Get it from the ENV var DDI_ADMIN_USER if this flag is not provided. "
                        "This option is required only if fetch-ddi-admin-credentials-from-vault is set to false"))
    az.add_argument("--ddi-admin-password",  default='',
                        help=("ddi password to change the token. "
                        "Get it from the ENV var DDI_ADMIN_PASSWORD if this flag is not provided. "
                        "This option is required only if fetch-ddi-admin-credentials-from-vault is set to false"))
    az.add_argument("--harvester-cluster-names", nargs='*', help="list of harvester cluster names")

    if len(sys.argv)==1:
        parser.print_help(sys.stderr)
        sys.exit(1)
    args = parser.parse_args()

    # Get vault address and token
    if not args.vault_addr:
        args.vault_addr = vault_secret_utils.get_environment_var('VAULT_ADDR')
    if not args.vault_token:
        args.vault_token = vault_secret_utils.get_environment_var('VAULT_TOKEN')

    # required parameters when performing global secret rotation
    if hasattr(args, 'rds_superuser') and not args.rds_superuser:
        args.rds_superuser = vault_secret_utils.get_environment_var('RDS_SUPERUSER')
    if hasattr(args, 'rds_superuser_password') and not args.rds_superuser_password:
        args.rds_superuser_password= vault_secret_utils.get_environment_var(
            'RDS_SUPERUSER_PASSWORD')
    if hasattr(args, 'rds_host') and not args.rds_host:
        args.rds_host = vault_secret_utils.get_environment_var('RDS_HOST')
    if hasattr(args, 'netbox_address') and not args.netbox_address:
        args.netbox_address = vault_secret_utils.get_environment_var('NETBOX_ADDR')

    # set up logger
    logger = vault_secret_utils.setup_logger(args.log_level)
    globals()["LOGGER"] = logger
    LOGGER.debug("args: %s", str(args))

    if not args.vault_addr:
        LOGGER.error("Failed to get the vault server address. use environmental variable "
                     "VAULT_ADDR or argument --vault-addr to set vault address")
        return 1
    if not args.vault_token:
        LOGGER.error("Failed to get the vault token address. use environmental variable "
                     "VAULT_TOKEN or argument --vault-token to set vault token")
        return 1

    vault_client = vault_secret_utils.get_vault_client_instance(
        args.vault_addr, args.vault_token, logger=LOGGER)
    LOGGER.info("Vault initialize status: %s", vault_client.sys.is_initialized())

    # rotate regional secrets
    if hasattr(args, 'rotate_all_regional_secrets') and args.rotate_all_regional_secrets:
        netbox_token_args_check(args)
        rotate_regional_secrets(vault_client, args)
        sys.exit(0)
    if hasattr(args, 'rotate_compute_api_db_role') and args.rotate_compute_api_db_role:
        rotate_compute_api_db_role(vault_client, args)
    if hasattr(args, 'rotate_netbox_token') and args.rotate_netbox_token:
        netbox_token_args_check(args)
        rotate_netbox_token(vault_client, args)

    # rotate global secrets
    if hasattr(args, 'rotate_all_global_secrets') and args.rotate_all_global_secrets:
        global_config_args_check(args)
        rotate_global_secrets(vault_client, args)
        sys.exit(0)
    if hasattr(args, 'rotate_billing_db_role') and args.rotate_billing_db_role:
        global_config_args_check(args)
        rotate_billing_db_role(vault_client, args)
    if hasattr(args, 'rotate_metering_db_role') and args.rotate_metering_db_role:
        global_config_args_check(args)
        rotate_metering_db_role(vault_client, args)
    if hasattr(args, 'rotate_cloudaccount_db_role') and args.rotate_cloudaccount_db_role:
        global_config_args_check(args)
        rotate_cloudaccount_db_role(vault_client, args)

    # rotate AZ secrets
    if hasattr(args, 'rotate_all_az_secrets') and args.rotate_all_az_secrets:
        ssh_operator_args_check(args)
        ddi_args_check(args)
        rotate_az_secrets(vault_client, args)
        sys.exit(0)
    if hasattr(args, 'rotate_ssh_proxy_operator_keys') and args.rotate_ssh_proxy_operator_keys:
        ssh_operator_args_check(args)
        rotate_ssh_proxy_operator_keys(vault_client, args)
    if hasattr(args, 'rotate_bm_ssh_proxy_operator_keys') and args.rotate_bm_ssh_proxy_operator_keys:
        ssh_operator_args_check(args)
        rotate_bm_ssh_proxy_operator_keys(vault_client, args)
    if hasattr(args, 'rotate_ddi_user') and args.rotate_ddi_user:
        ddi_args_check(args)
        rotate_ddi_user(vault_client, args)

if __name__ == "__main__":
    sys.exit(main())
