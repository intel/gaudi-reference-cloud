# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Load secrets into Vault.
# This script reads secrets from files in ${SECRETS_DIR} and writes them to Vault.

# Only meant to be executed as part of an UPGRADE to add the network-cluster to an existing IDC deployment.
# If you are installing all clusters, load-secrets.sh will load everything you need.

set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
: "${VAULT_TOKEN:?environment variable is required}"
: "${VAULT_ADDR:?environment variable is required}"
VAULT=${VAULT:-vault}
REGION=${REGION:-us-dev-1}
AVAILABILITY_ZONE=${AVAILABILITY_ZONE:-us-dev-1a}
SECRETS_DIR=${SECRETS_DIR:-local/secrets}
JWT_VALIDATION_PUBKEYS_FILE=${JWT_VALIDATION_PUBKEYS_FILE:-${SECRETS_DIR}/jwt_validation_pubkeys}
ENROLLMENT_KUBECONFIG=${ENROLLMENT_KUBECONFIG:-${SECRETS_DIR}/kubeconfig/kind-idc-${AVAILABILITY_ZONE}.yaml}
NETWORKING_KUBECONFIG=${NETWORKING_KUBECONFIG:-${SECRETS_DIR}/kubeconfig/kind-idc-${AVAILABILITY_ZONE}-network.yaml}
EAPI_USERNAME=$(cat ${SECRETS_DIR}/EAPI_USERNAME)
EAPI_PASSWD=$(cat ${SECRETS_DIR}/EAPI_PASSWD)
NETBOX_TOKEN=$(cat ${SECRETS_DIR}/NETBOX_TOKEN)


################################################################################
# Arista Switch eAPI
################################################################################
${VAULT} kv put -mount=controlplane \
  ${REGION}/${AVAILABILITY_ZONE}/nw-sdn-controller/eapi \
  username="${EAPI_USERNAME}" \
  password="${EAPI_PASSWD}"

################################################################################
# copy the az's kubeconfig to network cluster vault folder
################################################################################
${VAULT} kv put -mount=controlplane \
  ${REGION}/${AVAILABILITY_ZONE}/nw-sdn-controller/bmhkubeconfig \
  kubeconfig="@${ENROLLMENT_KUBECONFIG}"

################################################################################
# copy the network cluster's restricted kubeconfig to BM controller vault folder
################################################################################
# NETWORKING_KUBECONFIG is set in the "build/environments/<env_folder>/Makefile.environment" file
if [ -n "${NETWORKING_KUBECONFIG}" ]; then
    SDN_BMAAS_KUBECONFIG=${SECRETS_DIR}/restricted-kubeconfig/sdn-bmaas-kubeconfig.yaml
    echo "loading SDN Kubeconfig to Vault (for use by bm-instance-operator)"
    ${VAULT} kv put -mount=controlplane \
      ${AVAILABILITY_ZONE}-bm-instance-operator/sdnkubeconfig \
      kubeconfig="@${SDN_BMAAS_KUBECONFIG}"
fi

################################################################################
# List contents of Vault.
################################################################################

${VAULT} kv list public/
${VAULT} kv list controlplane/
${VAULT} kv get -mount=controlplane ${REGION}/baremetal/enrollment/approle


################################################################################
# Netbox token
################################################################################
${VAULT} kv put -mount=controlplane \
  ${REGION}/${AVAILABILITY_ZONE}/nw-sdn-controller/netboxtoken \
  token="${NETBOX_TOKEN}"

################################################################################
# Done.
################################################################################

echo vault-load-secrets-for-nw-cluster.sh: Done.
