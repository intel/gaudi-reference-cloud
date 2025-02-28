.. _harvester_deployment_procedure:

Harvester Deployment Procedure
##############################

This document describes how to install a new Harvester cluster and configure IDC to utilize it for VMaaS
in *production* and *staging* environments.

Related Documentation
*********************

* To install IDC services in a new production or staging environment, see :ref:`services_deployment_procedure`.

Assign Harvester Cluster ID
***************************

Assign a Harvester cluster ID to the new Harvester cluster.
Use a Harvester cluster ID following the example ``pdx05-k03-hv``.
Do not exceed 12 characters.

Install Harvester
*****************

Deploy Harvester Control Plane
==============================

#. Follow steps in `Harvester ISO Installation`_.

#. Change settings in the Harvester UI.

   a. Advanced -> Settings: overcommit-config: {"cpu": 200, "memory": 100, "storage": 200}

Add Harvester Nodes
===================

#. Configure the switch port to trunk all VLAN IDs corresponding to tenant VLANs,
   whether they are allocated to VNets or not.
   The VLAN for the Harvester node’s operating system should be untagged (native).

#. Follow steps in `Harvester ISO Installation`_.
   Be sure to use the option to join an existing cluster.

To open the Harvester UI in a browser
=====================================

.. code-block:: bash

   export CONTROL_PLANE_VAULT_OTP_PATH=ssh-harvester-services
   HARVESTER_VIP=100.64.20.35
   LOCAL_PORT=16503
   SSH_OPTS="-L ${LOCAL_PORT}:localhost:443" deployment/scripts/ssh-with-2otp.sh rancher@${HARVESTER_VIP}

If using VS Code to a remote SSH server, configure a port forward from 8250 to localhost:8250. This will be used for OIDC.

If using VS Code to a remote SSH server, configure a port forward from port ${LOCAL_PORT} to localhost:${LOCAL_PORT}.

Keep ssh running and open your browser to https://localhost:${LOCAL_PORT}/.
Login as admin.

Note: If above gives a port forward error, ensure that "AllowTcpForwarding yes" is in /etc/ssh/sshd_config in the Harvester node,
then run "systemctl restart sshd".

Update Wiki with Harvester Details
**********************************

Update the Wiki with the Harvester cluster details including the sections "Region Overview" and "How to connect to Harvester UI".
For example, see `IDC PDX04 us-region-1 Production Region`_.

Enable PCI Passthrough
======================

#. Login to the Harvester UI.

#. Click Addons -> pcidevices-controller. Check enabled and save it.

#. Click PCI Devices. Search for "gpu". Check all matching devices and click Enable Passthrough.

#. Repeat for "gaudi2" devices.

Obtain Harvester KubeConfigs
============================

Obtain the Harvester KubeConfig file using the steps below. Repeat for
each Harvester cluster.

#. Login to the Harvester UI.

#. Click *Support* in the bottom-left corner.

#. Click *Download KubeConfig*.

#. Save the file to
   ``local/secrets/${IDC_ENV}/harvester-kubeconfig/${HARVESTER_CLUSTER_ID}``.

#. In the Kubeconfig file, ensure the field ``clusters.cluster.server``
   has the Harvester VIP or FQDN that the VM Instance Operator in the AZ
   cluster can reach the Harvester cluster with. This must not be
   localhost or 127.0.0.1.

.. _run_kubectl_against_harvester:

Run Kubectl against Harvester
=============================

#. `SSH into a node in the Harvester cluster <#to-open-the-harvester-ui-in-a-browser>`__.

#. Sudo as root and test kubectl.

   .. code-block:: bash

      rancher@pdx04-c01-azvm001:~>
      sudo -i

      pdx04-c01-azvm001:~ #
      kubectl get nodes

Load Virtual Machine Images into Harvester
==========================================

Harvester uses VirtualMachineImage custom resources to copy machine image qcow files from an S3 bucket to the Harvester cluster.

#. Ensure that the machine image qcow file has been uploaded to the `Machine Image S3 Bucket`_.
   This should be done automatically by `Software Inflow`_.   

