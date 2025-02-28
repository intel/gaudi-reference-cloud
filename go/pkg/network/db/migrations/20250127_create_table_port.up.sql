-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- Port --
CREATE SEQUENCE port_resource_version_seq minvalue 1;

CREATE TABLE IF NOT EXISTS port (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resource_id if not specified by user
    name varchar(63) not null,
    created_timestamp timestamp DEFAULT NOW(),
    updated_timestamp timestamp DEFAULT NOW(),
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('port_resource_version_seq'),
    -- Protobuf VPC message serialized as JSON.
    value jsonb not null
);

-- Unique index prevents a record with the same name in the same cloud_account_id.
-- CREATE UNIQUE INDEX resource_idx_port on port (cloud_account_id, name);
