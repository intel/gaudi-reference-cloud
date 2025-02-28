#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
#  -d '{"serviceName": "compute"}' \
#  localhost:50051 proto.QuotaManagementService/GetStorageQuota
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

grpcurl --cacert local/secrets/pki/rohit/ca.pem --cert local/secrets/pki/rohit/cert.pem --key local/secrets/pki/rohit/cert.key \
-H '' -d '{"serviceName": "compute", "resourceType": "instances","cloudAccountId":"802429834639"}' \
dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:443 \
proto.QuotaManagementPrivateService/GetResourceQuotaPrivate
#cat <<EOF | \
#grpcurl --cacert local/secrets/pki/rohit/ca.pem --cert local/secrets/pki/rohit/cert.pem --key local/secrets/pki/rohit/cert.key \
#dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:443 proto.QuotaManagementPrivateService/GetResourceQuotaPrivate
#-H '' -d '{"serviceName": "compute", "resourceType": "instances", "cloudAccountId":"822189490624"}' \
