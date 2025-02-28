-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE k8scompatibility ADD COLUMN instancetype_name VARCHAR(63);
ALTER TABLE k8scompatibility ADD COLUMN lifecyclestate_id INT;

UPDATE k8scompatibility SET instancetype_name = (SELECT instancetype_name FROM instancetype order by instancetype_name LIMIT 1);

ALTER TABLE k8scompatibility DROP CONSTRAINT k8scompatibility_pkey;
ALTER TABLE k8scompatibility ADD PRIMARY KEY (runtime_name, k8sversion_name, osimage_name, provider_name, instancetype_name);
ALTER TABLE k8scompatibility ADD CONSTRAINT k8scompatibility_instancetype_name_fkey FOREIGN KEY (instancetype_name) REFERENCES instancetype (instancetype_name);

INSERT INTO k8scompatibility (instancetype_name, k8sversion_name, runtime_name, provider_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, instancetype_wk_override) 
SELECT a.instancetype_name, k.k8sversion_name, k.runtime_name, k.provider_name, k.osimage_name, k.cp_osimageinstance_name, k.wrk_osimageinstance_name, k.instancetype_wk_override from instancetype a CROSS JOIN k8scompatibility k WHERE a.instancetype_name != (SELECT instancetype_name from instancetype order by instancetype_name LIMIT 1);

UPDATE k8scompatibility SET lifecyclestate_id = (SELECT lifecyclestate_id FROM lifecyclestate WHERE name = 'Active');
ALTER TABLE k8scompatibility ALTER COLUMN lifecyclestate_id SET NOT NULL;