#. Create the VirtualMachineImage YAML file in ``build/environments/prod/us-region-1/VirtualMachineImage``.
   The host IP address in the url field should point to the regional NGINX caching proxy server.
   This server proxies requests to the above S3 bucket.

   The VirtualMachineImage name must match the MachineImage name and must be fewer than 32 characters.

#. Create a tar file containing all required VirtualMachineImage YAML files.

   .. code-block:: bash
      
      tar -C build/environments/staging/us-staging-1/VirtualMachineImage -cz . | base64

#. `Run Kubectl against Harvester <#run-kubectl-against-harvester>`__.

#. Extract VirtualMachineImage YAML files onto the Harvester node.
   You should copy/paste the base64 output from the prior step.

   .. code-block:: bash
      
      mkdir -p VirtualMachineImage
      base64 -d | tar -C VirtualMachineImage -xvz

#. Apply VirtualMachineImage Kubernetes resource.

   .. code-block:: bash

      kubectl apply -f VirtualMachineImage

#.	Confirm that the VirtualMachineImage was imported.

   .. code-block:: bash

      kubectl get VirtualMachineImage -o yaml

   "status.conditions[type=Imported].status" should be True.

Install Intel Device Plugin
***************************

A new device plugin called intel-device-plugin, has been created which will be utilized to advertise hardware resources, 
such as Habana gaudi2 accelerators PCIe cards, to the kubelet.
This allows these PCI devices to be made available for running AI-related workloads on a dedicated virtual machine.
In Addition, the device plugin performs a Function Level Reset (FLR) so as to clear Gaudi2 memory before VM instance creation

Pre-requisite
=============

- Access to *us-staging-1* gaudi2 harvester cluster **pdx05-k02-hv**
- On the harvester cluster make sure the *kubevirt* Custom Resource under the namespace *harvester-system* has ``externalResourceProvider: true`` 
  for the pciHostDevice for which device plugin is being deployed

  .. code-block:: bash

   ................
   ................
   spec:
    certificateRotateStrategy: {}
    configuration:
      developerConfiguration:
        featureGates:
        - LiveMigration
        - HotplugVolumes
        - HostDevices
      emulatedMachines:
      - q35
      - pc-q35*
      - pc
      - pc-i440fx*
      network:
        defaultNetworkInterface: bridge
        permitBridgeInterfaceOnPodNetwork: true
        permitSlirpInterface: true
      permittedHostDevices:
        pciHostDevices:
        - externalResourceProvider: true
          pciVendorSelector: 1da3:1020
          resourceName: habana.com/GAUDI2_AI_TRAINING_ACCELERATOR
    customizeComponents:
      patches:
    ................
    ................

Install Intel Device Plugin
===========================

- SSH into the gaudi2 harvester cluster *pdx05-k02-hv* from your workstation refer
  https://internal-placeholder.com/display/devcloud/IDC+PDX05+us-staging-1+Staging+Region#IDCPDX05usstaging1StagingRegion-ForHarvesterclusterpdx05-k02-hv

  .. code-block:: bash

     export VAULT_ADDR=https://internal-placeholder.com/
     export BASTION_IP=10.45.117.142
     export CONTROL_PLANE_VAULT_OTP_PATH=ssh-harvester-services
     HARVESTER_VIP=100.64.20.30
     LOCAL_PORT=16555
     SSH_OPTS="-L ${LOCAL_PORT}:localhost:443" deployment/scripts/ssh-with-2otp.sh rancher@${HARVESTER_VIP}
     
     rancher@pdx05-c01-bgan10:~>
     sudo -i

- Install Helm if not already present from the binary releases https://github.com/helm/helm/releases
   
  .. code-block:: bash

     curl -O https://get.helm.sh/helm-v3.14.2-linux-amd64.tar.gz
     tar -zxvf helm-v3.14.2-linux-amd64.tar.gz
     mv linux-amd64/helm /usr/local/bin/helm 
   
- Pull the latest helm chart from pre prod registry *amr-idc-registry-pre.infra-host.com*
  
  .. code-block:: console
     
     helm pull oci://amr-idc-registry-pre.infra-host.com/intelcloud/intel-device-plugin --version 0.0.1-b300d1cc01cae96950ba332b7e42789b1365d4b82e408dd2b1f2152e5b4b9b7a

  **NOTE**: 
  The latest helm charts and docker images can be downloaded from here. 

    - Pre-Prod: https://amr-idc-registry-pre.infra-host.com/harbor/projects/3/repositories/intel-device-plugin/artifacts-tab
    - caas-regsitry: https://internal-placeholder.com/harbor/projects/1963/repositories/intel-device-plugin/artifacts-tab

