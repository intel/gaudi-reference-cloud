import threading
import logging
import asyncio
import random
import httpx
import yaml
import time
import copy
import json
import sys
import os

from src.utils.wrappers.transformers import HfTokenizer
from src.utils.builders.prompt import GeneratePromptBuilder as PromptBuilder
from src.utils.wrappers.csv import CsvWriter
from dataclasses import dataclass, fields
from pathlib import Path
from typing import Dict, Optional, List, Any


@dataclass
class BenchResult:
    id: int
    http_code: int
    prompt: Optional[str] = None
    response_text: Optional[str] = None
    response_num_tokens: Optional[int] = None
    time_to_first_token_ms: Optional[float] = None
    total_latency_ms: Optional[float] = None

    def __post_init__(self):
        if self.http_code < 100 or self.http_code >= 600:
            raise ValueError("Invalid HTTP code")
        if (
            self.time_to_first_token_ms is not None and
            (self.time_to_first_token_ms < 0 or self.total_latency_ms < 0)
        ):
            raise ValueError("Time values must be non-negative")

    @classmethod
    def get_data_class_field_names(cls) -> List[str]:
        return [field.name for field in fields(cls)]


class BenchArguments:

    def __init__(self) -> None:

        url = os.environ.get("url")
        warmup = os.environ.get("warmup")
        headers = os.environ.get("headers")
        sleep_s = os.environ.get("sleep_s")
        model_id = os.environ.get("model_id")
        num_tests = os.environ.get("num_tests")
        hf_api_key = os.environ.get("hf_api_key")
        parameters = os.environ.get("parameters")
        output_dir = os.environ.get("output_dir")
        num_threads = os.environ.get("num_threads")
        data_file_path = os.environ.get("data_file_path")

        self._sleep_s = sleep_s if sleep_s is not None else 1.0

        if model_id is None:
            logging.error("'model_id' must be provided")
            sys.exit(1)
        self._model_id = model_id

        if hf_api_key is None:
            logging.error("'hf_api_key' must be provided")
            sys.exit(1)
        self._hf_api_key = hf_api_key

        if output_dir is None:
            logging.error("'output_dir' must be provided")
            sys.exit(1)
        self._output_dir = output_dir

        self._parameters = (
            parameters if parameters is not None
            else {"max_new_tokens": 128}
        )

        if num_tests is None:
            num_tests = 100
        else:
            num_tests = int(num_tests)
            num_tests = num_tests if num_tests > 0 else 100
        self._num_tests = num_tests

        if num_threads is None:
            self._num_threads = 1
        else:
            num_threads = int(num_threads)
            if num_threads > 0:
                self._num_threads = int(num_threads)
            else:
                logging.error("'threads' must be a positive integer")
                sys.exit(1)

        self._url = (
            url if url is not None
            else "http://localhost:8080/generate_stream"
        )

        self._headers = (
            headers if headers is not None
            else {"Content-Type": "application/json"}
        )

        self._curr_dir = os.path.dirname(__file__)
        self._data_file_path = (
            data_file_path if data_file_path is not None
            else os.path.join(self._curr_dir, "data.yaml")
        )

        if warmup is None:
            self._warmup = True
        else:
            self._warmup = warmup

    @property
    def url(self) -> str:
        return self._url

    @property
    def warmup(self) -> bool:
        return self._warmup

    @property
    def headers(self) -> Dict[str, str]:
        return self._headers

    @property
    def sleep_s(self) -> float:
        return self._sleep_s

    @property
    def model_id(self) -> str:
        return self._model_id

    @property
    def hf_api_key(self) -> str:
        return self._hf_api_key

    @property
    def parameters(self) -> Dict[str, Any]:
        return self._parameters

    @property
    def output_dir(self) -> str:
        return self._output_dir

    @property
    def curr_dir(self) -> str:
        return self._curr_dir

    @property
    def num_tests(self) -> int:
        return self._num_tests

    @property
    def num_threads(self) -> int:
        return self._num_threads

    @property
    def data_file_path(self) -> str:
        return self._data_file_path


