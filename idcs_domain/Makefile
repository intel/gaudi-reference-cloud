# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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

.PHONY: kind-deploy
kind-deploy: kind-deploy-operator kind-deploy-api-gateway ## Deploy all IDCS services in a new kind cluster.	

.PHONY: kind-deploy-operator
kind-deploy-operator: ## Deploy IDCS Operator in a new kind cluster.
	cd operator && POPULATE_SAMPLES=1 $(MAKE) kind-deploy

.PHONY: kind-deploy-api-gateway
kind-deploy-api-gateway: ## Deploy IDCS API Gateway in a new kind cluster.
	cd api_gateway && $(MAKE) kind-deploy

##@ Samples

KUBECTL_OPTIONS = --insecure-skip-tls-verify --cache-dir /dev/null
NAMESPACE1 = $(shell KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} create ns name-ignored -o jsonpath='{.metadata.name}')

.PHONY: require-KUBECONFIG1
require-KUBECONFIG1:
	@if [ "${KUBECONFIG1}" = "" ]; then \
		echo "Environment variable KUBECONFIG1 must be defined and point to a configuration file that connects to the IDC API Gateway."; \
		echo "See api_gateway/config/samples/kubeconfig-kind-api-gateway-user1.yaml."; \
		exit 1; \
	fi

.PHONY: test-samples
test-samples: clean-kube-cache require-KUBECONFIG1 ## Test with sample Kubernetes objects.
	@echo Testing create instance
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} apply -n ${NAMESPACE1} -f operator/config/samples/_v1alpha1_instance_my-tiny-vm-1.yaml
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} apply -n ${NAMESPACE1} -f operator/config/samples/_v1alpha1_instance_my-tiny-vm-2.yaml
	@echo Testing create sshpublickey
	operator/config/samples/generate_sshpublickeys.sh | KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} apply -n ${NAMESPACE1} -f -
	@echo Testing get single instance
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} get -n ${NAMESPACE1} instances my-tiny-vm-1
	@echo Testing list instances
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} get -n ${NAMESPACE1} instances
	@echo Testing watch instances
	-KUBECONFIG=${KUBECONFIG1} timeout 3s kubectl ${KUBECTL_OPTIONS} get -n ${NAMESPACE1} instances --watch
	@echo Testing delete single instance
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} delete -n ${NAMESPACE1} instances my-tiny-vm-2
	@echo Testing list instances
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} get -n ${NAMESPACE1} instances
	@echo Testing list sshpublickeys in namespace
	KUBECONFIG=${KUBECONFIG1} kubectl ${KUBECTL_OPTIONS} get -n ${NAMESPACE1} sshpublickeys
	@echo Testing successful

.PHONY: k9s1
k9s1: clean-kube-cache require-KUBECONFIG1 ## Run k9s as user 1
	KUBECONFIG=${KUBECONFIG1} k9s --insecure-skip-tls-verify -n ${NAMESPACE1} --logFile /tmp/k9s1.log --command instances

.PHONY: clean-kube-cache
clean-kube-cache: ## Clean Kubernetes client discover cache.
	rm -rf ${HOME}/.kube/cache ${HOME}/.kube/http-cache