- Install the chart
  
  .. code-block:: console

     helm install intel-device-plugin intel-device-plugin-0.0.1-b300d1cc01cae96950ba332b7e42789b1365d4b82e408dd2b1f2152e5b4b9b7a.tgz --set image.registry=amr-idc-registry-pre.infra-host.com
   
  **NOTE**: *image.registry* is required so as to point from which registry (pre-prod or caas) container image needs to be pulled.

   
  Upon successful installation of the chart , a *daemonset* should be created under the namespace *kube-system* inside the harvester k8s cluster with the following container logs

  .. code-block:: console

     {"component":"","level":"info","msg":"Discovered 8 PCI devices on the node for the resource: habana.com/GAUDI2_AI_TRAINING_ACCELERATOR","pos":"main.go:26","timestamp":"2024-02-24T08:57:55.737395Z"}  
     {"component":"","level":"info","msg":"Start Device Plugin","pos":"pci_device.go:87","timestamp":"2024-02-24T08:57:55.737482Z"}  
     {"component":"","level":"info","msg":"Registering the device plugin","pos":"pci_device.go:296","timestamp":"2024-02-24T08:57:55.738646Z"}  
     {"component":"","level":"info","msg":"habana.com/GAUDI2_AI_TRAINING_ACCELERATOR device plugin started","pos":"pci_device.go:126","timestamp":"2024-02-24T08:57:55.741885Z"}

Uninstalling Intel Device Plugin
================================

To Uninstall the device plugin from the harvester cluster, run the following commad

``helm uninstall intel-device-plugin``

Configure IDC to use the Harvester Cluster
******************************************

Configure Vault
===============

#. Create a new branch in `IDC monorepo`_.

#. Add Vault roles by adding the following lines to the file `deployment/common/vault/terraform/data/staging/jwt_roles_region_1.tfvars`.
   Be sure to use the appropriate region and Harvester cluster ID.

   .. code-block::

      "us-staging-1a-vm-instance-operator-pdx05-k03-hv-role" = {
         "role_name"       = "us-staging-1a-vm-instance-operator-pdx05-k03-hv-role"
         "backend"         = "cluster-auth"
         "token_policies"  = ["global-pki", "public", "us-staging-1a-vm-instance-operator-pdx05-k03-hv-policy"]
         "token_ttl"       = 3600
         "bound_subject"   = "system:serviceaccount:idcs-system:us-staging-1a-vm-instance-operator-pdx05-k03-hv"
         "bound_audiences" = ["https://kubernetes.default.svc.cluster.local"]
      }

#. Add Vault PKI roles by adding the following lines to the file `deployment/common/vault/terraform/data/staging/pki_roles_region_1.tfvars`.
   Be sure to use the appropriate region and Harvester cluster ID.

   .. code-block::

      "us-staging-1a-vm-instance-operator-pdx05-k03-hv" = {
         "backend"                     = "us-staging-1a-ca"
         "role_name"                   = "us-staging-1a-vm-instance-operator-pdx05-k03-hv"
         "allowed_domains"             = ["us-staging-1a-vm-instance-operator-pdx05-k03-hv.idcs-system.svc.cluster.local", "*.local", "*.internal-placeholder.com", "*.eglb.intel.com", "*.internal-placeholder.com", "*.internal-placeholder.com"]
         "allow_glob_domains"          = true
         "allow_bare_domains"          = true
         "ou"                          = ["us-staging-1a-vm-instance-operator-pdx05-k03-hv"]
         "allow_wildcard_certificates" = false
         "enforce_hostnames"           = true
         "allow_any_name"              = false
         "key_bits"                    = 2048
      }

