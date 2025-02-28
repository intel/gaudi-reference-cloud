from abc import ABC


class AbstractMaasError(Exception, ABC):
    """
    The basic format of MaaS custom errors.

    Args:
        log_err_msg (str):
            Log message for internal logging.
            Includes full stacktace (when needed).
        external_err_msg (str):
            External message to communicate with other services.
    """
    def __init__(
        self,
        log_err_msg: str,
        external_err_msg: str
    ) -> None:

        self._log_err_msg = log_err_msg
        self._external_err_msg = external_err_msg
        super().__init__(external_err_msg)

    @property
    def log_err_msg(self) -> str:
        return self._log_err_msg

    @property
    def external_err_msg(self) -> str:
        return self._external_err_msg


class MaasErrorWithStacktrace(AbstractMaasError):
    pass


class MaasErrorWithoutStacktrace(AbstractMaasError):
    pass
