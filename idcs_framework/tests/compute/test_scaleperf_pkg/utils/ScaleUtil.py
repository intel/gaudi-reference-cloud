
import json
import string
from jsonpath_ng import parse
from datetime import datetime, timedelta
import subprocess
import urllib3
import random
import time
import logging

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def read_json_file_as_string(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
        file.close()
        json_string = json.dumps(data)
    return json_string

def get_instance_name_and_payload(raw_payload, length):
    characters = string.ascii_lowercase + string.digits
    random_string = ''.join(random.choice(characters) for _ in range(length))
    random_string = "load-vm-"+random_string
    iteration_payload_string = raw_payload.replace("load-vm-<<index>>", random_string)
    return random_string, iteration_payload_string

def validate_status_code(response, expected_status_code):
    assert response.status_code == expected_status_code, f"Unexpected status code. Expected: {expected_status_code}, Actual: {response.status_code}"

def validate_response_body(response, expected_string):
    assert expected_string in response.text, f"Expected string '{expected_string}' not found in response body."

def is_instance_ready(http_client, creation_time, instance_endpoint, token, instance_name, proxies):
    instance_phase = ""
    while datetime.now() <= (creation_time + timedelta(minutes=5)):
        time.sleep(5)
        response = http_client.get_method(instance_endpoint, token, instance_name, proxies)
        validate_status_code(response, 200)
        instance_phase = (parse("$.status.phase").find(response.json())[0]).value
        if instance_phase == "Ready":
            break
        else:
            continue
    return instance_phase

def get_instance_details(response):
    proxy_ip = (parse("$.status.sshProxy.proxyAddress").find(response.json())[0]).value
    proxy_user = (parse("$.status.sshProxy.proxyUser").find(response.json())[0]).value
    instance_ip = (parse("$.status.interfaces[0].addresses[0]").find(response.json())[0]).value
    instance_user = (parse("$.status.userName").find(response.json())[0]).value
    return proxy_ip, proxy_user, instance_ip, instance_user

def is_instance_deleted(http_client, deletion_time, instance_endpoint, token, instance_name, proxies):
    isInstanceDeleted = False
    while datetime.now() <= (deletion_time + timedelta(minutes=5)):
        time.sleep(5)
        response = http_client.get_method(instance_endpoint, token, instance_name, proxies)
        if response.status_code == 404:
            isInstanceDeleted = True
            break
        else:
            continue
    return isInstanceDeleted


def run_ssh_commands(ssh_command):
    try:
        result = subprocess.run(ssh_command, capture_output=False, text=False)
        return result
    except subprocess.CalledProcessError as e:
        print("Error executing SSH command:", e.stderr)
        return None

def assert_ssh_output(result):
    assert result.returncode == 0

def read_env_file(filename):
    key_value_pairs = {}
    with open(filename, 'r') as file:
        for line in file:
            line = line.strip()
            if line and not line.startswith("#"):  # Skip empty lines and comments
                if '=' in line:
                        key, value = line.split('=', maxsplit=1)
                        key_value_pairs[key.strip()] = value.strip()
    return key_value_pairs