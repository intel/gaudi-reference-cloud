BIN_DIR := ./bin

ENROLLMENT_SERVICE_BIN:=bmaas-enrollment
ENROLLMENT_SERVICE_VERSION:=v1.0
IMAGE?=idcs/bmaas-enrollment:v1
NAMESPACE?=idcs-enrollment
VALUES_FILE?=charts/bmaas-enrollment-api/values.yaml
HELM_VERSION?=0.1.0
HTTP_PROXY?=http://internal-placeholder.com:911
HTTPS_PROXY?=http://internal-placeholder.com:912
NO_PROXY?=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,.intel.com,intel.com,localhost,127.0.0.1
GIN_MODE?=debug
REGISTRY?=internal-placeholder.com
REGISTRY_PROJECT?=intelcloud
REGISTRY_USER?=admin
REGISTRY_PW?=admin
CHART_PACKAGE_DIR?=/tmp/bmaas/enrollment/api
CHART_VERSION_SUFFIX?=aaaaaaa

$(BIN_DIR):
	mkdir -p ./bin

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
.DEFAULT_GOAL := help

## kind-deploy-enrollment: create kind cluster, build container image and deploy enrollment helm chart
.PHONY: kind-deploy-enrollment
kind-deploy-enrollment: kind-create-cluster build-image kind-load kind-customize-helm-chart helm-package kind-deploy-chart

## kind-undeploy-enrollment: delete kind cluster
.PHONY: kind-undeploy-enrollment
kind-undeploy-enrollment: kind-undeploy-chart
	kubectl delete -f config/kind/kind-k8s-service.yaml || true
	kind delete cluster --name idcs-dev

## kind-deploy-netbox: deploy netbox
.PHONY: kind-deploy-netbox
kind-deploy-netbox:
	KIND_CLUSTER=true ../netbox/install.sh

## kind-undeploy-netbox: undeploy netbox
.PHONY: kind-undeploy-netbox
kind-undeploy-netbox:
	helm del -n idcs-enrollment netbox --wait
	kubectl delete -f ../netbox/rbac.yaml

## kind-populate-netbox-samples: populate NetBox samples
.PHONY: kind-populate-netbox-samples
kind-populate-netbox-samples:
	../netbox/populate_samples.sh

## fmt: format go source code
.PHONY: fmt
fmt:
	go fmt ./...

## vet: examine go source code
.PHONY: vet
vet: fmt
	go vet ./...

## build: build enrollment api
.PHONY: build
build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(ENROLLMENT_SERVICE_BIN) ./cmd

## test: test go packages
.PHONY: test
test:
	go test ./... -v

## coverage: coverage reort of go packages
.PHONY: coverage
coverage:
	go test ./... -cover -covermode=count -coverprofile cover.out
	go tool cover -func cover.out

## lint-go: golang linter
.PHONY: lint-go
lint-go: fmt vet
	docker run --env HTTP_PROXY="$(HTTP_PROXY)" \
	--env HTTPS_PROXY="$(HTTPS_PROXY)" \
	--env NO_PROXY="$(NO_PROXY)" \
	--rm -v $(shell dirname $(shell pwd)):/app \
	-w /app/apiservice \
	internal-placeholder.com/intelcloud/golangci/golangci-lint:latest-alpine \
	golangci-lint run cmd api internal/...  --deadline=120s

## build-image: build bmaas-enrollment image
.PHONY: build-image
build-image:
	DOCKER_BUILDKIT=1 docker build --progress=plain --no-cache -t $(IMAGE) \
	    --build-arg http_proxy=$(HTTP_PROXY) \
		--build-arg https_proxy=$(HTTPS_PROXY) \
		--build-arg no_proxy=$(NO_PROXY) \
		--build-arg GIN_MODE=$(GIN_MODE) .

## ci-publish-image: build and publish bmaas-enrollment image to image repo
.PHONY: ci-publish-image
ci-publish-image: build-image image-registry-login image-push image-registry-logout

## ci-publish-chart: build and publish bmaas-enrollment helm chart to chart repo
.PHONY: ci-publish-chart
ci-publish-chart: helm-customize-chart helm-package helm-registry-login helm-push helm-registry-logout

