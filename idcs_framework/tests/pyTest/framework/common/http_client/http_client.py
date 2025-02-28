import requests
from pyTest.framework.common.log_utils import logger

LOG = logger.get_logger(__name__)
api_details = {}
HTTPS = "https://"
HTTP = "http://"

class Requests:
    """
    This class is to create request wrapper or the all the rest calls
    and the response is divided in to json, text, and
    Status Code and saved in an dictionary
    """

    def __init__(self):
        self.retry = 5

    def http_request(self, url, method=None, **kwargs):
        """
        This Method is to create a wrapper for POST rest call
        :param url:
        :param method:
        :return:
        """
        LOG.debug(str(method) + " call made on the resource URL: " + url)
        api_details.update({'url': str(url)})
        try:
            data = getattr(requests, method)(url, **kwargs)
            if method in ["post", "put", "patch"]:
                LOG.info("Request body used {}".format(kwargs['data']))
            # LOG.info("Header used {}".format(kwargs['headers']))
            api_details.update({
                'resp_time': str(data.elapsed.total_seconds())})
            api_details.update({'status_code': str(data.status_code)})
            api_details.update({'method': str(method.upper())})
            msg = \
                ('| {:<100} | {:<12} seconds | {:<11} | {:<5}'.format(
                    api_details['url'],
                    api_details['resp_time'],
                    api_details['status_code'],
                    api_details['method']))
            if "download" in url:
                return data
            response = self.return_values(data)
            return response
        except requests.exceptions.HTTPError as errh:
            print("Http Error:", errh)
        except requests.exceptions.ConnectionError as errc:
            print("Error Connecting:", errc)
        except requests.exceptions.Timeout as errt:
            print("Timeout Error:", errt)
        except requests.exceptions.RequestException as err:
            print("OOps: Something Else", err)

    def return_values(self, obj):
        """
        This method is segregates the response in to status code,
        json response, text response ans stores in a
        dictionary
        :param obj:
        :return: return_dict
        """
        text_response = obj.text
        try:
            json_response_body = obj.json()
        except ValueError:
            json_response_body = None
        response_code = obj.status_code
        response_time = obj.elapsed.total_seconds()
        LOG.debug("Status Code: " + str(response_code))

        if not text_response:
            pass
        else:
            LOG.debug("Response Text: " + text_response)
        return_dict = {
            "status_code": response_code,
            "json_response": json_response_body,
            "text_response": text_response,
            "response_time": response_time
        }
        return return_dict