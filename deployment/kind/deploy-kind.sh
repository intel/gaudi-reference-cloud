#!/usr/bin/env bash
#
# Attention:
# It is recommended to run the following instead of using this script directly.
#   make deploy-all-in-kind
#
set -ex
cd "$(dirname "$0")/../.."

: "${SECRETS_DIR:?environment variable is required}"

HELMFILE_DUMP=${HELMFILE_DUMP:-${SECRETS_DIR}/helmfile-dump.yaml}
YQ=${YQ:-yq}
KIND=${KIND:-kind}
KUBECTL=${KUBECTL:-kubectl}
LOCAL_REGISTRY_PORT=${LOCAL_REGISTRY_PORT:-5001}
LOCAL_REGISTRY_NAME=${LOCAL_REGISTRY_NAME:-idc-registry-${LOCAL_REGISTRY_PORT}.intel.com}
CREATE_ENABLED=${CREATE_ENABLED:-true}
PORT_FILE_DIRECTORY=${PORT_FILE_DIRECTORY:-local}

# This node image supports both amd64 and arm64. See: https://github.com/kubernetes-sigs/kind/releases
kind_sum=sha256:51a1434a5397193442f0be2a297b488b6c919ce8a3931be0ce822606ea5ca245
kind_version=1.29.2

user_kubeconfig=${KUBECONFIG:-${HOME}/.kube/config}
echo New clusters will be added to KUBECONFIG=${user_kubeconfig}

query_config() {
  ${YQ} "$@" ${HELMFILE_DUMP}
}

get_regions() {
  query_config ".Values.regions[].region"  
}

get_availability_zones_for_region() {
  region=$1
  query_config ".Values.regions.${region}.availabilityZones[].availabilityZone"
}

# Delete all kind clusters with matching CLUSTER_PREFIX in the name.
# Clusters are removed from KUBECONFIG sequentially, then deleted concurrently.
delete_clusters() {
  ${KUBECTL} config get-contexts
  ${KUBECTL} config get-contexts -o=name | grep "^kind-${CLUSTER_PREFIX}-" | xargs -i -P 1 sh -c "${KUBECTL} config delete-cluster {} || true"
  ${KUBECTL} config get-contexts -o=name | grep "^kind-${CLUSTER_PREFIX}-" | xargs -i -P 1 sh -c "${KUBECTL} config delete-user {} || true"
  ${KUBECTL} config get-contexts -o=name | grep "^kind-${CLUSTER_PREFIX}-" | xargs -i -P 1 sh -c "${KUBECTL} config delete-context {} || true"
  ${KUBECTL} config get-contexts
  ${KUBECTL} config get-clusters
  ${KUBECTL} config get-users
  ${KIND} get clusters
  ${KIND} get clusters | grep "^${CLUSTER_PREFIX}-" | xargs -i -P 0 ${KIND} delete cluster --name {}
  ${KIND} get clusters
}

# Create clusters concurrently.
create_clusters() {
    default_cluster_name=$1
    cluster_names=$@
    port_offset=0
    mkdir -p ${SECRETS_DIR}/kubeconfig
    mkdir -p ${PORT_FILE_DIRECTORY}
    all_kubeconfigs=${user_kubeconfig}

    for cluster_name in $cluster_names; do
        # Append new KUBECONFIG file name to list.
        kubeconfig=${SECRETS_DIR}/kubeconfig/kind-${CLUSTER_PREFIX}-${cluster_name}.yaml
        [ "${all_kubeconfigs}" == "" ] || all_kubeconfigs=${all_kubeconfigs}:
        all_kubeconfigs=${all_kubeconfigs}${kubeconfig}
        # Create cluster in background.
        create_cluster ${cluster_name} ${port_offset} ${kubeconfig} &
        port_offset=$((5000 + ${port_offset}))
    done

    # Wait for clusters to get created.
    wait

    # Merge KUBECONFIGs.
    echo Merging kubeconfigs: ${all_kubeconfigs}
    KUBECONFIG=${all_kubeconfigs} ${KUBECTL} config view --flatten > ${user_kubeconfig}.tmp
    chmod go-rwx ${user_kubeconfig}.tmp
    mv -f ${user_kubeconfig}.tmp ${user_kubeconfig}

    # Set default context to first created cluster.
    ${KUBECTL} config use-context kind-${CLUSTER_PREFIX}-${default_cluster_name}

    # Show contexts.
    ${KUBECTL} config get-contexts
}

