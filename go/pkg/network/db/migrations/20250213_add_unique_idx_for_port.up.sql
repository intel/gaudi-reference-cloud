-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation


-- Drop the existing index if it exists
DROP INDEX IF EXISTS idx_port_ipu_chassis_mac;

-- Create the new index with the additional column
CREATE UNIQUE INDEX idx_port_ipu_chassis_mac ON port ((value->'spec'->>'ipuSerialNumber'), (value->'spec'->>'chassisId'), (value->'spec'->>'macAddress'));
