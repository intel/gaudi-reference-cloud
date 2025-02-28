# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
# from pydata_sphinx_theme import index_toctree
import os
import sys
# sys.path.insert(0, os.path.abspath('..'))
sys.path.insert(0, os.path.abspath('../..'))

# -- Project information -----------------------------------------------------

project = 'Cloud Platform Docs'
copyright = '2023, Your Company'
author = 'Your Company'

# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
# extensions = ['sphinx.ext.autodoc']

# Add any paths that contain templates here, relative to this directory.
templates_path = ['_templates']

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store','CONTRIBUTING.rst']


# -- Options for HTML output -------------------------------------------------

master_doc = "index"

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = 'pydata_sphinx_theme'

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
# html_logo = '_static/assets/intel_logo_header.svg'
html_logo = '_static/assets/Alpha-Standalone-Icon-2.svg'

html_static_path = ['_static']

html_css_files = [
    'css/custom.css',
]

html_sidebars = {
    '**': ['search-field.html','sidebar-nav-bs.html',],
    # "index": [master_doc]
    # '**': ['globaltoc.html',]
}

#=================#
html_theme_options = {
"navbar_start": ["navbar-logo"],
"navbar_center": ["navbar-nav"],
# "navbar_end": ["version"],
"navigation_depth": 4,
# "show_toc_level": 2,
# "show_nav_level": 0,
"page_sidebar_items": ["page-toc"],
"show_prev_next": False ,
"navbar_align": "left",
"navbar_end" : ["login-button"],
# "navbar_end": release,
"collapse_navigation": True,
}

# version = release

rst_prolog = """
.. |IKS| replace:: Intel® Kubernetes Service
.. |INTC| replace:: Intel®
.. |INTG2SOLO| replace:: Intel® Gaudi® 2
.. |INTG2| replace:: Intel® Gaudi® 2 processor
.. |INTG2-ACC| replace:: Intel® Gaudi® AI accelerator
.. |INTG3| replace:: Intel® Gaudi® 3 processor
.. |ITDCSS| replace:: Intel® Tiber™ software and services
.. |ITDCS| replace:: Intel® Tiber™ solutions
.. |ITAC| replace:: Intel® Tiber™ AI Cloud
.. |ITAIS| replace:: Intel® Tiber™ AI Studio
.. |ITSS| replace:: Intel® Tiber™ software and services
.. |IXP| replace:: Intel® Xeon® processor family
.. |IXP3| replace:: 3rd Generation Intel® Xeon® Scalable Processor family
.. |IXP4| replace:: 4th Generation Intel® Xeon® Scalable Processor family
.. |GPUFLX| replace:: Intel® Data Center GPU Flex Series
.. |GPUMAX| replace:: Intel® Data Center GPU Max Series
.. |HBGAUD| replace:: Habana® Gaudi® Processor Training and Inference using OpenVINO™ Toolkit for U-Net 2D Model
"""
