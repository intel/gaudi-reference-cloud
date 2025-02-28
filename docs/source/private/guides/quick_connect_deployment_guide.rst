.. _quick_connect_deployment_guide:

Quick Connect Deployment Guide
##############################

Vault
-----

Add the following blocks to the deployment/common/vault/terraform/data files for the appropriate region.  Examples shown below are us-staging-1a.

approles_region_1.tfvars:

.. code-block:: terraform

  "us-staging-1a-quick-connect-api-server-role" = {
    # bind_secret_id =
    # local_secret_ids =
    # policies = policies.tfvars
    secret_id_bound_cidrs = []
    # secret_id_num_uses = 0
    # secret_id_ttl = 0
    token_bound_cidrs = []
    # token_explicit_max_ttl = 0
    token_max_ttl = "86400"
    # token_no_default_policy = false
    # token_num_uses = 0
    token_period   = 0
    token_policies = ["us-staging-1a-quick-connect-api-server-policy"]
    token_ttl      = 72000
    # token_type = default
  }

jwt_roles_region_1.tfvars:

.. code-block:: terraform

  "us-staging-1a-quick-connect-api-server-role" = {
    "role_name"       = "us-staging-1a-quick-connect-api-server-role"
    "backend"         = "cluster-auth"
    "token_policies"  = ["global-pki", "public", "us-staging-1a-quick-connect-api-server-policy"]
    "token_ttl"       = 3600
    "bound_subject"   = "system:serviceaccount:idcs-system:us-staging-1a-quick-connect-api-server"
    "bound_audiences" = ["https://kubernetes.default.svc.cluster.local"]
  }

pki_mounts_region_1.tfvars:

.. code-block:: terraform

  "us-staging-1a-quick-connect-client-ca" = {
    "path" = "us-staging-1a-quick-connect-client-ca"
  }

pki_roles_region_1.tfvars:

.. code-block:: terraform

  "us-staging-1a-quick-connect-api-server" = {
    "backend"                     = "us-staging-1a-ca"
    "role_name"                   = "us-staging-1a-quick-connect-api-server"
    "allowed_domains"             = ["us-staging-1a-quick-connect-api-server.idcs-system.svc.cluster.local", ".local", "*.internal-placeholder.com", "*.eglb.intel.com", "*.internal-placeholder.com", "*.internal-placeholder.com"]
    "allow_glob_domains"          = true
    "allow_bare_domains"          = true
    "ou"                          = ["us-staging-1a-quick-connect-api-server"]
    "allow_wildcard_certificates" = false
    "enforce_hostnames"           = true
    "allow_any_name"              = false
    "key_bits"                    = 2048
  }

  "us-staging-1a-quick-connect-client" = {
    "backend"                     = "us-staging-1a-quick-connect-client-ca"
    "role_name"                   = "us-staging-1a-quick-connect-client"
    "allowed_domains"             = ["us-staging-1a-quick-connect-client.idcs-system.svc.cluster.local"]
    "allow_glob_domains"          = true
    "allow_bare_domains"          = true
    "ou"                          = ["us-staging-1a-quick-connect-client"]
    "allow_wildcard_certificates" = false
    "enforce_hostnames"           = true
    "allow_any_name"              = false
    "key_bits"                    = 3072
  }

policies_region_1.tfvars:

.. code-block:: terraform

    "us-staging-1a-quick-connect-api-server-policy" = {
      policy = <<EOT
  path "controlplane/data/us-staging-1a-quick-connect-api-server/*" {
    capabilities = ["read", "list"]
  }
  path "controlplane/data/us-staging-1a/quick-connect-api-server/*" {
    capabilities = ["read", "list"]
  }
  path "controlplane/data/+/us-staging-1a/quick-connect-api-server/*" {
    capabilities = ["read", "list"]
  }
  path "us-staging-1a-ca/issue/us-staging-1a-quick-connect-api-server" {
    capabilities = ["update"]
  }
  path "us-staging-1a-quick-connect-client-ca/issue/us-staging-1a-quick-connect-client" {
    capabilities = ["update"]
  }
  EOT
    }

And add to each instance-operator policy (us-staging-1a-bm-instance-operator-policy, us-staging-1a-vm-instance-operator-harvester1-policy, etc.):

.. code-block:: terraform

    "us-staging-1a-bm-instance-operator-policy" = {
      policy = <<EOT
      # ...
  path "us-staging-1a-quick-connect-client-ca/cert/ca_chain" {
    capabilities = ["read"]
  }
  EOT
    }

NOTE: The new mount above requires additional steps after Vault terraform apply.

Register Callback Endpoint in Azure
-----------------------------------

NOTE: This has already been completed for existing regions as of 2024-01-14.

#. Login to https://portal.azure.com/intelcorpb2c.onmicrosoft.com

#. App registrations

#. IDC B2C Quick connect - Pre-production or IDC B2C Quick connect - Production

#. Redirect URIs

#. Add URI: enter new callback URI of form https://callback.connect.${REGION}.devcloudtenant.io/v1/callback

#. Save

Create Client Secret in Azure
-----------------------------

NOTE: This has already been completed for pre-production and production.

#. Login to https://portal.azure.com/intelcorpb2c.onmicrosoft.com

#. App registrations

#. IDC B2C Quick connect - Pre-production or IDC B2C Quick connect - Production

#. Client credentials

   The description should be pre-production or production, and expiration of 24 months.

Record approle credentials and created secret in Vault
------------------------------------------------------

The oauth2_client token and hmac values are can be obtained from existing deployments in Vault or in Azure.
Contact todd.malsbary@intel.com to obtain the values from Azure.


#. Record the approle credentials in Vault.

   .. code-block:: shell

      export VAULT_ADDR=https://internal-placeholder.com/
      export VAULT_TOKEN= # Obtain via Vault UI

      ROLE_ID=$(vault read auth/approle/role/us-staging-1a-quick-connect-api-server-role/role-id -format=json | jq -r .data.role_id)
      SECRET_ID=$(vault write -f auth/approle/role/us-staging-1a-quick-connect-api-server-role/secret-id -format=json | jq -r .data.secret_id)
      vault kv put -mount=controlplane us-staging-1a-quick-connect-api-server/approle secret_id=${SECRET_ID} role_id=${ROLE_ID}
      vault kv put -mount=controlplane us-staging-1a-quick-connect-api-server/oauth2_client token="..." hmac="..."

K8s Ingress Certificate Secret
------------------------------

Ensure the wildcard-devcloudtenant-tls Secret is deployed in the azqc cluster.

.. code-block:: shell

  $ kubectl get secret -n idcs-system wildcard-devcloudtenant-tls
  NAME                          TYPE                DATA   AGE
  wildcard-devcloudtenant-tls   kubernetes.io/tls   2      103d

Helmfile
--------

- In prod.yaml.gotmpl, set components.quickConnect.enabled: true

- In prod-region-us-region-3.yaml.gotmpl
  set enableQuickConnectClientCA: true for bmInstanceOperator and vmInstanceOperator 
  set quickConnect.enabled: true

- Add quickConnect component to prod.json

Deploy Service
--------------

Follow :ref:`services_upgrade_procedure` to deploy Quick Connect Service.  The quickConnect, computeVmInstanceOperator, and computeBmInstanceOperator components will need to be deployed.

Smoke Test
----------

#. Launch a VM compute instance with One-Click Connection enabled

#. Click the Connect button when available and confirm connection to JupyterLab

