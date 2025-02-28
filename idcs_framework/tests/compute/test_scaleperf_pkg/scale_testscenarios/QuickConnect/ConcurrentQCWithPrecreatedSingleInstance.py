from locust import HttpUser, task, between, events
from utils import ScaleUtil
import logging

class User(HttpUser):
    wait_time = between(1, 4)

    key_value_dict = ScaleUtil.read_env_file("resources/env.txt")
    host = key_value_dict["host"]
    cloudaccount=key_value_dict["cloudaccount"]
    qcBaseUrl = key_value_dict["qcurl"]
    OauthHMAC = key_value_dict["OauthHMAC"]
    OauthExpires = key_value_dict["OauthExpires"]
    BearerToken = key_value_dict["BearerToken"]
    IdToken = key_value_dict["IdToken"]
    RefreshToken = key_value_dict["RefreshToken"]
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.cookies_value = {
            "OauthHMAC": self.OauthHMAC,
            "OauthExpires": self.OauthExpires,
            "BearerToken": self.BearerToken,
            "IdToken": self.IdToken,
            "RefreshToken": self.RefreshToken
        }

    proxies = {
        "https": "http://internal-placeholder.com:912",
    }
    
    @task
    def make_request(self):
        qc_endpoint =  self.qcBaseUrl +"/"+self.cloudaccount+"/"+"<<instance-id>>"+"/lab?"
        qc_instance_response = self.client.get(qc_endpoint, cookies=self.cookies_value, verify=False, proxies=self.proxies)
        ScaleUtil.validate_status_code(qc_instance_response, 200)
        ScaleUtil.validate_response_body(qc_instance_response, "JupyterLab")

