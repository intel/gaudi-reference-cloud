.. _services_deployment_procedure_development:

IDC Services Deployment Procedure for Development Environments
##############################################################

.. warning::
   This document is mostly obsolete.
   Refer to `Deploying IDC Services in a Flexential Environment`_.

This document describes how to install IDC services in a *new* *development* environment.

This document omits some steps that are common with deploying to production and staging environments
as described in :ref:`services_deployment_procedure`

Related Documentation
*********************

* `Deploying IDC Services in a Flexential Environment`_

* To upgrade IDC services in an existing development environment, see :ref:`services_upgrade_procedure_development`.

* To install IDC services in a new *production* or *staging* environment, see :ref:`services_deployment_procedure`.

* To quickly deploy IDC in a single-node Kind cluster in the Flexential development environment, see
  `IDC Flexential Development Environments <https://internal-placeholder.com/x/MSPhs>`__.

Prepare your environment
************************

Unless otherwise specified, the steps in this document should be run in
a development VM in the Intel corporate environment.

Identify IDC Environment Name
******************************

Identify IDC environment name. Refer to `IDC Environments Master List <https://internal-placeholder.com/x/uyLhs>`__.

Prepare Monorepo for a new environment
**************************************

#. Create the directory structure build/environments/${IDC_ENV}. Copy
   from a similar environment such as dev4.

#. Create the file
   deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl.

#. Add the environment to deployment/helmfile/environments.yaml.

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

TODO

Populate Machine Images in Compute Database
===========================================

#. Edit MachineImage YAML files in the directory
   `go/pkg/compute_api_server/testdata/MachineImage <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/go/pkg/compute_api_server/testdata/MachineImage>`__.

#. If IDC has *not* been installed yet, then skip the remainder of this
   procedure. (The populate-machine-images job will be created later.)

   If you have already deployed IDC and only want to add or remove
   machine images, run:

   .. code-block:: bash

      make show-config
      HELMFILE_OPTS="destroy --selector name=${REGION}-populate-machine-images" make run-helmfile
      HELMFILE_OPTS="apply --selector name=${REGION}-populate-machine-images" make run-helmfile

#. Run ``make run-k9s`` and confirm that job
   *${REGION}-populate-machine-images-git-to-grpc-synchronizer* completed 1/1
   times.

#. Confirm that the machine image is now available in the IDC Console.

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

      aws sso login --profile idc-services-dev_390677890188-390677890188

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

Quick procedure to upgrade Portal (IDC UI)
******************************************

#. Checkout your branch in your development workstation.

#. Push your branch to Github.

#. Ensure that Jenkins successfully runs Bazel Container Push and Bazel
   Helm Push.

#. Change Portal chart and image versions in
   `deployment/helmfile/default.yaml.gotmpl <../deployment/helmfile/default.yaml.gotmpl>`__.
   Do *not* commit your changes yet since this will cause the version
   string to change.

   This step will become obsolete after
   https://internal-placeholder.com/browse/TWC4729-350.

#. Run the following in your development workstation.

   .. code-block:: bash

      export IDC_ENV=dev1
      make show-config
      HELMFILE_OPTS="apply --selector chart=idcs-portal" make run-helmfile

#. Portal should now be running.

#. Commit your changes, push to Github, and submit a PR to merge to
   main.

Quick procedure to upgrade a subset of IDC
******************************************

#. Checkout your branch in your development workstation.

#. Push your branch to Github.

#. Ensure that Jenkins successfully runs Bazel Container Push and Bazel
   Helm Push.

#. Run the following in your development workstation.

   Change the selector as needed.

   .. code-block:: bash

      export IDC_ENV=dev1
      HELMFILE_OPTS="apply --selector chart=billing" make run-helmfile

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

Create DNS CNAME records (Jones Farm only)
==========================================

For development environments in Jones Farm, create the following CNAME
records using https://internal-placeholder.com/.

-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}-internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com
-  {{ .Environment.Name }}internal-placeholder.com

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

Uninstall all components to avoid upgrading
*******************************************

.. warning:: THIS WILL DELETE ALL DATA.**

#. Delete all instances.

#. Run the following:

   .. code-block:: bash

      make show-config
      export KUBECONFIG=...
      kubectl get nodes
      HELMFILE_OPTS="destroy" make run-helmfile
      make undeploy-billing-db undeploy-compute-db undeploy-cloudaccount-db undeploy-metering-db undeploy-catalog-db undeploy-vault
      kubectl delete -n idcs-enrollment pvc/data-netbox-postgresql-0
      kubectl delete -n idcs-enrollment pvc/redis-data-netbox-redis-replicas-0
      kubectl delete -n idcs-enrollment pvc/redis-data-netbox-redis-master-0

