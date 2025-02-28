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
    status=1
  fi
}

vpcId="123e4567-e89b-12d3-a456-426614174001"

# Create VPC
output=$(./sdnctl create vpc $vpcId vpc1 1 1)
echo "$output"
check_db_query "SELECT vpc_id FROM vpc WHERE vpc_id='$vpcId';" "$vpcId"

# Create first subnet
output=$(./sdnctl create subnet test1 10.0.0.0/24 $vpcId)
echo "$output"
subnet_id1=$(extract_id "$output" "Subnet Created is")
check_db_query "SELECT subnet_id FROM subnet WHERE subnet_id='$subnet_id1';" "$subnet_id1"

# Create second subnet
output=$(./sdnctl create subnet test2 11.0.0.0/24 $vpcId)
echo "$output"
subnet_id2=$(extract_id "$output" "Subnet Created is")
check_db_query "SELECT subnet_id FROM subnet WHERE subnet_id='$subnet_id2';" "$subnet_id2"

# Create third subnet
output=$(./sdnctl create subnet test3 12.0.0.0/24 $vpcId)
echo "$output"
subnet_id3=$(extract_id "$output" "Subnet Created is")
check_db_query "SELECT subnet_id FROM subnet WHERE subnet_id='$subnet_id3';" "$subnet_id3"

# Create first port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 2 12 12.0.0.1)
echo "$output"
port_id1=$(extract_id "$output" "Port Created")
check_db_query "SELECT port_id FROM port WHERE port_id='$port_id1';" "$port_id1"

# Create second port in the third subnet
output=$(./sdnctl create port "$subnet_id3" 3 10 12.0.0.2)
echo "$output"
port_id2=$(extract_id "$output" "Port Created")
check_db_query "SELECT port_id FROM port WHERE port_id='$port_id2';" "$port_id2"

output=$(./sdnctl create port "$subnet_id3" 3 12 12.0.0.3)
echo "$output"
port_id3=$(extract_id "$output" "Port Created")
check_db_query "SELECT port_id FROM port WHERE port_id='$port_id3';" "$port_id3"

# Create first security group with the first two subnets and the first two ports
output=$(./sdnctl create securitygroup "group1" $vpcId port "$port_id1,$port_id3")
echo "$output"
security_group_id1=$(extract_id "$output" "Security Group Created")
check_db_query "SELECT security_group_id FROM security_group WHERE security_group_id='$security_group_id1';" "$security_group_id1"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id1';" "$port_id1|$security_group_id1"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id3' AND security_group_id='$security_group_id1';" "$port_id3|$security_group_id1"

# Create the first security rule
output=$(./sdnctl create securityrule "rule1" $vpcId $security_group_id1 100 egress deny "10.0.0.0/24,11.0.0.0/24" "192.168.3.0/24,192.168.2.0/24" udp 11 900)
echo "$output"
security_rule_id1=$(extract_id "$output" "Security Rule Created")
check_db_query "SELECT security_group_id,security_rule_id FROM  security_rule WHERE security_group_id='$security_group_id1' AND security_rule_id='$security_rule_id1';" "$security_group_id1|$security_rule_id1"
check_db_query "SELECT security_rule_id FROM acl WHERE security_rule_id='$security_rule_id1';" "$security_rule_id1"
check_db_query "SELECT security_rule_priority FROM security_rule WHERE security_rule_id='$security_rule_id1';" "100"
check_db_query "SELECT direction FROM security_rule WHERE security_rule_id='$security_rule_id1';" "2"
check_db_query "SELECT security_rule_action FROM security_rule WHERE security_rule_id='$security_rule_id1';" "2"
check_db_query "SELECT protocol FROM security_rule WHERE security_rule_id='$security_rule_id1';" "1"
check_db_query "SELECT source_ip_addresses FROM security_rule WHERE security_rule_id='$security_rule_id1';" "{10.0.0.0/24,11.0.0.0/24}"
check_db_query "SELECT destination_ip_addresses FROM security_rule WHERE security_rule_id='$security_rule_id1';" "{192.168.3.0/24,192.168.2.0/24}"
check_db_query "SELECT source_port FROM security_rule WHERE security_rule_id='$security_rule_id1';" "[11,901)"

# Update the first security rule
output=$(./sdnctl update securityrule $security_rule_id1 200 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443)
echo "$output"

