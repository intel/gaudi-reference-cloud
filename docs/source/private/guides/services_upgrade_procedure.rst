.. _services_upgrade_procedure:

IDC Services Upgrade Procedure
##############################

This document describes how to *upgrade* IDC services in *production* and *staging* environments.

Related Documentation
*********************

* To install IDC services in a *new* production or staging environment, see :ref:`services_deployment_procedure`.

* To install IDC services in a new *development* environment, see :ref:`services_deployment_procedure_development`.

* `IDC Environments Master List`_

Prepare for Upgrade
*******************

Unless otherwise specified, the steps in this document should be run in
a development VM in the Intel corporate environment.

Steps in this section should be performed before the RFC maintenance window begins.

Review Code Changes
===================

#. Checkout target commit.

   .. code:: bash

      git checkout main

#. View all changes to the monorepo between the currently-deployed commit and the target commit.

   For example:

   https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/compare/8475014a...main

#. Run the following command to identify changes to critical files.

   .. code:: bash

      hack/diff-critical-files.sh | less -R

#. Ensure that *existing* database migration files have not been
   changed. All database schema changes must be made with new .sql
   files, including rollback files.

   The following paths have database migration files.

   #. go/pkg/billing/db/migrations
   #. go/pkg/cloudaccount/sql
   #. go/pkg/compute_api_server/db/migrations
   #. go/pkg/metering/db/migration

#. Ensure that changes made to GRPC protobuf files are made in a compatible way.
   In addition to being used for RPCs, the JSON encoding of some
   Protobuf messages is used in the Compute API Server database.
   Therefore, changes should be backward and forward compatible.

   The following paths have GRPC protobuf files.

   #. https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/public_api/proto

   Compatible changes:

   #. Add a field to a message, assuming that the zero value is handled properly.
   #. Add an RPC
   #. Add a service

   Changes that may break compatibility and should be avoided or carefully reviewed:

   #. Change a field name, ID or data type
   #. Remove a field
   #. Change an RPC
   #. Remove an RPC
   #. Rename a message

   Refer to https://protobuf.dev/programming-guides/proto3/ and
   https://protobuf.dev/programming-guides/dos-donts/ for details.

#. Ensure that changes made to K8s APIs (custom resource definitions)
   are made in a compatible way. Changes should be backward and forward
   compatible.

   The following paths have K8s APIs:

   #. go/pkg/k8s/apis/private.cloud

   Compatible changes:

   #. Add a non-required field to a type, assuming that the zero value
      is handled properly. A non-required field contains the attribute
      "omitempty".
   #. Add a new type

#. Consider the effects of new code reading data persisted by older
   code.

#. Consider the effects of old code reading data persisted by newer
   code. This could be realized if an upgraded system is rolled back
   after some usage.

#. Identify changes to the vault/configure.sh and vault/load-secrets.sh scripts.

#. Create release notes.

Generate Argo CD Manifests
==========================

This procedure uses :ref:`universe_deployer`.

#. Checkout the **main** branch of `IDC monorepo`_.

#. Open the Universe Config file for the desired environment in your editor.

   #. `Universe Config prod`_

      .. warning::
         You must not merge changes to prod.json unless you have started the RFC implementation.

   #. `Universe Config staging`_

      .. warning::
         You must not merge changes to staging.json until you are ready for the changes to be deployed immediately.

#. (Optional) View the list of commits deployed for each environment, region, availability zone, and component.

   .. code-block:: bash

      make run-universe-config-print

   You may also export a CSV file with more detailed information about each commit.

   .. code-block:: bash

      make universe-config-csv

#. Identify the set of component instances that you want to change.
   For example, you may want to update the production us-region-1 instance of computeApiServer.

   The component corresponds to a set of Helm releases.
   Each Helm release has a label named "component".
   These labels are defined in the files ``deployment/helmfile/helmfile-*.yaml``.