create_cluster() {
    cluster_name=$1
    port_offset=$2
    kubeconfig=$3

    local port_file_prefix=${PORT_FILE_DIRECTORY}/${CLUSTER_PREFIX}-${cluster_name}_host_port_
    local api_server_port=$(fixed_or_dynamic_port ${port_file_prefix}6443 $((6443 + ${port_offset})))
    local forwarded_port_80=$(fixed_or_dynamic_port ${port_file_prefix}80 $((80 + ${port_offset})))
    local forwarded_port_443=$(fixed_or_dynamic_port ${port_file_prefix}443 $((443 + ${port_offset})))
    local forwarded_port_4566=$(fixed_or_dynamic_port ${port_file_prefix}4566 $((4566 + ${port_offset})))
    local forwarded_port_30960=$(fixed_or_dynamic_port ${port_file_prefix}30960 $((30960 + ${port_offset})))
    local forwarded_port_30965=$(fixed_or_dynamic_port ${port_file_prefix}30965 $((30965 + ${port_offset})))
    local forwarded_port_30970=$(fixed_or_dynamic_port ${port_file_prefix}30970 $((8970 + ${port_offset})))
    local forwarded_port_30980=$(fixed_or_dynamic_port ${port_file_prefix}30980 $((30980 + ${port_offset})))
    local forwarded_port_30990=$(fixed_or_dynamic_port ${port_file_prefix}30990 $((30990 + ${port_offset})))
    local forwarded_port_30444=$(fixed_or_dynamic_port ${port_file_prefix}30444 $((30444 + ${port_offset})))

    # Create a cluster with the following customizations:
    #  - Local registry enabled.
    #  - Do not check TLS certs when connecting to internal-placeholder.com.
    cat <<EOF | ${KIND} create cluster --name ${CLUSTER_PREFIX}-${cluster_name} --kubeconfig ${kubeconfig} --wait 5m --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${LOCAL_REGISTRY_PORT}"]
    endpoint = ["http://${LOCAL_REGISTRY_NAME}:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."internal-placeholder.com".tls]
    insecure_skip_verify = true
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["https://amr-idc-registry-pre.infra-host.com/v2/cache/"]
networking:
  apiServerAddress: "${apiserver_address}"
  apiServerPort: ${api_server_port}
nodes:
- role: control-plane
  image: kindest/node:v${kind_version}@${kind_sum}
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  - |
    kind: ClusterConfiguration
    apiServer:
        extraArgs:
          cors-allowed-origins: ".*"

  # Extra port mappings allow connections to the specified ports on the kind host to containers.
  extraPortMappings:
  # Port forward for an ingress controller.
  - containerPort: 80
    hostPort: ${forwarded_port_80}
    protocol: TCP
  - containerPort: 443
    hostPort: ${forwarded_port_443}
    protocol: TCP
  # Reserved for argocd-server service
  - containerPort: 30960
    hostPort: ${forwarded_port_30960}
  # Reserved for gitea service
  - containerPort: 30965
    hostPort: ${forwarded_port_30965}
  # Reserved for baremetal-enrollment service
  - containerPort: 30970
    hostPort: ${forwarded_port_30970}
  # Reserved for NetBox service
  - containerPort: 30980
    hostPort: ${forwarded_port_30980}
  # Reserved for vault service
  - containerPort: 30990
    hostPort: ${forwarded_port_30990}
  # Reserved for minio tenant console
  - containerPort: 30444
    hostPort: ${forwarded_port_30443}
  # Reserved for localstack
  - containerPort: 4566
    hostPort: ${forwarded_port_4566}
EOF

    ${KIND} export kubeconfig --name ${CLUSTER_PREFIX}-${cluster_name} --kubeconfig ${kubeconfig}

    # Write port mappings to files which can be used by test scripts.
    echo -n ${forwarded_port_80} > ${PORT_FILE_DIRECTORY}/${cluster_name}_host_port_80
    echo -n ${forwarded_port_443} > ${PORT_FILE_DIRECTORY}/${cluster_name}_host_port_443

    configure_cluster ${kubeconfig}
}

configure_cluster() {
    kubeconfig=$1

    while ! ${KUBECTL} --kubeconfig ${kubeconfig} get nodes ; do sleep 1; done

    # Document the local registry.
    # https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
    cat <<EOF | ${KUBECTL} --kubeconfig ${kubeconfig} apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${LOCAL_REGISTRY_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

    delete_coredns ${kubeconfig}
    install_nginx ${kubeconfig}
    wait_for_cluster ${kubeconfig}
}

# Delete coredns so we can install it with helmfile.
delete_coredns() {
    kubeconfig=$1
    ${KUBECTL} --kubeconfig ${kubeconfig} delete -n kube-system service/kube-dns
    ${KUBECTL} --kubeconfig ${kubeconfig} delete -n kube-system deployment/coredns
}

# Install NGINX Ingress Controller.
# Based on https://kind.sigs.k8s.io/docs/user/ingress/#ingress-nginx
install_nginx() {
    kubeconfig=$1
    ${KUBECTL} --kubeconfig ${kubeconfig} \
        apply -f deployment/kind/ingress-nginx.yaml
}

wait_for_cluster() {
    kubeconfig=$1
    while ! ${KUBECTL} --kubeconfig ${kubeconfig} wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=60m ;
    do sleep 1; done
}

create_multicluster() {
    echo Creating multiple kind clusters.
    cluster_names="global"
    regions=$(get_regions)
    for region in $regions; do
      cluster_names="${cluster_names} ${region}"
      availabilityZones=$(get_availability_zones_for_region ${region})
      for availabilityZone in $availabilityZones; do
        cluster_names="${cluster_names} ${availabilityZone} ${availabilityZone}-network"
      done
    done
    create_clusters ${cluster_names}
}

create_singlecluster() {
    echo Creating single kind cluster.
    cluster_names="global"
    create_clusters ${cluster_names}
    KUBECONFIG=${SECRETS_DIR}/kubeconfig/kind-${CLUSTER_PREFIX}-global.yaml write_config_singlecluster
}

main() {
    KIND_MULTICLUSTER=$(query_config ".Values.test.environment.kind.multicluster")
    CLUSTER_PREFIX=$(query_config ".Values.test.environment.kind.clusterPrefix")
    USE_DYNAMIC_PORTS=$(query_config ".Values.test.environment.kind.useDynamicPorts")

    delete_clusters
    clear_local_state

    if [ "${CREATE_ENABLED}" == "true" ]; then
        if [ "${KIND_MULTICLUSTER}" == "true" ]; then
            create_multicluster
        else
            create_singlecluster
        fi
    fi

    # Has to happen after the singlecluster config has been written, because that creates the kubeconfig directory.
    if [ -n "${EXTERNAL_NETWORKING_KUBECONFIG}" ]
    then
        cp ${EXTERNAL_NETWORKING_KUBECONFIG} ${NETWORKING_KUBECONFIG}
    fi
}

source deployment/common/deploy-on-k8s.sh

main
