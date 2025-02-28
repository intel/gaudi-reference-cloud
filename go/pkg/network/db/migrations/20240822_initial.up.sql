-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- VPC --
CREATE SEQUENCE vpc_resource_version_seq minvalue 1;

CREATE TABLE vpc (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resource_id if not specified by user
    name varchar(63) not null,
    created_timestamp timestamp DEFAULT NOW(),
    updated_timestamp timestamp DEFAULT NOW(),
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('vpc_resource_version_seq'), 
    -- Protobuf VPC message serialized as JSON.
    value jsonb not null
);

-- Unique index prevents a record with the same name in the same cloud_account_id.
CREATE UNIQUE INDEX resource_idx_vpc on vpc (cloud_account_id, name);

-- SUBNET --
CREATE SEQUENCE subnet_resource_version_seq minvalue 1;

CREATE TABLE subnet (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resource_id if not specified by user
    name varchar(63) not null,
    created_timestamp timestamp DEFAULT NOW(),
    updated_timestamp timestamp DEFAULT NOW(),
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('subnet_resource_version_seq'), 
    -- Protobuf VPCSubnet message serialized as JSON.
    value jsonb not null
);

-- Unique index prevents a record with the same name in the same cloud_account_id.
CREATE UNIQUE INDEX resource_idx_subnet on subnet (cloud_account_id, name);
