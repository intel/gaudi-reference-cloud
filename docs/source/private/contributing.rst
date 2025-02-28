.. _contributing:

Contributing
############

This page explains how to contribute to the |ITAC| Docs. The target audience is Intel contributors.

.. tip::
   This document only appears in the :file:`source/private` directory.

Prerequisites
*************

* `Python` 3.8+
* `Sphinx` 5.3.0
* `pip` 23.1.2+

#. To install Python, visit `python.org/downloads`_. Use Python3.8 or above for these docs.

   .. note::
      See also the `Python3 Installation and Setup Guide`_ fromm Real Python.

#. Install Sphinx on Linux OS, MacOS, or Windows.

   .. tabs::

      .. tab:: Linux\* OS / macOS\*

         .. code-block:: bash

            python -m pip install -U sphinx==5.3.0

      .. tab:: PowerShell

         .. code-block:: bash

            python -m pip install --user sphinx==5.3.0

#. On Windows OS, add the python script folder to your PATH system variable

#. Verify that the version of Sphinx installed matches the one found in :file:`requirements.txt`

   .. code-block:: bash

      sphinx-build --version

#. Verify ``pip`` is installed. A file path should appear in `stdout`.

   .. code-block:: bash

      python -m pip --version


#. If ``pip`` is not installed, install/upgrade it.

   .. tabs::

      .. tab:: Linux\* OS / macOS\*

         .. code-block:: bash

            python3 -m pip install --user --upgrade pip

      .. tab:: PowerShell

         .. code-block:: bash

            python.exe -m pip install --user --upgrade pip

Develop docs in venv
*********************

Use the python `venv` module, available by default in Python3.8. Learn more about using `Python virtual environments`_.

.. note::
   Using the `venv` helps you easily manage multiple versions of Python. It also allows you to manage Sphinx package dependencies **without impacting** your local development environment.

#. In your CLI, navigate to the `/docs` directory.

   .. code-block:: bash

      cd docs

#. Create a python virtual environment.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            python3 -m venv .venv

      .. tab:: PowerShell

         .. code-block:: bash

            py -m venv .venv

#. For Windows OS only, set policy.

   .. code-block:: bash

      Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

#. Activate the `venv`.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            source .venv/bin/activate

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .venv/Scripts/activate.ps1

#. Install Sphinx dependencies in the `venv`.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            python3 -m pip install -r requirements.txt --no-cache-dir

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            py -m pip install -r .\requirements.txt --no-cache-dir

   During development in PowerShell, you may disconnect from VPN if proxy issues occur.


#. Create or edit an existing :file:`RST` document.

   .. tip::
      See also the `Tutorial Demo`_ below to use a Template to create a document.

#. Run these commands to validate a successful build of new/existing docs in Sphinx for the public version.
   If you encounter errors, resolve the errors before submitting a pull request.

#. Build the `public` version of the documentation.

   .. note::
      As of December 2024, `Mermaid.js` was added to support **workflow diagrams** in public documentation.
      To include workflow diagrams, use the full command below. Otherwise, you may use :command:`make html`. It defaults to the public version.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            PROJECT=public NODE_ENV=dev make html

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat html

   .. tip::
      When you run `make html`, by default, you build the `public` version of the documentation.
      Follow the next step to build the private version of the documentation.

#. Build the `private` version of the documentation.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            PROJECT=private make html

      .. tab:: Windows/PowerShell

         .. code-block:: powershell

            $env:PROJECT = 'private'; .\make.bat html

   .. note::
      To **explicitly** build the `public` version, replace the string 'private', above,
      with 'public'.

#. Clean the `public` or `private` version of the documentation.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            make clean

      .. tab:: Windows/PowerShell

          .. code-block:: bash

             .\make.bat clean

#. Navigate to the :file:`source/_buid/html/` directory and double-click the :file:`index.html` to view docs site.

   a. You may also run a local python development server to view the docs site.

      .. code-block:: bash

         python3 -m http.server 8080 --bind 127.0.0.1

#. Clean the docs.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            make clean

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat clean

#. Deactivate the venv.

   .. code-block:: bash

      deactivate

#. Remove the venv.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            rm -rf .venv

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            Remove-Item .venv

         Enter Y to confirm.

.. note::
   To add figures or images, skip to `Figures and Images`_.

Tutorial Demo
****************

This demo shows you **how to create and integrate a tutorial** using reSTructuredText (RST).

If want to create a new tutorial, follow `Add New Tutorial`_.

What happens under the hood?:

* The :file:`tutorial_tmp` template is copied to the :file:`source/public/tutorials/` directory.
* The string :file:`tutorial_tmp` is inserted below the `toctree` directive in :file:`source/public/tutorials/index.rst`

You must enter **both commands**, per OS, before quitting tutorial:

