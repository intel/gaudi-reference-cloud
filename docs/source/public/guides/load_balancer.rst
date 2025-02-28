:alttitle: Load Balancer for application deployment
:category: load balancer
:keywords: distribute traffic, route traffic, listener, protocols, ports, health check
:phrase: Route traffic to one or more instances. Perform health checks. Prepare application deployment.
:rating: 0
:show_urls: true

.. _load_balancer:

Load Balancers
##############

Deploy an application or service by adding an IP address and exposing ports with Load Balancers.

* Route traffic to one or more instances
* Create listeners to handle connection requests
* Expose an application to other users or applications
* Enable health checks

Prerequisites
*************

* Running instance or defined labels
* :ref:`get_started`
* :ref:`manage_instance`

Create Load Balancer
********************

#. In the console main menu, click :command:`Compute -> Load Balancers`.

#. Click :guilabel:`Launch Load Balancer`.

   .. figure:: ../../_figures/guides/load_balancer_00.png
      :alt: Launch a Load Balancer - Input Form
      :align: center
      :class: sd-shadow-sm
      :scale: 75%

      Launch a Load Balancer - Input Form

#. In :guilabel:`Name`, enter a name using a combination of letters, numbers, or dashes, all lowercase.

#. In :guilabel:`Source IPs`, enter one of the following: `any`; IP address(es); or CIDR-format subnet(s).
   This **restricts** which public IPs are allowed to access the Load Balancer.

   .. caution::
      Be careful. Using `any` means no restriction applies, and anyone on the Internet can route to it.

#. Optional: Click :guilabel:`Add Source IP` to add more.

#. Complete subsections to launch a load balancer.

..
   .. tip::
      When a Load Balancer shows :guilabel:`Ready`, a public IP address is provisioned to it.

Listeners
=========

#. Enter :guilabel:`Listener` data using the example below. Your configuration may differ.

   a. :guilabel:`Listener Port`. We enter `80`.

   #. :guilabel:`Instance Port`. We enter  `8081`.

   #. :guilabel:`Monitor Type`. We enter `HTTPS`.
      Click `Monitor Type`_  for more information.

.. tip::
   Click `Listeners Reference`_ for more information.

Selector Type
=============

Choose one:

* **Instance Labels** - You define key/value pairs, which later are applied to an instance
* **Instances** - You must already have one or more instance(s) available

Instances
=========

Choose one or more items to which the Load Balancer applies.

* :guilabel:`Select All`
* :file:`my-instance`

Optional - Add New Listener
===========================

In this example, we create an **additional listener** for the same instance. The example shows how to do an HTTP redirect from `Port 80` to `Port 443`, ensuring a TLS connection.

Add Listener
-------------

#. Click :guilabel:`Add Listener`.

#. Let's add another Listener Port, Instance Port, and Monitor Type as follows:

   a. Listener Port - `443`
   #. Instance Port - `8443`
   #. Monitor Type - `HTTPS`

Instances
=========

Choose one or more items to which the Load Balancer applies.

Replace :file:`my-instance` with your instance name.

* :guilabel:`Select All`
* :file:`my-instance`

#. Finally, click :guilabel:`Launch`.

In the :guilabel:`Compute` dashboard, wait until the :guilabel:`State` shows **Active** for your Load Balancer.

The following tasks are optional. They represent limited steps you might perform.

Validate Load Balancer
**********************

To check the health of your instance/application, test your Load Balancer using `curl`.

.. tip::
   Ensure that the Load Balancer is :guilabel:`Active`.

#. Navigate to :command:`Compute --> Load Balancer`.

#. Choose the load balancer that you created in the previous steps.

#. In the :guilabel:`Details` tab, copy the **Virtual IP** address.

#. Run the command below, pasting *your own* **Virtual IP**, followed by the port number defined earlier.

#. Create a curl command, following an example below. Replace the URL with your **Virtual IP** address.

   .. code-block:: bash

      curl http://10.152.227.65

   .. code-block:: bash

      curl https://10.152.227.65:443

Edit Load Balancer
******************

For an existing Load Balancer, you can edit instances that receive traffic.

#. Create a new instance.

#. Navigate to :command:`Compute --> Load Balancer`.

#. Choose the load balancer that you created in the previous steps.

#. Look for the new instance under :guilabel:`Instances`.

#. Select the new instance name, or choose :guilabel:`Select All`.

   a. Optionally, you can deselect the instance to which the Load Balancer applies.

#. click :guilabel:`Save` to update the Load Balancer.

Delete Load Balancer
********************

#. Navigate to :command:`Compute --> Load Balancer`.

#. Find the :guilabel:`Name` of the Load Balancer you wish to remove.

#. Under :guilabel:`Actions`, click :guilabel:`Delete`.

#. Confirm delete.

Listeners Reference
*******************

Each listener routes a request from a client to instance(s) using the port and protocol that you configure.
A listener must include:

* **Listener Port** - Public port that is exposed to the Internet
* **Instance Port** - Internal port that is exposed on your instance
* **Monitor Type** -  Health check for an application or service. See table below.
* **Mode** - Load-balancing algorithm (e.g., Round Robin)

Monitor Type
============

.. list-table::
   :header-rows: 1
   :class: table-tiber-theme

   * - Protocol
     - Description

   * - TCP
     - A simple service port check

   * - HTTP
     - A request that expects an HTTP response of “200" or OK

   * - HTTPS
     - A secure request that expects an HTTP response of “200" or OK

.. collectfieldnodes::

.. meta::
   :description: Learn to deploy an application or service by adding an IP address and exposing ports with Load Balancers.
   :keywords: AI cloud load balancer, load balancer, route traffic, listener port

..
   Instance Labels
   ***************
..
      .. figure:: ../../_figures/guides/load_balancer_02.png
         :alt: Instance Labels for Load Balancer
         :align: center

         Instance Labels for Load Balancer
..
   TODO:
   Add Instance Labels section when ready
   Add FAQ to address whether to provision TLS certs.
