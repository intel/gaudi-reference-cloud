-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
SELECT osimageinstance_name, k8sversion_name, provider_name, nodegrouptype_name,imiartifact , lifecyclestate_id, created_date from osimageinstance;

INSERT INTO osimageinstance (osimageinstance_name, runtimeversion_name, runtime_name, osimage_name, k8sversion_name, provider_name, nodegrouptype_name,bootstrap_repo,imiartifact , instancetypecategory, instancetypefamiliy, lifecyclestate_id, created_date)
Select 'IMINAME','CONTAINERDNAME','Containerd','Ubuntu-22-04','K8SPATCHVERSION','iks','IMITYPE','https://github.com/intel-sandbox/k8aas-bootstrap/blob/main/bootstrap.sh','IMINAME','INSTANCETYPECAT','INSTANCETYPEFAM',lifecyclestate_id,NOW() from lifecyclestate where name = 'Active';

SELECT osimageinstance_name, k8sversion_name, provider_name, nodegrouptype_name,imiartifact , lifecyclestate_id, created_date from osimageinstance;
