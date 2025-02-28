-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
INSERT INTO k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version)
VALUES
('1.27.11',1,'1.27',1,false),
('1.28.5',2,'1.28',1,false),
('1.28.7',1,'1.28',1,false),
('1.29.2',1,'1.29',1,false);

INSERT INTO public.runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id)
VALUES
('containerd-1.7.1','Containerd','1.7.1',1),
('containerd-1.7.7','Containerd','1.7.7',1);
UPDATE runtimeversion SET lifecyclestate_id = 2 WHERE runtimeversion_name = 'containerd-rke2';


INSERT INTO osimageinstance (osimageinstance_name, k8sversion_name, osimage_name, runtimeversion_name, created_date, nodegrouptype_name, lifecyclestate_id, runtime_name, provider_name, bootstrap_repo, imiartifact, hw_type, instancetypecategory, instancetypefamiliy) 
VALUES
('iks-vm-u22-cd-cp-1-28-5-24-01-04','1.28.5','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:57.158989','controlplane',1,'Containerd','iks','','iks-vm-u22-cd-cp-1-28-5-24-01-04','standard','VirtualMachine',''),
('iks-vm-u22-cd-wk-1-28-5-24-01-04','1.28.7','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:34.751797','worker',1,'Containerd','iks','','iks-vm-u22-cd-wk-1-28-5-24-01-04','standard','VirtualMachine','4th Generation Intel Xeoncalable processors'),
('iks-vm-u22-cd-cp-1-27-11-v20240227','1.27.11','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:48.850763','controlplane',1,'Containerd','iks','','iks-vm-u22-cd-cp-1-27-11-v20240227','standard','VirtualMachine',''),
('iks-vm-u22-cd-cp-1-28-7-v20240227','1.28.7','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:57.158989','controlplane',1,'Containerd','iks','','iks-vm-u22-cd-cp-1-28-7-v20240227','standard','VirtualMachine',''),
('iks-vm-u22-cd-cp-1-29-2-v20240227','1.29.2','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:42:04.636743','controlplane',1,'Containerd','iks','','iks-vm-u22-cd-cp-1-29-2-v20240227','standard','VirtualMachine',''),
('iks-vm-u22-cd-wk-1-27-11-v20240227','1.27.11','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:21.802826','worker',1,'Containerd','iks','','iks-vm-u22-cd-wk-1-27-11-v20240227','standard','VirtualMachine','4th Generation Intel Xeoncalable processors'),
('iks-vm-u22-cd-wk-1-28-7-v20240227','1.28.7','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:34.751797','worker',1,'Containerd','iks','','iks-vm-u22-cd-wk-1-28-7-v20240227','standard','VirtualMachine','4th Generation Intel Xeoncalable processors'),
('iks-vm-u22-cd-wk-1-29-2-v20240227','1.29.2','Ubuntu-22-04','containerd-1.7.1','2024-03-04 20:41:42.638542','worker',1,'Containerd','iks','','iks-vm-u22-cd-wk-1-29-2-v20240227','standard','VirtualMachine','4th Generation Intel Xeoncalable processors'),
('iks-vm-u22-cd-wk-1-29-2-v20240227-test','1.29.2','Ubuntu-22-04','containerd-1.7.1','2024-04-09 06:34:38.270098','worker',3,'Containerd','iks','','iks-vm-u22-cd-wk-1-29-2-v20240227-test','standard','VirtualMachine','');

INSERT INTO instancetype (instancetype_name, memory, cpu, nodeprovider_name, storage, is_default, lifecyclestate_id, displayname, description, instancecategory, imi_override, instancetypefamiliy) 
VALUES
('vm-spr-sml','16','8','Compute','20',false,1,'Small VM - Intel Xeon 4th Gencalable processor','','VirtualMachine',false,'4th Generation Intel Xeoncalable processors'),
('vm-iks-tny','16','4','Compute','15',false,1,'Tiny VM - Intel Xeon Gen Scalable processor','','VirtualMachine',false,'4th Generation Intel Xeoncalable processors'),
('vm-spr-med','32','16','Compute','32',false,1,'Medium VM - Intel Xeon 4th Gencalable processor','','VirtualMachine',false,'4th Generation Intel Xeoncalable processors'),
('vm-spr-lrg','64','32','Compute','64',false,2,'Large VM - Intel Xeon 4th Gencalable processor','','VirtualMachine',false,'4th Generation Intel Xeoncalable processors'),
('vm-spr-tny','8','4','Compute','10',true,1,'Tiny VM - Intel Xeon 4th Gencalable processor','','VirtualMachine',false,'4th Generation Intel Xeoncalable processors');

