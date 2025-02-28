.. _services_upgrade_procedure_development:

IDC Services Upgrade Procedure
##############################

This document describes how to *upgrade* IDC services in *development* environments such as dev20 through dev39.

Related Documentation
*********************

* To install IDC services in a *new* development environment, see :ref:`services_deployment_procedure_development`.

* `IDC Environments Master List`_

How to Run the Deployer
***********************

#. Identify an existing branch or create a new branch in the `IDC monorepo`_.
   Ensure that you have pushed your local branch to Github.

#. Open `Flex Deployer`_ and enter the desired parameters.

   #. **IDC_ENV**:
      Select the IDC environment. 
      Ensure you have the permission of the environment owner identified in `IDC Environments Master List`_.

   #. **GIT_BRANCH**:
      Enter the Git branch or tag you wish to deploy.

   #. **DELETE_ALL_ARGO_APPLICATIONS**:
      If checked, delete (uninstall) all Argo CD Applications such as billing, billing-db, compute-api-server, compute-crds, etc..
      Check this box to completely delete all Postgres databases.
      This only applies to environments where this deployer deploys Argo CD, 
      which is where ``argocd.enabled`` is ``true`` in ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl``.

   #. **DELETE_ARGO_CD**:
      If checked, delete (uninstall) Argo CD (``argocd-server``, ``argocd-repo-server``, etc.).
      This only applies to environments where this deployer deploys Argo CD, 
      which is where ``argocd.enabled`` is ``true`` in ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl``.

   #. **DELETE_GITEA**:
      If checked, delete (uninstall) Gitea which is a dedicated Git repository for this environment used by Argo CD.
      This only applies to environments where this deployer deploys Argo CD, 
      which is where ``argocd.enabled`` is ``true`` in ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl``.

   #. **DELETE_VAULT**:
      If checked, delete (uninstall) Vault Server and Vault Agent Injector.
      This only applies to environments where this deployer deploys Vault Server, 
      which is where ``vault.enabled`` and ``global.vault.server.enabled`` are ``true`` in ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl``.

   #. **INCLUDE_DEPLOY**:
      This should usually be checked.
      To only uninstall, check desired DELETE_* flags and uncheck this flag.

   #. **INCLUDE_VAULT_CONFIGURE**:
      If checked, configure Vault roles, policies, and PKI.
      If you have manually changed the Vault configuration, this may overwrite your changes.

   #. **INCLUDE_VAULT_LOAD_SECRETS**:
      If checked, load secrets into Vault.
      If you have manually changed secrets in Vault, this may overwrite your changes.

   #. **PUSH_DEPLOYMENT_ARTIFACTS**:
      If checked, push the deployment artifacts (Helm charts and container images) to Harbor.
      This is always safe to do but if you know that the necessary artifacts have already been pushed,
      you can save some time by unchecking this.
      You'll know when this is required because Argo CD will be unable to download the chart or the pod will be unable to download the image.

#. Click Build.

How to quickly change Helm release values
*****************************************

The deployment procedure described above can deploy nearly any change automatically.
However, it can be too time consuming for some iterative development.
This procedure can be used to make changes to Helm release values within a few seconds.

#. Configure a port forward for Gitea.

   .. code:: bash

      export KUBECONFIG=...
      kubectl port-forward --namespace gitea service/gitea-http 50313:http &

#. Clone the Git repo.
   If you have already done this, you can skip this step.

   .. code:: bash

      GITEA_USERNAME=$(kubectl get secret --namespace argocd git-repo-creds -o jsonpath="{.data.username}" | base64 --decode)
      GITEA_PASSWORD=$(kubectl get secret --namespace argocd git-repo-creds -o jsonpath="{.data.password}" | base64 --decode)
      IDC_ENV=...
      cd
      git clone http://${GITEA_USERNAME}:${GITEA_PASSWORD}@localhost:50313/${GITEA_USERNAME}/idc-argocd.git ${IDC_ENV}-idc-argocd

#. Edit any file in this local repository.
   For example, to change the Helm values for billing-schedulers, edit
   ``~/${IDC_ENV}-idc-argocd/applications/idc-global-services/dev27-global/idc-dev-27/billing-schedulers/values.yaml``.

#. Commit and push your changes.

   .. code:: bash

      git add -A && \
      git commit -m "manual update" && \
      git push

#. Your changes should be detected by Argo CD within 10 seconds.

#. Repeat the last 3 steps as desired.

#. To avoid your changes from being replaced during the next deployment, be sure to commit your changes to `IDC monorepo`_.

Troubleshooting
***************

A pod or other Kubernetes resource is not getting deployed as expected
======================================================================

#. Identify the file ``/deployment/helmfile/helmfile-*.yaml`` that defines the Helm release that creates the Kubernetes resource.
   The name of this file is the **component** that is required.

#. Ensure that the Universe Config has this component enabled.

   #. If the file ``/universe_config/environment/${IDC_ENV}.json`` exists, then ensure that it includes the component in the correct sections.
      There is a different section for global, regional, and availability zone components.

   #. Otherwise, the Universe Config will come from ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl`` with defaults from ``/deployment/helmfile/defaults.yaml.gotmpl``.

      #. Global components are defined in ``global.components``.

      #. Regional components are defined in ``regions.${REGION}.components`` with defaults in ``defaults.region.components``.

      #. Availability zone components are defined in ``regions.${REGION}.availabilityZone.components`` with defaults in ``defaults.availabilityZone.components``.

      #. The component must have ``enabled`` set to ``true``.
         Keep in mind that some components are enabled by default and others are disabled by default.
         Only enable a component in ``defaults.yaml.gotmpl`` if it is essential for IDC to function.
         Otherwise, enable desired components in the environment-specific file ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl``.

   #. In general, ``commit`` should be set to ``HEAD`` to indicate that the deployer should use the Git branch or tag specified in ``GIT_BRANCH``.

#. Carefully review the ``{{ if ... }}`` sections in ``/deployment/helmfile/helmfile-*.yaml`` to identify all values that must be set
   in order to include the Helm release.
   Any such values must be defined in ``/deployment/helmfile/environments/${IDC_ENV}.yaml.gotmpl`` with defaults from ``/deployment/helmfile/defaults.yaml.gotmpl``.

   If you see an error such as shown below, this usually indicates that the ``{{ if ... }}`` conditions are not satisfied.

   .. code::

      err: no releases found that matches specified selector(geographicScope=global,component=telemetry) and environment(test-e2e-compute-bm), in any helmfile

#. To view the effective values used to render ``/deployment/helmfile/helmfile-*.yaml``,
   view the build artifacts for the Jenkins deployer pipeline, then choose the file ``local/build-artifacts/helmfile-dump.yaml``.
   The URL will be similar to https://internal-placeholder.com/satg-dcp-dcbmaas/job/Deployment/job/flex-deployer/441/artifact/local/build-artifacts/helmfile-dump.yaml.
   When reviewing this file to understand the effective values, ignore the values under the ``Values.defaults`` key.

Getting Help
************

If you need help with a deployment pipeline, send a message to the ``IDC Developer Support`` MS Teams chat.
At a minimum, include the URL to the Jenkins pipeline build.
The URL will be similar to https://internal-placeholder.com/satg-dcp-dcbmaas/job/Deployment/job/flex-deployer/441/.

See Also
********

* :ref:`deploy_all_in_k8s`
* :ref:`create_new_idc_service`



.. _IDC monorepo: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc
.. _IDC Environments Master List: https://internal-placeholder.com/x/uyLhs
.. _Flex Deployer: https://internal-placeholder.com/satg-dcp-dcbmaas/job/Deployment/job/flex-deployer/build
