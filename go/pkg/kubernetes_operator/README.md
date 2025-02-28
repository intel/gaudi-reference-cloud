<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Kubernetes Operator for Intel Kubernetes Service "IKS"

This operator provisions and manages the life cycle of Kubernetes clusters (managed clusters) built on top
of IDC compute instances.

## Getting Started

### Update CRDs

If the spec of a CRD API is changed we need to run the following command to generate the updated CRD.

```sh
make generate-k8s
```

### Install helm chart crds

```sh
make deploy-kubernetes-crds
```

### Install helm chart operator

Before installing the kubernetes operator ensure kubernetes crds, ilb crds and ilb operator are also installed.

```sh
make deploy-kubernetes-operator
```

### Uninstall helm chart crds

```sh
make undeploy-kubernetes-crds
```

### Uninstall helm chart operator

```sh
make undeploy-kubernetes-operator
```