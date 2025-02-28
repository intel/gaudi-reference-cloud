#!/usr/bin/env bash
#
# Attention:
# It is recommended to run the following instead of using this script directly.
#   make deploy-network-cluster-in-kind
#
set -ex
cd "$(dirname "$0")/../.."

CLUSTER_PREFIX=${CLUSTER_PREFIX:-idc}
#SECRETS_DIR=${SECRETS_DIR:-local/secrets}

# This node image supports both amd64 and arm64. See: https://github.com/kubernetes-sigs/kind/releases
kind_sum=sha256:dad5a6238c5e41d7cac405fae3b5eda2ad1de6f1190fa8bfc64ff5bb86173213
kind_version=1.28.0

user_kubeconfig=${KUBECONFIG:-${HOME}/.kube/config}
echo New clusters will be added to KUBECONFIG=${user_kubeconfig}

create_cluster() {
    cluster_name=$1
    port_offset=$2
    kubeconfig=$3

    # Create a cluster with the following customizations:
    #  - Local registry enabled.
    #  - Do not check TLS certs when connecting to internal-placeholder.com.
    cat <<EOF | kind create cluster --name ${CLUSTER_PREFIX}-${cluster_name} --kubeconfig ${kubeconfig} --wait 5m --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."internal-placeholder.com".tls]
    insecure_skip_verify = true
networking:
  apiServerAddress: "${apiserver_address}"
  apiServerPort: $((6443 + ${port_offset}))
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
    hostPort: $((80 + ${port_offset}))
    protocol: TCP
  - containerPort: 443
    hostPort: $((443 + ${port_offset}))
    protocol: TCP
  # Reserved for baremetal-enrollment service
  - containerPort: 30970
    hostPort: $((8970 + ${port_offset}))
  # Reserved for NetBox service
  - containerPort: 30980
    hostPort: $((30980 + ${port_offset}))
  # Reserved for vault service
  - containerPort: 30990
    hostPort: $((30990 + ${port_offset}))
EOF

    # Write port mappings to files which can be used by test scripts.
    echo -n $((80 + ${port_offset})) > local/${cluster_name}_host_port_80
    echo -n $((443 + ${port_offset})) > local/${cluster_name}_host_port_443

    configure_cluster ${kubeconfig}
}

configure_cluster() {
    kubeconfig=$1

    while ! kubectl --kubeconfig ${kubeconfig} get nodes ; do sleep 1; done

    # Connect the registry to the kind network if not already connected.
    docker network connect "kind" "${reg_name}" || true

    # Document the local registry.
    # https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
    cat <<EOF | kubectl --kubeconfig ${kubeconfig} apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

    delete_coredns ${kubeconfig}
    install_nginx ${kubeconfig}
    wait_for_cluster ${kubeconfig}
}

# Delete coredns so we can install it with helmfile.
delete_coredns() {
    kubeconfig=$1
    kubectl --kubeconfig ${kubeconfig} delete -n kube-system service/kube-dns
    kubectl --kubeconfig ${kubeconfig} delete -n kube-system deployment/coredns
}

# Install NGINX Ingress Controller.
# Based on https://kind.sigs.k8s.io/docs/user/ingress/#ingress-nginx
install_nginx() {
    kubeconfig=$1
    pwd
    ls -l deployment
    ls -l deployment/kind
    kubectl --kubeconfig ${kubeconfig} \
        apply -f deployment/kind/ingress-nginx.yaml
}

wait_for_cluster() {
    kubeconfig=$1
    while ! kubectl --kubeconfig ${kubeconfig} wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=60m ;
    do sleep 1; done
}

main() {
    if [ "${IDC_ENV}" == "kind-multicluster" ]; then
      kubeconfig=${SECRETS_DIR}/kubeconfig/kind-${CLUSTER_PREFIX}-us-dev-1a-network.yaml
      create_cluster "us-dev-1a-network" "30000" ${kubeconfig} &

      # Wait for clusters to get created.
      wait

      # Merge network cluster kubeconfig in to the existing kubeconfig
      all_kubeconfigs=${user_kubeconfig}:${kubeconfig}

      # Merge KUBECONFIGs.
      echo Merging kubeconfigs: ${all_kubeconfigs}
      KUBECONFIG=${all_kubeconfigs} kubectl config view --flatten > ${user_kubeconfig}.tmp
      chmod go-rwx ${user_kubeconfig}.tmp
      mv -f ${user_kubeconfig}.tmp ${user_kubeconfig}

      # Show contexts and nodes.
      kubectl config get-contexts
      kind get clusters | grep ${CLUSTER_PREFIX} | xargs -i -P 1 kubectl --context kind-{} get nodes
    fi


}

source deployment/common/deploy-on-k8s.sh

main