class TGIBenchmarkManager:

    def __init__(self, args: BenchArguments) -> None:

        self._args = args
        self._tokenizer = HfTokenizer(self._args.model_id)

        with open(self._args.data_file_path, "r") as file:
            data = yaml.safe_load(file)
        self._prompts = data["prompts"]

        self._output_file_name =\
            f"tgi-benchmark-{self._args.num_threads}t-{self._args.num_tests}r"

        for i in range(self._args.num_threads):
            if i == 0:
                self._results = {"thread-0": []}
            self._results[f"thread-{i}"] = []

    async def _get_ms_time_diff(self, start_ns: float, end_ns: float) -> float:
        return round((end_ns - start_ns) / 1_000_000, 3)

    async def _send_request(
        self,
        id: int,
        client: httpx.AsyncClient,
        req_json: Dict[str, str],
        raw_prompt: str
    ) -> BenchResult:

        try:
            async with (
                client.stream("POST", self._args.url, json=req_json)
            ) as response:

                if response.status_code == 200:

                    start_ns = time.time_ns()
                    first_token_received = False
                    response_text = ""
                    response_tokens = 0

                    async for resp_json in response.aiter_lines():

                        if resp_json:

                            if not first_token_received:
                                first_token_received = True
                                time_to_first_token_ms =\
                                    await self._get_ms_time_diff(
                                        start_ns, time.time_ns()
                                    )

                            json_string = resp_json[len("data:"):]
                            parsed_resp = json.loads(json_string)
                            response_text += parsed_resp["token"]["text"]
                            response_tokens += 1

                    latency = await self._get_ms_time_diff(
                        start_ns, time.time_ns()
                    )
                    return BenchResult(
                        id=id,
                        http_code=200,
                        prompt=raw_prompt,
                        response_text=response_text,
                        response_num_tokens=response_tokens,
                        time_to_first_token_ms=time_to_first_token_ms,
                        total_latency_ms=latency
                    )

                else:
                    return BenchResult(
                        id=id,
                        http_code=response.status_code,
                        prompt=req_json["inputs"]
                    )

        except Exception:
            return BenchResult(
                id=id,
                http_code=500,
                prompt=req_json["inputs"]
            )

    async def _warmup(self) -> None:
        
        warmup_prompts = self._prompts[:20]
        async with httpx.AsyncClient() as client:

            for i, prompt in enumerate(warmup_prompts, 1):

                logging.info(
                    f"Warmup: Sending request {i}/{len(warmup_prompts)}"
                )

                formated_prompt =\
                    await PromptBuilder(self._tokenizer)\
                        .with_user_single_prompt(prompt)\
                        .build(return_messages_list=False)

                req_json = {
                    "inputs": formated_prompt,
                    "parameters": self._args.parameters
                }

                await self._send_request(
                    id=i,
                    client=client,
                    req_json=req_json,
                    raw_prompt=prompt
                )

    async def run_test(self) -> None:

        if len(self._prompts) > self._args.num_tests:
            thread_prompts = copy.deepcopy(self._prompts)
            random.shuffle(thread_prompts)
            thread_prompts = thread_prompts[:self._args.num_tests]

        thread_name = threading.current_thread().name
        bench_results = []
        async with httpx.AsyncClient() as client:

            num_requests = len(thread_prompts)
            for i, prompt in enumerate(thread_prompts, 1):

                logging.info(
                    f"{thread_name}: Sending request {i}/{num_requests}"
                )

                formated_prompt =\
                    await PromptBuilder(self._tokenizer)\
                        .with_user_single_prompt(prompt)\
                        .build(return_messages_list=False)

                req_json = {
                    "inputs": formated_prompt,
                    "parameters": self._args.parameters
                }

                bench_result = await self._send_request(
                    id=i,
                    client=client,
                    req_json=req_json,
                    raw_prompt=prompt
                )
                bench_results.append(bench_result.__dict__)

                if (
                    i < num_requests
                    and self._args.sleep_s is not None
                    and self._args.sleep_s > 0
                ):
                    logging.info(f"Waiting {self._args.sleep_s} second(s)")
                    await asyncio.sleep(self._args.sleep_s)

            self._results[thread_name].extend(bench_results)

    def _start_async_event_loop(self) -> None:

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        logging.info(f"Starting {threading.current_thread().name}")
        loop.run_until_complete(self.run_test())
        logging.info(f"Closing {threading.current_thread().name}")
        loop.close()

    def run_benchmark_threads(self) -> None:

        if self._args.warmup is True:
            asyncio.run(self._warmup())

        threads = []
        for i in range(self._args.num_threads):
            thread = threading.Thread(
                target=self._start_async_event_loop,
                name=f"thread-{i}"
            )
            threads.append(thread)
            thread.start()

        # Wait for all threads to complete
        for thread in threads:
            thread.join()

        csv_headers = ["thread"]
        csv_headers.extend(BenchResult.get_data_class_field_names())

        CsvWriter.write(
            data=self._results,
            headers=csv_headers,
            write_dir=self._args.output_dir,
            file_name=self._output_file_name
        )
