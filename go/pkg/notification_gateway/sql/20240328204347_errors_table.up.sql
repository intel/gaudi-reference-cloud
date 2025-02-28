-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

CREATE TABLE errors (
    id VARCHAR(12) NOT NULL PRIMARY KEY,
    cloud_account_id VARCHAR(12),
    user_id VARCHAR(64),
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiration TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(12) NOT NULL,
    error_type VARCHAR(64) NOT NULL,
    severity VARCHAR(12) NOT NULL,
    service_name VARCHAR(64),
    message VARCHAR(64),
    properties jsonb,
    client_record_id VARCHAR(64) NOT NULL,
    region VARCHAR(64),
    UNIQUE (id, client_record_id)
);