-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- Address Translation --
CREATE SEQUENCE address_translation_resource_version_seq minvalue 1;

CREATE TABLE address_translation (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    created_timestamp timestamp DEFAULT NOW(),
    updated_timestamp timestamp DEFAULT NOW(),
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('address_translation_resource_version_seq'),
    -- Protobuf VPC message serialized as JSON.
    value jsonb not null
);
