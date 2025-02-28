<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Deployment

All services are designed to be run on the Kubernetes cluster this creates two requirements for the deployment:
1. OCI Image
2. Kubernetes Manifests

## OCI Image

OCI Image build is handled by `bazel` [build](development.md#oci-images).

## Manifests

Kubernetes manifests are part of the repository and resides in services `/manifests` folder. Having following structure:
```
./base
./stage
    ./<cluster-id>
        ./kustomization.yaml
        ./...
./prod
```

Following structure utilize [Kustomize](https://kustomize.io). To describe and collect final manifests. This manifests are referenced by
[Flux](https://fluxcd.io) controller inside the cluster. `Flux` monitors `Harbor` registry and based on the rules fetches OCI images with
manifests and apply them inside the cluster.

## Infrastructure

Common clusters infrastructure are resides in `infrastructure` folder, this folder have same structure as `manifests` folder for services with
one notable exception. This folder controls `Flux` `OCI Repository` and `Kustomization` resources. Examples:
```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: storage-controller
  namespace: flux-system
spec:
  interval: 1m0s
  url: oci://amr-idc-registry-pre.infra-host.com/idc-storage/flux/storage_controller
  ref:
    tag: latest
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: storage-controller
  namespace: flux-system
spec:
  interval: 1m
  prune: true
  sourceRef:
    kind: OCIRepository
    name: storage-controller
  path: ./stage/pdx05-k01-stcp-cluster/
```

Detailed explanation of the these resources can be found at the [Flux website](https://fluxcd.io/flux/cheatsheets/oci-artifacts/).

## Releases

Repository follows [trunk-based development](https://trunkbaseddevelopment.com/) branching model with main branch being automatically deployed.

This repository is responsible for building the image and setting correct SemVer tag to it. Flux is used to automate the process of commiting those new tag to [GitOps repository](https://github.com/intel-innersource/applications.infrastructure.idcstorage.flux).
By default each staging environment has 1.x.x-env99 semver range in it's image policy, meaning that any patch increments or minor releases are going to be deployed as soon as code is merged to main.
If one needs to deploy staging environment from the branch [Storage Controller Build](https://github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/actions/workflows/storage-controller-update.yml)
CI job should be invoked wich is going to create new tag with increased patch version and "env99" prerelease name ensuring that only intended environment receives the build.
Consecutive runs will produce increase in prerelease number (e.g. 1.0.0-pdx05.2)

If you need to restore environment to the main state tag the main with the semver increment.
```
git tag -a 1.0.0-pdx05.N+1 main # where N+1 means you have to check the latest tag and increment it's value

git push origin 1.0.0-pdx05.N+1
```

In order to deploy sds-controller to production one needs to publish new release on github.

Patch increases are deployed to production automatically after new github release is created.
Minor or Major releases will be deployed to production after you change semver range in the ImagePolicy for the environment you wish to update.