#. Identify the commit hash for each component instance you want to change.

   The commit hash for a component defines the precise and reproducible set of container images, Helm charts, and Helm release values.
   All binaries and configuration related to a component will be built from the specified commit for that component.

   Use *one* of the following processes.

   *  **Start rollout**

      The first time a commit in the main branch is deployed to any environment, it is considered the start of a rollout.
      This should only occur in a staging or other non-production environment.
      You may choose any commit in the main branch for the start of a rollout.
      Generally, you will choose the latest commit in the main branch and apply it only to a single region (us-staging-1).

   *  **Extend rollout**

      Once the result of the start of a rollout has been tested successfully, the rollout will be generally be extended to other
      staging regions, and then to production regions.
      This can generally be accomplished by copying and pasting the relevant commits in staging.json and prod.json.

      .. warning::
         If the rollout to a region requires a change to a configuration source file (e.g. /deployment/helmfile/environments)
         which is not in the existing commit, you cannot use this **extend rollout** process.
         Instead use the **hotfix** process.

   *  **Hotfix**

      In some cases you may want to make a limited change to a deployed component.
      This may be only a change to configuration, only a change to a binary, or both.
      In any case, this is referred to as a *hotfix*.

      Use the following steps to create a new commit for a hotfix.

      #. As a prerequisite to this whole procedure, follow the standard process to update the main branch with your fixes.
         If you fail to do this, then your hotfixes may be overwritten during the next rollout.

      #. Checkout the commit that is currently deployed and referenced in prod.json.

         For example:

         .. code-block:: bash

            git checkout dc4e58d9cdad4fce4a763e08b08531d7f539b375

      #. Create a new branch for your hotfix.

         For example:

         .. code-block:: bash

            git branch hotfix-change-compute-quotas
            git checkout hotfix-change-compute-quotas

      #. Cherry-pick the changes that should be included in your hotfix.

         For example:

         .. code-block:: bash

            git cherry-pick 8e3a8fff0b187c6ed92cebe479e47265ba8e8ac8

      #. Push the new branch to Github.

         For example:

         .. code-block:: bash

            git push --set-upstream origin hotfix-change-compute-quotas

         Do not create a PR for this branch since it will not be merged to main.

      #. Get the commit hash of the new branch.
         You will use this commit hash in the next step.

         .. code-block:: bash

            git rev-parse HEAD

      #. Checkout the **main** branch of `IDC monorepo`_.

         .. code-block:: bash

            git checkout main

   *  **New component instance**

      A new component instance can be added by adding the component to staging.json within the correct heirarchy (global, regional, or AZ).

   *  **Development**

      For development environments, including where a staging environment is used for development of some components,
      you may choose to upgrade and deploy a component using a short cut that requires only one PR approval in the `IDC monorepo`_.

      For this short cut, your branch should include a set of commits with your code or configuration changes.
      The last commit in your branch should update staging.json with the second-to-last commit hash in your branch,
      which will include your code or configuration changes.

      When your branch gets merged to main, these changes will be deployed.

      Since this commit is not in the main branch, it should not be deployed to production.

#. Identify the commit hash for the **grpc** component.

   .. warning::
      The **grpc** component commit must be in the main branch.

   The **grpc** component is shared between all gRPC services in the same global environment or region.
   It consists of grpc-rest-gateway, grpc-proxy-external, grpc-proxy-internal, and grpc-reflect Helm releases.
   All of these are dependent on the Protobuf files in ``/public_api/proto``.

   The **grpc** component should be at the latest commit among all gRPC services in the
   same global environment or region.
   To ensure that there is an obvious "latest" commit, you must ensure that the **grpc** component
   commit is in the main branch. It does not need to be the last commit.

   When IDC services makes gRPC changes that are forward and backward compatible, the **grpc** component can be safely
   upgraded to a later commit.
   However any downgrades must be carefully analyzed and tested.

   If you absolutely must perform development of a gRPC service in a staging region (*not recommended*),
   then you must merge your ``/public_api/proto`` changes (including changes from ``make generate``)
   into main and use this commit for the **grpc** component.
   The commit for the component that implements your service can be from any branch,
   or you may also choose to not implement the service or RPC.

