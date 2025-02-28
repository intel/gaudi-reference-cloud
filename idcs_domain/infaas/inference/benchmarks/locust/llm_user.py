from locust.clients import LocustResponse
from transformers import AutoTokenizer
from collections import deque
from locust import HttpUser, task, events, LoadTestShape
from typing import Dict, Any, List, Optional

import numpy as np
import logging
import time
import sys
import os


# Shared thread-safe list
all_users = deque()

class LlmUser(HttpUser):

    def __init__(self, *args, **kwargs) -> None:

        super().__init__(*args, **kwargs)

        self.model = os.getenv('MODEL')
        self.backend = os.getenv('BACKEND')
        self.input_tokens = int(os.getenv('INPUT_TOKENS'))
        self.output_tokens = int(os.getenv('OUTPUT_TOKENS'))
        self.hf_token = os.getenv('HF_TOKEN')
        self.llm_endpoint = os.getenv('LLM_ENDPOINT')
        self.client_id = os.getenv('IDC_CLIENT_ID')
        self.client_secret = os.getenv('IDC_CLIENT_SECRET')
        self.cloud_account_id = os.getenv('IDC_CLOUD_ACCOUNT_ID')
        self.product_name = os.getenv('MAAS_PRODUCT_NAME')
        self.product_id = os.getenv('MAAS_PRODUCT_ID')
        self.users = int(os.getenv("USERS", "10"))
        self.target_rps = int(os.getenv("TARGET_RPS", "5"))
        self.headers = self.get_headers()

        self.tokenizer = AutoTokenizer.from_pretrained(self.model)
        self.vocab = np.array(list(self.tokenizer.get_vocab().keys()))
        self.generate_inst = f"Generate a random sequence of {self.output_tokens} tokens, for example:\n"
        token_ids = self.tokenizer.encode(self.generate_inst)
        self.num_random_tokens = self.output_tokens - len(token_ids)

        logging.basicConfig(
            format="%(asctime)s %(levelname)s:%(filename)s:%(funcName)s: %(message)s",
            datefmt="%d-%M-%Y %H:%M:%S",
            level=logging.INFO,
            stream=sys.stdout
        )

        self.last_request_time = time.time()
        self.first_token_times = []
        self.response_times = []
        self.total_tokens = 0

        # tracking all users for stats calculation
        all_users.append(self)

    @task
    def query_llm(self) -> None:
        start_time = time.time()
        time_since_last_request = start_time - self.last_request_time
        
        prompt = self.generate_prompt()
        payload = self.get_request_payload(prompt)
        first_token_time = None
        token_count = 0

        with self.client.post(
            self.llm_endpoint,
            json=payload,
            headers=self.headers,
            stream=True,
            catch_response=True
        ) as response:
            
            if response.status_code == 200:
                response.success()
                token_count = 0

                for byte_payload in response.iter_lines():
                    if first_token_time is None:
                        first_token_time = self.report_ttft(start_time)

                    if self.try_parse_response(byte_payload=byte_payload):
                        token_count += 1

                response_time = self.report_request_stats(
                    first_token_time=first_token_time,
                    start_time=start_time,
                    token_count=token_count
                )

                self.handle_coordinated_omission(
                    time_since_last_request=time_since_last_request,
                    response_time=response_time
                )

            else:
                msg=f"Request failed with status {response.status_code}"
                response.failure(msg)
                self.response_times.append(time.time() - start_time)

    @events.test_stop.add_listener
    def on_test_stop(environment, **kwargs) -> None:
        all_response_times = []
        all_first_token_times = []
        all_users_tokens = 0

        for user in all_users:
            all_response_times.extend(user.response_times)
            all_first_token_times.extend(user.first_token_times)
            all_users_tokens += user.total_tokens

        if all_response_times:

            total_time = sum(all_response_times)
            throughput = all_users_tokens / total_time

            log_msg = LlmUser.get_test_stats_msg(
                all_response_times=all_response_times,
                all_first_token_times=all_first_token_times,
                total_tokens=all_users_tokens,
                total_time=total_time,
                throughput=throughput
            )

            logging.info(log_msg)

    def get_use_fast(self) -> bool:
        use_fast_dict = {"mistralai": False, "meta-llama": True}
        model_provider = self.model.split("/")[0]
        if use_fast_dict.get(model_provider) is not None:
            return use_fast_dict[model_provider]

    def get_headers(self) -> Dict[str, Any]:
        headers = {"Content-Type": "application/json"}
        if self.backend == "e2e":
            idc_access_token = self.get_e2e_token(
                self.client_id, self.client_secret
            )
            headers['Authorization'] = idc_access_token
        return headers

    def get_e2e_token(self, client_id: str, client_secret: str) -> str:
        import requests
        url = 'https://client-token.staging.api.idcservice.net/oauth2/token'
        payload = 'grant_type=client_credentials'
        headers = {'Content-Type': 'application/x-www-form-urlencoded'}
        response = requests.post(
            url=url, 
            data=payload,
            headers=headers,
            auth=(client_id, client_secret)
        )
        token_resp = response.json()
        return f"Bearer {token_resp['access_token']}"

    def generate_prompt(self) -> str:
        sampled_tokens = np.random.choice(self.vocab, self.num_random_tokens)
        random_tokens = self.tokenizer.convert_tokens_to_string(sampled_tokens.tolist())
        return f"{self.generate_inst}{random_tokens}"

    def get_request_payload(self, prompt: str) -> None:

        if self.backend == "tgi":
            if self.llm_endpoint.endswith("chat/completions"):
                return self.get_chat_completion_payload(prompt)
            else:
                raise ValueError("Unrecognized endpoint")

        elif self.backend == "e2e":
            if self.llm_endpoint.endswith("generatestream"):
                return self.get_generatestream_payload(prompt)
            else:
                raise ValueError("Unrecognized endpoint")

        else:
            raise ValueError("Unrecognized backend")

    def get_chat_completion_payload(self, prompt: str) -> Dict[str, Any]:
        return {
            "model": self.model,
            "messages": [{"role": "user", "content": prompt}],
            "temperature": 0.1,
            "max_tokens": self.output_tokens,
            "stream": True
        }

    def get_generatestream_payload(self, prompt: str) -> Dict[str, Any]:
        return {
            "model": self.model,
            "request": {
                "prompt": prompt,
                "params": {
                    "max_new_tokens": self.output_tokens
                }
            },
            "cloudAccountId": self.cloud_account_id,
            "productName": self.product_name,
            "productId": self.product_id
        }

    def report_ttft(self, start_time: float) -> float:
        first_token_time = time.time() - start_time
        self.environment.events.request.fire(
            request_type="POST",
            name=f"{self.llm_endpoint}_first_token",
            response_time=self.convert_to_ms(first_token_time),
            response_length=0
        )
        return first_token_time
    
    def report_request_stats(
        self,
        first_token_time:float,
        start_time: float,
        token_count: int
    ) -> float:

        response_time = time.time() - start_time
        self.response_times.append(response_time)
        self.total_tokens += token_count
        self.first_token_times.append(first_token_time)
        return response_time
    
    def handle_coordinated_omission(
        self, 
        time_since_last_request: float
    ) -> None:

        # 1st request
        if len(self.response_times) == 0:
            return

        # calculate a threshold for time_since_last_request
        # based on 90th percentile
        p90_response_time = np.percentile(self.response_times, 90)
        p90_threshold = 2 * p90_response_time
        if time_since_last_request <= p90_threshold:
            return

        # calculate missed requests
        expected_requests = time_since_last_request / p90_response_time
        missed_requests = int(expected_requests - 1)

        if missed_requests <= 0:
            return

        # use normal distribution based on recorded response times
        mean_time = np.mean(self.response_times)
        std_dev = np.std(self.response_times)
        
        for _ in range(missed_requests):
            simulated_time = max(0, np.random.normal(mean_time, std_dev))
            self.response_times.append(simulated_time)
            self.environment.events.request.fire(
                request_type="POST",
                name=f"{self.backend}_coordinated_omission",
                response_time=simulated_time,
                response_length=0,
                context={
                    "missed_request": True,
                    "simulated_response_time": simulated_time,
                    "time_since_last": time_since_last_request
                }
            )

    def convert_to_ms(self, seconds: float) -> float:
        return seconds * 1000
    
    def try_parse_response(
        self, 
        byte_payload: Dict[str, str] | None
    ) -> bool:
        """Response can have 'enpty lines' and other noise we don't want to count"""
        if (
            byte_payload is None or 
            byte_payload == b"\n"
        ):
            return False

        payload = byte_payload.decode("utf-8")

        if self.backend == "tgi":
            # TGI response looks like the following:
            # data:{...}
            return payload.startswith("data:")

        elif self.backend == "e2e":
            # E2E response looks like the following:
            # {"result":{...}
            return payload.startswith("result:")

        else:
            raise ValueError("Invalid 'backend'")

    @staticmethod
    def get_percentiles(
        data: List[float],
        percentiles: List[float]
    ) -> Dict[str, Any]:

        if data is None:
            return {p: 0.0 for p in percentiles}
        return dict(zip(percentiles, np.percentile(data, percentiles)))
    
    @staticmethod
    def get_test_stats_msg(
        all_response_times: List[float],
        all_first_token_times: List[float],
        total_tokens: int,
        total_time: float,
        throughput: float
    ) -> str:

        percentiles = [90, 99, 99.9]
        response_percentiles = LlmUser.get_percentiles(
            data=all_response_times, percentiles=percentiles
        )
        ttft_percentiles = LlmUser.get_percentiles(
            data=all_first_token_times, percentiles=percentiles
        )

        return "\nResponse Time Statistics:" \
            + f"\nAverage: {np.mean(all_response_times):.2f}s" \
            + f"\nMedian: {np.median(all_response_times):.2f}s" \
            + f"\nP90: {response_percentiles[90]:.2f}s" \
            + f"\nP99: {response_percentiles[99]:.2f}s" \
            + f"\nP99.9: {response_percentiles[99.9]:.2f}s" \
            + "\nFirst Token Statistics:" \
            + f"\nAverage: {np.mean(all_first_token_times):.2f}s" \
            + f"\nMedian: {np.median(all_first_token_times):.2f}s" \
            + f"\nP90: {ttft_percentiles[90]:.2f}s" \
            + f"\nP99: {ttft_percentiles[99]:.2f}s" \
            + f"\nP99.9: {ttft_percentiles[99.9]:.2f}s" \
            + "\nToken Throughput Test Results:" \
            + f"\nTotal Tokens Processed: {total_tokens}" \
            + f"\nTotal Time: {total_time:.2f}s" \
            + f"\nToken Throughput (Tokens per Second): {throughput:.2f}"

class FixedArrivalRate(LoadTestShape):
    """
    Ensures a steady request arrival rate instead of per-user requests.
    Dynamically reads `TARGET_RPS` from environment.
    """
    def __init__(self):
        super().__init__()
        self.users = int(os.getenv("USERS", "10"))
        self.target_rps = int(os.getenv("TARGET_RPS", "5"))
    
    def tick(self):
        return (self.users, self.target_rps)