INSERT INTO k8scompatibility (runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name, instancetype_name, lifecyclestate_id)
VALUES
('Containerd','1.27.11','Ubuntu-22-04','iks-vm-u22-cd-cp-1-27-11-v20240227','iks-vm-u22-cd-wk-1-27-11-v20240227','iks','vm-spr-tny',1),
('Containerd','1.27.11','Ubuntu-22-04','iks-vm-u22-cd-cp-1-27-11-v20240227','iks-vm-u22-cd-wk-1-27-11-v20240227','iks','vm-spr-lrg',1),
('Containerd','1.27.11','Ubuntu-22-04','iks-vm-u22-cd-cp-1-27-11-v20240227','iks-vm-u22-cd-wk-1-27-11-v20240227','iks','vm-spr-med',1),
('Containerd','1.27.11','Ubuntu-22-04','iks-vm-u22-cd-cp-1-27-11-v20240227','iks-vm-u22-cd-wk-1-27-11-v20240227','iks','vm-iks-tny',1),
('Containerd','1.27.11','Ubuntu-22-04','iks-vm-u22-cd-cp-1-27-11-v20240227','iks-vm-u22-cd-wk-1-27-11-v20240227','iks','vm-spr-sml',1);

INSERT INTO k8scompatibility (runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name, instancetype_name, lifecyclestate_id)
VALUES
('Containerd','1.28.5','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-5-24-01-04','iks-vm-u22-cd-wk-1-28-5-24-01-04','iks','vm-spr-tny',1),
('Containerd','1.28.5','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-5-24-01-04','iks-vm-u22-cd-wk-1-28-5-24-01-04','iks','vm-spr-lrg',1),
('Containerd','1.28.5','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-5-24-01-04','iks-vm-u22-cd-wk-1-28-5-24-01-04','iks','vm-spr-med',1),
('Containerd','1.28.5','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-5-24-01-04','iks-vm-u22-cd-wk-1-28-5-24-01-04','iks','vm-iks-tny',1),
('Containerd','1.28.5','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-5-24-01-04','iks-vm-u22-cd-wk-1-28-5-24-01-04','iks','vm-spr-sml',1);

INSERT INTO k8scompatibility (runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name, instancetype_name, lifecyclestate_id)
VALUES
('Containerd','1.28.7','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-7-v20240227','iks-vm-u22-cd-wk-1-28-7-v20240227','iks','vm-spr-tny',1),
('Containerd','1.28.7','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-7-v20240227','iks-vm-u22-cd-wk-1-28-7-v20240227','iks','vm-spr-lrg',1),
('Containerd','1.28.7','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-7-v20240227','iks-vm-u22-cd-wk-1-28-7-v20240227','iks','vm-spr-med',1),
('Containerd','1.28.7','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-7-v20240227','iks-vm-u22-cd-wk-1-28-7-v20240227','iks','vm-iks-tny',1),
('Containerd','1.28.7','Ubuntu-22-04','iks-vm-u22-cd-cp-1-28-7-v20240227','iks-vm-u22-cd-wk-1-28-7-v20240227','iks','vm-spr-sml',1);

INSERT INTO k8scompatibility (runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, provider_name, instancetype_name, lifecyclestate_id) 
VALUES
('Containerd','1.29.2','Ubuntu-22-04','iks-vm-u22-cd-cp-1-29-2-v20240227','iks-vm-u22-cd-wk-1-29-2-v20240227','iks','vm-spr-tny',1),
('Containerd','1.29.2','Ubuntu-22-04','iks-vm-u22-cd-cp-1-29-2-v20240227','iks-vm-u22-cd-wk-1-29-2-v20240227','iks','vm-spr-lrg',1),
('Containerd','1.29.2','Ubuntu-22-04','iks-vm-u22-cd-cp-1-29-2-v20240227','iks-vm-u22-cd-wk-1-29-2-v20240227','iks','vm-spr-med',1),
('Containerd','1.29.2','Ubuntu-22-04','iks-vm-u22-cd-cp-1-29-2-v20240227','iks-vm-u22-cd-wk-1-29-2-v20240227','iks','vm-iks-tny',1),
('Containerd','1.29.2','Ubuntu-22-04','iks-vm-u22-cd-cp-1-29-2-v20240227','iks-vm-u22-cd-wk-1-29-2-v20240227','iks','vm-spr-sml',1);

UPDATE nodeprovider SET nodeprovider_name = 'Harvester' WHERE nodeprovider_name = 'caas-harvester';
UPDATE nodeprovider SET is_default = 'true' WHERE nodeprovider_name = 'Compute';
UPDATE nodeprovider SET is_default = 'false' WHERE nodeprovider_name = 'Harvester';

