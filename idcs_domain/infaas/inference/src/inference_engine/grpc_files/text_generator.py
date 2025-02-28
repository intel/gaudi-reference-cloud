from src.utils.errors.abstract import (
    MaasErrorWithStacktrace,
    AbstractMaasError
)
from src.utils.entities.text_generation_context import TextGenerationContext
from src.utils.streamers.unsafe_prompt_response import UnsafePromptResponseStreamer
from src.utils.entities.safeguard_data import SafeguardData
from src.utils.entities.server_prob_data import ServerProbData
from src.utils.handlers.llm_response import LlmResponseHandler
from src.utils.factories.backends import BackendFactory
from src.utils.wrappers.transformers import HfTokenizer
from src.utils.errors.safeguard import SafeguardNotResponsiveError
from src.utils.builders.prompt import GeneratePromptBuilder, ChatPromptBuilder
import src.inference_engine.grpc_files.infaas_generate_pb2_grpc as pb2_grpc
import src.inference_engine.grpc_files.infaas_generate_pb2 as pb2

from google.protobuf.json_format import MessageToDict
from google.protobuf.message import Message
from collections.abc import AsyncIterator
from typing import Any, Optional, Dict, List, Tuple
import traceback
import logging
import asyncio
import grpc
import copy


class TextGenerator(pb2_grpc.TextGeneratorServicer):

    def __init__(
        self,
        safeguard_url: str,
        config: Dict[str, Any],
        base_url: Optional[str] = None,
        headers: Optional[Dict[str, str] | None] = None,
        cookies: Optional[Dict[str, str] | None] = None,
        timeout: Optional[int] = None,
        backend: Optional[str] = None
    ) -> None:

        self._config = config
        self._safeguard_config = config["server"]["safeguard"]

        self._llm_engine, self._sg_engine = BackendFactory.backend_wrapper(
            config=self._config,
            backend=backend,
            base_url=base_url,
            safeguard_url=safeguard_url,
            headers=headers,
            cookies=cookies,
            timeout=timeout
        )

        self._llm_tokenizer = HfTokenizer(
            model_id=self._config["server"]["model_id"]
        )

        self._sg_tokenizer = HfTokenizer(
            model_id=self._safeguard_config["model_id"]
        )

    @property
    def is_ready(self) -> bool:
        llm_tokenizer_is_ready = self._llm_tokenizer.is_ready
        sg_tokenizer_is_ready = self._sg_tokenizer.is_ready
        return llm_tokenizer_is_ready and sg_tokenizer_is_ready

    @property
    def server_prob_config(self) -> ServerProbData:
        return self._llm_engine.get_prob_data()

    async def ChatCompletionStream(
        self,
        request: pb2.ChatCompletionStreamRequest,
        context: grpc.RpcContext
    ) -> AsyncIterator[pb2.ChatCompletionStreamResponse]:

        try:
            await self._log_new_request(request)
            tokens_q = asyncio.Queue()
            raw_prompt = "\n".join(
                f"role: {message.role}\ncontent: {message.content}"
                for message in request.messages
            )

            await self._set_system_message(request.messages)
            safeguard_data = await self._create_safeguard_task(raw_prompt)

            cast_response_function = LlmResponseHandler(
                tokenizer=self._llm_tokenizer,
                inst_tokens=self._config["server"]["inference"]["inst_tokens"]
            ).handle_tgi_chat_completion_stream

            text_generator_context = TextGenerationContext(
                infer=self._llm_engine.chat_completion,
                cast_response=cast_response_function,
                cast_response_payload={
                    "messages": request.messages,
                    "requestID": request.requestID
                },
                unsafe_prompt_streamer=UnsafePromptResponseStreamer(
                    tokenizer=self._llm_tokenizer,
                    resp_type=pb2.ChatCompletionStreamResponse
                )
            )

            async for new_pred in self._infer(
                    request=request,
                    tokens_q=tokens_q,
                    safeguard_data=safeguard_data,
                    text_generator_context=text_generator_context
            ):
                yield new_pred

        except AbstractMaasError as e:
            await self._handle_abstract_maas_error(request, context, e)

        # Unexpected Error
        except Exception as e:
            await self._handle_general_exception(request, context, e)

    async def GenerateStream(
        self,
        request: pb2.GenerateStreamRequest,
        context: grpc.RpcContext
    ) -> AsyncIterator[pb2.GenerateStreamResponse]:

        try:
            await self._log_new_request(request)
            tokens_q = asyncio.Queue()
            raw_prompt = copy.deepcopy(request.prompt)

            request.prompt = \
                await GeneratePromptBuilder(tokenizer=self._llm_tokenizer) \
                    .with_user_single_prompt(user=request.prompt) \
                    .build(return_messages_list=False)

            safeguard_data = await self._create_safeguard_task(raw_prompt)
            cast_response_function = LlmResponseHandler(
                tokenizer=self._llm_tokenizer,
                inst_tokens=self._config["server"]["inference"]["inst_tokens"]
            ).handle_tgi_generate_stream

            text_generator_context = TextGenerationContext(
                infer=self._llm_engine.generate_stream,
                cast_response=cast_response_function,
                cast_response_payload={"requestID": request.requestID},
                unsafe_prompt_streamer=UnsafePromptResponseStreamer(
                    tokenizer=self._llm_tokenizer,
                    resp_type=pb2.GenerateStreamResponse
                )
            )

            async for new_pred in self._infer(
                    request=request,
                    tokens_q=tokens_q,
                    safeguard_data=safeguard_data,
                    text_generator_context=text_generator_context
            ):
                yield new_pred

        except AbstractMaasError as e:
            await self._handle_abstract_maas_error(request, context, e)

        # Unexpected Error
        except Exception as e:
            await self._handle_general_exception(request, context, e)

    async def _set_system_message(
        self,
        request_messages: List[pb2.ChatCompletionMessage]
    ) -> None:
        """uses the PromptBuilder specific build function to insert system
           messages into the user's messages list. works in place (changes
           input object).

        Args:
            request_messages (List[pb2.ChatCompletionMessage]): original 
                request list of messages.
        """
        messages_list = [
            MessageToDict(
                message=message,
                preserving_proto_field_name=True
            ) for message in request_messages
        ]

        messages_list = \
            await ChatPromptBuilder(tokenizer=self._llm_tokenizer) \
                .with_messages_list(messages=messages_list) \
                .build(return_messages_list=True)

        for req_message, list_message in \
                zip(request_messages, messages_list):
            req_message.role = list_message["role"]
            req_message.content = list_message["content"]

    async def _log_new_request(self, request: Message, ) -> None:

        logging.info(
            f"got new {request.__class__.__name__} | " +
            f"request id: {request.requestID}"
        )

        logging.debug(
            f"request body:\n" +
            str(
                MessageToDict(
                    message=request,
                    preserving_proto_field_name=True
                )
            )
        )

    async def _infer(
        self,
        request: Message,
        tokens_q: asyncio.Queue,
        safeguard_data: SafeguardData,
        text_generator_context: TextGenerationContext
    ) -> AsyncIterator[Message]:
        """main inference handling logic.
           executes two tasks that runs concurrently - 
           llm main tokens generation and safeguard tasks.
           we keep generated tokens in an async queue until the
           safeguard task is done and we can decide whether the prompt
           is safe or not.

        Args:
            request (Message): a gRPC message of `ChatCompletionMessage` or
                               `GenerateStream`.
            tokens_q (asyncio.Queue): llm generated tokens queue.
                                      keeps all generated tokens until the
                                      safeguard task is done, and we know
                                      if the prompt is safe.
            llm_function (Callable): TgiWrapper's function to create tokens
                                     iterator and iterate over generated tokens.
            safeguard_data (SafeguardData): holds all safeguard related fields-
                                            safeguard task object, safeguard
                                            prompt and raw prompt.

        Raises:
            SafeguardNotResponsiveError: safeguard task has to be done before
                                         llm generation is. if this condition
                                         wasn't met - we can't decide wether
                                         the prompt is safe and we raise this
                                         error. safeguard task is way simpler
                                         than llm's, hence this scinario
                                         inducates an error in the safeguard.

        Yields:
            Iterator[AsyncIterator[Message]]: iterator for llm's generated
                                              tokens.
        """

        request_dict = await self._get_request_as_dict(request)
        is_prompt_safe = None
        num_generated_tokens = 0

        async for new_pred in text_generator_context.infer(**request_dict):

            num_generated_tokens += 1

            await tokens_q.put(
                await text_generator_context.cast_response(
                    response=new_pred,
                    num_generated_tokens=num_generated_tokens,
                    **text_generator_context.cast_response_payload
                )
            )

            if is_prompt_safe is None and safeguard_data.task.done():
                is_prompt_safe, harm_category = await safeguard_data.task

            if is_prompt_safe is not None:
                if is_prompt_safe == True:
                    yield await tokens_q.get()
                else:
                    break

        if not safeguard_data.task.done():
            is_prompt_safe, harm_category = await self._wait_on_safeguard_task(
                safeguard_data=safeguard_data
            )

        if is_prompt_safe:
            # drain tokens queue if needed
            while not tokens_q.empty():
                yield await tokens_q.get()
        else:
            unsafe_streamer = text_generator_context.unsafe_prompt_streamer
            async for token in unsafe_streamer.stream(harm_category):
                yield token

    async def _wait_on_safeguard_task(
        self,
        safeguard_data: SafeguardData
    ) -> Tuple[bool, str]:
        """Waits until safeguard task is done for the configured safeguard timeout period.
           If safeguard task is completed during the period - returns task results.
           If not - throws 'SafeguardNotResponsiveError'.
        """
        safeguard_timeout = self._safeguard_config["timeout"]
        waited_count = 0
        while waited_count < safeguard_timeout:
            # wait 1s and check if safeguard task is done
            await asyncio.sleep(1)
            if safeguard_data.task.done():
                is_prompt_safe, harm_category = await safeguard_data.task
                return is_prompt_safe, harm_category
            waited_count += 1

        # if we got to the end of the loop without returning that means
        # the safeguard wasn't done during the timeout period so we cancel
        # it and throw exception
        safeguard_data.task.cancel()
        raise SafeguardNotResponsiveError(safeguard_data.raw_prompt)

    async def _create_safeguard_task(self, prompt: str) -> SafeguardData:

        safeguard_data = SafeguardData(raw_prompt=prompt)

        safeguard_data.set_messages(
            await ChatPromptBuilder(tokenizer=self._sg_tokenizer) \
                .with_safeguard_single_prompt(user=prompt) \
                .build(return_messages_list=True)
        )

        safeguard_data.set_task(
            asyncio.create_task(
                self._sg_engine.is_prompt_safe(safeguard_data.messages)
            )
        )

        return safeguard_data

    async def _get_request_as_dict(self, request: Message) -> Dict[str, Any]:
        # The if-else logic is due to the difference between
        # the OpenAI compatible and non-compatible requests.
        # In `ChatCompletionStreamRequest` all params are in the
        # request level, where in `GenerateStreamRequest`, there is
        # internal params object.
        if isinstance(request, pb2.ChatCompletionStreamRequest):
            request_dict = MessageToDict(
                message=request,
                preserving_proto_field_name=True
            )

        elif isinstance(request, pb2.GenerateStreamRequest):
            request_dict = MessageToDict(
                message=request.params,
                preserving_proto_field_name=True
            )
            request_dict["requestID"] = request.requestID
            request_dict["prompt"] = request.prompt

        return request_dict

    async def _log_error(
        self,
        e: Exception,
        request_id: str,
        add_stacktrace: bool,
        invocation_meta: Optional[Any] = None
    ) -> str:
        """
        Function to log error and handle enviroment information additions.

        Args:
            e (Exception): _description_
            request_id (str): _description_
            add_stacktrace (bool): _description_
            invocation_meta (Optional[Any], optional): _description_. Defaults to None.

        Returns:
            str: A 'raw' error without invocation meta or stacktrace
        """

        raw_err = f"Request ID: {request_id}"
        log_err = f"Request ID: {request_id}"

        if invocation_meta is not None:
            log_err += f" | Invocation Meta: {invocation_meta}"

        raw_err += f" | Error Msg: {e}"
        log_err += f" | Error Msg: {e}"

        if add_stacktrace:
            log_err += f" | Stack Trace:\n{traceback.format_exc()}"

        logging.error(log_err)
        return raw_err

    async def _extract_request_meta_safe(
        self,
        context: grpc.RpcContext
    ) -> Any:

        try:
            return context.invocation_metadata()
        except Exception:
            return

    async def _handle_general_exception(
        self,
        request: Message,
        context: grpc.RpcContext,
        e: Exception
    ) -> None:

        request_meta = await self._extract_request_meta_safe(context)

        err_msg = await self._log_error(
            request_id=request.requestID,
            invocation_meta=request_meta,
            add_stacktrace=True,
            e=e
        )

        await context.abort(
            grpc.StatusCode.UNKNOWN,
            err_msg
        )

    async def _handle_abstract_maas_error(
        self,
        request: Message,
        context: grpc.RpcContext,
        e: Exception
    ) -> None:

        request_meta = await self._extract_request_meta_safe(context)

        add_stacktrace = (
            True if isinstance(e, MaasErrorWithStacktrace)
            else False
        )

        err_msg = await self._log_error(
            request_id=request.requestID,
            invocation_meta=request_meta,
            add_stacktrace=add_stacktrace,
            e=e
        )

        await context.abort(
            grpc.StatusCode.INTERNAL,
            err_msg
        )
