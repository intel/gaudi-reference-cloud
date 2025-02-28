.. _fleet_management:

Fleet Management
################

.. _view_compute_node_pools:

View Compute Node Pools
***********************

The list of compute node pools is documented in the following pages:

* `Staging Compute Node Pools`_
* `Production Compute Node Pools`_

.. _create_compute_node_pool:

Create a Compute Node Pool
**************************

#. Edit the applicable document in :ref:`view_compute_node_pools` and add the new compute node pool.

Note: There is not an API to add a compute node pool. A compute node pool will be implicitly created when configuring pool labels
or associating a pool with a Cloud Account.

.. _configure_pool_labels_for_baremetalhosts:

Configure Pool Labels for BareMetalHosts
****************************************

#. Set KUBECONFIG for the AZCP (availability zone control plane) cluster that has the Metal3 BareMetalHost custom resources.

   .. code-block:: bash

      export KUBECONFIG=pdx05-k02-azcp.yaml

#. Run the following command to view each BareMetalHost, along with the compute node pools that it is assigned to.

   .. code-block:: bash

      kubectl get -A BareMetalHost -o yaml | \
      yq eval '[.items[] | {"name": .metadata.name, "namespace": .metadata.namespace, "labels": (.metadata.labels | with_entries(select(.key | test("pool\.cloud\.intel\.com/|instance-type\.cloud\.intel\.com/"))))}]'

   Example output:
   
   .. code-block:: yaml

      - name: pdx05-c01-bmas017
        namespace: metal3-1
        labels:
          instance-type.cloud.intel.com/bm-srf-sp-quanta: "true"
      - name: pdx05-c01-bmas003
        namespace: metal3-2
        labels:
          instance-type.cloud.intel.com/bm-spr-pvc-1100-8: "true"
          pool.cloud.intel.com/general: "true"
      - name: pdx03-c01-dspr001
        namespace: metal3-2
        labels:
          instance-type.cloud.intel.com/bm-spr: "true"
          pool.cloud.intel.com/general: "true"
          pool.cloud.intel.com/pool2: "true"

#. To add all BareMetalHosts to the general pool:

   .. code-block:: bash

      kubectl label --overwrite --all-namespaces --all BareMetalHosts pool.cloud.intel.com/general=true

#. To add a BareMetalHost to a pool:

   .. code-block:: bash

      kubectl label --overwrite -n metal3-1 BareMetalHosts/pdx05-c01-bspr006 pool.cloud.intel.com/general=true

#. To remove a BareMetalHost from a pool:

   .. code-block:: bash

      kubectl label --overwrite -n metal3-1 BareMetalHosts/pdx05-c01-bspr006 pool.cloud.intel.com/general-

.. _configure_pool_labels_for_harvester_nodes:

Configure Pool Labels for Harvester nodes
*****************************************

#. :ref:`run_kubectl_against_harvester`

#. Run the following command to view each node, along with the compute node pools that it is assigned to.

   .. code-block:: bash

      kubectl get nodes -o yaml | \
      yq eval '[.items[] | {"name": .metadata.name, "labels": (.metadata.labels | with_entries(select(.key | test("pool\.cloud\.intel\.com/|instance-type\.cloud\.intel\.com/"))))}]'

   Example output:
   
   .. code-block:: yaml

      - name: harvester-cp-dev1
        labels:
          instance-type.cloud.intel.com/vm-spr-med: "true"
          instance-type.cloud.intel.com/vm-spr-sml: "true"        
          pool.cloud.intel.com/general: "true"

#. To add all nodes to the general pool:

   .. code-block:: bash

      kubectl label --overwrite nodes -l node-role.kubernetes.io/master!=true pool.cloud.intel.com/general=true

#. To add a node to a pool:

   .. code-block:: bash

      kubectl label --overwrite nodes/harvester-cp-dev1 pool.cloud.intel.com/general=true

#. To remove a node from a pool:

   .. code-block:: bash

      kubectl label --overwrite nodes/harvester-cp-dev1 pool.cloud.intel.com/general-

.. _assign_cloud_account_to_compute_node_pools:

Assign Cloud Account to Compute Node Pools
******************************************

