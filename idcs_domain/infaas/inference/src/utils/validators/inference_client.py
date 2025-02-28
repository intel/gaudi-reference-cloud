from src.utils.errors.validation import ValidationError

from typing import Dict
from enum import StrEnum


class BackendOptions(StrEnum):
    TGI = "tgi"
    VLLM = "vllm"


class InferenceClientArgsValidator:
    """
    Validates the client arguments.
    If input arg is valid - do nothing.
    If input arg is not valid - throws invalid argument exception.
    """

    @staticmethod
    def validate_base_url(base_url: str, backend: str) -> None:
        reason = ""

        if backend == BackendOptions.TGI and base_url.endswith("/v1"):
            reason = "When using TGI, base url shouldn't include '/v1' suffix"
        elif backend == BackendOptions.VLLM and not base_url.endswith("/v1"):
            reason = "When using OpenAI, base url should include '/v1' suffix"

        if reason:
            raise ValidationError(
                validator=InferenceClientArgsValidator.__name__,
                argument="base_url",
                reason=reason
            )

    @staticmethod
    def validate_headers(headers: Dict[str, str] | None) -> None:
        InferenceClientArgsValidator._validate_dict_of_strings(
            input_dict=headers,
            arg_name="headers"
        )

    @staticmethod
    def validate_cookies(cookies: Dict[str, str] | None) -> None:
        InferenceClientArgsValidator._validate_dict_of_strings(
            input_dict=cookies,
            arg_name="cookies"
        )

    @staticmethod
    def validate_timeout(timeout: int) -> None:
        # timeout is optional - hence None is valid
        if timeout is None:
            return

        is_timeout_not_int = not isinstance(timeout, int)
        is_timeout_negative_or_zero = timeout <= 0

        if is_timeout_not_int or is_timeout_negative_or_zero:

            reason =\
                f"argument 'timeout' (value = {timeout})" +\
                "value must be a positive integer."

            raise ValidationError(
                validator=InferenceClientArgsValidator.__name__,
                argument="timeout",
                reason=reason
            )

    @staticmethod
    def validate_backend(backend: str) -> None:
        if backend not in BackendOptions:

            reason = \
                f"argument 'backend' (value = {backend})" + \
                "is not supported."

            raise ValidationError(
                validator=InferenceClientArgsValidator.__name__,
                argument="backend",
                reason=reason
            )


    @staticmethod
    def _validate_dict_of_strings(
        input_dict: Dict[str, str] | None,
        arg_name: str
    ) -> None:

        # input_dict is optional - hence None is valid
        if input_dict is None:
            return

        if not isinstance(input_dict, dict):

            reason = f"argument '{arg_name}' is not of type 'dict'."

            raise ValidationError(
                validator=InferenceClientArgsValidator.__name__,
                argument=arg_name,
                reason=reason
            )

        header_values_are_strings = [
            isinstance(v, str)
            for v in input_dict.values()
        ]

        if not all(header_values_are_strings):

            reason = f"argument '{arg_name}' has a non string value."

            raise ValidationError(
                validator=InferenceClientArgsValidator.__name__,
                argument=arg_name,
                reason=reason
            )
