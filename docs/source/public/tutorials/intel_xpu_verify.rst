.. _intel_xpu_verify:

XPU Verify Tool
###############

The `xpu-verify`_ tool provides a comprehensive set of tests and automated fixes to help ensure that |INTC| discrete GPUs have been set up correctly on Linux\* operating systems (OS). The tool supports various distributions, such as Ubuntu\* 20.04 and Ubuntu\* 22.04, Red Hat Enterprise Linux\* (RHEL) OS, and Fedora Linux\* OS.

Prerequisites
*************

* :ref:`manage_instance`
* Docker\* software is required to run AI tests

.. note::
   * If starting from scratch and you\'d like to set up intel dGPUs on Linux, check out the `XPU Setup Repo`_.
   * Ensure that the kernel, compute drivers, and other necessary components for Intel discrete GPUs have been set up according to the `Intel GPU documentation`_.
   * To run SYCL tests, install the :program:`intel-oneapi-compiler-dpcpp-cpp` package, which includes the oneAPI compiler.
     - For Ubuntu, find the `installation instructions on Ubuntu`_.
     - For RHEL and SUSE, find the `installation guide for Linux* OS`_.

Setup
*****

#. Complete all prerequisites to :ref:`initiate_instance_start` and launch a bare metal instance.

   a. Check the filter for :guilabel:`Bare metal` type.

   #. Select the filter for **GPU** processor.

#. Click :guilabel:`Select` to select a GPU compute instance.
   In this example, we use the |INTC| Max Series GPU.

#. Connect to your instance.

#. Continue in the next section.

Installation
************

Clone the Intel GPU sanity tests repository and navigate to the directory:

.. code-block:: bash

   git clone https://github.com/unrahul/xpu_verify && cd xpu_ver

Usage
*****

For help, run:

.. code-block:: bash
   
   ./xpu_verify.sh help 

Check System Setup
==================

To check if the system is set up correctly for Intel discrete GPUs, run the script with the -c option:

.. code-block:: bash

   ./xpu_verify.sh -c


Fix System Setup
================

To fix and augment the system setup with essential tools and libraries for Intel discrete GPUs, run the script with the -f option:

.. code-block:: bash

   ./xpu_verify.sh -f

Upon successful completion, a dialog appears: `Which services should be restarted?` 
To accept all defaults, press the Tab key to navigate to :guilabel:`OK` and press :guilabel:`Enter`.


Check and Fix System Setup
==========================

To check and fix the system setup for Intel discrete GPUs, run the script with the -p option:

.. code-block:: bash

   ./xpu_verify.sh -p

AI Libraries Installation
=========================

To install specific AI packages with XPU support (e.g., openvino_xpu, pytorch_xpu, tensorflow_xpu, ai_xpu), run:

.. code-block:: bash

   ./xpu_verify.sh -i pkg1, pkg2,...

Supported Tests
***************

You can perform the following tests.

Linux Kernel i915 Module and Graphics Microcode
================================================

This test checks if the Linux Kernel i915 module is loaded and the Graphics microcode for the GPU is loaded.

.. code-block:: bash

   ./check_device.sh


Check OS kernel and version
===========================

.. code-block:: bash

   ./check_os_kernel.sh


Compute Drivers
================

This test checks if the necessary Intel compute drivers are installed.

.. code-block:: bash

   ./check_compute_drivers.sh


GPU Devices Listing
====================

This test verifies if sycl-ls can list the GPU devices for :program:`OpenCL`` and Level-Zero backends. The oneAPI basekit is required for this test.

.. code-block:: bash

   ./syclls.sh --force


Check if Intel basekit is installed
===================================

.. code-block:: bash

   ./check_intel_basekit.sh


SYCL Programs Compilation
============================

This test checks if sycl programs can be compiled using icpx. The oneAPI basekit is required for this test.

.. code-block:: bash

   ./check_sycl.sh


Check scaling governer
======================

.. code-block:: bash
   
   ./scaling_governor.sh


PyTorch and TensorFlow XPU Device Detection
===========================================

This test checks if PyTorch\* software and TensorFlow\* software can detect the XPU device and run workloads using the device. Docker computer software is required for this test. 

.. tip::
   The test_tensorflow and pytorch scripts only work using a Docker image of the tensorflow and pytorch frameworks.  The main purpose is to verify that these frameworks can access the GPU. Results may differ based on your python `env`. 

For PyTorch:

.. code-block:: bash

   ./check_pytorch.sh

For TensorFlow:

.. code-block:: bash

   ./check_tensorflow.sh

Optional software
-----------------

* `Install Tensorflow\* software`_
* `Install PyTorch\* software`_ 


Additional Checks
*****************

Check if network proxy is setup, print the proxy, remove proxy settings and restore proxy settings:

.. code-block:: bash

   ./proxy_helper.sh

.. _xpu-verify: https://github.com/rahulunair/xpu_verify
.. _XPU Setup Repo: https://github.com/rahulunair/xpu_setup
.. _Intel GPU documentation: https://dgpu-docs.intel.com/installation-guides/index.html
.. _installation guide for Linux* OS: https://www.intel.com/content/www/us/en/docs/oneapi/installation-guide-linux/2023-0/yum-dnf-zypper.html
.. _installation instructions on Ubuntu: https://www.intel.com/content/www/us/en/docs/oneapi/installation-guide-linux/2023-0/apt.html
.. _Install Tensorflow\* software: https://www.tensorflow.org/install
.. _Install PyTorch\* software: https://pytorch.org/get-started/locally/
