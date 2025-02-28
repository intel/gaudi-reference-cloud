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
  echo "Bootstraps an instance into an IDC K8aaS cluster"
  echo ""
  echo "-h,--help print this help"
  echo "--provider-id Sets the provider ID to use. (default: 1)"
  echo "--registration-command The clusters registration command."
  echo "--configure-os Configures the OS to be used in an IDC K8aaS cluster. (default: true)"
  echo "--intel-proxies Sets the Intel  proxies. (default: false)"
  echo "--intel-ca Install the Intel certificate authority. (default: false)"
  echo "--ntp-server The NTP server to use. This option is valid only when --configure-os is true."
}

POSITIONAL=()

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -h | --help)
      print_help
      exit 1
      ;;
    --provider-id)
      PROVIDER_ID="$2"
      shift
      shift
      ;;
    --registration-command)
      REGISTRATION_COMMAND=$2
      shift
      shift
      ;;
    --configure-os)
      CONFIGURE_OS=$2
      shift
      shift
      ;;
    --intel-proxies)
      INTEL_PROXIES=$2
      shift
      shift
      ;;
    --intel-ca)
      INTEL_CA=$2
      shift
      shift
      ;;
    --ntp-server)
      NTP_SERVER=$2
      shift
      shift
      ;;
    *)                   # unknown option
      POSITIONAL+=("$1") # save it in an array for later
      shift              # past argument
      ;;
  esac
done

log_message "Starting the IDC K8aaS bootstrap script"


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

PROVIDER_ID="${PROVIDER_ID:-1}"
REGISTRATION_COMMAND="${REGISTRATION_COMMAND:-}"
CONFIGURE_OS="${CONFIGURE_OS:-true}"
INTEL_PROXIES="${INTEL_PROXIES:-false}"
INTEL_CA="${INTEL_CA:-false}"
NTP_SERVER="${NTP_SERVER:-}"


