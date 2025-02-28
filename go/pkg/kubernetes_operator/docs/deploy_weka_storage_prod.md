# Production Deployment of IKS and Weka Storage Integration

This document contains the steps that need to be run before deploying the Weka Storage changes in any of the Production environments.

## Controlplane Bootstrap

As part of this change, the roles used by the IKS operators to connect to IKS clusters have been changed to include more permissions.

Copy `bootstrap-iks-controlplane` from staging 1

## Worker Bootstrap

As part of this change, the workers will now install and configure
the weka client.

Copy `bootstrap-iks-worker` from staging 1

## Kubernetes Operator Config

This is a global variable that holds the IDC grpc endpoint, so that
grpc clients can be initialized when controllers are being created. 

```
# IDC grpc url to connect to the idc services.
idcGrpcUrl: "internal-placeholder.com:443"
```

This is a new section that holds weka settings. Change settings accordingly.

```
weka:
  recreateGracePeriod: 5m
  softwareVersion: "4.2.2"
  clusterUrl: "10.45.116.186:14000" # THIS MUST BE CHANGED
  clusterPort: "14000"
  scheme: "http"
  filesystemGroupName: "tenantfsgroup" # THIS MUST BE CHANGED
  initialFilesystemSizeGB: "2"
  reclaimPolicy: "Delete"
  helmchartRepoUrl: "https://weka.github.io/csi-wekafs"
  helmchartName: "weka-fs-plugin"
  instanceTypes:
    bm-icp-gaudi2:
      mode: "dpdk"
      numCores: "4"
    bm-spr-gaudi2:
      mode: "dpdk"
      numCores: "4"
```

## IKS DB and Weka Addons

As part of this change 2 more weka addons must be added to the IKS DB.

### Weka storage class

Ensure `weka-storageclass-1.template` is in the production S3 bucket. Copy the template from staging 1.

Associate `weka-storageclass-1.template` to each available k8s version in production DB.

```
INSERT INTO addonversion (addonversion_name, name, version, admin_only, install_type, artifact_repo, lifecyclestate_id,default_addonargs,tags,onbuild,addonversion_type) Values ('weka-storageclass-0.0.0','weka-storageclass','0.0.0','true','kubectl-apply','s3://weka-storageclass-1.template',1,null,null,false,'weka');

# Repeat this for every k8s version available in production.
INSERT INTO addoncompatibilityk8s (addonversion_name,k8sversion_name) values ('weka-storageclass-0.0.0', '1.28.7');
```

### Weka helm chart

Associate the `weka helm chart` to each available k8s version in production DB.

```
INSERT INTO addonversion (addonversion_name, name, version, admin_only, install_type, artifact_repo, lifecyclestate_id,default_addonargs,tags,onbuild,addonversion_type) Values ('weka-fs-plugin-2.3.4','weka-fs-plugin','2.3.4','true','helm','csi-wekafsplugin/v2.3.4',1,null,null,false,'weka');

# Repeat this for every k8s version available in production.
INSERT INTO addoncompatibilityk8s (addonversion_name,k8sversion_name) values ('weka-fs-plugin-2.3.4', '1.28.7');
```

### Steps to Deploy to Production

As a production deployment we are supposed to make change to prod.json file under `deployment/universe_deployer/environments` path. Below are the places where we need to make changes.

```
# IKS API's and gRPC Changes in REGION 2
For any of the IKS API's or gRPC changes, we need update the commit hash of the changes that we are trying to deploy in prod.json file.
The Changes should be under `us-region-2/components/iks` for any IKS API changes and `us-region-2/components/grpc` for any gRPC changes.

# IKS Operator changes in REGION 2
For any of the IKS Operator changes, we need update the commit hash of the changes that we are trying to deploy in prod.json file.
The Changes should be under `us-region-2/availabilityZones/us-region-2a/components/iks` for any IKS Operator changes including both Kubernetes Operators and ILB Operators.

# Repeat above steps for other REGIONS like REGION 1.
```
