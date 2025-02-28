-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--- a table for usage_report
CREATE TABLE IF NOT EXISTS invalid_metering_records (
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


CREATE INDEX  IF NOT EXISTS invalid_metering_records_idx ON invalid_metering_records (resource_id, cloud_account_id);