Follow the steps in this section to assign a Cloud Account to one or more Compute Node Pools.

Where to run this procedure
===========================

Run this from a workstation in the Intel corporate network.

Set environment variables for the environment
=============================================

.. code-block:: bash

   export IDC_ENV=staging
   export REGION=us-staging-1
   make show-config
   eval `make show-export`

Since the API servers are outside of the Intel corporate network (in Flex
or AWS), you will need to change your proxy configuration to force
requests to \*.intel.com to use the proxy.

.. code-block:: bash

   export no_proxy=10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16:10.165.28.33
   export NO_PROXY=${no_proxy}

Get IDC Admin API Token
========================

To obtain this token:

#. Login to the IDC Admin Console (https://admin.staging.console.idcservice.net/).

#. Click Admin Token in the Developer Tools box.

#. Copy the token to the clipboard.

#. Set the TOKEN environment variable.

   .. code-block:: bash

      export TOKEN="eyJhbGciOiJSUzI1NiIsI...7609g"

List the Compute Node Pools for the Cloud Account
=================================================

.. code-block:: bash

   CLOUDACCOUNT=... \
   go/pkg/fleet_admin/api_server/test-scripts/compute_node_pools_for_cloud_account_list.sh

Expected output:

.. code-block:: console

   {
      "computeNodePools": [
         {
            "poolId": "general"
         }
      ]
   }

List the Cloud Accounts for a Compute Node Pool
===============================================

.. code-block:: bash

   COMPUTE_NODE_POOL_ID=pool1 \
   go/pkg/fleet_admin/api_server/test-scripts/compute_node_pools_list_cloud_accounts_with_pool_id.sh

Expected output:

.. code-block:: console

   {
   "CloudAccountsForComputeNodePool": [
      {
         "cloudAccountId": "184547187159",
         "poolId": "pool1",
         "createAdmin": "idcadmin@intel.com"
      },
      {
         "cloudAccountId": "707654584302",
         "poolId": "pool1",
         "createAdmin": "idcadmin@intel.com"
      }
   ]
   }

Add Cloud Account to a Compute Node Pool
========================================

.. code-block:: bash

   CLOUDACCOUNT=... \
   CREATEADMIN=${USER}@intel.com \
   COMPUTE_NODE_POOL_ID=general \
   go/pkg/fleet_admin/api_server/test-scripts/compute_node_pools_add_cloud_account.sh

Delete Cloud Account from a Compute Node Pool
=============================================

.. code-block:: bash

   CLOUDACCOUNT=... \
   COMPUTE_NODE_POOL_ID=general \
   go/pkg/fleet_admin/api_server/test-scripts/compute_node_pools_delete_cloud_account.sh

Reserve BMaaS Capacity for a Single Cloud Account
=================================================

#. Define a compute node pool dedicated for a single Cloud Account.
   By convention, use the pool ID of "acct-" followed by the 12-digit Cloud Account ID. For example, "acct-123456789012".

   .. code-block:: bash

     export CLOUDACCOUNT=123456789012
     export COMPUTE_NODE_POOL_ID=acct-${CLOUDACCOUNT}

#. :ref:`create_compute_node_pool`.

#. Remove the node from the general pool.

   .. code-block:: bash

     kubectl label --overwrite -n metal3-1 BareMetalHosts/pdx05-c01-bspr006 pool.cloud.intel.com/general-

   Refer to :ref:`configure_pool_labels_for_baremetalhosts` for additional details.

#. Add the node to the new pool.

   .. code-block:: bash

      kubectl label --overwrite -n metal3-1 BareMetalHosts/pdx05-c01-bspr006 pool.cloud.intel.com/${COMPUTE_NODE_POOL_ID}=true

#. Add the Cloud Account to the new pool. See :ref:`assign_cloud_account_to_compute_node_pools`.



.. _Staging Compute Node Pools: https://internal-placeholder.com/display/devcloud/IDC+Staging+Environment#IDCStagingEnvironment-ComputeNodePools
.. _Production Compute Node Pools: https://internal-placeholder.com/display/devcloud/IDC+Production+Environment#IDCProductionEnvironment-ComputeNodePools
