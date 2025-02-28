.. _golang_upgrade_steps:

Steps to upgrade Golang version
###############################

Update Go Version
*****************

#.  Create a feature branch and update the following files with the desired Golang version.

    * deployment/common/install_requirements.sh
    * deployment/docker-images/deployer/Dockerfile
    * go.work
    * go/go.mod
    * idcs_domain/bmaas/enrollment/apiservice/go.mod
    * idcs_domain/catalog-validator/go.mod
    * idcs_framework/tests/compute/go.mod
    * idcs_framework/tests/goFramework/go.mod
    * idcs_framework/tests/Scaleperf/chaos-scripts/go.mod
    * idcs_shared/ansible_playbooks/roles/kind_deployment/vars/main.yml
    * idcs_shared/docker/Dockerfile.coverity
    * WORKSPACE

#.  Update dependencies.

    .. code-block:: bash
        
        make gazelle

Local Validation
****************

#.  Run the following commands to ensure there are no compatibility issues.

    .. code-block:: bash
        
        make generate build test

    **Note**: If there are any errors, update the dependencies and re-validate.
    See :ref:`updating_go_dependencies`.

#.  Ensure that deploy-all-in-kind-v2 works by performing the steps in
    :ref:`deploy_idc_core_services_in_local_kind_cluster`.

Create a PR
***********

#.  Create a PR and add the following labels:

    * bazel-bm
    * test-bazel-large
    * Ready for SDL scans

#.  Ensure that all Jenkins tests pass.

Merge the PR
************

#.  Merge the PR into main.

#.  Send a message in the "IDC Monorepo Announcements" Teams chat indicating that the Golang version has been updated.

Deployment and Testing
**********************

Each IDC team should use the standard process to deploy this change to staging.
Each upgrade should be followed by regression tests.
`Steps to upgrade the Golang version for regression tests <https://internal-placeholder.com/display/devcloud/Steps+to+upgrade+Golang+version+for+Regression+Tests>`__.
