-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

create table pool_cloud_account (
    pool_id varchar(42) not null,
    cloud_account_id varchar(12) not null,
    primary key (pool_id, cloud_account_id)
);
