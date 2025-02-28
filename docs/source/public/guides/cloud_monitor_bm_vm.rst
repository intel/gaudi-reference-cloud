:alttitle: Cloud Monitor for metrics and logging of your compute instances
:category: compute
:keywords: cloud monitor, metrics, logging, telemetry, CPU usage, network traffic, I/O traffic, read/write operations
:phrase: Cloud Monitor offers metrics and logging for bare metal and VM instances, exposing CPU usage, memory usage, I/O traffic, and more.
:rating: 0
:show_urls: true

.. _cloud_monitor_bm_vm:

Instance Cloud Monitor
######################

Cloud Monitor provides metrics and logging on both **bare metal** and **virtual machine** (VM) compute instances.

.. note::
   You must have a **compute instance running** to show Cloud Monitor. For help, see :ref:`get_started`.

Prerequisites
*************

* :ref:`get_started`
* :ref:`manage_instance`

Choose an **instance type** for Cloud Monitor:

* `Bare Metal Cloud Monitor`_
* `Virtual Machine Cloud Monitor`_

Bare Metal Cloud Monitor
************************

Cloud Monitor for bare metal instances offers a single dashboard for visualizing metrics such as CPU usage, memory usage, disk utilization, and more. Navigate to the metric you wish to monitor.

* `CPU Usage`_
* `Memory Usage`_
* `Network Usage Received`_
* `Network Usage Transmit`_
* `IO Traffic`_
* `Disk Utilization`_

Navigate to Instance
====================

#. In the console menu at left, click the :guilabel:`Cloud Monitor`.
#. Next, click :guilabel:`Instances`, which launches its own tab.
#. For each metric, select a **range**, or time span, under the :guilabel:`View` dropdown menu.

.. note::
   Dashboard metrics data are only retained for the previous 30 days.

CPU Usage
=========

This chart shows CPU utilization in % for bare metal instances. Use it to understand whether overall CPU usage is high or if there are overload scenarios during a specific time.

CPU utilization includes:

* **User**: Time spent executing user-space processes.
* **System**: Time spent executing kernel-space processes.
* **I/O Wait**: Time spent waiting for I/O operations.
* **Steal**: Time stolen by the hypervisor (in virtualized environments).
* **Nice**: CPU time used by low-priority processes.

Memory Usage
============

This chart shows **Memory** utilization in % for bare metal instances. Use it to understand how much memory (RAM) a bare metal instance uses over time.

Network Usage Received
======================

This chart shows network utilization in Mb/second that are **received** per network cards. Network metrics measure the overall network traffic for both inbound and outbound data.

Network Usage Transmit
======================

This chart shows network utilization in Mb/second that are **transmitted** per network cards. Network metrics measure the overall network traffic for both inbound and outbound data.

IO Traffic
==========

IO Traffic is the amount of data read from and written to disk over a period of time, measured in KiloBytes per second (KBps).
High IO traffic can indicate heavy usage, which may lead to bottlenecks or performance issues if not managed properly.

Key metrics:

* **Read IO Traffic**: The amount of data read from disk, measured in KiloBytes per second (KBps).
* **Write IO Traffic**: The amount of data written to disk, measured in KiloBytes per second (KBps).

Disk Utilization
================

The Disk utilization chart shows **Filesystem Usage per Mount Point** in percentage for a bare metal instance.

Virtual Machine Cloud Monitor
*****************************

Following are the Cloud Monitor metrics for **virtual machine** (VM) instances. Navigate to the metric you wish to monitor.

* `CPU Usage VM`_
* `Memory Usage VM`_
* `Network Usage VM`_
* `IO Traffic VM`_
* `IOPS Usage VM`_
* `IO Times VM`_

CPU Usage VM
============

This chart shows CPU utilization in % for virtual machine instances. Use it to understand whether overall CPU usage is high or if there are overload scenarios during a specific time.

Memory Usage VM
===============

This chart shows **Memory** utilization in % for virtual machine instances. Use it to understand how much memory (RAM) a bare metal instance uses over time.

Network Usage VM
================

This chart shows network utilization in Mb/second that are **received** and **transmitted**. Network metrics measure the overall network traffic for both inbound and outbound data.

IO Traffic VM
=============

IO Traffic is the amount of data read from and written to disk over a period of time, measured in KiloBytes per second (KBps).
High IO traffic may indicate heavy usage, which may lead to bottlenecks or performance issues if not managed properly.

Key metrics:

* **Read IO Traffic**: The amount of data read from disk, measured in KiloBytes per second (KBps).
* **Write IO Traffic**: The amount of data written to disk, measured in KiloBytes per second (KBps).

IOPS Usage VM
=============

:guilabel:`IOPS usage` shows the speed at which a storage device or network can read and write data per second (input/output operations per second).

Key IOPS for :guilabel:`IOPS usage`

* **Read IOPS**: The number of read operations per second.
* **Write IOPS**: The number of write operations per second.

IO Times VM
===========

:guilabel:`IO Times` is the amount of time taken to complete input/output operations on a storage device, per millisecond (ms). 
High :guilabel:`IO Times` may indicate potential performance issues, such as disk latency or bottlenecks.

Key metrics:

* **Read IO Time**: The amount of time it takes to complete read operations per millisecond (ms).
* **Write IO Time**: The amount of time it takes to complete write operations per millisecond (ms).

Next Steps
**********

For more help and troubleshooting, see :ref:`faq_cloud_monitor`.

.. meta::
   :description: Use Cloud Monitor for metrics and logging of your compute instances on Intel® Tiber™ AI Cloud.
   :keywords: metrics, logging, CPU usage, I/O traffic

.. collectfieldnodes::
