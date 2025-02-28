#!/usr/bin/env bash
#
# Start Docker container registry.
#
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
LOCAL_REGISTRY_PORT=${LOCAL_REGISTRY_PORT:-5001}
LOCAL_REGISTRY_PORT_HTTPS=${LOCAL_REGISTRY_PORT_HTTPS:-5443}
LOCAL_REGISTRY_NAME=${LOCAL_REGISTRY_NAME:-idc-registry-${LOCAL_REGISTRY_PORT}.intel.com}

create_docker_network() {
    docker network create kind || true
}

start_registry() {
    docker rm --force --volumes "${LOCAL_REGISTRY_NAME}" || true
    docker run \
        -d \
        --restart=always \
        -p "127.0.0.1:${LOCAL_REGISTRY_PORT}:5000" \
        --name "${LOCAL_REGISTRY_NAME}" \
        --read-only \
        registry:2
    docker network connect --alias local-registry kind ${LOCAL_REGISTRY_NAME}
}

# Create a self-signed TLS certificate.
create_tls_cert() {
    mkdir -p local/secrets/local-docker-registry
    openssl req -x509 -nodes --days 99999 -newkey rsa:4096 \
        -subj "/CN=${LOCAL_REGISTRY_NAME}" \
        -addext "subjectAltName = DNS:${LOCAL_REGISTRY_NAME}" \
        -keyout local/secrets/local-docker-registry/cert.key \
        -out local/secrets/local-docker-registry/cert.crt
}

# Start NGINX reverse proxy that listens on https and forwards to the http registry service.
start_nginx() {
    local nginx_container_name=nginx.${LOCAL_REGISTRY_NAME}
    docker rm --force --volumes "${nginx_container_name}" || true
    docker run \
        -d \
        --restart=always \
        -p "127.0.0.1:${LOCAL_REGISTRY_PORT_HTTPS}:443" \
        -v ${SCRIPT_DIR}/nginx.conf:/etc/nginx/nginx.conf:ro \
        -v $(pwd)/local/secrets/local-docker-registry:/tls:ro \
        --name "${nginx_container_name}" \
        nginx@sha256:f618a6de3e2c6464699f7f0cddeb5aff68534932cefad83e9c225b0db4024a03    
    docker network connect --alias nginx-local-registry kind ${nginx_container_name}
}

main() {
    create_docker_network
    start_registry
    create_tls_cert
    start_nginx
}

main
