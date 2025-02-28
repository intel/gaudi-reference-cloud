<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# VM cluster operator

# Development environment

The development environment is built upon the baremetal virtual stack ([BMVS](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/bmaas/bmvs)).

## Prerequisites

It is not necessary but recommended to set the GUEST_HOST_CPUSET environment variable.
This improves performance by ensuring all VM CPUs share the same L3 cache.

    $ lscpu -e
    CPU NODE SOCKET CORE L1d:L1i:L2:L3 ONLINE    MAXMHZ   MINMHZ      MHZ
      0    0      0    0 0:0:0:0          yes 3200.0000 800.0000  894.824
    ...
     31    0      0   31 31:31:31:0       yes 3200.0000 800.0000  800.000
     32    1      1   32 64:64:64:1       yes 3200.0000 800.0000 1010.021
    ...
     63    1      1   63 95:95:95:1       yes 3200.0000 800.0000  800.000
     64    0      0    0 0:0:0:0          yes 3200.0000 800.0000  800.000
    ...
     95    0      0   31 31:31:31:0       yes 3200.0000 800.0000  800.000
     96    1      1   32 64:64:64:1       yes 3200.0000 800.0000  800.000
    ...
    127    1      1   63 95:95:95:1       yes 3200.0000 800.0000  800.000

For the above example, good choices are either GUEST_HOST_CPUSET=0-31,64-95 or GUEST_HOST_CPUSET=32-63,96-127.

Create a cluster independent of BMaaS to emulate a KubeVirt control plane created using existing DevOps tooling.

    make -C idcs_domain/bmaas/bmvs install-requirements
    sudo usermod -aG libvirt $USER
    go/pkg/vm_cluster_operator/test-scripts/controlplane.sh start-rancher
    go/pkg/vm_cluster_operator/test-scripts/controlplane.sh create-cluster

Create a deployment environment derived from kind-jenkins.
Use $(hostname) instead of the example hostname internal-placeholder.com shown below.

    $ git diff deployment/helmfile/environments.yaml
    diff --git a/deployment/helmfile/environments.yaml b/deployment/helmfile/environments.yaml
    index 1f0f23339b..a49ca3b33a 100644
    --- a/deployment/helmfile/environments.yaml
    +++ b/deployment/helmfile/environments.yaml
    @@ -6,6 +6,11 @@
    # - shared/dev3-harvester-alias-dev1.yaml.gotmpl when region in us-dev-1

    environments:
    +  internal-placeholder.com:
    +    values:
    +      - defaults.yaml.gotmpl
    +      - environments/kind-jenkins.yaml.gotmpl
    +      - environments/internal-placeholder.com.yaml.gotmpl
      dev1:
        values:
          - defaults.yaml.gotmpl

Create deployment/helmfile/environments/$(hostname).yaml.gotmpl with the correct value for the dhcpFileServer IP address.
The correct address is that of Kind API server.
The in-use addresses can be found by inspecting the docker kind network:

    $ docker network inspect kind | grep IPv4Address
                    "IPv4Address": "172.18.0.2/16",
                    "IPv4Address": "172.18.0.3/16",

The next address will be 172.18.0.4, which is the what the Kind API server will use.

    $ cat deployment/helmfile/environments/internal-placeholder.com.yaml.gotmpl
    regions:
      us-dev-1:
        computeVmMachineImages:
          # Only deploy a single base image to conserve host disk space
          excludeSources:
            - ".*gaudi.*"
            - ".*pvc.*"
            - ubuntu-2204-jammy-v20240308
          urlPrefix: http://10.45.122.149/vmaas
        availabilityZones:
          us-dev-1a:
            computeVmMachineImages:
              enabled: true
            harvesterClusters: []
            kubeVirtClusters:
              - clusterId: sea01-k01-azkv
        dhcpProxy:
          dhcpProxyConfig:
            dhcpFileserver: 172.18.0.4

