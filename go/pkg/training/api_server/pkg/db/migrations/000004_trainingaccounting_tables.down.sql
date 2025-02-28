-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE user_accounting
DROP COLUMN country_code;

ALTER TABLE training_metrics
DROP COLUMN enterprise_id
