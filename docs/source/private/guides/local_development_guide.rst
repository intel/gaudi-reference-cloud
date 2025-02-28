.. _local_development_guide:

Local Development Guide
#######################

This guide can be used by Intel developers of IDC components to learn how to build and test IDC using a single development workstation.
External dependencies are minimized.

Prepare Development Workstation
-------------------------------

#. Follow the procedure in `Prepare Development Environment for IDC <https://internal-placeholder.com/x/8wmUlw>`__.

#. Run the following steps:

   .. code:: bash

      sudo apt install make unzip python3-pip
      make install-interactive-tools

#. Install additional tools.

   -  kind (use the version in /build/repositories/repositories.bzl)
   -  kubectl (use the version in /build/repositories/repositories.bzl)
   -  Go (use the Go toolchain version in /WORKSPACE)

#. `Configure k8s for large environments <https://github.com/kubeflow/manifests/issues/2087>`__

   .. code:: bash

      echo "fs.inotify.max_user_instances=1280" | sudo tee -a /etc/sysctl.d/idc.conf
      echo "fs.inotify.max_user_watches=655360" | sudo tee -a /etc/sysctl.d/idc.conf
      sudo sysctl --system

#. Configure docker to use internal mirror for images.

   Pulling images directly from docker is not recommended as it is very likely to fail because of rate limitations imposed by docker. See https://intel.sharepoint.com/sites/caascustomercommunity/sitepages/dockerhubcache.aspx?web=1#registry-mirror

   It is best practice to configure docker to use a mirror. You can do this by adding a registry-mirror to  /etc/docker/daemon.json on your host. For example

   .. code:: bash

      {"registry-mirrors": ["https://internal-placeholder.com"]}


   example steps:

   .. code:: bash

      sudo vi /etc/docker/daemon.json
      # add the line '{"registry-mirrors": ["https://internal-placeholder.com"]}'

      # save file

      sudo systemctl restart docker

Building
--------

To build all containers and Helm charts, run:

.. code:: bash

   make build

Running Tests
-------------

To run all unit tests and most integration tests, run the command:

.. code:: bash

   make test

Maintaining BUILD.bazel Files
-----------------------------

#. After making any changes to imports in Go (.go) files, run the command below to create or update BUILD.bazel files with
   ``go_binary``, ``go_test``, and other Bazel rules.

.. code:: bash

   make gazelle

.. _updating_go_dependencies:

Updating Go Dependencies
------------------------

#. Build the Go SDK and add it to the path.

   .. code:: bash

      eval `make go-sdk-export`
      go version
      cd go

#. Update a dependency in ``go.mod`` to the latest version with the following command:

   .. code:: bash

      go get example.com/pkg

   To use a specific version:

   .. code:: bash

      go get example.com/pkg@v1.2.3

#. Run Go Tidy to make sure go.mod matches the source code.

   .. code:: bash

      go mod tidy

#. Sometimes, the removal of a direct dependency will result in indirect dependencies getting downgraded.
   If this occurs, add a reference to the package in ``/go/pkg/force_import/main.go``.
   By adding a direct reference, ``go mod tidy`` will respect the minimal version in ``go.mod``.

   .. code:: golang

      import (
	      _ "example.com/pkg"
      )

#. Run Go Vet to examine the source code.

   .. code:: bash

      go vet ./svc/cloudaccount/...
      go vet ./...

#. Update the Bazel dependency list ``deps.bzl``.

   .. code:: bash

      cd ..
      make gazelle

#. Review the changes to ``deps.bzl`` to ensure that dependencies are not downgraded.

   If there are many changes, you may want to use ``make go-list`` and the script ``hack/go-mod-downgraded.sh``
   to automatically detect modules where the semver decreased.
   Note that this script only properly identifies versions in ``x.y.z`` format.
   If the version is in a different format, carefully review the output.
   Follow the steps below.

   .. code:: bash

      git checkout main
      make go-list > local/go-list-main.txt
      git checkout your-branch
      make go-list > local/go-list.txt
      hack/go-mod-downgraded.sh

#. Ensure that everything can be built and tests are successful.

   .. code:: bash

      make generate build test

