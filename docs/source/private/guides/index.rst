.. _guides:

Guides
#######

Learn the benefits of using Intel processors integrated with curated AI/ML software stacks.

These docs are hosted at `Internal Docs`_.

.. tip::
   Find basic user operations for the platform in :file:`source/public/guides`.

.. grid:: 3
   :gutter: 1
   :padding: 2
   :class-container: sd-container-fluid

   .. grid-item-card::
      :link: local_development_guide
      :link-alt: Local Development Guide
      :link-type: ref

      Local Development Guide
      ^^^
      * Local testing
      * Deployment in a kind cluster
      +++

   .. grid-item-card::
      :link: testing_guide
      :link-alt: Testing Guide
      :link-type: ref

      Testing Guide
      ^^^
      * Automated tests
      * Troubleshooting tests
      +++

   .. grid-item-card::
      :link: create_new_idc_service
      :link-alt: Create New IDC Service
      :link-type: ref

      Create New IDC Service
      ^^^
      * Create IDC service
      * Build and deploy
      * Test
      +++

   .. grid-item-card::
      :link: grpc_api_guide
      :link-alt: IDC gRPC API Guide
      :link-type: ref

      grpc API Guide
      ^^^
      * IDC add-on services
      * Use IDC gRPC APIs
      * Usage parameters
      +++

   .. grid-item-card::
      :link: enable_disable_grcp_service
      :link-alt: Enable/Disable GRCP Service
      :link-type: ref

      Enable/Disable GRCP Service
      ^^^
      Enable GRCP Service in All or Specific Environments
      +++

   .. grid-item-card::
      :link: services_upgrade_procedure
      :link-alt: IDC Services Upgrade Procedure
      :link-type: ref

      Services Upgrade
      ^^^
      * Upgrade IDC services
      +++

   .. grid-item-card::
      :link: logs_and_traces
      :link-alt: Logs and Traces development guide
      :link-type: ref

      Logs and Traces development guide
      ^^^
      Best practices about logs and traces for your Go service.
      +++

   .. grid-item-card::
      :link: services_deployment_procedure
      :link-alt: IDC Services Deployment Procedure
      :link-type: ref

      Services Deployment
      ^^^
      * Deploy IDC services to a new environment
      * Create DNS records
      * Create load balancer
      +++

   .. grid-item-card::
      :link: harvester_deployment_procedure
      :link-alt: Harvester Deployment Procedure
      :link-type: ref

      Harvester Deployment Procedure
      ^^^
      * Deploy a new Harvester cluster for VMaaS
      +++

   .. grid-item-card::
      :link: services_deployment_procedure_development
      :link-alt: IDC Services Deployment for Development
      :link-type: ref

      Services Deployment for Development
      ^^^
      * Deploy IDC services to a new development environment
      +++

   .. grid-item-card::
      :link: services_upgrade_procedure_development
      :link-alt: IDC Services Upgrade Procedure for Development
      :link-type: ref

      Services Upgrade for Development
      ^^^
      * Upgrade IDC services in a development environment

   .. grid-item-card::
      :link: universe_deployer
      :link-alt: Universe Deployer
      :link-type: ref

      Universe Deployer
      ^^^
      * Tool for deploying IDC services
      +++

   .. grid-item-card::
      :link: manage_k8s
      :link-alt: Create K8S Cluster
      :link-type: ref

      Create K8S Cluster
      ^^^
      Create and manage a Kubernetes cluster.
      +++

   .. grid-item-card::
      :link: golang_upgrade_steps
      :link-alt: IDC Golang version upgrade
      :link-type: ref

      Upgrade Golang Version
      ^^^
      Steps to upgrade golang version.
      +++

   .. grid-item-card::
      :link: vmaas_default_machine_image_update_procedure
      :link-alt: VMaaS Default Machine Image Update Procedure
      :link-type: ref

      Update VMaaS Default Machine Image\
      ^^^
      Steps to create and update vmaas default machine image
      +++

   .. grid-item-card::
      :link: vmaas_gaudi2_firmware_upgrade_procedure
      :link-alt: VMaaS |INTG2| Firmware Upgrade Procedure
      :link-type: ref

      Upgrade the VMaaS |INTG2| Firmware
      ^^^
      Steps to create and update vmaas default machine image
      +++

   .. grid-item-card::
      :link: fleet_management
      :link-alt: Fleet Management
      :link-type: ref

      Fleet Management
      ^^^
      * Compute Node Pools
      +++

   .. grid-item-card::
      :link: security_testing
      :link-alt: Security Testing
      :link-type: ref

      Security Testing
      ^^^
      * TLS
      * gRPC
      +++

   .. grid-item-card::
      :link: bazel
      :link-alt: Bazel
      :link-type: ref

      Bazel
      ^^^
      * Bazel caching
      +++

   .. grid-item-card::
      :link: quick_connect_deployment_guide
      :link-alt: Quick Connect Deployment Guide
      :link-type: ref

      Quick Connect Deployment
      ^^^
      * Deploy Quick Connect to a new environment
      +++

.. toctree::
   :maxdepth: 3
   :hidden:

   local_development_guide
   testing_guide
   create_new_idc_service
   grpc_api_guide
   enable_disable_grcp_service
   services_upgrade_procedure
   services_upgrade_procedure_development
   logs_and_traces
   services_deployment_procedure
   harvester_deployment_procedure
   services_deployment_procedure_development
   universe_deployer
   manage_k8s
   customizing_baremetal_instances
   golang_upgrade_steps
   vmaas_default_machine_image_update_procedure
   vmaas_gaudi2_firmware_upgrade_procedure
   fleet_management
   security_testing
   bazel
   quick_connect_deployment_guide


.. _Internal Docs: https://internal-placeholder.com/satg-dcp-dcbmaas/job/BMAAS-Orchestrator/job/main/lastSuccessfulBuild/artifact/bazel-bin/docs/source/private_docs_html/guides/index.html