#. Add Vault policies by adding the following lines to the file `deployment/common/vault/terraform/data/staging/policies_region_1.tfvars`.
   Be sure to use the appropriate region and Harvester cluster ID.

   .. code-block::

      "us-staging-1a-vm-instance-operator-pdx05-k03-hv-policy" = {
         policy = <<EOT
      path "controlplane/data/us-staging-1a-vm-instance-operator-pdx05-k03-hv/*" {
         capabilities = ["read", "list"]
      }
      path "us-staging-1a-ca/issue/us-staging-1a-vm-instance-operator-pdx05-k03-hv" {
         capabilities = ["update"]
      }
      EOT
         }      

#. Create a PR for the above changes and merge it into main.

#. Deploy the Vault changes by following the `Vault Terraform procedure`_.

#. Add secrets to Vault.

   #. VM Instance Operator

      * Path: secrets/controlplane/us-staging-1a-vm-instance-operator-pdx05-k03-hv/harvester_kubeconfig
      * Key: kubeconfig
      * Value: (Harvester KubeConfig)

   #. VM Instance Scheduler

      * Path: secrets/controlplane/us-staging-1a-vm-instance-scheduler/harvester_kubeconfig_pdx05-k03-hv
      * Key: kubeconfig
      * Value: (Harvester KubeConfig)

Deploy Helm Releases Using Argo CD
==================================

#. Create a new branch in `IDC monorepo`_ from the commit used to deploy the following components in `Universe Config prod`_.

   * computeVmInstanceOperator
   * computeVmInstanceScheduler
   * computePopulateInstanceType
   * computePopulateMachineImage

#. Add the new Harvester cluster ID to `Helmfile environment prod.yaml.gotmpl`_ or `Helmfile environment staging.yaml.gotmpl`_
   in the appropriate section such as
   ``regions.us-region-1.availabilityZones.us-region-1a.harvesterClusters[].clusterId``.

#. If needed, you may include new instance types by copying files from 
   ``build/environments/staging/InstanceType`` to
   ``build/environments/prod/InstanceType``.

#. If needed, you may include new machine images by copying files from
   ``build/environments/staging/MachineImage`` to
   ``build/environments/prod/MachineImage``.

#. Commit your changes to the branch.

#. Update the commits for the components listed above in `Universe Config prod`_.

#. Create a PR for the above changes and get it approved.
   During an RFC implementation, merge the PR into main.
   Refer to the :ref:`services_upgrade_procedure` for details on how to deploy this IDC update.

#. If needed, :ref:`update-product-catalog-definitions`.

.. _configure_labels_on_each_harvester_worker_node:

Configure labels on each Harvester worker node
==============================================

Each Harvester worker node should have a set of labels indicating the
instance types that it supports and its partition.
These labels will enable the VM Instance Scheduler to schedule instances on the node.

#. `Run Kubectl against Harvester <#run-kubectl-against-harvester>`__.

#. Run the following to assign partitions labels.
   This will use the node name as the partition label.

   .. code-block:: bash

      kubectl get nodes -o jsonpath={.items..metadata.name} | xargs -d " " -i \
      kubectl label --overwrite nodes/{} cloud.intel.com/partition={}

#. Run commands similar to the following to assign instance type labels:

   .. code-block:: bash

      kubectl get nodes --show-labels
      kubectl label --overwrite nodes/pdx04-c01-bmas018 instance-type.cloud.intel.com/vm-spr-lrg=true
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true instance-type.cloud.intel.com/vm-spr-lrg=true
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true instance-type.cloud.intel.com/vm-spr-med=true
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true instance-type.cloud.intel.com/vm-spr-sml=true
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true instance-type.cloud.intel.com/vm-spr-tny=true
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master=true instance-type.cloud.intel.com/vm-spr-lrg-
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master=true instance-type.cloud.intel.com/vm-spr-med-
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master=true instance-type.cloud.intel.com/vm-spr-sml-
      kubectl label --overwrite nodes -l node-role.kubernetes.io/master=true instance-type.cloud.intel.com/vm-spr-tny-

#. Run commands similar to the following to assign compute node pool labels:

   .. code-block:: bash

      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true pool.cloud.intel.com/general=true

Disable Longhorn replica creation on control plane nodes
========================================================

For each Harvester control plane node, disable the creation of Longhorn replicas to
prevent the overutilization of storage space on these nodes.

#. `Run Kubectl against Harvester <#run-kubectl-against-harvester>`__.