configure_os() {  
  # Check if host is already part of a k8s cluster, if true
  # stop execution. Host must be cleaned up first
  log_message "Check if host belongs to an existing k8s cluster"
  # If any of these dirs is found, execution will be stopped
  for f in "/etc/kubernetes" "/opt/rke" "/run/secrets/kubernetes.io" "/var/lib/etcd" "/var/lib/kubelet"; do
      test -d ${f} && echo "file ${f} was found, host seems to be in use. Clean the host and re-run the script" && exit 1
  done

  # This is to avoid prompts when updating the host, Ubuntu only
  log_message "Disable apt needrestart"
  test -f /etc/apt/apt.conf.d/99needrestart && mv /etc/apt/apt.conf.d/99needrestart{,.disabled} || echo "needrestart not present"

  log_message "Remove proxies from apt"
  test -f /etc/apt/apt.conf.d/proxy.conf && rm -f /etc/apt/apt.conf.d/proxy.conf || echo "no proxy.conf"
  test -f /etc/apt/apt.conf.d/90curtin-aptproxy && rm -f /etc/apt/apt.conf.d/90curtin-aptproxy || echo "no 90curtin-aptproxy"

  log_message "Disable IPV6"
  if ! grep "ipv6.disable=1" /etc/default/grub; then
      for line in "net.ipv6.conf.all.disable_ipv6" "net.ipv6.conf.default.disable_ipv6" "net.ipv6.conf.lo.disable_ipv6" "net.ipv6.conf.all.forwarding"; do
          if grep -E "^${line}" /etc/sysctl.conf; then
              sed -i "s/^${line}\s=\s[^1]/${line} = 1/g" /etc/sysctl.conf
          else
              # append line if it doesn't exist
              echo "add ${line} = 1"
              echo "" >> /etc/sysctl.conf
              echo "# caas configuration" >> /etc/sysctl.conf
              echo "${line} = 1" >> /etc/sysctl.conf
          fi
      done

      # reload options
      sysctl -p 
      # temporary disable ipv6
      echo 1 > /proc/sys/net/ipv6/conf/all/disable_ipv6
  else
      log_message "IPV6 has been disabled in /etc/default/grub"
  fi

  log_message "Set net.ipv4.ip_forward"
  line="net.ipv4.ip_forward"
  if grep -E "^${line}" /etc/sysctl.conf; then
      echo "Change ${line}"
      sed -i "s/^${line}.*/${line} = 1/g" /etc/sysctl.conf
  else
      # append line if it doesn't exist
      echo "Append ${line}"
      echo "# caas configuration" >> /etc/sysctl.conf
      echo "${line} = 1" >> /etc/sysctl.conf
  fi

  log_message "Set fs.inotify.max_user_instances"
  line="fs.inotify.max_user_instances"
  if grep -E "^${line}" /etc/sysctl.conf; then
      echo "Change ${line}"
      sed -i "s/^${line}.*/${line} = 1024/g" /etc/sysctl.conf
  else
      # append line if it doesn't exist
      echo "Append ${line}"
      echo "# caas configuration" >> /etc/sysctl.conf
      echo "${line} = 1024" >> /etc/sysctl.conf
  fi

  # Install minimum packages
  log_message "Update and install packages"
  apt update -y > /dev/null
  apt install -y unzip nfs-common ntp software-properties-common

  log_message "Stop firewall service"
  systemctl stop ufw.service
  systemctl disable ufw.service

  if [[ "${OS_VERSION}" == "22.04" ]]; then 
      log_message "Configure NTP server (timesyncd)"
      [ -n "${NTP_SERVER}" ] &&  sed -i "s/^\[Time\].*/&\nNTP=${NTP_SERVER}/g" /etc/systemd/timesyncd.conf || echo "No custom ntp server was set"

      log_message "Start NTP service"
      systemctl enable systemd-timesyncd
      systemctl restart systemd-timesyncd
  elif  [[ "${OS_VERSION}" == "20.04" ]]; then 
      log_message "Configure NTP server (ntp)"
      if [ -n "${NTP_SERVER}" ]; then
          sed -i "s/^pool.*/#&/g" /etc/ntp.conf 
    sed -i "s/^server.*/#&/g" /etc/ntp.conf
    echo "server ${NTP_SERVER}" >> /etc/ntp.conf
      else
    log_message "No custom ntp server was set"
      fi

      log_message "Start NTP service"
      systemctl enable ntp
      systemctl restart ntp
  fi
}

if [[ "${PROVIDER_ID}" == "1" ]]; then
  if [[ -z "${REGISTRATION_COMMAND}" ]]; then
    log_message "--registration-command must be provided for provider-id ${PROVIDER_ID}"
    exit 1
  fi
fi

if $INTEL_PROXIES; then
  log_message "Setting up the Intel proxies"
  export http_proxy=http://internal-placeholder.com:912; export https_proxy=http://internal-placeholder.com:912; export no_proxy=localhost,127.0.0.0/8,10.0.0.0/8,.intel.com
fi

if $CONFIGURE_OS; then
  configure_os
fi

if $INTEL_CA; then
  log_message "Installing the Intel CA"
  certs_file='IntelSHA2RootChain-Base64.zip'
  certs_url="http://certificates.intel.com/repository/certificates/$certs_file"
  if curl --output /dev/null --silent --head --fail "$certs_url"; then
    log_message "The ${certs_url} URL can be reached"
  else
    log_message "The ${certs_url} URL cannot be reached. Try setting up the proxies with --intel-proxies true"
    exit 1
  fi
  certs_folder='/usr/local/share/ca-certificates'
  cmd='/usr/sbin/update-ca-certificates'
  http_proxy='' && wget $certs_url -O $certs_folder/$certs_file
  unzip -u $certs_folder/$certs_file -d $certs_folder
  rm -f $certs_folder/$certs_file
  chmod 644 $certs_folder/*.crt
  /usr/sbin/update-ca-certificates
fi

if [[ "${PROVIDER_ID}" == "1" ]]; then
  registration_command_output=$(eval ${REGISTRATION_COMMAND})
  log_message "${registration_command_output}"
fi


log_message "The IDC K8aaS bootstrap script has completed."
