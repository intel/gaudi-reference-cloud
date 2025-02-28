#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

SECRETS_DIR=${SECRETS_DIR:-${SCRIPT_DIR}/../../../../local/secrets}
OPERATOR_DIR=${OPERATOR_DIR:-${SECRETS_DIR}/vm-cluster-operator}

# Execute a vault CLI command without polluting script environment
vault_helper() {
  local token=$1
  shift
  (
    export VAULT_ADDR=http://localhost:30990/
    export VAULT_TOKEN=${token}
    vault "$@"
  )
}

vault_admin() {
  vault_helper $(cat ${SECRETS_DIR}/VAULT_TOKEN) "$@"
}

vault_operator() {
  vault_helper ${OPERATOR_TOKEN} "$@"
}

OPERATOR_ROLE=$(vault_admin read -format=json auth/approle/role/${AZONE}-vm-cluster-operator-role/role-id | jq -r .data.role_id)
OPERATOR_SECRET=$(vault_admin write -force -format=json auth/approle/role/${AZONE}-vm-cluster-operator-role/secret-id | jq -r .data.secret_id)
OPERATOR_TOKEN=$(vault_admin write -format=json auth/approle/login role_id=${OPERATOR_ROLE} secret_id=${OPERATOR_SECRET} | jq -r .auth.client_token)

VAULT_ROLE=$(vault_operator read -format=json auth/approle/role/${AZONE}-vm-worker-role/role-id | jq -r .data.role_id)
WRAPPED_VAULT_TOKEN=$(vault_operator write -wrap-ttl=20m -force -format=json auth/approle/role/${AZONE}-vm-worker-role/secret-id | jq -r .wrap_info.token)

proxy_files=""
if [[ ! -z ${HTTPS_PROXY+x} ]]; then
  proxy_files=$(cat <<EOF
    - path: /etc/apt/apt.conf.d/90-proxy
      content: |
        Acquire::http::Proxy "${HTTP_PROXY}";
        Acquire::https::Proxy "${HTTPS_PROXY}";
    - path: /etc/profile.d/90-proxy.sh
      content: |
        export HTTP_PROXY=${HTTP_PROXY}
        export HTTPS_PROXY=${HTTPS_PROXY}
        export NO_PROXY=${NO_PROXY}
        export http_proxy=${HTTP_PROXY}
        export https_proxy=${HTTPS_PROXY}
        export no_proxy=${NO_PROXY}
EOF
  )
fi

# The curl wrapper is used to workaround a bug in quoting by Rancher's system-agent-install.sh script
cat <<EOF | yq -o json | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances --data-binary @- \
| jq .
metadata:
  name: ${NAME}
spec:
  availabilityZone: ${AZONE}
  instanceType: ${INSTANCE_TYPE}
  machineImage: ${MACHINE_IMAGE}
  runStrategy: RerunOnFailure
  sshPublicKeyNames:
  - ${KEYNAME}
  interfaces:
  - name: eth0
    vNet: ${VNETNAME}
  userData: |
    #cloud-init
    write_files:
${proxy_files}
    - path: /usr/local/bin/curl
      permissions: '755'
      content: |
$(sed -e 's/^/        /' ${SCRIPT_DIR}/curl-wrapper.sh)
    - path: /opt/idc/kubevirt-worker-node/bootstrap-kv-worker.py
      permissions: '755'
      content: |
$(sed -e 's/^/        /' ${SCRIPT_DIR}/bootstrap-kv-worker.py)
    runcmd:
    - /opt/idc/kubevirt-worker-node/bootstrap-kv-worker.py --cluster-name ${CLUSTER_NAME} --vault-addr ${VAULT_ADDR} --vault-token ${WRAPPED_VAULT_TOKEN} --vault-role ${VAULT_ROLE} --rancher-credentials-path ${REGION}/${AZONE}/rancher
EOF