#. Create a draft pull request (PR) to merge the updated Universe Config file to the
   **main** branch of `IDC monorepo`_.

   **DO NOT MERGE THIS PR YET!**

#. Jenkins will immediately start the CI/CD pipeline.

   The **Bazel Universe Deployer** stage will perform the following steps **automatically**.
   Wait for this stage to complete.

   #. Find unique commits in all Universe Config files.

   #. For each unique commit:

      #. Ensure that the commit has been authorized to deploy to the target environments.
         (Method TBD.)

      #. Clone `IDC monorepo`_ at the specific commit.

      #. Build containers and Helm charts.

      #. Build Argo CD manifests (config.json and Helm values.yaml).

      #. Push containers and Helm charts to Harbor. 
         Commits in `Universe Config prod`_ will be pushed to production Harbor.
         Commits in `Universe Config staging`_ will be pushed to staging Harbor.

      #. Combine Argo CD manifests from each commit.
         The combined result will be 100% authoritative and declarative for all IDC services in all production and staging clusters.
         It will reflect upgraded Helm releases, updated containers, updated values.yaml, new Helm releases, deleted Helm releases,
         and renamed Helm releases.

   #. Since this is not running in the **main** branch, this runs in dry run mode.
      It will not make any changes to the `idc-argocd`_ repository.

#. Review the log of the **Bazel Universe Deployer** stage in Jenkins.
   Search for "git diff BEGIN" and confirm that all changes are expected.

   #. Changed files in ``applications/idc-global-services/idc-global`` and ``applications/idc-region/us-region-*`` will affect production.
      A PR with a production change should only be merged after starting the RFC implementation.
      However, if only the value for `gitCommit` changes, the associated pods will not restart and this is safe to apply at any time
      without an RFC.

   #. A PR with a staging change should be merged only when you are ready for the changes to be deployed.
      If the change may affect other teams, announce the upcoming change in the "Staging rebuild" Teams chat.

   #. Pay particular attention to any Helm releases that are deleted or renamed.
      Depending on the configuration of the Argo CD ApplicationSet and Applications, deleted or renamed Helm releases
      may cause the associated Kubernetes resources to be permanently deleted.

Create Test Instances
=====================

Create a VM instance and a BM instance. Use instance names such as
"tiny-before-upgrade-1".

Begin Maintenance Window
************************

If your PR changes `Universe Config prod`_, stop here and wait for the RFC implementation to start.

Update Secrets
**************

If your PR requires changes to secrets in Vault, apply the changes now.

Deploy Helm Releases Using Argo CD
**********************************

#. Obtain approval of your PR, then merge it into the **main** branch of the `IDC monorepo`_.

   .. warning::
      Once changes to `Universe Config prod`_ or `Universe Config staging`_ are merged to main,
      the automatic process to update the environment will begin.
      There are no additional approvals nor manual steps in the process.
      Any updated pods will be restarted.

#. Jenkins will immediately start the CI/CD pipeline on the new commit in the **main** branch of the `IDC monorepo`_.

   The **Bazel Universe Deployer** stage will perform the same steps listed previously.
   Additionally, it will add a new commit to the **main** branch of `idc-argocd`_ with the combined Argo CD manifests.

#. Argo CD will automatically deploy the Helm releases.

#. You may monitor Argo CD through the Web UI. See `idc-argocd`_.

.. _update-product-catalog-definitions:

Update Product Catalog Definitions
**********************************

All product catalog definitions (Custom Resources) are managed through following repo:

https://github.com/intel-innersource/frameworks.cloud.devcloud.services.product-catalog

You must apply these definitions manually using the steps below:

