-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS user_accounting (
    cloud_account_id TEXT PRIMARY KEY NOT NULL,
    creation_date timestamptz NOT NULL,
    expiration_date timestamptz NOT NULL
);
