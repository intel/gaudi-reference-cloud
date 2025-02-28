# HEALTHCHECK is intentionally omitted because health checks are managed by Kubernetes probes.
# We undid EVERY user creation change - messes mounted S3 files consumption.

ARG PY_VERSION=3.12

# Build image
FROM chainguard/wolfi-base@sha256:2148be123cd047f10c93e2bc88010d4abba1fc56a367d6287a251099ed5f006a as build 
WORKDIR /build

ARG PY_VERSION 

# Install python and pip in build container
RUN apk add --no-cache python-${PY_VERSION} py${PY_VERSION}-pip

# Create venv
RUN python -m venv venv

# Copy over dependency relevant files
COPY requirements.txt /build/requirements.txt
# Install dependencies into venv
RUN venv/bin/pip install --no-cache-dir -r requirements.txt

# Now, modify the pydantic model in the `text-generation` package
RUN sed -i "s/tool_calls: Optional\[ChoiceDeltaToolCall\]/tool_calls: Optional[ChoiceDeltaToolCall] = None/" \
    /build/venv/lib/python${PY_VERSION}/site-packages/text_generation/types.py

# Runtime image
FROM chainguard/wolfi-base@sha256:2148be123cd047f10c93e2bc88010d4abba1fc56a367d6287a251099ed5f006a as final

ARG PY_VERSION

# Create a non-root user (1000:1000)
# RUN addgroup --gid 1000 appgroup
# RUN adduser -u 1000 -G appgroup --disabled-password --gecos '' appuser

# Install python in the runtime image
RUN apk add --no-cache python-${PY_VERSION} 
WORKDIR /app 

# Copy over dependencies
COPY --from=build /build/venv /app/venv 
# Copy over app sources
COPY src src/
COPY configs configs/
COPY main.py main.py

# Change ownership of the app directory to the non-root user
# RUN chown -R 1000:1000 /app

# Switch to the non-root user
# USER appuser

ENTRYPOINT ["/app/venv/bin/python", "main.py"]
