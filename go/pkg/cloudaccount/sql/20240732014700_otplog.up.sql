CREATE TABLE IF NOT EXISTS otp_logs (
    log_type SMALLINT,
    cloud_account_id VARCHAR(12),
    requested_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS requested_at_idx ON otp_logs(requested_at);
