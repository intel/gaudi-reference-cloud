-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE osimage ADD COLUMN cp_default BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE osimage ADD COLUMN wrk_default BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE cluster ADD COLUMN snapshot_retention INT NOT NULL DEFAULT 10;
ALTER TABLE addonversion ADD COLUMN admin_only BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE snapshot ADD COLUMN schedule_type VARCHAR(15);

UPDATE nodegroup SET osimage_name = 'ubuntu2204-k1.22.17-c1.0.0' WHERE osimage_name = 'ubuntu2204';
ALTER TABLE nodegroup RENAME COLUMN osimage_name TO osimageinstance_name;
ALTER TABLE nodegroup ADD CONSTRAINT nodegroup_osimageinstance_name_fkey FOREIGN KEY (osimageinstance_name) REFERENCES osimageinstance (osimageinstance_name);
ALTER TABLE nodegroup ALTER COLUMN sshkey TYPE JSONB USING to_jsonb(sshkey);

ALTER TABLE instancetype ADD COLUMN node_provider VARCHAR(30);
UPDATE instancetype SET node_provider = 'caas-harvester';
ALTER TABLE instancetype ALTER COLUMN node_provider SET NOT NULL;

ALTER TABLE osimageinstance ADD COLUMN provider_name VARCHAR(30);
ALTER TABLE osimageinstance ADD CONSTRAINT osimageinstance_provider_name_fkey FOREIGN KEY (provider_name) REFERENCES provider (provider_name);
UPDATE osimageinstance SET provider_name = 'rke2';
ALTER TABLE osimageinstance ALTER COLUMN provider_name SET NOT NULL;

ALTER TABLE k8scompatibility ADD COLUMN provider_name VARCHAR(30);
UPDATE k8scompatibility SET provider_name = 'rke2';
ALTER TABLE k8scompatibility ALTER COLUMN provider_name SET NOT NULL;
ALTER TABLE k8scompatibility DROP CONSTRAINT k8scompatibility_pkey;
ALTER TABLE k8scompatibility ADD PRIMARY KEY (runtime_name, k8sversion_name, osimage_name, provider_name);
ALTER TABLE k8scompatibility ADD CONSTRAINT k8scompatibility_provider_name_fkey FOREIGN KEY (provider_name) REFERENCES provider (provider_name);

CREATE TABLE IF NOT EXISTS osimageinstancecomponent (
   component_name VARCHAR(30) NOT NULL,
   osimageinstance_name VARCHAR(63) NOT NULL,
   version VARCHAR(30) NOT NULL,
   PRIMARY KEY (component_name, osimageinstance_name),
   FOREIGN KEY (osimageinstance_name)
       REFERENCES osimageinstance (osimageinstance_name)
);

ALTER TABLE k8sversion DROP COLUMN provider_name;
ALTER TABLE nodegroup DROP COLUMN cniversion_name;
ALTER TABLE nodegroup DROP COLUMN cni_args;
ALTER TABLE snapshot DROP COLUMN snapshottype;
DROP TABLE IF EXISTS cnicompatibilityk8s;
DROP TABLE IF EXISTS cniversion;