1. Clone the catalog definition repo locally.

   .. code:: bash

      git clone https://github.com/intel-innersource/frameworks.cloud.devcloud.services.product-catalog

2. Make sure you have kubeconfig setup properly to target cluster.

3. Apply product specs.

   .. code:: bash

      cd prod
      kubectl apply -f products/ -n idcs-system

4. Verify product specs on your cluster

   .. code:: bash

      kubectl get products.private.cloud.intel.com -n idcs-system 
      NAME                   AGE
      bm-spr                 41d
      vm-spr-lrg             41d
      vm-spr-med             41d
      vm-spr-sml             41d
      vm-spr-tny             29d

Deploy Console to S3 Bucket
***************************

Build the console and copy to `Console S3
Bucket <https://s3.console.aws.amazon.com/s3/buckets/s3-idc-console-production-spa?region=us-west-2&tab=objects>`__.

Smoke Test
**********

Create, SSH, and delete VM instance. Create, SSH, and delete BM
instance.

To Create VM instance, SSH and Delete
   Pipeline link: https://internal-placeholder.com/satg-dcp-dcbmaas/job/prod-compute-smoke/
   
   .. code:: bash
   
      Steps to execute VM smoke test:
      -------------------------------
      1) Build the pipeline (Build with parameters)
      2) Provide/Select the following input before running the pipeline
         a) Cloudaccount (use the default cloud account - recommended), 
         b) admin-token (fetch the token from admin console), 
         c) product-type - VM,
         d) instance-type can be any one of the following (vm-spr-sml(recommended for testing), vm-spr-med, vm-spr-lrg), 
         e) machine-image can be any one of the following (ubuntu-2204-jammy-v20230122, ubuntu-2204-jammy-v202403082), 
         f) vnet (If different cloud-account is chosen, provide the vnet present inside the cloud account)  and 
         g) region.
      3) Run the pipeline

To Create BM instance, SSH and Delete
   Pipeline link: https://internal-placeholder.com/satg-dcp-dcbmaas/job/prod-compute-smoke/

   .. code:: bash
   
      Steps to execute BM smoke test:
      -------------------------------
      1) Build the pipeline (Build with parameters)
      2) Provide/Select the following input before running the pipeline
         a) Cloudaccount (use the default cloud account - recommended), 
         b) admin-token (fetch the token from admin console), 
         c) product-type - BM,
         d) instance-type (For testing purpose - Use PVC (bm-spr-pvc-1100-8) or SPR (bm-spr) based on the availability of hardware in production), 
         e) machine-image can be any one of the following (ubuntu-22.04-pvc-metal-cloudimg-amd64-v20240319(pvc), ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115(spr)), 
         f) vnet (If different cloud-account is chosen, provide the vnet present inside the cloud account)  and 
         g) region.
      3) Run the pipeline

Final Steps
***********

#. Delete debug-tools and postgres-client pods.

Rollback Procedure
******************

#. Revert the change to `Universe Config prod`_ or `Universe Config staging`_.

#. Changes to the SQL schemas are not rolled back automatically.

   Use the steps below to manually rollback the billing DB.

   Connect to billing DB as shown in
   https://internal-placeholder.com/display/devcloud/IDC+Environment+dev3#IDCEnvironmentdev3-HowtoAccessGlobalDatabases.

   .. code:: bash

      root@postgres-client:/# export PGHOST=dev-idc-global-postgresqlv2.cluster-cb6hxdt0onur.us-west-2.rds.amazonaws.com
      root@postgres-client:/# psql
      Password for user billing_user: 

      billing=>
      drop table notifications;
      drop table alerts;
      update schema_migrations set version=20230511043259;



.. _IDC monorepo: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc
.. _Universe Config prod: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/universe_deployer/environments/prod.json
.. _Universe Config staging: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/universe_deployer/environments/staging.json
.. _idc-argocd: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc-argocd
.. _IDC Environments Master List: https://internal-placeholder.com/x/uyLhs
