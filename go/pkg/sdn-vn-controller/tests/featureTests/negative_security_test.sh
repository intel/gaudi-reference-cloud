#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Function to parse and extract IDs from output
extract_id() {
  local output="$1"
  local pattern="$2"
  echo "$output" | grep "$pattern" | awk -F': ' '{print $2}'
}
status=0
# Function to check if the output contains the expected error message
check_error() {
  local output="$1"
  local expected_message="$2"
  if echo "$output" | grep -F "$expected_message"; then
    echo -e "\033[0;32mPASS: Expected error '$expected_message' found.\033[0m"
  else
    echo -e "\033[0;31mFAIL: Expected error '$expected_message' not found.\033[0m"
    echo "Actual output: $output"
    status=1
  fi
}

vpcId="123e4567-e89b-12d3-a456-426614174001"
invalid_vpcId="123e4567-e89b-12d3-a456-426614174002"
invalid_sg="123e4567-e89b-12d3-a456-426614174003"
invalid_port="123e4567-e89b-12d3-a456-426614174004"
invalid_subnet="123e4567-e89b-12d3-a456-426614174005"

# Capture both stdout and stderr to a single variable
output=$(./sdnctl create vpc $vpcId vpc1 1 1 2>&1)
echo "$output"

# Expect VPC already exists error
output=$(./sdnctl create vpc $vpcId vpc1 1 1 2>&1)
check_error "$output" "already exists"

# Create Subnet
output=$(./sdnctl create subnet test 10.0.0.0/24 $vpcId 2>&1)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")

# Create Port
output=$(./sdnctl create port "$subnet_id1" 2 12 12.0.0.1 2>&1)
echo "$output"
port_id1=$(extract_id "$output" "Port Created")

# Test creating a security group with an invalid VPC
output=$(./sdnctl create securitygroup "group1" $invalid_vpcId port "$port_id1" 2>&1)
check_error "$output" "is not present in table \"vpc\""


# Test creating a security group with an invalid port
output=$(./sdnctl create securitygroup "group1" $vpcId port "$port_id1,$invalid_port"  2>&1)
check_error "$output" "is not present in table \"port\""

# Test creating a security group with an invalid subnet
output=$(./sdnctl create securitygroup "group1" $vpcId subnet "$invalid_subnet,$subnet_id1" 2>&1)
check_error "$output" "is not present in table \"subnet\""

# Create Security Group
output=$(./sdnctl create securitygroup "group1" $vpcId port "$port_id1" 2>&1)
echo "$output"
security_group_id=$(extract_id "$output" "Security Group Created")

# Expect security group already exists error
output=$(./sdnctl create securitygroup "group1" $vpcId port "$port_id1" 2>&1)
check_error "$output" "already exists"

# Test creating a security rule with an invalid VPC
output=$(./sdnctl create securityrule "rule1" $invalid_vpcId $security_group_id 100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
check_error "$output" "is not present in table \"vpc\""

# Test creating a security rule with an invalid security group
output=$(./sdnctl create securityrule "rule1" $vpcId $invalid_sg 100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
check_error "$output" "No group found with UUID: $invalid_sg"

# Create Security Rule with valid params
output=$(./sdnctl create securityrule "rule1" $vpcId $security_group_id  100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
echo "$output"
security_rule_id1=$(extract_id "$output" "Security Rule Created")

# Expect security rule already exists error
output=$(./sdnctl create securityrule "rule1" $vpcId $security_group_id  100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
check_error "$output" "already exists"

# Execute delete commands

output=$(./sdnctl delete securitygroup "$security_group_id" 2>&1)
check_error "$output" "still referenced"

output=$(./sdnctl delete securityrule "$security_rule_id1" 2>&1)
echo "$output"

output=$(./sdnctl delete port "$port_id1" 2>&1)
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id1" 2>&1)
echo "$output"

output=$(./sdnctl delete securitygroup "$security_group_id" 2>&1)
echo "$output"

output=$(./sdnctl delete vpc "$vpcId" 2>&1)
echo "$output"

# Test errors after deletion

# Expect security rule not found error
output=$(./sdnctl delete securityrule "$security_rule_id1" 2>&1)
check_error "$output" "no rows in result set"

# Expect port not found error
output=$(./sdnctl delete port "$port_id1" 2>&1)
check_error "$output" "no rows in result set"

# Expect subnet not found error
output=$(./sdnctl delete subnet "$subnet_id1" 2>&1)
check_error "$output" "no rows in result set"

# Expect security group not found error
output=$(./sdnctl delete securitygroup "$security_group_id" 2>&1)
check_error "$output" "no rows in result set"

# Expect VPC not found error
output=$(./sdnctl delete vpc "$vpcId" 2>&1)
check_error "$output" "No vpc found with UUID"

exit $status