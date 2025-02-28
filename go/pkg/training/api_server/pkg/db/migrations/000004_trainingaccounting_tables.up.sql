-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

-- For user_accounting table
ALTER TABLE user_accounting
DROP CONSTRAINT user_accounting_pkey;

ALTER TABLE user_accounting
ADD COLUMN country_code TEXT;

ALTER TABLE user_accounting
ADD PRIMARY KEY (enterprise_id);

-- For training_metrics table
ALTER TABLE training_metrics
DROP CONSTRAINT training_metrics_pkey;

ALTER TABLE training_metrics
ADD COLUMN enterprise_id TEXT;
