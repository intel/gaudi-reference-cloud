#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -o pipefail
set -o nounset
set -o errexit

err_report() {
  echo "Exited with error on line $1"
}
trap 'err_report $LINENO' ERR

log_message() {
    command echo $(date) "$@"
}

IFS=$'\n\t'

function print_help {
  echo "usage: $0 [options]"
  echo "Bootstraps an instance into an IDC Kubernetes cluster"
  echo ""
  echo "-h,--help print this help"
  echo "--ca-cert The Kubernetes CA certificate"
  echo "--ca-key The Kubernetes CA private key"
  echo "--apiserver-lb The Kubernetes apiserver loadbalancer"
  echo "--bootstrap-token The bootstrap token used by kubelet for joining the cluster"
  echo "--containerd-envvars The environment variables that will used to configure the containerd service"
  echo "--cluster-dns The coredns service ip"
  echo "--max-pods The maximum amount of pods for the worker node"
  echo "--storage-agent-url The url of the weka cluster"
  echo "--storage-weka-sw-version The version of the weka agent to configure"
  echo "--storage-weka-container-name The name of the weka container"
  echo "--storage-weka-num-cores The number of cores to use for creating weka container"
  echo "--kubelet-node-labels Labels to add to the node"
}

POSITIONAL=()

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -h | --help)
      print_help
      exit 1
      ;;
    --ca-cert)
      CA_CERT=$2
      shift
      shift
      ;;
    --apiserver-lb)
      APISERVER_LB=$2
      shift
      shift
      ;;
    --apiserver-lb-port)
      APISERVER_LB_PORT=$2
      shift
      shift
      ;;
    --bootstrap-token)
      BOOTSTRAP_TOKEN=$2
      shift
      shift
      ;;
    --containerd-envvars)
      CONTAINERD_ENVVARS=$2
      shift
      shift
      ;;
    --cluster-dns)
      CLUSTER_DNS=$2
      shift
      shift
      ;;
    --max-pods)
      MAX_PODS=$2
      shift
      shift
      ;;
    --storage-agent-url)
      STORAGE_AGENT_URL=$2
      shift
      shift
      ;;
    --storage-weka-sw-version)
      STORAGE_WEKA_SW_VERSION=$2
      shift
      shift
      ;;
    --storage-weka-container-name)
      STORAGE_WEKA_CONTAINER_NAME=$2
      shift
      shift
      ;;
    --storage-weka-num-cores)
      STORAGE_WEKA_NUM_CORES=$2
      shift
      shift
      ;;
    --storage-weka-mode)
      STORAGE_WEKA_MODE=$2
      shift
      shift
      ;;
    --kubelet-node-labels)
      KUBELET_NODE_LABELS=$2
      shift
      shift
      ;;
    *)                   # unknown option
      POSITIONAL+=("$1") # save it in an array for later
      shift              # past argument
      ;;
  esac
done

log_message "Starting the IDC Kubernetes bootstrap script"

