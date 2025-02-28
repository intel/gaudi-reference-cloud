--------------------------------------------------------------------------------
-- storage quota by account
--------------------------------------------------------------------------------

create table storage_quota_by_account (
    id bigserial primary key,

    cloud_account_id varchar(12) not null,

    cloud_account_type varchar(64) not null,

    updated_timestamp timestamp not null default now(),

    -- infinity means not deleted; set to 'now' when logically deleted
    deleted_timestamp timestamp not null default 'infinity',
    
    reason  text not null,

    -- filesize quota in TB
    filesize_quota_in_TB bigint not null,

    -- number of file volumes quota
    filevolumes_quota bigint not null,

    -- number of buckets quota
    buckets_quota bigint not null,
    unique (cloud_account_id, cloud_account_type, deleted_timestamp)
);

create index  if not exists storage_quota_by_account_idx on storage_quota_by_account (cloud_account_id);