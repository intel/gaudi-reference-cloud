# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
# Removed import statement for pydata theme that previously appeared on following line
import os
from os import environ
import sys
import time
from pathlib import Path
from jinja2 import Environment, FileSystemLoader, Template
from datetime import datetime
import json
# sys.path.insert(0, os.path.abspath('..'))
sys.path.insert(0, os.path.abspath('../..'))

# -- Project information -----------------------------------------------------

proj_selektor = os.environ.get('PROJECT', 'public')

if proj_selektor == 'private':
    print("\nUsing private templates.")
else:
    print("\nUsing public templates.")

project = 'Cloud Platform Docs'
copyright = '2023, Your company'
author = 'Your company'

# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.

project_extensions = []
if proj_selektor != 'private':
    project_extensions = ['nbsphinx', 'collectfieldnodes', 'sphinxcontrib.bibtex', 'sphinxcontrib.mermaid',
                          'createinstancespecs', 'json2table']

mermaid_use_local = '/docs/_static/js/mermaid/dist/mermaid.esm.min.mjs'
mermaid_version = '11.2.0'
mermaid_elk_use_local = '/docs/_static/js/@mermaid-js/layout-elk/dist/mermaid-layout-elk.esm.min.mjs'
mermaid_include_elk = '0.1.4'
d3_use_local = '/docs/_static/js/d3/dist/d3.min.js'
d3_version = '7.9.0'

extensions = ['sphinx.ext.autodoc',
              "sphinx_design",
              'sphinx_tabs.tabs',
              'multiproject',
              'sphinx_copybutton',
              ] + project_extensions


bibtex_bibfiles = ['reference/refs.bib']

bibtex_reference_style = "author_year"
myst_fence_as_directive = ["mermaid"]

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store']

# -- Options for HTML output -------------------------------------------------

# master_doc = "index"

multiproject_projects = {
    "public": {
        "path": "public",
    },
    "private": {
        "path": "private",
    },
}


# Add any paths that contain templates here, relative to this directory.
# templates_path = ['_templates', '_templates_private']

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.

html_theme = 'pydata_sphinx_theme'

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_logo = '_static/assets/intel_logo_header.svg'

html_static_path = ['_static']

html_css_files = [
    'css/custom.css',
    "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.1.1/css/all.min.css",
]


# -- JupyterLab Software Dependencies table --------------

table_data_file = 'jupyterlab.json'
jupyterlab_sw = json.load(open('jupyterlab.json'))
html_context = {'jupyter_sw' : jupyterlab_sw}


# Expose path, prepended with _static, for javascript.

html_js_files = ['js/custom.js',]

# Custom sidebar templates, maps document names to template names.

html_sidebars = {
    # '**': ['sidebar-nav-bs.html'],
    '**': ['globaltoc.html']
}

# Enable future customization with JavaScript or other


def setup(app):
    app.add_js_file('js/custom.js')
