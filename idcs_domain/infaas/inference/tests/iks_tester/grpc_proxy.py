from src.inference_engine.grpc_files.proxy_health import health_pb2
from src.inference_engine.client import (
    GrpcProxyClient,
    GrpcProxyHealthClient
)
from src.inference_engine.server import GrpcProxyServer
from typing import Dict, Optional, List

import src.inference_engine.grpc_files.infaas_generate_pb2 as pb2
import logging
import time


class GrpcProxyTester:

    _DEFAULT_MODEL_ID = "meta-llama/Llama-3.2-1B-Instruct"
    _CNVRG_TGI_SERVER_URL =\
        "https://tgi.acqmuxpkmadjvmazwq6yd2j.cloud.cnvrg.io"

    @staticmethod
    async def test_chat_completion(
        messages: List[Dict[str, str]],
        warmup_sleep: Optional[int] = None,
        print_full_response_object: Optional[bool] = False,
        **params
    ) -> None:

        if warmup_sleep is not None:
            logging.info(f"Starting a {warmup_sleep}[s] warmup")
            time.sleep(warmup_sleep)

        else:
            # applying minimal sleep for server loading
            time.sleep(1)

        logging.info("=" * 40 + " Starting `ChatCompletion` Test " + "=" * 40)
        await GrpcProxyClient.chat_completion(
            print_full_response_object=print_full_response_object,
            messages=[
                pb2.ChatCompletionMessage(
                    role=message["role"],
                    content=message["content"]
                ) for message in messages
            ],
            params=params
        )
        logging.info("=" * 40 + " Test is done " + "=" * 40)

    @staticmethod
    async def test_health_probes(
        is_server_connected: bool = False
    ) -> None:

        service_names = ["startup", "liveness", "readiness", "test"]
        results = {}

        for i, service_name in enumerate(service_names, 1):

            status = await GrpcProxyHealthClient.run_check(service_name)
            result_entry = {service_name: status}
            print(f"{i}/{len(service_names)}) status: {status}")
            results.update(result_entry)

        unknown_status = health_pb2.HealthCheckResponse.UNKNOWN
        serving_status = health_pb2.HealthCheckResponse.SERVING
        not_serving_status = health_pb2.HealthCheckResponse.NOT_SERVING

        assert results["test"] == unknown_status
        assert results["startup"] == serving_status
        assert results["liveness"] == serving_status

        if is_server_connected:
            assert results["readiness"] == serving_status
        else:
            assert results["readiness"] == not_serving_status

    @staticmethod
    async def test_generate_stream(
        prompt: str,
        warmup_sleep: Optional[int] = None,
        params: Optional[pb2.GenerateRequestParameters] = None,
        print_full_response_object: Optional[bool] = False,
    ) -> None:

        if warmup_sleep is not None:
            logging.info(f"Starting a {warmup_sleep}[s] warmup")
            time.sleep(warmup_sleep)

        if params is None:
            params = pb2.GenerateRequestParameters(
                do_sample=False,
                max_new_tokens=256
            )

        else:
            # applying minimal sleep for server loading
            time.sleep(1)

        logging.info("=" * 40 + " Starting `GenerateStream` Test " + "=" * 40)
        await GrpcProxyClient.generate(
            print_full_response_object=print_full_response_object,
            prompt=prompt,
            params=params
        )
        logging.info("=" * 40 + " Test is done " + "=" * 40)

    @staticmethod
    async def test_server_serve(
        model_id: str = _DEFAULT_MODEL_ID,
        base_url: Optional[str] = _CNVRG_TGI_SERVER_URL,
        hf_api_key: str | None = None,
        safeguard_url: str | None = None,
        headers: Optional[Dict[str, str] | None] = None,
        cookies: Optional[Dict[str, str] | None] = None,
        timeout: Optional[int | None] = None,
        backend: Optional[str] = None
    ) -> None:

        server = GrpcProxyServer(
            model_id=model_id,
            base_url=base_url,
            hf_api_key=hf_api_key,
            safeguard_url=safeguard_url,
            headers=headers,
            cookies=cookies,
            timeout=timeout,
            backend=backend
        )

        await server.serve()