INSERT INTO addonversion (addonversion_name, name, version, lifecyclestate_id, admin_only, install_type, artifact_repo, onbuild, addonversion_type)
VALUES
('konnectivity-agent-0.0.0','konnectivity-agent','0.0.0',1,true,'kubectl-apply','s3://konnectivity-agent.template',true, 'kubernetes'),
('kube-proxy-1.27.5','kube-proxy','1.27.1',1,true,'kubectl-apply','s3://kube-proxy-k1275-1.template',true, 'kubernetes'),
('coredns-1.11.1','coredns','1.11.1',1,true,'kubectl-apply','s3://coredns-1111-1.template',true, 'kubernetes'),
('calico-config-3.26.3','calico-config','3.26.3',1,true,'kubectl-apply','s3://calico-config-3263-1.template',true, 'kubernetes'),
('konnectivity-agent-0.1.4','konnectivity-agent','0.1.4',1,true,'kubectl-apply','s3://konnectivity-agent-014-1.template',true, 'kubernetes'),
('konnectivity-agent-0.1.6','konnectivity-agent','0.1.6',1,true,'kubectl-apply','s3://konnectivity-agent-016-1.template',true, 'kubernetes'),
('calico-operator-3.25.2','calico-operator','3.25.2',1,true,'kubectl-replace','s3://calico-operator-3252-1.template',true, 'kubernetes'),
('calico-operator-3.27.2','calico-operator','3.27.2',1,true,'kubectl-replace','s3://calico-operator-3272-1.template',true, 'kubernetes'),
('calico-config-general','calico-config','general',1,true,'kubectl-apply','s3://calico-config-general-1.template',true, 'kubernetes'),
('konnectivity-agent-0.1.7','konnectivity-agent','0.1.7',1,true,'kubectl-apply','s3://konnectivity-agent-017-1.template',true, 'kubernetes'),
('kube-proxy-1.27.11','kube-proxy','1.27.11',1,true,'kubectl-apply','s3://kube-proxy-12711-1.template',true, 'kubernetes'),
('kube-proxy-1.28.7','kube-proxy','1.28.7',1,true,'kubectl-apply','s3://kube-proxy-1287-1.template',true, 'kubernetes'),
('kube-proxy-1.29.2','kube-proxy','1.29.2',1,true,'kubectl-apply','s3://kube-proxy-1292-1.template',true, 'kubernetes');

INSERT INTO addoncompatibilityk8s (addonversion_name, k8sversion_name) 
VALUES
('kube-proxy-1.27.11','1.27.11'),
('coredns-1.11.1','1.27.11'),
('calico-operator-3.27.2','1.27.11'),
('calico-config-general','1.27.11'),
('konnectivity-agent-0.1.7','1.27.11'),
('kube-proxy-1.28.7','1.28.5'),
('coredns-1.11.1','1.28.5'),
('calico-operator-3.27.2','1.28.5'),
('calico-config-general','1.28.5'),
('konnectivity-agent-0.1.7','1.28.5'),
('kube-proxy-1.28.7','1.28.7'),
('coredns-1.11.1','1.28.7'),
('calico-operator-3.27.2','1.28.7'),
('calico-config-general','1.28.7'),
('konnectivity-agent-0.1.7','1.28.7'),
('kube-proxy-1.29.2','1.29.2'),
('coredns-1.11.1','1.29.2'),
('calico-operator-3.27.2','1.29.2'),
('calico-config-general','1.29.2'),
('konnectivity-agent-0.1.7','1.29.2');






UPDATE defaultconfig SET value = 'us-dev-1' WHERE name = 'region';
UPDATE defaultconfig SET value = 'us-dev-1a' WHERE name = 'availabilityzone';
UPDATE defaultconfig SET value = 'us-dev-1a-default' WHERE name = 'vnet';
UPDATE defaultconfig SET value = '182' WHERE name = 'ilb_environment';
UPDATE defaultconfig SET value = '2802' WHERE name = 'ilb_usergroup';
UPDATE defaultconfig SET value = '182' WHERE name = 'ilb_customer_environment';
UPDATE defaultconfig SET value = '2802' WHERE name = 'ilb_customer_usergroup';
-- Update cp_cloudaccount
-- update restricted


/*StorageProvider*/
INSERT INTO addonversion (addonversion_name, name, version, lifecyclestate_id, admin_only, install_type, artifact_repo,onbuild, addonversion_type)
VALUES
('weka-storageclass-1287-1','weka-storageclass','1.28.7',1,true,'kubectl-apply','s3://weka-storageclass-1287-1.template',false,'weka'),
('weka-fs-plugin-2-3-4','weka-fs-plugin','2.3.4',1,true,'helm','csi-wekafsplugin/v2.3.1',false,'weka');

INSERT INTO addoncompatibilityk8s (addonversion_name, k8sversion_name) 
VALUES
('weka-storageclass-1287-1','1.27.11'),
('weka-storageclass-1287-1','1.28.7'),
('weka-storageclass-1287-1','1.29.2'),
('weka-fs-plugin-2-3-4','1.27.11'),
('weka-fs-plugin-2-3-4','1.28.7'),
('weka-fs-plugin-2-3-4','1.29.2');


