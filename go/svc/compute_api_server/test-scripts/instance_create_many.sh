#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Create many instances in parallel.
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

FIRST=${FIRST:-1}
LAST=${LAST:-1}
PARALLELISM=${PARALLELISM:-1}
export CURL_OPTS="-k --silent"

seq --format=%04.0f ${FIRST} ${LAST} \
| xargs -i -n 1 --max-procs=${PARALLELISM} --verbose -- \
${SHELL} -c "NAME=instance-{} ${SCRIPT_DIR}/instance_create_with_name.sh"
