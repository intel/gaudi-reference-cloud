import sys
import os

parent_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, parent_dir)

from locust import HttpUser, task, between, events
from utils.HttpClient import HttpClient
from datetime import datetime, timedelta
from jsonpath_ng import parse
from utils import ScaleUtil
import logging
import os
import json
import time
import jq
import urllib3
import time

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class User(HttpUser):
    wait_time = between(1, 4)  # Time between consecutive requests
    
    # load the env variables
    key_value_dict = ScaleUtil.read_env_file("resources/env.txt")
    host = key_value_dict["host"]
    cloudaccount=key_value_dict["cloudaccount"]
    instance_endpoint = "/v1/cloudaccounts/"+cloudaccount+"/instances"
    http_client = HttpClient(base_url=host)
    token = payload_string =  ""
    script_start_time = time.time()

    proxies = {
        "https": "http://internal-placeholder.com:912",
    }

    def on_start(self):
        response = self.http_client.get_token()
        ScaleUtil.validate_status_code(response, 200)
        self.token = response.text
        self.raw_payload_string = ScaleUtil.read_json_file_as_string("resources/gaudi2-vm-with-realistic-load.json")
        start_time = int(time.time() * 1000)
        logging.info("start-time : %s" % start_time)
        
    @task
    def make_request(self):
        instance_phase = proxy_ip = proxy_user = instance_ip = instance_user = ""
        try:
            self.stop_if_time_exceeded()
            instance_name, iteration_payload_string = ScaleUtil.get_instance_name_and_payload(self.raw_payload_string, 6)

            # creation of the instance
            creation_time = datetime.now()
            instance_creation = self.http_client.post_method(self.instance_endpoint,self.token, iteration_payload_string, instance_name, self.proxies)
            ScaleUtil.validate_status_code(instance_creation, 200)
            logging.info("Instance creation details - response code: %s response body: %s" % (instance_creation.status_code, instance_creation.text))
            
            # wait for the instance to be in ready state
            instance_phase = ScaleUtil.is_instance_ready(self.http_client, creation_time, self.instance_endpoint, self.token, instance_name, self.proxies)
            
            # validate the instance and populate the data for SSH
            if instance_phase == "Ready":
                logging.info("Instance phase : %s" % instance_phase)
                response = self.http_client.get_method(self.instance_endpoint, self.token, instance_name, self.proxies)
                ScaleUtil.validate_status_code(response, 200)
                proxy_ip, proxy_user, instance_ip, instance_user = ScaleUtil.get_instance_details(response)
            else:
                raise Exception("Instance should be in ready state")

            self.stop_if_time_exceeded()
            
            #ssh into the instance and run commands
            ssh_command = ['ssh', '-J', f'{proxy_user}@{proxy_ip}', '-o', 'StrictHostKeyChecking=no', '-o', 'UserKnownHostsFile=/dev/null', f'{instance_user}@{instance_ip}', 'cd', '/', '&&', 'sudo', 'sh', '/tmp/run_container.sh', '>', '/tmp/loadoutput.txt']
            result = ScaleUtil.run_ssh_commands(ssh_command)
            if result.returncode == 0:
                print("Script executed successfully with exit code", result.returncode)
            else:
                logging.info("Script executed successfully with exit code %s" % result.returncode)
                raise Exception("ssh command exited with non zero code", result.stderr)
            time.sleep(60)

            # delete the instance
            deletion_time = datetime.now()
            instance_deletion = self.http_client.delete_method(self.instance_endpoint,self.token, instance_name, None)
            ScaleUtil.validate_status_code(instance_deletion, 200)

            # wait for the instance to be deleted
            isInstanceDeleted = ScaleUtil.is_instance_deleted(self.http_client, deletion_time, self.instance_endpoint, self.token, instance_name, None)
        
            if isInstanceDeleted == False:
                raise Exception("Instance is not deleted during iteration ", instance_name)
            
            logging.info("Instance deletion details - response code: %s response body: %s" % (instance_deletion.status_code, instance_deletion.text))
            time.sleep(10)

            #self.environment.runner.quit()
        except Exception as exception:
            logging.info("error is: %s" % exception)
            return
    
    def on_stop(self):
        end_time = int(time.time() * 1000)
        logging.info("end-time : %s" % end_time)
    
    def stop_if_time_exceeded(self):
        if time.time() - self.script_start_time >= 28800 :
            self.environment.runner.quit()
