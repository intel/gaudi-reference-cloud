#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
export PGPASSWORD=$(cat ../../../local/secrets/COMPUTE_DB_ADMIN_PASSWORD)
CONTAINER_NAME=${CONTAINER_NAME:-idc-db}
docker rm -f ${CONTAINER_NAME} | true
docker run -d --rm --name ${CONTAINER_NAME} -p 5432:5432 -e POSTGRES_PASSWORD=${PGPASSWORD} library/postgres:11
