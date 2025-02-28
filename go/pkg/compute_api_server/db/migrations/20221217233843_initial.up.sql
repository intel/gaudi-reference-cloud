--------------------------------------------------------------------------------
-- ssh_public_key
--------------------------------------------------------------------------------

create table ssh_public_key (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resource_id if not specified by user
    name varchar(63) not null,
    ssh_public_key text not null
);

-- Unique index prevents a record with the same name in the same cloud_account_id.
create unique index ssh_public_key_idx on ssh_public_key (cloud_account_id, name);

--------------------------------------------------------------------------------
-- instance
--------------------------------------------------------------------------------

create sequence instance_resource_version_seq minvalue 1;

create table instance (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    name varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('instance_resource_version_seq'),
    -- Protobuf InstancePrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/compute_api_server/instance/instance_serializer.go
    value jsonb not null
);

-- Unique index prevents a non-deleted instance with the same name in the same cloud_account_id.
create unique index instance_idx on instance (cloud_account_id, name, deleted_timestamp);

-- Index optimizes the watch query.
create unique index instance_resource_version_idx on instance (resource_version);

--------------------------------------------------------------------------------
-- instance_type
--------------------------------------------------------------------------------

create table instance_type (
    name varchar(253) primary key,
    -- Protobuf InstanceType message serialized as JSON.
    value jsonb not null
);

create index instance_type_idx on instance_type using gin (value);

--------------------------------------------------------------------------------
-- machine_image
--------------------------------------------------------------------------------

create table machine_image (
    name varchar(253) primary key,
    -- Protobuf MachineImage message serialized as JSON.
    value jsonb not null
);

create index machine_image_idx on machine_image using gin (value);


--------------------------------------------------------------------------------
-- vnet
--------------------------------------------------------------------------------

create table vnet (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resource_id if not specified by user
    name varchar(63) not null,
    value jsonb not null
);

-- Unique index prevents a record with the same name in the same cloud_account_id.
create unique index vnet_idx on vnet (cloud_account_id, name);