# Checking if OS is supported. Ubuntu is the only supported OS at this moment
DISTRO=$(lsb_release -i | cut -d: -f2 | sed s/'^\t'//)
OS_VERSION=$(lsb_release -sr)
if [[ ! "${DISTRO}" == "Ubuntu" ]]; then
  log_message "${DISTRO} is not supported"
  exit 1
fi

# Check if it's being executed by root
if [[ ! $(id -u) -eq 0 ]]; then
  log_message "The script must be run as root"
  exit 1
fi

# Mask the sanitize_ram.service because it overrides the value of vm.overcommit_memory (required for kubelet)
log_message "Stopping and masking the sanitize_ram.service"
SERVICE_NAME="sanitize_ram.service"
rm -rf /etc/systemd/system/$SERVICE_NAME
systemctl mask $SERVICE_NAME
systemctl stop $SERVICE_NAME
systemctl disable $SERVICE_NAME

# Mask the unattended-upgrades.service to prevent unwanted automatic upgrades
log_message "Stopping and masking the unattended-upgrades"
SERVICE_NAME="unattended-upgrades.service"
rm -rf /etc/systemd/system/multi-user.target.wants/$SERVICE_NAME
systemctl mask $SERVICE_NAME
systemctl stop $SERVICE_NAME
systemctl disable $SERVICE_NAME
if test -f /etc/apt/apt.conf.d/20auto-upgrades; then
  sed -i '/APT::Periodic::Unattended-Upgrade/d' /etc/apt/apt.conf.d/20auto-upgrades
  echo -e 'APT::Periodic::Unattended-Upgrade "0";' | tee -a /etc/apt/apt.conf.d/20auto-upgrades
fi

#(TEMP) Configure hugepages for Gaudi2 hosts
if lshw | grep -i gaudi >/dev/null; then
  log_message "Configuring hugepages for Gaudi2 hosts"
  sed -i '/^vm.nr_hugepages/d' /etc/sysctl.conf
  echo -e 'vm.nr_hugepages=156300' | tee -a /etc/sysctl.conf

  if lshw | grep -i gaudi | grep "1.15\|1.16\|1.17\|1.18" >/dev/null; then
  #Replace habana container runtime config file for 1.15 or 1.16 or 1.17 or 1.18
    log_message "Configuring habana container runtime for Gaudi2 version 1.15 or 1.16 or 1.17 or 1.18"
    mkdir -p /etc/habana-container-runtime
    cat > /etc/habana-container-runtime/config.toml <<END
disable-require = false
#accept-habana-visible-devices-envvar-when-unprivileged = true
#accept-habana-visible-devices-as-volume-mounts = false

## Uncomment and set to false if you are running inside kubernetes
## environment with Habana device plugin. Defaults to true
mount_accelerators = false

## Mount uverbs mounts the attached infiniband_verb device attached to
## the selected accelerator devices. Defaults to true.
#mount_uverbs = false

## [Optional section]
[network-layer-routes]
## Override the default path on hode for the network configuration layer.
## default:/etc/habanalabs/gaudinet.json
# path = "/etc/habanalabs/gaudinet.json"

[habana-container-cli]
#root = "/run/habana/driver"
#path = "/usr/bin/habana-container-cli"
environment = []

## Uncomment to enable logging
#debug = "/var/log/habana-container-hook.log"

[habana-container-runtime]

## Always try to expose devices on any container, no matter if requested the devices
## This is not recommended as it exposes devices and required metadata into any container
## Default: true
visible_devices_all_as_default = false

## Uncomment to enable logging
#debug = "/var/log/habana-container-runtime.log"

## Logging level. Supported values: "info", "debug"
#log_level = "debug"

## By default, runc creates cgroups and sets cgroup limits on its own (this mode is known as fs cgroup driver).
## By setting to true runc switches to systemd cgroup driver.
## Read more here: https://github.com/opencontainers/runc/blob/main/docs/systemd.md
#systemd_cgroup = false

## Use prestart hook for configuration. Valid modes: oci, legacy
## Default: oci
# mode = legacy
END
  fi
fi

if [ -f /etc/gaudinet.json ]; then
  log_message "Configuring layer 3 network configuration for Gaudi"
  sed -i 's/^# path = "\/etc\/habanalabs\/gaudinet.json"/path = "\/etc\/gaudinet.json"/g' /etc/habana-container-runtime/config.toml
fi

#(TEMP) Remove explicit disabling of ipv6 from sysctl.conf
sed -i '/^net.ipv6.conf/d' /etc/sysctl.conf
echo -e 'net.ipv6.conf.all.disable_ipv6 = 0\nnet.ipv6.conf.default.disable_ipv6 = 0\nnet.ipv6.conf.lo.disable_ipv6 = 0' | tee -a /etc/sysctl.conf
#(TEMP) Remove explicit disabling of ipv6 from grub
sed -i 's/ipv6.disable=1 //' /etc/default/grub

# Reload sysctl config
# This is to fix the vm.overcommit_memory in baremetal systems.
sysctl --system

# Store apiserver CA
echo "${CA_CERT}" | base64 --decode > /var/lib/kubernetes/pki/ca.crt

#Using Kubernetes default value of 110 for max-pods, if max-pods is not defined
if [ -z ${MAX_PODS+x} ]; then
  MAX_PODS=110
fi

# Funcionts used to calculate the minimum amount of resources to reserve for kubelet

# Helper function which calculates the amount of the given resource (either CPU or memory)
# to reserve in a given resource range, specified by a start and end of the range and a percentage
# of the resource to reserve. Note that we return zero if the start of the resource range is
# greater than the total resource capacity on the node. Additionally, if the end range exceeds the total
# resource capacity of the node, we use the total resource capacity as the end of the range.
# Args:
#   $1 total available resource on the worker node in input unit (either millicores for CPU or Mi for memory)
#   $2 start of the resource range in input unit
#   $3 end of the resource range in input unit
#   $4 percentage of range to reserve in percent*100 (to allow for two decimal digits)
# Return:
#   amount of resource to reserve in input unit
get_resource_to_reserve_in_range() {
  local total_resource_on_instance=$1
  local start_range=$2
  local end_range=$3
  local percentage=$4
  resources_to_reserve="0"
  if (($total_resource_on_instance > $start_range)); then
    resources_to_reserve=$(((($total_resource_on_instance < $end_range ? $total_resource_on_instance : $end_range) - $start_range) * $percentage / 100 / 100))
  fi
  echo $resources_to_reserve
}

# Calculates the amount of memory to reserve for kubeReserved in mebibytes. KubeReserved is a function of pod
# density so we are calculating the amount of memory to reserve for Kubernetes systems daemons by
# considering the maximum number of pods this instance type supports.
# Args:
#   $1 the max number of pods per instance type (MAX_PODS)
# Return:
#   memory to reserve in Mi for the kubelet
get_memory_mebibytes_to_reserve() {
  local max_num_pods=$1
  memory_to_reserve=$((11 * $max_num_pods + 255))
  echo $memory_to_reserve
}

# Calculates the amount of CPU to reserve for kubeReserved in millicores from the total number of vCPUs available on the instance.
# From the total core capacity of this worker node, we calculate the CPU resources to reserve by reserving a percentage
# of the available cores in each range up to the total number of cores available on the instance.
# We are using these CPU ranges from GKE (https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-architecture#node_allocatable):
# 6% of the first core
# 1% of the next core (up to 2 cores)
# 0.5% of the next 2 cores (up to 4 cores)
# 0.25% of any cores above 4 cores
# Return:
#   CPU resources to reserve in millicores (m)
get_cpu_millicores_to_reserve() {
  local total_cpu_on_instance=$(($(nproc) * 1000))
  local cpu_ranges=(0 1000 2000 4000 $total_cpu_on_instance)
  local cpu_percentage_reserved_for_ranges=(600 100 50 25)
  cpu_to_reserve="0"
  for i in "${!cpu_percentage_reserved_for_ranges[@]}"; do
    local start_range=${cpu_ranges[$i]}
    local end_range=${cpu_ranges[(($i + 1))]}
    local percentage_to_reserve_for_range=${cpu_percentage_reserved_for_ranges[$i]}
    cpu_to_reserve=$(($cpu_to_reserve + $(get_resource_to_reserve_in_range $total_cpu_on_instance $start_range $end_range $percentage_to_reserve_for_range)))
  done
  echo $cpu_to_reserve
}


# this is temporary fix for A100 node.
# TODO: apply this fix at os image level
# 1. do not install containerd by installing docker
# 2. make sure to apply correct containerd configuration
while [ -f /usr/bin/nvidia-ctk ]; do
  echo "nvidia-ctk found, assume gpu node, generating containerd config with nvidia runtime"
cat > /etc/containerd/config.toml <<END
version = 2
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "nvidia"
      discard_unpacked_layers = true
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia]
          privileged_without_host_devices = false
          runtime_engine = ""
          runtime_root = ""
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
            BinaryName = "/usr/bin/nvidia-container-runtime"
            SystemdCgroup = true
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"
  [plugins."io.containerd.grpc.v1.cri".cni]
    bin_dir = "/opt/cni/bin"
    conf_dir = "/etc/cni/net.d"
