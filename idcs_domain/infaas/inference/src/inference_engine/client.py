import src.inference_engine.grpc_files.infaas_generate_pb2_grpc as pb2_grpc
import src.inference_engine.grpc_files.infaas_generate_pb2 as pb2

from grpc_health.v1 import health_pb2_grpc, health_pb2
from typing import Optional, List, Dict, Any

import hashlib
import logging
import grpc
import os


class GrpcProxyClient:

    @staticmethod
    async def chat_completion(
        messages: List[pb2.ChatCompletionMessage],
        params: Dict[str, Any],
        server_address: Optional[str] = "localhost:50051",
        print_full_response_object: Optional[bool] = False
    ) -> None:

        async with grpc.aio.insecure_channel(server_address) as channel:

            stub = pb2_grpc.TextGeneratorStub(channel)
            req = await GrpcProxyClient._get_request(
                messages=messages, params=params
            )
            full_answer = ""

            async for new_pred in stub.ChatCompletionStream(request=req):
                if print_full_response_object:
                    print(new_pred)
                else:
                    new_token = new_pred.choices[0].delta.content
                    print(new_token, end="", flush=True)
                    full_answer += new_token

            logging.debug(
                "Request handling is done.\n" +
                f"Messages: {req.messages}\nAnswer: {full_answer}"
            )

    @staticmethod
    async def generate(
        prompt: str,
        params: pb2.GenerateRequestParameters,
        server_address: Optional[str] = "localhost:50051",
        print_full_response_object: Optional[bool] = False
    ) -> None:

        async with grpc.aio.insecure_channel(server_address) as channel:

            stub = pb2_grpc.TextGeneratorStub(channel)
            req = await GrpcProxyClient._get_request(
                prompt=prompt, params=params
            )
            full_answer = ""

            async for new_pred in stub.GenerateStream(request=req):
                if print_full_response_object:
                    print(new_pred)
                else:
                    print(new_pred.token.text, end="", flush=True)
                    full_answer += new_pred.token.text

            logging.debug(
                "Request handling is done.\n" +
                f"Prompt: {req.prompt}\nAnswer: {full_answer}"
            )

    @staticmethod
    async def _get_request(
        params: pb2.GenerateRequestParameters | Dict[str, Any],
        messages: List[pb2.ChatCompletionMessage] = None,
        prompt: Optional[str] = None,
        id_size: int = 16
    ) -> pb2.GenerateStreamRequest | pb2.ChatCompletionStreamResponse:

        # generating a random hash as requestID
        random_number = os.urandom(id_size)
        hash_object = hashlib.sha256(random_number)
        req_id = f"TEST-{hash_object.hexdigest()}"

        if (
            isinstance(params, pb2.GenerateRequestParameters) and
            prompt is not None
        ):
            return pb2.GenerateStreamRequest(
                requestID=req_id,
                prompt=prompt,
                params=params
            )

        if messages is not None:
            return pb2.ChatCompletionStreamRequest(
                requestID=req_id,
                messages=messages,
                **params
            )

        raise Exception(f"Params type `{params.__class__}` is not recognized")


class GrpcProxyHealthClient:

    @staticmethod
    async def run_check(
        service_name: str,
        server_address: Optional[str] = "localhost:50051"
    ) -> health_pb2.HealthCheckResponse:

        async with grpc.aio.insecure_channel(server_address) as channel:

            stub = health_pb2_grpc.HealthStub(channel)
            health_request = health_pb2.HealthCheckRequest(
                service=service_name
            )

            response = await stub.Check(health_request)
            return response.status
