-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
alter table workspace 
    drop column iks_cluster_name;

alter table workspace 
    drop column ssh_key_name;