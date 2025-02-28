from src.utils.errors.safeguard import (
    SafeguardResponseParseError,
    SafeguardUnrecognizedCategoryError
)
from src.utils.entities.server_prob_data import ServerProbData

from abc import ABC, abstractmethod
from typing import (
    Dict, Any, Tuple, Optional, AsyncIterator, List
)
from enum import Enum
import logging
import re


class PromptSafetyStatus(Enum):
    SAFE = "safe"
    UNSAFE = "unsafe"


class AbstractWrapper(ABC):

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
        Initiates and validates args.

        Args:
            config (Dict[str, Any]):
                config values
            base_url (str):
                OpenAI server's base url.
            model_id (str):
                the model ID
            headers (Dict[str, str], optional):
                default headers to the server requests.
            cookies (Dict[str, str], optional):
                cookies info to add with the server requests.
            timeout (int, optional):
                request timeout in seconds.
        """
        self._config = config
        self._base_url = base_url
        self._client = None

    @abstractmethod
    def generate_stream(
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
    ) -> AsyncIterator[Any]:
        """
        Main generation function that wraps the async clinet's
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
        pass  # Abstract method, no implementation here

    @abstractmethod
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
    ) -> AsyncIterator[Any]:
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
        pass  # Abstract method, no implementation here

    class _NumericParams:
        """
        A validation calss for numeric parameters.
        """

        def __init__(
            self,
            requestID: str,
            config: Dict[str, Any],
            max_new_tokens: int | None = None,
            repetition_penalty: float | None = None,
            frequency_penalty: float = None,
            seed: int | None = None,
            temperature: float | None = None,
            top_k: int | None = None,
            top_p: float | None = None,
            truncate: int | None = None,
            typical_p: float | None = None,
            top_n_tokens: int | None = None,
            top_logprobs: Optional[int] = None,
            n: Optional[int] = None
        ) -> None:

            def is_valid_number(num: Any) -> bool:
                return num is not None and num > 0

            if is_valid_number(max_new_tokens):
                self.max_new_tokens = max_new_tokens
            else:
                self.max_new_tokens = \
                    config["inference"]["defaults"]["max_new_tokens"]

                logging.warning(
                    f"request id: {requestID} | msg: " +
                    "'max_new_tokens' is negative, zero or None, " +
                    "and will be set to default value " +
                    f"({self.max_new_tokens})."
                )

            self.repetition_penalty = (
                repetition_penalty if is_valid_number(repetition_penalty)
                else None
            )
            self.frequency_penalty = (
                frequency_penalty if is_valid_number(frequency_penalty)
                else None
            )
            self.seed = (
                seed if is_valid_number(seed)
                else None
            )

            temperature = (
                temperature if is_valid_number(temperature)
                else None
            )
            if temperature is not None:
                self.temperature = temperature
            else:
                self.temperature = config["inference"]["defaults"]["temperature"]

            self.n = (
                n if is_valid_number(n)
                else 1
            )
            self.top_k = (
                top_k if is_valid_number(top_k)
                else None
            )
            self.top_p = (
                top_p if is_valid_number(top_p)
                else None
            )
            self.truncate = (
                truncate if is_valid_number(truncate)
                else None
            )
            self.typical_p = (
                typical_p if is_valid_number(typical_p)
                else None
            )
            self.top_n_tokens = (
                top_n_tokens if is_valid_number(top_n_tokens)
                else None
            )
            self._top_logprobes = (
                top_logprobs if is_valid_number(top_logprobs)
                else None
            )

    async def _check_safety_response(
            self,
            safety_response: Any,
    ) -> Tuple[bool, str | None]:
        """
        Function to set whether input prompt is safe or not.
        To be used with Safeguard.

        Args:
            safety_response (Any):
                The safety response
        Returns:
            Tuple[bool, str | None]
        """

        # response examples:
        # \n\nunsafe\nS2
        # \n\nsafe
        safety_response_text = \
            safety_response.choices[0].message.content.lower().strip()

        if "\n" in safety_response_text:
            response_components = safety_response_text.split("\n")
            safety_response_text = response_components[0]

        safety_status = PromptSafetyStatus(
            safety_response_text
        )

        if safety_status is PromptSafetyStatus.SAFE:
            return True, None

        # when response is "unsafe" it suppose to be:
        # unsafe\n<harm_code> - meaning -
        # response_components should be a valid list
        if safety_status is PromptSafetyStatus.UNSAFE:

            match = re.search(
                pattern=r"(S\d{1,2})",
                string=response_components[-1].upper().strip()
            )

            if match:
                harm_code = match.group(1)
                description = await self._map_unsafe_code(harm_code)
                return False, description

        raise SafeguardResponseParseError(safety_response_text)

    def _get_prob_data(self, prob_endpoint: str, model_id: str | None) -> ServerProbData:
        url = f"{self._base_url}/{prob_endpoint}"

        payload = {
            "model": model_id,
            "messages": [{"role": "user", "content": "test"}],
            "temperature": 0.01,
            "max_tokens": 4,
            "stream": False
        }

        return ServerProbData(url=url, payload=payload)

    async def _map_unsafe_code(self, harm_code: str) -> str:

        categories = self._config["safeguard"]["harm_categories"]
        description = categories.get(harm_code)
        if description is None:
            raise SafeguardUnrecognizedCategoryError(harm_code)
        return description