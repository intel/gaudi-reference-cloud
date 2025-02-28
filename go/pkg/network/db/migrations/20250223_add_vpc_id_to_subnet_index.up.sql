-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- Drop the existing index
DROP INDEX IF EXISTS resource_idx_subnet;

-- Create the new index with the additional column
CREATE UNIQUE INDEX resource_idx_subnet ON subnet (cloud_account_id, name, (value->'spec'->>'vpcId'), deleted_timestamp);
