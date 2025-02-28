--- add the TimescaleDB extension: 
CREATE EXTENSION IF NOT EXISTS timescaledb;

--- a table for product_registration
CREATE TABLE IF NOT EXISTS product_registration (
    id SERIAL PRIMARY KEY,
    product_id text NOT NULL,
    product_name text NOT NULL,
    vendor_id text NOT NULL,
    family_id text NOT NULL,
    description text NULL,
    UNIQUE (product_id, vendor_id)
);

--- a table for metering_records
CREATE TABLE IF NOT EXISTS metering_records (
    time TIMESTAMPTZ NOT NULL,
    product_id text NOT NULL,
    cloud_account_id text NOT NULL,
    comment text NULL,
    usage_unit text NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    attr jsonb NULL,
    UNIQUE (cloud_account_id, project_id, time)
);

--- a table for usage_report
CREATE TABLE IF NOT EXISTS usage_report (
    id BIGSERIAL PRIMARY KEY,
    transaction_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    cloud_account_id BIGSERIAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reported BOOLEAN DEFAULT FALSE,
    properties jsonb,
    UNIQUE (transaction_id, resource_id, cloud_account_id)
);


CREATE INDEX  IF NOT EXISTS usage_report_resource_account_idx ON usage_report (resource_id, cloud_account_id);

--- create hypertable for metering records
SELECT create_hypertable('metering_records','time');

CREATE INDEX idx_metering_records ON metering_records (product_id, time DESC);

CREATE INDEX idx_product_subscriptions ON product_subscription (product_id, vendor_id);