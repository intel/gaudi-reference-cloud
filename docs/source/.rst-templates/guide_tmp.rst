:orphan:

Guide Title
###########

- Edit this `RST` template and replace this section with an introduction.
- Replace the `:orphan:`, above, with this syntax `.. _guide_name`, where `_guide_name` is a unique `:ref:` label.
  For help, see `Cross-referencing arbitrary locations`_.
- Replace `Guide Title` above with your unique title.
- Rename the file name to be unique.
- Add the new file name under the `.. toctree::` in the appropriate parent :file:`index.rst` file.
  For example, see :file:`source/public/guides/index.rst`.

If possible, in the introduction:

* Describe the product feature/function and its relationship to the whole.
* Define any product names or unique terms necessary to implement the solution.

The `.. contents::` directive creates a documment-level Table of Contents.
Keep the `.. contents::` directive if your document is long. Otherwise, remove it.

.. contents::
   :local:
   :depth: 1

System Requirements
*******************

If applicable, add system requirements. Otherwise, remove this section.

* Item 1
* Item 2
* Item 3

Instructions
************

* Add numbered steps for instructions if applicable.
* Use the `hash`, followed by a period `.` for each step.
* Integrate code-blocks and figures as shown in the following steps.

Otherwise, remove this section.

#. First, add import statement:

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
==========

This is a subsection. Use it or remove it.

Subsection
----------

This is a sub-subsection. Use it or remove it.

Complex Data
************

To show complex relational data, add a table (see below).
Use the list-table directive for any tables.

Otherwise, remove this section.

HTTP Methods
============

.. list-table::
   :header-rows: 1

   * - HTTP Method
     - Description

   * - GET
     - Retrieve an existing resource.

   * - POST
     - Create a new resource.

   * - PUT
     - Update an existing resource.

   * - PATCH
     - Partially update an existing resource.

   * - DELETE
     - Delete an existing resource.

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

To show **short snippets** of code, use the :command:`code-block::` directive.

The argument to the right of :command:`code-block::` is the programming language.
For most code, follow these guidelines:

* To show **terminal input**, use `bash` like so:

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
.. _Cross-referencing arbitrary locations: https://www.sphinx-doc.org/en/master/usage/restructuredtext/roles.html#cross-referencing-arbitrary-locations
