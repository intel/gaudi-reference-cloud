#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation


# Description:
# This bash script demonstrates the creation, configuration, and teardown of a virtual network topology using the CMD client (`sdnctl` commands). 
# It begins by creating essential network components such as VPCs, subnets, routers, interfaces, and ports, then defines 
# security rules and groups, the NAT rules. Throughout the process, it accesses PostgreSQL and OVN databases to validate 
# the state of the networking components, ensuring that resources are properly configured.
# The script proceeds to clean up by deleting all created components, returning the system to its initial state.

# Function to extract IDs from output
extract_id() {
  local output="$1"
  local pattern="$2"
  echo "$output" | grep "$pattern" | awk -F': ' '{print $2}'
}

print_section_title() {
  echo
  echo -e "\033[0;32m============================================================================\033[0m"
  echo -e "\033[0;32m$1\033[0m"
  echo -e "\033[0;32m============================================================================\033[0m"
  echo
}

print_subsection_title() {
  echo
  echo -e "\033[0;34m_____________________________\033[0m"
  echo -e "\033[0;34m$1\033[0m"
  echo -e "\033[0;34m_____________________________\033[0m"
  echo
}

docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM gateway;"
echo
print_section_title "Continue to Create topology .."
read -n 1 -s

vpcId="123e4567-e89b-12d3-a456-426614174001"
./sdnctl create vpc $vpcId vpc1 1 1

output=$(./sdnctl create subnet subnet1 10.0.0.0/24 $vpcId)
echo $output
subnet_id1=$(extract_id "$output" "Subnet Created is")

output=$(./sdnctl create subnet subnet2 20.0.0.0/24 $vpcId)
echo $output
subnet_id2=$(extract_id "$output" "Subnet Created is")

output=$(./sdnctl create router r1 $vpcId)
echo $output
router_id1=$(extract_id "$output" "Router is")

output=$(./sdnctl create router interface "$router_id1" "$subnet_id1" 10.0.0.1/24 00:11:22:33:44:55)
echo $output
router_interface_id1=$(extract_id "$output" "Router interface uuid is")

output=$(./sdnctl create router interface "$router_id1" "$subnet_id2" 20.0.0.1/24 00:11:22:33:44:66)
echo $output
router_interface_id2=$(extract_id "$output" "Router interface uuid is")

echo
print_section_title "Continue to show DB.."
read -n 1 -s
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM vpc;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM subnet;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM router;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM router_interface;"

echo
print_section_title "Continue to show OVNDB.."
read -n 1 -s
print_subsection_title "Logical Switches:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_switch

print_subsection_title "Logical Routers:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router

print_subsection_title "Logical Ports:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_switch_port

print_section_title "Continue to create logical ports.."
read -n 1 -s

output=$(./sdnctl create port "$subnet_id1" 1 8 10.0.0.2 true)
echo $output
port_id1=$(extract_id "$output" "Port Created")

output=$(./sdnctl create port "$subnet_id1" 1 9 10.0.0.3 false)
echo $output
port_id2=$(extract_id "$output" "Port Created")

output=$(./sdnctl create port "$subnet_id2" 2 8 20.0.0.2 true)
echo $output
port_id3=$(extract_id "$output" "Port Created")

output=$(./sdnctl create port "$subnet_id2" 2 9 20.0.0.3 false)
echo $output
port_id4=$(extract_id "$output" "Port Created")

echo
print_section_title "Continue to show DB.."
read -n 1 -s
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM port;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM router;"

print_section_title "Continue to show OVNDB.."
read -n 1 -s
print_subsection_title "Routers:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router

print_subsection_title "Static Routes:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router_static_route

print_subsection_title "NAT rules:"
docker exec -it docker-local-ovn-1 ovn-nbctl list nat

print_section_title "Continue Update port to reset NAT"
read -n 1 -s

./sdnctl update port "$port_id1" 0 1
./sdnctl update port "$port_id3" 0 1

print_section_title "Continue to show DB.."
read -n 1 -s
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM port;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM router;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM nat_allocated_ips;"

print_section_title "Continue to show OVNDB.."
read -n 1 -s
print_subsection_title "Routers:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router

print_subsection_title "Static Routes:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router_static_route

print_subsection_title "NAT rules:"
docker exec -it docker-local-ovn-1 ovn-nbctl list nat

