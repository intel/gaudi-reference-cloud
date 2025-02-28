-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

ALTER TABLE pool_cloud_account ADD create_date timestamp not null default current_timestamp;
ALTER TABLE pool_cloud_account ADD create_admin varchar(253) not null default '';

create table pool (
    pool_id varchar(42) primary key,
    pool_name varchar(253) not null,
    pool_account_manager_ags_role varchar(253) not null,
    is_maintenance_pool boolean not null default false
);

create table node (
    node_id int primary key,
    region varchar(100) not null,
    availability_zone varchar(100) not null,
    cluster_id varchar(100) not null,
    namespace varchar(63) not null,
    node_name varchar(253) not null,
    override_instance_types boolean not null,
    override_pools boolean not null
);

create table node_pool (
    node_id int not null,
    pool_id varchar(42) not null,
    primary key (node_id, pool_id)
);

create table node_pool_override (
    node_id int not null,
    pool_id varchar(42) not null,
    primary key (node_id, pool_id)
);

create table node_instance_type (
    node_id int not null,
    instance_type varchar(253) not null,
    primary key (node_id, instance_type)
);

create table node_instance_type_override (
    node_id int not null,
    instance_type varchar(253) not null,
    primary key (node_id, instance_type)
);

create table node_stats (
    node_id int primary key,
    reported_time timestamp not null,
    source_group varchar(253) not null,
    source_version varchar(63) not null,
    source_resource varchar(63) not null,
    instance_category varchar(32) not null,
    partition varchar(100) not null,
    cluster_group varchar(100) not null,
    network_mode varchar(100) not null,
    free_millicpu int not null,
    used_millicpu int not null,
    free_memory_bytes bigint not null,
    used_memory_bytes bigint not null,
    free_gpu int not null,
    used_gpu int not null
);

create table node_instance_type_stats (
    node_id int not null,
    instance_type varchar(253) not null,
    running_instances int not null,
    max_new_instances int not null,
    primary key (node_id, instance_type)
);
