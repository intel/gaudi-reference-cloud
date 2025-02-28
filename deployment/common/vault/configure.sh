#!/usr/bin/env bash
# Configure Vault for IDC.
#
# This script does the following:
# - Enable secrets engines.
# - Enable auth methods.
# - Create roles for services.
# - Create policies for services.
# - Configure PKI.
#
# Run with "make deploy-vault-configure".
# This overwrites all configuration except it does not regenerate PKI CA certificates.
# Secrets are NOT loaded by this script. See load-secrets.sh.
#
set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
JQ=${JQ:-jq}
VAULT=${VAULT:-vault}
YQ=${YQ:-yq}
# AUTH_PATH_GLOBAL must match global.vault.authPath in Helmfile environment.
AUTH_PATH_GLOBAL=${AUTH_PATH_GLOBAL:-auth/cluster-auth}
# AUTH_PATH_REGIONAL must match regions[].vault.authPath in Helmfile environment.
AUTH_PATH_REGIONAL=${AUTH_PATH_REGIONAL:-auth/cluster-auth}
BOUND_AUDIENCES_GLOBAL=${BOUND_AUDIENCES_GLOBAL:-https://kubernetes.default.svc.cluster.local}
BOUND_AUDIENCES_REGIONAL=${BOUND_AUDIENCES_REGIONAL:-https://kubernetes.default.svc.cluster.local}
IDC_ENV=${IDC_ENV:-kind}
SECRETS_DIR=${SECRETS_DIR:-local/secrets}
VAULT_BACKUP_DIR=${VAULT_BACKUP_DIR:-${SECRETS_DIR}/vault-backup}
VAULT_BACKUP_DIR=${VAULT_BACKUP_DIR:-${SECRETS_DIR}/vault-backup}
DEFAULT_BAREMETAL_OPERATOR_NAMESPACE=("metal3-1")
BAREMETAL_OPERATOR_NAMESPACES=( "${BAREMETAL_OPERATOR_NAMESPACES[@]:-"${DEFAULT_BAREMETAL_OPERATOR_NAMESPACE[@]}"}" )
DEFAULT_NAMESPACE=idcs-system
# PKI common name limit is 64 characters.
PKI_ROOT_ID="$(uuidgen | head -c 8)"
PKI_ROOT_NAME="Intel IDC CA ${PKI_ROOT_ID} ${IDC_ENV}-root-ca"
PKI_INTERMEDIATE_NAME_PREFIX="Intel IDC CA ${PKI_ROOT_ID} "
# Service certificate expiration will be the lower of PKI_ROLE_TTL and the ttl value in the
# vault.hashicorp.com/agent-inject-template-certkey.pem pod annotation in deployment/charts/idc-common/templates/_common.tpl.
PKI_ROLE_TTL=24h
HELMFILE_DUMP=${HELMFILE_DUMP:-${SECRETS_DIR}/helmfile-dump.yaml}
: "${VAULT_TOKEN:?environment variable is required}"
: "${VAULT_ADDR:?environment variable is required}"
VAULT_DRY_RUN=${VAULT_DRY_RUN:-true}

query_config() {
  ${YQ} "$@" ${HELMFILE_DUMP}
}

get_regions() {
  query_config ".Values.regions[].region"
}

get_availability_zones_for_region() {
  local region=$1
  query_config ".Values.regions.${region}.availabilityZones[].availabilityZone"
}

get_harvester_clusterids_for_region_az() {
  local region=$1
  local availabilityZone=$2
  query_config ".Values.regions.${region}.availabilityZones.${availabilityZone}.harvesterClusters[].clusterId"
}

get_kubevirt_clusterids_for_region_az() {
  local region=$1
  local availabilityZone=$2
  query_config ".Values.regions.${region}.availabilityZones.${availabilityZone}.kubeVirtClusters[].clusterId"
}

wait_for_vault() {
  while ! ${VAULT} secrets list; do sleep 1; done
}

get_existing_policy_content() {
  local policy_name="$1"
  local existing_policy_content=$(${VAULT} policy read "$policy_name" 2>/dev/null)

  if [ "$?" -eq 0 ]; then
    echo "$existing_policy_content"
  else
    echo ""
  fi
}

process_policy_writes() {
  echo "Inside process policy writes for $pol"
  local pol=$1   # Policy name.
  local revised_policy_content="$2"
  local existing_policy_content=$(get_existing_policy_content "$pol")

  # TODO: Update code to check policy content
  if [ -n "$existing_policy_content" ]; then
    echo "Policy $pol with the same name already exists in Vault. Skipping policy write."
  else
    if [ "$VAULT_DRY_RUN" = true ]; then
      echo "vault dry run is enabled. skipping policy $pol."
    else
      echo "creating or updating policy $pol in Vault."
      ${VAULT} policy write ${pol} - <<EOF
$revised_policy_content
EOF
    fi
  fi
}

basic_service_global() {
  local ca=$1-ca	# Should match pki issuer ca
  local namespace=$2
  local svc=$3  	# Should match the Helm release name.
  local pol=${svc}-policy
  local endpoint=$4
  write_service_policy "${ca}" "${namespace}" "" "${svc}" "${pol}" "${endpoint}"
  write_service_roles "${ca}" "${namespace}" "${svc}" "${pol}" "${AUTH_PATH_GLOBAL}" "${BOUND_AUDIENCES_GLOBAL}"
}

basic_service_regional() {
  local ca=$1-ca	# Should match pki issuer ca
  local namespace=$2
  local scope=$3
  local svc=$4  	# Should match the Helm release name.
  local pol=${scope}-${svc}-policy
  local endpoint=$5 # Custom endpoint
  local quick_connect_ca=$6
  write_service_policy "${ca}" "${namespace}" "${scope}" "${svc}" "${pol}" "${endpoint}" "${quick_connect_ca}"
  write_service_roles "${ca}" "${namespace}" "${scope}-${svc}" "${pol}" "${AUTH_PATH_REGIONAL}" "${BOUND_AUDIENCES_REGIONAL}"
}

# Assign policy to access svc name as secrets endpoint
write_service_policy() {
  local ca=$1   # Should match pki issuer ca
  local namespace=$2
  local scope=$3 # region or region+az
  local svc=$4  # Should match the Helm release name.
  local pol=$5  # Policy name.
  local endpoint=$6   # Custom endpoint
  local quick_connect_ca=$7
  local scopedSvc
  if [ -n "${scope}" ]
  then
    scopedSvc=${scope}-${svc}
  else
    scopedSvc=${svc}
  fi
  # controlplane/data/service-name - Global services
  # controlplane/data/$region-service-name & controlplane/data/$AZ-service-name - Legacy (because vault doesn't support giving a policy for users like controlplane/data/*-service-name/*)
  # controlplane/data/$region/service-name - regional services
  # controlplane/data/$region/$az/service-name - az-level services
  local revised_policy_content=$(cat <<EOF
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
)
  local policy_change
  if [ -n "${endpoint}" ]; then
    policy_change=$(cat <<EOF
path "controlplane/data/${endpoint}/*" {
  capabilities = ["read", "list"]
}
EOF
)
  revised_policy_content="${revised_policy_content}
${policy_change}"
  fi

  local quick_connect_policy_change
  if [ -n "${quick_connect_ca}" ]; then
    quick_connect_policy_change=$(cat <<EOF
path "${quick_connect_ca}/cert/ca_chain" {
  capabilities = ["read"]
}
EOF
)
  revised_policy_content="${revised_policy_content}
${quick_connect_policy_change}"
  fi

  process_policy_writes "${pol}" "${revised_policy_content}"
}


product_catalog_service_global() {
  local ca=$1-ca	# Should match pki issuer ca
  local namespace=$2
  local svc=$3  	# Should match the Helm release name.
  local pol=${svc}-policy
  local endpoint=$4
  local endpoint2=$5
  write_product_catalog_service_policy "${ca}" "${namespace}" "" "${svc}" "${pol}" "${endpoint}" "${endpoint2}"
  write_service_roles "${ca}" "${namespace}" "${svc}" "${pol}" "${AUTH_PATH_GLOBAL}" "${BOUND_AUDIENCES_GLOBAL}"
}

# Assign policy to access svc name as secrets endpoint
write_product_catalog_service_policy() {
  local ca=$1   # Should match pki issuer ca
  local namespace=$2
  local scope=$3 # region or region+az
  local svc=$4  # Should match the Helm release name.
  local pol=$5  # Policy name.
  local endpoint=$6   # Custom endpoint
  local endpoint2=$7 # Custom endpoint
  local scopedSvc
  if [ -n "${scope}" ]
  then
    scopedSvc=${scope}-${svc}
  else
    scopedSvc=${svc}
  fi
  # controlplane/data/service-name - Global services
  # controlplane/data/$region-service-name & controlplane/data/$AZ-service-name - Legacy (because vault doesn't support giving a policy for users like controlplane/data/*-service-name/*)
  # controlplane/data/$region/service-name - regional services
  # controlplane/data/$region/$az/service-name - az-level services
  local revised_policy_content=$(cat <<EOF
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
)
  local policy_change
  if [ -n "${endpoint}" ]; then
    policy_change=$(cat <<EOF
path "controlplane/data/${endpoint}/*" {
  capabilities = ["read", "list"]
}
EOF
)
  revised_policy_content="${revised_policy_content}
${policy_change}"
  fi

  if [ -n "${endpoint2}" ]; then
    policy_change=$(cat <<EOF
path "controlplane/data/${endpoint2}/*" {
  capabilities = ["read", "list"]
}
EOF
)
    revised_policy_content="${revised_policy_content}
${policy_change}"
  fi

  process_policy_writes "${pol}" "${revised_policy_content}"
}


write_service_roles() {
  local ca=$1	  # Should match pki issuer ca
  local namespace=$2
  local svc=$3  # Should match the Helm release name.
  local pol=$4  # List of service-specific Vault policies.
  local auth_path=$5
  local bound_audiences=$6
  local policies="public,global-pki"
  if [ -n "$pol" ]; then
    policies=${policies},$pol
  fi
  local role_path="${auth_path}/role/${svc}-role"
  local role_attributes=(
    role_type="jwt"
    bound_audiences="${bound_audiences}"
    user_claim="sub"
    bound_subject="system:serviceaccount:${namespace}:${svc}"
    policies="${policies}"
    ttl="1h"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  role_path="${ca}/roles/${svc}"
  role_attributes=(
      allowed_domains="${svc}.${DEFAULT_NAMESPACE}.svc.cluster.local,*.local,*.internal-placeholder.com,*.eglb.intel.com,*.internal-placeholder.com,*.internal-placeholder.com,*.cloud.intel.com"
      allow_glob_domains=true
      allow_bare_domains=true
      allow_wildcard_certificates=false
      ou="${svc}"
      ttl="${PKI_ROLE_TTL}"
      max_ttl="${PKI_ROLE_TTL}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"
}

process_roles_writes_storage() {
    local role_path="$1" # Includes the path to the role
    shift # Shift the role_path argument out, leaving only the attributes map
    local attributes=("$@") # Store the remaining arguments as an array
    local attribute_list=()
    for attribute in "${attributes[@]}"; do
        attribute_list+=("$attribute")
    done

      # Use vault write with attribute key-value pairs
      if [ "$VAULT_DRY_RUN" = true ]; then
        echo "vault dry run is enabled. skipping creation for role $role_path with attributes ${attribute_list[@]}."
      else
        ${VAULT} write $role_path "${attribute_list[@]}"
        echo "Role $role_path created with attributes ${attribute_list[@]}."
      fi
}

process_roles_writes() {
    local role_path="$1" # Includes the path to the role
    shift # Shift the role_path argument out, leaving only the attributes map
    local attributes=("$@") # Store the remaining arguments as an array

    # Check if the role exists
    if ${VAULT} read $role_path &> /dev/null; then
      echo "Role $role_path already exists in Vault. Skipping role write."
    else
      echo "Role $role_path does not exist. Need to create.."
      local attribute_list=()
      for attribute in "${attributes[@]}"; do
          attribute_list+=("$attribute")
      done

      # Use vault write with attribute key-value pairs
      if [ "$VAULT_DRY_RUN" = true ]; then
        echo "vault dry run is enabled. skipping creation for role $role_path with attributes ${attribute_list[@]}."
      else
        ${VAULT} write $role_path "${attribute_list[@]}"
        echo "Role $role_path created with attributes ${attribute_list[@]}."
      fi
    fi
}

enable_secrets_global() {
  if [ "$VAULT_DRY_RUN" = true ]; then
      echo "vault dry run is enabled. skipping enable_secrets_global."
  else
      ${VAULT} secrets enable -path=controlplane -version=2 kv || true
      ${VAULT} secrets enable -path=public -version=2 kv || true
      ${VAULT} secrets enable -path=bmc -version=2 kv || true
      ${VAULT} secrets enable -path=storage -version=2 kv || true
  fi
}

enable_auth_methods() {
  if [ "$VAULT_DRY_RUN" = true ]; then
    echo "vault dry run is enabled. skipping enable_auth_methods."
  else
    ${VAULT} auth enable approle || true
    ${VAULT} auth enable -path=cluster-auth jwt || true
  fi
}

setup_intermediate_ca() {
  local ca=$1-ca  # Should match pki issuer ca
  local ca_path=${SECRETS_DIR}/pki/${ca}
  local max_lease_ttl=${2:-8760h}
  local ttl=${3:-8760h}
  local issuer_ref=${4:-"idc-${IDC_ENV}-root-ca"}
  local pki_intermediate_name_prefix=${5:-"${PKI_INTERMEDIATE_NAME_PREFIX}"}
  local root_ca_mount_path=${6:-"${IDC_ENV}-root-ca"}
  mkdir -p ${ca_path}

  if ${VAULT} read ${ca}/config/urls; then
    echo "CA ${ca} already exists. Skipping existing CA ${ca}"
  else
    if [ "$VAULT_DRY_RUN" = true ]; then
      echo "vault dry run is enabled. skipping CA ${ca}"
    else
      ${VAULT} secrets enable -path=${ca} -max-lease-ttl=${max_lease_ttl} pki

      # generate intermediate ca cert signing request
      ${VAULT} write -format=json ${ca}/intermediate/generate/internal \
        common_name="${pki_intermediate_name_prefix}${ca}" \
        issuer_name="idc-${ca}" \
        ttl=${ttl} | ${JQ} -r '.data.csr' > ${ca_path}/${ca}.csr

      # sign intermediate ca cert using the root-ca
      ${VAULT} write -format=json ${root_ca_mount_path}/root/sign-intermediate \
        issuer_ref="${issuer_ref}" \
        csr=@${ca_path}/${ca}.csr \
        format=pem_bundle ttl=${ttl} \
        | ${JQ} -r '.data.certificate' > ${ca_path}/${ca}.pem

      # set the ca cert signed by the root-ca
      ${VAULT} write ${ca}/intermediate/set-signed \
        certificate=@${ca_path}/${ca}.pem

      # configure urls for the issuer
      ${VAULT} write ${ca}/config/urls \
        issuing_certificates="${VAULT_ADDR}/v1/${ca}/ca" \
        crl_distribution_points="${VAULT_ADDR}/v1/${ca}/crl"
    fi
  fi

  if [ "$VAULT_DRY_RUN" = true ]; then
    echo "vault dry run is enabled. skipping saving ca-chain for the issuer"
  else
	# save ca-chain for the issuer
	  ${VAULT} read -format=json ${ca}/cert/ca_chain \
		  | ${JQ} -r '.data.ca_chain' > ${ca_path}/ca-chain.pem
  fi
}

# A real wildcard certificate such as used in the staging and production
# environments won't work for the kind environment since the ingress hosts
# do not share a single sub-domain. Instead, create a certificate with SANs
# of all ingress hosts.
#
# This cert will be signed by the environment global CA which is itself
# signed by the environment root CA. The root CA cert chain is available as
# part of the default vault agent annotations in each service
# (/vault/secrets/ca.pem in each Pod).
setup_wildcard_tls() {
  local cert=wildcard-tls
  local cert_path=${SECRETS_DIR}/${cert}
  local global_ca_mount_path="${IDC_ENV}-global-ca"
  local role="wildcard-tls"
  mkdir -p ${cert_path}

  role_path="${global_ca_mount_path}/roles/${role}"
  role_attributes=(
      allowed_domains="*.cloud.intel.com.kind.local"
      allow_glob_domains=true
      allow_bare_domains=true
      allow_wildcard_certificates=true
      ou="${role}"
      ttl="${PKI_ROLE_TTL}"
      max_ttl="${PKI_ROLE_TTL}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  [[ -f ${cert_path}/${cert} ]] && serial=$(${JQ} -r '.data.serial_number' ${cert_path}/${cert} 2>/dev/null) || serial=""
  if ${VAULT} read ${global_ca_mount_path}/cert/${serial}; then
    echo "certificate ${cert} already exists. skipping existing certificate ${cert}"
  else
    if [ "$VAULT_DRY_RUN" = true ]; then
      echo "vault dry run is enabled. skipping certificate ${cert}"
    else
      # issue cert
      ${VAULT} write -format=json ${global_ca_mount_path}/issue/${role} \
        common_name="*.cloud.intel.com.kind.local" \
        alt_names="dev.argocd.cloud.intel.com.kind.local, \
          dev.gitea.cloud.intel.com.kind.local, \
          dev.grpcapi.cloud.intel.com.kind.local, \
          dev.api.cloud.intel.com.kind.local, \
          dev-sdk.api.cloud.intel.com.kind.local, \
          dev.oidc.cloud.intel.com.kind.local, \
          dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local, \
          dev.compute.us-dev-1.api.cloud.intel.com.kind.local, \
          *.quick-connect.us-dev-1a.cloud.intel.com.kind.local, \
          dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local, \
          dev.vault.cloud.intel.com.kind.local" > ${cert_path}/${cert}
    fi

    # save ca-chain and key for the secret
    ${JQ} -r '.data.certificate,.data.ca_chain[]' ${cert_path}/${cert} > ${cert_path}/tls.crt
    ${JQ} -r '.data.private_key' ${cert_path}/${cert} > ${cert_path}/tls.key
  fi
}

configure_pki_global() {
  local ca=${IDC_ENV}-root-ca
  if ${VAULT} read ${ca}/config/urls; then
    echo "CA ${ca} already exists. Skipping existing CA ${ca}"
  else
    if [ "$VAULT_DRY_RUN" = true ]; then
      echo "vault dry run is enabled. skipping creation of ${ca}"
    else
      # self-signed root-ca setup
      mkdir -p ${SECRETS_DIR}/pki/${ca}

      ${VAULT} secrets enable -path=${ca} -max-lease-ttl=8760h pki

      # create a default root cert with ttl=1y
      # if root private key required, use /exported endpoint instead
      ${VAULT} write -field=certificate ${ca}/root/generate/internal \
          common_name="${PKI_ROOT_NAME}" \
          issuer_name="idc-${ca}" \
          ttl=8760h | openssl x509 -text > ${SECRETS_DIR}/pki/${ca}/ca.pem

      # configure urls for the issuer
      ${VAULT} write ${ca}/config/urls \
          issuing_certificates="${VAULT_ADDR}/v1/${ca}/ca" \
          crl_distribution_points="${VAULT_ADDR}/v1/${ca}/crl"

      # set revocation list update expiry=10s
      ${VAULT} write ${ca}/config/crl \
          expiry=10s
    fi
  fi

  # generate intermediate certificate, csr and sign with root-ca for each cluster
  # single Vault cluster to be used for multiple dev environments, hence root-ca and global-ca need env prefix
  setup_intermediate_ca ${IDC_ENV}-global
}

configure_global_services() {
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} authz
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} billing cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} billing-intel cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} billing-schedulers cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} billing-aria cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} billing-standard cloudcredits
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} notification-gateway
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} cloudaccount
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} cloudaccount-enroll
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} console
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} grpc-proxy-external gts-trade-compliance
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} grpc-proxy-internal gts-trade-compliance
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} app-client-api-grpc-proxy-external gts-trade-compliance
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} app-client-api-grpc-rest-gateway
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} grpc-reflect
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} grpc-rest-gateway
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} grpc-internal-rest-gateway
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} usage
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} metering
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} cloudcredits-worker
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} trade-scanner gts-trade-compliance
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} populate-inflow-component-git-to-grpc-synchronizer
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} populate-inflow-os-git-to-grpc-synchronizer
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} populate-inflow-recipe-git-to-grpc-synchronizer
  basic_service_global ${IDC_ENV}-global idcs-observability opentelemetry-collector
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} cloudmonitor
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} productcatalog-operator gts-trade-compliance
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} populate-product-git-to-grpc-synchronizer
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} rate-limit rate-limit-redis
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} rate-limit-redis
  basic_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} user-credentials cloudaccount

  basic_service_global ${IDC_ENV}-global cattle-monitoring-system rancher-monitoring-prometheus
  product_catalog_service_global ${IDC_ENV}-global ${DEFAULT_NAMESPACE} productcatalog cloudaccount gts-trade-compliance
}

