#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
export PGPASSWORD=$(cat ../../../local/secrets/compute_db_user_password)
DBNAME=${DBNAME:-postgres}
psql -h localhost -p 5432 -U dbuser -d ${DBNAME} $*
