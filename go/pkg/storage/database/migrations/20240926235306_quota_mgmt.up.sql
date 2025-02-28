-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

--------------------------------------------------------------------------------
--  quota by management service
--------------------------------------------------------------------------------

create type quota_scope_type AS ENUM (
    'QUOTA_ACCOUNT_TYPE',
    'QUOTA_ACCOUNT_ID'
);

create type quota_unit_type AS ENUM (
    'COUNT',
    'REQ_SEC',
    'REQ_MIN',
    'REQ_HOUR'
);

create table registered_services (
    id bigserial primary key,

    service_id varchar(64) not null,

    service_name varchar(24) not null,

    region varchar(50) not null,

    unique (service_name)
);

create unique index  if not exists service_id_idx on registered_services (service_id, service_name);


create table service_resources (
    id bigserial primary key,

    service_id varchar(64) not null,

    resource_name varchar(64) not null,

    quota_unit quota_unit_type not null,

    max_limit bigint not null,

    unique (service_id, resource_name)
);

create unique index  if not exists service_id_resource_idx on service_resources (service_id, resource_name);

create table service_resource_quotas (
    id bigserial primary key,

    service_id varchar(64) not null,

    resource_name varchar(64) not null,

    rule_id varchar(64) not null,

    limits	bigint	not null,

    quota_unit quota_unit_type not null,

    scope_type	quota_scope_type not null,

    scope_value	varchar(48)	not null,
    
    reason  text not null,

    created_timestamp timestamp not null default now(),

    updated_timestamp timestamp not null default now(),
    
    unique (service_id, resource_name, rule_id)

);
