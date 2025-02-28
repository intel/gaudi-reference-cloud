CREATE TABLE IF NOT EXISTS credit_usage (
    usage_id BIGINT NOT NULL PRIMARY KEY
);

ALTER TABLE cloud_credits_intel ADD reason SMALLINT;
