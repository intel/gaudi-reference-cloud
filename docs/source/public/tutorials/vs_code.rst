.. _vs_code:

Visual Studio Code\* Dev
########################

Develop using the `Visual Studio Code app`_ on |ITAC|. All you need is the app, a GitHub account, and the `Remote Tunnels` extension.

Prerequisites
*************

* :ref:`get_started`
* GitHub\* account
* `Visual Studio Code app`_
* `Remote Tunnel` extension

Instructions
************

#. Open the |ITAC| console.

#. Click on :guilabel:`Learning`.

#. Select any Jupyter Notebook.

#. Click :guilabel:`Launch Jupyter notebook.`

#. Once in a Notebook, select :command:`File --> New --> Terminal`.

#. From the Terminal, download, unzip, and execute Visual Studio Code\* CLI.

   .. code-block:: bash

      curl -Lk 'https://code.visualstudio.com/sha/download?build=stable&os=cli-alpine-x64' --output vscode_cli.tar.gz

      tar -xzf vscode_cli.tar.gz

      ./code tunnel --accept-server-license-terms

   .. tip::
      Only run the above command **with flags** the first time.  Afterwards, run :code:`./code tunnel`


#. Upon successful execution, your Terminal should resemble the screen below.

   .. figure:: ../../_figures/tutorials/vs_code_00.PNG
      :alt: Visit URL displayed and grant access
      :align: center

      Visit URL displayed and grant access

   .. note::
      You can use a browser, but the app won\'t have the same level of functionality.

#. In Device Activation, enter the code. Then follow the prompts.

   .. figure:: ../../_figures/tutorials/vs_code_01.PNG
      :alt: Device Activation
      :align: center
      :scale: 50%

      Device Activation

   .. figure:: ../../_figures/tutorials/vs_code_02.PNG
      :alt: Congratulations
      :align: center
      :scale: 50%

      Congratulations

#. After successful authentication, create a **name** for your tunnel.
   Your Terminal should prompt: **"What would you like to call this machine?"**

   .. figure:: ../../_figures/tutorials/vs_code_03.PNG
      :alt: Name your Tunnel
      :align: center

      Name your Tunnel

   .. note::
      You can name the tunnel anything. We recommend keeping the name short.

#. Launch the Visual Studio Code app locally.  If you do not have the  extension `Remote Tunnels`, install it.

   .. figure:: ../../_figures/tutorials/vs_code_04.PNG
      :alt: Remote Tunnels extension
      :align: center

      Remote Tunnels extension

   a. To install `Remote Tunnels`, select the :guilabel:`Extensions` icon at left.

      .. figure:: ../../_figures/tutorials/vs_code_09.PNG
         :alt: Extensions icon
         :align: center

         Extensions icon

   #. In the :guilabel:`Search` field, type `Remote Tunnels`.

   #. Click on the extension.

   #. Click on :guilabel:`Install` in the Extension page.

#. You will be asked to authorize `Remote - Tunnels` using GitHub.  This typically only happens on the very first connection.

   .. figure:: ../../_figures/tutorials/vs_code_05.PNG
      :alt: Allow extension
      :align: center

      Allow extension

#. Follow the prompts. Depending on how many organizations you belong to, there may be several pages.
   You do not need to authorize any of your other organizations. Just allow Visual Studio Code to talk to your instance on |ITAC|.

   .. figure:: ../../_figures/tutorials/vs_code_06.PNG
      :alt: Approve access
      :align: center

      Approve access

#. From the app, enter :kbd:`F1`. In the field at top, type  `Remote - Tunnels: Connect to Tunnel`.
   Select the name that you gave your tunnel. This example uses "vscode1".

   .. figure:: ../../_figures/tutorials/vs_code_07.PNG
      :alt: Type F1 and select Remote Tunnel
      :align: center

      Type F1 and select Remote Tunnel

#. There will be one final authorization.  Don\'t forget to check the box or you will be asked again.

   .. figure:: ../../_figures/tutorials/vs_code_08.PNG
      :alt: Click button "Yes, I trust the authors"
      :align: center

      Select checkbox and click button "Yes, I trust the authors"

Now you should have a successful connection. In the future, just log in to |ITAC|, establish a tunnel, and connect using :kbd:`F1` and `Remote Tunnels`.

Next Steps
**********

If you choose to launch a Jupyter Notebook from Visual Studio Code, you may be asked:

* To install extensions or plugins (to support running the Notebook).
* To "Type to choose a kernel source". For the kernel, type "base".

Troubleshooting
***************

.. list-table::
   :widths: auto
   :header-rows: 1
   :class: table-tiber-theme

   * - Problem
     - Solution/Command
     - Comments

   * - Experiencing problems with Terminal running the Remote Tunnel?
     - :code:`./code tunnel unregister`
     - This stops currently registered tunnel. Follow instructions to restart a new tunnel.

   * - Encountering issues with a corporate VPN?
     - Log off of your corporate VPN
     - A corporate VPN may require special configuration; therefore, we recommend not using it. However, consult your sys admin.

Resources
*********

* See also `Developing with Remote Tunnels`_

.. _Visual Studio Code app: https://code.visualstudio.com/
.. _Developing with Remote Tunnels: https://code.visualstudio.com/docs/remote/tunnels
