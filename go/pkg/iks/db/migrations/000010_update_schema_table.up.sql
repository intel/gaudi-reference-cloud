-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE nodegroup ADD COLUMN annotations JSON;
ALTER TABLE nodegroup ADD COLUMN hw_type VARCHAR(30) NOT NULL DEFAULT 'standard';
ALTER TABLE osimageinstance ADD COLUMN hw_type VARCHAR(30) NOT NULL DEFAULT 'standard';

ALTER TABLE instancetype ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE instancetype SET is_default = TRUE where instancetype_name = 'small';

UPDATE nodegroup SET instancetype_name='small' WHERE instancetype_name != 'small' AND instancetype_name != 'medium';
ALTER TABLE nodegroup ADD CONSTRAINT nodegroup_instancetype_name_fkey FOREIGN KEY (instancetype_name) REFERENCES instancetype (instancetype_name);

ALTER TABLE k8snode DROP COLUMN statedetails;
ALTER TABLE cluster ALTER COLUMN admin_kubeconfig TYPE TEXT;
ALTER TABLE clustermember ALTER COLUMN kubeconfig TYPE TEXT;

ALTER TABLE nodegroup ADD CONSTRAINT nodegroup_k8sversion_fkey FOREIGN KEY (k8sversion_name) REFERENCES k8sversion (k8sversion_name);

ALTER TABLE k8snode DROP CONSTRAINT k8snode_pkey;
ALTER TABLE k8snode ADD PRIMARY KEY (ip_address);

CREATE TABLE IF NOT EXISTS vipprovider (
   vipprovider_name VARCHAR(20) PRIMARY KEY,
   is_default BOOLEAN NOT NULL DEFAULT false,
   lifecyclestate_id INT NOT NULL,
   FOREIGN KEY (lifecyclestate_id)
       REFERENCES lifecyclestate (lifecyclestate_id)
);
INSERT INTO vipprovider(vipprovider_name, lifecyclestate_id, is_default)
VALUES ('Highwire', 1, true);

DROP TABLE IF EXISTS vip;

/*Nodegroup type is control plane or worker and vip type is private or public*/
CREATE TABLE IF NOT EXISTS vip (
   vip_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   cluster_id INT NOT NULL,
   dns_aliases JSON,
   vip_dns VARCHAR(63),
   vip_ip INET,
   viptype_name VARCHAR(15) NOT NULL,
   vip_status JSON,
   vipstate_name VARCHAR(15) NOT NULL,
   owner VARCHAR(30) NOT NULL,
   vipprovider_name VARCHAR(20) NOT NULL,
   vipinstance_id varchar(15),
   FOREIGN KEY (cluster_id)
       REFERENCES cluster (cluster_id),
   FOREIGN KEY (vipstate_name)
       REFERENCES vipstate (vipstate_name),
   FOREIGN KEY (vipprovider_name)
       REFERENCES vipprovider (vipprovider_name)
);

CREATE TABLE IF NOT EXISTS vipdetails (
   vip_id INT NOT NULL,
   vip_name VARCHAR(63) NOT NULL,
   pool_name VARCHAR(63),
   pool_id VARCHAR(20),
   description VARCHAR(250),
   port INT,
   pool_port INT,
   UNIQUE(pool_id),
   PRIMARY KEY(vip_id, vip_name),
   FOREIGN KEY (vip_id)
       REFERENCES vip (vip_id)
);

CREATE TABLE IF NOT EXISTS vipmembers (
   ip_address INET,
   pool_id VARCHAR(20),
   poolnode_status VARCHAR(63),
   PRIMARY KEY(ip_address, pool_id),
   FOREIGN KEY (ip_address)
       REFERENCES k8snode (ip_address),
   FOREIGN KEY (pool_id)
       REFERENCES vipdetails (pool_id)
);

ALTER table cluster drop column idc_vnet_id;
ALTER TABLE nodegroup DROP COLUMN availabilityzone_id;
ALTER TABLE nodegroup ADD COLUMN networkinterface_name VARCHAR(20) DEFAULT 'eth0';
ALTER TABLE k8snode ADD COLUMN dns_name VARCHAR(200);
ALTER TABLE nodegroup ADD COLUMN vnets JSON;
UPDATE nodegroup set vnets='[{"name": "vnet", "availabilityzone": "us-west-1"}]';
ALTER TABLE nodegroup ALTER COLUMN vnets SET NOT NULL;
ALTER TABLE cluster DROP COLUMN region_id;
ALTER TABLE cluster ADD COLUMN region_name VARCHAR(50);
UPDATE cluster SET region_name='us';
ALTER TABLE cluster ALTER COLUMN region_name SET NOT NULL;

INSERT INTO vipstate(vipstate_name)
VALUES('Pending');