END
  break
done


# Create containerd config file and set cgroup driver to systemd
while [ ! -f /etc/containerd/config.toml ]; do
cat > /etc/containerd/config.toml <<END
version = 2
root = "/var/lib/containerd"
state = "/run/containerd"

[grpc]
  address = "/run/containerd/containerd.sock"

[plugins]
  [plugins."io.containerd.grpc.v1.cri".containerd]
    default_runtime_name = "runc"
    discard_unpacked_layers = true

    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
        runtime_type = "io.containerd.runc.v2"

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
          SystemdCgroup = true

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.habana]
        runtime_type = "io.containerd.runc.v2"

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.habana.options]
          BinaryName = "/usr/bin/habana-container-runtime"

  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"

  [plugins."io.containerd.grpc.v1.cri".cni]
    bin_dir = "/opt/cni/bin"
    conf_dir = "/etc/cni/net.d"
END
done


# Create containerd service
while [ ! -f /etc/systemd/system/containerd.service ]; do
cat > /etc/systemd/system/containerd.service <<END
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
Environment=${CONTAINERD_ENVVARS}
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/containerd

Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
# TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
END
done

# Create bootstrap kubeconfig
export KUBECONFIG=/var/lib/kubelet/bootstrap-kubeconfig
kubectl config set-cluster bootstrap --server="https://${APISERVER_LB}:${APISERVER_LB_PORT}" --certificate-authority /var/lib/kubernetes/pki/ca.crt
kubectl config set-credentials kubelet-bootstrap --token ${BOOTSTRAP_TOKEN}
kubectl config set-context bootstrap --cluster bootstrap  --user kubelet-bootstrap
kubectl config use-context bootstrap

