#!/usr/bin/env python3
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

import argparse
import base64
import json
import os
import shutil
import ssl
import subprocess
import time
import yaml
import urllib.request
from datetime import datetime
from jinja2 import Template
from pathlib import Path

def _log_message(msg):
    print(str(datetime.now()) + " " + msg)

parser = argparse.ArgumentParser(
    description='Bootstraps an instance into a KubeVirt cluster'
)

parser.add_argument('--cluster-name', help='The name of the cluster in Rancher')
parser.add_argument('--vault-addr', help='The address of Vault')
parser.add_argument('--vault-role', help='The Vault role')
parser.add_argument('--vault-token', help='The wrapped single-use Vault token')
parser.add_argument('--rancher-credentials-path', help='The path to Rancher credentials in Vault')

args = parser.parse_args()

_rancher_url = ''
_rancher_access_key = ''
_rancher_secret_key = ''
_vault_token = args.vault_token

def _backup_dir(src_dir, dst_dir):
    dst_dir = os.path.join(dst_dir, src_dir[1:] if src_dir.startswith('/') else src_dir)
    shutil.move(src_dir, dst_dir)
    os.mkdir(src_dir)

def network_configure(**kwargs):
    existing_netplan_file = '/etc/netplan/50-cloud-init.yaml'
    new_netplan_file = '/etc/netplan/100-idc-kubevirt-worker-node.yaml'
    hooks_dir = '/etc/networkd-dispatcher/configured.d/'
    backup_dir = '/opt/idc/kubevirt-worker-node/old/'
    br_name = 'mgmt-br'

    # Get network info from existing netplan
    eth_info = {}
    with open(existing_netplan_file) as f:
        data = yaml.safe_load(f)
        eth_name = list(data['network']['ethernets'].keys())[0]
        eth_info = data['network']['ethernets'][eth_name]
        eth_info['interfaces'] = [eth_name]
        eth_info['version'] = data['network']['version']

    # Backup existing netplan directory
    _backup_dir(os.path.dirname(existing_netplan_file), backup_dir)

    # Create new netplan with management bridge
    new_netplan = {
        'network': {
            'version': eth_info['version'],
            'ethernets': {
                eth_name: {
                    'match': eth_info['match'],
                    'mtu': eth_info['mtu'],
                    'set-name': eth_info['set-name'],
                }
            },
            'bridges': {
                br_name: {
                    'interfaces': eth_info['interfaces'],
                    'mtu': eth_info['mtu'],
                    'addresses': eth_info['addresses'],
                    'nameservers': eth_info['nameservers'],
                    'routes': eth_info['routes'],
                }
            },
        }
    }
    with open(new_netplan_file, 'w') as f:
        yaml.dump(new_netplan, f, sort_keys=False)
        os.chmod(new_netplan_file, 0o600)

    # Install networkd-dispatcher hooks to configure bridge and attached interface
    fileTemplates = {}
    fileTemplates[os.path.join(hooks_dir, eth_info['set-name'])] = """#!/bin/bash
[[ "${IFACE}" == "{{ ETH_NAME }}" ]] || exit 0
ip link set dev {{ BR_NAME }} address $(ip link show dev ${IFACE} | grep link/ether | awk "{print $2}")
bridge vlan add vid 2-4094 dev ${IFACE}

"""
    fileTemplates[os.path.join(hooks_dir, br_name)] = """#!/bin/bash
[[ "${IFACE}" == "{{ BR_NAME }}" ]] || exit 0
ip link set ${IFACE} type bridge vlan_filtering 1
bridge vlan add vid 2-4094 dev ${IFACE} self

"""
    Path(hooks_dir).mkdir(parents=True, exist_ok=True)
    for file in fileTemplates:
        tm = Template(fileTemplates[file])
        file_content = tm.render(ETH_NAME=eth_info['set-name'], BR_NAME=br_name)
        with open(file, 'w') as f:
            f.write(file_content)
        os.chmod(file, 0o755)

    subprocess.run(['netplan', 'generate'])
    subprocess.run(['netplan', 'apply'])

def host_prepare_for_k8s():
    _log_message('Preparing host')
    file = '/etc/sysctl.d/60-kubevirt-worker-node.conf'
    content = """net.ipv4.ip_forward = 1
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1
fs.inotify.max_user_instances = 1024
vm.max_map_count = 262144
"""
    with open(file, 'w') as f:
        f.write(content)
    subprocess.run(['sysctl', '-p', file])
    subprocess.run(['systemctl', 'stop', 'ufw'])
    subprocess.run(['systemctl', 'disable', 'ufw'])

def vault_login():
    _log_message('Logging into Vault')
    global _vault_token
    request = urllib.request.Request(args.vault_addr+'v1/sys/wrapping/unwrap', method='POST')
    request.add_header('X-Vault-Token', _vault_token)
    response = json.load(urllib.request.urlopen(request))
    secret_id = response['data']['secret_id']
    request = urllib.request.Request(args.vault_addr+'v1/auth/approle/login', method='POST')
    request.add_header('Content-Type', 'application/json')
    response = json.load(urllib.request.urlopen(request, data=json.dumps({'role_id':args.vault_role,'secret_id':secret_id}).encode('utf-8')))
    _vault_token = response['auth']['client_token']

def vault_get_credentials():
    _log_message('Getting credentials from Vault')
    global _rancher_url, _rancher_access_key, _rancher_secret_key
    request = urllib.request.Request(args.vault_addr+f'v1/controlplane/data/{args.rancher_credentials_path}')
    request.add_header('X-Vault-Token', _vault_token)
    response = json.load(urllib.request.urlopen(request))
    _rancher_url = response['data']['data']['url']
    _rancher_access_key = response['data']['data']['access_key']
    _rancher_secret_key = response['data']['data']['secret_key']

def _rancher(path):
    # TODO enable cert validation
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE

    request = urllib.request.Request(_rancher_url+path)
    auth = base64.b64encode(f"{_rancher_access_key}:{_rancher_secret_key}".encode("ascii")).decode("ascii")
    request.add_header('Authorization', f'Basic {auth}')
    return json.load(urllib.request.urlopen(request, context=ctx))

def _registration_command():
    response = _rancher(f'/v3/cluster?name={args.cluster_name}')
    cluster_id = response['data'][0]['id']
    response = _rancher(f'/v3/clusterregistrationtoken?clusterId={cluster_id}')
    return response['data'][0]['insecureNodeCommand'] # TODO nodeCommand

def _rancher_get_node_registration_command():
    _log_message('Getting node registration command')
    command = _registration_command()
    while command == 'null':
        _log_message('Trying again to get the node registration command')
        time.sleep(2)
        command = _registration_command()
    return command

def rancher_register_node():
    _log_message(f'Registering worker node with cluster {args.cluster_name}')
    subprocess.run(['/bin/sh', '-c', _rancher_get_node_registration_command()+' --worker'])

_log_message('Starting the KubeVirt cluster bootstrap script')

network_configure()
host_prepare_for_k8s()
vault_login()
vault_get_credentials()
rancher_register_node()