# Verify that the updated values are present in the database
check_db_query "SELECT security_rule_priority FROM security_rule WHERE security_rule_id='$security_rule_id1';" "200"
check_db_query "SELECT direction FROM security_rule WHERE security_rule_id='$security_rule_id1';" "1"
check_db_query "SELECT security_rule_action FROM security_rule WHERE security_rule_id='$security_rule_id1';" "1"
check_db_query "SELECT protocol FROM security_rule WHERE security_rule_id='$security_rule_id1';" "0"
check_db_query "SELECT source_ip_addresses FROM security_rule WHERE security_rule_id='$security_rule_id1';" "{10.0.0.0/24}"
check_db_query "SELECT destination_ip_addresses FROM security_rule WHERE security_rule_id='$security_rule_id1';" "{192.168.1.0/24,192.168.2.0/24}"
check_db_query "SELECT source_port FROM security_rule WHERE security_rule_id='$security_rule_id1';" "[80,444)"


# Create the second security rule
output=$(./sdnctl create securityrule "rule2" $vpcId $security_group_id1 100 egress deny "10.0.1.0/24,10.0.2.0/24" "192.168.3.0/24" udp 90 100)
echo "$output"
security_rule_id2=$(extract_id "$output" "Security Rule Created")
check_db_query "SELECT security_group_id,security_rule_id FROM security_rule WHERE security_group_id='$security_group_id1' AND security_rule_id='$security_rule_id2';" "$security_group_id1|$security_rule_id2"



output=$(./sdnctl update securitygroup $security_group_id1 port "$port_id1,$port_id2" )
echo "$output"
# Verify port_security_group references
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id1' AND security_group_id='$security_group_id1';" "$port_id1|$security_group_id1"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id2' AND security_group_id='$security_group_id1';" "$port_id2|$security_group_id1"

# Create second security group with two subnets
output=$(./sdnctl create securitygroup "group2" $vpcId subnet "$subnet_id1,$subnet_id2")
echo "$output"
security_group_id2=$(extract_id "$output" "Security Group Created")
check_db_query "SELECT security_group_id FROM security_group WHERE security_group_id='$security_group_id2';" "$security_group_id2"

check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id1';" "$subnet_id1|$security_group_id2"
check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id2';" "$subnet_id2|$security_group_id2"

# Create the 3rd security rule
output=$(./sdnctl create securityrule "rule3" $vpcId $security_group_id1 100 egress deny "10.0.1.0/24,10.0.2.0/24" "192.168.3.0/24" udp 40 100)
echo "$output"
security_rule_id3=$(extract_id "$output" "Security Rule Created")

# Create 3d security group with two ports, port1 is shared shared with group1
output=$(./sdnctl create securitygroup "group3" $vpcId port "$port_id1,$port_id3")
echo "$output"
security_group_id3=$(extract_id "$output" "Security Group Created")
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id1' AND security_group_id='$security_group_id1';" "$port_id1|$security_group_id1"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id2' AND security_group_id='$security_group_id1';" "$port_id2|$security_group_id1"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id1' AND security_group_id='$security_group_id3';" "$port_id1|$security_group_id3"
check_db_query "SELECT port_id,security_group_id FROM port_security_group WHERE port_id='$port_id3' AND security_group_id='$security_group_id3';" "$port_id3|$security_group_id3"

# Create 4d security group with two ports, port1 is shared shared with group2
output=$(./sdnctl create securitygroup "group4" $vpcId subnet "$subnet_id1,$subnet_id3")
echo "$output"
security_group_id4=$(extract_id "$output" "Security Group Created")
check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id1' AND security_group_id='$security_group_id2';" "$subnet_id1|$security_group_id2"
check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id2' AND security_group_id='$security_group_id2';" "$subnet_id2|$security_group_id2"
check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id1' AND security_group_id='$security_group_id4';" "$subnet_id1|$security_group_id4"
check_db_query "SELECT subnet_id,security_group_id FROM subnet_security_group WHERE subnet_id='$subnet_id3' AND security_group_id='$security_group_id4';" "$subnet_id3|$security_group_id4"


#Get security components 
output=$(./sdnctl list securityrule)
echo "$output"

output=$(./sdnctl get securityrule "$security_rule_id1")
echo "$output"

output=$(./sdnctl get securityrule "$security_rule_id2")
echo "$output"

