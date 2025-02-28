# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line, and also
# from the environment for the first two.

SPHINXOPTS    ?=
SPHINXBUILD   ?= sphinx-build
SOURCEDIR     = .
BUILDDIR      = _build
IMAGEDIRS     = _images
ALLSPHINXOPTS   = -d $(BUILDDIR)/doctrees $(SPHINXOPTS) .
ERROROPTS       = -W --keep-going
.DEFAULT_GOAL := help
.PHONY: help

# Put it first so that "make" without argument is like "make help".
help:
	@echo
	@echo "Use \e[1;34mcustom build targets$<\e[0m."
	@echo
	@echo "Use 'make \e[1;34mtarget$<\e[0m' where \e[1;34mtarget$<\e[0m is one of:"
	@echo
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)##\(.*\)/\1\3/p' \
	| column -t  -s ' '
	@echo
	@echo Other defaults
	@echo
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).

html: ## to build the docs
	$(SPHINXBUILD) $(ERROROPTS) -b html $(ALLSPHINXOPTS) $(BUILDDIR)/html
ifneq ('$(PROJECT)', 'private')
	npm run post-build
endif
	@echo
	@echo "Build finished. The HTML pages are docs/source/$(BUILDDIR)/html."
	@echo

clean: ## to clean the docs
	rm -rf $(BUILDDIR)/*
	@echo "Directories cleaned."
	@echo

guide: ## to make guide from RST template
	python3 _scripts/make_guide.py
	@echo
	@echo "Guide template saved to /docs/source/public/guides."
	@echo

tut: ## to make tutorial from RST template
	python3 _scripts/make_tut.py
	@echo
	@echo "Tutorial template saved to /docs/source/public/tutorials."
	@echo

tutdemo: ## to make tutorial demo
	python3 _scripts/tut_demo.py
	@echo
	@echo "1. Double-click index.html in \033[1msource/_build/html\033[0m.\n2. In browser, click \033[1mTutorial Title\033[0m page to view Tutorial template."
	@echo
	@echo "Demo complete. Run \033[1mmake tutclean\033[0m to clean directory."

tutclean: ## to clean tutorial demo
	rm -rf $(BUILDDIR)/*
	@echo
	python3 _scripts/cln_demo.py	
	@echo
