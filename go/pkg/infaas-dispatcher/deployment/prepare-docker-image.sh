#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -e

if [ -z "$1" ]; then echo "invalid service (not provided)"; exit 1; fi
service=$1

echo building $service docker image...
make deploy-$service

# sets the env vars we need here
eval "$(teller sh)"

echo pushing image to cnvrg repo because why not...
# agent is created as infaas-inference, and the image is deployed to cnvrg/infaas. Don't ask... Need to fix that...
if [ "$service" == "infaas-inference" ]; then repo="cnvrg/infaas"; else repo="cnvrg/$service"; fi
echo deploying to dockerhub repo=$repo
imgTag=$(cat build/dynamic/DOCKER_TAG)
echo will use tag=$imgTag
docker pull localhost:5001/$service:$imgTag
docker tag localhost:5001/$service:$imgTag $repo:v666
docker push $repo:v666