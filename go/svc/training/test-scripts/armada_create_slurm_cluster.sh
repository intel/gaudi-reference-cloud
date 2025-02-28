#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/clusters --data-binary @- \
| jq .
{
  "cluster": {
    "name": "idc-static-slurm-cluster-test",
    "description": "Static slurm cluster replicating batch beta",
    "SSHKeyName": [
      "${KEYNAME}"
    ],
    "spec": {
            "region": "us-staging-1",
            "availabilityZone": "us-staging-1a",
            "prefixLength": 24
    },
    "storageNodes": [
      {
        "name": "weka-1",
        "description": "Weka Storage Node #1",
        "capacity": "50GB",
        "accessMode": "STORAGE_READ_WRITE",
        "mount": "STORAGE_WEKA",
        "localMountDir": "/share",
        "remoteMountDir": "/export/share"
      }
    ],
    "nodes": [
      {
        "count": 2,
        "imageType": "${MACHINE_IMAGE}",
        "machineType": "${SLURM_CLUSTER_MACHINE_TYPE}",
        "role": "JUPYTERHUB_NODE",
        "labels": {
          "role": "jupyterhub-node"
        }
      },
      {
        "count": 2,
        "imageType": "${MACHINE_IMAGE}",
        "machineType": "${SLURM_CLUSTER_MACHINE_TYPE}",
        "role": "LOGIN_NODE",
        "labels": {
          "role": "login-node"
        }
      },
      {
        "count": 1,
        "imageType": "${MACHINE_IMAGE}",
        "machineType": "${SLURM_CLUSTER_MACHINE_TYPE}",
        "role": "CONTROLLER_NODE",
        "labels": {
          "role": "controller-node"
        }
      },
      {
        "count": 1,
        "imageType": "${MACHINE_IMAGE}",
        "machineType": "${SLURM_CLUSTER_MACHINE_TYPE}",
        "role": "COMPUTE_NODE",
        "labels": {
          "role": "compute-node"
        }
      }
    ]
  }
}
EOF
