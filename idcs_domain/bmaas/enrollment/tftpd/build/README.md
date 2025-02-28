This directory contains files for building the tftp-server image for BMaaS.
To update the image, run the commands below:
```shell
export IMAGE_VERSION=vXXX
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f tftp-server.Dockerfile -t  amr-idc-registry-pre.infra-host.com/idc-devops/tftp-server:$IMAGE_VERSION .
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f tftp-server.Dockerfile -t  amr-idc-registry.infra-host.com/idc-devops/tftp-server:$IMAGE_VERSION .
docker push amr-idc-registry-pre.infra-host.com/idc-devops/tftp-server:$IMAGE_VERSION
docker push amr-idc-registry.infra-host.com/idc-devops/tftp-server:$IMAGE_VERSION
```
