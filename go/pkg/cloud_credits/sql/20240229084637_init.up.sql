-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- coupons
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS coupons (
    code VARCHAR(14) PRIMARY KEY NOT NULL,
    amount NUMERIC NOT NULL,
    creator VARCHAR,
    start TIMESTAMP,
    created TIMESTAMP,
    expires TIMESTAMP,
    disabled TIMESTAMP,
    num_uses INT,
    num_redeemed INT,
    is_standard BOOLEAN
);

--------------------------------------------------------------------------------
-- redemptions
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS redemptions (
    code VARCHAR(14) NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    redeemed TIMESTAMP NOT NULL,
    installed BOOLEAN
);

ALTER TABLE redemptions ADD CONSTRAINT pk_redempetions PRIMARY KEY (code, cloud_account_id);

--------------------------------------------------------------------------------
-- cloud_credits
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS cloud_credits (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    coupon_code VARCHAR(14) NOT NULL,
    original_amount NUMERIC NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expiry TIMESTAMP,
    remaining_amount NUMERIC,
    UNIQUE (cloud_account_id, coupon_code)
);

CREATE INDEX IF NOT EXISTS cloud_credits_account_id_idx ON cloud_credits(cloud_account_id);