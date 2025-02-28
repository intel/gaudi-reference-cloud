-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS product_usage_records (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    transaction_id TEXT NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    product_name TEXT DEFAULT '',
    region VARCHAR(128) NOT NULL,
    quantity NUMERIC NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reported BOOLEAN DEFAULT FALSE,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    properties jsonb,
    UNIQUE (transaction_id, cloud_account_id)
);

CREATE TABLE IF NOT EXISTS invalid_product_usage_records (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    record_id TEXT DEFAULT '',
    transaction_id TEXT DEFAULT '',
    region VARCHAR(128) DEFAULT '',
    product_name TEXT DEFAULT '',
    quantity NUMERIC NOT NULL,
    cloud_account_id VARCHAR(12) DEFAULT '',
    timestamp TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reported BOOLEAN DEFAULT FALSE,
    invalidity_reason SMALLINT,
    properties jsonb
);