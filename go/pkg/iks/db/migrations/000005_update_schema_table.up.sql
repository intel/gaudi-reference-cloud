-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE runtimecompatibilityk8s ALTER COLUMN cp_osinstanceimage_name SET NOT NULL;
ALTER TABLE runtimecompatibilityk8s ALTER COLUMN wrk_osinstanceimage_name SET NOT NULL;

ALTER TABLE k8sversion ADD COLUMN cp_bootstrap_repo VARCHAR(250);
UPDATE k8sversion SET cp_bootstrap_repo = 'https://github.com';
ALTER TABLE k8sversion ALTER COLUMN cp_bootstrap_repo SET NOT NULL;
ALTER TABLE k8sversion ADD COLUMN wrk_bootstrap_repo VARCHAR(250);
UPDATE k8sversion SET wrk_bootstrap_repo = 'https://github.com';
ALTER TABLE k8sversion ALTER COLUMN wrk_bootstrap_repo SET NOT NULL;

ALTER TABLE k8snode ADD COLUMN ip_address INET NOT NULL;

ALTER TABLE clusterrev ADD COLUMN created_by VARCHAR(30) NOT NULL DEFAULT 'ADMIN_DEFAULT';

CREATE TABLE IF NOT EXISTS clusteraddonstate (
   clusteraddonstate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO clusteraddonstate (clusteraddonstate_name)
VALUES('Active'), ('Unavailable'), ('Creating'), ('Updating'), ('Deleting'), ('Error') ON CONFLICT DO NOTHING;
ALTER TABLE clusteraddonversion ADD COLUMN kubernetes_status JSON;
ALTER TABLE clusteraddonversion ADD COLUMN clusteraddonstate_name VARCHAR(15);
UPDATE clusteraddonversion SET clusteraddonstate_name = 'Active';
ALTER TABLE clusteraddonversion ALTER COLUMN clusteraddonstate_name SET NOT NULL;
ALTER TABLE clusteraddonversion ADD CONSTRAINT clusteraddonversion_clusteraddonstate_name_fkey FOREIGN KEY (clusteraddonstate_name) REFERENCES clusteraddonstate (clusteraddonstate_name);

CREATE TABLE IF NOT EXISTS loadbalancerstate (
   loadbalancerstate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO loadbalancerstate (loadbalancerstate_name)
VALUES('Active'), ('Unavailable'), ('Creating'), ('Updating'), ('Deleting'), ('Error') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS loadbalancer (
   dns VARCHAR(63) PRIMARY KEY,
   cluster_id INT NOT NULL,
   loadbalancerstate_name VARCHAR(15) NOT NULL,
   nodegrouptype_name VARCHAR(15) NOT NULL,
   kubernetes_status JSON,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (loadbalancerstate_name)
       REFERENCES loadbalancerstate (loadbalancerstate_name)
);

ALTER TABLE cluster ADD COLUMN snapshot_schedule_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE cluster ADD COLUMN snapshot_schedule_cron VARCHAR(15) NOT NULL DEFAULT '0 0 * * *';
CREATE TABLE IF NOT EXISTS snapshotstate (
   snapshotstate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO snapshotstate (snapshotstate_name)
VALUES('Done'),('Running'), ('Deleting'),('Error') ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS snapshot (
   snapshot_name VARCHAR(63) PRIMARY KEY,
   cluster_id INT NOT NULL,
   snapshotstate_name VARCHAR(15) NOT NULL,
   kubernetes_status JSON,
   snapshottype VARCHAR(15) NOT NULL,
   snapshotfile VARCHAR(63),
   created_date TIMESTAMP NOT NULL,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (snapshotstate_name)
       REFERENCES snapshotstate (snapshotstate_name)
);
