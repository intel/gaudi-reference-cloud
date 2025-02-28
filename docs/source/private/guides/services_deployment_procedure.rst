.. _services_deployment_procedure:

IDC Services Deployment Procedure
#################################

This document describes how to install IDC services in a *new* *production* or *staging* environment.

Related Documentation
*********************

* To *upgrade* a production or staging environment, see :ref:`services_upgrade_procedure`.

* To install IDC services in a new *development* environment, see :ref:`services_deployment_procedure_development`.

Prepare your environment
************************

Unless otherwise specified, the steps in this document should be run in
a development VM in the Intel corporate environment.

Identify IDC Environment Name
******************************

Identify IDC environment name. Refer to `IDC Environments Master List <https://internal-placeholder.com/x/uyLhs>`__.

Prepare development workstation
===============================

#. Run the following to install the necessary tools in your development
   workstation.

   .. code-block:: bash

      make install-interactive-tools

#. Set IDC_ENV environment variable. Be sure to run this in all new Bash
   sessions.

   .. code-block:: bash

      export IDC_ENV=dev1
      SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}

Load Bare Metal Machine Images into the HTTP Server
===================================================

#. Ensure that the machine image qcow file has been uploaded to the `Machine Image S3 Bucket`_.
   This should be done automatically by `Software Inflow`_.   

Install Kubernetes
------------------

For all Kubernetes clusters except AWS EKS, RKE2 should be installed. In
production, this will be installed using Rancher.

The following NGINX Ingress Controller customizations are required to
allow use of NGINX snippets that set GRPC timeouts and to allow SSL
passthrough.

**WARNING!** If this step is not done, NGINX will silently ignore the
``nginx.ingress.kubernetes.io/ssl-passthrough: true`` annotation,
resulting in mTLS connection failures.

.. code-block:: bash

   kubectl apply -f deployment/rke2/root/var/lib/rancher/rke2/server/manifests/rke2-ingress-nginx-config.yaml
   helm get values -n kube-system rke2-ingress-nginx

Obtain Environment-specific Secrets
===================================

This sections shows how to configure environment-specific secrets in
``local/secrets/${IDC_ENV}``.

#. Download the Kubeconfig files for all global, regional, and AZ
   clusters and place them in the directory
   ``local/secrets/${IDC_ENV}/kubeconfig/``.

