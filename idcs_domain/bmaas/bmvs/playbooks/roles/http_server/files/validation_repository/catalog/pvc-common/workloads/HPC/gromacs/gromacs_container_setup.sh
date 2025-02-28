#!/bin/bash

docker build $(env | grep -E '(_proxy=|_PROXY)' | sed 's/^/--build-arg /') \
	-t gromacs2023 \
	-f Dockerfile \
	.

