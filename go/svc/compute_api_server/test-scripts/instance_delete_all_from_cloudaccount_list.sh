#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Delete all instances from all CloudAccounts in a file containing CloudAccount IDs.
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
CLOUDACCOUNTSFILE=${CLOUDACCOUNTSFILE:-${SCRIPT_DIR}/cloudaccounts.txt}
cat ${CLOUDACCOUNTSFILE} | xargs -i -n 1 -- env CLOUDACCOUNT={} ${SCRIPT_DIR}/instance_delete_all.sh
