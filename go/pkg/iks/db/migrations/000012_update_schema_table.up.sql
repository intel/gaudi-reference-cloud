-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
INSERT INTO defaultconfig(name, value)
VALUES ('kubecfg_cert_expiration_check','5'), ('ilb_allowed_ports','80,443');

INSERT INTO defaultconfig(name, value)
VALUES ('restrict_create_cluster','true');

ALTER TABLE cloudaccountextraspec ADD COLUMN active_account_create_cluster BOOLEAN DEFAULT FALSE;

/* NEW ENCRYPTION METHOD */
ALTER TABLE cluster_extraconfig ADD COLUMN nonce VARCHAR(63) NOT NULL;
ALTER TABLE cluster_extraconfig ADD COLUMN encryptionkey_id INT NOT NULL;

/* NEW SSH Key Logic */
ALTER TABLE cluster_extraconfig ADD COLUMN cluster_ssh_key_name VARCHAR(80);
ALTER TABLE cluster_extraconfig ADD COLUMN cluster_ssh_key TEXT;
ALTER TABLE cluster_extraconfig ADD COLUMN cluster_ssh_pub_key TEXT;

ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_cacrt;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_cakey;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_etcdcacrt;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_etcdcakey;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_etcdrotationkeys;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_sapub;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_sakey;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_cp_reg;
ALTER TABLE cluster_extraconfig DROP CONSTRAINT check_nospace_wk_reg;


/* METRICS UPDATE*/
ALTER TABLE defaultconfig ALTER COLUMN value TYPE VARCHAR(80);

/* FUTRE ABILITY TO CREATE OVERRIDES */
ALTER TABLE instancetype ADD COLUMN instancetypefamiliy VARCHAR(100);
ALTER TABLE instancetype ALTER COLUMN displayname TYPE VARCHAR(100);
ALTER TABLE osimageinstance ADD COLUMN instancetypecategory VARCHAR(50);
ALTER TABLE osimageinstance ADD COLUMN instancetypefamiliy VARCHAR(100);
