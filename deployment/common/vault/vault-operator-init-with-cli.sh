#!/usr/bin/env bash
# Same as vault-operator-init.sh but this uses the Vault CLI.

set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
NAME=vault
NAMESPACE=kube-system
SECRETS_DIR=${SECRETS_DIR:-/tmp}
JQ=${JQ:-jq}
KUBECTL=${KUBECTL:-kubectl}
VAULT=${VAULT:-vault}
VAULT_ADMIN_USERNAME="$(cat ${SECRETS_DIR}/vault_admin_username)"
VAULT_ADMIN_PASSWORD="$(cat ${SECRETS_DIR}/vault_admin_password)"

wait_for_vault() {
    set +e
    for (( ; ; ))
    do
        ${VAULT} status
        vault_status=$?
        if [[ $vault_status = 0 ]] || [[ $vault_status = 2 ]]; then
          break
        fi
        echo "Waiting for Vault..."
        sleep 1
    done
    set -e
}

vault_operator_init() {
  ${VAULT} operator init \
    -key-shares=1 -key-shares=1 \
    -key-threshold=1 \
    -format=json > ${SECRETS_DIR}/vault-cluster-keys.json
}

unseal() {
  # Get Vault unseal key.
  local vault_unseal_key=$(${JQ} -r ".unseal_keys_b64[]" ${SECRETS_DIR}/vault-cluster-keys.json)
  # Unseal Vault.
  ${VAULT} operator unseal ${vault_unseal_key}
  # Get root token.
  VAULT_ROOT_TOKEN=$(${JQ} -r ".root_token" ${SECRETS_DIR}/vault-cluster-keys.json)
  # Save root token to SECRETS_DIR. This should only be used for troubleshooting. It should not be used for deployment.
  # This root token will be deleted after the initial deployment.
  # Use the admin password to login instead of this root token.
  echo ${VAULT_ROOT_TOKEN} > ${SECRETS_DIR}/VAULT_ROOT_TOKEN
}

# Configure an admin user with a pre-defined password that can be used to login to Vault.
configure_admin() {
  VAULT_TOKEN=${VAULT_ROOT_TOKEN} ${VAULT} auth enable userpass

  VAULT_TOKEN=${VAULT_ROOT_TOKEN} ${VAULT} policy write admins - <<EOF
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF

  VAULT_TOKEN=${VAULT_ROOT_TOKEN} ${VAULT} write auth/userpass/users/${VAULT_ADMIN_USERNAME} \
    password="${VAULT_ADMIN_PASSWORD}" \
    policies=admins
}

login() {
  ${VAULT} login -method=userpass -token-only \
    username=${VAULT_ADMIN_USERNAME} \
    password="${VAULT_ADMIN_PASSWORD}" > ${SECRETS_DIR}/VAULT_TOKEN
  [[ $? -eq 0 ]] || return $?
  echo Vault admin token written to ${SECRETS_DIR}/VAULT_TOKEN.
}

if login; then
  echo Vault login succeeded. Vault initialization skipped.
else
  echo Vault login failed. Initializing Vault.
  wait_for_vault
  vault_operator_init
  unseal
  configure_admin
  login
fi

# Get public key of all Kubernetes clusters.
KUBECTL=${KUBECTL} ${SCRIPT_DIR}/get-kubernetes-public-keys.sh
