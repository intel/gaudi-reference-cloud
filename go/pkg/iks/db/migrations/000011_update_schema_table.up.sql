-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
INSERT INTO nodeprovider (nodeprovider_name, lifecyclestate_id)
VALUES ('Compute', 1);
 
CREATE TABLE IF NOT EXISTS defaultconfig (
   name VARCHAR(35) PRIMARY KEY,
   value VARCHAR(63)
);
 
INSERT INTO defaultconfig(name, value)
VALUES('region','us-region-1'),('vnet','us-region-1a'),('availabilityzone','us-region-1a-default'),('networkservice_cidr','100.66.0.0/16'),('networkpod_cidr','100.68.0.0/16'),('cluster_cidr','100.66.0.10');
INSERT INTO defaultconfig(name, value)
VALUES('networkinterfacename','eth0'),('ilb_environment','8'),('ilb_persist',''),('ilb_ipprotocol','tcp'),('ilb_usergroup','1545');
INSERT INTO defaultconfig(name, value)
VALUES('ilb_apiservername','apiserver'),('ilb_apiserverport','443'),('ilb_apiserverpool_port','443'),('ilb_apiserveriptype','private'),('ilb_etcdservername','etcd'),('ilb_etcdpool_port','2379'),('ilb_etcdport','443'),('ilb_etcdiptype','private'),('ilb_konnectivityname','konnectivity'),('ilb_konnectivityport','443'),('ilb_konnectivitypool_port','8132'),('ilb_konnectivityiptype','private'),('ilb_loadbalancingmode','least-connections-member'),('ilb_minactivemembers','1'),('ilb_monitor','i_tcp'),('ilb_memberConnectionLimit','0'),('ilb_memberPriorityGroup','0'),('ilb_memberratio','1'),('ilb_memberadminstate','enabled');
 
ALTER TABLE k8snode ADD COLUMN osimageinstance_name VARCHAR(63);
ALTER TABLE k8snode ADD CONSTRAINT k8snode_osimageinstance_name_fkey FOREIGN KEY (osimageinstance_name) REFERENCES osimageinstance (osimageinstance_name);
 
UPDATE nodeprovider SET nodeprovider_name = 'Harvester' WHERE nodeprovider_name = 'caas-harvester';
 
INSERT INTO defaultconfig(name, value)
VALUES('cp_cloudaccountid',''),('ilb_customer_environment','8'),('ilb_customer_usergroup','1545'),('cp_defaultsshkey','iks-cp-key');
 
INSERT INTO defaultconfig(name, value)
VALUES('ilb_public_apiservername', 'public-apiserver'),('ilb_public_apiserverpool_port','443'),('ilb_public_apiserverport','443'),('ilb_public_apiserveriptype','public'),('ilb_cp_owner','system');
ALTER TABLE k8scompatibility ADD COLUMN instancetype_wk_override JSON; /*If populated, should have type: imiwrk)*/

ALTER TABLE instancetype ADD COLUMN lifecyclestate_id INT DEFAULT 2;
UPDATE instancetype SET lifecyclestate_id = (Select lifecyclestate_id from lifecyclestate  where name = 'Staged');
ALTER TABLE instancetype ALTER COLUMN lifecyclestate_id SET NOT NULL;
ALTER TABLE instancetype ADD COLUMN displayname VARCHAR(63);
ALTER TABLE instancetype ADD COLUMN description VARCHAR(63);
ALTER TABLE instancetype ADD COLUMN instancecategory VARCHAR(63);
ALTER TABLE instancetype ADD COLUMN imi_override BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE instancetype ADD CONSTRAINT instancetype_lifecyclestate_name_fkey FOREIGN KEY (lifecyclestate_id) REFERENCES lifecyclestate (lifecyclestate_id);

INSERT INTO defaultconfig(name, value)
VALUES ('max_cust_cluster_ilb','2'),('max_cluster_ng','5'),('max_cluster_vm','50'),('max_nodegroup_vm','10');

INSERT INTO defaultconfig(name, value)
VALUES ('kubecfg_cert_expiration_days','30');
 
ALTER TABLE cloudaccountextraspec ADD COLUMN maxclusterilb_override INT;
ALTER TABLE cloudaccountextraspec ADD COLUMN maxclustervm_override INT;
ALTER TABLE cloudaccountextraspec ADD COLUMN maxclusterng_override INT;
ALTER TABLE cloudaccountextraspec ADD COLUMN maxnodegroupvm_override INT;

ALTER TABLE cluster ALTER COLUMN unique_id TYPE VARCHAR(15);
ALTER TABLE nodegroup ALTER COLUMN unique_id TYPE VARCHAR(15);

ALTER TABLE vip ADD COLUMN created_date TIMESTAMP NOT NULL DEFAULT Now();

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE cluster_extraconfig (
   cluster_id INT PRIMARY KEY,
   cluster_cacrt TEXT,
   cluster_cakey TEXT,
   cluster_sa_key TEXT,
   cluster_sa_pub TEXT,
   cluster_etcd_cacrt TEXT,
   cluster_etcd_cakey TEXT,
   cluster_etcd_rotation_keys TEXT,
   cluster_cp_reg_cmd TEXT,
   cluster_wk_reg_cmd TEXT,
   CONSTRAINT check_nospace_cacrt CHECK (cluster_cacrt LIKE '\\xc%'),
   CONSTRAINT check_nospace_cakey CHECK (cluster_cakey LIKE '\\xc%'),
   CONSTRAINT check_nospace_etcdcacrt CHECK (cluster_etcd_cacrt LIKE '\\xc%'),
   CONSTRAINT check_nospace_etcdcakey CHECK (cluster_etcd_cakey LIKE '\\xc%'),
   CONSTRAINT check_nospace_etcdrotationkeys CHECK (cluster_etcd_rotation_keys LIKE '\\xc%'),
   CONSTRAINT check_nospace_sapub CHECK (cluster_sa_pub LIKE '\\xc%'),
   CONSTRAINT check_nospace_sakey CHECK (cluster_sa_key LIKE '\\xc%'),
   CONSTRAINT check_nospace_cp_reg CHECK (cluster_cp_reg_cmd LIKE '\\xc%'),
   CONSTRAINT check_nospace_wk_reg CHECK (cluster_wk_reg_cmd LIKE '\\xc%'),
   /*CONSTRAINT check_nospace_wk_reg CHECK (cluster_wk_reg_cmd NOT LIKE '% %'),*/
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id)
);