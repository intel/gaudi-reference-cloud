.. _manage_k8s:

IKS and Usage
#############

|ITAC| has an Intel Kubernetes Service which supports creating and managing Kubernetes cluster.

Pre-Requisites:
*********************
* Create cloudaccount and SSH keys as requirements to create IKS cluster

CloudAccount:
*********************
* Create IDC User account, refer `User Account Guide`_.
* Login to IDC Portal and make sure you have a Cloud account associated with User account.
* CloudAccount is a 12 digit number.

Token:
*********************

* To get the bearer token:
* After signing in into `https://staging.console.idcservice.net/` , open the developer tool in your browser.
* Go to the Network tab. Refresh your browser tab once.
* You will notice an Requests URL and Headers tab appearing on the screen.
* Copy bearer token from the Headers/Request Headers/Authorization:Bearer tab and export the token as below.

#. 
   .. code-block:: bash

      export TOKEN=



SSH Keys
*********************
* Follow below document to create ssh keys for nodes.

For ssh key creation and upload , refer `SSH Key Guide`_. 


HTTP Methods
============

.. list-table::
   :header-rows: 1
  
   * - HTTP Method
     - Description
     
   * - GET
     - Retrieve available K8S versions supported by IKS.

   * - GET
     - Retrieve available Instance Types supported by IKS.
   
   * - GET
     - Retrieve available runtimes supported by IKS.
     
   * - POST
     - Create a a new IKS cluster.
    
   * - GET
     - Retrieve all available IKS clusters.

   * - GET
     - Retrieve an existing IKS cluster.

   * - GET
     - Retrieve an existing IKS cluster status.

   * - POST
     - Create a a new IKS nodegroup.

   * - GET
     - Retrieve an existing IKS nodegroup.

   * - PUT
     - Update an existing nodegroup.
   
   * - DELETE
     - Delete an existing IKS nodegroup.
   
   * - DELETE
     - Delete an existing IKS cluster.

Get Required Metadata to create IKS Kubernetes Cluster:
*******************************************************
* Follow below steps to Get Metadata required to create  IDC K8S cluster.

#. 
   .. code-block:: bash

        export URL_PREFIX=https://internal-placeholder.com
        export CLOUDACCOUNT=012345678912

#.
    .. code-block:: bash
      
        curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/metadata/runtimes -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

#.
    .. code-block:: bash
            
        curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/metadata/k8sversions -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"


Create Kubernetes Cluster:
**************************
* Follow below steps to Create Kubernetes cluster in IDC env once the required metadata is available.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912

   .. code-block:: bash

         curl -X POST \
         -H 'Content-type: application/json' \
         -H "Authorization: Bearer ${TOKEN}" \
         ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters -d \
         '{
            "name": "cls-iks",
            "description": "test-cls",
            "k8sversionname": "1.27",
            "runtimename": "Containerd",
            "tags": [
                  {
                     "key":"foo",
                     "value":"bbbbb"
                  }
            ]
          }'

Get All Kubernetes Clusters:
****************************
* Follow below steps to Get All Kubernetes Clusters in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912

   .. code-block:: bash

         curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

Get Specific Kubernetes Cluster:
********************************
* Follow below steps to Get Kubernetes Cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=

   .. code-block:: bash

         curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID} -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

Get kubernetes cluster status:
******************************
* Follow below steps to Get Kubernetes Cluster status in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=

   .. code-block:: bash

        curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/status -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

Delete Kubernetes Cluster:
**************************
* Follow below steps to Delete Kubernetes Cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=

   .. code-block:: bash

         curl -X DELETE ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID} -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"
         
Get Required Metadata to create Workers:
****************************************
* Follow below steps to Get Metadata required to create nodegroups.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912

   .. code-block:: bash
        
        curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/metadata/instancetypes -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

Create Vnets to create workers:
*******************************
* Follow below steps to Get Metadata required to create nodegroups.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912

   .. code-block:: bash
       
       curl -X POST \
       -H 'Content-type: application/json' \
       -H "Authorization: Bearer ${TOKEN}" \
       ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/vnets -d \
         '{
            "metadata": {
            "name": "test-vnet"
            },
            "spec": {
                  "availabilityZone": "us-region-1a",
                  "prefixLength": 1,
                  "region": "us-region-1"
            }
          }'

Create Nodegroup for Cluster:
*****************************
* Follow below steps to Create Nodegroup for available cluster in IDC env , use `SSH Keys`_ for creating ssh keys and vnet created from above step to create for nodegroup.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=

   .. code-block:: bash

         curl -X POST \
         -H 'Content-type: application/json' \
         -H "Authorization: Bearer ${TOKEN}" \
         ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups -d \
         '{
            "name": "ng-iks",
            "count": 2,
            "description": "test-nodegroup",
            "instancetypeid": "vm-spr-sml",
            "sshkeyname": [
                {
                "sshkey": "test-ssh-yogini"
                }
            ],
            "vnets": [
               {
                  "availabilityzonename": "us-staging-1a",
                  "networkinterfacevnetname": "us-staging-1a-default"
               }
            ]
          }'

Get Nodegroups:
*********************
* Follow below steps to Get Nodegroups for available cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=

   .. code-block:: bash

         curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"
        

Get Nodegroup:
*********************
* Follow below steps to Get Nodegroup for available cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=
         export NODEGROUPID=

   .. code-block:: bash

         curl -X GET ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups/${NODEGROUPID} -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"

Update Nodegroup Count:
***********************
* Follow below steps to Update Nodegroup for available cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=
         export NODEGROUPID=

   .. code-block:: bash

         curl -X PUT \
         -H 'Content-type: application/json' \
         -H "Authorization: Bearer ${TOKEN}" \
         ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups/${NODEGROUPID} -d '{"count":3}'


Delete Nodegroup:
*********************
* Follow below steps to Delete Nodegroup for available cluster in IDC env.

#. 
   .. code-block:: bash

         export URL_PREFIX=https://internal-placeholder.com
         export CLOUDACCOUNT=012345678912
         export CLUSTERID=
         export NODEGROUPID=

   .. code-block:: bash

         curl -X DELETE ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/nodegroups/${NODEGROUPID} -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"


Create Vip:
*********************
* Follow below steps to Create vip for available cluster in IDC env.

#.
  .. code-block:: bash

      export URL_PREFIX=https://internal-placeholder.com
      export CLOUDACCOUNT=012345678912
      export CLUSTERID=

  .. code-block:: bash

      curl -X POST \
         -H 'Content-type: application/json' \
         -H "Authorization: Bearer ${TOKEN}" \
         ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/vips -d \
         '{
            "name" : "ilb-defult",
            "description": "test-ilb",
            "port" : 80,
            "viptype" : "public"
          }'

Delete vip:
*********************
* Follow below steps to Delete vip fr available cluster in IDC env.

#.
  .. code-block:: bash

      export URL_PREFIX=https://internal-placeholder.com
      export CLOUDACCOUNT=012345678912
      export CLUSTERID=
      export VIPID=

  .. code-block:: bash

      curl -X DELETE ${URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/iks/clusters/${CLUSTERID}/vips/${VIPID} -H 'Content-type: application/json' -H "Authorization: Bearer ${TOKEN}"


.. _User Account Guide: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/7f82903653e162e201afbcaf9f98d5a316945a28/docs/source/public/guides/user_accounts.rst
.. _SSH Key Guide: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/7f82903653e162e201afbcaf9f98d5a316945a28/docs/source/public/guides/ssh_keys.rst
