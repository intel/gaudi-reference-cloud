:alttitle: Preview Catalog
:category: preview
:keywords: early-release, pre-release, processor preview
:phrase: Try pre-release and early-release CPU, GPU, and AI processors to evaluate and test workload performance. Request a preview instance.
:rating: 0
:show_urls: true

.. _preview_cat:

Preview Catalog
###############

Learn how to request and access a preview instance.

.. caution::

   Preview Catalog instances have an expiration date. Ensure you back up and save all instance data before the expiration date.

Prerequisites
*************

* |ITAC| account

View Preview Catalog
**********************

#. Visit the |ITAC| `console`_ home page.

#. Click the :command:`Preview > Preview Catalog` from the menu at left.

#. In the :guilabel:`Preview Catalog` tab, select an instance from the options.

.. _request_a_preview_instance:

Request a Preview Instance
**************************

.. _instance_family:

.. tip::
   For all Linux\* Operating Systems, Jupyterlab\* is launched via One-Click.
   For all Windows\* operating systems, remote desktop (RDP) is launched.

#. Use the form to :guilabel:`Request a Preview Instance`.

#. Follow the onscreen instructions and fill out all required fields.

   .. note::
      Required fields show an asterisk '*' character.

#. Select the :guilabel:`Instance family`.

   .. note::

      Depending on which :guilabel:`Instance family` you choose, Windows\* OS or Linux\* OS launches in your browser.

#. Under **Request Information**, select your preferences for :guilabel:`Use case` and :guilabel:`Duration`.

#. For :guilabel:`Deployment details`, type in a description (required).

#. Continue below.

Await Request Notification
**************************

.. _await_request_start:


After requesting an instance, you'll receive an email in 2-3 business days with request status.

.. important::

   | The email response indicates the request status: *approved*, *rejected*, or *waitlisted*.
   | If your request is pre-approved, your instance may be available immediately.

#. If your request is approved, log into your account.

#. From the left side menu, navigate to :guilabel:`Preview --> Preview Instances`.

#. View the table in the :guilabel:`Preview Instances` tab.

#. Click on your approved instance.

#. Continue.

.. _await_request_end:


Choose Method to Connect
************************

Choose a method to connect to a preview instance:

* `Connect via One-Click`_
* :ref:`Connect via SSH client <preview_keys>`

.. tip::
   You must check SSH Keys when :ref:`creating an instance <request_a_preview_instance>` for SSH option.


Connect via One-Click
=====================

Connect to a preview instance using One-Click connection. Your instance launches in a browser.

.. note::
   You are not required to add an SSH key using this method.

Click :guilabel:`Request`.

See next section.

.. _await_request_notification:

Launch instance
*****************

.. _launch_preview_instance-start:

#. Navigate to :guilabel:`Preview Instances` tab.

#. Wait until your instance :guilabel:`State` shows :guilabel:`Ready`.

#. Click on the instance name. Then choose a method to access.

   a. Click :guilabel:`Connect` for One-Click access.

      .. note::
         This action launches your preview instance in a browser. See also the :ref:`note regarding operating system <instance_family>`.

   #. In :guilabel:`Details` tab, click on :guilabel:`How to Connect via SSH`.
      Follow instructions.

.. _launch_preview_instance-end:

File Transfer Methods
=====================

Choose your operating system and follow commands for **upload/download**, or file transfers.

.. tabs::

   .. tab:: Linux Operating System

      #. To upload, select the :guilabel:`Upload Files`` icon in the upper left. Follow instructions.

      #. To download, right-click on the file. In the :guilabel:`pop-up` menu, select :guilabel:`Download`

   .. tab:: Windows Operating System

      #. To upload/download files, select :kbd:`CTRL+ALT+SHIFT``

      #. Select "Devices" in the dialog menu.

      #. Click :guilabel:`Upload Files` to upload.

      #. To download a file, double-click it and follow the dialog.

.. _console: https://console.cloud.intel.com/
