-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--- a table for usage_report
CREATE TABLE IF NOT EXISTS usage_report (
    id BIGSERIAL PRIMARY KEY,
    transaction_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reported BOOLEAN DEFAULT FALSE,
    properties jsonb,
    UNIQUE (transaction_id, resource_id, cloud_account_id)
);


CREATE INDEX  IF NOT EXISTS usage_report_resource_account_idx ON usage_report (resource_id, cloud_account_id);
