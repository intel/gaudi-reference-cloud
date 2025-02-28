# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Declare variables
ovnCentralIP="100.64.20.55"
CHASSIS="100.64.20.100"
IPU_ID="102E00007E90"
VF1="enp0s1f0d2"
VF2="enp0s1f0d3"

extract_id() {
  local output="$1"
  local pattern="$2"
  echo "$output" | grep "$pattern" | awk -F': ' '{print $2}'
}

docker-compose -f assets/docker/no_ovn.yaml down
OVN_IP="$ovnCentralIP" docker-compose -f assets/docker/no_ovn.yaml up --detach

# Configure the VFs
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns add vm1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns add vm2
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip link set $VF1 netns vm1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip link set $VF2 netns vm2
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm1 ip link set $VF1 up
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm2 ip link set $VF2 up
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm1 ip addr add 12.0.0.1/24 dev $VF1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm2 ip addr add 12.0.0.2/24 dev $VF2

# Create topo
vpcId="123e4567-e89b-12d3-a456-426614174001"
./sdnctl create vpc $vpcId vpc1 1 1
output=$(./sdnctl create subnet test1 12.0.0.0/24 $vpcId)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")
output=$(./sdnctl create port "$subnet_id1" $IPU_ID 3 12.0.0.1 "10:2e:00:00:7e:93")
echo "$output"
port_id1=$(extract_id "$output" "Port Created")
output=$(./sdnctl create port "$subnet_id1" $IPU_ID 4 12.0.0.2 "00:0a:00:03:39:18")
echo "$output"
port_id2=$(extract_id "$output" "Port Created")

ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR root@"$CHASSIS" ip netns exec vm1 ping -c 4 12.0.0.2

# Cleanup
./sdnctl delete port $port_id1
./sdnctl delete port $port_id2
./sdnctl delete subnet $subnet_id1
./sdnctl delete vpc $vpcId

ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm1 ip link set $VF1 netns 1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns exec vm2 ip link set $VF2 netns 1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns delete vm1
ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@"$CHASSIS" ip netns delete vm2
docker-compose -f assets/docker/no_ovn.yaml down

