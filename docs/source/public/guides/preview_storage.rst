:alttitle: Preview Storage
:category: preview
:keywords: preview storage, early-release, pre-release, processor preview
:phrase: Connect preview storage to your preview instance. Simplify data storage for training and inference on the latest |INTC| processors.
:rating: 0
:show_urls: true

.. _preview_storage:

Preview Storage
###############

You may connect a preview storage bucket to your preview instance. First, request a storage bucket and await approval.

Prerequisites
*************

* Preview instance (pre-exists)
* See also :ref:`request_a_preview_instance`

Overview
********

You must request a Storage Bucket and await approval.

#. To learn whether object storage is available for your instance type, select your instance name.
#.	Click :guilabel:`Details`.
#.	Read the value that follows :guilabel:`Object Storage`. If the value shows “Available…”, object storage is supported for your instance type.

Request Storage Bucket
**********************

#.	After your preview instance shows :guilabel:`Ready`, click the tab :guilabel:`Preview Storage`.
#.	Click on :guilabel:`Request Bucket`.
#.	In the Request storage bucket dialog, select the appropriate Bucket size
#.	Click :guilabel:`Request`.

   .. tip::
      During provisioning, the preview instance :guilabel:`State`` may change from :guilabel:`Requested` to :guilabel:`Provisioning`.
      When the :guilabel:`State` shows :guilabel:`Ready`, you may begin using it.

#. Navigate to the :guilabel:`Preview Storage` tab.

#. Confirm that the :guilabel:`State` shows :guilabel:`Ready`.

#. Continue.

Connect to Storage Bucket
*************************

* Follow this section after your preview storage bucket is approved.
* Configure and connect your instance to the storage bucket.

#. Click :guilabel:`How to use` to view commands.
#.	Follow the onscreen instructions.
#.	Click :guilabel:`Generate key`. This allows you to securely connect from your instance to your storage bucket.

   .. tip::
      Notice that commands are automatically generated after you click :guilabel:`Generate key`.

#. Save your key in a secure location in case of emergency (recommended).

#. Note that the key is valid for one week from time of creation.

   .. caution::
      If you subsequently click :guilabel:`Generate key`, all previous tokens will be deleted, and new ones will be generated.

To use preview storage, you must wait until the preview storage shows :guilabel:`Ready`.

Edit Storage
************

After a storage bucket is created, you may only expand storage.

#.	From :guilabel:`Preview Storage`, where your storage bucket :guilabel:`Name` appears, select :guilabel:`Edit`.
#.	Under :guilabel:`Bucket size`, select a larger bucket size in GB.
#.	Click :guilabel:`Request`.

To continue, you must await email notification and confirmation.

Extend Storage
**************

You will receive a notification via email for bucket storage:

* 1 month before expiration; and
* 1 week before expiration

If you wish to extend bucket storage, continue.

#.	From :guilabel:`Preview Storage`, where your storage bucket :guilabel:`Name` appears, select :guilabel:`Extend`.
#.	Follow instructions in the dialog to extend bucket storage.

To  continue, you must await email notification and confirmation.

Useful Tips
***********

* You may only connect to preview bucket storage from a preview instance.
* Preview storage may not be available for some preview instances.

.. _console: https://console.cloud.intel.com/
