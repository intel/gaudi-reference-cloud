#!/bin/bash

export IMAGE_NAME="intel/intel-extension-for-pytorch"
export IMAGE_TAG="xpu-max"
sudo docker pull "$IMAGE_NAME:$IMAGE_TAG"

cd /tmp/

CURRENT_DIR="$(pwd)"
sudo docker run -d -v "$CURRENT_DIR":/workspace -v /dev/dri/by-path:/dev/dri/by-path --device /dev/dri --privileged "$IMAGE_NAME:$IMAGE_TAG" tail -f /dev/null

# Get the container ID
CONTAINER_ID=$(sudo docker ps -q)

sudo docker exec "$CONTAINER_ID" python /workspace/train_cifar10.py >/tmp/dockeroutput2.txt 2>&1