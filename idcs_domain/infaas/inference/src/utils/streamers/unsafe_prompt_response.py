import src.inference_engine.grpc_files.infaas_generate_pb2 as pb2
from src.utils.entities.tokenization_result import TokenizationResult
from src.utils.wrappers.transformers import HfTokenizer

from google.protobuf.message import Message
from typing import AsyncIterator, List


class UnsafePromptResponseStreamer:

    def __init__(
        self,
        tokenizer: HfTokenizer,
        resp_type: Message
    ) -> None:

        self._unsafe_message_template = "I'm sorry but I can't help on any " +\
            "{harm_category} related topics."
        self._unsafe_message = None

        self._tokenizer = tokenizer
        self._resp_type = resp_type

    async def _tgi_chat_completion_stream(
        self,
        tokenized_unsafe_message: List[TokenizationResult]
    ) -> AsyncIterator[pb2.ChatCompletionStreamResponse]:

        for token in tokenized_unsafe_message:

            yield pb2.ChatCompletionStreamResponse(
                choices=[
                    pb2.ChatCompletionStreamChoice(
                        index=0,
                        delta={"role": "assistant", "content": token.text}
                    )
                ]
            )

        # After finished to stream the entire message -
        # stream the EOS token with generation details
        yield pb2.ChatCompletionStreamResponse(
            choices=[
                pb2.ChatCompletionStreamChoice(
                    index=0,
                    delta={
                        "role": "assistant",
                        "content": self._tokenizer.eos_token
                    },
                    finish_reason="eos_token",
                    usage=pb2.ChatCompletionUsage(
                        completion_tokens=len(tokenized_unsafe_message)
                    )
                )
            ]
        )

    async def _tgi_generate_stream(
        self,
        tokenized_unsafe_message: List[TokenizationResult]
    ) -> AsyncIterator[pb2.GenerateStreamResponse]:

        for token in tokenized_unsafe_message:
            yield pb2.GenerateStreamResponse(
                token=pb2.GenerateAPIToken(
                    id=token.id,
                    text=token.text
                )
            )

        # After finished to stream the entire message -
        # stream the EOS token with generation details
        yield pb2.GenerateStreamResponse(
            token=pb2.GenerateAPIToken(
                id=self._tokenizer.eos_token_id,
                text=self._tokenizer.eos_token,
                special=True
            ),
            details=pb2.StreamDetails(
                finish_reason=pb2.FinishReason.FINISH_REASON_EOS_TOKEN,
                generated_tokens=len(tokenized_unsafe_message)
            ),
            generated_text=self._unsafe_message
        )

    async def stream(self, harm_category: str) -> AsyncIterator[Message]:

        self._unsafe_message = self._unsafe_message_template.format(
            harm_category=harm_category.lower()
        )

        tokenized_unsafe_message = self._tokenizer.tokenize(
            text=self._unsafe_message
        )

        if self._resp_type is pb2.ChatCompletionStreamResponse:
            llm_function = self._tgi_chat_completion_stream
        elif self._resp_type is pb2.GenerateStreamResponse:
            llm_function = self._tgi_generate_stream

        async for pred in llm_function(
            tokenized_unsafe_message=tokenized_unsafe_message
        ):
            yield pred
