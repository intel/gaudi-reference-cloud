-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
UPDATE runtime SET is_default=TRUE WHERE runtime_name = 'Containerd';

INSERT INTO provider(provider_name, lifecyclestate_id, is_default)
VALUES ('iks', 1, true),('rke2', 1, false);

INSERT INTO osimage(osimage_name, osname, osversion, lifecyclestate_id, cp_default, wrk_default)
VALUES ('Ubuntu-22-04', 'Ubuntu', '22.04', 1, true, true);

INSERT INTO backuptype (backuptype_name, mandatoryargs, lifecyclestate_id)
VALUES('S3', '["bucketname", "accesskey", "secretkey","folder","region","endpoint"]', 1);

INSERT INTO clusterstate(clusterstate_name)
VALUES('DeletePending');
