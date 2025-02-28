This directory contains files for building the Docker image for nginx-s3-gateway.
To update the image, run the commands below:
```shell
export IMAGE_VERSION=vXXX
git clone https://github.com/nginxinc/nginx-s3-gateway.git
cp nginx-s3-gateway.Dockerfile nginx-s3-gateway
cd nginx-s3-gateway
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f nginx-s3-gateway.Dockerfile -t  amr-idc-registry-pre.infra-host.com/idc-devops/nginx-s3-gateway:$IMAGE_VERSION .
docker build --build-arg http_proxy=http://internal-placeholder.com:911 --build-arg https_proxy=http://internal-placeholder.com:912 -f nginx-s3-gateway.Dockerfile -t  amr-idc-registry.infra-host.com/idc-devops/nginx-s3-gateway:$IMAGE_VERSION .
docker push amr-idc-registry-pre.infra-host.com/idc-devops/nginx-s3-gateway:$IMAGE_VERSION
docker push amr-idc-registry.infra-host.com/idc-devops/nginx-s3-gateway:$IMAGE_VERSION
```
