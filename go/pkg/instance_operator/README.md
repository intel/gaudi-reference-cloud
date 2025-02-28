<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Instance Operator

## Description

This operator helps provision and deploy infrastructure components which includes virtual machines, bare metal instances and networking.

## Overview

See the following documents.

- [IDC Design Documents](https://internal-placeholder.com/x/KOlZoQ)

## Prerequisites

To install the Ginkgo CLI:
```sh
go install github.com/onsi/ginkgo/ginkgo
```

## Developer Information

#### Use Kubebuilder (Obsolete)

Kubebuilder was used to bootstrap Instance Operator but is no longer used in this project.
The information below is obsolete but may be useful.

```sh
kubebuilder version
# Version: main.version{KubeBuilderVersion:"3.6.0", KubernetesVendor:"1.24.1", GitCommit:"f20414648f1851ae97997f4a5f8eb4329f450f6d", BuildDate:"2022-08-03T11:47:17Z", GoOs:"linux", GoArch:"amd64"}

# Initialization.
kubebuilder init --domain intel.com \
--repo github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator \
--plugins=go/v4-alpha

# Enable multigroup (see https://kubebuilder.io/migration/multi-group.html)
kubebuilder edit --multigroup=true
```

#### Generate API types, controllers, configs, and deepcoppy objects

```
# Create Instance kind.
kubebuilder create api --group private.cloud --version v1alpha1 --kind Instance --resource --controller
# Create InstanceType kind.
kubebuilder create api --group private.cloud --version v1alpha1 --kind InstanceType --resource --controller=false
# Create a VNet kind.
kubebuilder create api --group private.cloud --version v1alpha1 --kind VNet --resource --controller=false
# Create a ProviderVlan kind which is global.
kubebuilder create api --group private.cloud --version v1alpha1 --kind ProviderVlan --namespaced=false --resource --controller=false
# Create SshPublicKey kind.
kubebuilder create api --group private.cloud --version v1alpha1 --kind SshPublicKey --resource --controller=false
# Create IpAddress kind.
kubebuilder create api --group private.cloud --version v1alpha1 --kind IpAddress --namespaced=false --resource --controller=false
```

After creating new apis, move the new files to ../k8s/apis.

#### Generate controller interfaces using `wrangler`

Add your new API to `pkg/codegen/main.go`, for example:
```go
Groups: map[string]args.Group{
			cloudv1alpha1.SchemeGroupVersion.Group: {
				Types: []interface{}{
					cloudv1alpha1.Instance{},
					cloudv1alpha1.InstanceType{},
					// Your new API type here
                                        cloudv1alpha1.Foo{},
				},
				GenerateTypes:     false,  // We are not generating the API types here since they are taken care by `kubebuilder`.
			},
```

#### Generate clients, informers, and listers using K8s `code-generator` tools

Add these tags above the API type definition in order to generate a client:
```go
// +genclient
type Foo struct {
}
```

Now you can run `make manifests generate-all` to generate both YAML manifests and Go code for your new API types.

## Storing and reading secrets in Operators

This is a reference documentation for developers for adding/reading secrets in Operators

### 1. Generate new secrets into file

Here is an example for generating ssk keys into files, the below command generates ssh private and public keys into given <secret_dir>/id_rsa and <secret_dir>/id_rsa.pub respectively

```sh
 $ ssh-keygen -f <secret dir> # -f  - "File" Specifies name of the file in which to store the created key.
 ```

> We can also have literals such as username, password etc saved in <secret_file>

Also, developer should update below targets in `Makefile` as best practices

 > `secrets`: should be generating new secrets

 > `clean-secrets`: should remove secrets that are generated

 > `show-secrets`: should be able to display the secrets in std output

 # 2. Create Kubernetes secrets for the generated/given keys

There are multiple ways to create secrets in Kubernetes as given in [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/), Below are 2 ways of them

1. Using `kubectl`
 Eg:
 ```sh
 $ kubectl create secret generic idcs-operator-ssh --from-file=ssh-privatekey=/path/to/id_rsa --from-file=ssh-publickey=/path/to/.ssh/id_rsa.pub

 $ kubectl create secret generic test-secret-literal --from-literal=username=testuser --from-literal=password=testpassword
 ```

### 3. Mount secrets in your deployment.yml file by mapping name of the secrets to container's file system path via volume mounts

```yml
# manager.yaml
spec:
  template:
    spec:
      containers:
        volumeMounts:
          - name: ssh-access
            mountPath: /home/idc/.ssh # container will have /home/idc/.ssh/id_rsa and /home/idc/.ssh/id_rsa.pub files mounted
          - name: auth-access
            mountPath: /var/run/secrets/kubernetes.io # container will have /var/run/secrets/kubernetes.io/username and /var/run/secrets/kubernetes.io/password files mounted
    volumes:
      - name: ssh-access
        secret:
          secretName: idcs-operator-ssh # secret name that was created in step 2
      - name: auth-access
        secret:
          secretName: auth-secret # secret name that was created in step 2
```

### 4. Read secret in your application to construct controller's object

```Go
// main.go
privateKey, err := os.ReadFile(PrivateKeyFilePath) // Here 'privateKeyFilePath' is referring to <secret dir>/id_rsa path
```