#. Run the following to get all the control plane nodes.

   .. code-block:: bash

      CONTROL_PLANE_NODES=$(kubectl get nodes -l node-role.kubernetes.io/control-plane=true -o jsonpath='{.items[*].metadata.name}')

#. Run the following to disable the scheduling of Longhorn replicas on control plane nodes.

   .. code-block:: bash

      for CONTROL_PLANE_NODE in $CONTROL_PLANE_NODES; do kubectl patch nodes.longhorn.io $CONTROL_PLANE_NODE -n longhorn-system -p '{"spec":{"allowScheduling":false}}' --type=merge; done

#. Repeat the above steps for each SPR harvester cluster across all the staging and prod regions

End-to-End Test Procedure Using API
***********************************

This section documents how to perform an end-to-end test using the API.
Alternatively, you may test using the IDC Console https://console.cloud.intel.com/.

Where to run this procedure
===========================

Run this from a workstation in the Intel corporate network.

Set environment variables for the environment
=============================================

.. code-block:: bash

   export IDC_ENV=prod
   export REGION=us-region-2
   make show-config
   eval `make show-export`

Since the API servers are outside of the Intel corporate network (in Flex
or AWS), you will need to change your proxy configuration to force
requests to \*.intel.com to use the proxy.

.. code-block:: bash

   export no_proxy=10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16:10.165.28.33
   export NO_PROXY=${no_proxy}

Get IDC API Token
==================

Get a token from Azure AD
-------------------------

This is the production configuration. A token from Azure AD must be
obtained.

To obtain this token:

#. Login to the IDC console https://console.cloud.intel.com/ using Chrome.

#. Press F12 to open developer tools.

#. Open the Network tab.

#. Click on the Compute tab in the IDC console menu. This will force an
   API call.

#. In the Network tab, click on the "instances" request.

#. In the Headers tab, expand the Request Headers, and locate the
   Authorization header. This will have the form "Bearer
   eyJhbGciOiJSUzI1NiIsI...7609g". Copy the all of the text after the
   word "Bearer". This will be around 1319 characters. This is your
   token.

#. Set the TOKEN environment variable.

   .. code-block:: bash

      export TOKEN="eyJhbGciOiJSUzI1NiIsI...7609g"

Determine your Cloud Account
============================

Login to https://console.cloud.intel.com/ and obtain your Cloud Account ID.

.. code-block:: bash

   export CLOUDACCOUNT=...

Use Compute API to create an Instance
=====================================

.. code-block:: bash

   go/svc/compute_api_server/test-scripts/instance_list.sh
   export INSTANCE_TYPE=vm-spr-pvc-1100-1
   export MACHINE_IMAGE=ubuntu-22.04-pvc-v1100-vm-v2
   export KEYNAME=claudiof
   FIRST=2 LAST=3 go/svc/compute_api_server/test-scripts/instance_create_many.sh
   watch go/svc/compute_api_server/test-scripts/instance_list_summary.sh

SSH to Instance
===============

.. code-block:: bash

   ssh -J guest-${IDC_ENV}@10.165.62.252 ubuntu@172.16.x.x

Run a Workload
==============

See `VMaaS Demos`_.

Delete Instance
===============

.. code-block:: bash

   go/svc/compute_api_server/test-scripts/instance_delete_by_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh



.. _IDC monorepo: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc
.. _Harvester ISO Installation: https://docs.harvesterhci.io/v1.1/install/iso-install
.. _Universe Config prod: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/universe_deployer/environments/prod.json
.. _Universe Config staging: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/universe_deployer/environments/staging.json
.. _Helmfile environment prod.yaml.gotmpl: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/helmfile/environments/prod.yaml.gotmpl
.. _Helmfile environment staging.yaml.gotmpl: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/helmfile/environments/staging.yaml.gotmpl
.. _Vault Terraform procedure: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/common/vault/terraform/README.md
.. _IDC PDX04 us-region-1 Production Region: https://internal-placeholder.com/x/yZ0_xg
.. _Machine Image S3 Bucket: https://s3.console.aws.amazon.com/s3/buckets/catalog-fs-dev
.. _Software Inflow: https://internal-placeholder.com/x/4rVCtQ
.. _VMaaS Demos: https://internal-placeholder.com/x/ZJnbp
