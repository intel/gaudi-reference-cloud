-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS lifecyclestate (
    lifecyclestate_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(30),
    UNIQUE(name)
);
INSERT INTO lifecyclestate (name)
VALUES('Active'),('Deprecated') ON CONFLICT DO NOTHING;


/* METADATA TABLES */

CREATE TABLE IF NOT EXISTS provider (
   provider_name VARCHAR(30) PRIMARY KEY,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS advanceconfigitem (
   advanceconfigitem_name VARCHAR(63) PRIMARY KEY,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);
INSERT INTO advanceconfigitem (advanceconfigitem_name, lifecyclestate_id)
VALUES('kubeletArgs',1),('kubeApiServerArgs',1),('controllerManagerArgs',1),('schedulerArgs',1),('kubeProxyArgs',1) ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS k8sversion (
   k8sversion_name VARCHAR(63) PRIMARY KEY,
   provider_name VARCHAR(63) NOT NULL,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (provider_name)
       REFERENCES provider (provider_name),
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS addonversion (
   addonversion_name VARCHAR(63) PRIMARY KEY,
   name VARCHAR(63) NOT NULL,
   version VARCHAR(63) NOT NULL,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS cniversion (
   cniversion_name VARCHAR(63) PRIMARY KEY,
   name VARCHAR(63) NOT NULL,
   version VARCHAR(63) NOT NULL,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS runtime (
   runtime_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO runtime (runtime_name)
VALUES('Containerd') ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS runtimeversion (
   runtimeversion_name VARCHAR(63) PRIMARY KEY,
   runtime_name VARCHAR(63) NOT NULL,
   version VARCHAR(63) NOT NULL,
   lifecyclestate_id INT NOT NULL,
   availableargs JSON,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id),
   FOREIGN KEY (runtime_name)
       REFERENCES runtime (runtime_name)
);
INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id)
VALUES('containerd-rke2','Containerd','rke2', 1) ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS addoncompatibilityk8s (
   addonversion_name VARCHAR(63) NOT NULL,
   k8sversion_name VARCHAR(63) NOT NULL,
   PRIMARY KEY (addonversion_name, k8sversion_name),
   FOREIGN KEY (addonversion_name)
       REFERENCES addonversion (addonversion_name),
   FOREIGN KEY (k8sversion_name)
       REFERENCES k8sversion (k8sversion_name)
);

CREATE TABLE IF NOT EXISTS cnicompatibilityk8s (
   cniversion_name VARCHAR(63) NOT NULL,
   k8sversion_name VARCHAR(63) NOT NULL,
   PRIMARY KEY (cniversion_name, k8sversion_name),
   FOREIGN KEY (cniversion_name)
       REFERENCES cniversion (cniversion_name),
   FOREIGN KEY (k8sversion_name)
       REFERENCES k8sversion (k8sversion_name)
);

CREATE TABLE IF NOT EXISTS runtimecompatibilityk8s (
   runtimeversion_name VARCHAR(63) NOT NULL,
   k8sversion_name VARCHAR(63) NOT NULL,
   PRIMARY KEY (runtimeversion_name, k8sversion_name),
   FOREIGN KEY (runtimeversion_name)
       REFERENCES runtimeversion (runtimeversion_name),
   FOREIGN KEY (k8sversion_name)
       REFERENCES k8sversion (k8sversion_name)
);

CREATE TABLE IF NOT EXISTS osimage (
   osimage_name VARCHAR(63) PRIMARY KEY,
   osname VARCHAR(63) NOT NULL,
   osversion VARCHAR(63) NOT NULL
);

CREATE TABLE IF NOT EXISTS runtimecompatibilityimg (
   runtimeversion_name VARCHAR(63) NOT NULL,
   osimage_name VARCHAR(63) NOT NULL,
   PRIMARY KEY (runtimeversion_name, osimage_name),
   FOREIGN KEY (runtimeversion_name)
       REFERENCES runtimeversion (runtimeversion_name),
   FOREIGN KEY (osimage_name)
       REFERENCES osimage (osimage_name)
);

CREATE TABLE IF NOT EXISTS backuptype (
   backuptype_name VARCHAR(15) PRIMARY KEY,
   mandatoryargs JSON,
   optionalargs JSON,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS nodetype (
   nodetype_name VARCHAR(15) PRIMARY KEY,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);

CREATE TABLE IF NOT EXISTS clusterrole (
   clusterrole_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   name VARCHAR(50) NOT NULL,
   permissions JSON,
   UNIQUE(name)
);

/* CLUSTER COMPONENTS */

CREATE TABLE IF NOT EXISTS clusterstate (
   clusterstate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO clusterstate (clusterstate_name)
VALUES('Active'),('New'),('Pending'),('Updating'),('Deleting'),('Deleted'),('Error') ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS cluster (
   cluster_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   unique_id VARCHAR(14) NOT NULL,
   backuptype_name VARCHAR(15),
   clusterstate_name VARCHAR(30) NOT NULL,
   provider_name VARCHAR(30) NOT NULL,
   idc_vnet_id INT NOT NULL,
   name VARCHAR(63) NOT NULL,
   description VARCHAR(250) NOT NULL,
   state_details VARCHAR(250),
   region_id INT NOT NULL,
   networkservice_cidr VARCHAR(18) NOT NULL,
   cluster_dns VARCHAR(63) NOT NULL,
   networkpod_cidr VARCHAR(18) NOT NULL,
   provider_args JSON,
   advanceconfigs JSON,
   tags JSON,
   backup_args JSON,
   labels JSON,
   annotations JSON,
   encryptionconig JSON,
   created_date TIMESTAMP NOT NULL,
   UNIQUE(unique_id),
   FOREIGN KEY (provider_name)
       REFERENCES provider (provider_name),
   FOREIGN KEY (backuptype_name)
       REFERENCES backuptype (backuptype_name),
   FOREIGN KEY (clusterstate_name)
       REFERENCES clusterstate (clusterstate_name)
);

CREATE TABLE IF NOT EXISTS clusterrev (
   clusterrev_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   cluster_id INT NOT NULL,
   currentjson JSON NOT NULL,
   desiredjson JSON NOT NULL,
   component_typegrp VARCHAR(30) NOT NULL,
   component_typename VARCHAR(30) NOT NULL,
   currentdata VARCHAR(100) NOT NULL,
   desireddata VARCHAR(100) NOT NULL,
   completionStatus VARCHAR(30),
   timestamp TIMESTAMP NOT NULL,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id)
);

CREATE TABLE IF NOT EXISTS loglevel (
   loglevel_name VARCHAR(15) PRIMARY KEY
);
INSERT INTO loglevel (loglevel_name)
VALUES('INFO'),('DEBUG'),('WARN'),('ERROR') ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS provisioningLog (
   provisioninglog_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   cluster_id INT NOT NULL,
   logentry VARCHAR(250) NOT NULL,
   loglevel_name VARCHAR(15) NOT NULL,
   logobject VARCHAR(30) NOT NULL,
   logtimestamp TIMESTAMP NOT NULL,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (loglevel_name)
       REFERENCES loglevel (loglevel_name)
);

CREATE TABLE IF NOT EXISTS clusteraddonversion (
   cluster_id INT NOT NULL,
   addonversion_name VARCHAR(63) NOT NULL,
   lastchangetimestamp TIMESTAMP NOT NULL,
   addOnArgs JSON,
   PRIMARY KEY (cluster_id, addonversion_name),
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (addonversion_name)
       REFERENCES addonversion (addonversion_name)
);

CREATE TABLE IF NOT EXISTS clustermember (
   clustermember_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   cluster_id INT NOT NULL,
   cloudaccount_id INT NOT NULL,
   kubeconfig JSON,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id)
);

CREATE TABLE IF NOT EXISTS clusterrolebinding (
   member_id INT NOT NULL,
   clusterrole_id INT NOT NULL,
   timestamp TIMESTAMP NOT NULL,
   FOREIGN KEY (member_id)
       REFERENCES clustermember (clustermember_id),
   FOREIGN KEY (clusterrole_id)
       REFERENCES clusterrole (clusterrole_id)
);

/* NODEGROUP COMPONENTS */

CREATE TABLE IF NOT EXISTS nodegroupstate (
   nodegroupstate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);
INSERT INTO nodegroupstate (nodegroupstate_name)
VALUES('Active'), ('Unavailable'), ('Creating'), ('Updating'), ('Deleting'), ('Error') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS instancetype (
   instancetype_name VARCHAR(63) PRIMARY KEY,
   memory INT NOT NULL,
   cpu INT NOT NULL
);

CREATE TABLE IF NOT EXISTS nodegroup (
   nodegroup_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   cluster_id INT NOT NULL,
   cniversion_name VARCHAR(63) NOT NULL,
   k8sversion_name VARCHAR(63) NOT NULL,
   nodegroupstate_name VARCHAR(15) NOT NULL,
   nodegrouptype_name VARCHAR(15) NOT NULL,
   runtime_name VARCHAR(15) NOT NULL,
   unique_id VARCHAR(14) NOT NULL,
   name VARCHAR(63) NOT NULL,
   description VARCHAR(250) NOT NULL,
   availabilityZone_id INT NOT NULL,
   stateDetails VARCHAR(250) NOT NULL,
   instancetype_name VARCHAR(63) NOT NULL,
   osimage_name VARCHAR(63) NOT NULL,
   nodecount INT,
   sshkey VARCHAR(250),
   runtime_args JSON,
   cni_args JSON,
   tags JSON,
   upgstrategydrainbefdel BOOLEAN NOT NULL,
   upgstrategymaxnodes INT NOT NULL,
   lifecyclestate_id INT NOT NULL,
   createddate TIMESTAMP NOT NULL,
   UNIQUE(unique_id),
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (cniversion_name)
       REFERENCES cniversion (cniversion_name),
   FOREIGN KEY (nodegroupstate_name)
       REFERENCES nodegroupstate (nodegroupstate_name),
   FOREIGN KEY (runtime_name)
       REFERENCES runtime (runtime_name)
);

/* NODE COMPONENTS */

CREATE TABLE IF NOT EXISTS k8snodestate (
   k8snodestate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);

CREATE TABLE IF NOT EXISTS k8snode (
   k8snode_name VARCHAR(63) PRIMARY KEY,
   nodetype_name VARCHAR(15) NOT NULL,
   cluster_id INT,
   k8snodestate_name VARCHAR(15) NOT NULL,
   nodegroup_id INT NOT NULL,
   idc_instance_id INT NOT NULL,
   stateDetails VARCHAR(250) NOT NULL,
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (nodegroup_id)
       REFERENCES nodegroup (nodegroup_id),
   FOREIGN KEY (nodetype_name)
       REFERENCES nodetype (nodetype_name),
   FOREIGN KEY (k8snodestate_name)
       REFERENCES k8snodestate (k8snodestate_name)
);
