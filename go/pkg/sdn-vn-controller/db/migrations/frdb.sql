-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- INTEL CONFIDENTIAL
-- Copyright (C) 2024 Intel Corporation
/* Database definition for Network Objects supported by Controller */

/* To Create a db using psql:
    psql -f <filespec of this file> --host=localhost --port=5432 --dbname=frdb --username=<username> --password
*/

/* TODO:  remove the following once out of initial development phase */

select 'drop table if exists "' || tablename || '" cascade;' 
  from pg_tables
 where schemaname = 'public';
\gexec

/* Instantiate the tables for basic objects */

CREATE TABLE IF NOT EXISTS vpc (
   vpc_id UUID PRIMARY KEY,
   vpc_name VARCHAR(255) NOT NULL UNIQUE,
   tenant_id VARCHAR(255) NOT NULL,
   region_id VARCHAR(100) NOT NULL
);


CREATE TABLE IF NOT EXISTS subnet(
   subnet_id UUID PRIMARY KEY,
   subnet_name VARCHAR(255) NOT NULL UNIQUE,
   subnet_cidr CIDR NOT NULL,
   availability_zone VARCHAR(100) NOT NULL,
   vpc_id UUID NOT NULL,
      FOREIGN KEY(vpc_id) 
        REFERENCES vpc(vpc_id)
);


CREATE TABLE IF NOT EXISTS gateway(
   gateway_id UUID PRIMARY KEY,
   chassis_id INTEGER NOT NULL,
   gateway_name VARCHAR(255) NOT NULL UNIQUE,
   subnet_id UUID NOT NULL,
   ext_port_name VARCHAR(255) NOT NULL,
   ext_network_name VARCHAR(255) NOT NULL,
   access_network INET NOT NULL,
   maxNATs INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS router(
   router_id UUID PRIMARY KEY,
   router_name VARCHAR(255) NOT NULL UNIQUE,
   snat_ip_address INET,
   gateway_id UUID,
   gw_interface_uuid UUID,
   gw_switch_port_id UUID,
   gw_port_ip_address INET,
   gw_route_uuid UUID,
   vpc_id UUID NOT NULL,
      FOREIGN KEY(vpc_id) 
        REFERENCES vpc(vpc_id),
      FOREIGN KEY(gateway_id) 
        REFERENCES gateway(gateway_id)
);

CREATE TABLE IF NOT EXISTS port(
   port_id UUID PRIMARY KEY,
   port_name VARCHAR(255) NOT NULL UNIQUE,
   subnet_id UUID NOT NULL,
      FOREIGN KEY(subnet_id) 
        REFERENCES subnet(subnet_id),
   /* TODO: revisit the data types for chassis id and device id once their
    definintion and role becomes clear*/
   chassis_id VARCHAR(255),  
   device_id  INTEGER,
   admin_state_up  boolean,
   mac_address MACADDR,
   internal_ip_address INET NOT NULL,
   isNAT boolean NOT NULL,
   snat_rule_id UUID,
   external_ip_address INET,
   external_mac_address MACADDR,
   availability_zone VARCHAR(100) NOT NULL,
   vpc_id UUID NOT NULL,
      FOREIGN KEY(vpc_id) 
        REFERENCES vpc(vpc_id)
);

CREATE TABLE IF NOT EXISTS static_route(
   static_route_id UUID PRIMARY KEY,
   router_id UUID NOT NULL,
   prefix CIDR NOT NULL,
   nexthop INET NOT NULL,
      FOREIGN KEY(router_id) 
        REFERENCES router(router_id)
);

CREATE TABLE IF NOT EXISTS router_interface(
   router_interface_id UUID PRIMARY KEY,
   subnet_id UUID NOT NULL UNIQUE,
   router_id UUID NOT NULL,
   router_port_id UUID NOT NULL,
   switch_port_id UUID NOT NULL,
   router_interface_ip_address INET NOT NULL,
   router_mac_address MACADDR,
      FOREIGN KEY(router_id) 
        REFERENCES router(router_id),
      FOREIGN KEY(subnet_id) 
        REFERENCES subnet(subnet_id)
);

CREATE TABLE IF NOT EXISTS security_group(
   security_group_id UUID PRIMARY KEY,
   security_group_name VARCHAR(255) NOT NULL UNIQUE,
   vpc_id UUID NOT NULL,
   security_group_type INTEGER NOT NULL,
      FOREIGN KEY(vpc_id) 
        REFERENCES vpc(vpc_id)
);

CREATE TABLE IF NOT EXISTS security_rule(
   security_rule_id UUID PRIMARY KEY,
   security_rule_name VARCHAR(255) NOT NULL UNIQUE,
   security_rule_priority INTEGER NOT NULL,
   direction INTEGER NOT NULL,
   protocol INTEGER ,
   source_ip_addresses INET[], 
   source_port INT4RANGE, 
   destination_ip_addresses INET[], 
   destination_port INT4RANGE, 
   security_rule_action INTEGER NOT NULL,
   src_address_set_uuid UUID,
   dst_address_set_uuid UUID,
   vpc_id UUID NOT NULL,
   security_group_id UUID NOT NULL,
      FOREIGN KEY(vpc_id) 
        REFERENCES vpc(vpc_id),
      FOREIGN KEY(security_group_id) 
        REFERENCES security_group(security_group_id)
);

/* Create relationship tables to keep track of dependencies */

/* A security rule maps to an acl in each of the AZ the security rule
 needs to be present in */
CREATE TABLE IF NOT EXISTS acl(
   ovn_acl_id UUID PRIMARY KEY,
   security_rule_id UUID,
   availability_zone VARCHAR(100) NOT NULL,
      FOREIGN KEY(security_rule_id) 
        REFERENCES security_rule(security_rule_id)
);


/* A Security Group maps to a Port Group in each of the AZ the Security Group
 needs to be present in */
CREATE TABLE IF NOT EXISTS port_group(
   port_group_id UUID PRIMARY KEY,
   security_group_id UUID,
   availability_zone VARCHAR(100) NOT NULL,
      FOREIGN KEY(security_group_id) 
        REFERENCES security_group(security_group_id)
);

/* Subnet to Security Group association  */

CREATE TABLE IF NOT EXISTS subnet_security_group(
   subnet_id UUID,
   security_group_id UUID,
   PRIMARY KEY (subnet_id, security_group_id),
      CONSTRAINT fk_subnet
      FOREIGN KEY(subnet_id) 
        REFERENCES subnet(subnet_id),
      CONSTRAINT fk_security_group
      FOREIGN KEY(security_group_id) 
        REFERENCES security_group(security_group_id)
);

/* Port to Security Group association  */

CREATE TABLE IF NOT EXISTS port_security_group(
   port_id UUID ,
   security_group_id UUID ,
   PRIMARY KEY (port_id, security_group_id),
      CONSTRAINT fk_port
      FOREIGN KEY(port_id) 
        REFERENCES port(port_id),
      CONSTRAINT fk_security_group
      FOREIGN KEY(security_group_id) 
        REFERENCES security_group(security_group_id)
);

CREATE TABLE IF NOT EXISTS nat_available_ips (
   ip_manager_uuid UUID NOT NULL,
   ip VARCHAR(15) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS nat_allocated_ips (
   ip_manager_uuid UUID NOT NULL,
   ip VARCHAR(15) PRIMARY KEY
);