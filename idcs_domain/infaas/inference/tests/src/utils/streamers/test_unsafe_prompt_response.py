from src.utils.streamers.unsafe_prompt_response import UnsafePromptResponseStreamer
from src.utils.wrappers.transformers import HfTokenizer
from tests.consts import TestConsts
import src.inference_engine.grpc_files.infaas_generate_pb2 as pb2

from typing import Dict, Any
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize("test_results", ["unsafe_stream.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
async def test_generate_stream_unsafe_resp(
    test_results: Dict[str, Any],
    hf_tokenizer: HfTokenizer
) -> None:
    """Tests 'GenerateStream' unsafe prompt stream."""

    streamer = UnsafePromptResponseStreamer(
        tokenizer=hf_tokenizer,
        resp_type=pb2.GenerateStreamResponse
    )

    i = 0
    async for chunk in streamer.stream(TestConsts.HARM_CATEGORY):
        assert isinstance(chunk, pb2.GenerateStreamResponse)
        assert isinstance(chunk.token, pb2.GenerateAPIToken)
        assert chunk.token.id == test_results["resp_tokens"][i]["id"]
        assert chunk.token.text == test_results["resp_tokens"][i]["text"]
        i += 1

    assert isinstance(chunk.details, pb2.StreamDetails)
    assert chunk.details.finish_reason == pb2.FinishReason.FINISH_REASON_EOS_TOKEN
    assert chunk.details.generated_tokens == i - 1
    assert chunk.generated_text == TestConsts.UNSAFE_MESSAGE

@pytest.mark.asyncio
@pytest.mark.parametrize("test_results", ["unsafe_stream.yaml"], indirect=True)
@pytest.mark.parametrize("hf_tokenizer", [TestConsts.QWEN_MODEL_ID], indirect=True)
async def test_chat_completion_stream_unsafe_resp(
    test_results: Dict[str, Any],
    hf_tokenizer: HfTokenizer
) -> None:
    """Tests 'ChatCompletion' unsafe prompt stream."""

    streamer = UnsafePromptResponseStreamer(
        tokenizer=hf_tokenizer,
        resp_type=pb2.ChatCompletionStreamResponse
    )

    i = 0
    async for chunk in streamer.stream(TestConsts.HARM_CATEGORY):
        assert isinstance(chunk.choices[0], pb2.ChatCompletionStreamChoice)
        assert chunk.choices[0].index == 0
        assert chunk.choices[0].delta.role == "assistant"
        assert chunk.choices[0].delta.content == test_results["resp_tokens"][i]["text"]
        i += 1

    usage = chunk.choices[0].usage
    assert isinstance(usage, pb2.ChatCompletionUsage)
    assert usage.completion_tokens == i - 1
    assert chunk.choices[0].finish_reason == "eos_token"
