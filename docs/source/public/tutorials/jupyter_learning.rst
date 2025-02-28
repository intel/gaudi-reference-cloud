.. _jupyter_learning:

JupyterLab Learning
###################

Run popular AI/ML tutorials in a JupyterLab\* environment on |ITAC|.

.. tip::
   Click :guilabel:`Learning` in the console main menu to view Jupyter Notebooks.

Prerequisites
*************

* :ref:`get_started`

Overview
********

A JupyterLab\* environment is provided for users who want to study, create, and share executable code. By default, Jupyter Notebooks are pre-loaded with popular libraries and tools, including |INTC| software and |INTC| hardware. See `Hardware and software`_.

#. From the console main menu, select :guilabel:`Learning`.

#. Choose an option.

   * `Launch Jupyter Notebook`_ -- **Most common**
      * Launch the default :guilabel:`Python 3 (ipykernel)`
      * Dependencies are loaded via :command:`import` statements
      * Pre-existing code is included

   * `Launch JupyterLab`_
      * Launch the default :guilabel:`Python 3 (ipykernel)`
      * Notebooks launch as bare, *without* pre-loaded modules or code.

Launch Jupyter Notebook
***********************

#. Select a :guilabel:`Jupyter Notebook` from a category, such as **AI with Intel® Gaudi® AI accelerator**.

   .. note::
      This is the most common option for students and developers.

#. Select the :guilabel:`Launch` button.

#. Follow the instructions in the Notebook.

#. Place your cursor in the cell to execute code.

#. Press :kbd:`SHIFT+ENTER` to run code and review output.

You're all set.

Launch JupyterLab
*****************

#. Click on the button :guilabel:`Launch JupyterLab`.

#. At the dialog, "Your JupyterLab access is ready", select :guilabel:`Launch`.

#. At the prompt :guilabel:`Select Kernel`, choose `base` to run the Notebook.

#. In a cell, enter your code and import modules.

#. Place your cursor in the cell to execute code.

#. Select :kbd:`SHIFT+ENTER` to run code and review output.

Notebook Tips
*************

* To change cell behavior, select the :guilabel:`Code` pull-down menu at top.
  Optional: Change to another option, like :guilabel:`Markdown` for text, or :guilabel:`Raw`.

* To change the kernel, select the pull-down menu in upper right.
  When :guilabel:`Select kernel` appears, select your preferred kernel. See also `Jupyter kernels`_.

Hardware and software
*********************

Hardware
========

The JupyterLab environment provides access to various types of hardware, including:

* Intel® Gaudi® 2
* Intel® Gaudi® AI Accelerator
* Intel® Data Center GPU Max Series

..
   @stevenfollis, decide if your intent is: Intel® Gaudi® 2 processor, or the Intel® Gaudi® AI Accelerator
   The intel product names databases distinguishes between these two things.

Hardware access operates in a shared model where each user accesses a portion of CPU, GPU, and memory resources.
For dedicated or exclusive access, use bare metal or virtual machine instances. See :ref:`manage_instance`.

Software
=========

The JupyterLab environment includes a variety of commonly used software to help accelerate learning.
The table below shows a subset of the available |INTC| software installed in the JupyterLab environment.

.. render-json-table::

Kernel options
==============

Jupyter utilizes a kernel model to allow different configurations of software that can be selected for a given Notebook.

Each notebook has a default kernel. Automatically, a kernel runtime is set for each Notebook upon launch. However, you can select a different kernel in a Notebook after you launch it. For a comprehensive list of Jupyter kernels, visit `Jupyter documentation`_

.. list-table:: Available kernels
   :header-rows: 1
   :class: table-tiber-theme

   * - Kernel
     - Intel® Max Series GPU
     - Intel® Gaudi® 2 AI Accelerator

   * - Base
     - ✓
     -

   * - Modin
     - ✓
     -

   * - Python 3
     - ✓
     - ✓

   * - PyTorch
     - ✓
     -

   * - PyTorch 2.6
     - ✓
     -

   * - PyTorch GPU
     - ✓
     -

   * - TensorFlow
     - ✓
     -

   * - TensorFlow GPU
     - ✓
     -

Sessions
********

A session is an environment provided to a single user to interact with Jupyter for a set duration of time. To conserve resources, the system will terminate a session:

* If you explicitly log out by clicking “File” on the Jupyter top navigation and selecting “Log Out.”
* If your Jupyter session is idle for 15 minutes the system will automatically log you out of a session.
* When the maximum 4-hour session duration is elapsed, the session is logged out.

You may create a new session at any time by following the same access instructions.

Session lifecycle
==================

Accounts that have not used Jupyter for 20 days are considered inactive and may have all local file contents removed. Each new session will reset this rolling 20-day timer, and users who have had their environments removed can always launch a notebook again.


Updating libraries
==================

Libraries are updated regularly to ensure you have access to the latest versions. If you require a specific library or version that does not exist, we encourage you to post on in our `Community page`_.

.. note::
   New libraries and tools are frequently added to the JupyterLab environment. Such dependencies are subject to change without notice.

User-installed packages
=======================

Optionally, you may install your own software via the Jupyter terminal. For example, you may do so by downloading binary files, cloning code repositories, or utilizing a package manager such as :command:`conda` or :command:`pip`. When executed, such code is installed by default in the :file:`~/.local` directory. Similarly, when Jupyter notebook guide users to pip install required libraries, these libraries are installed into :file:`~/.local` directory.

Troubleshooting
****************

If you encounter issues with the JupyterLab environment, try one of the following.

Restart Kernel
==============

In the Notebook, select the :guilabel:`Kernel` menu, and choose an action (e.g., "Restart kernel").

.. caution::
   Take care to save or back up the code entered in cells. Choosing some menu options may delete code depending on its state.

Logout
======

#. Select :guilabel:`File`` on the top navigation bar and then select :guilabel:`Log Out`.
#. Relaunch the Notebook by clicking the :guilabel:`Launch`.
#. Evaluate if a relaunch corrects the issue.

Policy
******

**The standard terms and conditions of the** `Intel® Tiber™ AI Cloud Service Agreement`_ **apply**.

**Fees** JupyterLab resources are currently provided free of charge.  However, you acknowledge and agree that this policy is subject to change without notice solely applies to JupyterLab resources, and not any other |ITAC| services, software, or professional services.

**Use Restriction.** As articulated in the Intel® Tiber™ AI Cloud Service Agreement, use of the JupyterLab resources for illegal or malicious activities is strictly prohibited, and Intel reserves the right to suspend or remove access at any time.

.. _Intel® Tiber™ AI Cloud Service Agreement: https://www.intel.com/content/www/us/en/content-details/785964/content-details.html
.. _Community page: https://community.intel.com/t5/Intel-Developer-Cloud/bd-p/developer-cloud
.. _Jupyter kernels: https://docs.jupyter.org/en/latest/projects/kernels.html
.. _Jupyter documentation: https://github.com/jupyter/jupyter/wiki/Jupyter-kernels
