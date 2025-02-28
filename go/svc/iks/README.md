<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## IDC Intel Kubernetes Service IKS

See [/README.md](../../../README.md) for common information.

### Overview

The service IKS runs the GRPC server.

The service IKS runs the GRPC-REST gateway that receives incoming REST requests,
converts to GRPC, and forwards the GRPC request to the IKS.

### Testing

#### Run Go tests

```bash
cd $(git rev-parse --show-toplevel)
make test
```

#### Deploy in kind-multicluster

```bash
export IDC_ENV=kind-multicluster
```


```bash
cd $(git rev-parse --show-toplevel)
make deploy-all-in-kind
```

#### Build and redeploy this service in an existing kind cluster

```bash
cd $(git rev-parse --show-toplevel)
make deploy-iks
```

#### Check the Ingress to get URL for Swagger and Postman 

```bash
kubectl config get-contexts  
CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE
*         kind-idc-global      kind-idc-global      kind-idc-global      
          kind-idc-us-dev-1    kind-idc-us-dev-1    kind-idc-us-dev-1    
          kind-idc-us-dev-1a   kind-idc-us-dev-1a   kind-idc-us-dev-1a  
```

## Switch to regional context

```bash
kubectl config use-context kind-idc-us-dev-1
```

```bash
kubectl get ing -n idcs-system
NAME                           CLASS   HOSTS                                                     ADDRESS     PORTS     AGE
us-dev-1-grpc-proxy-internal   nginx   dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local   localhost   80, 443   73m
us-dev-1-grpc-rest-gateway     nginx   dev.compute.us-dev-1.api.cloud.intel.com.kind.local       localhost   80, 443   73m
```

Add entry to Host machines /etc/hosts

```bash
<vm ip address > dev.vault.cloud.intel.com.kind.local dev.oidc.cloud.intel.com.kind.local dev.api.cloud.intel.com.kind.local dev.grpcapi.cloud.intel.com.kind.local dev.compute.us-dev-1.api.cloud.intel.com.kind.local
dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local dev.compute-api-server.us-dev-1.grpcapi.cloud.intel.com.kind.local dev.netbox.us-dev-1.api.cloud.intel.com.kind.local dev.baremetal-enrollment-api.us-dev-1.api.cloud.intel.com.kind.local
dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local dev.dhcp.us-dev-1a.infra.cloud.intel.com.kind.local
```

Swagger and postman will be available at :

```bash
https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local:10443/openapiv2/#/
```

```bash
kubectl get all -n idcs-system | grep iks
pod/us-dev-1-iks-545d747c78-44s48                               2/2     Running   0             17m
pod/us-dev-1-iks-db-postgresql-0                                1/1     Running   0             17m
service/us-dev-1-iks                        ClusterIP   10.96.183.12    <none>        8443/TCP   17m
service/us-dev-1-iks-db-postgresql          ClusterIP   10.96.134.13    <none>        5432/TCP   17m
service/us-dev-1-iks-db-postgresql-hl       ClusterIP   None            <none>        5432/TCP   17m
deployment.apps/us-dev-1-iks                                1/1     1            1           17m
replicaset.apps/us-dev-1-iks-545d747c78                               1         1         1       17m
statefulset.apps/us-dev-1-iks-db-postgresql       1/1     17m
```


#### Test IKS services using GRPC-REST gateway to the GRPC server

#### Test with Curl

```bash
#### Create a cloud account
export no_proxy=${no_proxy},.kind.local
export URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local
export TOKEN=$(curl "${URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
echo ${TOKEN}
 
export URL_PREFIX=https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local
export CLOUDACCOUNTNAME=${USER}@intel.com
go/svc/cloudaccount/test-scripts/cloud_account_create.sh
export CLOUDACCOUNT=$(go/svc/cloudaccount/test-scripts/cloud_account_get_by_name.sh | jq -r .id)
echo $CLOUDACCOUNT

### Create standard token
export URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local
export TOKEN=$(curl "${URL_PREFIX}/token?email=$USER@intel.com&group=IDC.Standard")
echo ${TOKEN}

### Create cluster
export URL_PREFIX=https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local:10443
go/svc/iks/test-scripts/cluster_create.sh

go/svc/iks/test-scripts/cluster_get.sh

#export cluster ID to create nodegroup
export CLUSTERID=
go/svc/iks/test-scripts/nodegroup_create.sh
### Use this token for IKS APIs in the postman along with cloud account ID.
API URL = https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local:10443
```


