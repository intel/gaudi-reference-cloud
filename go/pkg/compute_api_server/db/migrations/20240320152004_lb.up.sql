--------------------------------------------------------------------------------
-- load balancer
--------------------------------------------------------------------------------

create sequence if not exists loadbalancer_resource_version_seq minvalue 1;

create table if not exists loadbalancer (
                              resource_id uuid primary key,
                              cloud_account_id varchar(12) not null,
    -- will have same value as resourceId if not specified by user
                              name varchar(63) not null,
    -- infinity means not deleted; set to 'now' when logically deleted
                              deleted_timestamp timestamp not null default ('infinity'),
    -- provides the ordering of inserts and updates for reliable watching
                              resource_version bigint not null default nextval('loadbalancer_resource_version_seq'),
    -- Protobuf LoadBalancerPrivate message serialized as JSON.
                              value jsonb not null
);