from src.utils.errors.abstract import MaasErrorWithStacktrace


class ProxyInternalProbingError(MaasErrorWithStacktrace):

    def __init__(self, external_err: Exception = None):

        log_err_msg = "Proxy got an error while probing TGI server."
        if external_err:
            log_err_msg += f"\nError message: {external_err}"

        super().__init__(
            log_err_msg=log_err_msg,
            external_err_msg=ProxyInternalProbingError.__name__
        )