# Public path (public certificates, etc.)
configure_public() {
  local revised_policy_content=$(cat <<EOF
path "public/*" {
  capabilities = ["read", "list"]
}
EOF
  )
  process_policy_writes "public" "${revised_policy_content}"
}

configure_global() {
  wait_for_vault
  enable_secrets_global
  enable_auth_methods
  configure_pki_global
  configure_global_services
  configure_public
}

list_vault() {
  ${VAULT} secrets list
  ${VAULT} policy list
  ${VAULT} list ${AUTH_PATH_GLOBAL}/role
  ${VAULT} list ${AUTH_PATH_REGIONAL}/role
}

#################################################
# Baremetal enrollment task policy and roles
#################################################
configure_baremetal_enrollment_task() {
  local region=$1
  local availability_zone=$2
  local revised_policy_content=$(cat << EOF
path "${availability_zone}-ca/issue/${availability_zone}-baremetal-enrollment-task" {
  capabilities = ["update", "create"]
}
EOF
)
  process_policy_writes "${availability_zone}-baremetal-enrollment-task" "${revised_policy_content}"

  # Create Kubernetes cluster-auth role for enrollment jobs created by baremetal-enrollment.
  local role_path="${AUTH_PATH_REGIONAL}/role/${availability_zone}-baremetal-enrollment-task-role"
  local role_attributes=(
    role_type="jwt"
    bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
    user_claim="sub"
    bound_subject="system:serviceaccount:idcs-enrollment:${availability_zone}-baremetal-enrollment-task"
    ttl="1h"
    policies="public,${region}-baremetal-enrollment,${availability_zone}-baremetal-enrollment-task"
  )

  process_roles_writes "$role_path" "${role_attributes[@]}"

  local role_path="${availability_zone}-ca/roles/${availability_zone}-baremetal-enrollment-task"
  local role_attributes=(
    allow_localhost=true
    allow_any_name=true
    enforce_hostnames=false
    allow_ip_sans=true
    ou="${availability_zone}-baremetal-enrollment-task"
    ttl="2190h"
    max_ttl="2190h"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"
}

#################################################
# Baremetal enrollment operator
#################################################
#ToDo: move this as part of basic_service_regional(need to include ${region}-baremetal-enrollment policy)

configure_baremetal_enrollment_operator() {
  local region=$1
  local availability_zone=$2
  local revised_policy_content=$(cat << EOF
path "${availability_zone}-ca/issue/${availability_zone}-baremetal-enrollment-operator" {
  capabilities = ["update", "create"]
}
EOF
)
  process_policy_writes "${availability_zone}-baremetal-enrollment-operator" "${revised_policy_content}"

  local role_path="${AUTH_PATH_REGIONAL}/role/${availability_zone}-baremetal-enrollment-operator-role"
  local role_attributes=(
    role_type="jwt"
    bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
    user_claim="sub"
    bound_subject="system:serviceaccount:idcs-system:${availability_zone}-baremetal-enrollment-operator"
    ttl="1h"
    policies="public,${region}-baremetal-enrollment,${availability_zone}-baremetal-enrollment-operator"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  local role_path="${availability_zone}-ca/roles/${availability_zone}-baremetal-enrollment-operator"
  local role_attributes=(
    allow_localhost=true
    allow_any_name=true
    enforce_hostnames=false
    allow_ip_sans=true
    ou="${availability_zone}-baremetal-enrollment-operator"
    ttl="2190h"
    max_ttl="2190h"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"
}

#################################################
# Baremetal enrollment API policy and roles
#################################################
configure_baremetal_enrollment_api() {
  local region=$1
  local revised_policy_content=$(cat <<EOF
path "bmc/metadata/${region}/deployed/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "bmc/data/${region}/deployed/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "controlplane/metadata/${region}/baremetal/enrollment/*" {
  capabilities = ["read"]
}
path "controlplane/data/${region}/baremetal/enrollment/*" {
  capabilities = ["read"]
}
path "${region}-ca/issue/${region}-baremetal-enrollment-api" {
  capabilities = ["update", "create"]
}
EOF
)
  process_policy_writes "${region}-baremetal-enrollment" "${revised_policy_content}"

  local role_path="auth/approle/role/${region}-baremetal-enrollment-role"
  local role_attributes=(
    secret_id_num_uses=0
    secret_id_ttl=0
    token_ttl=20h
    token_max_ttl=24h
    token_num_uses=0
    policies="${region}-baremetal-enrollment"
  )

  process_roles_writes "$role_path" "${role_attributes[@]}"

  role_path="${AUTH_PATH_REGIONAL}/role/${region}-baremetal-enrollment-api-role"
  role_attributes=(
    role_type="jwt"
    bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
    user_claim="sub"
    bound_subject="system:serviceaccount:idcs-enrollment:${region}-baremetal-enrollment-api"
    policies="public,${region}-baremetal-enrollment"
    ttl="1h"
  )

  process_roles_writes "$role_path" "${role_attributes[@]}"

  role_path="${AUTH_PATH_REGIONAL}/role/${region}-enrollment-role"
  role_attributes=(
    role_type="jwt"
    bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
    user_claim="sub"
    bound_subject="system:serviceaccount:idcs-enrollment:enrollment"
    policies="public,${region}-baremetal-enrollment"
    ttl="1h"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  role_path="${region}-ca/roles/${region}-baremetal-enrollment-api"
  role_attributes=(
    allow_localhost=true
    allow_any_name=true
    enforce_hostnames=false
    allow_ip_sans=true
    ou="${region}-baremetal-enrollment-api"
    ttl="2190h"
    max_ttl="2190h"
  )

  process_roles_writes "$role_path" "${role_attributes[@]}"
}

#################################################
# Baremetal instance operator API policy and roles
# (For Ironic and BMO mTLS)
#################################################
configure_baremetal_operator() {
  local availability_zone=$1
  local az_ca=${availability_zone}-ca  # Should match cert issuer ca

  local revised_policy_content=$(cat <<EOF
path "${az_ca}/issue/${availability_zone}-baremetal-operator" {
  capabilities = ["update", "create"]
}
path "controlplane/data/${availability_zone}-baremetal-operator/*" {
  capabilities = ["read", "list"]
}
EOF
)
  process_policy_writes "${availability_zone}-baremetal-operator-policy" "${revised_policy_content}"
  local role_path
  local role_attributes
  for namespace in ${BAREMETAL_OPERATOR_NAMESPACES[@]}
    do
      role_path="${AUTH_PATH_REGIONAL}/role/${availability_zone}-baremetal-operator-${namespace}-ironic-role"
      role_attributes=(
        role_type="jwt"
        bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
        user_claim="sub"
        bound_subject="system:serviceaccount:${namespace}:baremetal-operator-ironic"
        policies="public,${availability_zone}-baremetal-operator-policy"
        ttl="24h"
      )
      process_roles_writes "$role_path" "${role_attributes[@]}"

      role_path="${AUTH_PATH_REGIONAL}/role/${availability_zone}-baremetal-operator-${namespace}-role"
      role_attributes=(
        role_type="jwt"
        bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
        user_claim="sub"
        bound_subject="system:serviceaccount:${namespace}:baremetal-operator"
        policies="public,${availability_zone}-baremetal-operator-policy"
        ttl="24h"
      )
      process_roles_writes "$role_path" "${role_attributes[@]}"
    done
      role_path="${az_ca}/roles/${availability_zone}-baremetal-operator"
      role_attributes=(
        allow_localhost=true
        allow_any_name=true
        enforce_hostnames=false
        allow_ip_sans=true
        ou="${availability_zone}-baremetal-operator"
        ttl="2190h"
        max_ttl="2190h"
      )
      process_roles_writes "$role_path" "${role_attributes[@]}"
}

configure_storage_kms() {
  local region=$1
  local revised_policy_content=$(cat <<EOF
path "controlplane/data/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "controlplane/metadata/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "storage/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "controlplane/metadata/${region}/storage/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "${region}-ca/issue/${region}-storage-kms" {
capabilities = ["update", "create"]
}
EOF
)
  process_policy_writes "${region}-storage-kms" "${revised_policy_content}"

  # Create an approle for storage.
  local role_path="auth/approle/role/${region}-storage-kms-role"
  local role_attributes=(
    secret_id_num_uses=0
    secret_id_ttl=0
    token_ttl=20h
    token_max_ttl=24h
    token_num_uses=0
    policies=${region}-storage-kms
  )
  process_roles_writes_storage "$role_path" "${role_attributes[@]}"

  # Create Kubernetes cluster-auth role for storage-api.
  local role_path="${AUTH_PATH_REGIONAL}/role/${region}-storage-kms-role"
  local role_attributes=(
    role_type="jwt"
    bound_audiences="${BOUND_AUDIENCES_REGIONAL}"
    user_claim="sub"
    bound_subject="system:serviceaccount:idcs-system:${region}-storage-kms"
    policies="public,${region}-storage-kms,global-pki"
    ttl="1h"
  )
  process_roles_writes_storage "$role_path" "${role_attributes[@]}"

  local role_path="${region}-ca/roles/${region}-storage-kms-role"
  local role_attributes=(
    allow_localhost=true
    allow_any_name=true
    enforce_hostnames=false
    allow_ip_sans=true
    ou="${region}-storage-kms"
    ttl="2190h"
    max_ttl="2190h"
  )
  process_roles_writes_storage "$role_path" "${role_attributes[@]}"

################################################################################
# Storage roles/ Policy End
################################################################################
}


configure_vm_clusters() {
  local region=$1
  local availability_zone=$2
  harvesters_clusterids=$(get_harvester_clusterids_for_region_az ${region} ${availability_zone})
  kubevirt_clusterids=$(get_kubevirt_clusterids_for_region_az ${region} ${availability_zone})

  for clusterid in $harvesters_clusterids $kubevirt_clusterids; do
      basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} vm-instance-operator-${clusterid} "" ${availability_zone}-quick-connect-client-ca
      basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} k8s-resource-patcher-${clusterid}
  done
}

configure_quick_connect_api_server() {
  local availability_zone=$1
  local ca=${availability_zone}-ca	# Should match pki issuer ca
  local clientCA=${availability_zone}-quick-connect-client-ca
  local namespace=${DEFAULT_NAMESPACE}
  local scope=${availability_zone}
  local svc=quick-connect-api-server  	# Should match the Helm release name.
  local pol=${scope}-${svc}-policy
  local scopedSvc=${scope}-${svc}
  local client=${availability_zone}-quick-connect-client  # Required to match terraform data.

  local revised_policy_content=$(cat <<EOF
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
path "${clientCA}/issue/${client}" {
  capabilities = ["update"]
}
EOF
)
  process_policy_writes "${pol}" "${revised_policy_content}"

  write_service_roles "${ca}" "${namespace}" "${scopedSvc}" "${pol}" "${AUTH_PATH_REGIONAL}" "${BOUND_AUDIENCES_REGIONAL}"

  role_path="${clientCA}/roles/${client}"
  role_attributes=(
      allowed_domains="${client}.${DEFAULT_NAMESPACE}.svc.cluster.local"
      allow_bare_domains=true
      allow_wildcard_certificates=false
      key_type="rsa"
      key_bits=3072
      signature_bits=384
      ou="${scopedSvc}"
      ttl="${PKI_ROLE_TTL}"
      max_ttl="${PKI_ROLE_TTL}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  local role_path="auth/approle/role/${scopedSvc}-role"
  local role_attributes=(
    secret_id_num_uses=0
    secret_id_ttl=0
    token_ttl=20h
    token_max_ttl=24h
    token_num_uses=0
    policies="${pol}"
  )

  process_roles_writes "$role_path" "${role_attributes[@]}"
}

configure_kubevirt() {
  local region=$1
  local availability_zone=$2

  local worker_policy=${availability_zone}-vm-worker-policy
  local worker_policy_content=$(cat <<EOF
path "controlplane/data/${region}/${availability_zone}/rancher" {
  capabilities = ["read", "list"]
}
EOF
)
  process_policy_writes "${worker_policy}" "${worker_policy_content}"

  local role_path="auth/approle/role/${availability_zone}-vm-worker-role"
  local role_attributes=(
    secret_id_num_uses=0
    secret_id_ttl=0
    token_ttl=20h
    token_max_ttl=24h
    token_num_uses=0
    policies="${worker_policy}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  local operator_policy=${availability_zone}-vm-cluster-operator-policy
  local operator_policy_content=$(cat <<EOF
path "auth/approle/role/${availability_zone}-vm-worker-role/role-id" {
  capabilities = ["read", "list"]
}
path "auth/approle/role/${availability_zone}-vm-worker-role/secret-id" {
    capabilities = ["create", "update"]
    min_wrapping_ttl = "1m"
    max_wrapping_ttl = "30m"
}
EOF
)
  process_policy_writes "${operator_policy}" "${operator_policy_content}"

  local role_path="auth/approle/role/${availability_zone}-vm-cluster-operator-role"
  local role_attributes=(
    secret_id_num_uses=0
    secret_id_ttl=0
    token_ttl=20h
    token_max_ttl=24h
    token_num_uses=0
    policies="${operator_policy}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"
}

configure_region_services() {
  local region=$1
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} grpc-proxy-external gts-trade-compliance
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} grpc-proxy-internal gts-trade-compliance
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} app-client-api-grpc-proxy-external gts-trade-compliance
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} app-client-api-grpc-rest-gateway
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} grpc-reflect
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} grpc-rest-gateway
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} cloudmonitor-logs-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} compute-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} fleet-admin-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} fleet-admin-ui-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-admin-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-scheduler
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-user
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-kms
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-resource-cleaner
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} quota-management-service
  configure_storage_kms ${region}
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} training-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} armada ${region}-training-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} compute-nginx-s3-gateway
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} iks
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} security-insights
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} kubescore
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} security-scanner
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} kfaas
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} infaas-inference
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} infaas-dispatcher
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} maas-gateway
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} dpai
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} dataloader
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} populate-instance-type-git-to-grpc-synchronizer
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} populate-machine-image-git-to-grpc-synchronizer
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} populate-subnet-git-to-grpc-synchronizer
  basic_service_regional ${region} idcs-observability ${region} opentelemetry-collector
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} rate-limit ${region}-rate-limit-redis
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} rate-limit-redis
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} storage-custom-metrics-service
  basic_service_regional ${region}-network ${DEFAULT_NAMESPACE} ${region} provider-sdn-controller
  basic_service_regional ${region}-network ${DEFAULT_NAMESPACE} ${region} provider-sdn-controller-rest ${region}/provider-sdn-controller
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} network-api-server
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} network-operator
  basic_service_regional ${region} ${DEFAULT_NAMESPACE} ${region} sdn-vn-controller
  configure_baremetal_enrollment_api ${region}
}