.PHONY: image-registry-login
image-registry-login:
	@echo $(REGISTRY_PW) | docker login $(REGISTRY)/$(REGISTRY_PROJECT) -u $(REGISTRY_USER) --password-stdin

.PHONY: image-push
image-push:
	docker push $(IMAGE)

.PHONY: image-registry-logout
image-registry-logout:
	@docker logout $(REGISTRY)/$(REGISTRY_PROJECT)

.PHONY: helm-registry-login
helm-registry-login:
	@echo $(REGISTRY_PW) | helm registry login $(REGISTRY)/$(REGISTRY_PROJECT) -u $(REGISTRY_USER) --password-stdin

.PHONY: helm-customize-chart
helm-customize-chart:
	$(shell rm -rf $(CHART_PACKAGE_DIR))
	$(shell mkdir -p $(CHART_PACKAGE_DIR))
	$(shell cp -r charts $(CHART_PACKAGE_DIR))
	$(shell sed -i "/version:/ s/$$/-$(CHART_VERSION_SUFFIX)/" $(CHART_PACKAGE_DIR)/charts/bmaas-enrollment-api/Chart.yaml)
	$(shell sed -i 's/tag: v1/tag: $(CHART_VERSION_SUFFIX)/g' $(CHART_PACKAGE_DIR)/charts/bmaas-enrollment-api/values.yaml)

.PHONY: helm-package
helm-package:
	helm package $(CHART_PACKAGE_DIR)/charts/bmaas-enrollment-api -d $(CHART_PACKAGE_DIR)

.PHONY: helm-registry-logout
helm-registry-logout:
	@helm registry logout $(REGISTRY)/$(REGISTRY_PROJECT)

.PHONY: helm-push
helm-push:
	$(eval HELM_CHART := $(shell ls -lt /tmp/bmaas/enrollment/api/*.tgz | awk '{ print $$9 }'))
	helm push $(HELM_CHART) oci://$(REGISTRY)/$(REGISTRY_PROJECT)

## deploy: deploy enrollment helm chart to an existing cluster
.PHONY: deploy
deploy:
	helm upgrade bmaas-enrollment-api charts/bmaas-enrollment-api \
    --install \
    --version $(HELM_VERSION) \
    --create-namespace \
    --namespace $(NAMESPACE) \
    --values $(VALUES_FILE)

## deploy-dry-run: deployment dry-run of enrollment helm chart
.PHONY: deploy-dry-run
deploy-dry-run:
	helm upgrade bmaas-enrollment-api charts/bmaas-enrollment-api \
    --install \
    --version $(HELM_VERSION) \
    --create-namespace \
    --namespace $(NAMESPACE) \
    --values $(VALUES_FILE) --debug --dry-run

## undeploy: undeploy enrollment helm chart
.PHONY: undeploy
undeploy:
	helm uninstall bmaas-enrollment-api -n $(NAMESPACE)

.PHONY: kind-create-cluster
kind-create-cluster:
	kind create cluster --config config/kind/kind-config.yaml --name idcs-dev

.PHONY: kind-load
kind-load:
	kind load docker-image $(IMAGE) --name idcs-dev

.PHONY: kind-customize-helm-chart
kind-customize-helm-chart:
	$(shell rm -rf $(CHART_PACKAGE_DIR))
	$(shell mkdir -p $(CHART_PACKAGE_DIR))
	$(shell cp -r charts $(CHART_PACKAGE_DIR))


.PHONY: kind-deploy-chart
kind-deploy-chart:
	$(eval HELM_CHART := $(shell ls -lt /tmp/bmaas/enrollment/api/*.tgz | awk '{ print $$9 }'))
	helm upgrade bmaas-enrollment-api $(HELM_CHART) \
    --install \
    --version $(HELM_VERSION) \
    --create-namespace \
    --namespace $(NAMESPACE) \
    --values $(CHART_PACKAGE_DIR)/charts/bmaas-enrollment-api/values.yaml

	kubectl apply -f config/kind/kind-k8s-service.yaml

.PHONY: kind-undeploy-chart
kind-undeploy-chart:
	helm uninstall bmaas-enrollment-api -n $(NAMESPACE) || true

.PHONY: all
all: clean build

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(CHART_PACKAGE_DIR)

## help: print help
.PHONY : help
help : Makefile
	@sed -n 's/^##//p' $<
