# OpenAPI Docker Image CLI

The release of OpenAPI CLI used (v6.2.1) creates code which does
not pass Coverity scan checker.

This folder contains scripts and patch(es) to generate a Docker image
with fixes for the flagged Coverity issues.

> ### **TODO**
> These patches should be upstreamed to the Open Source project.

## Download the sources and patch
```bash
git clone -b v6.2.1 --single-branch https://github.com/OpenAPITools/openapi-generator
cd openapi-generator
git am ../0001-Fix-gosec-ignoring-returned-errors.patch
```

## Install the build requisites

```bash
sudo apt install docker
```

## Build the Docker Image

```bash
docker build -t openapitools/openapi-generator-cli .
```

## Publish Locally and Test the Image

To test that new docker image generated the excepted code, you should
run a local Docker Registry, publish the image there, and update your
build environment to use you locally hosted image.

A guide for running a local Docker Registry in Docker is
https://docs.docker.com/registry/

Download and run a local Docker Registry:
```bash
docker run -d -p 5000:5000 --name registry registry:2.7
```

Now tag and publish the new OpenAPI Docker CLI image:
```bash
docker tag openapitools/openapi-generator-cli localhost:5000/openapitools/openapi-generator-cli
docker push localhost:5000/openapitools/openapi-generator-cli
```

Find the SHA256 of the new image:
```bash
docker inspect --format='{{index .RepoDigests 0}}' localhost:5000/openapitools/openapi-generator-cli
```

Update the your local build `Makefile.environment` with the new SHA and to
local Docker Registry by adding the following to your
`build/environments/pdx03-c01-azcp001-vm-X/Makefile.environment` file.

```diff
+OPENAPI_DOCKER_PREFIX ?= localhost:5000/
+OPENAPI_GENERATOR_TAG ?= @sha256:793fa1835cc9816a8bd905cd6bf64e915987c2c87ffd3dce485e2dd7a35cab71
+
```

Regenerate the code and review the changes:

```bash
make generate-go-openapi generate-go-openapi-raven fmt gazelle
git diff -a
```

Finally, confirm that the new code builds:
```bash
make build
```

It would be wise to also do some basic testings to ensure functionality.


## Publish the Docker Image (External)

To avoid proxy issues during the image build, it is suggested
to build this image externally, outside of the Intel firewall.

### Publish External to Intel

To get the CLI password to login with Harbor, open the website in a browser, login.
In the upper right-hand corner, select "User Profile" under you name.
Hit the "copy button" on the right of the "CLI Secret".
This is the password to supply to the interactive login.
export HARBOR_USERNAME="<YOUR_INTEL_EMAIL>"

```bash
docker tag openapitools/openapi-generator-cli amr-idc-registry-pre.infra-host.com/intelcloud/openapitools/openapi-generator-cli
docker login -u ${HARBOR_USERNAME} amr-idc-registry-pre.infra-host.com
docker push amr-idc-registry-pre.infra-host.com/intelcloud/openapitools/openapi-generator-cli
```

## Update the Makefile for the new image

Get the SHA256
```bash
docker inspect --format='{{index .RepoDigests 0}}' amr-idc-registry-pre.infra-host.com/intelcloud/openapitools/openapi-generator-cli
```

Update the `Makefile` with the new SHA in the `OPENAPI_GENERATOR_TAG`.

We also made a minor modification to always use the externally published OpenAPI Docker image.

```diff
index 6cf85e59ac7c..79a476876d83 100644
--- a/Makefile
+++ b/Makefile
@@ -48,8 +48,8 @@ BAZELISK_VERSION ?= v1.15.0
 GRPCURL_VERSION ?= 1.8.7
 TIMESCALE_CHART_VERSION ?= 0.27.5
 POSTGRES_CHART_VERSION ?= 12.2.6
-# Tag "v6.2.1" as of 2022-12-19
-OPENAPI_GENERATOR_TAG ?= @sha256:df74501bac3192a45a1b41d38e2db749b662882678d701666776ff3e1895b347
+# Tag "v6.2.1" plus Intel patch
+OPENAPI_GENERATOR_TAG ?= @sha256:632fa682a9802997b81c390d7edc383a58987909aa34e422667a597e48724530
 CONTROLLER_TOOLS_VERSION ?= v0.10.0
 CODEGEN_VERSION ?= v0.25.2
 KUSTOMIZE_VERSION ?= v4.5.5
@@ -180,6 +180,8 @@ DOCKERIO_REGISTRY_WITH_PREFIX ?= $(DOCKERIO_REGISTRY)
 else
 DOCKERIO_REGISTRY_WITH_PREFIX ?= $(DOCKERIO_REGISTRY)/$(DOCKERIO_REPOSITORY_PREFIX:/=)
 endif
+# We are using an Intel patched version
+OPENAPI_DOCKER_PREFIX ?= amr-idc-registry-pre.infra-host.com/intelcloud/

 # Calculated variables.
 UNAME_ARCH := $(shell uname -p)
@@ -543,7 +545,7 @@ generate-go-openapi: ## Generate Go OpenAPI clients.
        rm -rf go/pkg/compute_api_server/openapi
        docker run --rm -v $(shell pwd):/local \
                -u $(shell id -u $(USER)):$(shell id -g $(USER)) \
-               $(DOCKERIO_IMAGE_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
+               $(OPENAPI_DOCKER_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
                generate \
                        --input-spec /local/public_api/proto/compute.swagger.json \
                        --generator-name go \
@@ -558,7 +560,7 @@ generate-go-openapi-raven: ## Generate Go OpenAPI clients for Raven
        rm -rf go/pkg/raven/openapi
        docker run --rm -v $(shell pwd):/local \
                -u $(shell id -u $(USER)):$(shell id -g $(USER)) \
-               $(DOCKERIO_IMAGE_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
+               $(OPENAPI_DOCKER_PREFIX)openapitools/openapi-generator-cli$(OPENAPI_GENERATOR_TAG) \
                generate \
        --input-spec /local/go/pkg/raven/swagger.json \
        --generator-name go \
```

## Regenerate the code using the new image

```bash
make generate-go-openapi generate-go-openapi-raven fmt gazelle
```
