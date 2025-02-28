:alttitle: Preview Keys
:category: preview
:keywords: preview keys, early-release, pre-release, processor preview
:phrase: Learn How to create and add public SSH keys to your preview instance. Connect securely via SSH. Get started with preview.
:rating: 0
:show_urls: true

.. _preview_keys:

Preview Keys
#############

To connect to a preview instance via an SSH client, you must create preview keys. Use this guide to create and manage Preview Keys.

.. note::
   Preview keys are stored separately from a regular compute instance’s SSH keys.

Create SSH public keys
**********************

.. caution::
   Take care not to overwrite any local SSH public keys you may have.

#.	Follow the same process found in :ref:`gen_ssh_key_header` to create your :guilabel:`Preview Keys` for a preview instance.
#.	Click :guilabel:`How to create an SSH Key` and follow instructions.
#.	Continue.

Add Preview Keys
****************

.. note::
   In the :guilabel:`Preview Catalog`, adding a public SSH key to launch an instance is optional.

#.	Assure you created an SSH public key on your machine.

#.	Visit the |ITAC| `console`_ home page.

#.	In the console main menu, click the :guilabel:`Preview -> Preview Keys`.

#.	In the Preview Keys tab:

#. Click :guilabel:`Upload key`.

   .. note::
      You can add multiple keys.

#.	Switch to the :guilabel:`Preview Instances` tab.

#.	Click :guilabel:`Request instance`.

#.	Complete all required fields in the form :guilabel:`Request a preview instance`.

#. Complete :guilabel:`Deployment Details` (required).

#. Click :guilabel:`Request`.

   .. note::
      Upon successful completion of the request, you are shown :guilabel:`Preview Instances`.

Await Request Notification
**************************

.. include:: preview_cat.rst
   :start-after: await_request_start:
   :end-before: await_request_end:

Launch instance
*****************

.. include:: preview_cat.rst
   :start-after: launch_preview_instance-start:
   :end-before: launch_preview_instance-end:


.. _co_development:

Collaboration and co-development
********************************

Co-development is optional. If you want to add another user, get the user's email address. This instruction assumes that you are the account administrator.

Acceptable email domains include:

* Users with the **same email domain** as your company
* Intel staff with **Intel email domain** (for co-development and support)

Add Email Domain from your Company
==================================

#. Under :guilabel:`Associated Email`, add a user with **same email domain** as your company.

#. Paste the SSH public key in :guilabel:`Key contents`.

#. Click :guilabel:`Upload key`.


Add Email Domain from Intel
===========================

#. Under :guilabel:`Co-development`, click the toggles switch.

#. Under :guilabel:`Associated Email`, add the Intel staff email address for the person with whom you're collaborating.

#. Paste the SSH public key in :guilabel:`Key contents`.

#. Click :guilabel:`Upload key`.

.. _console: https://console.cloud.intel.com/
