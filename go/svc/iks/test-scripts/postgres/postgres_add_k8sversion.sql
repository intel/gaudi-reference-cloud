-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
/*
K8SVERSION i.e. 1.27
K8SPATCHVERSION i.e. 1.27.5
CPIMINAME i.e. iks-u22-cd-cp-1.27.5-23-9-18
*/

SELECT k8sversion_name, lifecyclestate_id FROM k8sversion;
SELECT osimageinstance_name, k8sversion_name, provider_name, nodegrouptype_name,imiartifact , lifecyclestate_id, created_date from osimageinstance;
SELECT k8sversion_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name FROM k8scompatibility;

BEGIN;

INSERT INTO k8sversion(k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version)
Select 'K8SPATCHVERSION',lifecyclestate_id, 'K8SVERSION', '1',false from lifecyclestate where name = 'Staged';

-- Commented when image already exists
--INSERT INTO osimageinstance (osimageinstance_name, runtimeversion_name, runtime_name, osimage_name, k8sversion_name, provider_name, nodegrouptype_name,bootstrap_repo,imiartifact , lifecyclestate_id, created_date)
--Select 'CPIMINAME','CONTAINERDNAME','Containerd','Ubuntu-22-04','K8SPATCHVERSION','iks','controlplane','https://github.com/intel-sandbox/k8aas-bootstrap/blob/main/bootstrap.sh','CPIMINAME',lifecyclestate_id,NOW() from lifecyclestate where name = 'Active';

--INSERT INTO osimageinstancecomponent(component_name, osimageinstance_name, version, artifact_repo)
--VALUES ('etcd', 'CPIMINAME', '3.5.9', 'somewhere'),('node-exporter', 'CPIMINAME', '1.6.0', 'somewhere');

WKSELECT

INSERT INTO addoncompatibilityk8s(addonversion_name, k8sversion_name)
VALUES ('ADDONPROXY','K8SPATCHVERSION'), ('ADDONCOREDNS','K8SPATCHVERSION'), ('ADDONCALOPER','K8SPATCHVERSION'), ('ADDONCALCONF','K8SPATCHVERSION'), ('ADDONKONN','K8SPATCHVERSION');

COMMIT;

SELECT k8sversion_name, lifecyclestate_id FROM k8sversion;
SELECT osimageinstance_name, k8sversion_name, provider_name, nodegrouptype_name,imiartifact , lifecyclestate_id, created_date from osimageinstance;
SELECT k8sversion_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name, instancetype_name FROM k8scompatibility;
