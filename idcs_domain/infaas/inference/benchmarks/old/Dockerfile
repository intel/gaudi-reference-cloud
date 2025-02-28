# HEALTHCHECK is intentionally omitted because health checks are managed by Kubernetes probes.

FROM internal-placeholder.com/cache/library/python:3.12-bookworm
WORKDIR /app

# Create a non-root user
RUN useradd -m appuser

RUN pip install --upgrade pip setuptools
RUN apt-get update && apt-get install -y --no-install-recommends python3-venv
RUN python -m venv /app/venv

COPY requirements/requirements-proxy.txt /app/requirements-tester.txt
RUN /app/venv/bin/pip install --no-cache-dir -r /app/requirements-tester.txt

# Create a directory for benchmark results
RUN mkdir -p /app/bench_results

# COPY tests/grpc_proxy.py /app/tests/grpc_proxy.py
COPY tests /app/tests/
COPY utils /app/utils/
COPY inference_engine /app/inference_engine
COPY test_runner.py /app/test_runner.py

# Change ownership of application files to the non-root user
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

RUN chmod +x /app/tests/bench_runner.sh
ENTRYPOINT ["/bin/bash", "/app/tests/bench_runner.sh"]