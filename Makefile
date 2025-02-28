###############################################################################
# Main IDC Makefile
###############################################################################

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

COMMA := ,

# Get directory that this Makefile is located in.
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Determine the IDC environment (IDC_ENV).
# If IDC_ENV is already set, then use that.
# Otherwise, if a directory that matches the hostname exists in build/environment, then use that.
# Otherwise, use "kind-singlecluster".
ifeq ($(IDC_ENV),)
HOSTNAME := $(shell hostname)
ifneq "$(wildcard build/environments/$(HOSTNAME))" ""
export IDC_ENV=$(HOSTNAME)
else
export IDC_ENV=kind-singlecluster
endif
endif

# IDC_ENV_DIR can be used for any environment-specific files.
export IDC_ENV_DIR ?= build/environments/${IDC_ENV}

# Load environment-specific Makefiles.
INCLUDE_FILES = \
	$(IDC_ENV_DIR)/Makefile.environment \
	local/environments/$(IDC_ENV)/Makefile.environment
-include $(INCLUDE_FILES)

TMPDIR := $(shell mktemp -d)

export GOPATH ?= $(ROOT_DIR)/local/go

# Ensure that binaries installed with "go install" are in the PATH.
ORIGINAL_PATH := $(PATH)
ifeq (,$(findstring :$(GOPATH)/bin, $(PATH)))
export PATH := $(GOPATH)/bin:$(PATH)
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
GOOGLE_API_DIR ?= public_api/proto/google/api

## Tool Versions
ARGOCD_VERSION ?= v2.9.3
# Releases are listed at https://github.com/metal3-io/baremetal-operator/releases
METAL3_BAREMETAL_OPERATOR_VERSION ?= 0.3.1
BAZELISK_VERSION ?= v1.22.0
GRPCURL_VERSION ?= 1.8.7
TIMESCALE_CHART_VERSION ?= 0.27.5
POSTGRES_CHART_VERSION ?= 12.2.6
# Tag "v6.2.1" plus Intel patch
OPENAPI_GENERATOR_TAG ?= @sha256:793fa1835cc9816a8bd905cd6bf64e915987c2c87ffd3dce485e2dd7a35cab71
CONTROLLER_TOOLS_VERSION ?= v0.16.4
CODEGEN_VERSION ?= 0.25.2
KUSTOMIZE_VERSION ?= v4.5.5
KUBECTL_VERSION ?= v1.24.9
# Releases are listed at https://github.com/helm/helm/releases
HELM_VERSION ?= v3.16.3
HELMFILE_VERSION ?= 0.169.2
VAULT_VERSION ?= 1.13.1
# Releases are listed at https://github.com/protocolbuffers/protobuf/releases
PROTOC_VERSION ?= 21.9
# Releases are listed at https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go
PROTOC_GEN_VERSION ?= 1.28
# Releases are listed at https://pkg.go.dev/google.golang.org/grpc/cmd/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION ?= 1.2
# Releases are listed at https://pkg.go.dev/github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
GRPC_GATEWAY_VERSION ?= 2.12.0
# Releases are listed at https://github.com/golang/mock/releases/
MOCKGEN_VERSION ?= v1.6.0
OPAL_VERSION ?= 0.5.0
OPAL_HELM_VERSION ?= 0.0.11
OPAL_CLIENT_DOCKER_TAG ?= b98db071b2bf60f98eebf75de4ff474c68baf887532fb52a2d3a55bc23c5811e
PROTO_VALIDATE_VERSION ?= v0.10.1
TERRAFORM_VERSION ?= 1.8.5
JQ_VERSION ?= 1.7.1
YQ_VERSION ?= v4.35.2
GOLANGCI_LINT_VERSION ?= 1.63.4

