-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

CREATE INDEX idx_subnet_vpcid ON subnet ((value->'spec'->>'vpcId'));
CREATE INDEX idx_subnet_cidrblock ON subnet ((value->'spec'->>'cidrBlock'));
