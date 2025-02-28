from src.utils.wrappers.backends.abstract import AbstractWrapper
from src.utils.validators.inference_client import BackendOptions
from src.utils.wrappers.backends.openai import OpenAiWrapper
from src.utils.wrappers.backends.tgi import TgiWrapper

from typing import Dict, Any, Tuple


class BackendFactory:

    @staticmethod
    def backend_wrapper(
        config: Dict[str, Any],
        backend: str,
        base_url: str,
        safeguard_url: str,
        headers: Dict[str, str] | None,
        cookies: Dict[str, str] | None,
        timeout: int
    ) -> Tuple[AbstractWrapper, AbstractWrapper]:

        llm_engine = None
        sg_engine = None
        safeguard_config = config["server"]["safeguard"]

        if backend == BackendOptions.TGI:
            llm_engine = TgiWrapper(
                config=config["server"],
                base_url=base_url,
                model_id="tgi",
                headers=headers,
                cookies=cookies,
                timeout=timeout
            )

            sg_engine = TgiWrapper(
                config=config["server"],
                base_url=safeguard_url,
                model_id="tgi",
                headers=safeguard_config["headers"],
                cookies=None,
                timeout=safeguard_config["timeout"]
            )

        elif backend == BackendOptions.VLLM:
            llm_engine = OpenAiWrapper(
                config=config["server"],
                base_url=base_url,
                model_id=config["server"]["model_id"],
                headers=headers,
                cookies=cookies,
                timeout=timeout
            )

            sg_engine = OpenAiWrapper(
                config=config["server"],
                base_url=safeguard_url,
                model_id=safeguard_config["model_id"],
                headers=safeguard_config["headers"],
                cookies=None,
                timeout=safeguard_config["timeout"]
            )

        return llm_engine, sg_engine
