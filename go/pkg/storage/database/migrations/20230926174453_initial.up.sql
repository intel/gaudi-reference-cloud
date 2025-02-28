-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- filesystem
--------------------------------------------------------------------------------

create sequence filesystem_resource_version_seq minvalue 1;

create table filesystem (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    name varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('filesystem_resource_version_seq'),
    -- Protobuf FilesystemPrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/storage/api_server/
    value jsonb not null
);

-- Unique index prevents a non-deleted filesystem with the same name in the same cloud_account_id.
create unique index filesystem_idx on filesystem (cloud_account_id, name, deleted_timestamp);

-- Index optimizes the watch query.
create unique index filesystem_resource_version_idx on filesystem (resource_version);

--------------------------------------------------------------------------------
-- storage -filesystem account table
--------------------------------------------------------------------------------

create table filesystem_namespace_user_account (
    id bigserial primary key,

    cloud_account_id varchar(12) not null,

    -- cluster assigned to this account
    cluster_name text not null, 

    -- cluster address assigned to this account
    cluster_addr text not null, 

    -- cluster uuid assigned to this account
    cluster_uuid text not null, 

    -- namespace on the cluster that is assigned to this account
    namespace_name text not null,

    -- namespace user for a given namespace
    namespace_user text not null,

    -- namespace password for a given namespace. namespace credentials are used
    -- only for admin purpose and not share with the user
    namespace_password text not null,

    -- filesystem user
    filesystem_user text not null,

    -- filesystem user
    filesystem_password text not null,

    -- time when this record was last updated. 
    updated_timestamp timestamptz not null default now(),

    -- time when this record was deleted. 
    deleted_timestamp timestamp not null default ('infinity'),

    unique (cloud_account_id, cluster_name, namespace_user, filesystem_user)
);

create index  if not exists filesystem_namespace_account_idx on filesystem_namespace_user_account (cloud_account_id);
