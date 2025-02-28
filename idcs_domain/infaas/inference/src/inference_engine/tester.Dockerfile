# HEALTHCHECK is intentionally omitted because health checks are managed by Kubernetes probes.
# A tester image --> NOT PRODUCTION

FROM internal-placeholder.com/cache/library/python:3.12-bookworm
WORKDIR /app

RUN pip install --upgrade pip setuptools
RUN python -m venv /app/venv

COPY requirements/requirements-tester.txt /app/requirements-tester.txt
RUN /app/venv/bin/pip install --no-cache-dir -r /app/requirements-tester.txt

# Create a directory for benchmark results
RUN mkdir -p /app/bench_results

COPY tests/grpc_proxy.py /app/tests/grpc_proxy.py
COPY utils /app/utils/
COPY inference_engine /app/inference_engine
COPY test_runner.py /app/test_runner.py

ENTRYPOINT ["/app/venv/bin/python", "test_runner.py"]