#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
#
# Bring up or down a VLAN sub-interface of a bridge and use the host as the default
# gateway for the subnet assigned to the VLAN.  The use case is to enable tenant or
# other VLANs to be routed succesfully with a KIND deployment (or more generally,
# anyone attached to the kind bridge).
#
# vlan up SUBNET_CIDR VLAN_ID BRIDGE_DEVICE_NAME
# vlan down SUBNET_CIDR VLAN_ID BRIDGE_DEVICE_NAME
set -eu -o pipefail
set -x

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

# vlan_up 192.168.201.0/24 201 br-071bfa8aef5d
vlan_up() {
  local -r subnet=$1
  local -r vlan=$2
  local -r bridge=$3
  local -r dev_prefix="${bridge:0:10}."
  local -r dev="${dev_prefix}${vlan}"
  local -r addr=$(ipcalc ${subnet} | awk '/HostMin:/ {print $2}') # Gateway address, typically a.b.c.1
  local -r class=$(ipcalc --class ${subnet})

  if ip link show dev ${dev}; then
    echo "Device ${dev} already exists. Skipping configuration."
    return
  else
    sudo ip link add link ${bridge} name ${dev} type vlan id ${vlan}
    sudo ip addr add ${addr}/${class} dev ${dev}
    sudo ip link set dev ${dev} up

    # The DOCKER-USER rule could use a single rule with ${dev_prefix}+ for both in and out arguments,
    # but specializing the first argument makes cleanup easier
    sudo iptables -I DOCKER-USER -i ${dev} ! -o ${dev_prefix}+ -j ACCEPT
    sudo iptables -I DOCKER-USER -o ${dev} ! -i ${dev_prefix}+ -j ACCEPT
    sudo iptables -t nat -A POSTROUTING -s ${subnet} ! -o ${dev} -j MASQUERADE
  fi
}

# vlan_down 192.168.201.0/24 201 br-071bfa8aef5d
vlan_down() {
  local -r subnet=$1
  local -r vlan=$2
  local -r bridge=$3
  local -r dev_prefix="${bridge:0:10}."
  local -r dev="${dev_prefix}${vlan}"

  if ! ip link show dev ${dev}; then
    echo "Device ${dev} does not exist. Skipping unconfiguration."
    return
  else
    sudo iptables -t nat -D POSTROUTING -s ${subnet} ! -o ${dev} -j MASQUERADE || true
    sudo iptables -D DOCKER-USER -i ${dev} ! -o ${dev_prefix}+ -j ACCEPT || true
    sudo iptables -D DOCKER-USER -o ${dev} ! -i ${dev_prefix}+ -j ACCEPT || true

    sudo ip link del name ${dev}
  fi
}

case $1 in
  up) vlan_up $2 $3 $4 ;;
  down) vlan_down $2 $3 $4 ;;
esac
