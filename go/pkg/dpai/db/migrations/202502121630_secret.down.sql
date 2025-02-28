-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- Add the new columns
ALTER TABLE secret
DROP COLUMN  nonce;

-- Remove the old column
ALTER TABLE secret
DROP COLUMN encrypted_password;

