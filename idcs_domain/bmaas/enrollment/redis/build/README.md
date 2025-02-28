This directory contains files for building the Docker image for Alpine based Redis for Netbox.
To update the image, run the commands below:
```shell
export IMAGE_VERSION=7.2.3-alpine3.19-bash
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f redis-alpine.Dockerfile -t  amr-idc-registry-pre.infra-host.com/idc-devops/redis:$IMAGE_VERSION .
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f redis-alpine.Dockerfile -t  amr-idc-registry.infra-host.com/idc-devops/redis:$IMAGE_VERSION .
docker push amr-idc-registry-pre.infra-host.com/idc-devops/redis:$IMAGE_VERSION
docker push amr-idc-registry.infra-host.com/idc-devops/redis:$IMAGE_VERSION
```