## Tool Binaries
ARGOCD = $(LOCALBIN)/argocd-$(ARGOCD_VERSION)
AWS = $(LOCALBIN)/aws
BAZELISK = $(LOCALBIN)/bazelisk-$(BAZELISK_VERSION)
BAZEL = $(BAZELISK)
PROTOC = $(HOME)/.local/bin/protoc
GRPCURL = $(LOCALBIN)/grpcurl-$(GRPCURL_VERSION)
CONTROLLER_GEN = $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
CODEGEN_DIR = $(LOCALBIN)/code-generator-$(CODEGEN_VERSION)
CODEGEN_GENERATE_GROUPS = $(CODEGEN_DIR)/generate-groups.sh
KUSTOMIZE = $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
HELM = $(LOCALBIN)/helm-$(HELM_VERSION)
HELMFILE = $(LOCALBIN)/helmfile-$(HELMFILE_VERSION)
VAULT = $(LOCALBIN)/vault-$(VAULT_VERSION)
KUBECTL = $(LOCALBIN)/kubectl
TERRAFORM = $(LOCALBIN)/terraform-$(TERRAFORM_VERSION)
JQ = $(LOCALBIN)/jq-$(JQ_VERSION)
YQ = $(LOCALBIN)/yq-$(YQ_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

# When deploying kind, a local Docker registry will be started that listens on this port.
KIND_REGISTRY_PORT ?= 5001
LOCAL_REGISTRY_NAME ?= idc-registry-$(KIND_REGISTRY_PORT).$(shell id -un).intel.com

# Push container images to this repo.
export DOCKER_REGISTRY ?= localhost:$(KIND_REGISTRY_PORT)

# When run in Jenkins, we must use the GIT_BRANCH and GIT_URL provided in the environment to determine these values.
# Otherwise, we can use the standard git commands.
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GIT_URL ?= $(shell git remote get-url origin)

# IDC version numbers
GIT_REV_COUNT := $(shell git rev-list --count HEAD)
GIT_SHORT_HASH := $(shell git rev-parse --short=8 HEAD)
export GIT_COMMIT := $(shell git rev-parse HEAD)
IDC_SEMANTIC_VERSION ?= 0.0.1
IDC_FULL_VERSION ?= $(IDC_SEMANTIC_VERSION)-$(GIT_COMMIT)

# Container images will have this prefix and tag.
DOCKER_IMAGE_PREFIX ?=
# Note that images used by Bazel-built Helm charts will be referenced using the SHA-256, not the tag.
export DOCKER_TAG ?= $(IDC_SEMANTIC_VERSION)-$(GIT_COMMIT)

# Helm settings.
# HELM_CHART_VERSION is the prefix. It will have a hash appended to it.
#   - The idc-versions Helm chart version will include the Git commit hash and look like: 0.0.1-d693c958428c1bc500f3f2a95f657d28eb11bd1a
#   - All other Helm chart versions will include a SHA-256 hash of the Helm chart and look like: 0.0.1-4ea989504fdda81d57ab36ebbf19bf938113d1d96753ff69c4b4fa6ae76c9e93
export HELM_CHART_VERSION ?= $(IDC_SEMANTIC_VERSION)

# Complete Helm chart versions will be downloaded as individual files to this directory with "make download-helm-chart-versions".
export HELM_CHART_VERSIONS_DIR ?= $(ROOT_DIR)/local/idc-versions/$(IDC_FULL_VERSION)/idc-versions/chart_versions
HELM_REGISTRY ?= localhost:$(KIND_REGISTRY_PORT)
HELM_PROJECT ?= intelcloud
HELMFILE_OPTS += --helm-binary $(HELM)

# Define resource attributes for traces when running tests.
OTEL_DEPLOYMENT_ENVIRONMENT ?= $(shell hostname)
export OTEL_RESOURCE_ATTRIBUTES ?= deployment.environment=$(OTEL_DEPLOYMENT_ENVIRONMENT)

# Disable proxy for addresses that end in .local.
ifeq (,$(findstring $(COMMA).local, $(no_proxy)))
export no_proxy := $(no_proxy),.local
endif
export NO_PROXY := $(no_proxy)

# Deployment configuration and secrets.
export REGION ?= us-dev-1
export AVAILABILITY_ZONE ?= us-dev-1a
export SECRETS_DIR ?= local/secrets
export SECRETS_DIR := $(abspath $(SECRETS_DIR))
export DB_ADMIN_USERNAME = postgres
export DB_USER_USERNAME = dbuser
export GRAFANA_ADMIN_USERNAME = admin
export HARVESTER_KUBECONFIG_DIR = $(SECRETS_DIR)/harvester-kubeconfig
export HARVESTER_KUBECONFIG = $(HARVESTER_KUBECONFIG_DIR)/harvester1
export SSH_PROXY_OPERATOR_ID_RSA = $(SECRETS_DIR)/ssh-proxy-operator/id_rsa
export VAULT_BACKUP_DIR = $(SECRETS_DIR)/vault-backup
export SSH_PROXY_SERVER_HOST_PUBLIC_KEY = $(SECRETS_DIR)/ssh-proxy-operator/host_public_key
export BM_INSTANCE_OPERATOR_ID_RSA = $(SECRETS_DIR)/bm-instance-operator/id_rsa
export BM_VALIDATION_OPERATOR_ID_RSA = $(SECRETS_DIR)/bm-validation-operator/id_rsa
export BM_INSTANCE_OPERATOR_HOST_PUBLIC_KEY = $(SECRETS_DIR)/bm-instance-operator/host_public_key
export VAULT_ADDR ?= http://dev.vault.cloud.intel.com.kind.local:80
export VAULT_TOKEN_FILE ?= $(SECRETS_DIR)/VAULT_TOKEN
export HOST_IP := $(shell ip -o route get to 10.248.2.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')
export KIND_API_SERVER_ADDRESS ?= $(HOST_IP)
export ARIA_CLIENT_ID_PREFIX ?= $(USER)
HELMFILE_ARGOCD_VALUES_DIR ?= $(SECRETS_DIR)/helm-values
HELMFILE_DUMP_YAML ?= $(SECRETS_DIR)/helmfile-dump.yaml
IKS_KUBERNETES_OPERATOR_CONFIG = $(SECRETS_DIR)/kubernetes-operator
IKS_KUBERNETES_RECONCILER_CONFIG = $(SECRETS_DIR)/kubernetes-reconciler
IKS_ILB_OPERATOR_CONFIG = $(SECRETS_DIR)/ilb-operator
BILLING_DRIVER_ARIA_SQS_ACCESS_KEY_ID = $(SECRETS_DIR)/billing_driver_aria_sqs_access_key_id
BILLING_DRIVER_ARIA_SQS_SECRET_ACCESS_KEY = $(SECRETS_DIR)/billing_driver_aria_sqs_secret_access_key
BILLING_DRIVER_ARIA_SQS_ACCOUNT_ID = $(SECRETS_DIR)/billing_driver_aria_sqs_account_id
SSH_PROXY_USER ?= guest-${USER}
# Helm releases matching this regex will be deleted when running "make upgrade-all-in-kind-v2".
# Example: us-dev-1-compute-api-server|us-dev-1-compute-db
DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE ?=
DEPLOY_ALL_IN_KIND_IDC_SERVICES_DEPLOYMENT_METHOD ?= argocd

# Universe Deployer configuration
UNIVERSE_DEPLOYER_BUILDS_PER_HOST ?= 1
UNIVERSE_DEPLOYER_JOBS_PER_PIPELINE ?= 1
UNIVERSE_DEPLOYER_POOL_DIR ?= /tmp
TEST_UNIVERSE_CONFIG_FILE ?= universe_deployer/environments/testing1.json
UNIVERSE_CONFIG ?= universe_deployer/environments/${IDC_ENV}.json
RUN_MANIFESTS_GENERATOR_GIT_COMMIT ?= $(GIT_COMMIT)

# OS http server port
HTTP_SERVER_PORT ?= 50001
GUEST_HOST_MEMORY_MB ?= 16384
GUEST_HOST_DEPLOYMENTS ?= 3

# Matrix for Docker registry values:
#                           DOCKER_REGISTRY                       DOCKERIO_REGISTRY                    DOCKERIO_REPOSITORY_PREFIX
# Inside Intel, kind        localhost:5001                        internal-placeholder.com          cache/
# Inside Intel, Jenkins     internal-placeholder.com           internal-placeholder.com          cache/
# Staging w/o Harbor, kind  localhost:5001                        docker.io
# Staging w/ Harbor         amr-idc-registry-pre.infra-host.com   amr-idc-registry-pre.infra-host.com  cache/

ifeq ($(DOCKERIO_REGISTRY),)
DOCKERIO_REGISTRY ?= internal-placeholder.com
DOCKERIO_REPOSITORY_PREFIX ?= cache/
endif

# internal-placeholder.com/cache/ OR docker.io/
export DOCKERIO_IMAGE_PREFIX ?= $(DOCKERIO_REGISTRY)/$(DOCKERIO_REPOSITORY_PREFIX)

# Same as above without trailing "/".
# internal-placeholder.com/cache OR docker.io
ifeq ($(DOCKERIO_REPOSITORY_PREFIX),)
DOCKERIO_REGISTRY_WITH_PREFIX ?= $(DOCKERIO_REGISTRY)
else
DOCKERIO_REGISTRY_WITH_PREFIX ?= $(DOCKERIO_REGISTRY)/$(DOCKERIO_REPOSITORY_PREFIX:/=)
endif
# We are using an Intel patched version
OPENAPI_DOCKER_PREFIX ?= amr-idc-registry-pre.infra-host.com/intelcloud/

# Calculated variables.
UNAME_ARCH := $(shell uname -p)
ifeq ($(UNAME_ARCH), aarch64)
  IDC_ARCH=arm64
  BAZELISK_ARCH := arm64
  HELM_ARCH := arm64
  HELMFILE_ARCH := arm64
  VAULT_ARCH := arm64
  OPAL_SERVER_HELM_OPTS := --set imageRegistry=localhost:$(KIND_REGISTRY_PORT) \
	--set postgresImageRegistry=arm64v8
  OPAL_CLIENT_DOCKER_REPOSITORY := permitio/opal-client-standalone
  OPAL_DOCKER_REGISTRY := localhost:$(KIND_REGISTRY_PORT)
  OPAL_CLIENT_DOCKER_TAG := $(OPAL_VERSION)
  # Tag "latest" as of 2023-03-16
  POSTGRES_IMAGE_TAG ?= @sha256:50a96a21f2992518c2cb4601467cf27c7ac852542d8913c1872fe45cd6449947
else
  IDC_ARCH=x86_64
  BAZELISK_ARCH := amd64
  HELM_ARCH := amd64
  HELMFILE_ARCH := amd64
  VAULT_ARCH := amd64
  OPAL_SERVER_HELM_OPTS :=
  OPAL_CLIENT_DOCKER_REPOSITORY := permitio/opal-client-standalone@sha256
  OPAL_DOCKER_REGISTRY := $(DOCKERIO_REGISTRY_WITH_PREFIX)
  # Tag "latest" as of 2023-03-10
  POSTGRES_IMAGE_TAG ?= @sha256:fbcec7ba704eb8c846b2a265953fa8faaa562fb172e6670018e35da77ef06057
endif

LOCAL_BAZEL_REMOTE_CACHE_NAME ?= bazel-remote-cache

BAZEL_STARTUP_OPTS ?=
BAZEL_EXTRA_OPTS ?=
BAZEL_OPTS = $(BAZEL_EXTRA_OPTS)

HELMFILE_ENVIRONMENT ?= $(IDC_ENV)

HELMFILE_ENVIRONMENT_FILES = \
	deployment/helmfile/environments/{prod,staging}.yaml.gotmpl \
	deployment/helmfile/environments/{prod,staging}-region-*.yaml.gotmpl

DEPLOY_ALL_IN_KIND_OPTS = \
	//go/pkg/universe_deployer/cmd/deploy_all_in_kind:deploy_all_in_kind -- \
	--bazel-binary $(BAZEL) \
	--commit $(shell git rev-parse HEAD) \
	--docker-image-prefix "$(DOCKER_IMAGE_PREFIX)" \
	--helm-project "$(HELM_PROJECT)" \
	--idc-env $(IDC_ENV) \
	--idc-services-deployment-method $(DEPLOY_ALL_IN_KIND_IDC_SERVICES_DEPLOYMENT_METHOD) \
	--local-registry-name $(LOCAL_REGISTRY_NAME) \
	--local-registry-port $(KIND_REGISTRY_PORT) \
	--secrets-dir $(SECRETS_DIR) \
	--semantic-version $(IDC_SEMANTIC_VERSION) \
	--temp-dir local/deploy-all-in-kind \
	--universe-config ${UNIVERSE_CONFIG}

###############################################################################
# DEFINE TARGET GROUPS FOR DEPLOY-ALL-IN-KIND (V1).
###############################################################################

# Containers built by Bazel.
BAZEL_CONTAINERS = \
	$(BAZEL_CONTAINERS_WITH_HELM) \
	$(BAZEL_CONTAINERS_WITHOUT_HELM)

# Containers built by Bazel with a Helm chart of the same name.
# Containers are built using Bazel.
# Helm charts are pushed using Bazel.
# Helm charts are deployed using Helmfile.
BAZEL_CONTAINERS_WITH_HELM = \
	$(BAZEL_CONTAINERS_FOUNDATION) \
	$(BAZEL_CONTAINERS_FINANCE) \
	$(BAZEL_CONTAINERS_COMPUTE) \
	$(BAZEL_CONTAINERS_STORAGE) \
	$(BAZEL_CONTAINERS_NETWORKING) \
	$(BAZEL_CONTAINERS_TRAINING) \
	$(BAZEL_CONTAINERS_MISC) \
	$(BAZEL_CONTAINERS_IKS) \
	$(BAZEL_CONTAINERS_INSIGHTS) \
	$(BAZEL_CONTAINERS_IKS_OPERATORS) \
	$(BAZEL_CONTAINERS_KFAAS) \
    $(BAZEL_CONTAINERS_CLOUDMONITOR) \
	$(BAZEL_CONTAINERS_CLOUDMONITOR_LOGS) \
	$(BAZEL_CONTAINERS_DATALOADER) \
	$(BAZEL_CONTAINERS_DPAI) \
	$(BAZEL_CONTAINERS_MAAS) \
	$(BAZEL_CONTAINERS_QUOTA_MANAGEMENT_SERVICE)

BAZEL_CONTAINERS_FOUNDATION = \
	cloudaccount \
	authz \
	user-credentials \
	grpc-rest-gateway \
	grpc-reflect \
	oidc

BAZEL_CONTAINERS_FINANCE = \
	cloudaccount-enroll \
	productcatalog \
	productcatalog-operator \
	usage \
	metering \
	cloudcredits \
	billing-standard \
	billing-aria \
	billing-intel \
	billing \
	billing-schedulers \
	notification-gateway \
	trade-scanner \
	cloudcredits-worker

BAZEL_CONTAINERS_COMPUTE = \
	baremetal-enrollment-api \
	baremetal-enrollment-operator \
	bm-instance-operator \
	bm-validation-operator \
	compute-api-server \
	compute-metering-monitor \
	firewall-operator \
	fleet-node-reporter \
	git-to-grpc-synchronizer \
	instance-replicator \
	loadbalancer-replicator \
	loadbalancer-operator \
	intel-device-plugin \
	ssh-proxy-operator \
	vm-instance-scheduler \
	vm-instance-operator

BAZEL_CONTAINERS_NETWORKING = \
	sdn-controller \
	sdn-controller-rest \
	sdn-integrity-checker \
	sdn-vn-controller \
	switch-config-saver

BAZEL_CONTAINERS_TRAINING = \
	training-api-server \
	armada

BAZEL_CONTAINERS_MISC = \
	console

BAZEL_CONTAINERS_IKS = \
	iks

BAZEL_CONTAINERS_INSIGHTS = \
	security-insights \
	kubescore \
	security-scanner

BAZEL_CONTAINERS_IKS_OPERATORS = \
	ilb-operator \
	kubernetes-operator \
	kubernetes-reconciler

BAZEL_CONTAINERS_KFAAS = \
     kfaas

BAZEL_CONTAINERS_CLOUDMONITOR = \
     cloudmonitor

BAZEL_CONTAINERS_CLOUDMONITOR_LOGS = \
    cloudmonitor-logs-api-server

BAZEL_CONTAINERS_DPAI = \
     dpai

BAZEL_CONTAINERS_MAAS = \
     infaas-dispatcher \
     infaas-inference \
     maas-gateway

BAZEL_CONTAINERS_STORAGE = \
	object-store-operator \
	bucket-metering-monitor \
	storage-api-server \
	storage-operator \
	vast-storage-operator \
	vast-metering-monitor \
	storage-replicator \
	bucket-replicator \
	storage-scheduler \
	storage-kms \
	storage-metering-monitor \
	storage-user \
	storage-custom-metrics-service \
	storage-admin-api-server \
	storage-resource-cleaner

BAZEL_CONTAINERS_QUOTA_MANAGEMENT_SERVICE = \
	quota-management-service

BAZEL_CONTAINERS_DATALOADER = \
	dataloader

# Containers built by Bazel without a Helm chart of the same name.
# Containers are built using Bazel.
# Helm charts are not pushed using Bazel.
# Helm charts are not deployed using Helmfile.
BAZEL_CONTAINERS_WITHOUT_HELM = \
	opa

# Helm charts that can be pushed (with Bazel).
HELM_CHARTS = \
	$(NON_BAZEL_CUSTOM_CHARTS) \
	$(BAZEL_CONTAINERS_FOUNDATION) \
	$(BAZEL_CONTAINERS_FINANCE) \
	$(BAZEL_CONTAINERS_COMPUTE) \
	$(BAZEL_CONTAINERS_STORAGE) \
	$(BAZEL_CONTAINERS_NETWORKING) \
	$(BAZEL_CONTAINERS_TRAINING) \
	$(BAZEL_CONTAINERS_MISC) \
	$(BAZEL_CONTAINERS_IKS) \
	$(BAZEL_CONTAINERS_INSIGHTS) \
	$(BAZEL_CONTAINERS_IKS_OPERATORS) \
	$(BAZEL_CONTAINERS_KFAAS) \
	$(BAZEL_CONTAINERS_CLOUDMONITOR) \
	$(BAZEL_CONTAINERS_CLOUDMONITOR_LOGS) \
	$(BAZEL_CONTAINERS_DATALOADER) \
	$(BAZEL_CONTAINERS_DPAI) \
	$(BAZEL_CONTAINERS_MAAS) \
	$(BAZEL_CONTAINERS_QUOTA_MANAGEMENT_SERVICE)

# Helm charts deployed using Helmfile.
HELMFILE_CHARTS = \
	$(HELM_CHARTS) \
	$(THIRD_PARTY_CHARTS)

# IDC custom Helm charts.
# Containers are not built using Bazel.
# Helm charts are pushed using Bazel.
# Helm charts are deployed using Helmfile.

NON_BAZEL_CUSTOM_CHARTS = \
	argo-cd-resources \
	compute-crds \
	debug-tools \
	productcatalog-crds \
	metal3-crds \
	baremetal-enrollment-task \
	baremetal-operator-ns \
	baremetal-operator \
	dhcp-proxy \
	metallb-custom-resources \
	bm-dnsmasq \
	idcs-init-k8s-resources \
	idcs-istio-custom-resources \
	grpc-proxy \
	nginx-s3-gateway\
	netbox \
	netbox-azuread-sso\
	tftp-server \
	database-creator \
	ilb-crds \
	loadbalancer-crds \
	firewall-crds \
	network-crds \
	kubernetes-crds \
	sdn-controller-crds \
	sdn-restricted-sa \
	idc-versions \
	local-path-provisioner \
	rate-limit \
	rate-limit-redis \
	vm-machine-image-resources

# 3rd-party Helm charts.
# Containers are not built using Bazel.
# Helm charts are not pushed using Bazel.
# Helm charts are deployed using Helmfile.
THIRD_PARTY_CHARTS = \
	argo-cd \
	coredns \
	gitea \
	opal \
	postgresql \
	timescaledb \
	cert-manager \
	external-secrets \
	metallb \
	grafana \
	minio \
	localstack \
	cloudnative-pg \
	operator \
	tenant

# Dependencies for container-build.
# This builds the Helm chart, which depends on the container image.
CONTAINER_BUILD_DEPS = \
	$(BAZEL_CONTAINERS_WITH_HELM:%=helm-build-%) \
	$(BAZEL_CONTAINERS_WITHOUT_HELM:%=container-build-%)

# Dependencies for container-push.
CONTAINER_PUSH_DEPS = \
	$(BAZEL_CONTAINERS:%=retry-container-push-%) \

# Dependencies for helm-push.
HELM_PUSH_DEPS = \
	$(HELM_CHARTS:%=retry-helm-push-%)

# Dependencies for deploy-all-in-kind.
DEPLOY_ALL_IN_KIND_DEPS = \
	show-make-config \
	secrets \
	deploy-registry \
	deploy-kind \
	deploy-k8s-infrastructure-helm-releases \
	deploy-vault \
	deploy-k8s-tls-secrets \
	deploy-idc \
	create-localstack-resources

# Dependencies for deploy-all-in-kind-v2.
DEPLOY_ALL_IN_KIND_V2_DEPS = \
	show-make-config \
	update-etc-hosts \
	deploy-registry

# Metal3 namespaces for vault roles
export BAREMETAL_OPERATOR_NAMESPACES ?= metal3-1 metal3-2

# Dependencies for deploy-metal-in-kind.
DEPLOY_METAL_IN_KIND_DEPS ?= \
	show-make-config \
	secrets \
	add-api-server-to-no-proxy \
	secrets \
	teardown-bmvs \
	deploy-kind \
	deploy-registry \
	deploy-k8s-infrastructure-helm-releases \
	deploy-vault \
	deploy-baremetal-operator-ns \
	deploy-cert-manager \
	deploy-external-secrets \
	deploy-metal3-crds \
	deploy-metallb \
	deploy-metallb-custom-resources \
	deploy-baremetal-operator \
	deploy-bm-dnsmasq \
	deploy-dhcp-proxy \
	setup-bmvs \
	deploy-baremetal-enrollment-api \
	deploy-baremetal-enrollment-task \
	deploy-netbox-kind \
	deploy-idc

DEPLOY_METAL_IN_RKE2_DEPS := \
	secrets \
	deploy-rke2 \
	deploy-k8s-infrastructure-helm-releases \
	deploy-vault \
	deploy-baremetal-operator-ns \
	deploy-cert-manager \
	deploy-external-secrets \
	deploy-metallb \
	deploy-metallb-custom-resources \
	deploy-metal3-crds \
	deploy-baremetal-operator \
	deploy-baremetal-enrollment-api \
	deploy-baremetal-enrollment-task \
	deploy-netbox-kind \
	deploy-idc

# Dependencies for deploy-idc.
DEPLOY_IDC_DEPS = \
	container-and-chart-push \
	helm-chart-versions \
	deploy-crds \
	deploy-all-helm-releases

# Dependencies for deploy-foundation.
DEPLOY_FOUNDATION_DEPS = \
	deploy-cloudaccount-db \
	deploy-opal \
	$(BAZEL_CONTAINERS_FOUNDATION:%=deploy-%) \
	deploy-grpc-proxy

# Dependencies for deploy-finance.
DEPLOY_FINANCE_DEPS = \
	deploy-metering-db \
	deploy-cloudcredits-db \
	deploy-billing-db \
	deploy-usage-db \
	deploy-notification-db \
	deploy-productcatalog-db \
	deploy-productcatalog-crds \
	$(BAZEL_CONTAINERS_FINANCE:%=deploy-%)

# Dependencies for deploy-compute.
DEPLOY_COMPUTE_DEPS = \
	deploy-compute-db \
	deploy-compute-crds \
	$(BAZEL_CONTAINERS_COMPUTE:%=deploy-%)

# Dependencies for deploy-storage.
DEPLOY_STORAGE_DEPS = \
	deploy-storage-db \
	$(BAZEL_CONTAINERS_STORAGE:%=deploy-%)

DEPLOY_TRAINING_DEPS = \
	deploy-training-db \
	$(BAZEL_CONTAINERS_TRAINING:%=deploy-%)

DEPLOY_IKS_DEPS = \
	$(BAZEL_CONTAINERS_IKS:%=deploy-%)

DEPLOY_INSIGHTS_DEPS = \
	deploy-insights-db \
	$(BAZEL_CONTAINERS_INSIGHTS:%=deploy-%)

DEPLOY_KFAAS_DEPS = \
  deploy-kfaas-db \
	$(BAZEL_CONTAINERS_KFAAS:%=deploy-%)

DEPLOY_CLOUDMONITOR_DEPS = \
  deploy-cloudmonitor-db \
	$(BAZEL_CONTAINERS_CLOUDMONITOR:%=deploy-%)

DEPLOY_CLOUDMONITOR_LOGS_DEPS = \
  	deploy-cloudmonitor-logs-db \
	$(BAZEL_CONTAINERS_CLOUDMONITOR_LOGS:%=deploy-%)

DEPLOY_DPAI_DEPS = \
  deploy-dpai-db \
	$(BAZEL_CONTAINERS_DPAI:%=deploy-%)

DEPLOY_IKS_OPERATORS_DEPS = \
	deploy-ilb-crds \
	deploy-kubernetes-crds \
	$(BAZEL_CONTAINERS_IKS_OPERATORS:%=deploy-%)

DEPLOY_DATALOADER_DEPS = \
    deploy-dataloader \
    $(BAZEL_CONTAINERS_DATALOADER:%=deploy-%)

DEPLOY_MISC_DEPS = \
	sdn-controller-crds \
	sdn-controller \
	sdn-controller-rest \
	sdn-restricted-sa \
	sdn-integrity-checker \
	switch-config-saver \
	deploy-console


DEPLOY_QUOTA_MANAGEMENT_DEPS = \
	deploy-quota-management-service-db \
	$(BAZEL_CONTAINERS_QUOTA_MANAGEMENT_SERVICE:%=deploy-%)


# Postgres databases.
DBS = \
	authz \
	billing \
	cloudcredits \
	cloudaccount \
	metering \
	$(REGION)-compute \
	$(REGION)-dpai \
	$(REGION)-insights \
	$(REGION)-storage \
	$(REGION)-training \
	$(REGION)-iks \
	$(REGION)-kfaas \
	$(REGION)-cloudmonitor \
	$(REGION)-cloudmonitor-logs \
	$(REGION)-quota-management-service

SECRETS = \
	$(IKS_KUBERNETES_OPERATOR_CONFIG) \
	$(IKS_KUBERNETES_RECONCILER_CONFIG) \
	$(IKS_ILB_OPERATOR_CONFIG)

###############################################################################
# CALCULATE BAZEL TEST TARGETS
###############################################################################

GO_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/...) \
except attr("tags", "manual", //go/...)')

BAZEL_TEST_TARGETS = \
	$(GO_TEST_TARGETS) \
	//deployment/universe_deployer:universe_config_tests

GO_NON_GINKGO_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/...) \
except rdeps(//go/..., @com_github_onsi_ginkgo_v2//:ginkgo) \
except attr("tags", "manual", //go/...)')

GINKGO_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/...) \
intersect rdeps(//go/..., @com_github_onsi_ginkgo_v2//:ginkgo) \
except attr("tags", "manual", //go/...)')

BAZEL_LARGE_JENKINS_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/...) \
intersect attr("size", "large", //go/...) \
intersect attr("tags", "jenkins", //go/...)')

BAZEL_ENORMOUS_JENKINS_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/test/compute/e2e/vm/...) \
intersect attr("size", "enormous", //go/test/compute/e2e/vm/...) \
intersect attr("tags", "jenkins", //go/test/compute/e2e/vm/...)')

BAZEL_ENORMOUS_JENKINS_BM_TEST_TARGETS = $(shell $(BAZEL) query '\
tests(//go/test/compute/e2e/bm/...) \
intersect attr("size", "enormous", //go/test/compute/e2e/bm/...) \
intersect attr("tags", "jenkins", //go/test/compute/e2e/bm/...)')

###############################################################################
# RULES
###############################################################################

.PHONY: all
all: help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: bazel generate-go generate-proto generate-go-openapi generate-go-openapi-raven generate-go-openapi-raven-sdn-private generate-k8s fmt tidy gazelle generate-copyright-headers ## Generate all generated files that should be committed.

delete-generated-files-go: ## Deleted output of Go code generators.
	rm -rf go/pkg/pb/*.{pb,gw,validate}.go

.PHONY: generate-go
generate-go: bazel protoc install-go-packages mockgen delete-generated-files-go ## Generate Go code.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- generate ./...

.PHONY: generate-proto
generate-proto: bazel ## Generate Protobuf files for IDE support.
	$(BAZEL) run $(BAZEL_OPTS) //go/pkg/storage/storagecontroller:update_gen

.PHONY: generate-copyright-headers
generate-copyright-headers: ## only ./go folder for now
	python3 ./hack/generate_copyright.py --dir ./go

# See https://github.com/OpenAPITools/openapi-generator/blob/master/docs/generators/go.md
.PHONY: generate-go-openapi
generate-go-openapi: ## Generate Go OpenAPI clients.
	rm -rf go/pkg/compute_api_server/openapi
	docker run --rm -v $(shell pwd):/local \
		-u $(shell id -u $(USER)):$(shell id -g $(USER)) \
		$(OPENAPI_DOCKER_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
		generate \
			--input-spec /local/public_api/proto/compute.swagger.json \
			--generator-name go \
			--output /local/go/pkg/compute_api_server/openapi
	rm go/pkg/compute_api_server/openapi/go.{mod,sum}
	rm -rf go/pkg/compute_api_server/openapi/test

# The input spec has been modified in order to generate the API client.
# See https://internal-placeholder.com/spec/devcloud for the latest spec.
.PHONY: generate-go-openapi-raven
generate-go-openapi-raven: ## Generate Go OpenAPI clients for Raven
	rm -rf go/pkg/raven/openapi
	docker run --rm -v $(shell pwd):/local \
		-u $(shell id -u $(USER)):$(shell id -g $(USER)) \
		$(OPENAPI_DOCKER_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
		generate \
	--input-spec /local/go/pkg/raven/swagger.json \
	--generator-name go \
			--output /local/go/pkg/raven/openapi
	rm -rf go/pkg/raven/openapi/go.{mod,sum}
	rm -rf go/pkg/raven/openapi/test
	rm -rf go/pkg/raven/openapi/git_push.sh
	rm -rf go/pkg/raven/openapi/.travis.yml

# The openapi spec below is for SDN only and not listed in https://internal-placeholder.com/spec/devcloud.
.PHONY: generate-go-openapi-raven-sdn-private
generate-go-openapi-raven-sdn-private: ## Generate Go OpenAPI Raven client for SDN-Controller.
	docker run --rm -v $(shell pwd):/local \
		-u $(shell id -u $(USER)):$(shell id -g $(USER)) \
		$(OPENAPI_DOCKER_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
		generate \
	--input-spec /local/go/pkg/sdn-controller/pkg/raven/openapi.yaml \
	--generator-name go \
			--output /local/go/pkg/sdn-controller/pkg/raven/openapi
	rm -rf go/pkg/sdn-controller/pkg/raven/openapi/go.{mod,sum}
	rm -rf go/pkg/sdn-controller/pkg/raven/openapi/test
	rm -rf go/pkg/sdn-controller/pkg/raven/openapi/git_push.sh
	rm -rf go/pkg/sdn-controller/pkg/raven/openapi/.travis.yml

.PHONY: delete-generated-files-k8s
delete-generated-files-k8s: ## Deleted output of Kubernetes code generators to ensure there are no circular dependencies.
	rm -f go/pkg/k8s/apis/private.cloud/v1alpha1/zz_generated.deepcopy.go
	rm -rf go/pkg/k8s/generated
	rm -rf go/pkg/k8s/config/crd/bases/*.yaml
	rm -rf go/pkg/productcatalog_operator/config/rbac/role.yaml
	rm -rf go/pkg/productcatalog_operator/config/crd/bases/*.yaml
	rm -rf go/pkg/instance_operator/bm/config/rbac/role.yaml
	rm -f deployment/charts/compute-crds/templates/*.yaml
	rm -f deployment/charts/firewall-crds/templates/*.yaml
	rm -f deployment/charts/productcatalog-crds/templates/*.yaml
	rm -f go/pkg/sdn-controller/api/v1alpha1/zz_generated.deepcopy.go
	rm -f deployment/charts/sdn-controller-crds/templates/*.yaml
	rm -f deployment/charts/ilb-crds/templates/*.yaml
	rm -f deployment/charts/loadbalancer-crds/templates/*.yaml
	rm -f deployment/charts/kubernetes-crds/templates/*.yaml

# Run all Kubernetes code generators. These must run in a precise order.
# This uses the following code generators:
#  - https://book.kubebuilder.io/reference/controller-gen.html
#  - https://github.com/rancher/wrangler
#  - https://github.com/kubernetes/code-generator
.PHONY: generate-k8s
generate-k8s: export PATH = $(shell $(BAZEL) info output_base)/external/go_sdk/bin:$(ORIGINAL_PATH)
generate-k8s: controller-gen codegen-tool delete-generated-files-k8s ## Run Kubernetes code generators.
	@echo Generating go/pkg/k8s/apis/private.cloud/v1alpha1/zz_generated.deepcopy.go
	cd go/pkg/k8s && $(CONTROLLER_GEN) object:headerFile="$(ROOT_DIR)/hack/boilerplate.go.txt" paths="./apis/..."

	@echo Generating go/pkg/k8s/generated/clientset and go/pkg/k8s/generated/controllers
	go run go/pkg/k8s/codegen/main.go

	@echo Generating go/pkg/k8s/generated/cloudintelclient
	cd go && $(CODEGEN_GENERATE_GROUPS) \
		client,lister,informer \
		github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient \
		github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis \
		private.cloud:v1alpha1 \
		--output-base $(TMPDIR) \
		--go-header-file $(ROOT_DIR)/hack/boilerplate.go.txt
	mv $(TMPDIR)/github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient \
		go/pkg/k8s/generated

	@echo Generating go/pkg/k8s/generated/metal3client
	cd go && $(CODEGEN_GENERATE_GROUPS) \
		all \
		github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client \
		github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis \
		metal3.io:v1alpha1 \
		--output-base $(TMPDIR) \
		--go-header-file $(ROOT_DIR)/hack/boilerplate.go.txt
	mv $(TMPDIR)/github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client \
		go/pkg/k8s/generated

	@echo Generating go/pkg/k8s/config/crd/bases/*.yaml
	cd go/pkg/k8s && $(CONTROLLER_GEN) crd paths="./apis/private.cloud/..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/productcatalog_operator/config/crd/bases/*.yaml
	cd go/pkg/productcatalog_operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/instance_operator/bm/config/rbac/role.yaml
	cd go/pkg/instance_operator && $(CONTROLLER_GEN) rbac:roleName=manager-role paths="./..." output:dir=bm/config/rbac

	@echo Generating go/pkg/productcatalog_operator/config/rbac/role.yaml
	cd go/pkg/productcatalog_operator && $(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..."

	@echo Generating go/pkg/ilb_operator/config/crd/bases/*.yaml
	cd go/pkg/ilb_operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/ilb_operator/config/rbac/role.yaml
	cd go/pkg/ilb_operator && $(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..."

	@echo Generating go/pkg/firewall_operator/config/crd/bases/*.yaml
	cd go/pkg/firewall_operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/loadbalancer_operator/config/crd/bases/*.yaml
	cd go/pkg/loadbalancer_operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/loadbalancer_operator/config/rbac/role.yaml
	cd go/pkg/loadbalancer_operator && $(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..."

	@echo Generating go/pkg/kubernetes_operator/config/crd/bases/*.yaml
	cd go/pkg/kubernetes_operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/kubernetes_operator/config/rbac/role.yaml
	cd go/pkg/kubernetes_operator && $(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..."

	@echo Generating go/pkg/network/operator/config/crd/bases/*.yaml
	cd go/pkg/network/operator && $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

	@echo Generating go/pkg/network/operator/config/rbac/role.yaml
	cd go/pkg/network/operator && $(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..."

	@echo Generating go/pkg/network/operator/apis/v1alpha1/zz_generated.deepcopy.go
	cd go/pkg/network/operator && $(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

	@echo Generating deployment/charts/compute-crds/templates/*.yaml
	cp go/pkg/k8s/config/crd/bases/*.yaml deployment/charts/compute-crds/templates

	@echo Generating deployment/charts/productcatalog-crds/templates/*.yaml
	cp go/pkg/productcatalog_operator/config/crd/bases/*.yaml deployment/charts/productcatalog-crds/templates

	@echo Generating deployment/charts/sdn-controller-crds/templates/*.yaml
	cd go/pkg/sdn-controller && $(CONTROLLER_GEN) rbac:roleName=manager-role crd paths="./..." output:crd:artifacts:config=../../../deployment/charts/sdn-controller-crds/templates

	@echo Generating go/pkg/sdn-controller/api/v1alpha1/zz_generated.deepcopy.go
	cd go/pkg/sdn-controller && $(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

	@echo Generating deployment/charts/ilb-crds/templates/*.yaml
	cp go/pkg/ilb_operator/config/crd/bases/*.yaml deployment/charts/ilb-crds/crds

	@echo Generating deployment/charts/loadbalancer-crds/templates/*.yaml
	cp go/pkg/loadbalancer_operator/config/crd/bases/*.yaml deployment/charts/loadbalancer-crds/templates/

	@echo Generating deployment/charts/network-crds/templates/*.yaml
	cp go/pkg/network/operator/config/crd/bases/*.yaml deployment/charts/network-crds/templates/

	@echo Generating deployment/charts/firewall-crds/templates/*.yaml
	cp go/pkg/firewall_operator/config/crd/bases/*.yaml deployment/charts/firewall-crds/templates

	@echo Generating deployment/charts/kubernetes-crds/templates/*.yaml
	cp go/pkg/kubernetes_operator/config/crd/bases/*.yaml deployment/charts/kubernetes-crds/crds

.PHONY: download-metal3-apis
download-metal3-apis: ## Download Metal3 APIs.
	wget -q https://github.com/metal3-io/baremetal-operator/archive/refs/tags/v$(METAL3_BAREMETAL_OPERATOR_VERSION).tar.gz -O - | \
	tar -xzv -C $(TMPDIR)
	cp -rv $(TMPDIR)/baremetal-operator-$(METAL3_BAREMETAL_OPERATOR_VERSION)/apis/metal3.io/* go/pkg/k8s/apis/metal3.io/
	$(MAKE) generate-copyright-headers

.PHONY: generate-docs
generate-docs: ## Run Protobuf documentation generator.
	rm -rf $(ROOT_DIR)/docs/generated
	mkdir -p $(ROOT_DIR)/docs/generated
	docker run --rm \
		-u $(shell id -u $(USER)):$(shell id -g $(USER)) \
		-v $(ROOT_DIR)/public_api/proto:/protos:ro \
		-v $(ROOT_DIR)/docs/generated:/out \
		${DOCKERIO_IMAGE_PREFIX}pseudomuto/protoc-gen-doc@sha256:779263a6dc01fbe375298c4e556d784640539e7b2bd433a50588285da88b74d5 \
		--doc_opt=html,index.html

.PHONY: docs
docs: bazel ## Build Sphinx documentation
	$(BAZEL) build $(BAZEL_OPTS) //docs

docs-http-server: docs ## Run an HTTP server to serve Sphinx documentation
	cd bazel-bin/docs/source/private_docs_html && \
	python3 -m http.server 8180 --bind 127.0.0.1

.PHONY: go-sdk
go-sdk: bazel ## Build Go SDK.
	cd go && $(BAZEL) build @go_sdk//:bin/go

.PHONY: go-sdk-export
go-sdk-export: ## Show command to export PATH to include the Go SDK. Run with "eval `make go-sdk-export`".
	@$(MAKE) go-sdk 1>&2
	@echo "export PATH=$(shell $(BAZEL) info output_base)/external/go_sdk/bin:$(ORIGINAL_PATH)"

.PHONY: go-get
go-get: bazel ## Update a Go module to the latest version. For example: GO_GET_OPTS=github.com/docker/docker make go-get gazelle
	cd go && $(BAZEL) run @go_sdk//:bin/go -- get $(GO_GET_OPTS) $(GO_GET_MODULE)

.PHONY: go-list
go-list: bazel ## List Go modules.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- list -m all

.PHONY: go-mod-graph
go-mod-graph: bazel ## Show Go module graph.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- mod graph

.PHONY: go-mod-why
go-mod-why: bazel ## Explain why a Go module is needed.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- mod why $(GO_MOD_WHY_OPTS)

.PHONY: gazelle
gazelle: bazel tidy ## Run Gazelle to update Bazel dependency list based on Go imports.
	$(BAZEL) run $(BAZEL_OPTS) //:gazelle
	$(BAZEL) run $(BAZEL_OPTS) //:gazelle-update-repos
	$(BAZEL) run $(BAZEL_OPTS) //:gazelle

.PHONY: tidy
tidy: bazel ## Update go.mod
	cd go && $(BAZEL) run @go_sdk//:bin/go -- mod tidy

# Store our own copy of Google API Protobuf files per https://github.com/grpc-ecosystem/grpc-gateway#usage.
.PHONY: download-google-apis
download-google-apis: ## Download the latest Google API Protobuf files.
	mkdir -p $(GOOGLE_API_DIR)
	wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -O $(GOOGLE_API_DIR)/annotations.proto
	wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/field_behavior.proto -O $(GOOGLE_API_DIR)/field_behavior.proto
	wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -O $(GOOGLE_API_DIR)/http.proto
	wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/httpbody.proto -O $(GOOGLE_API_DIR)/httpbody.proto

.PHONY: fmt
fmt: bazel ## Run go fmt against code.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- fmt ./...

.PHONY: test-fmt
test-fmt: bazel ## Run go fmt against code. Fail if any changes are made.
	@go_fmt_temp=`mktemp --dry-run`; \
	cd go && $(BAZEL) run @go_sdk//:bin/go -- fmt ./... | tee $$go_fmt_temp; \
	fmt_count=`cat $$go_fmt_temp | wc -l`; \
	rm -f $$go_fmt_temp; \
	if [ "$$fmt_count" != "0" ] ; then \
	echo "Format Failures found by 'go fmt': Please update the files above"; \
		/bin/false; \
	fi

.PHONY: helmfile-fmt
helmfile-fmt: $(YQ) ## Format Helmfile environment files.
	for file in $(HELMFILE_ENVIRONMENT_FILES); do \
	   	$(YQ) --inplace -P 'sort_keys(..)' $$file; \
	done

.PHONY: test-helmfile-fmt
test-helmfile-fmt: $(YQ) ## Run helmfile-fmt. Fail if any changes are made.
	for file in $(HELMFILE_ENVIRONMENT_FILES); do \
		$(YQ) -P 'sort_keys(..)' $$file > $(TMPDIR)/$$(basename $$file); \
		diff $$file $(TMPDIR)/$$(basename $$file) > /dev/null || (echo "File $$file is not formatted correctly. Run 'make helmfile-fmt'." && false); \
	done

.PHONY: helmfile-split
helmfile-split: $(YQ) ## Split Helmfile environment files by region. Intended for one-time migration.
	# See technique to prune tree in https://mikefarah.gitbook.io/yq/operators/path#set-path
	$(YQ) '(.regions.us-staging-1) as $$i ireduce({}; setpath($$i | path; $$i))' deployment/helmfile/environments/staging.yaml.gotmpl > deployment/helmfile/environments/staging-region-us-staging-1.yaml.gotmpl
	$(YQ) '(.regions.us-staging-2) as $$i ireduce({}; setpath($$i | path; $$i))' deployment/helmfile/environments/staging.yaml.gotmpl > deployment/helmfile/environments/staging-region-us-staging-2.yaml.gotmpl
	$(YQ) '(.regions.us-staging-3) as $$i ireduce({}; setpath($$i | path; $$i))' deployment/helmfile/environments/staging.yaml.gotmpl > deployment/helmfile/environments/staging-region-us-staging-3.yaml.gotmpl
	$(YQ) --inplace 'del(.regions)' deployment/helmfile/environments/staging.yaml.gotmpl
	$(YQ) '(.regions.us-region-1) as $$i ireduce({}; setpath($$i | path; $$i))' deployment/helmfile/environments/prod.yaml.gotmpl > deployment/helmfile/environments/prod-region-us-region-1.yaml.gotmpl
	$(YQ) '(.regions.us-region-2) as $$i ireduce({}; setpath($$i | path; $$i))' deployment/helmfile/environments/prod.yaml.gotmpl > deployment/helmfile/environments/prod-region-us-region-2.yaml.gotmpl
	$(YQ) --inplace 'del(.regions)' deployment/helmfile/environments/prod.yaml.gotmpl

.PHONY: test-generate
test-generate: ## Run generate. Report (but do not fail) if any changes are made.
	$(MAKE) generate
	git status

.PHONY: vet
vet: bazel ## Run go vet against code.
	cd go && $(BAZEL) run @go_sdk//:bin/go -- vet ./...

.PHONY: go-lint
go-lint: export PATH = $(shell $(BAZEL) info output_base)/external/go_sdk/bin:$(ORIGINAL_PATH)
go-lint: go-sdk golang-ci ## Run go linter.
	cd go && $(GOLANGCI_LINT) run ./... --timeout=15m || true

# DOCKERIO_IMAGE_PREFIX will be replaced by amr-registry.caas.intel.com/cache/ when run in Jenkins.
.PHONY: pull-images
pull-images: ## Pull dependent Docker images.
# No need to check every time with sha256 tag
ifneq ($(shell docker inspect --format='{{index .RepoDigests 0}}' postgres:latest 2>/dev/null || echo NONE), postgres$(POSTGRES_IMAGE_TAG))
	docker pull ${DOCKERIO_IMAGE_PREFIX}library/postgres$(POSTGRES_IMAGE_TAG)
	docker tag ${DOCKERIO_IMAGE_PREFIX}library/postgres$(POSTGRES_IMAGE_TAG) docker.io/library/postgres:latest
endif

.PHONY: test
test: test-fmt test-helmfile-fmt vet go-lint test-bazel test-copyright-headers ## Run all tests (except tests tagged "manual").

.PHONY: test-go
test-go: test-bazel ## Run Go tests.

.PHONY: test-bazel
test-bazel: pull-images bazel ## Run Bazel tests.
	$(BAZEL) test \
		$(BAZEL_OPTS) \
		$(BAZEL_TEST_TARGETS)

.PHONY: test-copyright-headers
test-copyright-headers: ## only ./go folder for now
	python3 ./hack/generate_copyright.py --dir ./go --verify-mode

.PHONY: test-go-non-ginkgo
test-go-non-ginkgo: pull-images bazel ## Run all Go non-Ginkgo tests.
	$(BAZEL) test \
		$(BAZEL_OPTS) \
		$(GO_NON_GINKGO_TEST_TARGETS)

.PHONY: test-go-ginkgo
test-go-ginkgo: pull-images bazel ## Run all Ginkgo tests (except tests labeled "large").
	$(BAZEL) test \
		$(BAZEL_OPTS) \
		$(GINKGO_TEST_TARGETS)

.PHONY: test-force
test-force: ## Run all tests, even if test result is cached (except tests tagged "manual").
	BAZEL_EXTRA_OPTS=--cache_test_results=no $(MAKE) test

.PHONY: test-streamed
test-streamed: test-fmt vet pull-images bazel ## Run all tests with output streamed to console (except tests tagged "manual").
	BAZEL_EXTRA_OPTS=--test_output=streamed $(MAKE) test

.PHONY: test-report-ginkgo
test-report-ginkgo: test-fmt vet pull-images bazel ## Run Ginkgo tests and generate reports.
	$(BAZEL) test \
		--cache_test_results=no \
		--test_arg=-ginkgo.json-report=/tmp/ginkgo-report.json \
		--test_arg=-ginkgo.junit-report=/tmp/junit-report.json \
		$(BAZEL_OPTS) \
		$(GINKGO_TEST_TARGETS)

# Use a fixed DOCKER_TAG to avoid Bazel test cache invalidation when the Git hash changes.
# Use a fixed HELM_CHART_VERSION (prefix) to avoid Bazel test cache invalidation if it is changed.
.PHONY: test-e2e-compute-vm
test-e2e-compute-vm: ## Run compute VMaaS E2E tests with a new local kind cluster.
	DOCKER_TAG=latest \
	HELM_CHART_VERSION=0.0.1 \
	IDC_ENV=test-e2e-compute-vm \
	LOCAL_REGISTRY_NAME=$(LOCAL_REGISTRY_NAME) \
	LOCAL_REGISTRY_PORT=$(KIND_REGISTRY_PORT) \
	SSH_PROXY_USER=$(SSH_PROXY_USER) \
	BAZEL_TEST_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 \
	//go/test/compute/e2e/vm:vm_test" \
	$(MAKE) secrets install-vault-requirements update-etc-hosts deploy-registry container-and-chart-push test-custom-quick

.PHONY: test-e2e-compute-vm-quick
test-e2e-compute-vm-quick: ## Run compute VMaaS E2E tests using a cluster deployed with test-e2e-compute-vm.
	KUBECONFIG=local/secrets/test-e2e-compute-vm/kubeconfig/config kind export kubeconfig --name test-e2e-compute-vm-global
	docker inspect test-e2e-compute-vm-global-control-plane | jq --raw-output '.[0].NetworkSettings.Ports."80/tcp"[0].HostPort' | tee local/test-e2e-compute-vm-global_host_port_80
	docker inspect test-e2e-compute-vm-global-control-plane | jq --raw-output '.[0].NetworkSettings.Ports."443/tcp"[0].HostPort' | tee local/test-e2e-compute-vm-global_host_port_443
	IDC_ENV=test-e2e-compute-vm \
	BAZEL_TEST_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 \
	//go/test/compute/e2e/vm:vm_quick_test" \
	$(MAKE) test-custom-quick

.PHONY: test-e2e-compute-bm-pre-commands
test-e2e-compute-bm-pre-commands:
	go/test/compute/e2e/bm/pre_commands.sh

.PHONY: test-e2e-compute-bm
test-e2e-compute-bm: ## Run compute BMaaS E2E tests with a new local kind cluster.
	DOCKER_TAG=latest \
	HELM_CHART_VERSION=0.0.1 \
	IDC_ENV=test-e2e-compute-bm \
	LOCAL_REGISTRY_NAME=$(LOCAL_REGISTRY_NAME) \
	LOCAL_REGISTRY_PORT=$(KIND_REGISTRY_PORT) \
	SECRETS_DIR=$(ROOT_DIR)/local/secrets/test-e2e-compute-bm \
	SSH_PROXY_USER=$(SSH_PROXY_USER) \
	BAZEL_TEST_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 \
	//go/test/compute/e2e/bm:bm_test" \
	$(MAKE) secrets test-e2e-compute-bm-pre-commands install-vault-requirements update-etc-hosts deploy-registry container-and-chart-push test-custom-quick

.PHONY: test-bazel-large-jenkins
test-bazel-large-jenkins: ## Run Bazel tests marked as size=large|enormous and tags=jenkins. Used by Jenkins.
	DOCKER_TAG=latest \
	HELM_CHART_VERSION=0.0.1 \
	SSH_PROXY_USER=$(SSH_PROXY_USER) \
	BAZEL_TEST_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 \
	$(BAZEL_LARGE_JENKINS_TEST_TARGETS) $(BAZEL_ENORMOUS_JENKINS_TEST_TARGETS)" \
	$(MAKE) install-vault-requirements update-etc-hosts test-custom-quick

.PHONY: test-bazel-large-bm-jenkins
test-bazel-large-bm-jenkins: ## Run Bazel tests marked as size=enormous and tags=jenkins for BM. Used by Jenkins.
	DOCKER_TAG=latest \
	HELM_CHART_VERSION=0.0.1 \
	SECRETS_DIR=$(ROOT_DIR)/local/secrets/test-e2e-compute-bm \
	BAZEL_TEST_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 \
	$(BAZEL_ENORMOUS_JENKINS_BM_TEST_TARGETS)" \
	$(MAKE) secrets test-e2e-compute-bm-pre-commands install-vault-requirements update-etc-hosts test-custom-quick

.PHONY: test-custom
test-custom: pull-images bazel ## Run bazel tests with parameters from BAZEL_EXTRA_OPTS and BAZEL_TEST_OPTS. See README.md.
	$(BAZEL) test $(BAZEL_OPTS) $(BAZEL_TEST_OPTS)

.PHONY: test-custom-quick
test-custom-quick: bazel ## Run bazel tests with parameters from BAZEL_EXTRA_OPTS and BAZEL_TEST_OPTS.
	$(BAZEL) test $(BAZEL_OPTS) $(BAZEL_TEST_OPTS)

.PHONY: test-universe-deployer
test-universe-deployer: bazel ## Run Universe Deployer unit tests.
	$(BAZEL) test $(BAZEL_OPTS) \
		//deployment/universe_deployer:universe_config_tests \
		//go/pkg/universe_deployer/manifests_generator:manifests_generator_test

.PHONY: clean
clean: ## Delete build outputs and local Bazel cache. Use for troubleshooting or freeing disk space. See also "make clean-bazel-remote-cache".
	-$(BAZEL) shutdown
	rm -rf $(LOCALBIN)
	$(MAKE) $(BAZEL)
	$(BAZEL) clean --expunge
	sudo rm -rf ~/.cache/bazel || true
	sudo rm -rf $(ROOT_DIR)/local/go || true
	@echo See also \"make clean-all\".

.PHONY: clean-bazel-remote-cache
clean-bazel-remote-cache: ## Delete Bazel "remote" cache volume.
	$(MAKE) stop-bazel-remote-cache
	-docker volume rm "${LOCAL_BAZEL_REMOTE_CACHE_NAME}"

.PHONY: clean-docker
clean-docker: ## Delete stopped containers, unused anonymous volumes, and unused images.
	docker system prune --all --volumes --force

.PHONY: clean-all
clean-all: clean clean-bazel-remote-cache clean-docker ## WARNING! Includes all "clean*" targets.

# instrumentation_filter contains regexes for Bazel targets whose lines will be instrumented for coverage reporting.
# It includes all code under /go except for generated or irrelevant code.
# The list of targets that follow should be the tests that should be executed for coverage reporting.
# The resulting coverage report can be served by running "cd local && python3 -m http.server".
.PHONY: coverage-compute
coverage-compute: pull-images bazel ## Generate coverage report for compute services.
	rm -rf local/coverage local/coverage-compute.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/baremetal_enrollment/mocks[/:],\
	-^//go/pkg/cloudaccount[/:],\
	-^//go/pkg/compute_api_server/openapi[/:],\
	-^//go/pkg/forked[/:],\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	-^//go/pkg/raven/openapi[/:],\
	^//go[/:],\
	" \
	//go/pkg/baremetal_enrollment/... \
	//go/pkg/compute_api_server/... \
	//go/pkg/compute_integration_test/... \
	//go/pkg/compute_metering_monitor/... \
	//go/pkg/conf/... \
	//go/pkg/git_to_grpc_synchronizer/... \
	//go/pkg/instance_operator/... \
	//go/pkg/instance_replicator/... \
	//go/pkg/instance_scheduler/... \
	//go/pkg/log/... \
	//go/pkg/manageddb/... \
	//go/pkg/protodb/... \
	//go/pkg/secrets/... \
	//go/pkg/ssh_proxy_operator/... \
	//go/pkg/tools/...

	genhtml --output local/coverage-compute --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-compute.tgz coverage-compute

.PHONY: coverage-productcatalog
coverage-productcatalog: pull-images bazel ## Generate coverage report for compute services.
	rm -rf local/coverage local/coverage-productcatalog.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	^//go[/:],\
	" \
	//go/pkg/productcatalog_operator/...

	genhtml --output local/coverage-productcatalog --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-productcatalog.tgz coverage-productcatalog

.PHONY: coverage-iks
coverage-iks: pull-images bazel ## Generate coverage report for compute services.
	rm -rf local/coverage local/coverage-iks.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	^//go[/:],\
	" \
	//go/pkg/iks/...

	genhtml --output local/coverage-iks --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-iks.tgz coverage-iks

.PHONY: coverage-billing
coverage-billing: pull-images bazel ## Generate coverage report for billing services.
	rm -rf local/coverage local/coverage-billing.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/cloudaccount[/:],\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	^//go[/:],\
	" \
	//go/pkg/billing/... \
	//go/pkg/billing_common/... \
	//go/pkg/billing_driver_intel/... \
	//go/pkg/billing_driver_aria/... \
	//go/pkg/billing_driver_standard/...

	genhtml --output local/coverage-billing --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-billing.tgz coverage-billing

.PHONY: coverage-storage
coverage-storage: pull-images bazel ## Generate coverage report for filesystem services.
	rm -rf local/coverage local/coverage-storage.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/observability[/:],\
	-^//go/pkg/cloudaccount[/:],\
	-^//go/pkg/authutil[/:],\
	-^//go/pkg/authz[/:],\
	-^//go/pkg/grpcutil[/:],\
	-^//go/pkg/conf[/:],\
	-^//go/pkg/log[/:],\
	-^//go/pkg/tools[/:],\
	-^//go/pkg/tlsutil[/:],\
	-^//go/pkg/manageddb[/:],\
	-^//go/pkg/mineral-river[/:],\
	-^//go/pkg/protodb[/:],\
	-^//go/pkg/secrets[/:],\
	-^//go/pkg/storage/secrets[/:],\
	-^//go/pkg/storage/storagecontroller/test/mocks[/:],\
	-^//go/pkg/notification_gateway[/:],\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/storage/secrets/vault_mocks.go[/:],\
	-^//go/pkg/storage/storage_custom_metrics_service[/:],\
	-^//go/pkg/pb[/:],\
	-^//go/pkg/utils[/:],\
	-^//go/pkg/sesutil[/:],\
	-^//go/pkg/snsutil[/:],\
	-^//go/pkg/sqsutil[/:],\
	-^//go/pkg/billing[/:],\
	-^//go/pkg/billing_common[/:],\
	-^//go/pkg/billing_driver_aria[/:],\
	-^//go/pkg/billing_driver_intel[/:],\
	-^//go/pkg/billing_driver_standard[/:],\
	-^//go/pkg/metering[/:],\
	-^//go/pkg/usage[/:],\
	^//go[/:],\
	" \
	//go/pkg/storage/... \

	genhtml --output local/coverage-storage --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-storage.tgz coverage-storage

.PHONY: coverage-all
coverage-all: pull-images bazel ## Generate coverage report for all code inside of go directory except for generated files
	$(BAZEL) coverage $(BAZEL_OPTS) \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/compute_api_server/openapi[/:],\
	-^//go/pkg/k8s/generated[/:],\
	-^//go/pkg/mmws/openapi[/:],\
	-^//go/pkg/pb[/:],\
	-^//go/pkg/raven/openapi[/:],\
	-^//go/test[/:],\
	^//go[/:],\
	" \
	//go/...

.PHONY: coverage-fleet-admin
coverage-fleet-admin: pull-images bazel ## Generate coverage report for Fleet Admin Service.
	rm -rf local/coverage local/coverage-fleet-admin.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	^//go[/:],\
	" \
	//go/pkg/fleet_admin/...

	genhtml --output local/coverage-fleet-admin --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-fleet-admin.tgz coverage-fleet-admin

.PHONY: coverage-all-genhtml
coverage-all-genhtml: ## generate html report for all golang code from previously generated bazel coverage report
	genhtml --output local/coverage-all --show-details --legend "./bazel-out/_coverage/_coverage_report.dat"
	cd local && python3 -c "import shutil; shutil.make_archive('coverage-all', 'zip', 'coverage-all')"


.PHONY: coverage-quota-management-service
coverage-quota-management-service: pull-images bazel ## Generate coverage report for Quota Management Service
	rm -rf local/coverage local/coverage-quota-management-service.tgz

	$(BAZEL) coverage $(BAZEL_OPTS) \
	--cache_test_results=no \
	--combined_report=lcov \
	--instrumentation_filter="\
	-^//go/pkg/k8s[/:],\
	-^//go/pkg/pb[/:],\
	-^//go/pkg/observability[/:],\
	-^//go/pkg/grpcutil[/:],\
	-^//go/pkg/manageddb[/:],\
	-^//go/pkg/authutil[/:],\
	-^//go/pkg/conf[/:],\
	-^//go/pkg/log[/:],\
	-^//go/pkg/mineral-river[/:],\
	^//go[/:],\
	" \
	//go/pkg/quota_management/...

	genhtml --output local/coverage-quota-management-service --show-details --legend "$(shell bazel info output_path)/_coverage/_coverage_report.dat"
	cd local && tar -czf coverage-quota-management-service.tgz coverage-quota-management-service

##@ Build

.PHONY: build
build: bazel ## Build everything.
	$(BAZEL) build $(BAZEL_OPTS) \
		//go/... \
		//deployment:all_chart_versions_except_idc_versions \
		//docs

.PHONY: build-go
build-go: bazel ## Build Go.
	$(BAZEL) build $(BAZEL_OPTS) //go/...

.PHONY: container-build
container-build:  ## Build Docker images.
	$(BAZEL) build $(BAZEL_OPTS) //deployment/push:all_container_push

# container-build-%
$(BAZEL_CONTAINERS_WITHOUT_HELM:%=container-build-%): container-build-%: bazel
	$(BAZEL) build $(BAZEL_OPTS) //deployment/push:$*_container_push

# Help for generated targets
ifeq (1,0)
container-build-opa: ## Build the specified Docker image.
endif

.PHONY: container-and-chart-push
container-and-chart-push: bazel ## Combined container-push and chart-push.
	@echo Pushing container images to $(DOCKER_REGISTRY) with prefix \"$(DOCKER_IMAGE_PREFIX)\" and tag $(DOCKER_TAG)
	@echo Pushing Helm charts to $(HELM_REGISTRY) with project $(HELM_PROJECT) and version $(IDC_FULL_VERSION)
	$(BAZEL) $(BAZEL_STARTUP_OPTS) run $(BAZEL_OPTS) //deployment/push:all_container_and_chart_push

.PHONY: container-push
container-push: bazel ## Build container images and push to container registry. If pushing to a private registry, this assumes that "docker login" has already been run.
	@echo Pushing container images to $(DOCKER_REGISTRY) with prefix \"$(DOCKER_IMAGE_PREFIX)\" and tag $(DOCKER_TAG)
	$(BAZEL) run $(BAZEL_OPTS) //deployment/push:all_container_push

##@ deploy-all-in-kind (v1) (deprecated)

# container-push-%
# If pushing to a private registry, this assumes that "docker login" has already been run.
$(BAZEL_CONTAINERS:%=container-push-%): container-push-%: bazel
	$(BAZEL) run $(BAZEL_OPTS) //deployment/push:$*_container_push

# container-push-%
# No-op container-push for charts without custom containers.
$(NON_BAZEL_CUSTOM_CHARTS:%=container-push-%): container-push-%:
$(THIRD_PARTY_CHARTS:%=container-push-%): container-push-%:

# helm-build-%
$(HELM_CHARTS:%=helm-build-%): helm-build-%: bazel
	$(BAZEL) build $(BAZEL_OPTS) //deployment/charts/$*:chart

.PHONY: helm-login
helm-login: helm ## Login to Helm registry. Uses $HELM_USERNAME and $HELM_PASSWORD to authenticate.
	echo "$(HELM_PASSWORD)" | $(HELM) registry login -u $(HELM_USERNAME) --password-stdin $(HELM_REGISTRY)

# If pushing to a private registry, this assumes that "helm registry login" has already been run.
.PHONY: helm-push
helm-push: bazel ## Build Helm charts and push to Helm chart registry.
	@echo Pushing Helm charts to $(HELM_REGISTRY) with project $(HELM_PROJECT) and version $(IDC_FULL_VERSION)
	$(BAZEL) run $(BAZEL_OPTS) //deployment/push:all_chart_push

# helm-push-%
$(HELM_CHARTS:%=helm-push-%): helm-push-%: bazel
	$(BAZEL) run $(BAZEL_OPTS) //deployment/push:$*_chart_push

# helm-template-%
$(HELM_CHARTS:%=helm-template-%): helm-template-%: helm-push-% ## Render a Helm chart. Override name with HELM_RELEASE_NAME.
	$(BAZEL) build $(BAZEL_OPTS) //deployment:$*_chart_version
	$(HELM) template --debug $${HELM_RELEASE_NAME:-us-region-1a-$*} oci://$(HELM_REGISTRY)/$(HELM_PROJECT)/$* \
		--version $(shell cat bazel-bin/deployment/chart_versions/$*.version)

# %-chart-version updates only a single chart and copies it (& other most-recently built charts) into the helm chart versions directory.
$(HELMFILE_CHARTS:%=%-chart-version): %-chart-version: bazel
	rm -rf local/idc-versions/$(IDC_FULL_VERSION)
	$(BAZEL) build $(BAZEL_OPTS) //deployment:$*_chart_version
	mkdir -p $(HELM_CHART_VERSIONS_DIR)
	cp bazel-bin/deployment/chart_versions/* $(HELM_CHART_VERSIONS_DIR)

# helm-push-%
# No-op helm-push for charts that do not need to be pushed.
$(THIRD_PARTY_CHARTS:%=helm-push-%): helm-push-%:

.PHONY: deploy-compute
deploy-compute: $(DEPLOY_DEPS) $(DEPLOY_COMPUTE_DEPS) ## Deploy compute components in an existing kind cluster

.PHONY: deploy-all-in-kind
deploy-all-in-kind: $(DEPLOY_ALL_IN_KIND_DEPS) ## Build all components, deploy a new kind cluster, and deploy all components (deprecated).
	@echo deploy-all-in-kind completed.

.PHONY: deploy-all-in-kind-v2
deploy-all-in-kind-v2: bazel $(DEPLOY_ALL_IN_KIND_V2_DEPS) ## Build all components, deploy a new kind cluster, and deploy all components.
	$(BAZEL) run $(BAZEL_OPTS) $(DEPLOY_ALL_IN_KIND_OPTS)
	@echo deploy-all-in-kind-v2 completed.

.PHONY: upgrade-all-in-kind-v2
upgrade-all-in-kind-v2: bazel ## Build and upgrade all components in an existing kind cluster. Use after "make deploy-all-in-kind-v2". Delete applications instead of upgrading by setting DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE to a regex matching the names of Argo Application custom resources.
	$(BAZEL) run $(BAZEL_OPTS) $(DEPLOY_ALL_IN_KIND_OPTS) \
		--applications-to-delete "$(DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE)" \
		--upgrade=true
	@echo upgrade-all-in-kind-v2 completed.

.PHONY: undeploy-all-in-kind
undeploy-all-in-kind: ## Destroy all kind clusters
	kind get clusters | grep idc | xargs -i -P 0 kind delete cluster --name {}

.PHONY: deploy-metal-in-kind
deploy-metal-in-kind: $(DEPLOY_METAL_IN_KIND_DEPS) ## Build all components for bmaas in a new kind cluster and a virtual BM stack

.PHONY: deploy-metal-in-rke2
deploy-metal-in-rke2: $(DEPLOY_METAL_IN_RKE2_DEPS) ## Build all components for bmaas in a new rke2 cluster

.PHONY: deploy-all-in-k8s
deploy-all-in-k8s: export PATH := $(LOCALBIN):$(PATH)
deploy-all-in-k8s: bazel install-vault-requirements ## Build and deploy all components to an existing Kubernetes cluster.
	UNIVERSE_DEPLOYER_JOBS_PER_PIPELINE=$(UNIVERSE_DEPLOYER_JOBS_PER_PIPELINE) \
	UNIVERSE_DEPLOYER_POOL_DIR=$(UNIVERSE_DEPLOYER_POOL_DIR) \
	$(BAZEL) run $(BAZEL_OPTS) \
	//go/pkg/universe_deployer/cmd/deploy_all_in_k8s:deploy_all_in_k8s -- \
	--bazel-binary $(BAZEL) \
	--build-artifacts-dir local/build-artifacts \
	--cache-dir local/deploy-all-in-k8s/cache \
	--commit $(shell git rev-parse HEAD) \
	--delete-all-argo-applications="$${DELETE_ALL_ARGO_APPLICATIONS:-false}" \
	--delete-argo-cd=$${DELETE_ARGO_CD:-false} \
	--delete-gitea=$${DELETE_GITEA:-false} \
	--delete-vault=$${DELETE_VAULT:-false} \
	--git-pusher-dry-run=$${GIT_PUSHER_DRY_RUN:-false} \
	--idc-env "$(IDC_ENV)" \
	--include-deploy=$${INCLUDE_DEPLOY:-true} \
	--include-push=$${PUSH_DEPLOYMENT_ARTIFACTS:-true} \
	--include-vault-configure=$${INCLUDE_VAULT_CONFIGURE:-false} \
	--include-vault-load-secrets=$${INCLUDE_VAULT_LOAD_SECRETS:-false} \
	--secrets-dir $(SECRETS_DIR) \
	--semantic-version $(IDC_SEMANTIC_VERSION) \
	--universe-config "$(UNIVERSE_CONFIG)"
	@echo deploy-all-in-k8s completed.

.PHONY: deploy-idc
deploy-idc: $(DEPLOY_IDC_DEPS) ## Deploy IDC components in an existing kind cluster
	@echo deploy-idc completed.

.PHONY: deploy-all-in-kind-go
deploy-all-in-kind-go: bazel ## Run deploy_all_in_kind (Go version)
	$(BAZEL) run $(BAZEL_OPTS) $(DEPLOY_ALL_IN_KIND_OPTS)

.PHONY: show-cluster-info
show-cluster-info: $(KUBECTL) ## Deletes all IDC releases using helmfile
	$(KUBECTL) cluster-info
	$(KUBECTL) version

.PHONY: delete-all-instances
delete-all-instances: secrets $(HELM) $(KUBECTL) $(JQ) ## Deletes all IDC releases using helmfile
	HELM=$(HELM) \
	KUBECTL=$(KUBECTL) \
	JQ=$(JQ) \
	deployment/jenkinsfile/deployer-clean-instances.sh

.PHONY: delete-all-pvcs
delete-all-pvcs: secrets $(HELM) $(KUBECTL) ## Deletes all IDC releases using helmfile
	HELM=$(HELM) \
	KUBECTL=$(KUBECTL) \
	deployment/jenkinsfile/undeploy/undeploy.sh

.PHONY: get-all-helm-releases
get-all-helm-releases: secrets $(HELM) $(KUBECTL) ## Gets all installed Helm releases
	$(HELM) ls -a -A
	$(KUBECTL) get all -A
	$(KUBECTL) get crds

.PHONY: deploy-vendors
deploy-vendors: $(KUBECTL) ## Applies all the vendors from the dev/vendors/ folder
	$(KUBECTL) apply -f $(PRODUCTS_PATH)/dev/vendors/ -n idcs-system

.PHONY: deploy-products ## Applies all the products from the dev/products/ folder
deploy-products: $(KUBECTL)
	$(KUBECTL) apply -f $(PRODUCTS_PATH)/dev/products/ -n idcs-system

.PHONY: get-products ## Gets all vendors and products from the cluster
get-products: $(KUBECTL)
	$(KUBECTL) get products,vendors -A

.PHONY: undeploy-jobs
undeploy-jobs: # Delete Kubernetes jobs
	-cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=git-to-grpc-synchronizer --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: deploy-all-helm-releases
deploy-all-helm-releases: undeploy-jobs ## Apply all helmfile releases. You must run `make helm-push` before using this.
	cd deployment/helmfile && $(HELMFILE) apply --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: deploy-foundation
deploy-foundation: $(DEPLOY_FOUNDATION_DEPS) ## Deploy foundational components in an existing kind cluster

.PHONY: deploy-finance
deploy-finance: $(DEPLOY_FINANCE_DEPS) ## Deploy financial components in an existing kind cluster

.PHONY: deploy-iks
deploy-iks: $(DEPLOY_IKS_DEPS)

.PHONY: deploy-insights
deploy-insights: $(DEPLOY_INSIGHTS_DEPS)

.PHONY: deploy-kfaas
deploy-kfaas: $(DEPLOY_KFAAS_DEPS)

.PHONY: deploy-cloudmonitor
deploy-cloudmonitor: $(DEPLOY_CLOUDMONITOR_DEPS)

.PHONY: deploy-cloudmonitor-logs
deploy-cloudmonitor-logs: $(DEPLOY_CLOUDMONITOR_LOGS_DEPS)

.PHONY: deploy-dpai
deploy-dpai: $(DEPLOY_DPAI_DEPS)

.PHONY: dpai-sqlc-generate
dpai-sqlc-generate:
	@echo "dpai-sqlc-generate is needed only for the development purpose. Make sure dpai-db is up and running."
	./go/svc/dpai/util-scripts/sqlc.sh

.PHONY: deploy-dataloader
deploy-kfaas: $(DEPLOY_DATALOADER_DEPS)

.PHONY: deploy-compute
deploy-compute: $(DEPLOY_COMPUTE_DEPS) ## Deploy compute components in an existing kind cluster

.PHONY: deploy-storage
deploy-storage: $(DEPLOY_STORAGE_DEPS) ## Deploy storage components in an existing kind cluster

.PHONY: deploy-training
deploy-training: $(DEPLOY_TRAINING_DEPS) ## Deploy training components in an existing kind cluster

.PHONY: deploy-iks-operators
deploy-iks-operators: $(DEPLOY_IKS_OPERATORS_DEPS)

.PHONY: deploy-sdn-vn-all
deploy-sdn-vn-all: helmfile helm-push-sdn-vn-controller secrets undeploy-sdn-vn-all container-push-sdn-vn-controller helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --allow-no-matching-release --selector component=sdnVN --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: undeploy-sdn-vn-all
undeploy-sdn-vn-all: helmfile helm-push-sdn-vn-controller secrets helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector component=sdnVN --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)


.PHONY: deploy-quota-management-service
deploy-quota-management-service: $(DEPLOY_QUOTA_MANAGEMENT_DEPS) ## Deploy quota management components in an existing kind cluster

# undeploy-%
$(HELMFILE_CHARTS:%=undeploy-%): undeploy-%: helmfile helm-push-% secrets helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

# deploy-%
$(HELMFILE_CHARTS:%=deploy-%): deploy-%: helmfile helm-push-% secrets undeploy-% container-push-% helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --allow-no-matching-release --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

# apply-%, like deploy-% but without undeploying first (eg. to avoid deleting CRs when deploying new version of -crd chart)
$(HELMFILE_CHARTS:%=apply-%): apply-%: helmfile helm-push-% container-push-% helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

# undeploy-only-%, like undeploy-% but without building ALL charts & chart-versions. For faster redeploy if just changing a single chart.
$(HELMFILE_CHARTS:%=undeploy-only-%): undeploy-only-%: helm-push-% secrets %-chart-version
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS) --skip-deps

# deploy-only-%, like deploy-% but without building ALL charts & chart-versions. For faster redeploy if just changing a single chart.
$(HELMFILE_CHARTS:%=deploy-only-%): deploy-only-%: helm-push-% secrets %-chart-version undeploy-only-% container-push-%
	cd deployment/helmfile && $(HELMFILE) apply --allow-no-matching-release --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS) --skip-deps

# apply-only-%, like deploy-only-% but without undeploying first (eg. to avoid deleting CRs when deploying new version of -crd chart)
$(HELMFILE_CHARTS:%=apply-only-%): apply-only-%: helm-push-% container-push-% %-chart-version
	cd deployment/helmfile && $(HELMFILE) apply --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS) --skip-deps

# diff-only-%, like apply-only-%, but without applying changes
$(HELMFILE_CHARTS:%=diff-only-%): diff-only-%: helm-push-% container-push-% %-chart-version
	cd deployment/helmfile && $(HELMFILE) diff --selector chart=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS) --skip-deps

# Help for generated targets
ifeq (1,0)
deploy-opal: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-grpc-rest-gateway: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-grpc-reflect: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-oidc: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-billing: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-billing-schedulers: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-billing-standard: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-billing-aria: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-billing-intel: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-notification-gateway: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudaccount: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudaccount-enroll: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-productcatalog: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-productcatalog-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-metering: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudcredits-worker: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudcredits: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-dataloader: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-usage: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-iks: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster
deploy-baremetal-enrollment-api: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-baremetal-enrollment-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-baremetal-enrollment-task: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-bm-instance-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-bm-validation-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-compute-api-server: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-compute-metering-monitor: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-instance-replicator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-loadbalancer-replicator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-loadbalancer-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-ssh-proxy-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-vm-instance-scheduler: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-vm-instance-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-fleet-node-reporter: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-console: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-compute-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-loadbalancer-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-firewall-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-firewall-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-productcatalog-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-metal3-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-metallb-custom-resources: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-bm-dnsmasq: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-dhcp-proxy: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-init-k8s-resources: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-idcs-istio-custom-resources: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cert-manager: ## Deploy the specified Helm chart to an existing cluster.
deploy-external-secrets: ## Deploy the specified Helm chart to an existing cluster.
deploy-metallb: ## Deploy the specified Helm chart to an existing cluster.
deploy-git-to-grpc-synchronizer: ## Populate Compute databases using git-to-grpc-synchronizer.
deploy-trade-scanner: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-training-api-server: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-armada: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-iks-operators: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-kfaas: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudmonitor: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-cloudmonitor-logs-api-server: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-dpai: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-api-server: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-vast-storage-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-bucket-replicator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-replicator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-scheduler: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-user: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-admin-api-server: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-kms: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-metering-monitor: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-vast-metering-monitor: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-custom-metrics-service: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-storage-resource-cleaner: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-quota-management-service: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
apply-sdn-controller-crds: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-sdn-controller: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-sdn-controller-rest: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-sdn-integrity-checker: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-switch-config-saver: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-intel-device-plugin: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-object-store-operator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-bucket-metering-monitor: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-pgoperator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-miniooperator: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-user-credentials: ## Build container, build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
deploy-security-insights: ## Deploy IKS security insights APIs
deploy-kubescore: ## Deploy IKS kubescore scheduler
deploy-security-scanner: ## Deploy IKS security scanner scheduler
deploy-vm-machine-image-resources: ## Build Helm chart, push, and deploy the specified Helm chart to an existing cluster.
endif

deploy-grpc-proxy: container-push-opa ## Deploy grpc-proxy (uses a custom OPA container)

##@ Deployment

.PHONY: run-helmfile
run-helmfile: secrets download-helm-chart-versions ## Run a custom helmfile command. Set arguments with HELMFILE_OPTS.
	cd deployment/helmfile && \
	IDC_USERNAME=$(HELM_USERNAME) IDC_PASSWORD=$(HELM_PASSWORD) \
	$(HELMFILE) --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: run-helmfile-only
run-helmfile-only: helmfile download-helm-chart-versions ## Run a custom helmfile command. Set arguments with HELMFILE_OPTS.
	cd deployment/helmfile && \
	IDC_USERNAME=$(HELM_USERNAME) IDC_PASSWORD=$(HELM_PASSWORD) \
	$(HELMFILE) --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: run-helmfile-nodeps
run-helmfile-nodeps: ## Run a custom helmfile command (no dependencies). Set arguments with HELMFILE_OPTS.
	cd deployment/helmfile && \
	IDC_USERNAME=$(HELM_USERNAME) IDC_PASSWORD=$(HELM_PASSWORD) \
	$(HELMFILE) --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: test-helmfile
test-helmfile: secrets $(HELMFILE)  ## Test that helmfile can render Kubernetes resources.
	cd deployment/helmfile && \
	$(HELMFILE) build --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: helmfile-dump
helmfile-dump: $(HELMFILE) $(HELM)	## Dump environment-specific values used by helmfile to a yaml file.
	cd deployment/helmfile && \
		$(HELMFILE) \
		write-values \
		--file helmfile-dump.yaml \
		--environment $(HELMFILE_ENVIRONMENT) \
		--output-file-template $(HELMFILE_DUMP_YAML) \
		$(HELMFILE_OPTS)

.PHONY: helmfile-generate-argocd-values
helmfile-generate-argocd-values: download-helm-chart-versions helmfile-generate-argocd-values-delete helmfile-generate-argocd-global-values helmfile-generate-argocd-regional-values helmfile-generate-argocd-az-values ## Write Helm release values for ArgoCD. Deprecated. Use run-manifests-generator instead.
	tar -czvf $(SECRETS_DIR)/$(IDC_ENV)-helm-values-${GIT_SHORT_HASH}.tgz -C $(HELMFILE_ARGOCD_VALUES_DIR) .

.PHONY: helmfile-generate-argocd-values-delete
helmfile-generate-argocd-values-delete:
	rm -rf "$(HELMFILE_ARGOCD_VALUES_DIR)"

.PHONY: helmfile-generate-argocd-global-values
helmfile-generate-argocd-global-values: secrets
	cd deployment/helmfile && \
		$(HELMFILE) \
		write-values \
		--file helmfile.yaml \
		--selector geographicScope=global \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-deps \
		--output-file-template "$(HELMFILE_ARGOCD_VALUES_DIR)/idc-global-services/{{ .Release.Labels.environmentName }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml" \
		$(HELMFILE_OPTS)
	cd deployment/helmfile && \
		$(HELMFILE) \
		list \
		--file helmfile.yaml \
		--selector geographicScope=global \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-charts \
		--output json > $(TMPDIR)/releaselist-global.json
	ENVIRONMENT_NAME="${HELMFILE_ENVIRONMENT}" ENVIRONMENT_TYPE="global" RELEASE_LIST_JSON="$(TMPDIR)/releaselist-global.json" OUTPUT_DIR=$(HELMFILE_ARGOCD_VALUES_DIR) \
		deployment/helmfile/scripts/generate_config_jsons.sh

.PHONY: helmfile-generate-argocd-regional-values
helmfile-generate-argocd-regional-values: secrets
	cd deployment/helmfile && \
		$(HELMFILE) \
		write-values \
		--file helmfile.yaml \
		--selector geographicScope=regional \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-deps \
		--output-file-template "$(HELMFILE_ARGOCD_VALUES_DIR)/idc-regional/{{ .Release.Labels.region }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml" \
		$(HELMFILE_OPTS)
	cd deployment/helmfile && \
		$(HELMFILE) \
		list \
		--file helmfile.yaml \
		--selector geographicScope=regional \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-charts \
		--output json > $(TMPDIR)/releaselist-regional.json
	ENVIRONMENT_NAME="${HELMFILE_ENVIRONMENT}" ENVIRONMENT_TYPE="regional" RELEASE_LIST_JSON="$(TMPDIR)/releaselist-regional.json" OUTPUT_DIR=$(HELMFILE_ARGOCD_VALUES_DIR) \
		deployment/helmfile/scripts/generate_config_jsons.sh

.PHONY: helmfile-generate-argocd-az-values
helmfile-generate-argocd-az-values: secrets
	cd deployment/helmfile && \
		$(HELMFILE) \
		write-values \
		--file helmfile.yaml \
		--selector geographicScope=az \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-deps \
		--output-file-template "$(HELMFILE_ARGOCD_VALUES_DIR)/idc-regional/{{ .Release.Labels.region }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml" \
		$(HELMFILE_OPTS)
	cd deployment/helmfile && \
		$(HELMFILE) \
		list \
		--file helmfile.yaml \
		--selector geographicScope=az \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-charts \
		--output json > $(TMPDIR)/releaselist-az.json
	ENVIRONMENT_NAME="${HELMFILE_ENVIRONMENT}" ENVIRONMENT_TYPE="az" RELEASE_LIST_JSON="$(TMPDIR)/releaselist-az.json" OUTPUT_DIR=$(HELMFILE_ARGOCD_VALUES_DIR) \
		deployment/helmfile/scripts/generate_config_jsons.sh

.PHONY: helmfile-generate-argocd-az-network-values
helmfile-generate-argocd-az-network-values: secrets download-helm-chart-versions
	cd deployment/helmfile && \
		$(HELMFILE) \
		write-values \
		--file helmfile.yaml \
		--selector geographicScope=az-network \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-deps \
		--output-file-template "$(HELMFILE_ARGOCD_VALUES_DIR)/idc-network/{{ .Release.Labels.region }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml" \
		$(HELMFILE_OPTS)
	cd deployment/helmfile && \
		$(HELMFILE) \
		list \
		--file helmfile.yaml \
		--selector geographicScope=az-network \
		--environment $(HELMFILE_ENVIRONMENT) \
		--skip-charts \
		--output json > $(SECRETS_DIR)/releaselist-az-network.json
	ENVIRONMENT_NAME="${HELMFILE_ENVIRONMENT}" ENVIRONMENT_TYPE="az-network" RELEASE_LIST_JSON="$(SECRETS_DIR)/releaselist-az-network.json" OUTPUT_DIR=$(HELMFILE_ARGOCD_VALUES_DIR) \
		deployment/helmfile/scripts/generate_config_jsons.sh

.PHONY: add-api-server-to-no-proxy
add-api-server-to-no-proxy:
	@-[ ! -f ${HOME}/.local/kind-env ] && mkdir -p ${HOME}/.local && touch ${HOME}/.local/kind-env
	@-[ ! "$$(grep -w $(KIND_API_SERVER_ADDRESS) ${HOME}/.local/kind-env)" ] && echo "no_proxy=\"$$no_proxy,$(KIND_API_SERVER_ADDRESS)\"" > ${HOME}/.local/kind-env
	@-[ ! "$$(grep -w kind-env ${HOME}/.bashrc)" ] && echo source '${HOME}/.local/kind-env' >> ${HOME}/.bashrc
	@source ${HOME}/.local/kind-env

.PHONY: install-requirements
install-requirements:
	deployment/common/install_requirements.sh
	$(MAKE) -C idcs_domain/bmaas/bmvs install-requirements

.PHONY: setup-bmvs
setup-bmvs:
ifneq ($(SKIP_BMVS),1)
	GUEST_HOST_DEPLOYMENTS=$(GUEST_HOST_DEPLOYMENTS) GUEST_HOST_MEMORY_MB=$(GUEST_HOST_MEMORY_MB) HTTP_SERVER_PORT=$(HTTP_SERVER_PORT) $(MAKE) -C idcs_domain/bmaas/bmvs setup
endif

.PHONY: teardown-bmvs
teardown-bmvs:
ifneq ($(SKIP_BMVS),1)
	GUEST_HOST_DEPLOYMENTS=$(GUEST_HOST_DEPLOYMENTS) HTTP_SERVER_PORT=$(HTTP_SERVER_PORT) $(MAKE) -C idcs_domain/bmaas/bmvs teardown
endif

.PHONY: deploy-netbox-kind
deploy-netbox-kind: deploy-netbox deploy-netbox-samples ## Deploy Netbox on kind.

.PHONY: deploy-netbox-samples
deploy-netbox-samples: ## Populate Netbox database with sample records.
	GUEST_HOST_DEPLOYMENTS=$(GUEST_HOST_DEPLOYMENTS) deployment/common/netbox/populate_samples.sh

.PHONY: undeploy-vault
undeploy-vault: helmfile ## Undeploy Vault Server and Agent Injector.
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=vault --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)
	-kubectl delete pvc -n kube-system -l app.kubernetes.io/instance=vault

.PHONY: deploy-vault
deploy-vault: secrets undeploy-vault deploy-vault-releases deploy-vault-operator-init deploy-vault-configure deploy-vault-secrets ## Deploy Vault Server and Agent Injector.

.PHONY: deploy-vault-releases
deploy-vault-releases: ## Deploy Vault Server and Agent Injector Helm releases.
	cd deployment/helmfile && $(HELMFILE) apply --selector chart=vault --environment $(HELMFILE_ENVIRONMENT) --wait $(HELMFILE_OPTS)

.PHONY: deploy-vault-operator-init
deploy-vault-operator-init: $(KUBECTL) $(JQ) ## Initialize (unseal) Vault Server.
	KUBECTL=$(KUBECTL) \
	JQ=$(JQ) \
	deployment/common/vault/vault-operator-init.sh

.PHONY: deploy-vault-configure
deploy-vault-configure: secrets helmfile-dump $(VAULT) $(YQ) $(JQ) ## Configure Vault.
	VAULT=$(VAULT) \
	YQ=$(YQ) \
	JQ=$(JQ) \
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) \
	VAULT_DRY_RUN=false \
	deployment/common/vault/configure.sh

.PHONY: test-vault-configure
test-vault-configure: secrets helmfile-dump $(VAULT) $(YQ) $(JQ) ## Configure Vault (dry-run).
	VAULT=$(VAULT) \
	YQ=$(YQ) \
	JQ=$(JQ) \
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) \
	VAULT_DRY_RUN=true \
	deployment/common/vault/configure.sh

.PHONY: deploy-vault-secrets
deploy-vault-secrets: secrets helmfile-dump $(VAULT) $(YQ) ## Deploy all secrets to Vault.
	VAULT=$(VAULT) \
	YQ=$(YQ) \
	JQ=$(JQ) \
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) \
	deployment/common/vault/load-secrets.sh

.PHONY: deploy-vault-secrets-for-nw-cluster
deploy-vault-secrets-for-nw-cluster: secrets $(VAULT) ## Deploy those secrets needed for the SDN-controller to vault.
	VAULT=$(VAULT) \
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) \
	go/pkg/sdn-controller/vault-load-secrets-for-nw-cluster.sh

.PHONY: deploy-vault-to-network-cluster
deploy-vault-to-network-cluster: ## Deploy Vault Agent Injector Helm release to the network services cluster (only)
	cd deployment/helmfile && $(HELMFILE) apply --selector chart=vault,geographicScope=az-network --environment $(HELMFILE_ENVIRONMENT) --wait $(HELMFILE_OPTS)

.PHONY: undeploy-vault-from-network-cluster
undeploy-vault-from-network-cluster: ## Deploy Vault Agent Injector Helm release to the network services cluster (only)
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=vault,geographicScope=az-network --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: generate-vault-pki-test-cert
generate-vault-pki-test-cert: $(VAULT) ## Generate test certificate and key from Vault PKI.
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) COMMON_NAME=testclient1 CREATE_ROLE=1 hack/generate-vault-pki-cert.sh

.PHONY: generate-vault-pki-cert
generate-vault-pki-cert: $(VAULT) ## Generate certificate and key from Vault PKI. Set COMMON_NAME.
	VAULT_TOKEN=$(shell cat $(VAULT_TOKEN_FILE)) hack/generate-vault-pki-cert.sh

.PHONY: deploy-k8s-image-pull-secrets
deploy-k8s-image-pull-secrets: deploy-k8s-image-pull-secrets-idcs-system deploy-k8s-image-pull-secrets-idcs-enrollment ## Create image pull secrets in Kubernetes cluster.

deploy-k8s-image-pull-secrets-%: ## Create image pull secret in specified namespace in Kubernetes cluster.
	-kubectl create namespace $*
	-kubectl delete secret -n $* idc-image-pull-secret
	kubectl create secret docker-registry \
		-n $* \
		idc-image-pull-secret \
		--docker-server="$(DOCKER_REGISTRY)" \
		--docker-username="$(HARBOR_USERNAME)" \
		--docker-password="$(HARBOR_PASSWORD)"

.PHONY: deploy-k8s-tls-secrets
deploy-k8s-tls-secrets: ## Create TLS secrets in Kubernetes cluster.
	hack/deploy-k8s-tls-secrets.sh

.PHONY: deploy-crds
deploy-crds: helmfile ## Deploy Kubernetes Custom Resource Definitions
	cd deployment/helmfile && $(HELMFILE) sync --selector crd=true --environment $(HELMFILE_ENVIRONMENT) --wait $(HELMFILE_OPTS)

.PHONY: deploy-restricted-serviceaccounts
deploy-restricted-serviceaccounts: helm-push-sdn-restricted-sa helmfile ## Deploy Kubernetes Restricted ServiceAccounts (eg. accounts with limited RBACs)
	cd deployment/helmfile && $(HELMFILE) sync --allow-no-matching-release --selector restrictedServiceaccount=true --environment $(HELMFILE_ENVIRONMENT) --wait $(HELMFILE_OPTS)

.PHONY: deploy-registry
deploy-registry: ## Deploy a local Docker registry
	LOCAL_REGISTRY_NAME=$(LOCAL_REGISTRY_NAME) \
	LOCAL_REGISTRY_PORT=$(KIND_REGISTRY_PORT) \
	deployment/registry/start_registry.sh

.PHONY: create-localstack-resources
create-localstack-resources:
	deployment/localstack/init.sh

.PHONY: deploy-kind
deploy-kind: helmfile-dump update-etc-hosts $(YQ) ## Deploy a new kind cluster.
ifneq ($(KUBECONFIG),)
	@echo "KUBECONFIG must not be set" && false
endif
	YQ=$(YQ) \
	LOCAL_REGISTRY_NAME=$(LOCAL_REGISTRY_NAME) \
	LOCAL_REGISTRY_PORT=$(KIND_REGISTRY_PORT) \
	sg docker -c "deployment/kind/deploy-kind.sh"

.PHONY: deploy-network-cluster-in-kind
deploy-network-cluster-in-kind: update-etc-hosts ## Deploy a new kind cluster.
ifneq ($(KUBECONFIG),)
	@echo "KUBECONFIG must not be set" && false
endif
	sg docker -c "deployment/kind/deploy-network-cluster-in-kind.sh"

.PHONY: deploy-k3s
deploy-k3s: update-etc-hosts ## Deploy a new k3s cluster.
	sg docker -c "deployment/k3s/deploy-k3s.sh"

.PHONY: deploy-rke2
deploy-rke2: update-etc-hosts ## Deploy a new rke2 cluster.
	sg docker -c "deployment/rke2/deploy-rke2.sh"

deploy-k8s-infrastructure-helm-releases: helm-chart-versions deploy-coredns deploy-restricted-serviceaccounts generate-sdn-restricted-kubeconfigs ## Deploy Kubernetes infrastructure Helm releases (coredns)

generate-sdn-restricted-kubeconfigs:
	deployment/network/generate-restricted-kubeconfigs.sh

.PHONY: deploy-coredns-to-network-cluster
deploy-coredns-to-network-cluster: helmfile helm-push-coredns undeploy-coredns-from-network-cluster
	cd deployment/helmfile && $(HELMFILE) apply --selector chart=coredns,cluster=network --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: undeploy-coredns-from-network-cluster
undeploy-coredns-from-network-cluster: helmfile helm-push-coredns
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector chart=coredns,cluster=network --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

## Undeploy database.
UNDEPLOY_DBS := $(DBS:%=undeploy-%-db)
.PHONY: $(UNDEPLOY_DBS)
$(UNDEPLOY_DBS): undeploy-%: helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)
	-kubectl delete pvc -n idcs-system -l app.kubernetes.io/instance=$*

# Deploy database.
DEPLOY_DBS := $(DBS:%=deploy-%-db)
.PHONY: $(DEPLOY_DBS)
$(DEPLOY_DBS): deploy-%: undeploy-% helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: undeploy-pgoperator
undeploy-pgoperator: undeploy-$(REGION)-pgoperator ## Undeploy pg operator.

.PHONY: deploy-pgoperator
deploy-pgoperator: deploy-$(REGION)-pgoperator ## Deploy pg operator.

.PHONY: undeploy-$(REGION)-pgoperator
undeploy-$(REGION)-pgoperator: undeploy-%: helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: deploy-$(REGION)-pgoperator
deploy-$(REGION)-pgoperator: deploy-%: undeploy-% helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: undeploy-miniooperator
undeploy-minioOperator: undeploy-$(REGION)-minioOperator undeploy-$(REGION)-minioTenant ## Undeploy minio operator along with minio tenant.

.PHONY: deploy-miniooperator
deploy-miniooperator: deploy-$(REGION)-miniooperator deploy-$(REGION)-miniotenant ## Deploy minio operator alsong with minio tenant.

.PHONY: undeploy-$(REGION)-miniooperator
undeploy-$(REGION)-miniooperator: undeploy-%: helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: deploy-$(REGION)-miniooperator
deploy-$(REGION)-miniooperator: deploy-%: undeploy-% helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: undeploy-$(REGION)-miniotenant
undeploy-$(REGION)-miniotenant: undeploy-%: helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) destroy --allow-no-matching-release --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

.PHONY: deploy-$(REGION)-miniotenant
deploy-$(REGION)-miniotenant: deploy-%: undeploy-% helmfile helm-chart-versions
	cd deployment/helmfile && $(HELMFILE) apply --selector name=$* --environment $(HELMFILE_ENVIRONMENT) $(HELMFILE_OPTS)

# Help for generated targets
ifeq (1,0)
undeploy-cloudaccount-db: Undeploy cloudaccount database.
deploy-cloudaccount-db: ## Deploy cloudaccount database for kind.
undeploy-usage-db: ## Undeploy usage database.
deploy-usage-db: ## Deploy usage database for kind.
undeploy-metering-db: ## Undeploy metering database.
deploy-metering-db: ## Deploy metering database for kind.
undeploy-cloudcredits-db: ## Undeploy cloudcredits database.
deploy-cloudcredits-db: ## Deploy cloudcredits database for kind.
undeploy-billing-db: ## Undeploy billing database.
deploy-billing-db: ## Deploy billing database for kind.
undeploy-notification-db: ## Undeploy notifications database.
deploy-notification-db: ## Deploy billing notifications for kind.
undeploy-productcatalog-db: ## Undeploy productcatalog database.
deploy-productcatalog-db: ## Deploy productcatalog for kind.
endif

.PHONY: undeploy-compute-db
undeploy-compute-db: undeploy-$(REGION)-compute-db ## Undeploy compute database.

.PHONY: deploy-compute-db
deploy-compute-db: deploy-$(REGION)-compute-db ## Deploy compute database.

.PHONY: undeploy-storage-db
undeploy-storage-db: undeploy-$(REGION)-storage-db ## Undeploy storage database.

.PHONY: deploy-storage-db
deploy-storage-db: deploy-$(REGION)-storage-db ## Deploy compute database.
.PHONY: undeploy-training-db
undeploy-training-db: undeploy-$(REGION)-training-db ## Undeploy training database.


.PHONY: undeploy-quota-management-service-db
undeploy-quota-management-service-db: undeploy-$(REGION)-quota-management-service-db ## Undeploy quota management database.

.PHONY: deploy-quota-management-service-db
deploy-quota-management-service-db: deploy-$(REGION)-quota-management-service-db ## Deploy quota management database.

.PHONY: deploy-training-db
deploy-training-db: deploy-$(REGION)-training-db ## Deploy training database.

.PHONY: undeploy-iks-db
undeploy-iks-db: undeploy-$(REGION)-iks-db ## undeploy IKS database

.PHONY: deploy-iks-db
deploy-iks-db: deploy-$(REGION)-iks-db

.PHONY: deploy-insights-db
deploy-insights-db: deploy-$(REGION)-insights-db

.PHONY: undeploy-insights-db
undeploy-insights-db: undeploy-$(REGION)-insights-db ## undeploy Insights database

.PHONY: undeploy-kfaas-db
undeploy-kfaas-db: undeploy-$(REGION)-kfaas-db ## undeploy KFAAS database

.PHONY: deploy-kfaas-db
deploy-kfaas-db: deploy-$(REGION)-kfaas-db

.PHONY: undeploy-cloudmonitor-db
undeploy-cloudmonitor-db: undeploy-cloudmonitor-db ## undeploy cloudmonitor database

.PHONY: deploy-cloudmonitor-db
deploy-cloudmonitor-db: deploy-cloudmonitor-db

.PHONY: undeploy-cloudmonitor-logs-db
undeploy-cloudmonitor-logs-db: undeploy-$(REGION)-cloudmonitor-logs-db ## undeploy cloudmonitor-logs database

.PHONY: deploy-cloudmonitor-logs-db
deploy-cloudmonitor-logs-db: deploy-$(REGION)-cloudmonitor-logs-db

.PHONY: undeploy-dpai-db
undeploy-dpai-db: undeploy-$(REGION)-dpai-db ## undeploy DPAI database

.PHONY: deploy-dpai-db
deploy-dpai-db: deploy-$(REGION)-dpai-db


.PHONY: deploy-network-services
deploy-network-services: deploy-vault-to-network-cluster apply-sdn-controller-crds deploy-sdn-controller deploy-sdn-controller-rest deploy-sdn-integrity-checker deploy-switch-config-saver

.PHONY: undeploy-network-services
undeploy-network-services: undeploy-vault-from-network-cluster undeploy-sdn-controller-crds undeploy-sdn-controller undeploy-sdn-controller-rest undeploy-sdn-integrity-checker undeploy-switch-config-saver

# opal doesn't have docker images for arm64
.PHONY: build-opal-images
build-opal-images: ## build opal docker images
ifeq ($(UNAME_ARCH), aarch64)
	if ! curl -sL http://localhost:$(KIND_REGISTRY_PORT)/v2/permitio%2fopal-server/tags/list \
		| jq -c .tags | grep -q $(OPAL_VERSION); then \
		set -e; \
		rm -rf local/opa-build; \
		mkdir -p local/opa-build; \
		(cd local/opa-build; git clone -b $(OPAL_VERSION) https://github.com/permitio/opal.git); \
		docker build -t localhost:$(KIND_REGISTRY_PORT)/permitio/opal-client:$(OPAL_VERSION) --target client \
			-f local/opa-build/opal/docker/Dockerfile local/opa-build/opal; \
		docker build -t localhost:$(KIND_REGISTRY_PORT)/permitio/opal-client-standalone:$(OPAL_VERSION) --target client-standalone \
			-f local/opa-build/opal/docker/Dockerfile local/opa-build/opal; \
		docker build -t localhost:$(KIND_REGISTRY_PORT)/permitio/opal-server:$(OPAL_VERSION) --target server \
			--build-arg TRUST_POLICY_REPO_HOST_SSH_FINGERPRINT=false \
			-f local/opa-build/opal/docker/Dockerfile local/opa-build/opal; \
		docker image push localhost:$(KIND_REGISTRY_PORT)/permitio/opal-client:$(OPAL_VERSION); \
		docker image push localhost:$(KIND_REGISTRY_PORT)/permitio/opal-client-standalone:$(OPAL_VERSION); \
		docker image push localhost:$(KIND_REGISTRY_PORT)/permitio/opal-server:$(OPAL_VERSION); \
		rm -rf local/opa-build; \
	fi
endif

.PHONY: deploy-populate-compute-db
deploy-populate-compute-db: deploy-git-to-grpc-synchronizer ## Populate MachineImage, InstanceType, and other records in the Compute DB.

.PHONY: run-vault-validation
run-vault-validation: $(TERRAFORM)
	TERRAFORM=$(TERRAFORM) \
	deployment/common/vault/terraform/scripts/update_validation.sh

##@ Universe Deployer

.PHONY: universe-deployer
universe-deployer: main-universe-deployer-git-pusher ## Run Universe Deployer.

.PHONY: main-universe-deployer-git-pusher
main-universe-deployer-git-pusher: bazel
	$(BAZEL) run \
		--jobs $(UNIVERSE_DEPLOYER_JOBS_PER_PIPELINE) \
		--sandbox_writable_path=${UNIVERSE_DEPLOYER_POOL_DIR} \
		$(BAZEL_OPTS) \
		//deployment/universe_deployer/main_universe:main_universe_deployer_git_pusher -- \
		--patch-command deployment/universe_deployer/main_universe/main_universe_patches.sh \
		--push-state-file-name push_state_main_universe_deployer.json \
		--source-sequence-number $(GIT_REV_COUNT) \
		--source-git-remote $(GIT_URL) \
		--source-git-branch $(GIT_BRANCH)

.PHONY: universe-deployer-set-commit-main
universe-deployer-set-commit-main: universe-deployer-set-commit-prod universe-deployer-set-commit-staging ## Update all commits in all Universe Config files.

.PHONY: universe-deployer-set-commit-prod
universe-deployer-set-commit-prod: ## Update all commits in Universe Config file prod.json.
	sed -i "s/\"commit\": \".*\"/\"commit\": \"`git rev-parse HEAD`\"/" universe_deployer/environments/prod.json

.PHONY: universe-deployer-set-commit-staging
universe-deployer-set-commit-staging: ## Update all commits in Universe Config file staging.json.
	sed -i "s/\"commit\": \".*\"/\"commit\": \"`git rev-parse HEAD`\"/" universe_deployer/environments/staging.json

.PHONY: test-universe-deployer-git-push
test-universe-deployer-git-push: ## Update all commits in Universe Config File and push to Git
	# Commit all changes if needed.
	git add -A -v
	git diff --quiet && git diff --staged --quiet || git commit -m "git-push: add all"

	# Update universe config to reference the latest commit.
	sed -i "s/\"commit\": \".*\"/\"commit\": \"`git rev-parse HEAD`\"/" $(TEST_UNIVERSE_CONFIG_FILE)

	# Commit universe config.
	git add -A -v
	git commit -m "git-push: update $(TEST_UNIVERSE_CONFIG_FILE)"
	git push

.PHONY: test-universe-deployer-git-pusher
test-universe-deployer-git-pusher: test-universe-deployer-git-push universe-deployer ## Update all commits in Universe Config File and run universe_deployer

.PHONY: build-manifests-generator
build-manifests-generator: bazel
	rm -rf local/manifests-generator
	$(BAZEL) build $(BAZEL_OPTS) //deployment/universe_deployer/deployment_artifacts:deployment_artifacts_tar
	mkdir -p local/manifests-generator/commit
	tar -C local/manifests-generator/commit -x -f bazel-bin/deployment/universe_deployer/deployment_artifacts/deployment_artifacts_tar.tar
	$(BAZEL) build $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/manifests_generator:manifests_generator

.PHONY: run-manifests-generator
run-manifests-generator: build-manifests-generator ## Run Universe Deployer Manifests Generator with IDC_ENV and all components set to current commit. Use for manual testing.
	bazel-bin/go/pkg/universe_deployer/cmd/manifests_generator/manifests_generator_/manifests_generator \
		--commit $(RUN_MANIFESTS_GENERATOR_GIT_COMMIT) \
		--commit-dir $(ROOT_DIR)/local/manifests-generator/commit \
		--output $(ROOT_DIR)/local/manifests-generator/manifests.tar \
		--secrets-dir $(SECRETS_DIR) \
		--snapshot \
		--universe-config $(ROOT_DIR)/${UNIVERSE_CONFIG}
	mkdir -p local/manifests-generator/manifests
	tar -C local/manifests-generator/manifests -x -f local/manifests-generator/manifests.tar
	@echo Manifests have been generated in local/manifests-generator/manifests

.PHONY: test-manifests-generator
test-manifests-generator: bazel  ## Test Manifests Generator.
	$(BAZEL) test \
		$(BAZEL_OPTS) \
		//go/pkg/universe_deployer/manifests_generator:manifests_generator_test

.PHONY: run-universe-config-annotate
run-universe-config-annotate: bazel ## Annotate Universe Config files.
	$(BAZEL) run $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/universe_config_cli:universe_config_cli -- \
		annotate $(ROOT_DIR)/universe_deployer/environments/*.json \
		--git-repository-dir $(ROOT_DIR)

.PHONY: test-universe-config-annotation
test-universe-config-annotation: bazel ## Test that Universe Config file annotations are up-to-date.
	$(BAZEL) run $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/universe_config_cli:universe_config_cli -- \
		test-annotate $(ROOT_DIR)/universe_deployer/environments/*.json \
		--git-repository-dir $(ROOT_DIR)

.PHONY: run-universe-config-print
run-universe-config-print: bazel ## Print Universe Config files. Run with `EXTRA_OPTS="--help"` for more options.
	$(BAZEL) run $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/universe_config_cli:universe_config_cli -- \
		print $(ROOT_DIR)/universe_deployer/environments/{prod,staging}.json \
		--git-repository-dir $(ROOT_DIR) \
		--render-mode pretty \
		$(EXTRA_OPTS)

.PHONY: run-universe-config-print-by-date
run-universe-config-print-by-date: bazel ## Print Universe Config files, sorted by date. Run with `EXTRA_OPTS="--help"` for more options.
	$(BAZEL) run $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/universe_config_cli:universe_config_cli -- \
		print $(ROOT_DIR)/universe_deployer/environments/{prod,staging}.json \
		--git-repository-dir $(ROOT_DIR) \
		--render-mode pretty \
		--sort authorDate \
		$(EXTRA_OPTS)

.PHONY: universe-config-csv
universe-config-csv: bazel ## Export Universe Config files to a CSV file.
	$(BAZEL) run $(BAZEL_OPTS) \
		//go/pkg/universe_deployer/cmd/universe_config_cli:universe_config_cli -- \
		print $(ROOT_DIR)/universe_deployer/environments/*.json \
		--git-repository-dir $(ROOT_DIR) \
		--render-mode csv \
		$(EXTRA_OPTS) \
		> local/universe_config.csv
	@echo Wrote local/universe_config.csv

##@ Build Dependencies

.PHONY: bazel ## Install Bazel and start Bazel remote cache.
bazel: bazelisk start-bazel-remote-cache generate-dynamic

.PHONY: bazelisk
bazelisk: $(BAZELISK) ## Download Bazelisk locally if necessary.
$(BAZELISK):
	mkdir -p $(LOCALBIN)
	test -s $(BAZELISK) || { wget --tries 1 --timeout=30 -q \
	https://github.com/bazelbuild/bazelisk/releases/download/$(BAZELISK_VERSION)/bazelisk-linux-$(BAZELISK_ARCH) \
	-O $(BAZELISK).tmp && mv $(BAZELISK).tmp $(BAZELISK) && chmod +x $(BAZELISK); }

.PHONY: start-bazel-remote-cache
start-bazel-remote-cache: ## Start Bazel "remote" cache locally in a Docker container. This caches build outputs for the entire local machine.
	LOCAL_BAZEL_REMOTE_CACHE_GRPC_PORT=$(LOCAL_BAZEL_REMOTE_CACHE_GRPC_PORT) \
	LOCAL_BAZEL_REMOTE_CACHE_NAME=$(LOCAL_BAZEL_REMOTE_CACHE_NAME) \
	build/bazel-remote-cache/start-bazel-remote-cache.sh

.PHONY: stop-bazel-remote-cache
stop-bazel-remote-cache: ## Stop Bazel remote cache. See also "make clean-bazel-remote-cache" to delete the volume.
	docker rm --force --volumes "$(LOCAL_BAZEL_REMOTE_CACHE_NAME)"

.PHONY: generate-dynamic
generate-dynamic: ## Generate dynamic files used by Bazel.
	echo -n "$(DOCKER_TAG)" > build/dynamic/DOCKER_TAG
	echo -n "$(HELM_CHART_VERSION)" > build/dynamic/HELM_CHART_VERSION
	echo -n "$(IDC_FULL_VERSION)" > build/dynamic/IDC_FULL_VERSION
	echo "DOCKER_REGISTRY = \"${DOCKER_REGISTRY}\"" > build/dynamic/docker_registry.bzl
	echo "DOCKER_IMAGE_PREFIX = \"${DOCKER_IMAGE_PREFIX}\"" >> build/dynamic/docker_registry.bzl
	echo "HELM_REGISTRY = \"${HELM_REGISTRY}\"" > build/dynamic/helm_registry.bzl
	echo "HELM_PROJECT = \"${HELM_PROJECT}\"" >> build/dynamic/helm_registry.bzl
	echo "MAX_POOL_DIRS = ${UNIVERSE_DEPLOYER_BUILDS_PER_HOST}" > build/dynamic/universe_deployer.bzl
	echo "POOL_DIR = \"${UNIVERSE_DEPLOYER_POOL_DIR}\"" >> build/dynamic/universe_deployer.bzl

.PHONY: protoc
protoc: $(PROTOC) ## Install Protobuf and GRPC if necessary.
$(PROTOC):
	($(PROTOC) --version | grep -q --line-regexp --fixed-strings "libprotoc 3.$(PROTOC_VERSION)") || { \
	    echo Installing protoc $(PROTOC_VERSION) && \
		wget -q https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip \
			-O /tmp/protoc-$(PROTOC_VERSION)-linux-x86_64.zip && \
		unzip -o /tmp/protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d $(HOME)/.local ; }
	$(PROTOC) --version | grep -q --line-regexp --fixed-strings "libprotoc 3.$(PROTOC_VERSION)"

install-go-packages: bazel
	$(BAZEL) run @go_sdk//:bin/go -- install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_VERSION)
	$(BAZEL) run @go_sdk//:bin/go -- install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)
	$(BAZEL) run @go_sdk//:bin/go -- install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v$(GRPC_GATEWAY_VERSION)
	$(BAZEL) run @go_sdk//:bin/go -- install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v$(GRPC_GATEWAY_VERSION)
	$(BAZEL) run @go_sdk//:bin/go -- install github.com/envoyproxy/protoc-gen-validate@$(PROTO_VALIDATE_VERSION)

grpcurl: $(GRPCURL) ## Download gRPCurl locally if necessary.
$(GRPCURL):
	mkdir -p $(LOCALBIN)
	test -s $(GRPCURL) || { wget -q \
	https://github.com/fullstorydev/grpcurl/releases/download/v$(GRPCURL_VERSION)/grpcurl_$(GRPCURL_VERSION)_linux_x86_64.tar.gz \
	-O - | tar -xzv -C $(TMPDIR) && mv $(TMPDIR)/grpcurl $(GRPCURL); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): bazel
	mkdir -p $(LOCALBIN)
	test -s $(CONTROLLER_GEN) || { \
	GOBIN=$(TMPDIR) $(BAZEL) run @go_sdk//:bin/go -- install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION) && \
	mv $(TMPDIR)/controller-gen $(CONTROLLER_GEN) ; }

.PHONY: mockgen
mockgen: bazel ## Download mockgen locally.
	$(BAZEL) run @go_sdk//:bin/go -- install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: codegen-tool
codegen-tool: bazel ## Download Kubernetes codegen tool.
	mkdir -p $(LOCALBIN)
	test -s $(CODEGEN_GENERATE_GROUPS) || { \
	wget -q https://github.com/kubernetes/code-generator/archive/refs/tags/v$(CODEGEN_VERSION).tar.gz -O - | \
	tar -xzv -C $(LOCALBIN) ; }

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE):
	mkdir -p $(LOCALBIN)
	test -s $(KUSTOMIZE) || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(TMPDIR) && \
	mv $(TMPDIR)/kustomize $(KUSTOMIZE) ; }

.PHONY: helm
helm: $(HELM) ## Download helm locally if necessary.
$(HELM):
	mkdir -p $(LOCALBIN)
	test -s $(HELM) || { wget -q \
	https://get.helm.sh/helm-$(HELM_VERSION)-linux-$(HELM_ARCH).tar.gz \
	-O - | tar -xzv -C $(TMPDIR) && mv $(TMPDIR)/linux-$(HELM_ARCH)/helm $(HELM); }

.PHONY: helmfile
helmfile: $(HELMFILE) helm ## Install helmfile including plugins.
	-$(HELM) plugin update diff
	$(HELMFILE) init --force $(HELMFILE_OPTS)

$(HELMFILE):
	mkdir -p $(LOCALBIN)
	test -s $(HELMFILE) || { wget -q \
	https://github.com/helmfile/helmfile/releases/download/v$(HELMFILE_VERSION)/helmfile_$(HELMFILE_VERSION)_linux_$(HELMFILE_ARCH).tar.gz \
	-O - | tar -xzv -C $(TMPDIR) && mv $(TMPDIR)/helmfile $(HELMFILE); }

.PHONY: vault
vault: $(VAULT) FORCE ## Download vault (CLI) locally if necessary.
	mkdir -p $(LOCALBIN)
	ln -f -s $(VAULT) $(LOCALBIN)/vault

FORCE: # Placeholder target to force a rebuild of any target that depends on it.

$(VAULT):
	mkdir -p $(LOCALBIN)
	test -s $(VAULT) || { wget -q \
	https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_linux_$(VAULT_ARCH).zip \
	-O /tmp/vault-$(VAULT_VERSION).zip && \
	unzip -o /tmp/vault-$(VAULT_VERSION).zip -d /tmp && \
	mv /tmp/vault $(VAULT); }

.PHONY: argocd
argocd: $(ARGOCD) ## Download argocd (CLI) locally if necessary.
$(ARGOCD):
	mkdir -p $(LOCALBIN)
	test -s $(ARGOCD) || { \
  	curl -Lo /tmp/argocd-$(ARGOCD_VERSION) https://github.com/argoproj/argo-cd/releases/download/$(ARGOCD_VERSION)/argocd-linux-amd64 && \
  	chmod +x /tmp/argocd-$(ARGOCD_VERSION) && \
  	mv /tmp/argocd-$(ARGOCD_VERSION) $(ARGOCD); }

.PHONY: kubectl
kubectl: $(KUBECTL) ## Download kubectl locally if necessary.
$(KUBECTL):
	mkdir -p $(LOCALBIN)
	test -s $(KUBECTL) || { \
  	curl -Lo /tmp/kubectl https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/linux/amd64/kubectl && \
  	chmod +x /tmp/kubectl && \
  	mv /tmp/kubectl $(KUBECTL); }

.PHONY: terraform
terraform: $(TERRAFORM) ## Download terraform (CLI) locally if necessary.
$(TERRAFORM):
	mkdir -p $(LOCALBIN)
	test -s $(TERRAFORM) || { \
  	wget -q https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_linux_amd64.zip \
	-O /tmp/terraform-$(TERRAFORM_VERSION).zip && \
	unzip -o -d /tmp /tmp/terraform-$(TERRAFORM_VERSION).zip terraform && \
	mv /tmp/terraform $(TERRAFORM); }

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary.
$(YQ):
	mkdir -p $(LOCALBIN)
	test -s $(YQ) || { wget -q \
	https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_linux_amd64 \
	-O $(YQ).tmp && mv $(YQ).tmp $(YQ) && chmod +x $(YQ); }

.PHONY: jq
jq: $(JQ) ## Download yq locally if necessary.
$(JQ):
	mkdir -p $(LOCALBIN)
	test -s $(JQ) || { wget -q \
	https://github.com/jqlang/jq/releases/download/jq-$(JQ_VERSION)/jq-linux-amd64 \
	-O $(JQ).tmp && mv $(JQ).tmp $(JQ) && chmod +x $(JQ); }

.PHONY: aws
aws: $(AWS) ## Download aws locally if necessary.
$(AWS):
	mkdir -p $(LOCALBIN)
	test -s $(AWS) || { wget -q \
	https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip \
	-O /tmp/awscli-exe-linux-x86_64.zip && \
	unzip -o /tmp/awscli-exe-linux-x86_64.zip -d /tmp && \
	/tmp/aws/install --bin-dir $(LOCALBIN) --install-dir $(LOCALBIN)/aws-cli; }
	$(AWS) --version

.PHONY: golang-ci
golang-ci: $(GOLANGCI_LINT) ## Download golang-ci locally if necessary.
$(GOLANGCI_LINT):
	mkdir -p $(LOCALBIN)
	test -s $(GOLANGCI_LINT) || { curl -sL \
	https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz | \
  tar xzf - golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64/golangci-lint && \
	mv golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64/golangci-lint $(GOLANGCI_LINT);}

##@ Misc.

.PHONY: install-interactive-tools
install-interactive-tools: protoc $(BAZELISK) $(GRPCURL) $(HELM) helmfile vault argocd yq install-vault-requirements ## Install tools useful for interactive development.
	cp $(ARGOCD) $(HOME)/.local/bin/argocd
	cp $(BAZELISK) $(HOME)/.local/bin/bazel
	cp $(GRPCURL) $(HOME)/.local/bin/grpcurl
	cp $(HELM) $(HOME)/.local/bin/helm
	cp $(HELMFILE) $(HOME)/.local/bin/helmfile
	cp $(VAULT) $(HOME)/.local/bin/vault
	cp $(YQ) $(HOME)/.local/bin/yq
	sudo apt update -y
	sudo apt install -y \
		jq \
		lcov \
		python3-pip \
		unzip \
		yarn

install-vault-requirements: ## Install Vault Python dependencies
	pip3 install --user --requirement deployment/common/vault/requirements.txt

install-testing-dashboards-requirements: ## Install testing dashboard Python dependencies
	pip3 install --user pymongo
	pip3 install --user argparse
	pip3 install --user numpy

.PHONY: update-etc-hosts
update-etc-hosts: ## If needed, add required entries to /etc/hosts (may require sudo).
	deployment/common/etc-hosts/update-etc-hosts.sh

.PHONY: show-make-config
show-make-config: show-config

.PHONY: show-config
show-config:  ## Show key Makefile configuration variables.
	@echo "show-config: Running with the following make configuration:"
	@echo "show-config:   IDC_ENV=$(IDC_ENV)"
	@echo "show-config:   HELMFILE_ENVIRONMENT=$(HELMFILE_ENVIRONMENT)"
	@echo "show-config:   SECRETS_DIR=$(SECRETS_DIR)"
	@echo "show-config:   GIT_COMMIT=$(GIT_COMMIT)"
	@echo "show-config:   IDC_FULL_VERSION=$(IDC_FULL_VERSION)"
	@echo "show-config:   DOCKER_REGISTRY=$(DOCKER_REGISTRY)"
	@echo "show-config:   DOCKER_TAG=$(DOCKER_TAG)"
	@echo "show-config:   DOCKERIO_IMAGE_PREFIX=$(DOCKERIO_IMAGE_PREFIX)"
	@echo "show-config:   HELM_CHART_VERSION=$(HELM_CHART_VERSION)"
	@echo "show-config:   OTEL_DEPLOYMENT_ENVIRONMENT=$(HELMFILE_ENVIRONMENT)"
	@echo "show-config:   REGION=$(REGION)"
	@echo "show-config:   VAULT_ADDR=$(VAULT_ADDR)"
	@echo "show-config:   HOST_IP=$(HOST_IP)"
	@echo "show-config:   KUBECONFIG=$(KUBECONFIG)"
	@echo "show-config:   NETWORKING_KUBECONFIG=$(NETWORKING_KUBECONFIG)"

.PHONY: show-bazel-info
show-bazel-info: $(BAZEL) ## Show output of "bazel info"
	$(BAZEL) info

.PHONY: show-export
show-export:  ## Show commands to export environment-specific variables. Run with "eval `make show-export`".
	@echo "export IDC_ENV=$(IDC_ENV)"
	@echo "export SECRETS_DIR=$(SECRETS_DIR)"
	@echo "export REGION=$(REGION)"
	@echo "export PREFIXLENGTH=$(PREFIXLENGTH)"
	@echo "export IDC_GLOBAL_URL_PREFIX=$(IDC_GLOBAL_URL_PREFIX)"
	@echo "export IDC_REGIONAL_URL_PREFIX=$(IDC_REGIONAL_URL_PREFIX)"
	@echo "export IDC_OIDC_URL_PREFIX=$(IDC_OIDC_URL_PREFIX)"

dump-makefile: ## Dump Makefile variables (verbose).
	@echo DEPLOY_IDC_DEPS = $(DEPLOY_IDC_DEPS)
	@echo BAZEL_CONTAINERS = $(BAZEL_CONTAINERS)

helm-chart-version: ## Print the version of IDC Helm releases (DEPRECATED, use helm-chart-versions instead).
	@echo $(IDC_FULL_VERSION)

idc-version: ## Print the IDC version
	@echo $(IDC_FULL_VERSION)

.PHONY: helm-chart-versions
helm-chart-versions: bazel ## Generate Helm chart version files.
	rm -rf local/idc-versions/$(IDC_FULL_VERSION)
	$(BAZEL) build $(BAZEL_OPTS) //deployment:all_chart_versions_except_idc_versions
	mkdir -p $(HELM_CHART_VERSIONS_DIR)
	cp bazel-bin/deployment/chart_versions/* $(HELM_CHART_VERSIONS_DIR)

.PHONY: download-helm-chart-versions
download-helm-chart-versions: helm ## Download Helm chart version files.
	rm -rf local/idc-versions/$(IDC_FULL_VERSION)
	$(HELM) pull --untar --destination local/idc-versions/$(IDC_FULL_VERSION) oci://$(HELM_REGISTRY)/$(HELM_PROJECT)/idc-versions --version $(IDC_FULL_VERSION)

run-k9s: ## Run k9s.
	@echo KUBECONFIG=$(KUBECONFIG)
	k9s

# See also: https://bazel.build/query/guide
# Other examples:
#  bazel query "somepath(//go/test/compute/e2e/vm:vm_test, //go/pkg/baremetal_enrollment/bmc:bmc)"
bazel-query-deps: ## List dependencies of Bazel target. For example: BAZEL_QUERY_DEPS=//go/test/compute/e2e/vm:vm_test make bazel-query-deps
	$(BAZEL) query "deps($(BAZEL_QUERY_DEPS))"

.PHONY: bazel-query-rdeps-go
bazel-query-rdeps-go: ## List Go targets that depend on a set of Go modules. For example: BAZEL_QUERY_RDEPS=@sh_helm_helm_v3//... make bazel-query-rdeps-go
	$(BAZEL) query "rdeps(//go/..., $(BAZEL_QUERY_RDEPS))"

##@ Secrets

.PHONY: secrets
secrets: secrets-sh $(SECRETS) ## Generate missing secrets.

.PHONY: secrets-sh
secrets-sh: helmfile-dump $(YQ) ## Generate missing secrets using make-secrets.sh.
	YQ=$(YQ) \
	deployment/common/vault/make-secrets.sh

.PHONY: access-dev-vault
access-dev-vault: vault
	@NO_PROXY= no_proxy= $(VAULT) write auth/approle/login role_id=${VAULT_DEV_CREDENTIALS_USR} secret_id=${VAULT_DEV_CREDENTIALS_PSW} | grep token | head -1 | awk '{print$$2}' > $(SECRETS_DIR)/DEV_VAULT_TOKEN
	NO_PROXY= no_proxy= $(VAULT) status

.PHONY: get-dev-secrets
get-dev-secrets: vault yq ## Get secrets needed to deploy IDC to a development environment.
	AVAILABILITY_ZONE=$(REGION)a \
	GET_DB_USER_SECRETS=true \
	SECRET_DIR=$(SECRETS_DIR) \
	WORKSPACE=$(ROOT_DIR) \
	VAULT=$(VAULT) \
	VAULT_ADDR=https://internal-placeholder.com \
	YQ=$(YQ) \
	no_proxy= \
	NO_PROXY= \
	deployment/jenkinsfile/vault/remote/get-secrets-remote.sh

.PHONY: save-dev-db-admin-secrets
save-dev-db-admin-secrets: vault ## Get secrets needed to deploy IDC to a development environment.
	AVAILABILITY_ZONE=$(REGION)a \
	SAVE_DB_USER_SECRETS=true \
	SECRET_DIR=$(SECRETS_DIR) \
	WORKSPACE=$(ROOT_DIR) \
	VAULT=$(VAULT) \
	VAULT_ADDR=https://internal-placeholder.com \
	no_proxy= \
	NO_PROXY= \
	deployment/jenkinsfile/vault/remote/save-db-admin-secrets-remote.sh

.PHONY: save-dev-secrets
save-dev-secrets: vault ## Get secrets needed to deploy IDC to a development environment.
	AVAILABILITY_ZONE=$(REGION)a \
	SECRET_DIR=$(SECRETS_DIR) \
	WORKSPACE=$(ROOT_DIR) \
	VAULT=$(VAULT) \
	VAULT_ADDR=https://internal-placeholder.com \
	no_proxy= \
	NO_PROXY= \
	deployment/jenkinsfile/vault/remote/save-secrets-remote.sh

.PHONY: get-dev3-secrets
get-dev3-secrets: vault ## Get secrets needed to deploy IDC to dev3.
	AVAILABILITY_ZONES="us-dev3-1a us-dev3-2a" \
	REGIONS="us-dev3-1 us-dev3-2" \
	SECRET_DIR=$(SECRETS_DIR) \
	WORKSPACE=$(ROOT_DIR) \
	VAULT=$(VAULT) \
	VAULT_ADDR=$(VAULT_DEV_ADDR) \
	no_proxy= \
	NO_PROXY= \
	deployment/jenkinsfile/vault/remote/get-dev3-secrets.sh

$(SECRETS_DIR):
	mkdir -p $@

$(IKS_KUBERNETES_OPERATOR_CONFIG):
	@echo "Warning: Add missing values to kubernetes-operator config.yaml"
	mkdir -p $(IKS_KUBERNETES_OPERATOR_CONFIG)
	cp go/pkg/kubernetes_operator/config.yaml $(IKS_KUBERNETES_OPERATOR_CONFIG)/config.yaml
	cp go/pkg/kubernetes_operator/bootstrap-scripts/* $(IKS_KUBERNETES_OPERATOR_CONFIG)/

$(IKS_KUBERNETES_RECONCILER_CONFIG):
	@echo "Warning: Add missing values to kubernetes-reconciler config.yaml"
	mkdir -p $(IKS_KUBERNETES_RECONCILER_CONFIG)
	cp go/pkg/kubernetes_reconciler/config.yaml $(IKS_KUBERNETES_RECONCILER_CONFIG)/config.yaml

$(IKS_ILB_OPERATOR_CONFIG):
	@echo "Warning: Add missing values to ilb-operator config.yaml"
	mkdir -p $(IKS_ILB_OPERATOR_CONFIG)
	cp go/pkg/ilb_operator/config.yaml $(IKS_ILB_OPERATOR_CONFIG)/config.yaml

## Makefile utilities
bazel-run-%: # Run a generic Bazel target.
	$(BAZEL) run $(BAZEL_OPTS) $*

retry-%: # Attempt to build Makefile target multiple times.
	hack/retry.sh $(MAKE) $*

### CICD
.PHONY: get-target-images
get-target-images: bazelisk ## Get target images from the latest build.
	$(BAZEL) build $(BAZEL_OPTS) //hack/scanner-reporter:scanner_reporter

.PHONY: deployment-artifacts
deployment-artifacts: bazel ## Build deployment artifacts tar file.
	$(BAZEL) build $(BAZEL_OPTS) \
		deployment/universe_deployer/deployment_artifacts:deployment_artifacts_tar

create-coupon: ## Create a coupon. For example: TOKEN="..." IDC_ENV=dev27 make create-coupon
	go/pkg/billing/test-scripts/create_coupon.sh

##@ BMaaS utilities
.PHONY: show-bm-validation-versions
show-bm-validation-versions:
	@pip3 -q install tabulate
	@python3 ./idcs_domain/bmaas/scripts/bmaas_validation_versions.py

.PHONY: save-bm-validation-versions
save-bm-validation-versions:
	@pip3 -q install tabulate
	@pip3 -q install openpyxl
	@python3 ./idcs_domain/bmaas/scripts/bmaas_validation_versions.py --save
