from src.utils.wrappers.transformers import HfTokenizer
from tests.consts import TestConsts

from typing import Dict, Any
import pytest
import yaml
import os


@pytest.fixture
def test_results(request) -> Dict[str, Any]:
    """Pytest fixture to load the specific test results.
       Gets the results file to load, returns parsed yaml as Dict[str, Any].
    """

    res_file_name = request.param

    if res_file_name is None:
        pytest.fail("Missing 'res_file_name'", pytrace=False)

    if (
        not res_file_name.endswith(".yaml") and
        not res_file_name.endswith(".yml")
    ):
        res_file_name = f"{res_file_name}.yaml"

    res_file_path = f"{TestConsts.RESULTS_DIR_PATH}/{res_file_name}"
    if not os.path.exists(res_file_path):
        pytest.fail(
            f"Test results file '{res_file_path}' not found.",
            pytrace=False
        )

    with open(res_file_path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)
    
    if not isinstance(data, dict):
        pytest.fail(
            "Invalid format: Expected dictionary but got " \
                + f"{type(data)} in '{res_file_path}'."
        )

    return data

@pytest.fixture
def server_config() -> Dict[str, Any]:
    """Pytest fixture to load the dev server config."""

    with open(TestConsts.SERVER_DEV_CONFIG_PATH, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)

    if not isinstance(data, dict):
        pytest.fail(
            "Invalid format: Expected dictionary but got " \
                + f"{type(data)} in '{TestConsts.SERVER_DEV_CONFIG_PATH}'."
        )

    return data

@pytest.fixture
def hf_tokenizer(request) -> Dict[str, Any]:
    """Pytest fixture to create a HfTokenizer for testing.
       Gets the model_id to load as tokenizer, returns the loaded HfTokenizer.
    """

    hf_token = os.getenv("HF_TOKEN")
    if hf_token is None:
        pytest.fail("Missing 'hf_token'", pytrace=False)

    model_id = request.param
    if model_id is None:
        pytest.fail("Missing 'model_id'", pytrace=False)

    return HfTokenizer(model_id)