Create Secrets
==============

.. code-block:: bash

   make secrets

Configure kubectl
=================

.. code-block:: bash

   make show-config

Copy/paste the line containing ``KUBECONFIG``, with ``export`` before
it.

.. code-block:: bash

   export KUBECONFIG=...
   kubectl get nodes

Install Vault
==============

Skip this section if Vault is already installed.

*Note: Requires longhorn.*

.. code-block:: bash

   export IDC_ENV=dev1
   SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace --version 1.4.1 
   kubectl create namespace kube-system
   kubectl delete secret -n kube-system wildcard-tls
   kubectl create secret tls \
       -n kube-system \
       wildcard-tls \
       --cert=${SECRETS_DIR}/wildcard-tls/tls.crt \
       --key=${SECRETS_DIR}/wildcard-tls/tls.key
   make test-helmfile
   make deploy-vault

OBSOLETE - Load TLS secrets (Development)
*****************************************

These steps should be repeated for all regional and AZ Kubernetes
clusters.

.. code-block:: bash

   export IDC_ENV=dev-jf
   SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   export KUBECONFIG=...
   kubectl create namespace idcs-system
   kubectl delete secret -n idcs-system wildcard-tls
   kubectl create secret tls \
       -n idcs-system \
       wildcard-tls \
       --cert=${SECRETS_DIR}/wildcard-tls/tls.crt \
       --key=${SECRETS_DIR}/wildcard-tls/tls.key

Load secrets into Vault
***********************

This section is required only when new secrets are required or existing
secrets are changed.

#. Browse to the Vault UI.

   For development environments in Flexential, use
   https://internal-placeholder.com/ui/vault/auth?with=oidc.

   Some environments have dedicated Vault clusters such as
   https://internal-placeholder.com.

#. Open the top-right menu and click "Copy token".

#. Create the file
   `../local/secrets/dev1/VAULT_TOKEN <../local/secrets/dev1/VAULT_TOKEN>`__
   and paste the token.

#. Run:

   If Vault is outside the Intel corporate network, add
   ``no_proxy= NO_PROXY=`` immediately before the following ``make``
   commands.

   .. code-block:: bash

      export IDC_ENV=dev1
      export IPA_IMAGE_SSH_PRIV=local/secrets/${IDC_ENV}/ipa-ssh
      export IPA_IMAGE_SSH_PUB=local/secrets/${IDC_ENV}/ipa-ssh.pub
      make show-config
      make deploy-vault-configure
      make deploy-vault-secrets

Create Database
================

#. Get psqlcompute_admin password from
   https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-compute/customer.
   Use this value for PGUSER.

#. Create database. Run this in a Postgres client pod in the regional
   cluster.

   .. code-block:: bash

      export PGUSER=psqlcompute_admin
      export PGPASSWORD=...
      export PGHOST=100.64.17.215
      export PGDATABASE=postgres
      DB_USER_USERNAME=dbuser
      psql -c "grant ${DB_USER_USERNAME} to psqlcompute_admin;"
      psql -c "create database main;"
      psql -c "alter database main owner to ${DB_USER_USERNAME};"
      psql -c "alter role ${DB_USER_USERNAME} with login;"
      psql -c "grant connect on database main to ${DB_USER_USERNAME};"
      psql -c "grant all privileges on database main to ${DB_USER_USERNAME};"

Configure Postgres database user (AWS)
**************************************

Shell into postgres-client pod.

Get password from Vault path controlplane/show/billing/aws-database.

.. code-block:: bash

   export PGUSER=billing_user
   export PGHOST=dev-idc-global-postgresqlv2.cluster-cb6hxdt0onur.us-west-2.rds.amazonaws.com
   export PGDATABASE=billing
   psql

Run Postgres Client in Kubernetes (development Postgres deployed by Helmfile)
******************************************************************************

This section applies to development envirionments in which the Postgres
database is deployed with Helmfile.

For development, the ``postgres`` user password is not rotated by Vault.
It can be used with the Postgres client in the \*-db-postgresql-0 pod.

.. code-block:: bash

   PGPASSWORD=$POSTGRES_POSTGRES_PASSWORD psql -U postgres -d ${POSTGRES_DB}