Generating Code
---------------

After updating ``public_api/proto/*.proto``, ``go/svc/*/*.templ``, or other sources of generated code, run:

.. code:: bash

   make generate

If you only made changes to Protobuf (.proto) files, you can run just a subset of the generation process with:

.. code:: bash

   make generate-go

Then commit any changed files.
The Jenkins job "Check generated files" will fail if generated files have not been checked in.

.. _deploy_idc_core_services_in_local_kind_cluster:

Deploying IDC Core Services in a Local kind Cluster
---------------------------------------------------

Most IDC services can be deployed to a development workstation using `kind <https://sigs.k8s.io/kind>`__.
This environment can be used for iterative development of IDC services.

To build core components, deploy a new local kind cluster, and deploy core components, run:

.. code:: bash

   make deploy-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/deploy-all-in-kind-v2.log

To enable verbose logs, run ``export ZAP_LOG_LEVEL=-127`` before running the previous command.

To run core end-to-end tests.

.. code:: bash

   export no_proxy=${no_proxy},.local
   source go/pkg/tools/oidc/test-scripts/get_token.sh
   go/svc/cloudaccount/test-scripts/cloud_account_create.sh
   export CLOUDACCOUNT=$(go/svc/cloudaccount/test-scripts/cloud_account_get_by_name.sh | jq -r .id)
   go/svc/compute_api_server/test-scripts/vnet_create_with_name.sh
   go/svc/compute_api_server/test-scripts/sshpublickey_create_with_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh

.. _upgrade_services_in_local_kind_cluster:

Upgrading Services in a Local kind Cluster
------------------------------------------

Anytime after running ``make deploy-all-in-kind-v2``, you can modify the source code of any service and upgrade the service
running in kind using the steps below.

.. code:: bash

   make upgrade-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/upgrade-all-in-kind-v2.log

In some cases, you may want to completely uninstall an IDC service and then reinstall it.
This is particularly useful when iterating on changes to a database schema.
The environment variable ``DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE`` can be set to any
regular expression that matches any number of Helm release names.

.. code:: bash

   DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE=".*-compute-db|.*-compute-api-server" \
   make upgrade-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/upgrade-all-in-kind-v2.log

Alternatively, you can checkout a different commit before running ``make deploy-all-in-kind-v2``, then checkout
your latest commit to manually test an upgrade.

Deploy All In Kind V2 Overview
------------------------------

When ``make deploy-all-in-kind-v2`` is executed, the following occurs.

#. Generate random secrets in ``local/secrets`` if needed. Existing secrets are unchanged.

#. Deploy a Docker registry as a Docker container. This will be used by containers and OCI Helm charts.

#. Run the Go application `deploy_all_in_kind <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/go/pkg/universe_deployer/cmd/deploy_all_in_kind/main.go>`_.
   This performs the following.

   #. Run Bazel to build deployment artifacts (see :ref:`deployment_artifacts`).

   #. Run Bazel to push container images and Helm charts to the local Docker registry.

   #. Generate Argo CD manifests which define the Helm releases that will be deployed.

   #. Start a kind cluster.

   #. Deploy CoreDNS, Vault, and Gitea.
   
   #. Push Argo CD manifests to a repo in Gitea.
   
   #. Deploy Argo CD and configure it to watch the repo in Gitea.

   #. Wait for Argo CD to deploy IDC services.

Enable VMaaS in a Local kind Cluster
------------------------------------

To enable VMaaS in a local kind cluster, follow the steps in this section.

Obtain Harvester KubeConfigs
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you do not have a valid Harvester KubeConfig, the VM Instance
Operator will fail to start. Obtain the KubeConfig file using the
steps below.

Method 1
^^^^^^^^

#. Download the KubeConfig from `Vault <https://internal-placeholder.com/ui/vault/secrets/dev-idc-env/kv/shared%2Fharvester1%2Fkubeconfig/details?version=1>`__.

#. Save the file to ``local/secrets/harvester-kubeconfig/harvester1``.

Method 2
^^^^^^^^

#. If your development workstation is connected to the Intel corporate network:

   Login to `Harvester1 <https://10.165.57.245>`__.

