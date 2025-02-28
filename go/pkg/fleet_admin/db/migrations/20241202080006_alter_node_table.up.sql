-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

INSERT INTO pool (pool_id, pool_name, pool_account_manager_ags_role, is_maintenance_pool)
VALUES('general','General Purpose', 'IDC Pool Account Manager - General', false) ON CONFLICT DO NOTHING;
