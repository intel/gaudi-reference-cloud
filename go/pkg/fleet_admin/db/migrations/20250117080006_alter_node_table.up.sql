-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

DELETE FROM pool WHERE  pool_id = 'general';

INSERT INTO pool (pool_id, pool_name, pool_account_manager_ags_role, is_maintenance_pool)
VALUES('general','General Purpose', '', false);
