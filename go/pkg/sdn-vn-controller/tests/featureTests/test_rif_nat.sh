#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# TODO: More descriptions and the setup guide

#!/bin/bash

# Function to parse and extract IDs from output
extract_id() {
  local output="$1"
  local pattern="$2"
  echo "$output" | grep "$pattern" | awk -F': ' '{print $2}'
}

# Function to run a query and check for expected results
check_db_query() {
  local query="$1"
  local expected="$2"
  local result
  result=$(docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -t -c "$query" | tr -d '[:space:]')
  if [[ "$result" == *"$expected"* ]]; then
    echo -e "\033[0;32mPASS: Expected result found in the database.\033[0m"
  else
    echo -e "\033[0;31mFAIL: Expected result not found in the database.\033[0m"
    echo "Query result: $result"
  fi
}

# Function to check if the output contains the expected error message
check_error() {
  local output="$1"
  local expected_message="$2"
  if echo "$output" | grep -F "$expected_message"; then
    echo -e "\033[0;32mPASS: Expected error '$expected_message' found.\033[0m"
  else
    echo -e "\033[0;31mFAIL: Expected error '$expected_message' not found.\033[0m"
    echo "Actual output: $output"
  fi
}

# Get UUIDs for gw-1 and gw-2
output=$(docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT gateway_id, gateway_name FROM gateway;")
gw1_uuid=$(echo "$output" | grep 'gw-1' | awk '{print $1}')
gw2_uuid=$(echo "$output" | grep 'gw-2' | awk '{print $1}')

# Print the UUIDs
echo "UUID for gw-1: $gw1_uuid"
echo "UUID for gw-2: $gw2_uuid"

vpcId="123e4567-e89b-12d3-a456-426614174001"

output=$(./sdnctl create vpc $vpcId vpc1 1 1)
echo "$output"

# Create the first subnet
output=$(./sdnctl create subnet test 10.0.0.0/24 $vpcId)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")

# Create the first router
output=$(./sdnctl create router r1 $vpcId)
echo "$output"
router_id1=$(extract_id "$output" "Router is")

# Create a router interface for the first subnet
output=$(./sdnctl create router interface "$router_id1" "$subnet_id1" 10.0.0.1/24 00:11:22:33:44:55)
echo "$output"
router_interface_id1=$(extract_id "$output" "Router interface uuid is")
check_db_query "SELECT router_interface_id FROM router_interface WHERE router_interface_id='$router_interface_id1';" "$security_rule_id2"

# Create the first router
output=$(./sdnctl create router r2 $vpcId)
echo "$output"
router_id2=$(extract_id "$output" "Router is")

output=$(./sdnctl create router interface "$router_id2" "$subnet_id1" 10.0.0.7/24 00:11:22:33:44:55 2>&1)
check_error "$output" "Key (subnet_id)=($subnet_id1) already exists"


# Create the first port 
output=$(./sdnctl create port "$subnet_id1" 1 8 10.0.0.1 true)
echo "$output"
port_id1=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

# Create the second port
output=$(./sdnctl create port "$subnet_id1" 1 9 10.0.0.2)
echo "$output"
port_id2=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "f"
check_db_query "SELECT snat_rule_id FROM port WHERE port_id='$port_id2';" ""

# Update the second port
output=$(./sdnctl update port "$port_id2" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id2';" "t"

# Create the third port
output=$(./sdnctl create port "$subnet_id1" 1 10 10.0.0.2)
echo "$output"
port_id3=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT snat_rule_id FROM port WHERE port_id='$port_id3';" ""

output=$(./sdnctl delete router interface "$router_interface_id1" )
echo "$output"

check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "t"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT snat_rule_id FROM port WHERE port_id='$port_id1';" ""
check_db_query "SELECT snat_rule_id FROM port WHERE port_id='$port_id2';" ""
check_db_query "SELECT snat_rule_id FROM port WHERE port_id='$port_id3';" ""
check_db_query "SELECT gateway_id FROM router WHERE router_id='$router_id1';" ""

# Rereate the same router interface again and verify the NATs are set
output=$(./sdnctl create router interface "$router_id1" "$subnet_id1" 10.0.0.1/24 00:11:22:33:44:55)
echo "$output"
router_interface_id1=$(extract_id "$output" "Router interface uuid is")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "t"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id2';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

output=$(./sdnctl delete port "$port_id1" )
echo "$output"

output=$(./sdnctl delete port "$port_id2" )
echo "$output"

output=$(./sdnctl delete port "$port_id3" )
echo "$output"

output=$(./sdnctl delete router interface "$router_interface_id1" )
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id1")
echo "$output"

output=$(./sdnctl delete router "$router_id1")
echo "$output"

output=$(./sdnctl delete router "$router_id2")
echo "$output"


output=$(./sdnctl delete vpc "$vpcId")