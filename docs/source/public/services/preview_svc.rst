.. _preview_svc:

Preview
#######

What it is
**********

.. _preview_start:

`Preview`_ offers subscribers the ability to try out pre- and early-release |INTC| hardware on a limited basis.
Available preview hardware is organized by common use cases like **AI**, **CPU**, and **GPU**.

.. _preview_end:

.. note::
   The Preview service is completely separate from :ref:`default compute instances <manage_instance>`
   in the `hardware catalog`_.

Why use
*******

* Deploy and test AI/ML workloads on cutting-edge |INTC| processors at low to no cost.
* Connect to your preview instance via one-click (without uploading SSH keys).
* Evaluate |INTC| hardware in a safe and secure development environment.

.. mermaid::
   :caption: Preview catalog
   :alt: Preview catalog
   :align: center

   graph LR
    subgraph "Request Preview instance"
    direction LR
      a[Request] -->|Await approval|b[Receive email confirmation]
    end

    subgraph "**Connect** Bucket 1"
      direction LR
      e[Preview] <--> f{{Storage}}
    end

    subgraph "Access preview instance"
    direction LR
      c[Preview] -->|Connect via One-click| d[instance]
      c[Preview] -->|Connect via SSH| d[instance]
    end

Where to start
***************

In `Preview`_, you can:

* Try the |INTG2-ACC| for managing advanced AI workloads.
* Learn the advantages of the Intel® Core™ Ultra processor family.
* Test AI workloads on the |GPUMAX| and |GPUFLX|.

Related services
****************

* Visit `Preview`_ to request access to pre-release hardware
* Connect `Preview Storage`_ to your preview instance

Related documents
*****************

* Request a :ref:`Preview Instance <preview_cat>`
* Set up :ref:`preview_keys` to securely connect to your preview instance
* Get :ref:`preview_storage` for your preview instance

.. _hardware catalog: https://console.cloud.intel.com/hardware
.. _Preview: https://console.cloud.intel.com/preview/hardware
.. _Preview Storage: https://console.cloud.intel.com/preview/storage
.. _Intel® Data Center GPU Max Series: https://www.intel.com/content/www/us/en/developer/articles/technical/intel-data-center-gpu-max-series-overview.html