print_section_title "Continue to Add Security for ports"
read -n 1 -s

output=$(./sdnctl create securitygroup "group1" $vpcId port "$port_id1,$port_id2")
echo "$output"
security_group_id1=$(extract_id "$output" "Security Group Created")

output=$(./sdnctl create securityrule "rule1" $vpcId $security_group_id1 100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
echo "$output"
security_rule_id1=$(extract_id "$output" "Security Rule Created")

output=$(./sdnctl create securityrule "rule2" $vpcId $security_group_id1 100 egress deny "10.0.1.0/24,10.0.2.0/24" "192.168.3.0/24" udp 90 100)
echo "$output"
security_rule_id2=$(extract_id "$output" "Security Rule Created")


output=$(./sdnctl create securitygroup "group2" $vpcId subnet  "$subnet_id1,$subnet_id2")
echo "$output"
security_group_id2=$(extract_id "$output" "Security Group Created")

output=$(./sdnctl create securityrule "rule3" $vpcId $security_group_id2 100 ingress allow "10.0.0.0/24" "192.168.1.0/24,192.168.2.0/24" tcp 80 443 2>&1)
echo "$output"
security_rule_id3=$(extract_id "$output" "Security Rule Created")

output=$(./sdnctl create securityrule "rule4" $vpcId $security_group_id2 100 egress deny "10.0.1.0/24,10.0.2.0/24" "192.168.3.0/24" udp 90 100)
echo "$output"
security_rule_id4=$(extract_id "$output" "Security Rule Created")

echo
print_section_title "Continue to show DB"
read -n 1 -s
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM acl;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM security_rule;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM security_group;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM port_security_group;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM subnet_security_group;"

echo
print_section_title "Continue to show OVNDB.."
read -n 1 -s
echo
print_subsection_title "ACL rules:"
docker exec -it docker-local-ovn-1 ovn-nbctl list acl

print_subsection_title "Address set:"
docker exec -it docker-local-ovn-1 ovn-nbctl list address_set

print_subsection_title "Port groups:"
docker exec -it docker-local-ovn-1 ovn-nbctl list port_group

echo
print_section_title "Continue to delete components"
read -n 1 -s
echo

output=$(./sdnctl delete securityrule "$security_rule_id1")
echo $output
output=$(./sdnctl delete securityrule "$security_rule_id2")
echo $output
output=$(./sdnctl delete securityrule "$security_rule_id3")
echo $output
output=$(./sdnctl delete securityrule "$security_rule_id4")
echo $output
output=$(./sdnctl delete securitygroup "$security_group_id1")
echo $output
output=$(./sdnctl delete securitygroup "$security_group_id2")
echo $output
echo $output
output=$(./sdnctl delete port "$port_id4" )
echo $output
output=$(./sdnctl delete port "$port_id3" )
echo $output
output=$(./sdnctl delete port "$port_id2" )
echo $output
output=$(./sdnctl delete port "$port_id1" )
echo $output

output=$(./sdnctl delete router interface "$router_interface_id1" )
echo $output
output=$(./sdnctl delete router interface "$router_interface_id2" )
echo $output
output=$(./sdnctl delete subnet "$subnet_id1")
echo $output
output=$(./sdnctl delete subnet "$subnet_id2")
echo $output
output=$(./sdnctl delete router "$router_id1")
echo $output
output=$(./sdnctl delete vpc "$vpcId")

print_section_title "Continue to show DB.."
read -n 1 -s
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM port;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM subnet;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM router;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM security_rule;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM security_group;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM port_security_group;"
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * FROM subnet_security_group;"

print_section_title "Continue to show OVNDB.."
read -n 1 -s
print_subsection_title "Logical Switches:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_switch

print_subsection_title "Logical Routers:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router

print_subsection_title "Logical Ports:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_switch_port

print_subsection_title "Static Routes:"
docker exec -it docker-local-ovn-1 ovn-nbctl list logical_router_static_route

print_subsection_title "NAT rules:"
docker exec -it docker-local-ovn-1 ovn-nbctl list nat

print_subsection_title "ACL rules:"
docker exec -it docker-local-ovn-1 ovn-nbctl list acl

print_subsection_title "Address set:"
docker exec -it docker-local-ovn-1 ovn-nbctl list address_set

print_subsection_title "Port groups:"
docker exec -it docker-local-ovn-1 ovn-nbctl list port_group
