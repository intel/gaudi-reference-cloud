from src.utils.errors.abstract import MaasErrorWithoutStacktrace


class SafeguardNotResponsiveError(MaasErrorWithoutStacktrace):

    def __init__(self, prompt: str):

        log_err_msg =\
            "Safeguard failed to provide classification befor LLM was done." +\
            f"\nPrompt: {prompt}"

        super().__init__(
            log_err_msg=log_err_msg,
            external_err_msg=SafeguardNotResponsiveError.__name__
        )


class SafeguardResponseParseError(MaasErrorWithoutStacktrace):

    def __init__(self, safeguard_response: str):

        log_err_msg = "Failed to parse the following safeguard response: " +\
            safeguard_response

        super().__init__(
            log_err_msg=log_err_msg,
            external_err_msg=SafeguardResponseParseError.__name__
        )


class SafeguardUnrecognizedCategoryError(MaasErrorWithoutStacktrace):

    def __init__(self, harm_category: str):

        log_err_msg = "Safeguard returned unrecognized 'harm category': " +\
            harm_category

        super().__init__(
            log_err_msg=log_err_msg,
            external_err_msg=SafeguardUnrecognizedCategoryError.__name__
        )
