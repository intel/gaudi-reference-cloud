CREATE TABLE IF NOT EXISTS admin_otp (
    id BIGSERIAL PRIMARY KEY,
    cloud_account_id VARCHAR(12) NOT NULL,
    member_email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(12) NOT NULL CHECK(otp_code != ''),
    otp_state SMALLINT,
    expiry TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS admin_otp_idx ON admin_otp(cloud_account_id);