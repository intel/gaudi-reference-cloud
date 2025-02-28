-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- for notifications 
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(12) NOT NULL PRIMARY KEY,
    cloud_account_id VARCHAR(12),
    user_id VARCHAR(64),
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiration TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(12) NOT NULL,
    notification_type VARCHAR(64) NOT NULL,
    severity VARCHAR(12) NOT NULL,
    service_name VARCHAR(64),
    message VARCHAR(64),
    properties jsonb,
    client_record_id VARCHAR(64) NOT NULL,
    UNIQUE (id, client_record_id)
);


--------------------------------------------------------------------------------
-- for alerts
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(12) NOT NULL PRIMARY KEY,
    cloud_account_id VARCHAR(12),
    user_id VARCHAR(64),
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiration TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(12) NOT NULL,
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(12) NOT NULL,
    service_name VARCHAR(64),
    message VARCHAR(64),
    properties jsonb,
    client_record_id VARCHAR(64) NOT NULL,
    UNIQUE (id, client_record_id)
);
