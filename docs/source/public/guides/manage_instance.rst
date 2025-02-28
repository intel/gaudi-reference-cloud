.. _manage_instance:

Manage Instance
################

Learn basic user operations to manage a compute instance after creating one. Available :ref:`Instance Types <instance_types>` are bare metal and virtual machines (VM). To view a full description of compute services, see :ref:`compute_svc`.

.. tip::
   Follow :ref:`stop_instance` to temporarily pause your instance.

.. _instance_states:

Instance States
****************

The :guilabel:`State` reflects the status of a compute instance.

Provisioning
============

#. After launching an instance, :guilabel:`State` shows **Provisioning** in the dashboard.
#. Wait until :guilabel:`State` shows :guilabel:`Ready`.

Ready
=====

#. When :guilabel:`State` shows **Ready**, you may launch your instance.
#. A connection is established with the host.
#. If desired, continue at :ref:`connect_to_instance`.

.. _stop_instance:

Stop Instance
*************

After you launch an instance and its :guilabel:`Status` shows :guilabel:`Ready`, you may wish to **stop the instance**.

#. In the :guilabel:`Instances` tab, scroll to show the :guilabel:`Actions` column.

#. Under :guilabel:`Actions`, click the :guilabel:`Stop` button to pause your instance.

   .. tip::
      When the confirmation dialog appears, follow the instructions.

#. Click :guilabel:`Stop` to pause your instance.

Restart Instance
****************

After your instance is stopped, it shows a :guilabel:`Start` button.

#. Click :guilabel:`Start` to restart your instance.

   .. tip::
      When the confirmation dialog appears, follow the instructions.

#. Click :guilabel:`Start`.

#. In the :guilabel:`State` column, your instance shows :guilabel:`Starting`.

#. Wait until your instance :guilabel:`State` shows :guilabel:`Ready` again.

#. Click on your instance under :guilabel:`Instance Name`.

#. Select :guilabel:`Details` to display options on how to connect.

Alternatively, choose a method to :ref:`connect_to_instance`.

Edit Instance
*************

#. Navigate to :command:`Compute > Instances` from main console.

#. In the :guilabel:`Instances` tab, assure that your instance appears.

#. With your instance, select :guilabel:`Edit` under :guilabel:`Actions`.

#. In the page "Edit Instance", modify settings as desired.


Delete an Instance
*******************

#. Navigate to :command:`Compute > Instances` from main console.

#. Under :guilabel:`Actions`, select the :guilabel:`Delete` button.

#. At the dialog "Delete instance", select :guilabel:`Delete` to confirm your choice.

#. Select :guilabel:`Cancel` if you do not wish to delete your instance.

Optional - Update OS and Add Packages
*************************************

While in an SSH session, you can add Ubuntu 22.04 packages and update your OS.

#. To update and upgrade your OS, enter one command at a time.

   .. code-block:: bash

      sudo apt-get update -y
      sudo apt-get upgrade -y

#. Add `net-tools` or `curl`

   .. code-block:: bash

      sudo apt-get install net-tools

   .. code-block:: bash

      sudo apt-get install curl

