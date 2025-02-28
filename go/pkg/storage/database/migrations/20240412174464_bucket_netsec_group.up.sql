--------------------------------------------------------------------------------
-- vnet object
--------------------------------------------------------------------------------

create sequence vnet_resource_version_seq minvalue 1;

create type bucket_subnet_event_states as ENUM (
    'UNSPECIFIED',
    'ADDING',
    'ADDED',
    'DELETING',
    'DELETED',
    'FAILED'
);

create table bucket_network_security_group (
    id bigserial primary key,

    resource_id uuid,
    
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    vnetName varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
    updated_timestamp timestamp not null default ('infinity'),

    event_status bucket_subnet_event_states not null,
    
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('vnet_resource_version_seq'),
    -- Protobuf BucketPrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/storage/api_server/
    value jsonb not null
);

-- Unique index prevents a non-deleted bucket with the same name in the same cloud_account_id.
-- create unique index bucket_sec_group_idx on bucket_network_security_group (cloud_account_id, vnetName, updated_timestamp, event_status);
