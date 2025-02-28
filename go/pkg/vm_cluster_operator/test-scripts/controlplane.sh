#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -eu -o pipefail
set -x

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

KIND_GATEWAY=${KIND_GATEWAY:-172.18.0.1}
NODE_ADDR=${GUEST_ADDR:-172.18.255.201/16}
SSH_PUBLIC_KEY=${SSH_PUBLIC_KEY:-${HOME}/.ssh/id_rsa.pub}
if [[ ! -f ${SSH_PUBLIC_KEY} ]]; then
  echo "ERROR: missing ${SSH_PUBLIC_KEY}"
  exit 1
fi

RANCHER_URL=${RANCHER_URL:-https://$(ip route get 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p'):8443}
rancher_host=$(echo ${RANCHER_URL} | awk -F[/:] '{print $4}')

# The host is required as curl doesn't support CIDR values in no_proxy
if [[ ! -z ${NO_PROXY+x} ]]; then
  export NO_PROXY=${NO_PROXY},172.18.0.0/16,${rancher_host}
  export no_proxy=${no_proxy},172.18.0.0/16,${rancher_host}
fi

dev=$(resolvectl query intel.com | awk '/intel.com/ {print $5}')
nameservers=$(resolvectl status ${dev} | awk '/DNS Servers/ {print "[" $3 ", " $4 ", " $5 "]"}')
if [[ ${nameservers} == "" ]]; then
  nameservers="[1.1.1.1]"
fi

SECRETS_DIR=${SECRETS_DIR:-${SCRIPT_DIR}/../../../../local/secrets}
OPERATOR_DIR=${OPERATOR_DIR:-${SECRETS_DIR}/vm-cluster-operator}
RANCHER_DIR="/tmp/vm-cluster-operator/rancher"
mkdir -p ${OPERATOR_DIR}

start_rancher() {
  proxy_args=""
  if [[ ! -z ${HTTPS_PROXY+x} ]]; then
    proxy_args=" --env HTTPS_PROXY=${HTTPS_PROXY} --env HTTP_PROXY=${HTTP_PROXY} --env NO_PROXY=${NO_PROXY} --env https_proxy=${HTTPS_PROXY} --env http_proxy=${HTTP_PROXY} --env no_proxy=${NO_PROXY}"
  fi
  docker run --privileged -d --restart=unless-stopped -p 8080:80 -p 8443:443 \
    -v ${RANCHER_DIR}:/var/lib/rancher \
    ${proxy_args} \
    --name rancher rancher/rancher:v2.7.7

  # Wait for Rancher to be ready.
  while ! docker logs rancher  2>&1 | grep "Bootstrap Password:"; do
    sleep 2
  done
  bootstrap_password=$(docker logs rancher  2>&1 | grep "Bootstrap Password:" | sed -e 's/.*Bootstrap Password: \(.*\)/\1/')

  # TODO Ideally this would be automated, but it does not look like
  # Rancher supports the non-interactive configuration needed.
  cat <<EOF

Visit https://localhost:8443/dashboard/ and follow instructions to complete Rancher installation.
Enter "${bootstrap_password}" for the bootstrap password.
Enter "${RANCHER_URL}" for the Server URL.

Then create an API key for use by IDC:
  1. Account icon in the upper right
  2. Account & API Keys
  3. Create API Key
  4. Create, default settings are acceptable

Create Rancher cluster:
  RANCHER_ACCESS_KEY=... RANCHER_SECRET_KEY=... $0 create-cluster

EOF
}

stop_rancher() {
  if docker ps -f name=rancher | grep -w rancher; then
    docker stop rancher
    docker rm rancher
  fi
  sudo rm -rf ${RANCHER_DIR}
}

create_mgmt_network() {
  if virsh net-list | grep -w mgmt; then
    echo "Network mgmt already exists. Skipping network creation."
    return
  else
    kind_bridge=$(ip addr | grep ${KIND_GATEWAY} | awk '{print $7}')

    cat <<EOF | tee ${OPERATOR_DIR}/net-mgmt.xml
<network>
  <name>mgmt</name>
  <forward mode='bridge'></forward>
  <bridge name='${kind_bridge}'/>
</network>
EOF

    virsh net-define ${OPERATOR_DIR}/net-mgmt.xml
    virsh net-start mgmt
    virsh net-autostart mgmt
  fi
}

destroy_mgmt_network() {
  if virsh net-list | grep -w mgmt; then
    virsh net-undefine mgmt
    virsh net-destroy mgmt
  fi
}

create_controlplane_node() {
  local -r name="gh-node-cp"

  if virsh list | grep -w ${name}; then
    echo "Control plane node ${name} already exists. Skipping node creation."
    return
  else
    base_image="/var/lib/libvirt/images/base/jammy-server-cloudimg-amd64.img"
    sudo mkdir -p /var/lib/libvirt/images/base
    [[ -f ${base_image} ]] || sudo -E wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img -O ${base_image}

    sudo mkdir -p /var/lib/libvirt/images/${name}
    sudo qemu-img create -f qcow2 -F qcow2 -o backing_file=${base_image} /var/lib/libvirt/images/${name}/${name}.qcow2
    sudo qemu-img resize /var/lib/libvirt/images/${name}/${name}.qcow2 32G

    cat <<EOF | sudo tee /var/lib/libvirt/images/${name}/meta-data
EOF

    proxy_files=""
    if [[ ! -z ${HTTPS_PROXY+x} ]]; then
      proxy_files=$(cat <<EOF
- path: /etc/apt/apt.conf.d/90-proxy
  content: |
    Acquire::http::Proxy "${HTTP_PROXY}";
    Acquire::https::Proxy "${HTTPS_PROXY}";
- path: /etc/profile.d/90-proxy.sh
  content: |
    export HTTPS_PROXY=${HTTPS_PROXY}
    export HTTP_PROXY=${HTTP_PROXY}
    export NO_PROXY=${NO_PROXY}
    export https_proxy=${HTTPS_PROXY}
    export http_proxy=${HTTP_PROXY}
    export no_proxy=${NO_PROXY}
EOF
      )
    fi
    passwd=$(mkpasswd --method=SHA-512 --rounds=4096 devcloud)
    cat <<EOF | sudo tee /var/lib/libvirt/images/${name}/user-data
#cloud-config

fqdn: ${name}.staging.devcloud.intel.com
hostname: ${name}
preserve_hostname: False

users:
- name: devcloud
  ssh_authorized_keys: ['$(cat ${SSH_PUBLIC_KEY})']
  sudo: ALL=(ALL) NOPASSWD:ALL
  groups: sudo
  passwd: ${passwd}
  lock_passwd: false
  shell: /bin/bash
ssh_pwauth: True

packages:
- cloud-initramfs-growroot

write_files:
- path: /usr/local/bin/curl
  content: |
$(sed -e 's/^/    /' ${SCRIPT_DIR}/curl-wrapper.sh)
  permissions: '0755'
${proxy_files}
- path: /etc/networkd-dispatcher/configured.d/enp1s0
  content: |
    #!/bin/bash
    [[ "\${IFACE}" == "enp1s0" ]] || exit 0
    ip link set dev mgmt-br address \$(ip link show dev \${IFACE} | grep link/ether | awk '{print \$2}')
    bridge vlan add vid 2-4094 dev \${IFACE}
  permissions: '0755'
- path: /etc/networkd-dispatcher/configured.d/mgmt-br
  content: |
    #!/bin/bash
    [[ "\${IFACE}" == "mgmt-br" ]] || exit 0
    ip link set \${IFACE} type bridge vlan_filtering 1
    bridge vlan add vid 2-4094 dev \${IFACE} self
  permissions: '0755'
- path: /etc/sysctl.d/60-ipv6-disable.conf
  content: |
    net.ipv6.conf.all.disable_ipv6 = 1
    net.ipv6.conf.default.disable_ipv6 = 1
  permissions: '0644'

runcmd:
- sysctl -p /etc/sysctl.d/60-ipv6-disable.conf
- echo "AllowUsers devcloud" >> /etc/ssh/sshd_config
- systemctl restart sshd
EOF
    cat <<EOF | sudo tee /var/lib/libvirt/images/${name}/network-config
version: 2
renderer: networkd
ethernets:
  enp1s0:
    match:
      name: enp1s0
bridges:
  mgmt-br:
    interfaces: [enp1s0]
    addresses: ['${NODE_ADDR}']
    nameservers:
      addresses: ${nameservers}
    routes: [{'to': 'default', 'via': '${KIND_GATEWAY}'}]
EOF

    cpuset=""
    if [[ ! -z ${GUEST_HOST_CPUSET+x} ]]; then
      cpuset="--cpuset ${GUEST_HOST_CPUSET}"
    fi
    virt-install --name ${name} --os-variant ubuntu22.04 \
      --vcpus 8 ${cpuset} \
      --ram 16384 --import --disk path=/var/lib/libvirt/images/${name}/${name}.qcow2,format=qcow2 \
      --network network=mgmt \
      --cloud-init meta-data=/var/lib/libvirt/images/${name}/meta-data,user-data=/var/lib/libvirt/images/${name}/user-data,network-config=/var/lib/libvirt/images/${name}/network-config --noautoconsole

    # Wait for guest to be up.
    local -r host=${NODE_ADDR%%/*}
    while ! ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null devcloud@${host} -- cloud-init status --wait; do
      sleep 2
    done
  fi
}

destroy_controlplane_node() {
  local name="gh-node-cp"

  if virsh list --all | grep -w ${name}; then
    virsh destroy ${name} || true
    virsh undefine ${name} --remove-all-storage
  fi
}

make_rancher_secrets() {
  for availability_zone in us-dev-1a us-dev-1b; do
    echo -n "${RANCHER_ACCESS_KEY}" >${SECRETS_DIR}/${availability_zone}-rancher-access_key
    echo -n "${RANCHER_SECRET_KEY}" >${SECRETS_DIR}/${availability_zone}-rancher-secret_key
    echo -n "${RANCHER_URL}" >${SECRETS_DIR}/${availability_zone}-rancher-url
  done
}

create_rancher_cluster() {
  local -r access_key=$(cat ${SECRETS_DIR}/us-dev-1a-rancher-access_key)
  local -r secret_key=$(cat ${SECRETS_DIR}/us-dev-1a-rancher-secret_key)
  cat <<EOF >${OPERATOR_DIR}/netrc
machine ${rancher_host} login ${access_key} password ${secret_key}
EOF
  chmod 600 ${OPERATOR_DIR}/netrc

  curl -k --netrc-file ${OPERATOR_DIR}/netrc -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' -d @${OPERATOR_DIR}/cluster.json ${RANCHER_URL}/v1/provisioning.cattle.io.clusters
}

cluster_name() {
  cat ${OPERATOR_DIR}/cluster.json | jq -r '.metadata.name'
}

cluster_namespace() {
  cat ${OPERATOR_DIR}/cluster.json | jq -r '.metadata.namespace'
}

destroy_rancher_cluster() {
  local -r namespace=$(cluster_namespace)
  local -r name=$(cluster_name)
  curl -k --netrc-file ${OPERATOR_DIR}/netrc -X DELETE ${RANCHER_URL}/v1/provisioning.cattle.io.clusters/${namespace}/${name} || true
}

register_controlplane_node() {
  local -r name=$(cluster_name)
  registration_command="null"
  while [[ ${registration_command} == "null" ]]; do
    sleep 2
    cluster_id=$(curl -k --netrc-file ${OPERATOR_DIR}/netrc ${RANCHER_URL}/v3/cluster?name=${name} | jq -r .data[0].id)
    registration_command=$(curl -k --netrc-file ${OPERATOR_DIR}/netrc ${RANCHER_URL}/v3/clusterregistrationtoken?clusterId=${cluster_id} | jq -r .data[0].insecureNodeCommand)
  done

  local -r host=${NODE_ADDR%%/*}
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null devcloud@${host} -- ${registration_command} --etcd --controlplane --worker

  # Wait for cluster to be ready.
  while [[ $(curl -k --netrc-file ${OPERATOR_DIR}/netrc ${RANCHER_URL}/v3/cluster?name=${name} | jq -r .data[0].state) != "active" ]]; do
    sleep 20
  done
}

download_kubeconfig() {
  local -r name=$(cluster_name)
  local -r cluster_id=$(curl -k --netrc-file ${OPERATOR_DIR}/netrc ${RANCHER_URL}/v3/cluster?name=${name} | jq -r .data[0].id)
  mkdir -p ${OPERATOR_DIR}/kubeconfig
  curl -k --netrc-file ${OPERATOR_DIR}/netrc -X POST ${RANCHER_URL}/v3/clusters/${cluster_id}?action=generateKubeconfig | jq -r .config >${OPERATOR_DIR}/kubeconfig/${name}.yaml
  chmod 600 ${OPERATOR_DIR}/kubeconfig/${name}.yaml

  # Merge kubeconfig into user's kubeconfig
  user_kubeconfig=${KUBECONFIG:-${HOME}/.kube/config}
  all_kubeconfigs=${user_kubeconfig}
  kubeconfig=${OPERATOR_DIR}/kubeconfig/$(cluster_name).yaml
  [ "${all_kubeconfigs}" == "" ] || all_kubeconfigs=${all_kubeconfigs}:
  all_kubeconfigs=${all_kubeconfigs}${kubeconfig}
  KUBECONFIG=${all_kubeconfigs} kubectl config view --flatten > ${user_kubeconfig}.tmp
  chmod go-rwx ${user_kubeconfig}.tmp
  mv -f ${user_kubeconfig}.tmp ${user_kubeconfig}
  kubectl config get-contexts
}

deploy_nfs_storage() {
  kubectl --context $(cluster_name) apply -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/example/nfs-provisioner/nfs-server.yaml

  helm repo add csi-driver-nfs https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts
  helm install csi-driver-nfs csi-driver-nfs/csi-driver-nfs --namespace kube-system --version v4.9.0 --set externalSnapshotter.enabled=true
  # TODO A clone-strategy of csi-clone is preferred but appears to be unsupported by nfs-csi:
  #   Warning ExternalExpanding 2m54s volume_expand Ignoring the PVC: didn't find a plugin capable of expanding the volume; waiting for an external controller to process this PVC.
  # The ControllerExpandVolume and NodeExpandVolume functions return unimplemented.
  # The snapshot strategy also requires volume expansion, so that leaves copy as the only choice.
  cat <<EOF | kubectl --context $(cluster_name) apply -f -
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    storageclass.kubevirt.io/is-default-virt-class: "true"
    cdi.kubevirt.io/clone-strategy: copy
  name: nfs-csi
provisioner: nfs.csi.k8s.io
parameters:
  server: nfs-server.default.svc.cluster.local
  share: /
  # csi.storage.k8s.io/provisioner-secret is only needed for providing mountOptions in DeleteVolume
  # csi.storage.k8s.io/provisioner-secret-name: "mount-options"
  # csi.storage.k8s.io/provisioner-secret-namespace: "default"
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
mountOptions:
  - nfsvers=4.1
EOF
}

deploy_kubevirt() {
  export RELEASE=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
  kubectl --context $(cluster_name) apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml
  cat <<EOF | kubectl --context $(cluster_name) apply -f -
---
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  certificateRotateStrategy: {}
  configuration:
    architectureConfiguration:
      amd64:
        emulatedMachines:
        - q35
        - pc-q35*
        - pc
        - pc-i440fx*
    developerConfiguration:
      featureGates:
      - ExpandDisks
  customizeComponents: {}
  imagePullPolicy: IfNotPresent
  workloadUpdateStrategy: {}
EOF
}

deploy_cdi() {
  export VERSION=$(curl -s https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  kubectl --context $(cluster_name) apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-operator.yaml
  kubectl --context $(cluster_name) apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-cr.yaml

  cat <<EOF | kubectl --context $(cluster_name) apply -f -
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cdi-cloner
rules:
- apiGroups: ["cdi.kubevirt.io"]
  resources: ["datavolumes/source"]
  verbs: ["create"]
EOF
}

install_virtctl() {
  LOCALBIN=${SCRIPT_DIR}/../../../../bin
  VIRTCTL_VERSION=v1.3.1
  VIRTCTL=${LOCALBIN}/virtctl-${VIRTCTL_VERSION}
  mkdir -p ${LOCALBIN}
  test -s ${VIRTCTL} || { wget -q https://github.com/kubevirt/kubevirt/releases/download/${VIRTCTL_VERSION}/virtctl-${VIRTCTL_VERSION}-linux-amd64 -O ${VIRTCTL}; }
  chmod +x ${VIRTCTL}
  cp ${VIRTCTL} ${HOME}/.local/bin/virtctl
}

configure_idc_to_use_cluster() {
  CLUSTER_ID=harvester1 # KIND development clusters expect cluster ID to be harvester1
  cp ${OPERATOR_DIR}/kubeconfig/$(cluster_name).yaml ${SECRETS_DIR}/harvester-kubeconfig/${CLUSTER_ID}
}

create_cluster() {
  create_mgmt_network
  create_controlplane_node
  make_rancher_secrets
  create_rancher_cluster
  register_controlplane_node
  download_kubeconfig
  deploy_nfs_storage
  deploy_kubevirt
  deploy_cdi
  install_virtctl
  configure_idc_to_use_cluster
}

destroy_cluster() {
  destroy_rancher_cluster
  destroy_controlplane_node
  destroy_mgmt_network
}

# Some of the functions read from cluster.json, expand it first
envsubst <${SCRIPT_DIR}/cluster.json.envsubst >${OPERATOR_DIR}/cluster.json

case $1 in
  start-rancher) start_rancher ;;
  stop-rancher) stop_rancher ;;

  # The below commands requre RANCHER_ACCESS_KEY and RANCHER_SECRET_KEY to
  # be defined in the environment.
  create-cluster) create_cluster ;;
  destroy-cluster) destroy_cluster ;;
esac