# Get the amount of resources to reserve for kubelet
mebibytes_to_reserve=$(get_memory_mebibytes_to_reserve $MAX_PODS)
cpu_millicores_to_reserve=$(get_cpu_millicores_to_reserve)

# Create kubelet config
while [ ! -f /var/lib/kubelet/kubelet-config.yaml ]; do
cat > /var/lib/kubelet/kubelet-config.yaml <<END
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  anonymous:
    enabled: false
  webhook:
    enabled: true
  x509:
    clientCAFile: /var/lib/kubernetes/pki/ca.crt
authorization:
  mode: Webhook
cgroupDriver: systemd
cgroupRoot: /
clusterDNS:
  - ${CLUSTER_DNS}
clusterDomain: "cluster.local"
resolvConf: /run/systemd/resolve/resolv.conf
runtimeRequestTimeout: "15m"
rotateCertificates: true
serverTLSBootstrap: true
protectKernelDefaults: true
readOnlyPort: 0
featureGates:
  RotateKubeletServerCertificate: true
tlsCipherSuites:
  - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
  - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  - TLS_AES_256_GCM_SHA384
  - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
  - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
kubeReserved:
  cpu: ${cpu_millicores_to_reserve}m
  memory: ${mebibytes_to_reserve}Mi
  ephemeral-storage: 1Gi
systemReservedCgroup: /system
kubeReservedCgroup: /runtime
END
done

KUBELET_START_COMMAND="/usr/local/bin/kubelet --bootstrap-kubeconfig=/var/lib/kubelet/bootstrap-kubeconfig --config=/var/lib/kubelet/kubelet-config.yaml --kubeconfig=/var/lib/kubelet/kubeconfig --container-runtime-endpoint=unix:///var/run/containerd/containerd.sock --v=2"
if [ ! -z ${KUBELET_NODE_LABELS+x} ]; then
  KUBELET_START_COMMAND="$KUBELET_START_COMMAND --node-labels=$KUBELET_NODE_LABELS"
fi


# Create kubelet service
while [ ! -f /etc/systemd/system/kubelet.service ]; do
cat > /etc/systemd/system/kubelet.service <<END
[Unit]
Description=Kubernetes Kubelet
Documentation=https://github.com/kubernetes/kubernetes
After=containerd.service
Requires=containerd.service

[Service]
ExecStart=${KUBELET_START_COMMAND}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
END
done

