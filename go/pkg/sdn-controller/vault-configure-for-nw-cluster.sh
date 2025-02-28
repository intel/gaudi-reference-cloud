# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Configure Vault for IDC.
#
# This script does the following:
# - Enable secrets engines.
# - Enable auth methods.
# - Create roles for services.
# - Create policies for services.
#
# Secrets are NOT loaded by this script. See load-secrets.sh.
#
set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
VAULT=${VAULT:-vault}
# AUTH_PATH_GLOBAL must match global.vault.authPath in Helmfile environment.
AUTH_PATH_GLOBAL=${AUTH_PATH_GLOBAL:-auth/cluster-auth}
# AUTH_PATH_REGIONAL must match regions[].vault.authPath in Helmfile environment.
AUTH_PATH_REGIONAL=${AUTH_PATH_REGIONAL:-auth/cluster-auth}
BOUND_AUDIENCES_GLOBAL=${BOUND_AUDIENCES_GLOBAL:-https://kubernetes.default.svc.cluster.local}
BOUND_AUDIENCES_REGIONAL=${BOUND_AUDIENCES_REGIONAL:-https://kubernetes.default.svc.cluster.local}
IDC_ENV=${IDC_ENV:-kind}
REGION=${REGION:-us-dev-1}
AVAILABILITY_ZONE=${AVAILABILITY_ZONE:-us-dev-1a}
SECRETS_DIR=${SECRETS_DIR:-local/secrets}
DEFAULT_BAREMETAL_OPERATOR_NAMESPACE=("metal3-1")
BAREMETAL_OPERATOR_NAMESPACES=( "${BAREMETAL_OPERATOR_NAMESPACES[@]:-"${DEFAULT_BAREMETAL_OPERATOR_NAMESPACE[@]}"}" )
DEFAULT_NAMESPACE=idcs-system
# PKI common name limit is 64 characters.
PKI_ROOT_ID="$(uuidgen | head -c 8)"
PKI_ROOT_NAME="Intel IDC CA ${PKI_ROOT_ID} ${IDC_ENV}-root-ca"
PKI_INTERMEDIATE_NAME_PREFIX="Intel IDC CA ${PKI_ROOT_ID} "
: "${VAULT_TOKEN:?environment variable is required}"
: "${VAULT_ADDR:?environment variable is required}"

mkdir -p ${SECRETS_DIR}/pki

basic_service_global() {
  ca=$1-ca	# Should match pki issuer ca
  namespace=$2
  svc=$3  	# Should match the Helm release name.
  pol=${svc}-policy
  write_service_policy ${ca} ${namespace} "" ${svc} ${pol}
  write_service_roles ${ca} ${namespace} ${svc} ${pol} ${AUTH_PATH_GLOBAL} ${BOUND_AUDIENCES_GLOBAL}
}

basic_service_regional() {
  ca=$1-ca	# Should match pki issuer ca
  namespace=$2
  scope=$3
  svc=$4  	# Should match the Helm release name.
  pol=${scope}-${svc}-policy
  write_service_policy ${ca} ${namespace} ${scope} ${svc} ${pol}
  write_service_roles ${ca} ${namespace} ${scope}-${svc} ${pol} ${AUTH_PATH_REGIONAL} ${BOUND_AUDIENCES_REGIONAL}
}

# Creates a role and policy with a custom secrets endpoint
# Should not be used for adding another secrets endpoints to the policy (use add_service_policy() instead)
basic_service_with_policy_global() {
  ca=$1-ca  # Should match pki issuer ca
  namespace=$2
  svc=$3
  custom_endpoint=$4
  pol=${svc}-policy
  assign_service_policy ${ca} ${namespace} ${svc} ${custom_endpoint} ${pol}
  write_service_roles ${ca} ${namespace} ${svc} ${pol} ${AUTH_PATH_GLOBAL} ${BOUND_AUDIENCES_GLOBAL}
}

# Assign policy to access custom secrets endpoint-name
assign_service_policy() {
  ca=$1    # Should match pki issuer ca
  namespace=$2
  svc=$3   # Should match the Helm release name.
  custom_endpoint=$4   # Custom endpoint name for secrets.
  pol=$5   # Policy name.
  ${VAULT} policy write ${pol} - <<EOF
path "controlplane/data/${custom_endpoint}/*" {
  capabilities = ["read", "list"]
}
path "${ca}/issue/${svc}" {
  capabilities = ["update"]
}
EOF
}

# Add policy to access a secrets endpoint at /controlplane.
add_service_policy() {
  svc=$1  # Should match the Helm release name.
  endpoint=$2   # Should match the custom endpoint.
  pol=${svc}-policy
  existing_policy=$(${VAULT} read -field=rules sys/policy/${pol})
  policy_change="
path \"controlplane/data/${endpoint}/*\" {
  capabilities = [\"read\", \"list\"]
}"
  # update the existing svc default policy
  ${VAULT} write sys/policy/${pol} policy="${existing_policy}${policy_change}"
}

# Assign policy to access svc name as secrets endpoint
write_service_policy() {
  ca=$1   # Should match pki issuer ca
  namespace=$2
  scope=$3 # region or region+az
  svc=$4  # Should match the Helm release name.
  pol=$5  # Policy name.
  if [ -n "${scope}" ]
  then
    scopedSvc=${scope}-${svc}
  else
    scopedSvc=${svc}
  fi
  # controlplane/data/service-name - Global services
  # controlplane/data/$REGION-service-name & controlplane/data/$AZ-service-name - Legacy (because vault doesn't support giving a policy for users like controlplane/data/*-service-name/*)
  # controlplane/data/$region/service-name - regional services
  # controlplane/data/$region/$az/service-name - az-level services
  ${VAULT} policy write ${pol} - <<EOF
path "controlplane/data/${scopedSvc}/*" {
  capabilities = ["read", "list"]
}
path "controlplane/data/${scope}/${svc}/*" {
  capabilities = ["read", "list"]
}
path "controlplane/data/+/${scope}/${svc}/*" {
  capabilities = ["read", "list"]
}
path "${ca}/issue/${scopedSvc}" {
  capabilities = ["update"]
}
EOF
}

write_service_roles() {
  ca=$1	  # Should match pki issuer ca
  namespace=$2
  svc=$3  # Should match the Helm release name.
  pol=$4  # List of service-specific Vault policies.
  auth_path=$5
  bound_audiences=$6
  policies="public,global-pki"
  if [ -n "$pol" ]; then
    policies=${policies},$pol
  fi
  ${VAULT} write ${auth_path}/role/${svc}-role \
    role_type="jwt" \
    bound_audiences="${bound_audiences}" \
    user_claim="sub" \
    bound_subject="system:serviceaccount:${namespace}:${svc}" \
    policies="${policies}" \
    ttl="1h"
  ${VAULT} write ${ca}/roles/${svc} \
    allowed_domains="${svc}.${DEFAULT_NAMESPACE}.svc.cluster.local,*.local,*.internal-placeholder.com,*.eglb.intel.com,*.internal-placeholder.com" \
    allow_glob_domains=true \
    allow_bare_domains=true \
    allow_wildcard_certificates=false \
    ou="${svc}" \
    ttl="1h" \
    max_ttl="1h"
}

setup_intermediate_ca() {
	ca=$1-ca  # Should match pki issuer ca
	ca_path=${SECRETS_DIR}/pki/${ca}
	mkdir -p ${ca_path}

	# generate intermediate ca cert signing request
	${VAULT} write -format=json ${ca}/intermediate/generate/internal \
		common_name="${PKI_INTERMEDIATE_NAME_PREFIX}${ca}" \
		issuer_name="idc-${ca}" \
		ttl=8760h | jq -r '.data.csr' > ${ca_path}/${ca}.csr

	# sign intermediate ca cert using the root-ca
	${VAULT} write -format=json ${IDC_ENV}-root-ca/root/sign-intermediate \
		issuer_ref="idc-${IDC_ENV}-root-ca" \
		csr=@${ca_path}/${ca}.csr \
		format=pem_bundle ttl=8760h \
		| jq -r '.data.certificate' > ${ca_path}/${ca}.pem

	# set the ca cert signed by the root-ca
	${VAULT} write ${ca}/intermediate/set-signed \
		certificate=@${ca_path}/${ca}.pem

	# configure urls for the issuer
	${VAULT} write ${ca}/config/urls \
		issuing_certificates="${VAULT_ADDR}/v1/${ca}/ca" \
		crl_distribution_points="${VAULT_ADDR}/v1/${ca}/crl"

	# save ca-chain for the issuer
	${VAULT} read -format=json ${ca}/cert/ca_chain \
		| jq -r '.data.ca_chain' > ${ca_path}/ca-chain.pem
}

################################################################################
# Wait for Vault to come up.
################################################################################

while ! ${VAULT} secrets list; do sleep 1; done

################################################################################
# Enable Secrets Engines
################################################################################

${VAULT} secrets enable -path=${AVAILABILITY_ZONE}-network-ca -max-lease-ttl=8760h pki || true

# If Vault doesn't already have a ${IDC_ENV}-root-ca for this environment, uncomment this section to create one
# (Useful if reusing a vault that already has secrets for ANOTHER environment in it)
#################################################################################
## Configure PKI
#################################################################################
#
#${VAULT} secrets enable -path=${IDC_ENV}-root-ca -max-lease-ttl=8760h pki || true
#
## self-signed root-ca setup
#mkdir -p ${SECRETS_DIR}/pki/${IDC_ENV}-root-ca
#
## create a default root cert with ttl=1y
## if root private key required, use /exported endpoint instead
#${VAULT} write -field=certificate ${IDC_ENV}-root-ca/root/generate/internal \
#    common_name="${PKI_ROOT_NAME}" \
#    issuer_name="idc-${IDC_ENV}-root-ca" \
#    ttl=8760h | openssl x509 -text > ${SECRETS_DIR}/pki/${IDC_ENV}-root-ca/ca.pem || true
#
## configure urls for the issuer
#${VAULT} write ${IDC_ENV}-root-ca/config/urls \
#    issuing_certificates="${VAULT_ADDR}/v1/${IDC_ENV}-root-ca/ca" \
#    crl_distribution_points="${VAULT_ADDR}/v1/${IDC_ENV}-root-ca/crl"
#
## set revocaton list update expiry=10s
#${VAULT} write ${IDC_ENV}-root-ca/config/crl \
#    expiry=10s

# generate intermediate certificate, csr and sign with root-ca for each cluster
# single Vault cluster to be used for multiple dev environments, hence root-ca and global-ca need env prefix
setup_intermediate_ca ${AVAILABILITY_ZONE}-network

################################################################################
# Basic services
################################################################################

basic_service_regional ${AVAILABILITY_ZONE}-network ${DEFAULT_NAMESPACE} ${AVAILABILITY_ZONE} nw-sdn-controller


################################################################################
# List contents of Vault.
################################################################################

${VAULT} secrets list
${VAULT} policy list
${VAULT} list ${AUTH_PATH_GLOBAL}/role
${VAULT} list ${AUTH_PATH_REGIONAL}/role

################################################################################
# Done.
################################################################################

echo vault-configure-for-nw-cluster.sh: Done.
