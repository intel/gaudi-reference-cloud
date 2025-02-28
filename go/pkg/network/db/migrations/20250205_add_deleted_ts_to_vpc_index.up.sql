-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- Drop the existing index
DROP INDEX IF EXISTS resource_idx_vpc;

-- Create the new index with the additional column
CREATE UNIQUE INDEX resource_idx_vpc on vpc (cloud_account_id, name, deleted_timestamp);