#. Click *Support* in the bottom-left corner.

#. Click *Download KubeConfig*.

#. Save the file to ``local/secrets/harvester-kubeconfig/harvester1``.

Obtain Host Public Key (RSA)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Both SSH Proxy Operator and BM Instance Operator needs the public key of
the SSH Proxy Server to verify it before establishing a connection.

Obtain and update the host public key secret using the following command:

.. code:: bash

   ssh-keyscan -t rsa ${HOST_IP} | awk '{print $2, $3}'> local/secrets/ssh-proxy-operator/host_public_key

**NOTE**: Here **HOST_IP** corresponds to IP address of the bastion server or
SSH proxy server through which user will be connecting to the reserved
instances.

Deploying IDC Core and VMaaS Services in a Local kind Cluster
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Perform the steps in :ref:`deploy_idc_core_services_in_local_kind_cluster`.

After running the core end-to-end tests, run the following additional steps.

.. code:: bash

   go/svc/compute_api_server/test-scripts/instance_create_with_name.sh
   watch go/svc/compute_api_server/test-scripts/instance_list_summary.sh
   ssh -J guest-${USER}@10.165.62.252 ubuntu@172.16.x.x
   go/svc/compute_api_server/test-scripts/instance_delete_by_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh

Enable BMaaS in a Local kind Cluster
------------------------------------

Enable NGINX S3 Gateway and configure BMaaS to use it (optional)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This step is optional and is primarily intended for BMaaS developers who
will be requesting BM instances.

If you want baremetal-operator to pull OS images directly from S3
bucket, instead of deploying a dedicated HTTP server you can enable
NGINX S3 Gateway.

To do this, modify your environment Helmfile to include following
section in regional services (or edit `default
settings <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/3c2c15e5a0d94e19faa5f110c4f5fee0f0046ffa/deployment/helmfile/defaults.yaml.gotmpl#L409>`__)

.. code:: yaml

   nginxS3Gateway:
       enabled: true
       s3_bucket_name: {{ env "NGINX_S3_GATEWAY_BUCKET" | default "catalog-fs-dev" }}

You can set S3 bucket name that should be used to pull images directly
in Helmfile, or overwrite it through ``NGINX_S3_GATEWAY_BUCKET`` env var
before deployment.

NGINX S3 Gateway Helm chart will create k8s NodePort service available
on port ``31969``. Next, it’s necessary to configure baremetal-operator
to use this service for pulling images. This can be accomplished by
modifying ``bmInstanceOperator`` configuration in your environment
Helmfile.

.. code:: yaml

   bmInstanceOperator:
       osHttpServerUrl: {{ env "OS_IMAGES_HTTP_URL" | default (printf "http://%s:31969" (requiredEnv "KIND_API_SERVER_ADDRESS")) }}

This configuration is already included in
`bmaas-flex-dev <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/3c2c15e5a0d94e19faa5f110c4f5fee0f0046ffa/deployment/helmfile/environments/bmaas-flex-dev.yaml.gotmpl#L92>`__
environment settings.

The last step is to provide AWS credentials that will be used by NGINX
S3 Gateway for authentication. Save AWS access key ID to
``local/secrets/NGINX_S3_GATEWAY_ACCESS_KEY_ID`` and AWS secret key to
``local/secrets/NGINX_S3_GATEWAY_SECRET_KEY`` before triggering
deployment. Those files will be used to populate Vault secret.

Deploy a new kind cluster, and deploy baremetal operator, baremetal virtual stack
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This deployment setup baremetal-operator running in kind connected to a
virtual baremetal stack all in one single instance. BMaaS developers can
use this setup to develop BM instance operator

baremetal-operator includes the following services - ironic - ironic
inspector - ironic http,tftp serving iPXE, iPXE profiles and ironic
python agent - dhcp server

It will also include a `virtual baremetal stack <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/idcs_domain/bmaas/bmvs>`__ (vBMC + quemu-kvm nodes)
to run this deployment on a baremetal instance(reserve a baremetal
instance in onecloud) Set BMC default credential using
``DEFAULT_BMC_USERNAME`` ``DEFAULT_BMC_PASSWD`` env variables

