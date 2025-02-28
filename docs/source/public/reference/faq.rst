.. _faq:

FAQ
###

Find answers to frequently asked questions (FAQs) from our experts and community. See also :ref:`support` to explore our *knowledge base* and *community* forum.

Billing
*******

.. dropdown:: How am I billed if both cloud credits and a credit card are in the system?
   :animate: fade-in
   :color: info

   The system will deplete the cloud credits first. Thereafter, the balance is charged to the credit card.

.. dropdown:: What happens to services consumed when the account is out of credit?
   :animate: fade-in
   :color: info

   The instances will be shut down.

Services
********

.. dropdown:: What services are currently available on the IDC Console?
   :animate: fade-in
   :color: info

   Visit :ref:`svc_overview` for an overview of all services.

   * Bare metal and Virtual Machine instances. See :ref:`compute_svc`.
   * Object storage and file storage. See :ref:`storage_svc`
   * Intel Kubernetes\* Services. See  :ref:`k8s_svc`
   * Preview Catalog for |INTC| pre-release hardware. See :ref:`preview_svc`
   * Free pre-loaded Juypter Notebooks. See :ref:`learning_svc`


Software
*********

.. dropdown:: What if I cannot SSH into a reserved instance?
   :animate: fade-in
   :color: info

   * First, go to the main console and select the :guilabel:`key` icon.

     - In the table, your existing SSH key data should appear below **Name** and **Type** columns.
       If no data appears, follow next section.

   * Second, verify that you successfully uploaded your SSH key (public) to the IDC console.

     If in doubt, follow the instructions in :ref:`ssh_keys`.

     - Try uploading your local SSH Key to the IDC console again.
     - If **no SSH key exists**, regenerate an SSH key.

   * Third, if you exhausted the above options, contact :ref:`support`.

.. _faq_cloud_monitor:

Cloud Monitor
*************

.. dropdown:: Which resources support Cloud Monitor?
   :animate: fade-in
   :color: info

   * Cloud Monitor is supported for virtual machines, bare metal, the |IKS| control plane, and instance groups.

.. dropdown:: Why is that my metrics show spikes when I select Last one hour, but the spikes are not shown with larger time ranges?
   :animate: fade-in
   :color: info

   * If the spikes are for a very small interval, say 30 seconds, These may be sampled out. Sampling is applied for metrics.

.. dropdown:: Why is my Cloud Monitor graph giving "Data is Unavailable error"?
   :animate: fade-in
   :color: info

   * You may see this error in the following cases:

     - It may take a few mins for the metrics to start appearing after the resource is in Ready State.  Please retry after sometime.

     - For bare metal isntances, this error would occur if the Cloud Monitor agent is not enabled. The agent is only enabled in the latest machine images. See the "Image equipped with" section of the bare metal image selected in the Machine Image dropdown when you create an instance to check if OpenTelemetry Collector is present.

     - If you continue to see this error, contact Support.

.. dropdown:: Is there any sampling done for Cloud Monitor metrics?
   :animate: fade-in
   :color: info

   * Yes, sampling is applied for metrics.

.. dropdown:: Are the metrics shown in graphs rounded off?
   :animate: fade-in
   :color: info

   * Yes, the metrics are rounded off to 2 decimal places.

.. dropdown:: How do I enable/disable  a certain legend in the graph?
   :animate: fade-in
   :color: info

   * Legends can be selectively enabled/disabled by clicking on them.

.. dropdown:: My Intel® Kubernetes Service cluster does not have the Cloud Monitor tab. How do I enable it?
   :animate: fade-in
   :color: info

   * Cloud Monitor for Intel Kubernetes Service is enabled only for clusters created after Feb 4th 2025. For clusters created before this date, please contact Support.


.. dropdown:: Can I get metrics for other Kubernetes Control Plane components in IKS?
   :animate: fade-in
   :color: info

   * Currently Kubernetes API Server and etcd metrics are supported. Please contact support.

.. dropdown:: Which images of BM support Cloud Monitor?
   :animate: fade-in
   :color: info

   * See the "Image equipped with" section of the BM image selected in the Machine Image dropdown when you create an instance to check if OpenTelemetry Collector is present.

.. dropdown:: Is Cloud Monitor enabled for all virtual machines?
   :animate: fade-in
   :color: info

   * Yes, Cloud Monitor is enabled for all virtual machines.


.. meta::
   :description: Get answers to frequently asked questions (FAQ) for services, billing, and more on Intel® Tiber™ AI Cloud.
   :keywords: AI Cloud FAQ, AI Cloud, FAQ, AI Cloud billing
