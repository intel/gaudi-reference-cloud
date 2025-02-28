#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Function to parse and extract IDs from output
extract_id() {
  local output="$1"
  local pattern="$2"
  echo "$output" | grep "$pattern" | awk -F': ' '{print $2}'
}

# Execute the command to create the first subnet
output=$(./sdnctl create subnet test 10.0.0.0/24)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")

# Execute the command to create the first port in the first subnet
output=$(./sdnctl create port "$subnet_id1" 1 8 10.0.0.1)
echo "$output"
port_id1=$(extract_id "$output" "Port Created")

# Execute the command to create the second port in the first subnet
output=$(./sdnctl create port "$subnet_id1" 1 9 10.0.0.2)
echo "$output"
port_id2=$(extract_id "$output" "Port Created")

# Execute the command to create a router
output=$(./sdnctl create router r1)
echo "$output"
router_id1=$(extract_id "$output" "Router is")

# Execute the command to create a router interface for the first subnet
output=$(./sdnctl create router interface "$router_id1" "$subnet_id1" 10.0.0.1/24 00:11:22:33:44:55)
echo "$output"
router_interface_id1=$(extract_id "$output" "Router interface uuid is")

# Execute the command to update the first port
output=$(./sdnctl update port "$port_id1" 1 1)
echo "$output"
updated_port_id1=$(extract_id "$output" "Port updated")

# Execute the command to update the second port
output=$(./sdnctl update port "$port_id2" 1 1)
echo "$output"
updated_port_id2=$(extract_id "$output" "Port updated")

# Execute the command to create the second subnet
output=$(./sdnctl create subnet test2 20.0.0.0/24)
echo "$output"
subnet_id2=$(extract_id "$output" "Subnet Created is")

# Execute the command to create the first port in the second subnet
output=$(./sdnctl create port "$subnet_id2" 2 10 20.0.0.1)
echo "$output"
port_id3=$(extract_id "$output" "Port Created")

# Execute the command to create the second port in the second subnet
output=$(./sdnctl create port "$subnet_id2" 2 11 20.0.0.2)
echo "$output"
port_id4=$(extract_id "$output" "Port Created")

# Execute the command to create a router interface for the second subnet
output=$(./sdnctl create router interface "$router_id1" "$subnet_id2" 20.0.0.1/24 00:11:22:33:44:66)
echo "$output"
router_interface_id2=$(extract_id "$output" "Router interface uuid is")

# Execute the command to update the first port in the second subnet
output=$(./sdnctl update port "$port_id3" 1 1)
echo "$output"
updated_port_id3=$(extract_id "$output" "Port updated")

# Execute the command to update the second port in the second subnet
output=$(./sdnctl update port "$port_id4" 1 1)
echo "$output"
updated_port_id4=$(extract_id "$output" "Port updated")

# Execute the command to create the third subnet
output=$(./sdnctl create subnet test3 32.0.0.0/24)
echo "$output"
subnet_id3=$(extract_id "$output" "Subnet Created is")

# Execute the command to create the first port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 3 12 32.0.0.1)
echo "$output"
port_id5=$(extract_id "$output" "Port Created")

# Execute the command to create the second port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 3 13 32.0.0.2)
echo "$output"
port_id6=$(extract_id "$output" "Port Created")

# Execute the command to create a second router
output=$(./sdnctl create router r2)
echo "$output"
router_id2=$(extract_id "$output" "Router is")

# Execute the command to create a router interface for the third subnet
output=$(./sdnctl create router interface "$router_id2" "$subnet_id3" 32.0.0.1/24 00:11:22:33:44:77)
echo "$output"
router_interface_id3=$(extract_id "$output" "Router interface uuid is")

# Execute the command to update the first port in the third subnet
output=$(./sdnctl update port "$port_id5" 1 1)
echo "$output"
updated_port_id5=$(extract_id "$output" "Port updated")

# Execute the command to update the second port in the third subnet
output=$(./sdnctl update port "$port_id6" 1 1)
echo "$output"
updated_port_id6=$(extract_id "$output" "Port updated")

# Execute the command to create the fourth subnet
output=$(./sdnctl create subnet test4 40.0.0.0/24)
echo "$output"
subnet_id4=$(extract_id "$output" "Subnet Created is")

# Execute the command to create the first port in the fourth subnet
output=$(./sdnctl create port "$subnet_id4" 4 7 40.0.0.1)
echo "$output"
port_id7=$(extract_id "$output" "Port Created")

# Execute the command to create the third router
output=$(./sdnctl create router r3)
echo "$output"
router_id3=$(extract_id "$output" "Router is")

# Execute the command to create a router interface for the fourth subnet
output=$(./sdnctl create router interface "$router_id3" "$subnet_id4" 40.0.0.1/24 00:11:22:33:44:88)
echo "$output"
router_interface_id4=$(extract_id "$output" "Router interface uuid is")

# Optional: Execute the command to update the new port (if needed)
output=$(./sdnctl update port "$port_id7" 1 1)
echo "$output"
updated_port_id7=$(extract_id "$output" "Port updated")


# #RESET NATs
output=$(./sdnctl update port "$port_id1" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id2" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id3" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id4" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id5" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id6" 0 1)
echo "$output"  
updated_port_id=$(extract_id "$output" "Port updated")

output=$(./sdnctl update port "$port_id7" 0 1)
echo "$output"
updated_port_id=$(extract_id "$output" "Port updated")