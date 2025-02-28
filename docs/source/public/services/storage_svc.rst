.. _storage_svc:

Storage
#######

What it is
**********

Object storage and file storage support model training and inference, analytics, and testing AI/ML workloads.

Why use
*******

* Integrate data, whether object or file storage, for training and inference pipelines.
* Analyze any object in a bucket by functional characteristics with :ref:`Object Storage <staas_object>`.
* Develop cloud-native applications with :ref:`Object Storage <staas_object>` or :ref:`File Storage <staas_file>`

.. mermaid::
   :caption: Object and File Storage
   :alt: Object and File Storage
   :align: center

   flowchart TD

    A --> B{{ Storage bucket }}
    A --> C{{ Storage bucket }}
    A{{ Block storage }}

    D --> E[/ Files in folders /]
    D --> F[/ Files in folders /]
    D --> G[/ Files in folders /]
    D{{ File storage }}

Where to start
***************

* Integrate storage with your LLM inference workflow at low cost

In `Object Storage`_, you can:

* Create a bucket for unstructured data and connect other services.

In `File Storage`_, you can:

* Create a volume for data in shared directories and connect other services.

Related services
****************

* Combine `Object Storage`_ with a :ref:`RAG Inference Engine <tutorials>`.

* Review the :ref:`model_matrix` to choose the best processor for a training and inference pipeline, including a :ref:`storage solution <staas_overview>`.

* Upload a dataset to `Object Storage`_ or `File Storage`_. Then follow inference and training on the :ref:`Intel® Gaudi® 2 processor<tutorials>`.

Related documents
******************

* :ref:`Storage Overview <staas_overview>`
* :ref:`Object Storage Guide <staas_object>`
* :ref:`File Storage Guide <staas_file>`

.. _File Storage: https://console.cloud.intel.com/storage
.. _Object Storage: https://console.cloud.intel.com/buckets
.. _Intel® Data Center GPU Max Series: https://www.intel.com/content/www/us/en/developer/articles/technical/intel-data-center-gpu-max-series-overview.html