Configure Tenant SSH Proxy Server
*********************************

Jones Farm (JF)
===============

Create a user named ``guest-dev1``. See
`ssh-proxy-vm <../idcs_domain/ssh-proxy-vm/README.md>`__.

Flexential
==========

See `Deploy Tenant SSH Proxy Server <../../../../deployment/deploy-ssh-proxy-server.md>`__.

Obtain Host Public Key
======================

Both SSH Proxy Operator and BM Instance Operator needs the public key of the SSH Proxy Server to verify it before establishing a connection.

Obtain the host public key secret using the following command:

.. code-block:: bash

   ssh-keyscan -t rsa ${SSH_PROXY_IP} | awk '{print $2, $3}' > local/secrets/${IDC_ENV}/ssh-proxy-operator/host_public_key

Create TLS secrets in Kubernetes clusters
*****************************************

This section must be repeated for each Kubernetes cluster (except
Harvester).

.. code-block:: bash

   KUBECONFIG=$(pwd)/local/secrets/${IDC_ENV}/kubeconfig/${IDC_ENV}.yaml make deploy-k8s-tls-secrets

Create image pull secrets in Kubernetes clusters (Flexential and AWS only)
**************************************************************************

This section must be repeated for each Kubernetes cluster (except
Harvester).

.. code-block:: bash

   export IDC_ENV=dev3
   export SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
   export HARBOR_USERNAME="$(cat ${SECRETS_DIR}/HARBOR_USERNAME)"
   export HARBOR_PASSWORD="$(cat ${SECRETS_DIR}/HARBOR_PASSWORD)"
   KUBECONFIG=$(pwd)/local/secrets/${IDC_ENV}/kubeconfig/${IDC_ENV}.yaml make deploy-k8s-image-pull-secrets

Allocate Tenant Subnets to Region (Flexential only)
***************************************************

This section is not required for Jones Farm.

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

#. Check Allocated.

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

#. If IDC has *not* been installed yet, then skip the remainder of this
   procedure. (The populate-subnet job will be created later.)

   If you have already deployed IDC and only want to add or remove
   subnets, run:

   .. code-block:: bash

      make show-config
      export DOCKER_TAG=n2161-hdb22404b
      HELMFILE_OPTS="diff --selector name=${REGION}-populate-subnet" make run-helmfile
      HELMFILE_OPTS="destroy --selector name=${REGION}-populate-subnet" make run-helmfile
      HELMFILE_OPTS="apply --selector name=${REGION}-populate-subnet" make run-helmfile

#. Run ``make run-k9s`` and confirm that job
   *${REGION}-populate-subnet-git-to-grpc-synchronizer* completed 1/1
   times.

Push containers and Helm charts
*******************************

Jones Farm
===========

This is normally performed by Jenkins. If Jenkins is not available, you
may run the steps below.

.. code-block:: bash

   export IDC_ENV=dev1
   docker login internal-placeholder.com
   helm registry login internal-placeholder.com
   make container-push helm-push

.. _flexential-1:

Flexential Development
======================

This is normally performed by Jenkins. If Jenkins is not available, you
may run the steps below.

#. Create Harbor robot acccount for intelcloud project.

#. Create files ${SECRETS_DIR}/HARBOR_USERNAME and HARBOR_PASSWORD.

#. Run:

   .. code-block:: bash

      export IDC_ENV=dev3
      export SECRETS_DIR=$(pwd)/local/secrets/${IDC_ENV}
      export DOCKER_REGISTRY=amr-idc-registry-pre.infra-host.com
      export HARBOR_USERNAME="$(cat ${SECRETS_DIR}/HARBOR_USERNAME)"
      export HARBOR_PASSWORD="$(cat ${SECRETS_DIR}/HARBOR_PASSWORD)"
      echo "${HARBOR_PASSWORD}" | docker login -u ${HARBOR_USERNAME} --password-stdin ${DOCKER_REGISTRY}
      make container-push helm-push

Production
==========

#. Jenkins currently does not push containers and Helm charts to Harbor directly from a branch.
   To work around this, edit /Jenkinsfile and remove the `when` sections for "Bazel Container Push" and "Bazel Helm Push" stages.

Configure and check Helmfile
*****************************

By default, this will use the Helm charts tagged with the the current
git commit.

To use a different commit, run ``export DOCKER_TAG=nXXXX-hXXXXXXXX``.
You can get the git commit hash from the log of any Jenkins Bazel Helm
Push job.

