:alttitle: Use Cloud Monitor on Intel® Kubernetes Service to manage the control plane and monitor cluster performance
:category: intel K8S service
:keywords: monitor kubernetes, cluster, API servers, visualize clusters, cluster performance
:phrase: Use Cloud Monitor to track API servers, monitor Kubernetes* clusters, or troubleshoot performance on Intel® Tiber™  AI Cloud.
:rating: 0
:show_urls: true

.. _cloud_monitor_iks:

|IKS| Cloud Monitor
###################

Use Cloud Monitor on the |IKS| Control Plane, where you can view Kubernetes clusters.

#. Select the Kubernetes tab for the selected cluster in the |ITAC| `console`_.
#. Select the appropriate cluster in the dropdown.

Kubernetes API Server
*********************

CPU Usage
=========

This chart shows **CPU utilization** in % for the |IKS| Control plane API Server by ``nodename``.  Use it to understand whether overall CPU usage is high or if there are overload scenarios during a specific time.

Memory Usage
============

This chart shows **Memory utilization** in megabytes (MB) for the Kubernetes Control plane API Server by ``nodename``.

Requests by code
================

This chart shows the total number of requests handled by the Kubernetes API server by ``http code``.

Requests by verb
================

This chart shows the total number of requests handled by the Kubernetes API server by ``HTTP verb``.  These verbs reflect standard API request names, like POST, GET, and PATCH.


Latency by node name
=====================

This chart shows the average latency by dividing the total latency by the number of requests, in milliseconds, grouped by the ``nodename``.

Errors by nodename
==================

This chart shows the percentage of error requests grouped by ``nodename``.
An error request displays with error code **4xx**, **5xx**.

The number of error requests is divided by total number of requests to get the percentage error.

Errors by verb
==============

This chart shows the percentage of error requests grouped by ``HTTP verb``.
An error request displays with error code **4xx**, **5xx**.

The number of error requests is divided by total number of requests to get the percentage error.

HTTP Requests by nodename
=========================

This chart shows the total number of HTTP requests handled by the Kubernetes API server by ``nodename``.

Kubernetes etcd Metrics
************************

CPU Usage
=========

This chart shows CPU utilization in % for the IKS Control plane ``etcd``,
grouped by ``nodename``.

Use it to understand whether overall CPU usage is high or if there are overload scenarios during a specific time.

Memory usage
=============

This is a chart that which provides Memory utilization for the Kubernetes Control plane etcd grouped by nodename.

Client traffic In
=================

This chart shows the rate of data received over gRPC by ``etcd`` from all clients grouped by ``nodename``, measured in KiloBytes per second (KBps).


Client traffic Out
==================

This chart shows the rate of data sent, in number of KiloBytes per second (KBps), over gRPC by ``etcd`` to its clients by ``etcd`` grouped by ``nodename``.

Peer traffic In
===============

This chart shows the rate of data received, in number of KiloBytes per second (KBps), over gRPC by ``etcd`` from all its peers grouped by ``nodename``.

Peer traffic Out
================

This chart shows the rate of data number of bytes sent, in number of KiloBytes per second (KBps), over gRPC by etcd to its peers by ``etcd`` grouped by ``nodename``.

DB Size in use
==============

This chart shows the total size of the ``etcd`` database currently in use, in bytes. This metric reflects how much disk space the ``etcd`` key-value store is utilizing for storing data.

Total heartbeat send failures
=============================

This chart shows the total number of failed attempts by an ``etcd`` node to send heartbeats to its peers in the cluster grouped by ``nodename``.

Next Steps
**********

For more help and troubleshooting, see :ref:`faq_cloud_monitor`.

.. meta::
   :description: Use Cloud Monitor on Intel® Kubernetes Service to manage clusters, API services, and visualize performance.
   :keywords: monitor kubernetes, kubernetes API servers, visualize clusters, cluster performance

.. collectfieldnodes::

.. _console: https://console.cloud.intel.com/