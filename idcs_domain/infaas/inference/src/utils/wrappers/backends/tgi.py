from src.utils.wrappers.backends.abstract import AbstractWrapper
from src.utils.entities.server_prob_data import ServerProbData

from text_generation.types import StreamResponse, ChatCompletionChunk, Message
from text_generation import AsyncClient
from typing import (
    AsyncIterator, Dict, List, Any, Optional, Tuple
)


class TgiWrapper(AbstractWrapper):

    _PROB_ENDPOINT = "v1/chat/completions"

    def __init__(
        self,
        config: Dict[str, Any],
        base_url: str,
        model_id: str | None,
        headers: Dict[str, str] | None,
        cookies: Dict[str, str] | None,
        timeout: int
    ) -> None:
        """
        Initiates a TGI async client and exposes its streaming.

        Args:
            config (Dict[str, Any]):
                config values
            base_url (str):
                OpenAI server's base url.
            model_id (str):
                the model ID
            headers (Dict[str, str], optional):
                default headers to the TGI server requests.
            cookies (Dict[str, str], optional):
                cookies info to add with the TGI server requests.
            timeout (int, optional):
                request timeout in seconds.
        """

        super().__init__(config, base_url, None, headers, cookies, timeout)

        # if all validations were passed - create a protected client
        self._client = AsyncClient(
            base_url=base_url,
            headers=headers,
            cookies=cookies,
            timeout=timeout
        )

        self._model_id = model_id

    def _messages_to_tgi_obj(
            self,
            messages: List[Dict[str, str]],
    ) -> list[Message]:
        """
        Function to set whether input prompt is safe or not.
        To be used with Safeguard.

        Args:
            messages (List[Dict[str, str]]):
                Messages list of messages.
        Returns:
            list[Message]
        """
        tgi_messages = [
            Message(role=message["role"], content=message["content"])
            for message in messages
        ]
        return tgi_messages

    async def is_prompt_safe(
            self,
            messages: List[Dict[str, str]],
            max_tokens: int = 10,
            temperature: float = 0.01
    ) -> Tuple[bool, str | None]:
        """
        Function to set whether input prompt is safe or not.
        To be used with Safeguard.

        Args:
            messages (List[Dict[str, str]]):
                Messages list of safeguard system and user message.
            max_new_tokens (int, optional):
                Maximum number of generated tokens.
                Safeguard's output suppose to be short - hence defaults to 10.
            temperature (float, optional):
                The value used to module the logits distribution.
                Effects the model level of "creativity".
                Defaults is set to 0.01 - very low value - to make the model
                stick to internal instructions as much as possible.

        Returns:
            Tuple[bool, str | None]
        """

        safety_response = await self._client.chat(
            messages=self._messages_to_tgi_obj(messages),
            max_tokens=max_tokens,
            temperature=temperature,
            stream=False
        )

        return await self._check_safety_response(safety_response)

    def get_prob_data(self) -> ServerProbData:

        return self._get_prob_data(prob_endpoint=self._PROB_ENDPOINT, model_id=self._model_id)

    async def generate_stream(
        self,
        requestID: str,
        prompt: str,
        do_sample: bool = False,
        max_new_tokens: int | None = None,
        repetition_penalty: float | None = None,
        frequency_penalty: float = None,
        return_full_text: bool = True,
        seed: int | None = None,
        stop_sequences: List[str] = [],
        temperature: float | None = None,
        top_k: int | None = None,
        top_p: float | None = None,
        truncate: int | None = None,
        typical_p: float | None = None,
        watermark: bool = False,
        top_n_tokens: int | None = None
    ) -> AsyncIterator[StreamResponse]:
        """
        Main generation function that wraps the TGI async clinet's
        generate_stream. Streams out the model generated tokens one
        by one, as a StreamResponse objects.

        Args:
            requestID (`str`):
                Internal MaaS request ID for logging.
            prompt (str):
                The prompt (textual instructions) used for generation.
            do_sample (bool, optional):
                Wether to use any sampling strategy other than greedy decoding.
            max_new_tokens (int | None, optional):
                Maximum number of generated tokens.
                If input is None (default) - will be set using internal
                _NumericParams.
            repetition_penalty (float | None, optional):
                The parameter for repetition penalty. 1.0 means no penalty.
                See [this paper](https://arxiv.org/pdf/1909.05858.pdf)
                for more details.
            frequency_penalty (float, optional):
                The parameter for frequency penalty. 1.0 means no penalty
                Penalize new tokens based on their existing frequency in
                the text so far, decreasing the model's likelihood to repeat
                the same line verbatim.
            return_full_text (bool, optional):
                Whether to prepend the prompt to the generated text.
            seed (int | None, optional):
                Random sampling seed, in use when `do_sample` is true and
                model doesn't do greedy decoding.
            stop_sequences (List[str], optional):
                Stop generating tokens if a member of `stop_sequences`
                is generated.
            temperature (float | None, optional):
                The value used to module the logits distribution.
                Effects the model level of "creativity".
            top_k (int | None, optional):
                The number of highest probability vocabulary tokens
                to keep for top-k-filtering.
            top_p (float | None, optional):
                If set to < 1, only the smallest set of most probable tokens
                with probabilities that add up to `top_p` or higher are kept
                for generation.
            truncate (int | None, optional):
                Truncate inputs tokens to the given size.
            typical_p (float | None, optional):
                Typical Decoding mass.
                See [Typical Decoding for Natural Language Generation]
                (https://arxiv.org/abs/2202.00666) for more information.
            watermark (bool, optional):
                Watermarking with [A Watermark for Large Language Models]
                (https://arxiv.org/abs/2301.10226).
            top_n_tokens (int | None, optional):
                Used to return information about the the `n` most likely
                tokens at each generation step, instead of just the
                sampled token.

        Yields:
            AsyncIterator[StreamResponse]
        """

        # tgi numeric params must be greater than zero or None
        # gRPC by default would give 0 to optional not populated fields
        tgi_np = self._NumericParams(
            requestID=requestID,
            config=self._config,
            max_new_tokens=max_new_tokens,
            repetition_penalty=repetition_penalty,
            frequency_penalty=frequency_penalty,
            seed=seed,
            temperature=temperature,
            top_k=top_k,
            top_p=top_p,
            truncate=truncate,
            typical_p=typical_p,
            top_n_tokens=top_n_tokens
        )

        # TGI client has an internal validation of the generation params
        # If a validation fails - validator will raise a 'ValidationError'
        stream_iterator = self._client.generate_stream(
            prompt=prompt,
            do_sample=do_sample,
            max_new_tokens=tgi_np.max_new_tokens,
            repetition_penalty=tgi_np.repetition_penalty,
            frequency_penalty=tgi_np.frequency_penalty,
            return_full_text=return_full_text,
            seed=tgi_np.seed,
            stop_sequences=stop_sequences,
            temperature=tgi_np.temperature,
            top_k=tgi_np.top_k,
            top_p=tgi_np.top_p,
            truncate=tgi_np.truncate,
            typical_p=tgi_np.typical_p,
            watermark=watermark,
            top_n_tokens=tgi_np.top_n_tokens
        )

        async for new_pred in stream_iterator:
            yield new_pred

    async def chat_completion(
        self,
        requestID: str,
        messages: List[Dict[str, str]],
        repetition_penalty: Optional[float] = None,
        frequency_penalty: Optional[float] = None,
        logit_bias: Optional[List[float]] = None,
        return_logprobs: Optional[bool] = None,
        top_logprobs: Optional[int] = None,
        max_tokens: Optional[int] = None,
        num_completions: Optional[int] = None,
        presence_penalty: Optional[float] = None,
        seed: Optional[int] = None,
        temperature: Optional[float] = None,
        top_p: Optional[float] = None,
        # We're not supporting tools at the moment
        # tools: Optional[List[Tool]] = None,
        # tool_choice: Optional[str] = None,
    ) -> AsyncIterator[ChatCompletionChunk]:
        """
        Given a list of messages, generate a response asynchronously

        Args:
            requestID (`str`):
                Internal MaaS request ID for logging.
            messages (`List[Dcit[str, str]]`):
                List of messages
            repetition_penalty (`float`):
                The parameter for frequency penalty. 0.0 means no penalty.
                See [this paper](https://arxiv.org/pdf/1909.05858.pdf)
                for more details.
            frequency_penalty (`float`):
                The parameter for frequency penalty. 0.0 means no penalty
                Penalize new tokens based on their existing frequency in
                the text so far, decreasing the model's likelihood to repeat
                the same line verbatim.
            logit_bias (`List[float]`):
                Adjust the likelihood of specified tokens
            return_logprobs (`bool`):
                Include log probabilities in the response
            top_logprobs (`int`):
                Include the `n` most likely tokens at each step
            max_tokens (`int`):
                Maximum number of generated tokens
            num_completions (`int`):
                Generate `n` completions
            presence_penalty (`float`):
                The parameter for presence penalty. 0.0 means no penalty.
                See [this paper](https://arxiv.org/pdf/1909.05858.pdf)
                for more details.
            stream (`bool`):
                Stream the response
                ** NOT SUPPORTED AT THE MOMENT (WE DO STREAMING ONLY)
            seed (`int`):
                Random sampling seed
            temperature (`float`):
                The value used to module the logits distribution.
            top_p (`float`):
                If set to < 1, only the smallest set of most probable tokens
                with probabilities that add up to `top_p` or higher are kept
                for generation.
            tools (`List[Tool]`):
                List of tools to use
                ** NOT SUPPORTED AT THE MOMENT
            tool_choice (`str`):
                The tool to use
                ** NOT SUPPORTED AT THE MOMENT
        """

        # tgi numeric params must be greater than zero or None
        # gRPC by default would give 0 to optional not populated fields
        tgi_np = self._NumericParams(
            requestID=requestID,
            config=self._config,
            max_new_tokens=max_tokens,
            repetition_penalty=repetition_penalty,
            frequency_penalty=frequency_penalty,
            seed=seed,
            temperature=temperature,
            top_p=top_p,
            top_logprobs=top_logprobs,
            n=num_completions
        )

        chat_completion = self._client.chat(
            messages=self._messages_to_tgi_obj(messages),
            repetition_penalty=None,
            frequency_penalty=tgi_np.frequency_penalty,
            logit_bias=logit_bias,
            logprobs=return_logprobs,
            top_logprobs=top_logprobs,
            max_tokens=tgi_np.max_new_tokens,
            n=tgi_np.n,
            presence_penalty=presence_penalty,
            # We support only streaming at the moment
            stream=True,
            seed=tgi_np.seed,
            temperature=tgi_np.temperature,
            top_p=tgi_np.top_p,
            tools=None,
            tool_choice=None
        )

        async for new_pred in await chat_completion:
            yield new_pred
