--------------------------------------------------------------------------------
-- object bucket
--------------------------------------------------------------------------------

create sequence bucket_resource_version_seq minvalue 1;

create table bucket (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    name varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('bucket_resource_version_seq'),
    -- Protobuf BucketPrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/storage/api_server/
    value jsonb not null
);

-- Unique index prevents a non-deleted bucket with the same name in the same cloud_account_id.
create unique index bucket_idx on bucket (cloud_account_id, name, deleted_timestamp);

-- Index optimizes the watch query.
create unique index bucket_resource_version_idx on bucket (resource_version);

create sequence object_user_resource_version_seq minvalue 1;

create table object_user (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    name varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('object_user_resource_version_seq'),
    -- Protobuf BucketPrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/storage/api_server/
    value jsonb not null
);

-- Unique index prevents a non-deleted object_user with the same name in the same cloud_account_id.
create unique index object_user_idx on object_user (cloud_account_id, name, deleted_timestamp);

-- Index optimizes the watch query.
create unique index object_user_resource_version_idx on object_user (resource_version);

create table object_user_bucket_mapping (
    id bigserial primary key,

    bucket_id uuid references bucket,

    user_id uuid references object_user
);

create sequence bucket_lifecycle_rule_seq minvalue 1;

create table bucket_lifecycle_rule (
    resource_id uuid primary key,
    cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
    name varchar(63) not null,

    bucket_id uuid references bucket,

    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
    resource_version bigint not null default nextval('bucket_lifecycle_rule_seq'),
    -- Protobuf BucketPrivate message serialized as JSON.
    -- See https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/storage/api_server/
    value jsonb not null
);