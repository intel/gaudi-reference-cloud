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
    num_redeemed INT
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
-- cloud_credits_intel
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS cloud_credits_intel (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    coupon_code VARCHAR(14) NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    created TIMESTAMP NOT NULL,
    expiration TIMESTAMP NOT NULL,
    original_amount NUMERIC NOT NULL,
    remaining_amount NUMERIC NOT NULL
);
CREATE INDEX IF NOT EXISTS cloud_credits_intel_account_id_idx ON cloud_credits_intel(cloud_account_id);