**Important**: Since production global services are deployed using
ArgoCD, not helmfile, ensure that ``global.enabled`` is false in
deployment/helmfile/environments/prod.yaml.gotmpl.

.. code-block:: bash

   export IDC_ENV=dev3
   make show-config
   make test-helmfile

If you are performing an upgrade, review the differences between the
Kubernetes resources that are currently installed and the resources that
will be applied.

To check all clusters:

.. code-block:: bash

   HELMFILE_OPTS="diff" make run-helmfile |& tee local/${IDC_ENV}-helmfile-diff.log
   egrep '^\-|^\+' local/${IDC_ENV}-helmfile-diff.log | less

To check only a single region:

.. code-block:: bash

   HELMFILE_OPTS="diff --selector region=us-dev3-2" make run-helmfile |& tee local/${IDC_ENV}-helmfile-diff.log
   egrep '^\-|^\+' local/${IDC_ENV}-helmfile-diff.log | less

Deploy Helm releases with Custom Resource Definitions (CRDs)
************************************************************

Helm releases with CRDs must be installed before other Helm releases.

To deploy to all clusters:

.. code-block:: bash

   make deploy-crds |& tee -a local/${IDC_ENV}-deploy-crds.log

To deploy only to a single region:

.. code-block:: bash

   HELMFILE_OPTS="sync --selector region=us-dev3-2,crd=true" make run-helmfile |& tee local/${IDC_ENV}-deploy-crds.log

Deploy Vault Agent Injector
****************************

To deploy to all clusters:

.. code-block:: bash

   make deploy-vault-releases |& tee -a local/${IDC_ENV}-deploy-vault-releases.log

To deploy only to a single region:

.. code-block:: bash

   HELMFILE_OPTS="apply --selector region=us-dev3-2,chart=vault" make run-helmfile |& tee local/${IDC_ENV}-deploy-vault-releases.log

Deploy all remaining Helm releases
**********************************

To deploy to all clusters:

.. code-block:: bash

   make deploy-all-helm-releases |& tee -a local/${IDC_ENV}-deploy-all-helm-releases.log

To deploy only to a single region:

.. code-block:: bash

   HELMFILE_OPTS="apply --selector region=us-dev3-2" make run-helmfile |& tee local/${IDC_ENV}-deploy-all-helm-releases.log

Deploy a subnet of components
*****************************

These commands may be useful when only a subset of components should be
updated.

.. code-block:: bash

   export IDC_ENV=dev3
   make show-config
   HELMFILE_OPTS="apply --selector chart=debug-tools" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=vault" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=opentelemetry-collector-agent" make run-helmfile
   HELMFILE_OPTS="apply --selector name=cloudaccount" make run-helmfile
   HELMFILE_OPTS="apply --selector name=grpc-proxy" make run-helmfile
   HELMFILE_OPTS="apply --selector name=us-dev3-1-grpc-rest-gateway" make run-helmfile
   HELMFILE_OPTS="apply --selector name=us-dev3-1-grpc-proxy" make run-helmfile
   HELMFILE_OPTS="apply --selector name=us-dev-1-grpc-proxy" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=compute-api-server" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=bm-instance-operator" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=vm-instance-operator" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=vm-instance-scheduler" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=compute-metering-monitor" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=instance-replicator" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=ssh-proxy-operator" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=compute-api-server --selector chart=compute-crds" make run-helmfile
   HELMFILE_OPTS="apply --selector geographicScope=regional" make run-helmfile
   HELMFILE_OPTS="apply --selector geographicScope=az,service=compute --selector geographicScope=az,service=compute-vm" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=metal3-crds" make run-helmfile
   HELMFILE_OPTS="apply --selector geographicScope=az" make run-helmfile
   HELMFILE_OPTS="apply --selector geographicScope=global" make run-helmfile
   HELMFILE_OPTS="apply --selector service!=compute-bm" make run-helmfile
   HELMFILE_OPTS="apply --selector chart!=metallb-custom-resources" make run-helmfile
   HELMFILE_OPTS="destroy --selector geographicScope=global,chart=vault" make run-helmfile
   HELMFILE_OPTS="destroy --selector name=oidc" make run-helmfile
   HELMFILE_OPTS="destroy --selector chart=git-to-grpc-synchronizer" make run-helmfile
   HELMFILE_OPTS="apply --selector chart=git-to-grpc-synchronizer" make run-helmfile

Create AWS Route53 DNS records
==============================

