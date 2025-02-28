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


vpcId="123e4567-e89b-12d3-a456-426614174001"

output=$(./sdnctl create vpc $vpcId vpc1 1 1)
echo "$output"

# Create the first subnet
output=$(./sdnctl create subnet test 10.0.0.0/24 $vpcId)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")

# Create the first port in the first subnet
output=$(./sdnctl create port "$subnet_id1" 1 8 10.0.0.1)
echo "$output"
port_id1=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "f"

# Create the second port in the first subnet
output=$(./sdnctl create port "$subnet_id1" 1 9 10.0.0.2)
echo "$output"
port_id2=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "f"

# Create the first router
output=$(./sdnctl create router r1 $vpcId)
echo "$output"
router_id1=$(extract_id "$output" "Router is")

# Create a router interface for the first subnet
output=$(./sdnctl create router interface "$router_id1" "$subnet_id1" 10.0.0.1/24 00:11:22:33:44:55)
echo "$output"
router_interface_id1=$(extract_id "$output" "Router interface uuid is")

# Update the first port and check the gateway
output=$(./sdnctl update port "$port_id1" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

# Update the second port and check the gateway
output=$(./sdnctl update port "$port_id2" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

# Create the second subnet
output=$(./sdnctl create subnet test2 20.0.0.0/24 $vpcId)
echo "$output"
subnet_id2=$(extract_id "$output" "Subnet Created is")

# Create the first port in the second subnet
output=$(./sdnctl create port "$subnet_id2" 2 10 20.0.0.1)
echo "$output"
port_id3=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id3';" "f"

# Create the second port in the second subnet
output=$(./sdnctl create port "$subnet_id2" 2 11 20.0.0.2)
echo "$output"
port_id4=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id4';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id4';" "f"

# Create a router interface for the second subnet
output=$(./sdnctl create router interface "$router_id1" "$subnet_id2" 20.0.0.1/24 00:11:22:33:44:66)
echo "$output"
router_interface_id2=$(extract_id "$output" "Router interface uuid is")

# Update the first port in the second subnet and check the gateway
output=$(./sdnctl update port "$port_id3" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

# Update the second port in the second subnet and check the gateway
output=$(./sdnctl update port "$port_id4" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id4';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

# Create the third subnet
output=$(./sdnctl create subnet test3 32.0.0.0/24 $vpcId)
echo "$output"
subnet_id3=$(extract_id "$output" "Subnet Created is")

# Create the first port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 3 12 32.0.0.1)
echo "$output"
port_id5=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id5';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id5';" "f"

# Create the second port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 3 13 32.0.0.2)
echo "$output"
port_id6=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id6';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id6';" "f"

# Create a second router
output=$(./sdnctl create router r2 $vpcId)
echo "$output"
router_id2=$(extract_id "$output" "Router is")

# Create a router interface for the third subnet
output=$(./sdnctl create router interface "$router_id2" "$subnet_id3" 32.0.0.1/24 00:11:22:33:44:77)
echo "$output"
router_interface_id3=$(extract_id "$output" "Router interface uuid is")

# Update the first port in the third subnet and check the gateway
output=$(./sdnctl update port "$port_id5" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id5';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id5';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id2';" "t"

# Update the second port in the third subnet and check the gateway
output=$(./sdnctl update port "$port_id6" 1 1)
echo "$output"
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id6';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id6';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id2';" "t"

# Create the fourth subnet
output=$(./sdnctl create subnet test4 40.0.0.0/24 $vpcId)
echo "$output"
subnet_id4=$(extract_id "$output" "Subnet Created is")

# Create the third router
output=$(./sdnctl create router r3 $vpcId)
echo "$output"
router_id3=$(extract_id "$output" "Router is")

# Create a router interface for the fourth subnet
output=$(./sdnctl create router interface "$router_id3" "$subnet_id4" 40.0.0.1/24 00:11:22:33:44:88)
echo "$output"
router_interface_id4=$(extract_id "$output" "Router interface uuid is")

# Create the first port in the fourth subnet with isNAT true and check the gateway
output=$(./sdnctl create port "$subnet_id4" 4 7 40.0.0.2 true)
echo "$output"
port_id7=$(extract_id "$output" "Port Created")
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id7';" "t"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id7';" "t"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id3';" "t"

# RESET NATs for the first router
output=$(./sdnctl update port "$port_id1" 0 1)
echo "$output"
# Verify that gateway_id is still present when not all NATs are reset
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id1';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id1';" "f"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

output=$(./sdnctl update port "$port_id2" 0 1)
echo "$output"
# Verify that gateway_id is still present when not all NATs are reset
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id2';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id2';" "f"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

output=$(./sdnctl update port "$port_id3" 0 1)
echo "$output"
# Verify that gateway_id is still present when not all NATs are reset
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id3';" "f"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id1';" "t"

output=$(./sdnctl update port "$port_id4" 0 1)
echo "$output"

# Verify that gateway_id is set to NULL after resetting NATs
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id4';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id4';" "f"
check_db_query "SELECT gateway_id FROM router WHERE router_id='$router_id1';" ""

# RESET NATs for the second router
output=$(./sdnctl update port "$port_id5" 0 1)
echo "$output"
# Verify that gateway_id is still present when not all NATs are reset
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id5';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id5';" "f"
check_db_query "SELECT gateway_id IS NOT NULL FROM router WHERE router_id='$router_id2';" "t"

output=$(./sdnctl update port "$port_id6" 0 1)
echo "$output"

# Verify that gateway_id is set to NULL after resetting NATs
check_db_query "SELECT isnat FROM port WHERE port_id='$port_id6';" "f"
check_db_query "SELECT snat_rule_id IS NOT NULL FROM port WHERE port_id='$port_id5';" "f"
check_db_query "SELECT gateway_id FROM router WHERE router_id='$router_id2';" ""

# DELETE ports and reset NATs for the third router
output=$(./sdnctl delete port "$port_id7")
echo "$output"

# Verify that gateway_id is set to NULL after resetting NATs
check_db_query "SELECT gateway_id FROM router WHERE router_id='$router_id3';" ""

output=$(./sdnctl delete port "$port_id6" )
echo "$output"

output=$(./sdnctl delete port "$port_id5" )
echo "$output"

output=$(./sdnctl delete port "$port_id4" )
echo "$output"

output=$(./sdnctl delete port "$port_id3" )
echo "$output"

output=$(./sdnctl delete port "$port_id2" )
echo "$output"

output=$(./sdnctl delete port "$port_id1" )
echo "$output"


output=$(./sdnctl delete router interface "$router_interface_id1" )
echo "$output"

output=$(./sdnctl delete router interface "$router_interface_id2" )
echo "$output"

output=$(./sdnctl delete router interface "$router_interface_id3" )
echo "$output"

output=$(./sdnctl delete router interface "$router_interface_id4" )
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id1")
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id2")
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id3")
echo "$output"

output=$(./sdnctl delete subnet "$subnet_id4")
echo "$output"

output=$(./sdnctl delete router "$router_id1")
echo "$output"

output=$(./sdnctl delete router "$router_id2")
echo "$output"

output=$(./sdnctl delete router "$router_id3")
echo "$output"

output=$(./sdnctl delete vpc "$vpcId")