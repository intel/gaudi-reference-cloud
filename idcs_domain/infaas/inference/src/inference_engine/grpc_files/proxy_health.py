from src.inference_engine.grpc_files.text_generator import TextGenerator
from src.utils.threading.workers.liveness import InferenceEngineLivenessWorker

from grpc_health.v1 import health, health_pb2
from enum import Enum
import logging
import grpc


class ServiceType(Enum):
    LIVENESS = "liveness"
    STARTUP = "startup"
    READINESS = "readiness"


class ProxyHealthService(health.aio.HealthServicer):

    def __init__(
        self,
        main_servicer: TextGenerator
    ) -> None:

        super().__init__()
        self._main_servicer = main_servicer
        self._liveness_worker = InferenceEngineLivenessWorker(prob_data=main_servicer.server_prob_config)
        self._liveness_worker.run()

    async def Check(
        self,
        request: health_pb2.HealthCheckRequest,
        context: grpc.RpcContext
    ) -> health_pb2.HealthCheckResponse:

        try:
            service_type = ServiceType(request.service)
        except:
            logging.warning("Unknown probe method was detected")
            unknown = health_pb2.HealthCheckResponse.UNKNOWN
            return health_pb2.HealthCheckResponse(status=unknown) 

        if service_type is ServiceType.LIVENESS:
            serving = health_pb2.HealthCheckResponse.SERVING
            return health_pb2.HealthCheckResponse(status=serving)

        if service_type is ServiceType.STARTUP:

            return self._get_serving_status(
                is_serving=self._main_servicer.is_ready
            )

        if service_type is ServiceType.READINESS:

            tgi_server_is_alive_and_ready = (
                self._liveness_worker.is_working and
                self._liveness_worker.server_status_code == 200
            )

            return self._get_serving_status(
                is_serving=tgi_server_is_alive_and_ready
            )

        else:
            logging.warning("Unknown probe method was detected")
            unknown = health_pb2.HealthCheckResponse.UNKNOWN
            return health_pb2.HealthCheckResponse(status=unknown)

    async def enter_graceful_shutdown(self) -> None:
        self._liveness_worker.stop()
        return await super().enter_graceful_shutdown()

    def _get_serving_status(
        self,
        is_serving: bool
    ) -> health_pb2.HealthCheckResponse:

        if is_serving:
            serving = health_pb2.HealthCheckResponse.SERVING
            return health_pb2.HealthCheckResponse(status=serving)

        not_serving = health_pb2.HealthCheckResponse.NOT_SERVING
        return health_pb2.HealthCheckResponse(status=not_serving)
