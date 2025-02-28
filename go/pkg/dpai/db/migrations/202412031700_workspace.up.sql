-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
alter table workspace
    add column if not exists iks_cluster_name text, 
    add column if not exists ssh_key_name text;