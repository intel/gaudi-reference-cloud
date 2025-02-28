-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS osinstanceimage (
   osinstanceimage_name VARCHAR(63) PRIMARY KEY,
   k8sversion_name VARCHAR(63),
   osimage_name VARCHAR(63) NOT NULL,
   runtimeversion_name VARCHAR(63) NOT NULL,
   created_date TIMESTAMP NOT NULL,
   nodegrouptype_name VARCHAR(15) NOT NULL,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (osimage_name)
       REFERENCES osimage (osimage_name),
   FOREIGN KEY (runtimeversion_name)
       REFERENCES runtimeversion (runtimeversion_name),
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

ALTER TABLE k8sversion ADD COLUMN major_version VARCHAR(20);
UPDATE k8sversion SET major_version=SUBSTRING(k8sversion_name,1,5);
ALTER TABLE k8sversion ALTER COLUMN major_version SET NOT NULL;

ALTER TABLE provider ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE runtime ALTER COLUMN runtime_name TYPE VARCHAR(30);
ALTER TABLE runtimeversion ALTER COLUMN runtime_name TYPE VARCHAR(30);

ALTER TABLE runtimecompatibilityk8s DROP CONSTRAINT runtimecompatibilityk8s_runtimeversion_name_fkey;
ALTER TABLE runtimecompatibilityk8s RENAME COLUMN runtimeversion_name TO runtime_name;
UPDATE runtimecompatibilityk8s SET runtime_name='Containerd' WHERE runtime_name like '%containerd%';
ALTER TABLE runtimecompatibilityk8s ADD CONSTRAINT runtimecompatibilityk8s_runtime_name_fkey FOREIGN KEY (runtime_name) REFERENCES runtime (runtime_name);
ALTER TABLE runtimecompatibilityk8s ADD COLUMN cp_osinstanceimage_name VARCHAR(63);
ALTER TABLE runtimecompatibilityk8s ADD COLUMN wrk_osinstanceimage_name VARCHAR(63);
ALTER TABLE runtimecompatibilityk8s ADD CONSTRAINT runtimecompatibilityk8s_cp_osinstanceimage_name_fkey FOREIGN KEY (cp_osinstanceimage_name) REFERENCES osinstanceimage (osinstanceimage_name);
ALTER TABLE runtimecompatibilityk8s ADD CONSTRAINT runtimecompatibilityk8s_wrk_osinstanceimage_name_fkey FOREIGN KEY (wrk_osinstanceimage_name) REFERENCES osinstanceimage (osinstanceimage_name);

ALTER TABLE cluster DROP COLUMN state_details;
ALTER TABLE cluster ADD COLUMN kubernetes_status JSON;

ALTER TABLE clusterrev DROP COLUMN completionStatus;
ALTER TABLE clusterrev ADD COLUMN change_applied BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE osimage ADD COLUMN lifecyclestate_id INT NOT NULL DEFAULT 1;
ALTER TABLE osimage ADD CONSTRAINT osimage_lifecyclestate_id_fkey FOREIGN KEY (lifecyclestate_id) REFERENCES lifecyclestate (lifecyclestate_id);

ALTER TABLE clusterrev ADD COLUMN nodegroup_id INT;
ALTER TABLE clusterrev RENAME COLUMN currentjson TO currentspec_json;
ALTER TABLE clusterrev RENAME COLUMN desiredjson TO desiredspec_json;
ALTER TABLE clusterrev ADD CONSTRAINT clusterrev_nodegroup_id_fkey FOREIGN KEY (nodegroup_id) REFERENCES nodegroup (nodegroup_id);

INSERT INTO lifecyclestate (name) VALUES('Staged');
UPDATE lifecyclestate SET name='Archived' WHERE name='Deprecated';

DROP TABLE IF EXISTS runtimecompatibilityimg;
