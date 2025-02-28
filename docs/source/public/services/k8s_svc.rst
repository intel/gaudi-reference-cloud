.. _k8s_svc:

Intel Kubernetes Service
########################

What it is
**********

A fully integrated Kubernetes platform used to manage and scale workloads, including load balancing, cluster management, failover support, batch execution, storage orchestration, and more.

Why use
*******

* Create clusters and automate AI workload deployments at low cost.
* Orchestrate batches of parallel workloads in an LLM inference workflow.
* Benchmark AI/ML workloads, using few to many nodes, to determine the most performant hardware.
* Discover methods to optimize inference when serving an open LLM.

.. mermaid::
   :caption: Kubernetes Control Plane with Two Nodes
   :alt: Kubernetes Control Plane with Two Nodes
   :align: center

   graph LR
    subgraph "Control Plane"
      apiSrv[API Server] -->|etcd| etcd[Etcd]
      cntrlMgr[Controller Manager] -->|apiSrv| apiSrv
      sched[Scheduler] -->|apiSrv| apiSrv
    end

    subgraph "Node 1"
      kubelet1[Kubelet] -->|apiSrv| apiSrv
      container1[Container] -->|kubelet1| kubelet1
    end

    subgraph "Node 2"
      kubelet2[Kubelet] -->|apiSrv| apiSrv
      container2[Container] -->|kubelet2| kubelet2
    end

Where to start
***************

In `Kubernetes`_, you can:

* Create a cluster and access the Kubernetes control plane.
* Add worker node groups that use AI Accelerators.
* Use :file:`kubeconfig` to connect to your clusters.
* Deploy services and run workloads at scale.

Related services
****************

* Review the `Kubernetes`_ Overview
* Launch a `Kubernetes Cluster`_

Related documents
*****************

* Follow the :ref:`Intel Kubernetes Service Guide <k8s_guide>`

.. _Kubernetes: https://console.cloud.intel.com/cluster/overview
.. _Kubernetes Cluster: https://console.cloud.intel.com/cluster
.. _Compute: https://console.cloud.intel.com/compute
.. _File Storage: https://console.cloud.intel.com/storage
.. _Load Balancer: https://console.cloud.intel.com/load-balancer
.. _Object Storage: https://console.cloud.intel.com/buckets
.. _Preview: https://console.cloud.intel.com/preview
.. _Intel® Data Center GPU Max Series: https://www.intel.com/content/www/us/en/developer/articles/technical/intel-data-center-gpu-max-series-overview.html