To enable access to the Ironic installer image, set ssh keys in Vault by
using ``IPA_IMAGE_SSH_PRIV`` and ``IPA_IMAGE_SSH_PUB`` env variables set
to the ssh key files. They will default to /dev/null and will require
manually updating in Vault if not set.

Add Intel certs
~~~~~~~~~~~~~~~

If this is a newly provisioned node, you might need the Intel certs
applied

.. code:: bash

   curl -LO --insecure -s https://internal-placeholder.com/artifactory/it-btrm-local/intel_cacerts/install_intel_cacerts_linux.sh
   chmod +x install_intel_cacerts_linux.sh
   sudo ./install_intel_cacerts_linux.sh
   rm install_intel_cacerts_linux.sh

Set root password
~~~~~~~~~~~~~~~~~

.. code:: bash

   sudo passwd root

Add http_proxy variables to /etc/environment
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

   https_proxy="http://internal-placeholder.com:912"
   http_proxy="http://internal-placeholder.com:912"
   no_proxy="intel.com,.intel.com,10.0.0.0/8,192.168.0.0/16,localhost,127.0.0.0/8,134.134.0.0/16,172.16.0.0/16,192.168.150.0/24,.kind.local"

Setup OS images
~~~~~~~~~~~~~~~

.. code:: bash

   pushd idcs_domain/bmaas/bmvs/playbooks/roles/http_server/files
   wget https://internal-placeholder.com/artifactory/intelcloudservices-or-local/images/ubuntu-22.04-server-cloudimg-amd64-latest.qcow2
   wget https://internal-placeholder.com/artifactory/intelcloudservices-or-local/images/ubuntu-22.04-server-cloudimg-amd64-latest.qcow2.md5sum
   popd

.. code:: bash

   sudo apt install make gcc
   make secrets
   export DEFAULT_BMC_USERNAME=admin
   export DEFAULT_BMC_PASSWD=password
   export SSH_PROXY_IP=$(hostname -f)
   export SSH_USER_PASSWORD=$(uuidgen)
   sudo useradd -m -p $SSH_USER_PASSWORD guest-${USER}
   sudo -u guest-${USER} mkdir /home/guest-${USER}/.ssh
   sudo -u guest-${USER} cp local/secrets/ssh-proxy-operator/id_rsa.pub /home/guest-${USER}/.ssh/authorized_keys
   sudo useradd -m -p $SSH_USER_PASSWORD bmo-${USER}
   sudo -u bmo-${USER} mkdir /home/bmo-${USER}/.ssh
   sudo -u bmo-${USER} cp local/secrets/bm-instance-operator/id_rsa.pub /home/bmo-${USER}/.ssh/authorized_keys
   make install-requirements
   export PATH=/home/${USER}/.local/bin:/usr/local/go/bin:$PATH
   make install-interactive-tools
   sudo iptables -I INPUT -p tcp -m tcp --dport 6443 -j ACCEPT
   sudo iptables -I INPUT -p tcp -m tcp --dport 443 -j ACCEPT
   sudo iptables -I INPUT -p tcp --match multiport --dports 8001,8002,8003,50001 -j ACCEPT
   export IDC_ENV='kind-jenkins'
   make deploy-metal-in-kind

   ###### NOTE: deploy-metal-in-kind creates the KIND cluster with the routable host interface IP as the API server address by default. Use the following command to create a KIND cluster with a specific interface IP address,
   make deploy-metal-in-kind KIND_API_SERVER_ADDRESS=<IP address>
   make deploy-metal-in-kind KIND_API_SERVER_ADDRESS=127.0.0.1

   ###### NOTE: For billing aria driver deployment, `make secrets` will create two files under local/secrets - 1) aria_auth_key 2) aria_client_no  3) aria_api_crt 4) aria_api_key
   - Update the aria_auth_key file using this command (******* -> these values you can get it from your supervisor):
       `echo "*********"> local/secrets/aria_auth_key`
   - Update the aria_client_no file using this command (******* -> these values you can get it from your supervisor):
       `echo "*********"> local/secrets/aria_client_no`
   - Update the aria_api_crt file using this command (******* -> these values you can get it from your supervisor):
       `echo "*********"> local/secrets/aria_api_crt`
   - Update the aria_api_key file using this command (******* -> these values you can get it from your supervisor):
       `echo "*********"> local/secrets/aria_api_key`