configure_az_services() {
  local region=$1
  local availability_zone=$2
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} bm-instance-operator "" ${availability_zone}-quick-connect-client-ca
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} bm-validation-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} compute-metering-monitor
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} fleet-node-reporter
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} k8s-resource-patcher-bm
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} storage-metering-monitor
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} bucket-metering-monitor
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} storage-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} vast-storage-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} vast-metering-monitor
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} object-store-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} storage-replicator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} bucket-replicator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} instance-replicator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} ssh-proxy-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} vm-instance-scheduler
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} ilb-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} kubernetes-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} kubernetes-reconciler
  basic_service_regional ${availability_zone} idcs-observability   ${availability_zone} opentelemetry-collector
  basic_service_regional ${availability_zone}-network ${DEFAULT_NAMESPACE} ${availability_zone} nw-sdn-controller
  basic_service_regional ${availability_zone}-network ${DEFAULT_NAMESPACE} ${availability_zone} nw-sdn-controller-rest +/${availability_zone}/nw-sdn-controller
  basic_service_regional ${availability_zone}-network ${DEFAULT_NAMESPACE} ${availability_zone} nw-sdn-integrity-checker +/${availability_zone}/nw-sdn-controller
  basic_service_regional ${availability_zone}-network ${DEFAULT_NAMESPACE} ${availability_zone} nw-switch-config-saver +/${availability_zone}/nw-sdn-controller
  configure_baremetal_operator ${availability_zone}
  configure_vm_clusters ${region} ${availability_zone}
  configure_baremetal_enrollment_task ${region} ${availability_zone}
  configure_baremetal_enrollment_operator ${region} ${availability_zone}
  configure_sdn ${region} ${availability_zone}
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} firewall-operator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} loadbalancer-replicator
  basic_service_regional ${availability_zone} ${DEFAULT_NAMESPACE} ${availability_zone} loadbalancer-operator
  configure_quick_connect_api_server ${availability_zone}
  configure_kubevirt ${region} ${availability_zone}
}

