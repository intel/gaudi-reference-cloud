:orphan:

Tutorial Title
##############

- Edit this `RST` template and replace this section with an introduction.
- Replace the `:orphan:`, above, by using this exact syntax `.. _tutorial_name`, where `_tutorial_name` is your document's unique name. 
  For help, see `Cross-referencing arbitrary locations`_.
- Replace `Tutorial Title` above with a unique title.
- Rename the filename to be unique.
- Add the new filename under the `.. toctree::` directive in the appropriate parent :file:`index.rst` file.
  To add a new tutorial in `public` docs, add file name in :file:`source/public/tutorials/index.rst`.

If possible, in the introduction, explain:

* How the Intel processor (or chipset) benefits the workload/stack
* Generally, **do not make performance claims** unless you have reliable, third-party documentation.

The `.. contents::` directive creates a documment-level Table of Contents.
Keep the `.. contents::` directive if your document is long. Otherwise, remove it.

.. contents::
   :local:
   :depth: 1

Compute Instance
****************

If applicable, include the “Compute Instance” details for the tutorial.
Use the following list-table directive for any tables.

Otherwise, remove this section.

.. list-table:: Compute Instance
   :widths: auto
   :header-rows: 1

   * - Instance type
     - Processor
     - Cores
     - RAM
     - Disk

   * - Bare Metal (BM)/ Virtual Machine (VM)
     - Intel processor brand name
     - N cores
     - N GB
     - N GB/TB

Prerequisites
*************

If applicable, add prerequisites required to run this tutorial. Otherwise, remove this section.
Most tutorials require adding :ref:`get_started` in list below. As applicable, include any third-party libraries.

* :ref:`get_started`
* Item 2
* Item 3

Instructions
************

* Add numbered steps `#.` for instructions if applicable.
* Integrate code-blocks or figures as shown in the examples below.

Otherwise, remove this section.

Numbered Steps - Example
========================

#. First, add an import statement:

   .. code-block:: python

      import tensorflow as tf
      print("TensorFlow version:", tf.__version__)

#. Next, see figure below.

   .. figure:: ../../_figures/_archive/python150.png
      :alt: Add image description here
      :align: center

      Add caption here.

.. note::
   See more examples of `Code blocks`_ and `Figures`_ below.

Subsection
===========

This is a subsection. Use it or remove it.

Sub-subsection
---------------

This is a sub-subsection. Use it or remove it.

Hyperlinks
**********

The  `Instructions`_ above reference code from a `Tensorflow Tutorial`_.
Did you catch the hyperlink in the previous sentence?

After making an in-line hyperlink (e.g. :command:`Tensoflow Tutorial_`, as above),
make its corresponding **hyperlink value** at the bottom of the page, like so:

.. code-block:: console


   .. _Tensorflow Tutorial: https://www.tensorflow.org/tutorials/quickstart/beginner

Code blocks
***********

To show **short snippets** of code, use the :command:`.. code-block::` directive.

The argument to the right of :command:`.. code-block::` is the programming language.
Alternatively, for command-line code, follow these examples:

* To show simple **terminal input**, use `bash` like so:

.. code-block:: bash

   echo "Hello World!" > hello-world.txt
   cat hello-world.txt

* To show **terminal ouput**, use `console` like so:

.. code-block:: console

   Hello World!

Long Code Samples
*****************

To show the entire code of a file, use the :command:`.. literalinclude::` directive.

Figures
*******

To add a figure:

* Use the :command:`.. figure::` directive as shown
* Place the new figure in root-level `_figures` directory in the `public` or `private` subdirectory.

.. figure:: ../../_figures/_archive/python150.png
   :alt: Add image description here
   :align: center

   Add caption here.

.. note::
   The filepath argument to `.. figure::` assumes all document figures are stored in `docs/_figures` directory .

.. _Tensorflow Tutorial: https://www.tensorflow.org/tutorials/quickstart/beginner
.. _Cross-referencing arbitrary locations: https://www.sphinx-doc.org/en/master/usage/referencing.html#cross-referencing-arbitrary-locations