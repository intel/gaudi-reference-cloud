<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# sdn-controller

## Description

### Running on a kind cluster

Follow the steps [here](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc#prepare-build-workstation) to prepare the environment. 


Set Raven username and password
```
export RAVEN_USERNAME=<your Raven user name>
export RAVEN_PASSWD=<your Raven password>
```

Update the helmfile `Values.managerConfig.controllerManagerConfigYaml.controllerConfig.switchBackendMode` to make SDN-Controller talk to a mock switch or a real switch. Set this `mock` to talk to mock switches. `mock` switch is designed for testing the SDN locally during the development. When switchBackendMode is set to `eapi` SDN will talk to a real switch via eAPI. This is the recommended approach for production. When switchBackendMode is set to `raven`, SDN will talk to Raven for making switch config changes. (please refer to the Configuration section for more details on the SDN-Controller configuration). 

Update the helmfile `Values.managerConfig.controllerManagerConfigYaml.ravenConfig.environment` to set the Raven environment. 

Go to the IDC project root folder `/<your_workspace>/frameworks.cloud.devcloud.services.idc/` and run below commands to prepare secrets: 
```
export IDC_ENV=kind-multicluster
export KIND_MULTICLUSTER=1
make secrets
```
Modify `${SECRETS_DIR}/EAPI_USERNAME` & `${SECRETS_DIR}/EAPI_PASSWD` to your eapi username & password.

Create kind clusters, build & deploy all helm charts / containers & deploy:
```
make deploy-all-in-kind
```

### Running as a local process (for development and testing)

SDN-Controller read configuration from a config file. We have created one in the folder `/<your_workspace>/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/config/controller_manager_config.yaml` for running local process.

If you need to talk to a Raven server during your local test, make change to `/<your_workspace>/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/config/raven_credentials.yaml` to provide the username and password. 

To run SDN-Controller, go to the project folder `/<your_workspace>/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller`

run `make install` to install the CRDs

run `make run` or `go run main.go` to run the sdn-controller


### Configuration
During SDN-Controller’s start-up, it reads the configurations from a SDNControllerConfig file which is injected into the Pod via a ConfigMap. 
The ConfigMap can be found here `/<your_workspace>/frameworks.cloud.devcloud.services.idc/deployment/charts/sdn-controller/templates/manager-config.yaml`. 

Below is an example of the configMap with description of each field
```
apiVersion: idcnetwork.intel.com/v1alpha1
kind: SDNControllerConfig
controllerConfig:
  # SwitchSecretsPath is the location of the shared eAPI secret file
  switchSecretsPath: /vault/secrets/eapi
  # SwitchBackendMode specifies which type of switch backend that will be used.
  # set switchBackendMode to `eapi` to allow SDN to talk to a switch directly via eAPI.
  # set it to `raven` to use Raven as backend,
  # and `mock` to use a mock switch.  
  switchBackendMode: "mock"
  # DataCenter is a list of data center names with ":" as the separator. e.g., "fxhb3p3s:fxhb3p3r"
  # During SDN-Controller's start-up, it calls Raven to fetches the switches data and persist it in the format of CRD. The field `dataCenter` here specifies which data center the SDN-Controller should fetch the switches from.
  dataCenter: fxhb3p3r
  # PortResyncPeriodInSec specifies the frequency of the periodic SwitchPort reconciliation.
  portResyncPeriodInSec: 300
  # NetworkNodeResyncPeriodInSec specifies the frequency of the periodic networkNode reconciliation.
  networkNodeResyncPeriodInSec: 300
  # BMHResyncPeriodInSec specifies the frequency of the periodic BareMetalHost reconciliation.
  bmhResyncPeriodInSec: 300
  # SwitchResyncPeriodInSec specifies the frequency of the periodic Switch reconciliation.
  switchResyncPeriodInSec: 300
  # NodeGroupResyncPeriodInSec specifies the frequency of the periodic NodeGroup reconciliation.
  nodeGroupResyncPeriodInSec: 300
  # MaxConcurrentReconciles specifies the number of threads of the SwitchPort reconciler.
  maxConcurrentReconciles: 10
  # MaxConcurrentSwitchReconciles specifies the number of threads of the Switch reconciler.
  maxConcurrentSwitchReconciles: 10
  # MaxConcurrentNetworkNodeReconciles specifies the number of threads of the NetworkNode reconciler.
  maxConcurrentNetworkNodeReconciles: 10
  # MaxConcurrentNodeGroupReconciles specifies the number of threads of the NodeGroup reconciler.
  maxConcurrentNodeGroupReconciles: 10
  # BMHClusterKubeConfigFilePath is the location of the BMH Cluster kubeConfig file.  This is required if `switchPortImportSource` is set to `bmh`
  bmhClusterKubeConfigFilePath: "/vault/secrets/bmhkubeconfig"
  # EnableReadOnlyMode specifies if read only mode is enabled or not.
  # (Raven backend is unaffected)
  enableReadOnlyMode: false
  # SwitchImportPeriodInSec specifies the interval of getting switches from Raven/Netbox and then try to create SwitchCR for it.
  switchImportPeriodInSec: 300
  # PoolsConfigFilePath is the location of the pool configuration file
  poolsConfigFilePath: "/pool_config.json"
  # NodeGroupToPoolMappingSource specifies which source SDN Controller should get the Group to Pool mappings. options are "local" and "crd"
  nodeGroupToPoolMappingSource: "local"
  # NodeGroupToPoolMappingConfigFilePath specifies the location of the NodeGroup to Pool mapping file.
  # This is used for the LocalPoolMapping, we don't need this if we store the mapping in Netbox.
  nodeGroupToPoolMappingConfigFilePath: "/group_pool_mapping_config.json"
  # UseDefaultValueInPoolForMovingNodeGroup specifies if NOOP Vlan or BGP value should be used a NodeGroup is created or moved to a new Pool.
  useDefaultValueInPoolForMovingNodeGroup: false
  # SwitchImportSource specifies where to import the Switch data. options: "netbox", "raven" and "none"
  switchImportSource: raven
  # SwitchPortImportSource specifies where to import the SwitchPort data. options: "netbox", "bmh" and "none"
  switchPortImportSource: bmh
  # StatusReportPeriodInSec specifies the interval of getting switch config and updating the status of Switch and SwitchPort CRs.
	statusReportPeriodInSec int
  
  /* Netbox related */
  # NetboxServer
  netboxServer: "https://netbox.idcs-enrollment.svc.cluster.local"
  # NetboxTokenPath
  netboxTokenPath: "/vault/secrets/netboxtoken"
  # NetboxProviderServersFilterFilePath - the filters that define which switches the Netbox Controller maintains
  netboxProviderServersFilterFilePath: /netbox/netbox_provider_servers_filter.json
  # note: filtering the interface also require the netboxProviderServersFilterFilePath value, as we need to first find the devices(switches), and then find the interfaces.
  # NetboxProviderInterfacesFilterFilePath - the filters that define which switch ports/interfaces the Netbox Controller maintains
  netboxProviderInterfacesFilterFilePath: /netbox/netbox_provider_interfaces_filter.json
  # NetboxSwitchesFilterFilePath - the filters that define which switches the Netbox Controller maintains
  netboxSwitchesFilterFilePath: /netbox/switches_filter.json
  # NetboxClientInsecureSkipVerify specifies the InsecureSkipVerify setting
  netboxClientInsecureSkipVerify: false
  
ravenConfig:
  # environment specifies the Raven environment, options: mock/rnd/staging/prod. Where `mock` is using mock Raven, `rnd` is for development, `staging` is for staging environment, and `prod` is for production.
  environment: mock
  # credentialsFilePath specifies the location of Raven credential file. 
  credentialsFilePath: tests/config/raven_credentials.yaml
  # host specifies the Raven API server host name. 
  host: raven-devcloud.app.intel.com
```

By default, TLS is enabled and will be used for connections to Raven, eAPI, and Vault.

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

The manifests will be generated to the folder `frameworks.cloud.devcloud.services.idc/deployment/charts/sdn-controller/templates`

# Valid values for CRDs & fields

| CRD field                            | Meaning of -1                         | Meaning of "missing field"                    | Meaning of 0 / "" / [] (depending on field type)                        | Value to set to remove from switch config |
|--------------------------------------|---------------------------------------|-----------------------------------------------|-------------------------------------------------------------------------|-------------------------------------------|
| NetworkNode.*Fabric.Spec.VlanId      | Will not make change to switchport CR | Will default to -1 (& not modify SP CR)       | Will not make change to switchport CR                                   | 1                                         |
| NetworkNode.*Fabric.Spec.Mode        | N/A (string field, can't be -1)       | Will not make change to switchport CR         | Will not make change to switchport CR                                   | "access"                                  |
| NetworkNode.*Fabric.Spec.TrunkGroups | N/A (list field)                      | Will remove switchport.Spec.TrunkGroups field | Set TrunkGroups in child switchports to [] - ie. remove all trunkGroups | []                                        |
| NetworkNode.*Fabric.Spec.NativeVlan  | Will not make change to switchport CR | Will default to -1 (& not modify SP CR)       | Will not make change to switchport CR                                   | 1                                         |
| Switchport.Spec.Description          | N/A (string field, can't be -1)       | Will not make change on switch                | Will remove description (set it to "")                                  | ""                                        |
| Switchport.Spec.VlanId               | Will not make change on switch        | Will not make change on switch                | 0 is ignored (no change made to switch)                                 | 1                                         |
| Switchport.Spec.Mode                 | N/A (string field, can't be -1)       | Will fail validation. Set "" for "no action"  | "" will not make change on switch                                       | "access"                                  | 
| Switchport.Spec.TrunkGroups          | N/A (list field)                      | Will not make change on switch                | Set TrunkGroups on switch to [] - ie. remove all trunkGroups            | []                                        |
| Switchport.Spec.NativeVlan           | Will not make change on switch        | Will not make change on switch                | 0 is ignored (no change made to switch)                                 | 1                                         |
| Switch.Spec.BGPCommunity             | Will not make change on switch        | Will not make change on switch                | 0 is ignored (no change made to switch)                                 | Not supported                             |

### Deploying into an existing IDC deployment

Check out the `TWC4721-73/Implement-Network-Controller` branch (or main, once it's merged)

Populate local/secrets with existing secrets if there are any & run:
```
make secrets
```
Edit these files to add the correct passwords (DO NOT ADD A NEWLINE AT THE END (in vi, use ":set binary", ":set noeol", then save)):
* `local/secrets/RAVEN_USERNAME`
* `local/secrets/RAVEN_PASSWD`
* `local/secrets/EAPI_USERNAME`
* `local/secrets/EAPI_PASSWD`

Manually create your new network cluster & deploy coredns through whatever process you create kubernetes clusters
(devops automation ansible scripts, or using Kind).
If using kind during development, there's a script that will deploy the kind cluster & add it to the kubeconfig contexts:
```
cd ~/idc
export IDC_ENV=kind-multicluster
make deploy-network-cluster-in-kind
make deploy-coredns-to-network-cluster
```

Put the kubeconfig file for accessing this cluster at the location set in `build/environments/${IDC_ENV}/Makefile.environment`
(eg. for kind, put it at `local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml`)

If it's a Rancher-managed cluster, set it to use the context that connects DIRECTLY to a node in the cluster 
(not via Rancher control-plane in the cloud):
```
kubectl --kubeconfig $(pwd)/local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml config get-contexts
kubectl --kubeconfig $(pwd)/local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml config use-context <eg. pdx05-k01-nwcp-cluster-pdx05-c01-nwcp001>
```

Set `$KUBECONFIG` so that you deploy to the networking k8s cluster:
```
export KUBECONFIG=$(pwd)/local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml
kubectl config get-contexts
```

Deploy the "restricted access service accounts" to the network cluster:
```
make deploy-restricted-serviceaccounts
```

Either manually create & put the restricted kubeconfig file for accessing the Network CP at `local/secrets/restricted-kubeconfig/sdn-bmaas-kubeconfig.yaml`
OR generate a new "restricted access" kubeconfig file at this location based on the existing "full" kubeconfig file:
```
make generate-sdn-restricted-kubeconfigs
cat local/secrets/restricted-kubeconfig/sdn-bmaas-kubeconfig.yaml 
```

Set the vault env vars to the Vault instance for your environment (eg. in Kind, set):
```
export VAULT_ADDR=http://localhost:30990/
export VAULT_TOKEN=$(cat local/secrets/VAULT_TOKEN)
```

Run the `go/pkg/sdn-controller/vault-configure-for-nw-cluster.sh` script. This will ADD the vault roles needed for the network cluster to an existing IDC vault deployment.
Run `make deploy-vault-secrets-for-nw-cluster` to ADD the passwords needed for the network cluster:
These 2 scripts should not modify existing values in vault used by other clusters.

If you are using a new EXTERNAL networking cluster (eg. prod, staging, you're using kind-multicluster or have a
separately provisioned nwcp cluster), add the JWK for this cluster to Vault's `auth/cluster-auth/config` to allow the 
cluster agent to log in to Vault to retrieve secrets etc:
```
# Grab the network-cluster public key from kubernetes
export NETWORKING_KUBECONFIG=$(pwd)/local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml
export SECRETS_DIR=$(pwd)/local/secrets
# If this fails, try to grab it from one of the nodes directly (not using rancher) by changing your current kube context.
kubectl get --kubeconfig ${NETWORKING_KUBECONFIG} --raw \
"$(kubectl get --kubeconfig ${NETWORKING_KUBECONFIG} --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')" > ${SECRETS_DIR}/vault-jwk-validation-public-keys/network.jwk

# Regenerate the vault-jwt-validation-public-keys-plus-nw.json 
./deployment/common/vault/jwk-to-vault.py --input-dir ${SECRETS_DIR}/vault-jwk-validation-public-keys --output-file ${SECRETS_DIR}/vault-jwt-validation-public-keys-plus-nw.json

# View:
cat ${SECRETS_DIR}/vault-jwt-validation-public-keys-plus-nw.json

# It should have MULTIPLE JWTs in there - one for the network, others for each cluster. Check & compare with the EXISTING Vault config to make sure nothing has been removed:
vault read --format=json --field=jwt_validation_pubkeys auth/cluster-auth/config

# Save to vault
cat ${SECRETS_DIR}/vault-jwt-validation-public-keys-plus-nw.json | vault write auth/cluster-auth/config -
```

Select the network cluster & deploy coredns, vault-agent & the SDN controller to it:
```
cd ~/idc
export KUBECONFIG=$(pwd)/local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml
make deploy-network-services
```

Check that the sdn-controller pod is running & check logs (possibly of the vault-init container) if not:
```
k9s
```

#### Rollback steps

```
make undeploy-network-services
make undeploy-sdn-restricted-sa
```

Check deleted:
```
kubectl get pods -A | grep sdn
kubectl get switch -A
kubectl get secrets -A | grep sdn
```

Redeploy main:
```
git checkout main
unset KUBECONFIG DOCKER_TAG
make deploy-idc
```

You can clean up Vault if desired:
```
export VAULT_ADDR=http://localhost:30990/
export VAULT_TOKEN=$(cat local/secrets/VAULT_TOKEN)

vault secrets disable us-dev-1a-network-ca/
vault kv delete controlplane/us-dev-1/us-dev-1a/nw-sdn-controller/raven
vault kv delete controlplane/us-dev-1/us-dev-1a/nw-sdn-controller/eapi
vault kv delete controlplane/us-dev-1/us-dev-1a/nw-sdn-controller/bmhkubeconfig
vault kv delete controlplane/us-dev-1a-bm-instance-operator/sdnkubeconfig
vault policy delete us-dev-1a-nw-sdn-controller-policy
vault delete /auth/cluster-auth/role/us-dev-1a-nw-sdn-controller-role
```

# Handling `TrunkGroups` in the reconciliation. 

The `TrunkGroups` field (a slice of string `*[]string`) defined in the NetworkNode and SwitchPort CRD's Spec, represents the `Static Trunk Groups` of an interface on a Arista switch. 
```
Name: Et5
Switchport: Enabled
Administrative Mode: trunk
Operational Mode: trunk
MAC Address Learning: enabled
Dot1q ethertype/TPID: 0x8100 (active)
Dot1q VLAN Tag: Allowed
Access Mode VLAN: 1 (default)
Trunking Native Mode VLAN: 1 (default)
Administrative Native VLAN tagging: disabled
Trunking VLANs Enabled: ALL
Static Trunk Groups: Tenant_Nets
Dynamic Trunk Groups:
...
```

## handling `TrunkGroups` in SwitchPort(SP) Controller

When `SwitchPort.Spec.TrunkGroups` field is nil, SP Controller won't make any change to the trunk groups of an interface

When `SwitchPort.Spec.TrunkGroups` field is NOT nil, but is empty (ie, []string{}), SP Controller will remove all the trunk groups for an interface

When `SwitchPort.Spec.TrunkGroups` field is NOT empty (ie, []string{"Tenant_Nets"}), SP Controller will reconcile to ensure the trunk groups of an interface meet the desired value. 

## handling `TrunkGroups` in NetworkNode(NN) Controller

NN Controller's responsibility is to ensure the `NetworkNode.Spec.TrunkGroups` value is in-sync with SwitchPort CRs it managed. For example, when `NetworkNode.Spec.TrunkGroups` is nil, then `SwitchPort.Spec.TrunkGroups` also need to be nil. 

Beside, it needs to report the `NetworkNode.Status.readiness` field accurately. Especially when comparing the `nil` and `empty` values between the NN and SP.

For example, if `NetworkNode.Spec.TrunkGroups` is `nil`, and `SwitchPort.Spec.TrunkGroups` is `[]string{}`, then they are NOT in-sync.
