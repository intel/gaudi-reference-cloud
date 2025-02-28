from src.utils.wrappers.transformers import HfTokenizer
from tests.consts import TestConsts

from transformers import AutoTokenizer
from typing import Dict, Any
import pytest


@pytest.mark.parametrize("test_results", ["tokenizer.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
def test_tokenizer_initialization(
    hf_tokenizer: HfTokenizer,
    test_results: Dict[str, Any]
) -> None:
    """Tests initialization for the HF AutoTokenizer for all supported 
       tokenizers, and the HfTokenizer object initialization.
    """

    reference = test_results["test_tokenizer_initialization"]
    for ref in reference:
        # test all tokenizer loading
        tokenizer = AutoTokenizer.from_pretrained(
            ref["model_id"], 
            use_fast=ref["family"]
        )

        # test the HF object's fields
        assert tokenizer is not None
        assert tokenizer.eos_token_id == ref["eos_token_id"]
        assert tokenizer.eos_token == ref["eos_token"]

        # test MaaS HfTokenizer initialization with Qwen
        if ref["model_id"] == TestConsts.QWEN_MODEL_ID:
            assert hf_tokenizer is not None
            assert hf_tokenizer.is_ready == True
            assert hf_tokenizer.model_id == TestConsts.QWEN_MODEL_ID
            assert hf_tokenizer.eos_token_id == ref["eos_token_id"]
            assert hf_tokenizer.eos_token == ref["eos_token"]

@pytest.mark.parametrize("test_results", ["tokenizer.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
def test_tokenize(
    hf_tokenizer: HfTokenizer,
    test_results: Dict[str, Any]
) -> None:
    """Tests HfTokenizer 'tokenize' function"""

    ref = test_results["test_tokenize"]
    res = hf_tokenizer.tokenize(TestConsts.QUERY)
    assert len(res) == len(ref)

    for i in range(len(res)):    
        assert res[i].id == ref[i]["id"]
        assert res[i].text == ref[i]["text"]

@pytest.mark.parametrize("test_results", ["tokenizer.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)  
def test_apply_chat_template(
    hf_tokenizer: HfTokenizer,
    test_results: Dict[str, Any]
) -> None:
    """Tests HfTokenizer 'test_apply_chat_template' function"""

    ref = test_results["test_apply_chat_template"]
    messages = [{"role": "user", "content": TestConsts.QUERY}]
    
    ref_raw = ref["raw"]
    res_raw = hf_tokenizer.apply_chat_template(
        messages, add_generation_prompt=True, tokenize=False
    )
    assert ref_raw == res_raw
    
    ref_tokenized = ref["tokenized"]
    res_tokenized = hf_tokenizer.apply_chat_template(
        messages, return_dict=True
    )
    assert len(res_tokenized["input_ids"]) == len(ref_tokenized["input_ids"])
    assert len(res_tokenized["attention_mask"]) == ref_tokenized["attention_mask"]["len"]
    for i in range(len(ref_tokenized["input_ids"])):
        assert res_tokenized["input_ids"][i] == ref_tokenized["input_ids"][i]
        assert res_tokenized["attention_mask"][i] == ref_tokenized["attention_mask"]["value"]

@pytest.mark.parametrize("test_results", ["tokenizer.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
def test_count_tokens(
    hf_tokenizer: HfTokenizer,
    test_results: Dict[str, Any]
) -> None:
    """Tests HfTokenizer 'test_count_tokens' function"""

    messages = [{"role": "user", "content": TestConsts.QUERY}]
    ref_num_tokens = test_results["test_count_tokens"]
    res_num_tokens = hf_tokenizer.count_tokens(messages)
    assert ref_num_tokens == res_num_tokens