Patch the scheduler to ignore missing Harvester CRDs (this will be fixed properly with https://internal-placeholder.com/browse/IDCCOMP-3424).

    diff --git a/go/pkg/instance_scheduler/vm/server/scheduling_service.go b/go/pkg/instance_scheduler/vm/server/scheduling_service.go
    index 125192245e..ed92445be1 100644
    --- a/go/pkg/instance_scheduler/vm/server/scheduling_service.go
    +++ b/go/pkg/instance_scheduler/vm/server/scheduling_service.go
    @@ -97,7 +97,7 @@ func NewSchedulingService(ctx context.Context, cfg *privatecloudv1alpha1.VmInsta
                            return nil, fmt.Errorf("unable to create clientset for %s: %w", kubeConfigFilename, err)
                    }
                    // Patch Harvester overcommit config settings
    -               err = patchHarvesterOvercommitConfig(ctx, kubeConfig, cfg)
    +               err = nil // TODO patchHarvesterOvercommitConfig(ctx, kubeConfig, cfg)
                    if err != nil {
                            return nil, err
                    }

## Deploy Kind

Follow instructions to [deploy a new kind cluster, and deploy baremetal operator, baremetal virtual stack](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/docs/source/private/guides/local_development_guide.rst#deploy-a-new-kind-cluster-and-deploy-baremetal-operator-baremetal-virtual-stack), with the following changes:

- Update the host public key secret.
  This only needs to be done once.

    ssh-keyscan -t rsa ${SSH_PROXY_IP} | awk '{print $2, $3}'> local/secrets/ssh-proxy-operator/host_public_key

- Edit the sshd config to add guest-${USER} and bmo-${USER} to AllowUsers.
  This only needs to be done once.

    echo "AllowUsers guest-${USER}" | sudo tee /etc/ssh/sshd_config.d/guest-${USER}.conf
    echo "AllowUsers bmo-${USER}" | sudo tee /etc/ssh/sshd_config.d/bmo-${USER}.conf
    sudo systemctl restart sshd

- Increase the CPU and memory of guests to accommodate nested KubeVirt VMs.

    export GUEST_HOST_DEPLOYMENTS=1
    export GUEST_HOST_VCPU=8 GUEST_HOST_MEMORY_MB=32768

## Create management cloud account and VNet

Management here refers to management of KubeVirt worker nodes (in contrast to tenant accounts and VNets).

    export no_proxy=${no_proxy},.kind.local
    export URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local

    export TOKEN=$(curl "${URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
    echo ${TOKEN}

    export URL_PREFIX=https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local
    export CLOUDACCOUNTNAME=${USER}@intel.com
    go/svc/cloudaccount/test-scripts/cloud_account_create.sh
    export CLOUDACCOUNT=$(go/svc/cloudaccount/test-scripts/cloud_account_get_by_name.sh | jq -r .id)
    echo $CLOUDACCOUNT

    go/svc/compute_api_server/test-scripts/sshpublickey_create_with_name.sh

    export AZONE=us-dev-1b
    export VNETNAME=us-dev-1b-mgmt
    go/pkg/vm_cluster_operator/test-scripts/mgmt_vnet_create_with_name.sh

## Enroll the BMVS virtual machine

- Side note: speed up things by disabling bm-validation.

    kubectl -n idcs-system scale --replicas 0 deployment/us-dev-1a-bm-validation-operator

- Port forward 30980 if using VS Code.

- Open Netbox at http://localhost:30980.
  The username:password is admin:admin.
  Go to Devices, select device-1, Edit and set Device role to bmaas and BM Enrollment Status to Enroll.

- Wait for BareMetalHost resources in the IDC cluster.
  It takes approximately 15m on my VM for the machines to be ready.

- Side note: label the node as verified if bm-validation was disabled earlier.
  This will also need to be done after deleting an Instance.

    kubectl -n metal3-1 label bmh device-1 cloud.intel.com/verified=true
    kubectl -n metal3-1 label bmh device-1 instance-type.cloud.intel.com/bm-virtual=true

Debugging tips:

- To see early stage netboot: virt-manager --connect qemu:///system --show-domain-console gh-node-1.
  If this fails, confirm that the dhcpFileServer address from above is correct.

- If validation fails, confirm that the host ssh config is correct (i.e. AllowUsers mentioned above).

## Create a worker node Instance

   export NAME=worker-instance-1
   export INSTANCE_TYPE=bm-virtual
   export MACHINE_IMAGE=ubuntu-22.04-server-cloudimg-amd64-latest
   go/pkg/vm_cluster_operator/test-scripts/instance_create_with_name.sh

## Delete a worker node Instance

To ensure proper cleanup of the Rancher cluster before deleting the Instance:

    go/pkg/vm_cluster_operator/test-scripts/instance_delete_by_name.sh

## Create a KubeVirt virtual machine

Create a test KubeVirt VirtualMachine.
Note that the VLAN and and networkData in test-vm.yaml may need to be modified depending on which subnet was chosen for the management network.

    kubectl --context sea01-k01-azkv apply -f go/pkg/vm_cluster_operator/test-scripts/test-vm.yaml

When ready, the VM can be accessed with the virtctl console:

    virtctl --context sea01-k01-azkv console test-vm -n test-vm
