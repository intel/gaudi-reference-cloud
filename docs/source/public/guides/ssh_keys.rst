.. _ssh_keys:

SSH Keys
########

This document describes our acceptable use policy for managing SSH Keys.

What is an ssh key?
*******************

A key is a set of security credentials that you use to prove your identity when connecting to an instance. Before launching an instance, you must specify a key.Credentials allow you to run a secure shell (SSH) session in your compute instance.

* You can choose an existing key or upload a new one.
* Depending on how you manage your security, you can specify the same key for all your instances, or you can specify different keys per instance.
* To connect to an instance using SSH, follow :ref:`connect_to_instance`.


View my account keys
********************

#. Sign into the console.

#. Next, click on the left side menu :guilabel:`Compute keys` icon in the sidebar menu (shown below).

.. note::

   Your keys will be shown. If your account does not have a key, a button will appear to help you upload a new key.

.. _gen_ssh_key_header:

Generate an SSH Key
*******************

.. _gen_ssh_key_start:

Launch a Terminal on your system to generate an SSH Key.
For MacOS\* systems, follow the Linux OS instructions.

Click the tab for your operating system (OS).

.. tabs::

   .. tab:: Linux\* OS

      #. Launch a Terminal on your local system.

      #. To generate an SSH key, copy and paste the following to your Terminal. 

         .. code-block:: bash

            ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa

      #. If you're prompted to overwrite, select No.

      #. Click the :guilabel:`copy` icon. Then paste this command in your Terminal to show the generated SSH key.

         .. code-block:: bash

            cat ~/.ssh/id_rsa.pub

      #. Continue below.

   .. tab:: Windows\* PowerShell

      #. Launch a new PowerShell window on your local system.

      #. Optional: If you haven\'t generated a key before, create an `.ssh` directory.

         .. code-block:: bash

            mkdir $env:UserProfile\.ssh

      #. Copy and paste the following to your terminal to generate SSH Keys

         .. code-block:: bash

            ssh-keygen -t rsa -b 4096 -f $env:UserProfile\.ssh\id_rsa

      #. If you are prompted to overwrite, select No.

      #. Click the :guilabel:`copy` icon. Then paste paste this command **to show** the generated SSH key.

         .. code-block:: bash

            cat $env:UserProfile\.ssh\id_rsa.pub

      #. Continue below.

.. _gen_ssh_key_end:

Upload an SSH Key
******************

.. _upload_ssh_key_start:

#. In the |ITAC|, click on :command:`Compute > Keys`.

#. In the :guilabel:`Keys` tab, click :guilabel:`Upload`.

#. Enter a name for your key in :guilabel:`Key name`.

#. Paste the contents of your SSH key in :guilabel:`Key contents`.

#. Select :guilabel:`Upload`.

#. Verify that your SSH key appears in the table.

.. _upload_ssh_key_end:

Delete an SSH Key
*****************

#. Go to the :guilabel:`Keys` tab.
#. Locate your key in the table.
#. Click :guilabel:`Delete` next to your key name.
#. Click :guilabel:`Delete` to confirm deletion.

.. note::
   You cannot delete a key for an instance that is currently running.
   Close the instance first. Then delete the SSH key.

For information on deleting SSH keys in shared instances, see :ref:`multi_user_accounts`.

Update an SSH Key
*****************

To update an SSH key while the instance is running, follow these steps.

#. Click :command:`Compute > Instances`.

#. In the console, click in the :guilabel:`Instances`.

#. For your desired instance, find the :guilabel:`Actions` column and click :guilabel:`Edit`.

#. From :guilabel:`Public keys` shown, select the key that you wish to update.

#. Save changes.

   .. note::
      Next, a dialog appears showing the instructions on how to complete configuration.

#. Follow the onscreen instructions to update your SSH key.

