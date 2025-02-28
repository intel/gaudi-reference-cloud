-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE user_accounting
DROP COLUMN latest_ssh_access
DROP COLUMN latest_jupyter_access
DROP COLUMN cloud_account_type
DROP COLUMN enterprise_id
DROP COLUMN linux_user_id
DROP COLUMN user_email;

DROP table training_metrics;
