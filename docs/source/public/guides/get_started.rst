.. _get_started:

Get Started
###########

Sign up for |ITAC|. Sign in. Launch a compute instance.

.. tip::
   To explore the compute platform in a learning node only, try :ref:`quick_start`.

Create Account, Sign In
***********************

#. Visit the |ITAC| `Console`_.

#. Choose :guilabel:`Create an Account` or :guilabel:`Sign in`.

   .. tabs::

      .. tab:: Create an Account

         Create an account with your email address. This becomes the account ID.

         #. Click :guilabel:`Create an Account`.

         #. Follow instructions in the dialog and complete required fields.

         #. Sign in with the email address you used to register your account.

      .. tab:: Sign In

         #. View the previous figure.

         #. In the :guilabel:`Email` field, enter your email address.

         #. Click :guilabel:`Sign In`.

         #. In the next screen, enter your password and press :guilabel:`<Enter>`

.. _initiate_instance_start:

Initiate Instance
*****************

#. Log in to |ITAC|.

#. In the left side menu, click :guilabel:`Compute -> Instances`.

#. In the :guilabel:`Instances` tab, click :guilabel:`Launch instance`.

Launch Instance
***************

Next, complete a form named :guilabel:`Launch a compute instance`.

#. From :guilabel:`Instance family`, select your choice.

#. From :guilabel:`Instance Type`, select your choice.

   .. note::
      Optionally, select :guilabel:`Compare instance types` to compare instances' specifications.

#. From :guilabel:`Machine image`, select your choice.

   .. tip::
      This option may be pre-selected based on your choice of :guilabel:`Instance family`.

#. In :guilabel:`Instance name`, enter a name in lowercase. Optional: Use hyphens.

#. You may use **One-Click connection**. If yes, skip to the next step.

   a. Continue below if you wish to upload an SSH Key.

   #. Under :guilabel:`Public Keys`, select the checkbox for your public key.

      .. tip::
         If you wish **to connect using SSH public keys**, :ref:`follow the instructions <set_up_ssh_keys>` below.

#. **One-Click connection**. Select the radio button to **enable single-click access** for future use.

   .. note::
      Selecting One-Click connection is only performed once. You cannot add or remove this functionality after you launch an instance.

#. Click :guilabel:`Launch`.

#. Continue in next section.

.. _connect_to_instance:

Connect to Instance
********************

#. View your instance in the :guilabel:`Instances` tab.

#. Assure your instance :guilabel:`State` shows **Ready**.

   .. tip::
      See :ref:`instance_states` for more details.

#. Choose a method to connect to your instance:

   * :ref:`one_click_header`
   * :ref:`Local Terminal Connection<set_up_ssh_keys>`

.. _one_click_header:

One-Click Connection
====================

Using this method, you connect to your instance in a JupyterLab environment.

#. Navigate to the :guilabel:`Instances` tab.

#. Wait until your instance name shows :guilabel:`Ready`.

#. Click on your instance name.

#. Click :guilabel:`Connect`.

   a. Optional: Click on the :guilabel:`Connect` button in the row where your instance appears. A JupyterLab environment will launch.

   .. tip::
      You must select "One-Click connection" before launching an instance for this option to be available.

#. Select the :guilabel:`Terminal` icon to access your instance.

#. You're all set. Start exploring.

.. _set_up_ssh_keys:

Set Up SSH Keys
===============

Configure SSH keys in two steps. First, create an SSH key locally. Second, upload it to your account. Setting up an SSH Key is a one-time task.

.. warning::

   Never share your private keys with anyone. Never create a SSH Private key without a passphrase.

Create an SSH Key
-----------------

.. include:: ssh_keys.rst
   :start-after: gen_ssh_key_start:
   :end-before: gen_ssh_key_end:

Upload an SSH Key
-----------------

.. include:: ssh_keys.rst
   :start-after: upload_ssh_key_start:
   :end-before: upload_ssh_key_end:

Local Terminal Connection
=========================

#. **Connect via Terminal** from your local machine.

   a. Under :guilabel:`Instance Name`, click on your instance name.

   #. Select :guilabel:`How to Connect via SSH`.

   #. A new pop-up dialog appears:"How to connect to your instance".

   #. Continue in next section.

#. Follow the onscreen instructions in the dialog.

   a. Select your operating system (OS).

   #. Follow instructions.

   #. Copy the command shown to connect to your instance.

   #. Open a Terminal.

#. In the Terminal, paste the command you copied and press enter.

#. If prompted to add your public key, select `Yes`.

#. After launching instance, run command to confirm Ubuntu 22.04 (or other).

   .. code-block:: bash

      cat /etc/os-release

Next Steps
**********

* Learn how to :ref:`manage your instance <manage_instance>`
* Apply a :ref:`load_balancer` to your instance
* Start exploring use cases in :ref:`tutorials`.

Optional Steps
**************

Access via Corporate Network
============================

If you're connecting to |ITAC| from your company Corporate Network, you'll need to update SSH config file.

.. note::

   If you connect using PowerShell on Microsoft Windows Operating System, you must install `gitforwindows`.

#. SSH Configuration is an one-time task.

#. Your SSH configuration file is in a folder named :file:`.ssh` under your user's home folder. If the file is not present, create one.

#. Copy and paste the following to SSH config file:

   .. code-block:: console

      ~/.ssh/config

   .. tabs::

      .. tab:: Linux\* OS

         .. code-block:: console

            Host 146.152.*.*
            ProxyCommand /usr/bin/nc -x PROXYSERVER:PROXYSPORT %h %p

      .. tab:: Windows

         .. code-block:: console

            Host 146.152.*.*

            ProxyCommand "C:\Program Files\Git\mingw64\bin\connect.exe" -S PROXYSERVER:PROXYSPORT %h %p 

#. From your Lab Administrator, get PROXYSERVER and PROXYPORT in your Corporate Network for SSH, NOT for HTTP/HTTPS Proxy.

#. Replace PROXYSERVER and PROXYPORT with the information you received from your lab administrator and save the SSH Config file.

.. _Console: https://console.cloud.intel.com/