## Manage DBs for each environment
This documentation will list out all the changes that need to be done in order to test out various DBs.
IDC has multiple Helm Chart environments where it can pull different configuration details for each component. In this case, we will use pull different DB details. Unless the information is specified a specific environment, IDC will use `deployment/helmfile/defaults.yaml.gotmpl` to pull information.

### DB Vault Secrets 
These are the secrets that we need to set up or change before depoying the IKS API or Managed DB Pod.

#### **ManagedDb**
These secrets are used only for Managed DB instances and Unit testing. You can find these files in `local/secrets` folder.
- SO Username - `{env}-iks_db_username` -- Postgres 
- SO Password - `{env}-iks_db_user_password` -- {Any}
- RW Username - `{env}-iks_db_username` -- iks_user1
- RW Password - `{env}-iks_db_user_password` -- OnlyUseForTest!
- Encryption - `{env}-iks_db_encryption_keys` -- {Any of char length 32}

#### **DBaaS Instance (IDC Deployment)**
These secrets must be set in the Vault manually in order for the IKS API to connect. These values should have already been created.
- SO Username - `secrets/controlplane/{env}-iks/database/username` --  { Admin User provided by IDC DBaaS }
- SO Password - `secrets/controlplane/{env}-iks/databse/password` -- { Admin User Password provided by IDC DBaaS }
- RW Username - `secrets/controlplane/{env}-iks/database/username_rw` -- { One of: iks_user1 or iks_user2 depending on which is active }
- RW Password - `secrets/controlplane/{env}-iks/databse/password_rw` -- { Password of current active user }
- Encryption - `secrets/controlplane/{env}-iks/encryption_keys/{Key}` -- { Key-Value pair of ID and Char length 32 -- I.E "1":"abcd..789" }

#### **DBaaS Instance (Internal/DBaaS UI)**
These secrets must be set in the Vault manually in order for the IKS API to connect. These values should have already been created.
- SO Username - `secrets/controlplane/{env}-iks/database/username` --  { Admin User provided by DBaaS UI }
- SO Password - `secrets/controlplane/{env}-iks/databse/password` -- { Admin User Password provided by DBaaS UI }
- RW Username - `secrets/controlplane/{env}-iks/database/username_rw` -- { One of: iks_user1 or iks_user2 **NEEDS TO BE CREATED MANUALLY IN DBAAS.INTEL.COM UI** }
- RW Password - `secrets/controlplane/{env}-iks/databse/password_rw` -- { Password of user created in dbaas.intel.com}
- Encryption - `secrets/controlplane/{env}-iks/encryption_keys/{Key}` -- { Key-Value pair of ID and Char length 32 -- I.E "1":"abcd..789" }

### DB Connection Details:
Each environment has it's own configuration you can add to override the defaults configuration that comes from `deployment/helmfile/defaults.yaml.gotmpl`. You can find these files in `deployment/helmfile/environment/{env}.yaml.gotmpl`
- service - DB Hostname 
- name - DB name
- arg - Normally just sslmode
- port - DB Port 

#### EXAMPLE: Connect to Internal DBaaS instance in kind-singlecluster.yaml.gotmpl
```

regions:
  us-dev-1:
    iks:
      database:
        service: postgres5327-lb-or-in.dbaas.intel.com
        name: proto_idc_k8aas_db_local
        arg: sslmode=require
        port: 5433
```

### Steps to change the configuration
1. Update Vault Secrets. Depending on the environment, it can be replaced using /local/secrets and running `make deploy-vault-secrets` or by manually connecting to Vault and changing them manually if using DBaaS instance.
2. Update the specific environment in deployment/helmfile/environments to contain the DB connection details you wish to connect to.
    a. By default we configure the environment to use a postgres pod deployed in the cluster, but can also use a hosted DB instance such as a DBaaS DB.
3. (Optional) Prevent postgres pod from deploying. If you use a hosted DB, you can disable the postgres pod from deploying by removing the "deploy-iks-db" option in the main Makefile and setting the specific environemt's components.iksdb.enable to false.  Alternatively you can deploy the postgres pod, but the API will connect to the hosted DB, ignoring the pod.
