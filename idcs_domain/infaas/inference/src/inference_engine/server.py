from src.inference_engine.grpc_files.text_generator import TextGenerator
from src.inference_engine.grpc_files.proxy_health import ProxyHealthService
import src.inference_engine.grpc_files.infaas_generate_pb2_grpc as\
    generate_pb2_grpc
import src.inference_engine.grpc_files.infaas_generate_pb2 as generate_pb2

from grpc_reflection.v1alpha import reflection
from grpc_health.v1 import health_pb2_grpc
from typing import Optional, Dict, Tuple, Any
import logging
import grpc


class GrpcProxyServer:
    """
    Python gRPC proxy server for HuggingFace's TGI server.
    This class uses the gRPC 'TextGenerator' class,
    that handles RPC errors internally.

    This class DOESN'T catch unexpected exceptions,
    it is suppose to be triggered by external file (proxy_runner.py)
    that handles both KeyboardImterrupt and general Exception.

    This class exposes a gracefull shutdown to close the sever on
    an unexpected exception.
    """

    def __init__(
        self,
        config: Dict[str, Any],
        base_url: str,
        safeguard_url: str,
        headers: Optional[Dict[str, str]] | None = None,
        cookies: Optional[Dict[str, str]] | None = None,
        timeout: Optional[int] = None,
        backend: Optional[str] = None
    ) -> None:

        self._server = None
        self._health_servicer = None

        self._safeguard_url = safeguard_url
        self._base_url = base_url
        self._cookies = cookies
        self._timeout = timeout
        self._headers = headers
        self._backend = backend
        self._config = config

        # if user passed Content-Type - keep it
        # else: add "Content-Type": "application/json" header
        if self._headers is None:
            self._headers = {"Content-Type": "application/json"}
        else:
            is_content_type_header = [
                k.lower() == "content-type"
                for k in self._headers.keys()
            ]
            if not any(is_content_type_header):
                self._headers["Content-Type"] = "application/json"

    async def _configure_servicers(
        self
    ) -> Tuple[grpc.aio.Server, ProxyHealthService]:

        server = grpc.aio.server()
        main_servicer = TextGenerator(
            config=self._config,
            safeguard_url=self._safeguard_url,
            base_url=self._base_url,
            headers=self._headers,
            cookies=self._cookies,
            timeout=self._timeout,
            backend=self._backend,
        )

        generate_pb2_grpc.add_TextGeneratorServicer_to_server(
            server=server,
            servicer=main_servicer
        )

        health_servicer = ProxyHealthService(
            main_servicer=main_servicer,
        )

        health_pb2_grpc.add_HealthServicer_to_server(
            server=server,
            servicer=health_servicer
        )

        service_full_name =\
            generate_pb2.DESCRIPTOR.services_by_name["TextGenerator"].full_name

        service_names = (
            service_full_name,
            reflection.SERVICE_NAME
        )
        reflection.enable_server_reflection(service_names, server)

        listen_addr = self._config["server"]["listen_address"]
        server.add_insecure_port(listen_addr)
        await server.start()

        logging.info(f"Main gRPC Server Started on {listen_addr}")
        return server, health_servicer

    async def serve(self) -> None:

        self._server, self._health_servicer =\
            await self._configure_servicers()
        await self._server.wait_for_termination()

    async def stop_server_gracefully(self) -> None:

        if self._health_servicer is not None:
            await self._health_servicer.enter_graceful_shutdown()

        if self._server is not None:
            # Running RPCs has 5 second to copmlete
            await self._server.stop(grace=5)
