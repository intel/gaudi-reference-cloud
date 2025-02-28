from src.utils.errors.probing import ProxyInternalProbingError
from src.utils.entities.server_prob_data import ServerProbData

from aiohttp import ClientSession, ClientTimeout
from typing import Optional, Dict, Any, Coroutine
from http import HTTPStatus

import asyncio
import logging


class InferenceEngineLivenessWorker:
    """
    Worker to run internal liveness check from the proxy.
    """

    def __init__(
        self,
        prob_data: ServerProbData,
        sleep: Optional[int] = 10
    ) -> None:

        self._prob_data = prob_data
        self._sleep = sleep
        self._is_working = False
        self._server_status_code = HTTPStatus.SERVICE_UNAVAILABLE.value

    @property
    def is_working(self) -> bool:
        return self._is_working

    @property
    def server_status_code(self) -> int:
        return self._server_status_code

    def run(self) -> None:
        self._is_working = True
        logging.info(f"{InferenceEngineLivenessWorker.__name__} is working...")
        logging.info(f"Probing Address: {self._prob_data.url}")
        logging.info(f"Probing Sleep: {self._sleep}[s]")
        asyncio.create_task(self._run_async())

    def stop(self):
        self._is_working = False

    async def _run_async(self) -> None:
        # startup probe - probe every 1s until first '200 - OK' response
        logging.info("Strating startup probing task...")
        await self._probe_server(
            sleep=1,
            execution_condition=self._execute_if_not_ready
        )

        # readiness probe - probe every given 'sleep' value as long as
        # 'self._is_working' is True
        logging.info(
            "Startup probing task completed. " +
            "Strating readiness probing task..."
        )
        await self._probe_server(
            sleep=self._sleep,
            execution_condition=self._execute_if_working
        )

    async def _probe_server(
        self, 
        sleep: int,
        execution_condition: Coroutine
    ) -> None:
        try:
            timeout = ClientTimeout(total=None)
            async with ClientSession(timeout=timeout) as session:
                while await execution_condition():
                    await self._post_payload(session)
                    await asyncio.sleep(sleep)

        # general exception - stops the worker
        except Exception as e:
            self._is_working = False
            raise ProxyInternalProbingError(e)

    async def _post_payload(
        self, 
        session: ClientSession
    ) -> None:
        try:
            async with session.post(
                self._prob_data.url,
                json=self._prob_data.payload
            ) as response:
                self._server_status_code = response.status

        # single request exception - doesn't stop the worker
        except Exception:
            self._server_status_code =\
                HTTPStatus.SERVICE_UNAVAILABLE.value

    async def _execute_if_not_ready(self) -> bool:
        return self._server_status_code != HTTPStatus.OK.value

    async def _execute_if_working(self) -> bool:
        return self._is_working
