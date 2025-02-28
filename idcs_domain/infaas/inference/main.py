from src.utils.validators.inference_client import(
    InferenceClientArgsValidator, BackendOptions
)
from src.inference_engine.server import GrpcProxyServer

import traceback
import logging
import asyncio
import json
import yaml
import sys
import os


class _GrpcProxyServerEnvArgs:

    def __init__(self) -> None:

        server_config_path = os.environ.get("server_config")
        if server_config_path is None:
            server_config_path = "configs/server_dev.yaml"
            logging.warning(
                "External server config was not found." \
                + f" Loading dev config: {server_config_path}"
            )

        with open(server_config_path, "r") as f:
            self.server_config = yaml.safe_load(f)

        show_full_stacktrace_str =\
            os.environ.get("show_full_stacktrace", default="True")
        self.show_full_stacktrace = bool(show_full_stacktrace_str)

        self.base_url = os.environ.get("base_url", default="http://localhost:8080")
        self.safeguard_url = os.environ.get("safeguard_url", default="http://localhost:9000")
        self.headers = os.environ.get("headers")
        self.cookies = os.environ.get("cookies")

        timeout_str = os.environ.get("timeout")
        self.timeout = int(timeout_str) if timeout_str is not None else 60
        self.backend = os.environ.get("backend", default=BackendOptions.TGI)

        # validator will raise a 'ValidationError' if field is not valid
        InferenceClientArgsValidator.validate_backend(self.backend)
        InferenceClientArgsValidator.validate_base_url(self.base_url, self.backend)
        InferenceClientArgsValidator.validate_base_url(self.safeguard_url, self.backend)
        InferenceClientArgsValidator.validate_headers(self.headers)
        InferenceClientArgsValidator.validate_cookies(self.cookies)
        InferenceClientArgsValidator.validate_timeout(self.timeout)

        # report the loaded config
        logging.info(
            "Environment Variables:\n" +
            json.dumps(vars(self), indent=4)
        )


if __name__ == "__main__":

    server = None
    env_args = None

    try:
        env_args = _GrpcProxyServerEnvArgs()
        server = GrpcProxyServer(
            config=env_args.server_config,
            safeguard_url=env_args.safeguard_url,
            base_url=env_args.base_url,
            headers=env_args.headers,
            cookies=env_args.cookies,
            timeout=env_args.timeout,
            backend=env_args.backend
        )
        asyncio.run(server.serve())

    except KeyboardInterrupt:
        if server is not None:
            asyncio.run(server.stop_server_gracefully())
        sys.exit(130)

    except Exception as e:
        err_msg = f"Inference-proxy terminated with an unexpected exception: {e}"

        if env_args is not None and env_args.show_full_stacktrace:
            err_msg += f"\nStacktrace:\n{traceback.format_exc()}"

        if server is not None:
            asyncio.run(server.stop_server_gracefully())

        logging.error(err_msg)
