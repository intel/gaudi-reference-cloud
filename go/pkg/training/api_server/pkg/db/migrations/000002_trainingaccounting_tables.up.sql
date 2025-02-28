-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE user_accounting
ADD COLUMN latest_ssh_access timestamptz,
ADD COLUMN latest_jupyter_access timestamptz,
ADD COLUMN cloud_account_type TEXT,
ADD COLUMN enterprise_id TEXT,
ADD COLUMN linux_user_id TEXT,
ADD COLUMN user_email TEXT;

CREATE TABLE IF NOT EXISTS training_metrics (
    cloud_account_id TEXT NOT NULL,
    training_id TEXT NOT NULL,
    training_name TEXT NOT NULL,
    access_time timestamptz NOT NULL,
    PRIMARY KEY (cloud_account_id, training_id)
);
