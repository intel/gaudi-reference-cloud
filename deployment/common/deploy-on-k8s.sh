#!/usr/bin/env bash
#
set -e
cd "$(dirname "$0")/../.."

apiserver_address=${KIND_API_SERVER_ADDRESS:-127.0.0.1}
PORT_FILE_DIRECTORY=${PORT_FILE_DIRECTORY:-local}

# Use Harbor caching proxy by default.
# Caller can set DOCKERIO_IMAGE_PREFIX=docker.io to not use Harbor caching proxy.
DOCKERIO_IMAGE_PREFIX=${DOCKERIO_IMAGE_PREFIX:-internal-placeholder.com/cache/}

# Start Docker container registry.
start_registry() {
    reg_port=${KIND_REGISTRY_PORT:-5001}
    reg_name="idc-registry-${reg_port}.$(id -un).intel.com"
    docker rm -f "${reg_name}" || true
    docker run \
        -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
        --read-only \
        registry:2
}

upload_local_docker_images() {
  for filename in deployment/docker-images/*.tar; do
     response=$(docker load -i $filename)
     docker push "${response/Loaded image: /}"
  done
}

write_config_singlecluster() {
    mkdir -p ${SECRETS_DIR}/kubeconfig
    KUBECONFIG=${KUBECONFIG:-${HOME}/.kube/config}
    cp -v ${KUBECONFIG} ${SECRETS_DIR}/kubeconfig/kind-idc-us-dev-1a.yaml
    cp -v ${KUBECONFIG} ${SECRETS_DIR}/kubeconfig/kind-idc-us-dev-1a-network.yaml

    echo -n 80 > ${PORT_FILE_DIRECTORY}/global_host_port_80
    echo -n 80 > ${PORT_FILE_DIRECTORY}/us-dev-1_host_port_80
    echo -n 80 > ${PORT_FILE_DIRECTORY}/us-dev-1a_host_port_80
    echo -n 443 > ${PORT_FILE_DIRECTORY}/global_host_port_443
    echo -n 443 > ${PORT_FILE_DIRECTORY}/us-dev-1_host_port_443
    echo -n 443 > ${PORT_FILE_DIRECTORY}/us-dev-1a_host_port_443
}

clear_local_state() {
  rm -rvf ${SECRETS_DIR}/kubeconfig
  rm -rvf ${SECRETS_DIR}/vault-jwk-validation-public-keys
  rm -vf ${PORT_FILE_DIRECTORY}/*_host_port_*
}

# Output a TCP port number. If USE_DYNAMIC_PORTS=true, a free dynamic port will be allocated.
# Otherwise, return the provided fixed port.
# The port number is saved to a file for use by automated tests.
fixed_or_dynamic_port() {
  local port_file=$1
  local fixed_port=$2
  if [ "${USE_DYNAMIC_PORTS}" == "true" ]; then
    allocate_port ${port_file}
  else
    echo ${fixed_port} > ${port_file}
    cat ${port_file}
  fi
}

# Output a free dynamic port.
# The port number is read from the provided file (if possible) and saved to the provided file so that it doesn't change.
allocate_port() {
  local port_file=$1
  if [ ! -f ${port_file} ]; then
    deployment/common/freeport.py > ${port_file}
    echo "allocate_port: Allocated free port $(cat ${port_file}) for ${port_file}." >&2
  else
    echo "allocate_port: Using already allocated port $(cat ${port_file}) for ${port_file}." >&2
  fi
  cat ${port_file}
}