#. Set permissions of Kubeconfig files.

   .. code-block:: bash

      chmod 600 local/secrets/${IDC_ENV}/kubeconfig/*

#. Obtain Kubernetes public key (see `Get Kubernetes public
   key <#get-kubernetes-public-key>`__).

#. Download TLS wildcard certificates to
   ``local/secrets/${IDC_ENV}/wildcard-tls/`` (ask for them to the team
   lead).

#. If needed, authenticate to AWS. See `How to connect to AWS to manage
   IDC global services <https://internal-placeholder.com/x/LSl2sQ>`__ for more
   details.

   .. code-block:: bash

      aws sso login --profile AWSAdministratorAccess-045705861988

#. Ensure proper Kubernetes context names. In particular, EKS context
   names should not have ``:`` characters.

   .. code-block:: bash

      kubectl config rename-context arn:aws:eks:us-west-2:390677890188:cluster/dev-idc-global dev-idc-global

Get Kubernetes public key
*************************

This must be repeated for each Kubernetes cluster (except Harvester).

Copy deployment/common/vault/get-kubernetes-public-keys.sh to an RKE2
node, as shown below.

.. code-block:: bash

   scp deployment/common/vault/get-kubernetes-public-keys.sh sdp@10.165.161.212:

Execute get-kubernetes-public-keys.sh in the RKE2 node as shown below.

.. code-block:: bash

   ssh sdp@10.165.161.212
   sudo apt install jq
   sudo -i
   export KUBECONFIG=/etc/rancher/rke2/rke2.yaml
   export PATH=${PATH}:/var/lib/rancher/rke2/bin
   kubectl get nodes
   ~sdp/get-kubernetes-public-keys.sh

Copy the file in /tmp/vault-jwk-validation-public-keys/ to the following
locations:

-  ${SECRETS_DIR}/vault-jwk-validation-public-keys/${CLUSTER_NAME}.jwk
-  build/environments/${IDC_ENV}/vault-jwk-validation-public-keys/${CLUSTER_NAME}.jwk

These public keys should be checked into source control.

Create DNS Records and Load Balancers
*************************************

The following lists the DNS records and load balancers required for
ingress.

+---------+-----------+-----------+------+-----------+-----------+
| Scope   | Helmfile  | Targets   | Port | Re        | Notes     |
|         | en        |           |      | commended |           |
|         | vironment |           |      | FQDN      |           |
|         | parameter |           |      |           |           |
+=========+===========+===========+======+===========+===========+
| public  | global.p  | global    | 443  | dev.cons  |           |
|         | ortal.ing | C         |      | ole.idcse |           |
|         | ress.host | loudFront |      | rvice.net |           |
+---------+-----------+-----------+------+-----------+-----------+
| public  | global.g  | global    | 443  | dev.      |           |
|         | rpcRestGa | ingress   |      | api.idcse |           |
|         | teway.ing |           |      | rvice.net |           |
|         | ress.host |           |      |           |           |
+---------+-----------+-----------+------+-----------+-----------+
| private | gl        | global    | 443  | dev.grpc  |           |
|         | obal.grpc | ingress   |      | api.idcse |           |
|         | Proxy.int |           |      | rvice.net |           |
|         | ernal.ing |           |      |           |           |
|         | ress.host |           |      |           |           |
+---------+-----------+-----------+------+-----------+-----------+
| public  | re        | regional  | 443  | de        |           |
|         | gions[].g | K8s       |      | v3-comput |           |
|         | rpcRestGa | ingress   |      | e-us-dev3 |           |
|         | teway.ing |           |      | -1-api-cl |           |
|         | ress.host |           |      | oud.eglb. |           |
|         |           |           |      | intel.com |           |
+---------+-----------+-----------+------+-----------+-----------+
| private | regio     | regional  | 443  | dev3-c    |           |
|         | ns[].grpc | K8s       |      | ompute-us |           |
|         | Proxy.int | ingress   |      | -dev3-1-g |           |
|         | ernal.ing |           |      | rpcapi-cl |           |
|         | ress.host |           |      | oud.eglb. |           |
|         |           |           |      | intel.com |           |
+---------+-----------+-----------+------+-----------+-----------+
| private | reg       | regional  | 443  | dev3-com  |           |
|         | ions[].co | K8s       |      | pute-api- |           |
|         | mputeApiS | ingress   |      | server-us |           |
|         | erver.ing |           |      | -dev3-1-g |           |
|         | ress.host |           |      | rpcapi-cl |           |
|         |           |           |      | oud.eglb. |           |
|         |           |           |      | intel.com |           |
+---------+-----------+-----------+------+-----------+-----------+
| private | re        | regional  | 443  | d         | Used only |
|         | gions[].n | K8s       |      | ev3-netbo | for       |
|         | etbox.ing | ingress   |      | x-us-dev3 | Netbox    |
|         | ress.host |           |      | -1-api-cl | d         |
|         |           |           |      | oud.eglb. | eployment |
|         |           |           |      | intel.com | in dev    |
|         |           |           |      |           | envi      |
|         |           |           |      |           | ronments. |
+---------+-----------+-----------+------+-----------+-----------+
| private | regions[  | AZ K8s    | 443  | dev3-co   |           |
|         | ].availab | ingress   |      | mpute-us- |           |
|         | ilityZone |           |      | dev3-1a-g |           |
|         | s[].vmIns |           |      | rpcapi-cl |           |
|         | tanceSche |           |      | oud.eglb. |           |
|         | duler.ing |           |      | intel.com |           |
|         | ress.host |           |      |           |           |
+---------+-----------+-----------+------+-----------+-----------+
| private | re        | AZ K8s    | 443  | de        | Called by |
|         | gions[].a | ingress   |      | v3-bareme | Netbox.   |
|         | vailabili |           |      | tal-enrol |           |
|         | tyZones[] |           |      | lment-api |           |
|         | .baremeta |           |      | -us-dev3- |           |
|         | lEnrollme |           |      | 1a-api-cl |           |
|         | ntApi.ing |           |      | oud.eglb. |           |
|         | ress.host |           |      | intel.com |           |
+---------+-----------+-----------+------+-----------+-----------+
| public  | region    | tenant    | 22   | d         |           |
|         | s[].avail | SSH proxy |      | ev3.ssh-1 |           |
|         | abilityZo |           |      | .us-dev3- |           |
|         | nes[].ssh |           |      | 1a.cloud. |           |
|         | Proxy.pro |           |      | intel.com |           |
|         | xyAddress |           |      |           |           |
+---------+-----------+-----------+------+-----------+-----------+

Create Load Balancers (Flexential only)
=======================================

For environments in Flexential, create load balancer VIPs using
https://internal-placeholder.com/.

#. dev3-compute-us-dev3-1-api-cloud

   -  Environment: DevCloud Staging - Infra Public - OR

      -  VIP:

         -  App Name: dev3-compute-us-dev3-1-api-cloud
         -  Port: 443
         -  Address Type: Public (Internet Routable)

      -  Pool:

         -  App Name: (same as VIP App Name)
         -  Members (repeat for all Kubernetes nodes in this cluster):

            -  IP: for example, 100.64.16.88
            -  Port: 443

#. dev3-ssh-1-us-dev3-1a-cloud

   -  Environment: DevCloud Staging - Tenant - OR

      -  VIP:

         -  App Name: dev3-ssh-1-us-dev3-1a-cloud
         -  Port: 22
         -  Address Type: Public (Internet Routable)

      -  Pool:

         -  App Name: (same as VIP App Name)
         -  Members (repeat for all tenant SSH proxy servers):

            -  IP: for example, 100.64.16.88
            -  Port: 22

TODO: document private load balancer configuration.

Load TLS secrets
****************

If needed, convert file in ``BEGIN ENCRYPTED PRIVATE KEY`` or
``BEGIN PRIVATE KEY`` format to ``BEGIN RSA PRIVATE KEY`` format.

.. code-block:: bash

   openssl rsa -in tls-encrypted.key -out tls.key -text

These steps should be repeated for all regional and AZ Kubernetes
clusters.

.. code-block:: bash

   export IDC_ENV=dev-jf
   SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   export KUBECONFIG=...
   kubectl create namespace idcs-system

   kubectl delete secret -n idcs-system wildcard-idcmgt-tls
   kubectl create secret tls \
       -n idcs-system \
       wildcard-idcmgt-tls \
       --cert=${SECRETS_DIR}/wildcard-idcmgt-tls/tls.crt \
       --key=${SECRETS_DIR}/wildcard-idcmgt-tls/tls.key

   kubectl delete secret -n idcs-system wildcard-cloud-tls
   kubectl create secret tls \
       -n idcs-system \
       wildcard-cloud-tls \
       --cert=${SECRETS_DIR}/wildcard-cloud-tls/tls.crt \
       --key=${SECRETS_DIR}/wildcard-cloud-tls/tls.key

These steps should be repeated for all Quick Connect AZ Kubernetes
clusters.

.. code-block:: bash

   export IDC_ENV=dev-jf
   SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   export KUBECONFIG=...
   kubectl create namespace idcs-system

   kubectl delete secret -n idcs-system wildcard-devcloudtenant-tls
   kubectl create secret tls \
       -n idcs-system \
       wildcard-devcloudtenant-tls \
       --cert=${SECRETS_DIR}/wildcard-devcloudtenant-tls/tls.crt \
       --key=${SECRETS_DIR}/wildcard-devcloudtenant-tls/tls.key

Configure Postgres database for Compute API Server
**************************************************

In the steps below, be sure to use the environment-specific Vault URL.

Create Database
================

#. Get the ``psqlcompute_admin`` password from
   https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-compute/customer.   
   This ``usernames`` and ``passwords`` fields are comma-separated values that correspond to each other.
   For example, the password for the first value in ``usernames`` is the first value in ``passwords``.
   Use this value for PGPASSWORD.

#. Create database. Run this in a Postgres client pod in the regional
   cluster.

   .. code-block:: bash

      export PGUSER=psqlcompute_admin
      export PGPASSWORD=...
      export PGHOST=100.64.17.215
      export PGDATABASE=postgres
      DB_ADMIN_USERNAME=psqlcompute_admin
      DB_USER_USERNAME=dbuser
      psql -c "grant ${DB_USER_USERNAME} to ${DB_ADMIN_USERNAME};"
      psql -c "create database main;"
      psql -c "alter database main owner to ${DB_USER_USERNAME};"
      psql -c "alter role ${DB_USER_USERNAME} with login;"
      psql -c "grant connect on database main to ${DB_USER_USERNAME};"
      psql -c "grant all privileges on database main to ${DB_USER_USERNAME};"

Configure Postgres database for Fleet Admin Server
**************************************************

In the steps below, be sure to use the environment-specific Vault URL.

Create Database
================

#. Get the ``psqlfleetadmindb_admin`` password from
   https://internal-placeholder.com/ui/vault/secrets/secret/kv/dbaas%2Fus-staging-1%2Fpsql-fleet-admin-db%2Fcustomer/details?version=1.
   This ``usernames`` and ``passwords`` fields are comma-separated values that correspond to each other.
   For example, the password for the first value in ``usernames`` is the first value in ``passwords``.
   Use this value for PGPASSWORD.

#. Create database. Run this in a Postgres client pod in the regional
   cluster.

   .. code-block:: bash

      export PGUSER=psqlfleetadmindb_admin
      export PGPASSWORD=...
      export PGHOST=100.64.17.221
      export PGDATABASE=postgres
      DB_ADMIN_USERNAME=psqlfleetadmindb_admin
      DB_USER_USERNAME=fleetadmindb_user
      psql -c "grant ${DB_USER_USERNAME} to ${DB_ADMIN_USERNAME};"
      psql -c "create database main;"
      psql -c "alter database main owner to ${DB_USER_USERNAME};"
      psql -c "alter role ${DB_USER_USERNAME} with login;"
      psql -c "grant connect on database main to ${DB_USER_USERNAME};"
      psql -c "grant all privileges on database main to ${DB_USER_USERNAME};"

Add Secrets to Vault for Fleet Admin Server
*******************************************

In the steps below, be sure to use the environment-specific Vault URL.

#. Obtain the database username and password from
   https://internal-placeholder.com/ui/vault/secrets/secret/kv/dbaas%2Fus-staging-1%2Fpsql-fleet-admin-db%2Fcustomer/details?version=1

#. Open the Vault UI at
   https://internal-placeholder.com/ui/vault/secrets/controlplane/kv/list.

#. Click Create secret.

   #. Path for this secret: us-staging-1-fleet-admin-api-server/database

   #. User name:

      #. Secret data key: username

      #. Secret data value: fleetadmindb_user

   #. Password:

      #. Secret data key: password

      #. Secret data password: (password from previous step)

   #. Click Save.

Add Staas Database Secrets to Vault
***********************************

In the steps below, be sure to use the environment-specific Vault URL.

#. Create database secrets for storage-api-server and storage-admin-api-server

   #. Obtain database username and password from
      https://internal-placeholder.com/ui/vault/secrets/secret/kv/dbaas%2Fus-region-2%2Fpsql-staas%2Fcustomer/details?version=2

#. Open the Vault UI at
   https://internal-placeholder.com//ui/vault/secrets/controlplane/kv/list.

#. Click Create secret.

   #. Paths for these secrets: 

      #. us-region-2-storage-api-server/database

      #. us-region-2-storage-admin-api-server/database

   #. User name:

      #. Secret data key: username

      #. Secret data value: (username from previous step)

   #. Password:

      #. Secret data key: password

      #. Secret data password: (password from previous step)

   #. Click Save.


Add Staas Cognito Secrets to Vault
**********************************

In the steps below, be sure to use the environment-specific Vault URL.

#. Generate new cognito client id and client secret for each staas service: reach out to vishnu.v.ravi@intel.com for these instructions 

#. Open the Vault UI at
   https://internal-placeholder.com//ui/vault/secrets/controlplane/kv/list.

#. Click Create secret.

   #. Create cognito secrets for following staas services with paths:

      #. us-region-2-storage-api-server/cognito

      #. us-region-2-storage-admin-api-server/cognito

      #. us-region-2-storage-resource-cleaner/cognito

      #. us-region-2a-storage-metering-monitor/cognito

      #. us-region-2a-bucket-metering-monitor/cognito

   #. Client Id:

      #. Secret data key: client_id

      #. Secret data value: (client_id from previous step)

   #. Client Secret:

      #. Secret data key: client_secret

      #. Secret data client_secret: (client_secret from previous step)

   #. Click Save.

Create database for STaaS
**************************

#. Get ``psqlstaas_admin`` password from
   https://internal-placeholder.com/ui/vault/secrets/secret/kv/dbaas%2Fpsql-staas%2Fcustomer
   This ``usernames`` and ``passwords`` fields are comma-separated values that correspond to each other.
   For example, the password for the first value in ``usernames`` is the first value in ``passwords``.
   Use this value for PGPASSWORD.

#. Create database. Run this in a Postgres client pod in the regional
   cluster.

   .. warning::
      The code below is for staging. For production, use the appropriate region.

   .. code-block:: bash

      export PGUSER=psqlstaas_admin
      export PGPASSWORD=...
      export PGHOST=100.64.17.218
      export PGDATABASE=postgres
      DB_ADMIN_USERNAME=psqlstaas_admin
      DB_USER_USERNAME=dbuser
      psql -c "grant ${DB_USER_USERNAME} to ${DB_ADMIN_USERNAME};"
      psql -c "create database main;"
      psql -c "alter database main owner to ${DB_USER_USERNAME};"
      psql -c "alter role ${DB_USER_USERNAME} with login;"
      psql -c "grant connect on database main to ${DB_USER_USERNAME};"
      psql -c "grant all privileges on database main to ${DB_USER_USERNAME};"

Create Vault Rules for KMS STaaS
*********************************

#. Apply rules in the vault for role ID and secret ID.

   .. warning::
      The code below is for staging. For production, use the appropriate region.

   .. code-block:: bash

      STORAGE_ROLE_KMS_ID=$(vault read auth/approle/role/us-staging-3-storage-kms-role/role-id -format=json | jq -r .data.role_id)
      STORAGE_SECRET_KMS_ID=$(vault write -f auth/approle/role/us-staging-3-storage-kms-role/secret-id -format=json | jq -r .data.secret_id)
      vault kv put -mount=controlplane us-staging-3/storage/kms/approle  secret_id=${STORAGE_SECRET_KMS_ID}  role_id=${STORAGE_ROLE_KMS_ID}

Configure Postgres database user (AWS)
**************************************

Shell into postgres-client pod.

Get password from Vault path controlplane/show/billing/aws-database.

.. code-block:: bash

   export PGUSER=billing_user
   export PGHOST=dev-idc-global-postgresqlv2.cluster-cb6hxdt0onur.us-west-2.rds.amazonaws.com
   export PGDATABASE=billing
   psql

Configure Tenant SSH Proxy Server
*********************************

See `Deploy Tenant SSH Proxy Server`_.

Obtain Host Public Key
======================

Both SSH Proxy Operator and BM Instance Operator needs the public key of the SSH Proxy Server to verify it before establishing a connection.

Obtain the host public key secret using the following command:

.. code-block:: bash

   ssh-keyscan -t rsa ${SSH_PROXY_IP} | awk '{print $2, $3}' > local/secrets/${IDC_ENV}/ssh-proxy-operator/host_public_key

Create TLS secrets in Kubernetes clusters
*****************************************

This section must be repeated for each Kubernetes cluster (except Harvester).

.. code-block:: bash

   KUBECONFIG=$(pwd)/local/secrets/${IDC_ENV}/kubeconfig/${IDC_ENV}.yaml make deploy-k8s-tls-secrets

Create image pull secrets in Kubernetes clusters
************************************************

This section must be repeated for each Kubernetes cluster (except Harvester).

.. code-block:: bash

   export IDC_ENV=dev3
   export SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   export HARBOR_USERNAME="$(cat ${SECRETS_DIR}/HARBOR_USERNAME)"
   export HARBOR_PASSWORD="$(cat ${SECRETS_DIR}/HARBOR_PASSWORD)"
   KUBECONFIG=$(pwd)/local/secrets/${IDC_ENV}/kubeconfig/${IDC_ENV}.yaml make deploy-k8s-image-pull-secrets

Allocate Tenant Subnets to Region
*********************************

.. _allocate-tenant-subnets-in-ddi-men--mice:

Allocate Tenant Subnets in DDI (Men & Mice)
===========================================

#. Login to https://internal-placeholder.com/ (development and staging) or
   https://internal-placeholder.com (production).

#. Change address space to pdx03-c01-tenant.

#. Click IPAM.

#. Enter "/24" in the Quick filter field.

#. Shift-click on the desired ranges within 100.80.0.0/14. You must
   choose only /24 ranges. Ensure that you only select ranges that are
   not already allocated.

#. Click the Edit Properties button.

#. Set the Description and ConsumerID fields to the region name such as
   ``us-dev3-1``.

#. Save.

Populate Subnets in Compute Database
====================================

#. Prepare environment variables.

   For development and staging only:

   .. code-block:: bash

      export IDC_ENV=staging
      REGION=us-${IDC_ENV}-1
      MEN_AND_MICE_URL=https://internal-placeholder.com

   For production only:

   .. code-block:: bash

      export IDC_ENV=prod
      REGION=us-region-1
      MEN_AND_MICE_URL=https://internal-placeholder.com

   For all:

   .. code-block:: bash

      AVAILABILITY_ZONE=${REGION}a
      SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
      MEN_AND_MICE_USERNAME=$(cat ${SECRETS_DIR}/MEN_AND_MICE_USERNAME)
      MEN_AND_MICE_PASSWORD=$(cat ${SECRETS_DIR}/MEN_AND_MICE_PASSWORD)

#. Extract Men & Mice ranges into IDC subnet files.

   .. code-block:: bash

      go/pkg/compute_api_server/ip_resource_manager/mmws-extract-subnets.py \
      --mmws-url "${MEN_AND_MICE_URL}" \
      --mmws-password "${MEN_AND_MICE_PASSWORD}" \
      --mmws-username "${MEN_AND_MICE_USERNAME}" \
      --output-dir build/environments/${IDC_ENV}/${REGION}/Subnet \
      --region ${REGION} \
      --availability-zone ${AVAILABILITY_ZONE}

Deploy Helm Releases Using Argo CD
**********************************

See :ref:`services_upgrade_procedure`.

Create AWS Route53 DNS records
******************************

#. Create a CNAME or A record for the DNS name specified in
   ``global.grpcProxy.internal.ingress.host`` to resolve to the AWS
   Application Load Balancer for GRPC. This has a name such as
   ``dualstack.k8s-devgrpcidcglobal-814970eab7-907006937.us-west-2.elb.amazonaws.com``.

#. Create a CNAME or A record for the DNS name specified in
   ``global.grpcRestGateway.ingress.host`` to resolve to the AWS
   Application Load Balancer for REST. This has a name such as
   ``dualstack.k8s-devrestidcglobal-2c3c04b017-1130606621.us-west-2.elb.amazonaws.com.``.

Deploy Harvester for VMaaS
**************************

See :ref:`harvester_deployment_procedure`.

Update Product Catalog Definitions
**********************************

All product catalog definitions (Custom Resources) are managed through
following repo:

https://github.com/intel-innersource/frameworks.cloud.devcloud.services.product-catalog

And, it currently requires you to apply these specs manually onto your
dev cluster. Follow these instructions to achieve that:

#. Clone the catalog definition repo locally

   .. code-block:: bash

      git clone https://github.com/intel-innersource/frameworks.cloud.devcloud.services.product-catalog

#. Make sure you have kubeconfig setup properly to target cluster

#. Apply product specs

   .. code-block:: bash

      cd staging
      kubectl apply -f vendors/ -n idcs-system
      kubectl apply -f products/ -n idcs-system

#. Verify product specs on your cluster

   .. code-block:: bash

      kubectl get products.private.cloud.intel.com -n idcs-system 
      NAME                AGE
      bm-icx              161m
      bm-icx-atsm-170-1   161m
      bm-icx-gaudi2       161m
      ...

Configure Intel SSO to allow redirects to the IDC console
*********************************************************

The IDC console (portal) uses the address specified in the Helmfile
environment parameter ``global.portal.ingress.host``. This will
generally have the form https://console-${IDC_ENV}.internal-placeholder.com.

Intel SSO must be configured to allow redirects to this address.
Procedure TBD.

Add Coupons (AWS)
*****************

.. code-block:: bash

   make show-config
   export KUBECONFIG=...
   kubectl apply -n idcs-system -f deployment/hack/postgres-client-pod.yaml

Note: Please delete postgres-client post this activity as it is a security risk.

Shell into postgres-client pod (``make run-k9s``).

Get password from Vault path controlplane/show/billing/aws-database.

.. code-block:: bash

   export PGUSER=billing_user
   export PGHOST=dev-idc-global-postgresqlv2.cluster-cb6hxdt0onur.us-west-2.rds.amazonaws.com
   export PGDATABASE=billing
   psql

.. code-block:: sql

   delete from coupons where num_redeemed=0;
   insert into coupons (code, amount, creator, start, created, expires, disabled, num_uses, num_redeemed)
   values ('52XL-RZ73-YLAA', 2500, 'gopesh-intel', current_timestamp, current_timestamp, current_timestamp + INTERVAL '30 day',NULL, 1, 0);
   select count(*) from coupons where num_redeemed=0;
   select code from coupons where num_redeemed=0 order by code limit 20;

For development, you can load sample coupons from go/pkg/test-data/100-coupons.txt.

End-to-End Test Procedure
*************************

Where to run this procedure
===============================

Run this from a workstation in the Intel corporate network.

Set environment variables for the environment
=============================================

.. code-block:: bash

   export IDC_ENV=dev1
   make show-config
   eval `make show-export`

If the API servers are outside of the Intel corporate network (in Flex
or AWS), you will need to change your proxy configuration to force
requests to \*.intel.com to use the proxy.

.. code-block:: bash

   export no_proxy=10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16:10.165.28.33
   export NO_PROXY=${no_proxy}

Get IDC API Token
==================

Use one of these methods to obtain an IDC API token.

.. _get-a-token-from-azure-ad-grpcproxyexternalinsecureskipjwtvalidationfalse-and-grpcproxyexternalinsecuredevenvironmentfalse:

Get a token from Azure AD (grpcProxy.external.insecureSkipJwtValidation=false and grpcProxy.external.insecureDevEnvironment=false)
-----------------------------------------------------------------------------------------------------------------------------------

This is the production configuration. A token from Azure AD must be
obtained.

To obtain this token:

#. Login to the IDC console using Chrome.

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

Create a Cloud Account
======================

*Note:* The Cloud Account created using this method will not be able to
login to the IDC Console. This procedure will create a Cloud Account
using your hostname to avoid conflicts with your @intel.com email
address.

.. code-block:: bash

   export CLOUDACCOUNTNAME=${USER}@$(hostname -f)
   go/svc/cloudaccount/test-scripts/cloud_account_create.sh
   export CLOUDACCOUNT=$(go/svc/cloudaccount/test-scripts/cloud_account_get_by_name.sh | jq -r .id)
   echo ${CLOUDACCOUNT}

Use Compute API to create VNet, SSH Public Key, and Instance
==============================================================

.. code-block:: bash

   go/svc/compute_api_server/test-scripts/vnet_create_with_name.sh
   go/svc/compute_api_server/test-scripts/sshpublickey_create_with_name.sh
   go/svc/compute_api_server/test-scripts/instance_create_with_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh

Configure your SSH Client (Flex)
================================

Configure your SSH client to use the Intel SOCKS proxy to reach the IDC
tenant SSH proxy. Add the following to ~/.ssh/config.

.. code-block:: console

   Host 146.152.*.*
     ProxyCommand /usr/bin/nc -x internal-placeholder.com:1080 %h %p

SSH to Instance
===============

.. code-block:: bash

   ssh -J guest-${IDC_ENV}@10.165.62.252 ubuntu@172.16.x.x

Delete Instance
===============

.. code-block:: bash

   go/svc/compute_api_server/test-scripts/instance_delete_by_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh

Open IDC Console (Portal)
=========================

Open your browser to the IDC Console (portal) address.

Troubleshooting
***************

Vault Errors
============

Symptom
-------

.. code-block:: console

   2023-06-13T03:37:31.328Z [INFO]  agent.auth.handler: authenticating
   2023-06-13T03:37:31.340Z [ERROR] agent.auth.handler: error authenticating:
     error=
     | Error making API request.
     | 
     | URL: PUT https://internal-placeholder.com:443/v1/auth/cluster-auth/login
     | Code: 400. Errors:
     | 
     | * error validating token: error verifying token signature: no known key successfully validated the token signature
      backoff=1m23.66s

Cause
-----

Kubernetes public key is not in Vault.

Resolution
----------

See `Get Kubernetes Public Key <#get-kubernetes-public-key>`__.



.. _Harvester ISO Installation: https://docs.harvesterhci.io/v1.1/install/iso-install
.. _Software Inflow: https://internal-placeholder.com/x/4rVCtQ
.. _Machine Image S3 Bucket: https://s3.console.aws.amazon.com/s3/buckets/catalog-fs-dev
.. _Deploy Tenant SSH Proxy Server: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/deploy-ssh-proxy-server.md
