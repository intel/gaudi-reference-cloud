from locust import HttpUser, task
import json

class HttpClientLocust:
    def __init__(self, base_url, proxy=None):
        self.base_url = base_url
        self.proxy = proxy
    
    def post_method(self, post_endpoint, token, payload, random_string):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + post_endpoint
        json_payload = json.loads(payload)
        response = self.client.post(url, json=json_payload, headers=headers_value)
        return response

    def get_method(self, get_endpoint, token, instance_name):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + get_endpoint + "/" + instance_name
        response = self.client.get(url, headers=headers_value)
        return response

    def put_method(self, put_endpoint, token, payload):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + put_endpoint
        json_payload = json.loads(payload)
        response = self.client.put(url, json=json_payload, headers=headers_value)
        return response

    def delete_method(self, delete_endpoint, token, instance_name):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + delete_endpoint + "/" + instance_name
        response = self.client.delete(url, headers=headers_value)
        return response

    def delete_all_instances(self, delete_endpoint, token, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + delete_endpoint +"/id/{}"
        response =self.client.delete(url, headers=headers_value, verify=False)
        return response

    def get_token(self):
        url = "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin"
        headers_value = {
            "Content-Type": "application/json"
        }
        response = self.client.get(url, headers=headers_value, verify=False)
        return response
    