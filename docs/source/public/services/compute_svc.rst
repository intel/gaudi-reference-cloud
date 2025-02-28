.. _compute_svc:

Compute
########

What it is
**********

Dedicated compute, as a bare metal or virtual machine instance, running on |INTC| processors.

Why use
*******

* Accelerate AI development for training and inference, applications, and micro-services
* Evaluate performance of AI/ML workloads running on |INTC| processors
* Get compute resources for free (on some processors), coupon credits, and competitive price-performance
* Leverage AI Accelerators for premium or enterprise accounts

.. mermaid::
   :caption: Instance use cases: Bare metal and Virtual Machines
   :alt: Instance use cases: Bare metal and Virtual Machines
   :align: center

   flowchart TD
     b("Bare Metal")  --> d("General use: CPU, GPU, AI-accelerated processors")
     c("Virtual Machines") --> e("Manage workloads: CPU and AI applications")

.. _instance_types:

.. list-table:: Instance Types
   :header-rows: 1
   :class: table-tiber-theme

   * - Type
     - Description
     - Memory
     - Typical Use Case(s)

   * - Bare Metal (BM)
     - Bare metal compute instances are intended for general use and deploying CPU, GPU, and AI-accelerated processors.
       Current offerings include: 4th Generation Intel® Xeon® Scalable processors; Intel® Max Series GPU (PVC); Habana Gaudi2 Deep Learning Server.
     - 256GB and 1TB
     - AI and core computing.

   * - Virtual Machine (VM)
     - VM compute instances are intended for managing workloads in CPU and AI applications.
       They help support developers world-wide to test and experiment with on-demand workloads and applications.
       A VM requires a hypervisor, which consumes some of its computing power.
       Current offerings include: 4th Generation Intel® Xeon® Scalable processors.
     - 16GB (small), 32GB (medium), and 64GB (large).
     - Workload testing and application development using CPUs, GPUs, and memory in the Intel ecosystem.

..
  TODO: Add next section in Phase 2
.. Specifications & Pricing
 ************************
 TBD

Where to start
***************

In `Compute`_, you can:

* Launch an instance to test and evaluate AI/ML workloads in a few steps
* Discover performance improvements using |INTG2-ACC| for GenAI and LLMs
* Learn how Intel® Advanced Vector Extensions 512 (Intel® AVX-512) speed development on the |IXP|

Related services
****************

* Apply a `Load Balancer`_ to your instance to run and scale services behind an IP address.
* Combine `Object Storage`_ or `File Storage`_ with your instance in an LLM inference pipeline.
* Get access to reservation-based access to `Preview`_  for free for proof of concepts or experiments.

Related documents
*****************

* :ref:`quick_start`
* :ref:`manage_instance`
* :ref:`load_balancer`
* :ref:`preview_cat`

.. _Compute: https://console.cloud.intel.com/compute
.. _File Storage: https://console.cloud.intel.com/storage
.. _Load Balancer: https://console.cloud.intel.com/load-balancer
.. _Object Storage: https://console.cloud.intel.com/buckets
.. _Preview: https://console.cloud.intel.com/preview
.. _Intel® Data Center GPU Max Series: https://www.intel.com/content/www/us/en/developer/articles/technical/intel-data-center-gpu-max-series-overview.html