#. Create a CNAME or A record for the DNS name specified in
   ``global.grpcProxy.internal.ingress.host`` to resolve to the AWS
   Application Load Balancer for GRPC. This has a name such as
   ``dualstack.k8s-devgrpcidcglobal-814970eab7-907006937.us-west-2.elb.amazonaws.com``.

#. Create a CNAME or A record for the DNS name specified in
   ``global.grpcRestGateway.ingress.host`` to resolve to the AWS
   Application Load Balancer for REST. This has a name such as
   ``dualstack.k8s-devrestidcglobal-2c3c04b017-1130606621.us-west-2.elb.amazonaws.com.``.

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
==========================================================

The IDC console (portal) uses the address specified in the Helmfile
environment parameter ``global.portal.ingress.host``. This will
generally have the form https://console-${IDC_ENV}.internal-placeholder.com.

Intel SSO must be configured to allow redirects to this address.
Procedure TBD.

Add Coupons (AWS)
=================

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

.. _token-not-required-grpcproxyexternalinsecureskipjwtvalidationtrue:

Token not required (grpcProxy.external.insecureSkipJwtValidation=true)
-----------------------------------------------------------------------

When ``global.grpcProxy.external.insecureSkipJwtValidation`` is
``true``, a token is not required. The ``TOKEN`` environment variable
can have any value or it can be not set at all.

.. _get-a-token-from-the-development-only-zitadel-oidc-provider:

Get a token from the development-only Zitadel OIDC provider 
-----------------------------------------------------------

(grpcProxy.external.insecureSkipJwtValidation=false and grpcProxy.external.insecureDevEnvironment=true)

The environment variable ``IDC_OIDC_URL_PREFIX`` must be the value from
``global.oidc.ingress.host``.

