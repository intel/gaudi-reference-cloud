<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Compute BMaaS E2E Tests in Bazel and Jenkins

This package is used to run Compute BMaaS end-to-end tests in Bazel and Jenkins.
It runs all required IDC services in a single kind cluster.

## External Dependencies

You must run Ubuntu 22.04. Previous versions are not supported.

Enable hardware virtualization on your VM by raising a [Developer Infrastructure & Operations request](https://opensource.intel.com/jira/servicedesk/customer/portal/1/group/12).

Run the following command to verify if hardware virtualization is enabled:

```bash
sudo kvm-ok
```

This suite is dependent on the following external systems.

- Harvester cluster (required because instance-scheduler must have at least one working Harvester cluster)

## Prerequisites for running BMaaS E2E tests

Create the following files with the appropriate secrets.

- local/secrets/test-e2e-compute-bm/harvester-kubeconfig/harvester1

## Ensure to remove all the existing running and exited e2e containers to avoid network issues during enrollment.

Run the following command.

```bash
docker rm -f $(docker ps -a -q --filter "name=nginx*")
docker rm -f $(docker ps -a -q --filter "name=idc-registry*")
docker rm -f $(docker ps -a -q --filter "name=idc-global*")
```

## Deploy all in Kind and run BMaaS E2E tests

This step can be used to run the E2E tests in a local workstation.
The variable `SKIP_TEAR_DOWN=1` will keep the Kind cluster running after the test.

```bash
SKIP_TEAR_DOWN=1 make test-e2e-compute-bm |& tee local/test-e2e-compute-bm.log
```

## Access the Kind cluster if deployed using idc TestEnvId in Bazel (1 kind Cluster format)

```bash
kind export kubeconfig --name idc-global
```