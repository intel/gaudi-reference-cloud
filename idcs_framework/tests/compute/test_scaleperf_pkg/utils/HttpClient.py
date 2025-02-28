import requests
import json
#import urllib3

#urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class HttpClient:
    def __init__(self, base_url):
        self.base_url = base_url

    def post_method(self, post_endpoint, token, payload, random_string, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + post_endpoint
        json_payload = json.loads(payload)
        response = requests.post(url, json=json_payload, headers=headers_value, verify=False, proxies=proxies)
        #assert response.status_code == 200, "status code is not 200"
        return response

    def get_method(self, get_endpoint, token, get_instance, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + get_endpoint + "/name/" + get_instance
        response = requests.get(url, headers=headers_value, verify=False, proxies=proxies)
        #assert response.status_code == 200, "status code is not 200"
        return response

    def get_all_method(self, get_endpoint, token, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + get_endpoint
        response = requests.get(url, headers=headers_value, verify=False, proxies=proxies)
        #assert response.status_code == 200, "status code is not 200"
        return response

    def put_method(self, put_endpoint, token, payload, proxies):
        pass  # Implement the put method as needed

    def delete_method(self, delete_endpoint, token, delete_instance, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + delete_endpoint + "/name/" + delete_instance
        response = requests.delete(url, headers=headers_value, verify=False, proxies=proxies)
        #assert response.status_code == 200, "status code is not 200"
        return response

    def delete_all_instances(self, delete_endpoint, token, proxies):
        headers_value = {
            "Authorization": "Bearer " + str(token),
            "Content-Type": "application/json"
        }
        url = self.base_url + delete_endpoint +"/id/{}"
        response = requests.delete(url, headers=headers_value, verify=False, proxies=proxies)
        return response

    def get_token(self):
        url = "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin"
        headers_value = {
            "Content-Type": "application/json"
        }
        response = requests.get(url, headers=headers_value, verify=False)
        return response
    