#!/bin/bash

# Build the Docker image
docker build -t my_pytorch_image -f /tmp/Dockerfile .

# Run a container using the built image
docker run --runtime=habana -e HABANA_VISIBLE_DEVICES=all -e OMPI_MCA_btl_vader_single_copy_mechanism=none \
           --cap-add=sys_nice --net=host --ipc=host my_pytorch_image