configure_sdn() {
  local region=$1

  role_path="${region}-network-ca/roles/${region}-provider-sdn-readonly"
  role_attributes=(
      allowed_domains="*.idcs-system.svc.cluster.local,*.internal-placeholder.com,*.eglb.intel.com,*.internal-placeholder.com,*.internal-placeholder.com"
      allow_glob_domains=true
      allow_bare_domains=true
      allow_wildcard_certificates=false
      server_flag=false
      client_flag=true
      ou="provider-sdn-readonly"
      ttl="${PKI_ROLE_TTL}"
      max_ttl="${PKI_ROLE_TTL}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"

  role_path="${region}-network-ca/roles/${region}-provider-sdn-readwrite"
  role_attributes=(
      allowed_domains="*.idcs-system.svc.cluster.local,*.internal-placeholder.com,*.eglb.intel.com,*.internal-placeholder.com,*.internal-placeholder.com"
      allow_glob_domains=true
      allow_bare_domains=true
      allow_wildcard_certificates=false
      server_flag=false
      client_flag=true
      ou="provider-sdn-readwrite"
      ttl="${PKI_ROLE_TTL}"
      max_ttl="${PKI_ROLE_TTL}"
  )
  process_roles_writes "$role_path" "${role_attributes[@]}"
}

configure_region() {
  local region=$1
  setup_intermediate_ca ${region}
  setup_intermediate_ca ${region}-network
  configure_region_services ${region}
}

configure_az() {
  local region=$1
  local availability_zone=$2
  setup_intermediate_ca ${availability_zone}
  setup_intermediate_ca ${availability_zone}-network
  setup_intermediate_ca ${availability_zone}-quick-connect-client 87600h 87600h
  configure_az_services ${region} ${availability_zone}
}

# TODO: Update function to include all Roles
backup_vault() {
  # Create a backup directory with the current date as a postfix
  CURRENT_DATE=$(date +'%Y%m%d_%H%M%S')
  BACKUP_DIR_POLICY="$VAULT_BACKUP_DIR/$CURRENT_DATE/policy"
  BACKUP_DIR_ROLES_GLOBAL="$VAULT_BACKUP_DIR/$CURRENT_DATE/roles/${AUTH_PATH_GLOBAL}"
  BACKUP_DIR_ROLES_REGIONAL="$VAULT_BACKUP_DIR/$CURRENT_DATE/roles/${AUTH_PATH_REGIONAL}"

  mkdir -p "$BACKUP_DIR_POLICY"
  mkdir -p "$BACKUP_DIR_ROLES_GLOBAL"
  mkdir -p "$BACKUP_DIR_ROLES_REGIONAL"

  # Backup Vault policies
  ${VAULT} policy list | while read -r policy_name; do
    if ${VAULT} policy read "$policy_name" > "$BACKUP_DIR_POLICY/$policy_name.hcl" 2>&1; then
      echo "Policy $policy_name backed up at $BACKUP_DIR_POLICY successfully"
    else
      echo "Failed to back up policy $policy_name"
    fi
  done

  # Backup Vault roles
    ${VAULT} list ${AUTH_PATH_GLOBAL}/role | while read -r role_name; do
      if ${VAULT} read ${AUTH_PATH_GLOBAL}/role/"$role_name" > "$BACKUP_DIR_ROLES_GLOBAL/${role_name}.json" 2>&1; then
        echo "Role ${role_name} backed at $BACKUP_DIR_ROLES_GLOBAL successfully"
      else
        echo "Failed to back up role ${AUTH_PATH_GLOBAL}/${role_name}"
      fi
    done

      ${VAULT} list ${AUTH_PATH_REGIONAL}/role | while read -r role_name; do
      if ${VAULT} read ${AUTH_PATH_REGIONAL}/role/"$role_name" > "$BACKUP_DIR_ROLES_REGIONAL/${role_name}.json" 2>&1; then
        echo "Role ${AUTH_PATH_REGIONAL}/${role_name} backed at $BACKUP_DIR_ROLES_REGIONAL successfully"
      else
        echo "Failed to back up role ${AUTH_PATH_REGIONAL}/${role_name}"
      fi
    done

  echo "Vault policies, and roles have been backed successfully"
}

main() {
  if [ "$VAULT_DRY_RUN" = true ]; then
    echo "vault dry run is enabled. skipping backup."
  else
    echo "executing backup_vault."
    backup_vault
    echo "backup_vault done."
  fi

  configure_global

  local regions=$(get_regions)
  local availability_zones
  for region in ${regions}; do
    configure_region ${region}
    availability_zones=$(get_availability_zones_for_region ${region})
    for availability_zone in $availability_zones; do
      configure_az ${region} ${availability_zone}
    done
  done

  setup_wildcard_tls

  list_vault
  echo configure.sh: Done.
}

main
