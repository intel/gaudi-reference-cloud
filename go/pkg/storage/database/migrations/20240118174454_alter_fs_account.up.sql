-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE filesystem_namespace_user_account DROP CONSTRAINT filesystem_namespace_user_acc_cloud_account_id_cluster_name_key;

ALTER TABLE filesystem_namespace_user_account DROP COLUMN namespace_password;
ALTER TABLE filesystem_namespace_user_account DROP COLUMN namespace_user;
ALTER TABLE filesystem_namespace_user_account DROP COLUMN filesystem_user;
ALTER TABLE filesystem_namespace_user_account DROP COLUMN filesystem_password;

ALTER TABLE filesystem_namespace_user_account ADD  CONSTRAINT fs_acc_uniq UNIQUE (cloud_account_id, cluster_name);
