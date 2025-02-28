# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

FORCE_CLEAN ?= false
MUNGE_KEY = munge.key

.PHONY: all
all: help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: install-requirements
install-requirements: ## Install requirements
	sudo apt install openssh-client ansible ansible-lint python3-pip
	pip3 install --ignore-installed PyYAML
	pip3 install -q -r requirements.txt
	ansible-galaxy install -r requirements.yml
	ansible-galaxy collection install kubernetes.core cloud.common

##@ Ansible Playbooks

.PHONY: basic_workers
basic_workers: ## Deploy a generic cluster
	cd playbooks && ansible-playbook -e force_clean=$(FORCE_CLEAN) basic_workers.yml

.PHONY: slurm
slurm: secrets ## Deploy a Slurm cluster
	cd playbooks && ansible-playbook -e force_clean=$(FORCE_CLEAN) -e slurm_munge_key=../$(MUNGE_KEY) slurm.yml

##@ Secrets
.PHONY: secrets
secrets: $(MUNGE_KEY) ## Generate missing secrets.
$(MUNGE_KEY):
	dd if=/dev/urandom bs=1 count=128 status=none > $(MUNGE_KEY)
clean-secrets: ## Delete secrets.
	-rm $(MUNGE_KEY)
show-secrets: ## Show secrets.
	hexdump $(MUNGE_KEY)
