from src.utils.builders.prompt import (
    ChatPromptBuilder, GeneratePromptBuilder
)
from src.utils.wrappers.transformers import HfTokenizer
from tests.consts import TestConsts

from typing import Dict, Any
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize("test_results", ["prompt_builder.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
async def test_generate_with_user_single_prompt(
    test_results: Dict[str, Any],
    hf_tokenizer: HfTokenizer
) -> None:
    """Tests 'GeneratePromptBuilder', 'with_user_single_prompt'."""

    ref = test_results["generate_with_user_single_prompt"]
    res = await GeneratePromptBuilder(tokenizer=hf_tokenizer) \
        .with_user_single_prompt(user=TestConsts.QUERY) \
        .build(return_messages_list=False)
    assert ref == res

@pytest.mark.asyncio
@pytest.mark.parametrize("test_results", ["prompt_builder.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
async def test_chat_completion_with_messages_list(
    test_results: Dict[str, Any],
    hf_tokenizer: HfTokenizer
) -> None:
    """Tests 'ChatPromptBuilder', 'with_messages_list'."""

    ref = test_results["chat_completion_with_messages_list"]
    messages = [{"role": "user", "content": TestConsts.QUERY}]
    res = await ChatPromptBuilder(tokenizer=hf_tokenizer) \
        .with_messages_list(messages=messages) \
        .build(return_messages_list=True)

    for res_element, ref_element in zip(res, ref):
        assert res_element["role"] == ref_element["role"]
        assert res_element["content"] == ref_element["content"]

@pytest.mark.asyncio
@pytest.mark.parametrize("test_results", ["prompt_builder.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.LLAMA_GUARD_MODEL_ID], indirect=True)
async def test_chat_completion_with_safeguard_single_prompt(
    test_results: Dict[str, Any],
    hf_tokenizer: HfTokenizer
) -> None:
    """Tests 'ChatPromptBuilder', 'with_safeguard_single_prompt'."""

    ref = test_results["chat_completion_with_safeguard_single_prompt"]
    res = await ChatPromptBuilder(tokenizer=hf_tokenizer) \
        .with_safeguard_single_prompt(user=TestConsts.QUERY) \
        .build(return_messages_list=False)
    assert ref == res
