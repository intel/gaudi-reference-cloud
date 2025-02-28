#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
NAME=vault
NAMESPACE=kube-system
SECRETS_DIR=${SECRETS_DIR:-/tmp}
JQ=${JQ:-jq}
KUBECTL=${KUBECTL:-kubectl}

wait_pod_vault_0() {
    # wait for 'vault-0` pod, it would show 0/1 until we enable it
    set +e
    for (( ; ; ))
    do
        ${KUBECTL} get pods --namespace $NAMESPACE | grep vault-0 | grep Running
        if [[ $? = 0 ]]
        then
          break
        fi
        echo "waiting..."
        sleep 4
    done
    set -e
}

vault_operator_init() {
  ${KUBECTL} exec vault-0 \
    --namespace $NAMESPACE  \
    -- vault operator init \
    -key-shares=1 -key-shares=1 \
    -key-threshold=1 \
    -format=json > ${SECRETS_DIR}/vault-cluster-keys.json
}

wait_pod_vault_0
vault_operator_init

# Create a variable named VAULT_UNSEAL_KEY to capture the Vault unseal key
VAULT_UNSEAL_KEY=$(${JQ} -r ".unseal_keys_b64[]" ${SECRETS_DIR}/vault-cluster-keys.json)
# Unseal Vault running on the vault-0 pod
${KUBECTL} exec vault-0 --namespace $NAMESPACE -- vault operator unseal $VAULT_UNSEAL_KEY
# capture root key
ROOT_KEY=$(${JQ} -r ".root_token" ${SECRETS_DIR}/vault-cluster-keys.json)
echo $ROOT_KEY
# login to vault
${KUBECTL} exec  -n $NAMESPACE vault-0 -- vault login $ROOT_KEY

### save root key in ${SECRETS_DIR}
echo "save root key in ${SECRETS_DIR}/VAULT_ROOT_KEY"
echo ${ROOT_KEY} > ${SECRETS_DIR}/VAULT_ROOT_KEY
cp ${SECRETS_DIR}/VAULT_ROOT_KEY ${SECRETS_DIR}/VAULT_TOKEN

# Get public key of all Kubernetes clusters.
KUBECTL=${KUBECTL} ${SCRIPT_DIR}/get-kubernetes-public-keys.sh