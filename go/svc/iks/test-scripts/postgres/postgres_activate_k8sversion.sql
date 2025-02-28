-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
/*
K8SPATCHVERSION i.e. 1.27.5
*/

SELECT k8sversion_name, minor_version, lifecyclestate_id FROM k8sversion;

BEGIN;

UPDATE k8sversion SET lifecyclestate_id = (SELECT lifecyclestate_id FROM lifecyclestate where name = 'Archived') WHERE minor_version = 'K8SVERSION' and lifecyclestate_id = (SELECT lifecyclestate_id FROM lifecyclestate where name = 'Active');
UPDATE k8sversion SET lifecyclestate_id = (SELECT lifecyclestate_id FROM lifecyclestate where name = 'Active') WHERE k8sversion_name = 'K8SPATCHVERSION';

COMMIT;

SELECT k8sversion_name, minor_version, lifecyclestate_id FROM k8sversion;
