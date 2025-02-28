-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
INSERT INTO clusterrole (name, permissions)
VALUES ('admin','{"access": "admin"}');

CREATE TABLE IF NOT EXISTS nodeprovider (
   nodeprovider_name VARCHAR(20) PRIMARY KEY,
   is_default BOOLEAN NOT NULL DEFAULT false,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);
INSERT INTO nodeprovider(nodeprovider_name, lifecyclestate_id)
VALUES ('caas-harvester', 1);

ALTER TABLE instancetype RENAME COLUMN node_provider TO nodeprovider_name;
UPDATE instancetype SET nodeprovider_name = 'caas-harvester';
ALTER TABLE instancetype ADD CONSTRAINT instancetype_nodeprovider_name_fkey FOREIGN KEY (nodeprovider_name) REFERENCES nodeprovider (nodeprovider_name);

CREATE TABLE IF NOT EXISTS cloudaccountextraspec (
   cloudaccount_id VARCHAR(60) PRIMARY KEY,
   provider_name VARCHAR(20),
   sshkeys JSON,
   FOREIGN KEY (provider_name)
       REFERENCES provider (provider_name)
);

ALTER TABLE loadbalancerstate RENAME COLUMN loadbalancerstate_name TO vipstate_name;
ALTER TABLE loadbalancerstate RENAME TO vipstate;

DROP TABLE IF EXISTS loadbalancer;

/*Nodegroup type is control plane or worker and vip type is private or public*/
CREATE TABLE IF NOT EXISTS vip (
   vip_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   vip_name VARCHAR(63),
   cluster_id INT NOT NULL,
   vipstate_name VARCHAR(15) NOT NULL,
   kubernetes_status JSON,
   backend_ports INT NOT NULL,
   frontend_ports INT NOT NULL,
   dns_aliases JSON,
   highwire_id varchar(15),
   viptype_name VARCHAR(15) NOT NULL,
   nodegroup_type VARCHAR(15) NOT NULL,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (vipstate_name)
       REFERENCES vipstate (vipstate_name)
);

ALTER TABLE k8sversion RENAME COLUMN major_version TO minor_version;
ALTER TABLE k8sversion ADD COLUMN major_version VARCHAR(2) NOT NULL DEFAULT 'v1';
ALTER TABLE k8sversion ADD COLUMN test_version BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE osimageinstancecomponent ADD COLUMN artifact_repo VARCHAR(250);

ALTER TABLE osimageinstance ADD COLUMN bootstrap_repo VARCHAR(250);

DROP TABLE IF EXISTS clusterrolebinding; 
DROP TABLE IF EXISTS clustermember;

CREATE TABLE IF NOT EXISTS member (
   member_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   username VARCHAR(100) NOT NULL,
   UNIQUE(username)
);

CREATE TABLE IF NOT EXISTS clustermember (
   member_id INT NOT NULL,
   cluster_id INT NOT NULL,
   kubeconfig JSON,
   PRIMARY KEY (cluster_id, member_id),
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (member_id)
       REFERENCES member (member_id)
);

CREATE TABLE IF NOT EXISTS clusterrolebinding (
   member_id INT NOT NULL,
   clusterrole_id INT NOT NULL,
   timestamp TIMESTAMP NOT NULL,
   cluster_id INT NOT NULL,
   PRIMARY KEY (cluster_id, member_id, clusterrole_id),
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (member_id)
       REFERENCES member (member_id),
   FOREIGN KEY (clusterrole_id)
       REFERENCES clusterrole (clusterrole_id)
);

ALTER TABLE cluster ADD COLUMN admin_kubeconfig JSON;
ALTER TABLE cluster ADD COLUMN cloudaccount_id VARCHAR(60);
UPDATE cluster SET cloudaccount_id = '1';
ALTER TABLE cluster ALTER COLUMN cloudaccount_id SET NOT NULL;

ALTER TABLE addonversion ADD COLUMN install_type VARCHAR(30);
UPDATE addonversion SET install_type = 'kubectl-apply';
ALTER TABLE addonversion ALTER COLUMN install_type SET NOT NULL;
ALTER TABLE addonversion ADD COLUMN artifact_repo VARCHAR(250);
ALTER TABLE addonversion ADD COLUMN default_addonargs JSON;
ALTER TABLE addonversion ADD COLUMN tags JSON;

ALTER TABLE nodegroup ADD COLUMN kubernetes_status JSON;
ALTER TABLE osimageinstance ADD COLUMN imiartifact VARCHAR(250);

ALTER TABLE k8snode DROP CONSTRAINT k8snode_nodetype_name_fkey;
ALTER TABLE k8snode DROP COLUMN nodetype_name;
ALTER TABLE k8snode ADD COLUMN nodeprovider_name VARCHAR(20) NOT NULL;
ALTER TABLE k8snode ADD CONSTRAINT k8snode_nodeprovider_name_fkey FOREIGN KEY (nodeprovider_name) REFERENCES nodeprovider (nodeprovider_name);
DROP TABLE nodetype;

INSERT INTO k8snodestate (k8snodestate_name)
VALUES('Active'), ('Unavailable'), ('Creating'), ('Updating'), ('Deleting'), ('Error') ON CONFLICT DO NOTHING;

ALTER TABLE k8snode ADD COLUMN kubernetes_status JSON;
ALTER TABLE osimageinstancecomponent ALTER COLUMN component_name TYPE VARCHAR(80);
ALTER TABLE k8snode Add COLUMN created_date TIMESTAMP;
UPDATE k8snode SET created_date = NOW();
ALTER TABLE k8snode ALTER COLUMN created_date SET NOT NULL;

ALTER TABLE instancetype ADD COLUMN storage INT;
UPDATE instancetype SET storage = 0;
ALTER TABLE instancetype ALTER COLUMN storage SET NOT NULL;

ALTER TABLE runtime ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE k8sversion DROP COLUMN cp_bootstrap_repo;
ALTER TABLE k8sversion DROP COLUMN wrk_bootstrap_repo;
