-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--- a table for invalidated metering records.

CREATE TABLE IF NOT EXISTS invalidated_metering_records (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGSERIAL,
    transaction_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    region TEXT NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reported BOOLEAN DEFAULT FALSE,
    invalidity_reason SMALLINT,
    properties jsonb,
    UNIQUE (transaction_id, resource_id, cloud_account_id)
);