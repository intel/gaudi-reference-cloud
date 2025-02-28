-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--------------------------------------------------------------------------------
-- Usercredentials
--------------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS client_id_idx ON cloudaccount_user_credentials(client_id);
CREATE INDEX IF NOT EXISTS user_email_idx ON cloudaccount_user_credentials(user_email);

ALTER TABLE cloudaccount_user_credentials
ALTER COLUMN user_email TYPE varchar(255);

ALTER TABLE cloudaccount_user_credentials
ALTER COLUMN client_id TYPE varchar(128);

