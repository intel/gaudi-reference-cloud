#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Create necessary directories
mkdir -p /var/run/openvswitch
mkdir -p /var/log/openvswitch
mkdir -p /var/lib/openvswitch
mkdir -p /var/lib/ovn
mkdir -p /var/log/ovn
mkdir -p /var/run/ovn

# Remove old OVS database and create a new one
rm -rf /var/lib/openvswitch/conf.db
ovsdb-tool create /var/lib/openvswitch/conf.db /usr/share/openvswitch/vswitch.ovsschema

# Start the OVSDB server with adjusted logging
ovsdb-server /var/lib/openvswitch/conf.db \
    -vANY:INFO \
    -vconsole:WARN \
    -vreconnect:WARN \
    --remote=punix:/var/run/openvswitch/db.sock \
    --private-key=db:Open_vSwitch,SSL,private_key \
    --certificate=db:Open_vSwitch,SSL,certificate \
    --bootstrap-ca-cert=db:Open_vSwitch,SSL,ca_cert \
    --no-chdir \
    --log-file=/var/log/openvswitch/ovsdb-server.log \
    --pidfile=/var/run/openvswitch/ovsdb-server.pid \
    --detach

unset http_proxy
unset https_proxy

# Start OVS switch daemon with adjusted logging
ovs-vswitchd unix:/var/run/openvswitch/db.sock \
    -vANY:INFO \
    -vconsole:WARN \
    -vreconnect:WARN \
    --log-file=/var/log/openvswitch/ovs-vswitchd.log \
    --pidfile \
    --detach

set -x

# Initialize variables
PORT_COUNT=6
ENCAP_VLAN=$(cat /etc/ovn/encap_vlan 2>/dev/null || echo "")
HOSTNAME=n1
ENCAP_IP=10.0.0.2
BRIDGE_NAME=br-int

# Set up OVS bridge and external-ids for OVN
ovs-vsctl add-br ${BRIDGE_NAME}
ovs-vsctl set open . external-ids:ovn-encap-type=geneve
ovs-vsctl set open . external-ids:ovn-encap-ip=${ENCAP_IP}
ovs-vsctl set open . external-ids:ovn-bridge=${BRIDGE_NAME}
ovs-vsctl set open . external-ids:system-id=${HOSTNAME}
ovs-vsctl set open . external-ids:ovn-remote=tcp:${OVN_CENTRAL_IP}:6642

# Create veth pairs
for PORT in $(seq 1 $PORT_COUNT); do
    ip link delete ovnp_${PORT} || true
    ip link add ovnp_${PORT} type veth peer name port${PORT}
    ip link set ovnp_${PORT} up
    ip link set port${PORT} up
done

# Add ports to the OVS bridge
for PORT in $(seq 1 $PORT_COUNT); do
    ovs-vsctl add-port ${BRIDGE_NAME} ovnp_${PORT}
    ovs-vsctl set Interface ovnp_${PORT} external_ids:iface-id=${HOSTNAME}@port${PORT}
done

# Start OVN controller in the background with adjusted logging
ovn-controller unix:/var/run/openvswitch/db.sock \
    -vANY:INFO \
    -vconsole:WARN \
    -vreconnect:WARN \
    --no-chdir \
    --log-file=/var/log/ovn/ovn-controller.log \
    --pidfile=/var/run/ovn/ovn-controller.pid \
    --detach

tail -f /dev/null
