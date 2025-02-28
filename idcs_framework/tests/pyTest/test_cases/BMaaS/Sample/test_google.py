import pytest
from pyTest.framework.common.http_client.http_client import Requests
from pyTest.framework.common.log_utils.logger import get_logger

LOG = get_logger(__name__)

class TestGoogle:
    pytestmark = [pytest.mark.BMasS_login]
    @classmethod
    def _handler_creation(cls, request):
        LOG.info(f"***** TESTING SAMPLE GOOGLE API *****")
        cls.request_obj = Requests()

    def test_get(self, request):
        url = "http://google.com"
        self._handler_creation(request)
        resp = self.request_obj.http_request(
            url, method="get")
        assert resp["status_code"] == 200


