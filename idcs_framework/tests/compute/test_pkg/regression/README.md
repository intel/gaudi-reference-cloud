<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Compute VMaaS & BMaaS Regression Tests in Ginkgo

This package is used to run Compute VMaaS & BMaaS regression tests in Ginkgo.

## How to run the entire suite against staging or actual env

This step can be used to run the Regression tests in Jenkins.

### VMaaS
```bash
./run_tests.sh vm_directories --region=<<region>> --cloudAccountId=<<cloud-account>> --vnetName=<<vnet-name>> --test_env=<<environment>>
```

### BMaaS
```bash
./run_tests.sh bm_directories --region=<<region>> --cloudAccountId=<<cloud-account>> --vnetName=<<vnet-name>> --test_env=<<environment>>
```

## How to run the entire suite against local kind env

This step can be used to run the Regression tests in a local workstation. Cloudaccount and vnet is created automatically if it is kind.

### VMaaS
```bash
./run_tests.sh vm_directories --test_env=kind
```

### BMaaS
```bash
./run_tests.sh bm_directories --test_env=kind
```

## How to run the specific test case from local workstation against any environment

This step can be used to run the Specific Regression test or suite in a local workstation. Navigate to the path where test case and suite file is present.

Note: Skip cloudaccount and vnet if it is kind.

```bash
ginkgo --label-filter="<<ginkgo-label>>" -- --sshPublicKey="<<public-key>>" --cloudAccountId=<<cloud-account>> --vnetName=<<vnet-name>> --test_env=<<environment>>
```