output=$(./sdnctl get securityrule "$security_rule_id3")
echo "$output"

output=$(./sdnctl list securitygroup)
echo "$output"

output=$(./sdnctl get securitygroup "$security_group_id1")
echo "$output"

output=$(./sdnctl get securitygroup "$security_group_id2")
echo "$output"

output=$(./sdnctl get securitygroup "$security_group_id3")
echo "$output"

output=$(./sdnctl get securitygroup "$security_group_id4")
echo "$output"

# Execute delete commands
output=$(./sdnctl delete securityrule "$security_rule_id1")
echo "$output"
output=$(./sdnctl delete securityrule "$security_rule_id2")
echo "$output"
output=$(./sdnctl delete securityrule "$security_rule_id3")
echo "$output"
output=$(./sdnctl delete securityrule "$security_rule_id4")
echo "$output"
check_db_query "SELECT COUNT(*) FROM security_rule WHERE security_rule_id='$security_rule_id1';" "0"
check_db_query "SELECT COUNT(*) FROM acl WHERE security_rule_id='$security_rule_id1';" "0"
check_db_query "SELECT COUNT(*) FROM security_rule WHERE security_rule_id='$security_rule_id2';" "0"
check_db_query "SELECT COUNT(*) FROM acl WHERE security_rule_id='$security_rule_id2';" "0"
check_db_query "SELECT COUNT(*) FROM security_rule WHERE security_rule_id='$security_rule_id3';" "0"
check_db_query "SELECT COUNT(*) FROM acl WHERE security_rule_id='$security_rule_id3';" "0"

output=$(./sdnctl delete port "$port_id1")
echo "$output"
check_db_query "SELECT COUNT(*) FROM port WHERE port_id='$port_id1';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE port_id='$port_id1';" "0"

output=$(./sdnctl delete subnet "$subnet_id1")
echo "$output"
check_db_query "SELECT COUNT(*) FROM subnet WHERE subnet_id='$subnet_id1';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE subnet_id='$subnet_id1';" "0"

output=$(./sdnctl delete securitygroup "$security_group_id1")
echo "$output"
check_db_query "SELECT COUNT(*) FROM security_group WHERE security_group_id='$security_group_id1';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE security_group_id='$security_group_id1';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE security_group_id='$security_group_id1';" "0"

output=$(./sdnctl delete securitygroup "$security_group_id2")
echo "$output"
check_db_query "SELECT COUNT(*) FROM security_group WHERE security_group_id='$security_group_id2';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE security_group_id='$security_group_id2';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE security_group_id='$security_group_id2';" "0"

output=$(./sdnctl delete securitygroup "$security_group_id3")
echo "$output"
check_db_query "SELECT COUNT(*) FROM security_group WHERE security_group_id='$security_group_id3';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE security_group_id='$security_group_id3';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE security_group_id='$security_group_id3';" "0"

output=$(./sdnctl delete securitygroup "$security_group_id4")
echo "$output"
check_db_query "SELECT COUNT(*) FROM security_group WHERE security_group_id='$security_group_id4';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE security_group_id='$security_group_id4';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE security_group_id='$security_group_id4';" "0"

output=$(./sdnctl delete port "$port_id2")
echo "$output"
check_db_query "SELECT COUNT(*) FROM port WHERE port_id='$port_id2';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE port_id='$port_id2';" "0"

output=$(./sdnctl delete port "$port_id3")
echo "$output"
check_db_query "SELECT COUNT(*) FROM port WHERE port_id='$port_id3';" "0"
check_db_query "SELECT COUNT(*) FROM port_security_group WHERE port_id='$port_id3';" "0"

output=$(./sdnctl delete subnet "$subnet_id2")
echo "$output"
check_db_query "SELECT COUNT(*) FROM subnet WHERE subnet_id='$subnet_id2';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE subnet_id='$subnet_id2';" "0"

output=$(./sdnctl delete subnet "$subnet_id3")
echo "$output"
check_db_query "SELECT COUNT(*) FROM subnet WHERE subnet_id='$subnet_id3';" "0"
check_db_query "SELECT COUNT(*) FROM subnet_security_group WHERE subnet_id='$subnet_id3';" "0"

output=$(./sdnctl delete vpc "$vpcId")
echo "$output"
check_db_query "SELECT COUNT(*) FROM vpc WHERE vpc_id='$vpcId';" "0"

exit $status
