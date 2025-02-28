-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS product_usages_report (
    id TEXT NOT NULL PRIMARY KEY,
    product_usage_id TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    product_id TEXT DEFAULT '',
    product_name TEXT DEFAULT '',
    region VARCHAR(128) NOT NULL,
    quantity NUMERIC NOT NULL,
    rate NUMERIC NOT NULL,
    unreported_quantity NUMERIC NOT NULL,
    usage_unit_type VARCHAR(64),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reported BOOLEAN DEFAULT FALSE,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
