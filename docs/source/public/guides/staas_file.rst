.. _staas_file:

File Storage
############

File storage offers shared files across multiple systems, including
development environments, media stores, and user directories.

Prerequisites
*************

Complete these guides.

* :ref:`get_started`
* :ref:`manage_instance`

Create volume
*************

#. In the left side menu, select :command:`Storage > File Storage`.

#. In the :guilabel:`File Storage` tab, click :guilabel:`Create volume`.

#. Enter data in the following fields:

   - **Name**
   - **Storage size**

#. After data entry is accepted, click :guilabel:`Launch`.

Delete volume
*************

#. In the left side menu, select :command:`Storage > File Storage`.

#. In the :guilabel:`Actions` column, select :guilabel:`Delete`.

#. A dialog box requests confirmation.

#. Select :guilabel:`Delete` to proceed, or select :guilabel:`Cancel` to abort.

Mount volume
************

#. In the left side menu, select :command:`Storage > File Storage`.

#. Click the name of your storage volume to show details.

#. Select :guilabel:`How to mount`.

   Choose between *single instance* and *instance group*.

   .. note::
      You must have an instance available from which you
      can connect to the storage volume.

#. Click the :guilabel:`Instance` drop-down menu.

   Your instance should appear. Return to :ref:`manage_instance`
   if you need to create a new instance.

#. Choose :guilabel:`Mount your volume`.

#. Make sure to launch an instance from which you will mount a volume.

#. Follow the onscreen instructions to mount your volume.

#. After entering "login to the weka cluster", you\'re instructed to' "Enter your password at prompt".

#. In the :guilabel:`Security` tab, click :guilabel:`Generate password`.

#. In the pop-out dialog, click the copy button beside **Password**.

#. Paste that **password** into your Terminal instance.

   .. note::
      On most systems, use right-click, or CTRL+V, to paste the password.
      Once pasted, the password **will not appear**.

#. Press :guilabel:`<Enter>` to login.

#. Allow any downloads to complete as applicable.

#. Your Terminal output should show **Mount completed successfully**.

Verify Storage
**************

Assure that you're logged into your instance. Next, you can view mount points and perform basic file I/O operation.

View filesystem mount points:

.. code-block::  bash

   df -h

To view more details, like *Type* and *Size*, try:

.. code-block::  bash

   df -aTh

Change directory into the storage volume.

.. code-block::  bash

   cd /mnt/test

Suppose you want to download a sample dataset from the `UC Irvine Machine Learning Repository`_
In this example, we download `Online Retail dataset` as a CSV file.

.. code-block::  bash

   sudo curl -o data_online_retail.csv https://archive.ics.uci.edu/dataset/352/online+retail

FAQ
***

.. list-table::
   :header-rows: 1
   :class: table-tiber-theme

   * - Question
     - Answer

   * - Can I share file-system data across multiple instances?
     - Yes. You can mount the same storage on multiple instances to share the data.

   * - Do you recommend bare metal or virtual machine storage for specific use cases?
     - Mounting on bare metal will be slightly faster. Note: Mounting on a Guadi2 instance requires different mount commands.

   * - Is there a limit (e.g., quota) to *the number* or *the size* of storage volumes?
     - Yes. The number of instances available varies across account type and processor type. Generally, the number of instances available increases from Standard to Enterprise accounts.

       If you exceed the quota, please contact :ref:`support`. Enterprise users, please contact your representative.

.. _UC Irvine Machine Learning Repository: https://archive.ics.uci.edu/
.. _Online Retail dataset: https://archive.ics.uci.edu/dataset/352/online+retail