.. code-block:: bash

   export TOKEN=$(curl "${IDC_OIDC_URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
   echo ${TOKEN}

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

Generate ArgoCD Manifests
*************************

Use this procedure to generate ArgoCD manifests (Helm values and
metadata).

If ``global.enabled`` is false, set it to true temporarily for this
procedure.

.. code-block:: bash

   export IDC_ENV=dev3
   make helmfile-generate-argocd-values

Uninstalling
============

Delete instances
----------------

If there are breaking changes to the Instances schema (in Compute DB or
CRDs), delete all instances before continuing.

.. code-block:: bash

   PGPASSWORD=$POSTGRES_POSTGRES_PASSWORD psql -U postgres -d ${POSTGRES_DB}
   UPDATE instance SET value = jsonb_set(value,'{metadata, deletionTimestamp}', JSONB '"2023-06-30T13:34:26.948057750Z"');
   UPDATE instance SET resource_version = resource_version + 1000000000;

Delete Baremetalhosts
======================

TODO

Delete all Helm releases
========================

.. code-block:: bash

   export IDC_ENV=dev1

   ## Tag of the version getting deleted
   export DOCKER_TAG=n1340-h8475014a
   HELMFILE_OPTS="destroy --selector crd!=true" make run-helmfile
   HELMFILE_OPTS="destroy" make run-helmfile

Delete all databases
====================

By default, Postgres PVCs are not deleted when deleting Helm releases.
When reinstalling Postgres, it will re-use an existing PVC. To
permanently delete all databases, perform the steps below.

.. code-block:: bash

   make undeploy-billing-db
   make undeploy-compute-db
   make undeploy-cloudaccount-db
   make undeploy-metering-db

Delete AWS Database Schemas
===========================

#.  Run psql and connect to AWS databases (billing, cloudaccount, metering).
    See `IDC Environment dev3 <https://internal-placeholder.com/x/UyPhs>`__ for example.

#.  Drop tables.

   .. code-block:: console

      root@postgres-client:/# export PGDATABASE=billing
      root@postgres-client:/# export PGUSER=billing_user
      root@postgres-client:/# psql
      Password for user billing_user:

      billing=> drop table cloud_credits_intel;
      DROP TABLE
      billing=> drop table coupons;
      DROP TABLE
      billing=> drop table credit_usage;
      DROP TABLE
      billing=> drop table redemptions;
      DROP TABLE
      billing=> drop table schema_migrations;
      DROP TABLE
      billing=> \dt
      Did not find any relations.

   .. code-block:: console

      root@postgres-client:/# export PGDATABASE=cloudaccount
      root@postgres-client:/# export PGUSER=cloudacct_user
      root@postgres-client:/# psql
      Password for user cloudacct_user:

      cloudaccount=> drop table cloud_account_members;
      DROP TABLE
      cloudaccount=> drop table cloud_accounts;
      DROP TABLE
      cloudaccount=> drop table schema_migrations;
      DROP TABLE
      cloudaccount=> \dt
      Did not find any relations.

   .. code-block:: console

      root@postgres-client:/# export PGDATABASE=metering
      root@postgres-client:/# export PGUSER=metering_user
      root@postgres-client:/# psql
      Password for user metering_user:

      metering=> \dt
                     List of relations
      Schema |       Name        | Type  |     Owner     
      --------+-------------------+-------+---------------
      public | schema_migrations | table | metering_user
      public | usage_report      | table | metering_user
      (2 rows)

      metering=> drop table schema_migrations;
      DROP TABLE
      metering=> drop table usage_report;
      DROP TABLE

Delete Vault Secrets
=====================

Delete the following secret mounts:

- controlplane
- anything with ${IDC_ENV}
- public

Other
-----

.. code-block:: bash

   export KUBECONFIG=local/secrets/${IDC_ENV}/kubeconfig/${IDC_ENV}.yaml
   kubectl delete namespace idcs-system
   kubectl delete namespace idcs-enrollment
   kubectl delete namespace opal-server
   kubectl delete namespace metal3-1
   kubectl delete namespace metallb-system
   kubectl delete namespace idcs-observability
   kubectl delete namespace vault
   kubectl delete crd bminstanceoperatorconfigs.private.cloud.intel.com
   kubectl delete crd computemeteringmonitorconfigs.private.cloud.intel.com
   kubectl delete crd instancereplicatorconfigs.private.cloud.intel.com
   kubectl delete crd instances.private.cloud.intel.com
   kubectl delete crd products.private.cloud.intel.com
   kubectl delete crd sshproxyoperatorconfigs.private.cloud.intel.com
   kubectl delete crd sshproxytunnels.private.cloud.intel.com
   kubectl delete crd vendors.private.cloud.intel.com
   kubectl delete crd vminstanceoperatorconfigs.private.cloud.intel.com
   kubectl delete crd vminstanceschedulerconfigs.private.cloud.intel.com
   kubectl delete crd baremetalhosts.metal3.io
   kubectl delete crd bmceventsubscriptions.metal3.io
   kubectl delete crd firmwareschemas.metal3.io
   kubectl delete crd hardwaredata.metal3.io
   kubectl delete crd hostfirmwaresettings.metal3.io
   kubectl delete crd preprovisioningimages.metal3.io
   kubectl delete crd authorizationpolicies.security.istio.io
   kubectl delete crd destinationrules.networking.istio.io
   kubectl delete crd envoyfilters.networking.istio.io
   kubectl delete crd gateways.networking.istio.io
   kubectl delete crd istiooperators.install.istio.io
   kubectl delete crd peerauthentications.security.istio.io
   kubectl delete crd proxyconfigs.networking.istio.io
   kubectl delete crd requestauthentications.security.istio.io
   kubectl delete crd serviceentries.networking.istio.io
   kubectl delete crd sidecars.networking.istio.io
   kubectl delete crd telemetries.telemetry.istio.io
   kubectl delete crd virtualservices.networking.istio.io
   kubectl delete crd wasmplugins.extensions.istio.io
   kubectl delete crd workloadentries.networking.istio.io
   kubectl delete crd workloadgroups.networking.istio.io

Troubleshooting
===============

-  If only a ConfigMap is changed, the pod will not be restarted. You
   will need to delete the pod so that it gets recreated with the
   updated ConfigMap.

-  If breaking changes are made to the Compute DB schema
   (https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/go/pkg/compute_api_server/db/migrations),
   it should be recreated as follows:

   #. Open k9s.
   #. Delete pvc idcs-system/data-compute-db-postgresql-0. It will show
      as terminating.
   #. Delete pod idcs-system/compute-db-postgresql-0.
   #. Delete pod `compute-api-server-*`.

-  You may run psql with the following command:

.. code-block:: bash

   PGPASSWORD=$POSTGRES_POSTGRES_PASSWORD psql -U postgres -d ${POSTGRES_DB}



.. _Harvester ISO Installation: https://docs.harvesterhci.io/v1.1/install/iso-install
.. _Software Inflow: https://internal-placeholder.com/x/4rVCtQ
.. _Deploying IDC Services in a Flexential Environment: https://internal-placeholder.com/x/8oyY1
