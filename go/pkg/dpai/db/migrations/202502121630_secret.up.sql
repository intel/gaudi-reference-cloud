-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- Add the new columns
ALTER TABLE secret
ADD COLUMN encrypted_password TEXT NOT NULL,
ADD COLUMN nonce TEXT NOT NULL;

-- Remove the old column
ALTER TABLE secret
DROP COLUMN value;

-- Set the primary key on id
ALTER TABLE secret
ADD CONSTRAINT secret_pkey PRIMARY KEY (id);

-- Rename the column and change its data type
ALTER TABLE airflow
RENAME COLUMN webserver_admin_password TO webserver_admin_password_secret_id;

ALTER TABLE airflow
ALTER COLUMN webserver_admin_password_secret_id TYPE INT USING webserver_admin_password_secret_id::INTEGER;