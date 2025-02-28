import sys
import os

parent_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, parent_dir)

from locust import HttpUser, task, between, events
from utils.HttpClient import HttpClient
from datetime import datetime, timedelta
from jsonpath_ng import parse
import jsonpath_ng
from utils import ScaleUtil
import logging
import os
import json
import time
import jq
import urllib3
import time
import random

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class User(HttpUser):
    wait_time = between(1, 4)  # Time between consecutive requests
    
    # load the env variables
    key_value_dict = ScaleUtil.read_env_file("resources/env_local.txt")
    host = key_value_dict["host"]
    cloudaccount=key_value_dict["cloudaccount"]
    instance_endpoint = "/v1/cloudaccounts/"+cloudaccount+"/instances"
    http_client = HttpClient(base_url=host)
    token = payload_string =  ""

    def on_start(self):
        response = self.http_client.get_token()
        ScaleUtil.validate_status_code(response, 200)
        self.token = response.text
        self.raw_payload_string = ScaleUtil.read_json_file_as_string("resources/instance-payload.json")
        start_time = int(time.time() * 1000)
        logging.info("start-time : %s" % start_time)

    # proxies = {
    #      "https": "http://internal-placeholder.com:912",
    # }

    proxies = None
        
    @task
    def make_request(self):
        instance_phase = proxy_ip = proxy_user = instance_ip = instance_user = ""
        try:

            # Get the existing instances
            all_instances = self.http_client.get_all_method(self.instance_endpoint, self.token, self.proxies)
            jsonpath_expr = jsonpath_ng.parse("$.items[*].metadata.name")
            matches = [match.value for match in jsonpath_expr.find(all_instances.json())]
            captured_response = matches
            print("Instances present :", captured_response)

            # Delete one of the instance randomly
            random_item = random.choice(captured_response)
            print("Instance to be deleted :",random_item)

            deletion_time = datetime.now()
            instance_deletion = self.http_client.delete_method(self.instance_endpoint,self.token, random_item, self.proxies)
            ScaleUtil.validate_status_code(instance_deletion, 200)

            # wait for the instance to be deleted
            isInstanceDeleted = ScaleUtil.is_instance_deleted(self.http_client, deletion_time, self.instance_endpoint, self.token, random_item, self.proxies)
            logging.info("Instance deletion details - response code: %s response body: %s" % (instance_deletion.status_code, instance_deletion.text))
            time.sleep(10)

            instance_name, iteration_payload_string = ScaleUtil.get_instance_name_and_payload(self.raw_payload_string, 6)

            if isInstanceDeleted == True:
                # create the instance
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
            
        except Exception as exception:
            logging.info("error is: %s" % exception)
            return
    
    def on_stop(self):
        end_time = int(time.time() * 1000)
        logging.info("end-time : %s" % end_time)