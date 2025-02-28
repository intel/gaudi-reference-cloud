from src.utils.errors.abstract import MaasErrorWithoutStacktrace


class ValidationError(MaasErrorWithoutStacktrace):

    def __init__(self, validator: str, argument: str, reason: str = None):

        log_err_msg = f"Validator '{validator}' failed on " +\
            f"argument '{argument}'."
        if reason is not None and reason != "":
            log_err_msg += f"\nFailure reason: {reason}"

        super().__init__(
            log_err_msg=log_err_msg,
            external_err_msg=reason
        )