Run samples
~~~~~~~~~~~

Create a cloud account
^^^^^^^^^^^^^^^^^^^^^^

.. code:: bash

   export no_proxy=${no_proxy},.kind.local
   export URL_PREFIX=http://dev.oidc.cloud.intel.com.kind.local

   export TOKEN=$(curl "${URL_PREFIX}/token?email=admin@intel.com&groups=IDC.Admin")
   echo ${TOKEN}

   export URL_PREFIX=https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local
   export CLOUDACCOUNTNAME=${USER}@intel.com
   go/svc/cloudaccount/test-scripts/cloud_account_create.sh
   export CLOUDACCOUNT=$(go/svc/cloudaccount/test-scripts/cloud_account_get_by_name.sh | jq -r .id)
   echo $CLOUDACCOUNT

Create a vNet
^^^^^^^^^^^^^

.. code:: bash

   export AZONE=us-dev-1b
   export VNETNAME=us-dev-1b-metal
   go/svc/compute_api_server/test-scripts/vnet_create_with_name.sh

Create an instance
^^^^^^^^^^^^^^^^^^

.. code:: bash

   export NAME=my-metal-instance-1
   export INSTANCE_TYPE=bm-virtual
   export MACHINE_IMAGE=ubuntu-22.04-server-cloudimg-amd64-latest
   go/svc/compute_api_server/test-scripts/sshpublickey_create_with_name.sh
   go/svc/compute_api_server/test-scripts/instance_create_with_name.sh
   go/svc/compute_api_server/test-scripts/instance_list.sh
   go/svc/compute_api_server/test-scripts/instance_get_status.sh
   ssh -J guest-${USER}@$(hostname -f) sdp@172.18.10.x
   go/svc/compute_api_server/test-scripts/instance_delete_by_name.sh

Create a Load Balancer
^^^^^^^^^^^^^^^^^^^^^^

.. code:: bash
   
   export LB_MONITOR=tcp
   export LB_PORT=8080
   export NAME=mylb1

   go/svc/compute_api_server/test-scripts/loadbalancer_create_with_name.sh
   go/svc/compute_api_server/test-scripts/loadbalancer_list.sh  
   go/svc/compute_api_server/test-scripts/loadbalancer_delete_by_name.sh 

Multi-cluster Testing with kind
-------------------------------

For some testing, it may be important to deploy a separate kind cluster for global and regional services.
This uses the original (v1) version of ``make deploy-all-in-kind``.

#. Deploy in kind.

   To test VMaaS only, with a multicluster (1 region) kind environment:

   .. code:: bash

      export IDC_ENV=kind-multicluster
      make show-config
      make deploy-all-in-kind |& ts | ts -i | ts -s | tee local/deploy-all-in-kind-multicluster.log

   To test VMaaS only, with a 2-region kind environment:

   .. code:: bash

      export IDC_ENV=kind-2regions
      make show-config
      make deploy-all-in-kind |& ts | ts -i | ts -s | tee local/deploy-all-in-kind-2regions.log

#. Check for pods that are not healthy.

   .. code:: bash

      watch 'kind get clusters | grep idc | xargs -i kubectl --context kind-{} get pods -A | egrep -v "NAMESPACE|Running|Completed"'

#. You may view all pods with the following command.

   .. code:: bash

      kind get clusters | grep idc | xargs -i kubectl --context kind-{} get pods -A

#. To create an instance in a different region, run;

   .. code:: bash

      export REGION=us-dev-2

Testing Techniques
------------------

Running a Single Ginkgo Test
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Mark one or more tests with ``Focus`` as shown below. See https://onsi.github.io/ginkgo/#focused-specs for details.

.. code:: go

   It("should work", Focus, func() {
       ...
   })

Run the test suite with maximum verbosity.

.. code:: bash

   BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_integration_test/..." make test-custom

Excluding Go Test Suites
~~~~~~~~~~~~~~~~~~~~~~~~

