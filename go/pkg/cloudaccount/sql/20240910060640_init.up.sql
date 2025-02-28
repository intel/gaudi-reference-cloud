-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- Usercredentials
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS cloudaccount_user_credentials (
    cloudaccount_id VARCHAR(12) NOT NULL,
    user_email VARCHAR(32) NOT NULL CHECK (user_email != ''),
    client_id VARCHAR(32) NOT NULL CHECK (client_id != '') UNIQUE,
    appclient_name VARCHAR(255),
    country_code VARCHAR(255),
    revoked BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
       PRIMARY KEY (
        user_email,
        client_id,
        cloudaccount_id
    )
);
