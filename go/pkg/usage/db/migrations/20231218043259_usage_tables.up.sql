-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS resource_usages (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    cloud_account_id VARCHAR(12) NOT NULL,
    resource_id VARCHAR(64) NOT NULL,
    resource_name TEXT,
    product_id VARCHAR(64) NOT NULL,
    product_name TEXT,
    transaction_id VARCHAR(64) NOT NULL,
    region VARCHAR(128) NOT NULL,
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiration TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    quantity NUMERIC NOT NULL,
    unreported_quantity NUMERIC NOT NULL,
    rate NUMERIC NOT NULL,
    usage_unit_type VARCHAR(64),
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reported BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS resource_metering (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    resource_id VARCHAR(64) NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    transaction_id VARCHAR(64) NOT NULL,
    region VARCHAR(128) NOT NULL,
    last_recorded TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_usages (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    cloud_account_id VARCHAR(12) NOT NULL,
    product_id VARCHAR(64) NOT NULL,
    product_name TEXT,
    region VARCHAR(128) NOT NULL,
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiration TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    quantity NUMERIC NOT NULL,
    rate NUMERIC NOT NULL,
    usage_unit_type VARCHAR(64),
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);