# Reload systemd services and start kubelet
systemctl daemon-reload

# Start services
systemctl start containerd
systemctl enable containerd

systemctl start kubelet
systemctl enable kubelet

# Get network information.
INTERFACE_NAME=$(ip -o -4 route show to default | awk '{print $5}')
GATEWAY=$(ip -o -4 route show to default | awk '{print $3}')

# Some nodes will have a dedicated interface for storage configuration named "storage0-tenant".
# If it is found, it will be used instead of default interface.
if ip addr | grep "storage0-tenant"; then
  log_message "Dedicated storage interface found"
  INTERFACE_NAME="storage0-tenant"
  GATEWAY=$(ip -o -4 route show | grep -E ".*via.*storage0-tenant.*" | awk '{print $3}')
fi

IP_ADDR=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f1 | head -n 1)
NETMASK=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f2 | head -n 1)

echo "Network information: $INTERFACE_NAME $IP_ADDR $NETMASK $GATEWAY"

#Install nfs-common
apt install -y nfs-common

# Install and configure stunnel for VAST
apt install -y stunnel4

while [ ! -f /etc/stunnel/stunnel.conf ]; do
cat > /etc/stunnel/stunnel.conf <<END
pid = /var/run/stunnel4/stunnel.pid
socket = r:TCP_NODELAY=1

[nfs4]
client=yes
accept=0.0.0.0:2049
connect=vip1.vast-pdx05-1.us-staging-1.cloud.intel.com:2049
ciphers=ALL
sslVersionMin=TLSv1.2
sslVersionMax=TLSv1.3
verifyChain=yes
CAPath = /etc/ssl/certs
END
done

# Ensure service starts on boot
systemctl enable stunnel4
systemctl restart stunnel4

if [[ "${STORAGE_WEKA_MODE}" == "enable" ]]; then
  # Weka storage agent installation
  if [ -z ${STORAGE_AGENT_URL+x} ]; then
    STORAGE_AGENT_URL="internal-placeholder.com:14000"
    echo "Using default storage agent URL: $STORAGE_AGENT_URL"
  fi
  if [ -z ${STORAGE_WEKA_SW_VERSION+x} ]; then
    STORAGE_WEKA_SW_VERSION="4.2.2"
    echo "Using default storage agent version: $STORAGE_WEKA_SW_VERSION"
  fi

  # Hardcoding this to client because weka only supports that name for the local container.
  # Any other name won't work.
  STORAGE_WEKA_CONTAINER_NAME="client"

  # Ensure WEKA MACHINE ID is created (unique)
  WEKA_MACHINE_IDENTIFIER="/opt/weka/data/agent/machine-identifier"
  #check if machine identifier exist or not , if not create one
  if [ ! -f "$WEKA_MACHINE_IDENTIFIER" ];then
    mkdir -p '/opt/weka/data/agent'
    dbus-uuidgen --ensure=$WEKA_MACHINE_IDENTIFIER
    echo "Created weka machine identifier successfully"
  fi

  # Print args
  echo "--storage-agent-url: $STORAGE_AGENT_URL"
  echo "--storage-weka-sw-version: $STORAGE_WEKA_SW_VERSION"
  echo "--storage-weka-container-name: $STORAGE_WEKA_CONTAINER_NAME"
  echo "--storage-weka-num-cores: $STORAGE_WEKA_NUM_CORES"

  curl -vk "http://$STORAGE_AGENT_URL/dist/v1/install" | sh
  weka version get $STORAGE_WEKA_SW_VERSION
  weka version set $STORAGE_WEKA_SW_VERSION
  weka local setup container --name $STORAGE_WEKA_CONTAINER_NAME --cores $STORAGE_WEKA_NUM_CORES --only-frontend-cores --net "$INTERFACE_NAME/$IP_ADDR/$NETMASK/$GATEWAY"
  if weka local ps | grep -i stem >/dev/null; then
    log_message "The storage agent has been setup successfully"
  else
    log_message "There was an issue installing the required storage agent. Please check the output of 'weka local ps' for additional details"
  fi
fi

log_message "The IDC Kubernetes bootstrap script has completed."