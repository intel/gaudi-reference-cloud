<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Compute VMaaS E2E Tests in Bazel and Jenkins

This package is used to run Compute VMaaS end-to-end tests in Bazel and Jenkins.
It runs all required IDC services in a single kind cluster.

## External Dependencies

This suite is dependent on the following external systems.

- Harvester cluster
- SSH proxy server

## Prerequisites for running VMaaS E2E tests

Create the following files with the appropriate secrets.

- local/secrets/test-e2e-compute-vm/harvester-kubeconfig/harvester1
- local/secrets/test-e2e-compute-vm/ssh-proxy-operator/host_public_key
- local/secrets/test-e2e-compute-vm/ssh-proxy-operator/id_rsa
- local/secrets/test-e2e-compute-vm/ssh-proxy-operator/id_rsa.pub

## Prepare Jenkins

Copy secrets to Jenkins executors.

```bash
tar -czvf local/secrets/secrets-test-e2e-compute-vm.tgz local/secrets/test-e2e-compute-vm/{harvester-kubeconfig/harvester1,ssh-proxy-operator/*}
scp local/secrets/secrets-test-e2e-compute-vm.tgz sdp@10.165.160.171:/tmp
ssh sdp@10.165.160.171 sudo -u intelcloudservices cp /tmp/secrets-test-e2e-compute-vm.tgz /home/intelcloudservices
```

## Deploy all in Kind and run VMaaS E2E tests

This step can be used to run the E2E tests in a local workstation.
The variable `SKIP_TEAR_DOWN=1` will keep the Kind cluster running after the test.

```bash
SKIP_TEAR_DOWN=1 make test-e2e-compute-vm |& tee local/test-e2e-compute-vm.log
```

## Use a previously deployed Kind cluster to run VMaaS E2E tests

To speed up iterative development and testing, this section can be followed to run E2E tests in a previously deployed Kind cluster.
IDC services will not be redeployed.

Search `local/test-e2e-compute-vm.log` for "VaultToken" and copy the token to the file `local/secrets/test-e2e-compute-vm/VAULT_TOKEN`.

```bash
SKIP_TEAR_DOWN=1 make test-e2e-compute-vm-quick |& tee local/test-e2e-compute-vm-quick.log
```

## Access a Kind cluster deployed in Bazel

```bash
kind export kubeconfig --name test-e2e-compute-vm-global
```

## Access a Kind cluster deployed in Bazel and Jenkins

```bash
ssh sdp@${HOST_IP}
sudo -u intelcloudservices -i
cd /home/intelcloudservices/workspace/BMAAS-Orchestrator_PR-*
./bazel-bin/go/test/compute/e2e/vm/vm_test_/vm_test.runfiles/kind_linux_amd64/file/kind export kubeconfig --name test-e2e-compute-vm-global
./bazel-bin/go/test/compute/e2e/vm/vm_test_/vm_test.runfiles/kubectl_linux_amd64/file/kubectl describe -n idcs-system deployment/cloudaccount
```

## Run debug pod

```bash
kubectl apply -f deployment/hack/debug-tools-pod.yaml
```
