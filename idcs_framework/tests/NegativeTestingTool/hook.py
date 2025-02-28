import pytest
import json
from hypothesis import settings
import sys
import schemathesis
from datetime import timedelta
from hypothesis import strategies as st

CONFIG_PATH = "./config.json"
environment = print(sys.argv[1])
token = print(sys.argv[2])

def get_environment_url(environment):
    f = open(CONFIG_PATH)
    json_data = json.loads(f.read())
    url = json_data[environment]
    f.close()
    return url


open_api_url = get_environment_url(environment)

schema = schemathesis.from_uri(open_api_url)

@pytest.fixture(scope="session")
def token():
    return token

@schema.parametrize()
@settings(max_examples=25)
def test_app(case, token):
    case.headers = {"Authorization": f"Bearer {token}"}
    response = case.call()
    case.validate_response(response)
    
@schemathesis.check
def not_so_slow(response, case):
    """Custom response check."""
    assert response.elapsed < timedelta(milliseconds=100), "Response is slow!"

@schemathesis.target
def big_response(context):
    """Custom data generation target."""
    return float(len(context.response.content))