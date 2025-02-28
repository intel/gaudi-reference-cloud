
# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
# From https://stackoverflow.com/questions/10858261/how-to-abort-makefile-if-variable-not-set 
# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

$(call check_defined, HARVESTER_CLUSTER_NAME)
$(call check_defined, SECRETS_DIR, Example: export SECRETS_DIR=${shell pwd}/local/secrets/${IDC_ENV})
export HARVESTER_KUBECONFIG_DIR=$(SECRETS_DIR)/harvester-kubeconfig
export KUBECONFIG=$(HARVESTER_KUBECONFIG_DIR)/$(HARVESTER_CLUSTER_NAME)
export VALUES_FILE=${shell pwd}/deployment/helm/ssh-proxy-vm/values.yaml
export OVERRIDES_FILE=${shell pwd}/deployment/helm/ssh-proxy-vm/$(HARVESTER_CLUSTER_NAME).overrides.yaml

ID_RSA = secrets/ssh_host_rsa_key
ID_ECDSA = secrets/ssh_host_ecdsa_key
ID_ED25519 = secrets/ssh_host_ed25519_key
PROXY_SERVER_NAME ?= proxy1


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

##@ Deployment

.PHONY: redeploy
redeploy: undeploy deploy ## Redeploy SSH Proxy.

.PHONY: deploy
deploy: secrets $(KUBECONFIG) ## Deploy SSH Proxy.
	helm upgrade \
		--install \
		--set-file ssh.ssh_host_rsa_key=$(ID_RSA) \
		--set-file ssh.ssh_host_rsa_key_pub=$(ID_RSA).pub \
		--set-file ssh.ssh_host_ecdsa_key=$(ID_ECDSA) \
		--set-file ssh.ssh_host_ecdsa_key_pub=$(ID_ECDSA).pub \
		--set-file ssh.ssh_host_ed25519_key=$(ID_ED25519) \
		--set-file ssh.ssh_host_ed25519_key_pub=$(ID_ED25519).pub \
		--namespace proxy \
		--create-namespace \
		$(PROXY_SERVER_NAME) \
		-f $(VALUES_FILE) \
		-f $(OVERRIDES_FILE) \
		deployment/helm/ssh-proxy-vm 

.PHONY: undeploy $(KUBECONFIG)
undeploy: ## Undeploy SSH Proxy.
	-helm delete $(PROXY_SERVER_NAME) --namespace proxy --wait
	-kubectl -n proxy delete --wait virtualmachineinstance/$(PROXY_SERVER_NAME)
	-kubectl -n proxy delete --wait pvc/$(PROXY_SERVER_NAME)-disk-0

##@ Secrets
KEY_COMMENT = dev-ssh-proxy-$(shell uuidgen | head -c 8)@intel.com
.PHONY: secrets
secrets: $(ID_RSA) $(ID_RSA).pub $(ID_ECDSA) $(ID_ECDSA).pub $(ID_ED25519) $(ID_ED25519).pub ## Generate missing secrets.
$(ID_RSA) $(ID_RSA).pub:
	mkdir -p secrets
	ssh-keygen -f $(ID_RSA) -N '' -t rsa -C $(KEY_COMMENT)
	echo Key pair created on $(shell date -u) by $(USER)@$(shell hostname) with comment $(KEY_COMMENT) >> $(ID_RSA)
$(ID_ECDSA) $(ID_ECDSA).pub:
	ssh-keygen -f $(ID_ECDSA) -N '' -t ecdsa -C $(KEY_COMMENT)
	echo Key pair created on $(shell date -u) by $(USER)@$(shell hostname) with comment $(KEY_COMMENT) >> $(ID_ECDSA)
$(ID_ED25519) $(ID_ED25519).pub:
	ssh-keygen -f $(ID_ED25519) -N '' -t ed25519 -C $(KEY_COMMENT)
	echo Key pair created on $(shell date -u) by $(USER)@$(shell hostname) with comment $(KEY_COMMENT) >> $(ID_ED25519)
clean-secrets: ## Delete secrets.
	-rm $(ID_RSA) $(ID_RSA).pub $(ID_ECDSA) $(ID_ECDSA).pub $(ID_ED25519) $(ID_ED25519).pub
show-secrets: ## Show secrets.
	cat $(ID_RSA)
	cat $(ID_RSA).pub
	cat $(ID_ECDSA)
	cat $(ID_ECDSA).pub
	cat $(ID_ED25519)
	cat $(ID_ED25519).pub

.PHONY: show-config
show-config:  ## Show key Makefile configuration variables.
	@echo "show-config: Running with the following make configuration:"
	@echo "show-config:   HARVESTER_CLUSTER_NAME=$(HARVESTER_CLUSTER_NAME)"
	@echo "show-config:   SECRETS_DIR=$(SECRETS_DIR)"
	@echo "show-config:   VALUES_FILE=$(VALUES_FILE)"
	@echo "show-config:   OVERRIDES_FILE=$(OVERRIDES_FILE)"
	@echo "show-config:   KUBECONFIG=$(KUBECONFIG)"