Tests that have external dependencies that are not widely available to all users should be excluded from ``make test``.
Entire Go test suites can be excluded by adding ``tags = ["manual"]`` to the ``go_test()`` definition in the ``BUILD.bazel`` file.

Such a test suite can be executed manually with the command below.

.. code:: bash

   BAZEL_EXTRA_OPTS="--test_output=streamed //go/pkg/compute_integration_test/..." make test-custom

How to use the Vault CLI in kind
--------------------------------

.. code:: bash

   export VAULT_ADDR=http://localhost:30990/
   export VAULT_TOKEN=$(cat local/secrets/VAULT_TOKEN)
   vault secrets list

How to access services on kind from your laptop
-----------------------------------------------

#. Edit the hosts file in your laptop running your browser (``C:\Windows\System32\drivers\etc\hosts``).
   It should have the line from
   `deployment/common/etc-hosts/hosts <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/common/etc-hosts/hosts>`__
   but with the IP address pointing to the host running kind. 

#. For token generation, visit:

   https://dev.oidc.cloud.intel.com.kind.local

#. For invoking global APIs (grpc-rest-gateway), visit:

   https://dev.api.cloud.intel.com.kind.local

#. For invoking regional APIs (grpc-rest-gateway), visit:

   https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local

#. Other URLs:

   -  https://dev.vault.cloud.intel.com.kind.local
   -  http://dev.netbox.us-dev-1.api.cloud.intel.com.kind.local/

How to access the Vault UI in kind
----------------------------------

Use VS Code to forward port 30990 to localhost:30990. Then visit:

http://localhost:30990/

Login using the Vault token in ``local/secrets/VAULT_ROOT_KEY``.

Argo CD
-------

Argo CD is deployed with ``make deploy-all-in-kind-v2``.

How to access Argo CD UI in kind
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Get ``admin`` password:

.. code:: bash

   ARGOCD_PASSWORD=$(kubectl get secret -n argocd argocd-initial-admin-secret -o go-template='{{.data.password | base64decode}}')
   echo ${ARGOCD_PASSWORD}

Use VS Code to forward port 30960 to localhost:30960. Then visit:

http://localhost:30960/

How to use the Argo CD CLI in kind
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The recommended way of monitoring and controlling Argo CD is through the Kubernetes CRDs such as Applications and ApplicationSets.
If you need to use the Argo CD CLI, follow the steps below.

.. code:: bash

   export ARGOCD_SERVER=localhost:30960
   export ARGOCD_OPTS="--plaintext"
   ARGOCD_PASSWORD=$(kubectl get secret -n argocd argocd-initial-admin-secret -o go-template='{{.data.password | base64decode}}')
   argocd login ${ARGOCD_SERVER} --username admin --password "${ARGOCD_PASSWORD}"
   argocd app list

For more information, see
`deployment/argocd/README.md <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/argocd/README.md>`__.

Gitea
-----

Gitea provides a Github-like environment locally. It is deployed with ``make deploy-all-in-kind-v2``.

How to access Gitea UI in kind
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Get ``gitea_admin`` password:

.. code:: bash

   GITEA_ADMIN_PASSWORD=$(cat local/secrets/gitea_admin_password)
   echo ${GITEA_ADMIN_PASSWORD}

Use VS Code to forward port 30965 to localhost:30965. Then visit:

http://localhost:30965/

Common Issues
-------------

Issue: E1010 08:57:43.772801   78004 memcache.go:265] "Unhandled Error" err="couldn't get current server API group list: Get \"https://10.98.58.75:6443/api?timeout=32s\": Forbidden"
Remedy: Check that your no_proxy env var is set correctly. It should include the IP address of this node.


Issue: github.com/google/cel-go/interpreter: github.com/aws/aws-sdk-go-v2@v1.30.4: Get "https://proxy.golang.org/github.com/aws/aws-sdk-go-v2/@v/v1.30.4.mod": tls: failed to verify certificate: x509: certificate signed by unknown authority
Remedy: sudo apt install ca-certficates && sudo update-ca-certificates

See Also
--------

-  `IDC Deployment Architectures for Testing on a Single Host <https://internal-placeholder.com/x/Q9jvs>`__
