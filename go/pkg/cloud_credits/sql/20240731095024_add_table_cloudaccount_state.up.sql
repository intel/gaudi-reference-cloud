-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- credits_state_log
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS credits_state_log (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    state SMALLINT CHECK(state > 0),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    event_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS credits_state_log_id_idx ON credits_state_log(cloud_account_id);