.. list-table:: Tutorial demo commands
   :widths: auto
   :header-rows: 1

   * - Purpose
     - Linux\* OS/ MacOS\*
     - PowerShell\*/Windows\* OS

   * - Build tutorial demo
     - :command:`make tutdemo`
     - :command:`.\make.bat tutdemo`

   * - Clean tutorial demo
     - :command:`make tutclean`
     - :command:`.\make.bat tutclean`

#. Demo the *process of adding a new tutorial* using the tutorial template.

   .. tabs::

      .. tab:: Linux OS/ MacOS

         .. code-block:: bash

            make tutdemo

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat tutdemo

#. Clean the tutorial demo.

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            make tutclean

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat tutclean

Add New Guide
*************

#. Start a new **guide** using our guide template.

   .. note::
      The guide template is saved to :file:`source/public/guides` by default.
      Move the copied template to a different directory if you desire.

#. Create a new guide using the template:

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            make guide

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat guide

#. Follow the instructions in the template, whose title is "Guide Title".

#. Add the :file:`filename` of the guide under the :command:`.. toctree::` in the appropriate :file:`index.rst` file.

#. Remember to add the new, and modified, files using Git.


Add New Tutorial
****************

#. Start a new tutorial using our tutorial template.

   .. note::
      The tutorial template is saved to :file:`source/public/tutorials` by default.
      Move the copied template to a different directory if you desire.

#. Create a new tutorial using the template:

   .. tabs::

      .. tab:: Linux\* OS/macOS\*

         .. code-block:: bash

            make tut

      .. tab:: Windows/PowerShell

         .. code-block:: bash

            .\make.bat tut

#. Follow the instructions in the template, whose title is "Tutorial Title".

#. Add the :file:`filename` of the tutorial under the :command:`.. toctree::` in the appropriate :file:`index.rst` file.

#. Remember to add the new, and modified, files using Git.

.. tip::
   **Choose carefully** where you save the tutorial, whether in the `public` or `private` directories.

Figures and Images
*******************

Place all new figures (e.g., graphics, images) in the :file:`_figures` directory, as shown below.

* The :file:`_figures` sub-directory structure reflects all *common* subdirectories
  in the :file:`public` and :file:`private` directories.

Follow these steps for adding new or editing existing images.

Add image
=========

* Place the new figure (or image) in its associated subdirectory (e.g., "_figures/guide", "_figures/tutorials", etc.).

  - For example, the first ``.PNG`` file for guides is :file:`_figures/guides/get_started_00.png`

* Name the figure\'s filename using the :file:`ref` target, which occurs at the top of any RST document.

  - The :file:`ref` target becomes the figure\'s prefix.

  - For example, :file:`_figures/guides/get_started_00.png` is the first figure that appears in :file:`get_started.rst`, whose :file:`ref`
    target is :file:`.. _get_started`.  Therefore, its filename prefix is **get_started_**. The **00** is added because it\'s the first figure that appears in the document.

.. note::
   If a new figure must **appear only the private version** of documents, place that image in its :file:`private` subdirectory.

*  Name the new figure\'s filename **suffix** using the next available number.

   - For example, if adding a new figure in :file:`get_started.rst`, rename a new filename as :file:`_figures/guides/get_started_07.png`
     if :file:`_figures/guides/get_started_06.png` were the last known figure.

* Update the *relative file-path* argument to the :file:`.. figure::` in your RST document.
  The filepath must reflect the correct subdirectory inside :file:`_figures`.

  Example:

  .. code-block:: console

     .. figure:: ../../_figures/guides/get_started_00.png
        :alt: |ITAC| Sign-In
        :align: center
        :scale: 75%

        |ITAC| Sign-In

Edit image
==========

* Edit existing images, or versions thereof, and keep them in their original directory.
  **DO NOT** create new subdirectories within the :file:`_figures` directory.

* After any :file:`_figures` subdirectory has one or more images, remove the :file:`.gitkeep` in Git.


.. _GitHub Workflows Developer: https://internal-placeholder.com/pages/viewpage.action?spaceKey=InnerSource&title=InnerSource+WG+Documentation
.. _python.org/downloads: https://www.python.org/downloads/
.. _Python3 Installation and Setup Guide: https://realpython.com/installing-python/
.. _Python virtual environments: https://docs.python.org/3/library/venv.html
.. _PowerShell - Windows Proxy: https://internal-placeholder.com/display/proxy/Windows+Proxy+Environment+Variables
.. _Linux OS Proxy: https://internal-placeholder.com/display/GTAIGKLabs/How+to+set+up+proxy+in+Linux+and+Windows
.. _Proxy at Intel: https://intelpedia.intel.com/Proxy_at_Intel#Proxy_Environment_Variables_to_set
.. _Build private docs in Linux: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/README.md#build-private-docs-in-linux
