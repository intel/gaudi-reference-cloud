<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Training Service in IKS

## Cluster Setup

### Label Nodes

All nodes in the cluster need to be labeled as either control-plane or data-plane. Use the label-node.sh script to add labels.

Example:

```sh
bash label-node.sh ng-azttpspmxq-1bbb5 --data-plane
bash label-node.sh ng-xrenjy24ie-760c1 --control-plane
```

### PVC Device Plugin

If this is a PVC-based cluster, follow the below steps to install the Intel Device Plugin.

```sh
#Note: Replace <RELEASE_VERSION> with the desired release tag or main to get devel images.
#Note: Add --dry-run=client -o yaml to the kubectl commands below to visualize the yaml content being applied.

# Start NFD - if your cluster doesn't have NFD installed yet
$ kubectl apply -k 'https://github.com/intel/intel-device-plugins-for-kubernetes/deployments/nfd?ref=<RELEASE_VERSION>'

# Create NodeFeatureRules for detecting GPUs on nodes
$ kubectl apply -k 'https://github.com/intel/intel-device-plugins-for-kubernetes/deployments/nfd/overlays/node-feature-rules?ref=<RELEASE_VERSION>'

# Create GPU plugin daemonset
$ kubectl apply -k 'https://github.com/intel/intel-device-plugins-for-kubernetes/deployments/gpu_plugin/overlays/nfd_labeled_nodes?ref=<RELEASE_VERSION>'
```


### Gaudi Device Plugin

If this is a Gaudi-based cluster, Use this command to install the Gaudi Device Plugin:

```sh
curl -s https://vault.habana.ai/artifactory/docker-k8s-device-plugin/habana-k8s-device-plugin.yaml | \
yq 'select(di == 1).spec.template.spec.nodeSelector."training.cloud.intel.com/role" = "data-plane"' | \
kubectl create -f -
```

Source: https://docs.habana.ai/en/v1.18.0/Installation_Guide/Additional_Installation/Kubernetes_Installation/#intel-gaudi-device-plugin-for-kubernetes


### CRD Installations

```sh
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update
helm install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm delete istio-base -n istio-system
kubectl delete namespace istio-system
```

```sh
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm repo update
helm install cnpg cnpg/cloudnative-pg -n cnpg-system --create-namespace
helm delete cnpg -n cnpg-system
kubectl delete namespace cnpg-system
```

### Registration DB Setup

If the Registration DB needs to be set up and there are no tables, complete the following steps:

- From a shell in the Registration API pod:

```sh
apt update && apt install -y postgresql
psql $(cat /etc/secrets/app/uri)

=> CREATE TABLE IF NOT EXISTS training_user (
    hashed_enterprise_id varchar(32) CONSTRAINT key PRIMARY KEY,
    random_user_id char(32) NOT NULL,
    launch_time timestamp with time zone NOT NULL DEFAULT 'epoch',
    coupon_expiration timestamp with time zone NOT NULL DEFAULT 'epoch',
    selected_training varchar(32) NOT NULL DEFAULT 'base'
);

=> CREATE TABLE IF NOT EXISTS feature_controller (
    coupon_enforce boolean NOT NULL DEFAULT FALSE
);
```

## Local Development Tips


### Docker signins needed 

```sh
docker login icir.cps.intel.com
docker login amr-fmext-registry.caas.intel.com
```
