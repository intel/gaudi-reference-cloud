#!/usr/bin/env bash
# Load secrets into Vault.
# This script reads secrets from files in ${SECRETS_DIR} and writes them to Vault.
# Run with "make deploy-vault-secrets".
# By default, this does not overwrite existing secrets.
# To overwrite secrets, run "OVERWRITE_SECRETS=1 make deploy-vault-secrets".

set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
: "${VAULT_TOKEN:?environment variable is required}"
: "${VAULT_ADDR:?environment variable is required}"
JQ=${JQ:-jq}
VAULT=${VAULT:-vault}
YQ=${YQ:-yq}
AVAILABILITY_ZONE0=${AVAILABILITY_ZONE:-us-dev-1a}
SECRETS_DIR=${SECRETS_DIR:-local/secrets}
JWT_VALIDATION_PUBKEYS_FILE=${JWT_VALIDATION_PUBKEYS_FILE:-${SECRETS_DIR}/jwt_validation_pubkeys}
DEFAULT_BMC_USERNAME=${DEFAULT_BMC_USERNAME}
DEFAULT_BMC_PASSWD=${DEFAULT_BMC_PASSWD}
ENROLLMENT_KUBECONFIG=${ENROLLMENT_KUBECONFIG:-${SECRETS_DIR}/kubeconfig/kind-idc-${AVAILABILITY_ZONE0}.yaml}
SSH_PROXY_OPERATOR_ID_RSA=${SSH_PROXY_OPERATOR_ID_RSA:-${SECRETS_DIR}/ssh-proxy-operator/id_rsa}
SSH_PROXY_SERVER_HOST_PUBLIC_KEY=${SSH_PROXY_SERVER_HOST_PUBLIC_KEY:-${SECRETS_DIR}/ssh-proxy-operator/host_public_key}
BM_INSTANCE_OPERATOR_ID_RSA=${BM_INSTANCE_OPERATOR_ID_RSA:-${SECRETS_DIR}/bm-instance-operator/id_rsa}
BM_VALIDATION_OPERATOR_ID_RSA=${BM_VALIDATION_OPERATOR_ID_RSA:-${SECRETS_DIR}/bm-validation-operator/id_rsa}
RAVEN_USERNAME=$(cat ${SECRETS_DIR}/RAVEN_USERNAME)
RAVEN_PASSWD=$(cat ${SECRETS_DIR}/RAVEN_PASSWD)
# ICP_USERNAME and ICP_PASSWORD correspond to the API Key and Secret that can be obtained from API portal
# https://internal-placeholder.com. Currently we are using https://internal-placeholder.com/my-apps/661c06a7-90c7-414b-8c63-e289a79fdffd
# as the reference app.
ICP_USERNAME=$(cat ${SECRETS_DIR}/ICP_USERNAME)
ICP_PASSWORD=$(cat ${SECRETS_DIR}/ICP_PASSWORD)
GTS_USERNAME=$(cat ${SECRETS_DIR}/GTS_USERNAME)
GTS_PASSWORD=$(cat ${SECRETS_DIR}/GTS_PASSWORD)
MEN_AND_MICE_USERNAME=$(cat ${SECRETS_DIR}/MEN_AND_MICE_USERNAME)
MEN_AND_MICE_PASSWORD=$(cat ${SECRETS_DIR}/MEN_AND_MICE_PASSWORD)
NETBOX_TOKEN=$(cat ${SECRETS_DIR}/NETBOX_TOKEN)
GRAFANA_USERNAME=$(cat ${SECRETS_DIR}/grafana_admin_username)
GRAFANA_PASSWORD=$(cat ${SECRETS_DIR}/grafana_admin_password)
BM_ENROLLMENT_APISERVICE_USERNAME=$(cat ${SECRETS_DIR}/BM_ENROLLMENT_APISERVICE_USERNAME)
BM_ENROLLMENT_APISERVICE_PASSWORD=$(cat ${SECRETS_DIR}/BM_ENROLLMENT_APISERVICE_PASSWORD)
EAPI_USERNAME=$(cat ${SECRETS_DIR}/EAPI_USERNAME)
EAPI_PASSWD=$(cat ${SECRETS_DIR}/EAPI_PASSWD)
HELMFILE_DUMP=${HELMFILE_DUMP:-${SECRETS_DIR}/helmfile-dump.yaml}
OVERWRITE_SECRETS=${OVERWRITE_SECRETS:-0}
: "${VAULT_TOKEN:?environment variable is required}"
: "${VAULT_ADDR:?environment variable is required}"

query_config() {
  ${YQ} "$@" ${HELMFILE_DUMP}
}

get_regions() {
  query_config ".Values.regions[].region"
}

get_availability_zones_for_region() {
  region=$1
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

# Write a new secret.
vault_kv_put() {
  mount=$1
  key=$2
  shift 2
  if [ "${OVERWRITE_SECRETS}" == "0" ]; then
    if ${VAULT} kv get ${mount} ${key} > /dev/null; then
      echo "Key ${key} in ${mount} already exists and will not be replaced"
    else
      ${VAULT} kv put --cas=0 ${mount} ${key} "$@"
    fi
  else
    ${VAULT} kv put ${mount} ${key} "$@"
  fi
}

basic_postgres() {
  db=$1
  svc=${2:-${db}}
  vault_kv_put -mount=controlplane \
    ${svc}/database \
    username="dbuser" \
    password="$(cat ${SECRETS_DIR}/${db}_db_user_password)"
}

basic_dpai() {
  db=$1
  svc=${2:-${db}}
  echo $db
  vault_kv_put -mount=controlplane \
    ${svc}/secrets \
    postgres_password="$(cat ${SECRETS_DIR}/${db}_db_user_password)"\
    external_caas_password="$(cat ${SECRETS_DIR}/${db}_external_caas_password)"
}

basic_postgres_cloudmonitor() {
  db=$1
  svc=${2:-${db}}
  vault_kv_put -mount=controlplane \
    ${svc}/database \
    username="dbuser" \
    password="$(cat ${SECRETS_DIR}/${db}_db_user_password)" \
    vmtoken="ABCDEF"
}


basic_iks() {
  db=$1
  svc=$1
  echo $db
  vault_kv_put -mount=controlplane \
    ${svc}/database \
    username="$(cat ${SECRETS_DIR}/${db}_db_username)"\
    password="$(cat ${SECRETS_DIR}/${db}_db_user_password)" \
    username_rw="$(cat ${SECRETS_DIR}/${db}_db_username_rw)" \
    password_rw="$(cat ${SECRETS_DIR}/${db}_db_user_password_rw)"
  vault_kv_put -mount=controlplane \
    ${svc}/encryption_keys \
    1="$(cat ${SECRETS_DIR}/${db}_db_encryption_keys)"
  vault_kv_put -mount=controlplane \
    ${svc}/admin_key \
    1="$(cat ${SECRETS_DIR}/${db}_admin_key)" 
}

basic_insights() {
  db=$1
  svc=$1
  echo $db
  vault_kv_put -mount=controlplane \
    ${svc}/database \
    username="$(cat ${SECRETS_DIR}/${db}_db_username)"\
    password="$(cat ${SECRETS_DIR}/${db}_db_user_password)" \
    username_rw="$(cat ${SECRETS_DIR}/${db}_db_username_rw)" \
    password_rw="$(cat ${SECRETS_DIR}/${db}_db_user_password_rw)"
  vault_kv_put -mount=controlplane \
    ${svc}/github_key \
    github_key="1234567890"  
}

# Basic Auth GTS creds secrets for Product Catalog Operator
basic_gts() {
  svc=$1
  vault_kv_put -mount=controlplane \
    ${svc}/apigee \
    username="${GTS_USERNAME}" \
    password="${GTS_PASSWORD}"
}

wait_for_vault() {
  while ! ${VAULT} secrets list; do sleep 1; done
}

# Load public key of all Kubernetes clusters into Vault.
# TODO: utilize OVERWRITE_SECRETS flag
write_cluster_auth_config() {
  local config_path="auth/cluster-auth/config"
  if existing_data=$(${VAULT} read $config_path); then
    echo "$config_path already exists in Vault with $existing_data. Skipping $config_path write."
  else
    echo "$config_path does not exist. Need to create.."
    ${SCRIPT_DIR}/jwk-to-vault.py --input-dir ${SECRETS_DIR}/vault-jwk-validation-public-keys --output-file ${SECRETS_DIR}/vault-jwt-validation-public-keys.json
    cat ${SECRETS_DIR}/vault-jwt-validation-public-keys.json | ${VAULT} write $config_path -
  fi
}

# Public data (public certificates, etc.)
write_public() {
  # Load otel-ca.pem (public certificate).
  OTEL_CA_FILE=${SCRIPT_DIR}/../../../go/pkg/observability/otel-ca.pem
  vault_kv_put -mount=public otel otel_ca_pem=@${OTEL_CA_FILE}

  # Load ICP cert apis-intel-com.pem (public certificate).
  ICP_CA_FILE=${SCRIPT_DIR}/../../../go/pkg/cloudaccount_enroll/apis-intel-com.pem
  vault_kv_put -mount=public icp icp_ca_pem=@${ICP_CA_FILE}

  # Load AWS RDS certificate bundles (public certificates).
  # https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html
  vault_kv_put -mount=public globaldbssl sslcert=@${SCRIPT_DIR}/public/aws-rds-global-bundle.pem
}

write_cloudaccount_enroll() {
  vault_kv_put -mount=controlplane \
    cloudaccount-enroll/icp \
    username="${ICP_USERNAME}" \
    password="${ICP_PASSWORD}"
}

write_grafana() {
  vault_kv_put -mount=controlplane \
    grafana/secrets \
    username="${GRAFANA_USERNAME}" \
    password="${GRAFANA_PASSWORD}"
}

write_billing_aria() {
  vault_kv_put -mount=controlplane \
    billing-aria/secrets \
    authKey="$(cat ${SECRETS_DIR}/aria_auth_key)" \
    clientNo="$(cat ${SECRETS_DIR}/aria_client_no)" \
    apiCrt="$(cat ${SECRETS_DIR}/aria_api_crt)" \
    apikey="$(cat ${SECRETS_DIR}/aria_api_key)" \

  vault_kv_put -mount=controlplane \
    billing-aria/awssecrets \
    access_key_id="$(cat ${SECRETS_DIR}/billing_driver_aria_sqs_access_key_id)" \
    secret_access_key="$(cat ${SECRETS_DIR}/billing_driver_aria_sqs_secret_access_key)" \
    accountId="$(cat ${SECRETS_DIR}/billing_driver_aria_sqs_account_id)"
}
write_notification_gateway() {
  vault_kv_put -mount=controlplane \
    notification-gateway/awssecrets \
    access_key_id="$(cat ${SECRETS_DIR}/aws_credentials/access_key_id)" \
    secret_access_key="$(cat ${SECRETS_DIR}/aws_credentials/secret_access_key)" \
    accountId="$(cat ${SECRETS_DIR}/aws_credentials/account_id)"
}
write_user_credentials() {
  vault_kv_put -mount=controlplane \
    user-credentials/awssecrets \
    access_key_id="$(cat ${SECRETS_DIR}/aws_credentials/access_key_id)" \
    secret_access_key="$(cat ${SECRETS_DIR}/aws_credentials/secret_access_key)" \
    accountId="$(cat ${SECRETS_DIR}/aws_credentials/account_id)"
}
write_rate_limit_redis() {
  vault_kv_put -mount=controlplane \
    rate-limit-redis/credentials \
    password="$(cat ${SECRETS_DIR}/rate_limit_redis_password)"
}

write_rate_limit_redis_region() {
  vault_kv_put -mount=controlplane \
    ${1}-rate-limit-redis/credentials \
    password="$(cat ${SECRETS_DIR}/${1}-rate_limit_redis_password)"
}

write_global() {
  write_cluster_auth_config
  write_public
  basic_postgres authz
  basic_postgres cloudaccount
  basic_postgres metering
  basic_postgres cloudcredits
  basic_postgres usage
  basic_postgres billing
  basic_postgres notification notification-gateway
  basic_postgres_cloudmonitor cloudmonitor
  basic_postgres productcatalog
  basic_gts gts-trade-compliance
  write_cloudaccount_enroll
  write_grafana
  write_billing_aria
  write_notification_gateway
  write_user_credentials
  write_rate_limit_redis
}

process_roles_secrets() {
    local mount="$1"
    local kv_secret_path="$2" # KV secret in the mount that will be created such as "${region}/baremetal/enrollment/approle"
    local auth_approle_path="$3" # Auth approle such as "auth/approle/role/${region}-baremetal-enrollment-role"

    echo "process_roles_secrets:  kv_secret_path: $kv_secret_path auth_approle_path: $auth_approle_path"
    # TODO: Utilize OVERWRITE_SECRETS flag
    # Check if the role exists
    if ${VAULT} kv get ${mount} $kv_secret_path > /dev/null; then
      echo "Role $kv_secret_path already exists in Vault. Skipping role write."
    else
      echo "Role $kv_secret_path does not exist. Need to create.."
      ROLE_ID=$(${VAULT} read $auth_approle_path/role-id -format=json | ${JQ} -r .data.role_id)
      SECRET_ID=$(${VAULT} write -f $auth_approle_path/secret-id -format=json | ${JQ} -r .data.secret_id)
      echo "Creating role $kv_secret_path with secret_id:$SECRET_ID and role_id:$ROLE_ID"
      ${VAULT} kv put ${mount} $kv_secret_path secret_id=${SECRET_ID} role_id=${ROLE_ID}
    fi
}

write_baremetal_enrollment() {
  region=$1
  availability_zone=${region}a

  # create secret for region kubeconfig
  vault_kv_put -mount=controlplane ${region}/baremetal/enrollment/${availability_zone} kubeconfig="@${ENROLLMENT_KUBECONFIG}"

  # Create secret for netbox access, this secret is used only for integration testing
  vault_kv_put -mount=controlplane ${region}/baremetal/enrollment/netbox token="0123456789abcdef0123456789abcdef01234567"

  # Create  enrollment api service secrets
  vault_kv_put -mount=controlplane ${region}/baremetal/enrollment/apiservice username="${BM_ENROLLMENT_APISERVICE_USERNAME}" password="${BM_ENROLLMENT_APISERVICE_PASSWORD}"

  # create secret for virtual BMC, it is only used for virtual baremetal stack
  if [ -d ${IDC_ENV_DIR}/vault/bmc ]; then
      for MAC_ADDR in $(ls -p ${IDC_ENV_DIR}/vault/bmc/ | grep -E '^([0-9a-f]{2}-){5}[0-9a-f]{2}$' | tr '-' ':'); do
          vault_kv_put -mount=bmc ${region}/deployed/${MAC_ADDR}/default username="${DEFAULT_BMC_USERNAME}" password="${DEFAULT_BMC_PASSWD}"
      done
  else
      vault_kv_put -mount=bmc ${region}/deployed/virtual/default username="${DEFAULT_BMC_USERNAME}" password="${DEFAULT_BMC_PASSWD}"
  fi

  # create men and mice secrets
  vault_kv_put -mount=controlplane ${region}/baremetal/enrollment/menandmice username="${MEN_AND_MICE_USERNAME}" password="${MEN_AND_MICE_PASSWORD}"

  # load secret for IPA Image ssh key
  if [ -n "${IPA_IMAGE_SSH_PUB}" -a -n "${IPA_IMAGE_SSH_PRIV}" ]; then
      vault_kv_put -mount=controlplane ${region}/baremetal/enrollment/ipaimage/sshkey \
        publicKey=@${IPA_IMAGE_SSH_PUB} \
        privateKey=@${IPA_IMAGE_SSH_PRIV}
  fi

  process_roles_secrets -mount=controlplane "${region}/baremetal/enrollment/approle" "auth/approle/role/${region}-baremetal-enrollment-role"
}

write_ssh_proxy_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-ssh-proxy-operator/ssh \
    publickey="@${SSH_PROXY_OPERATOR_ID_RSA}.pub" \
    privatekey="@${SSH_PROXY_OPERATOR_ID_RSA}" \
    host_public_key="@${SSH_PROXY_SERVER_HOST_PUBLIC_KEY}"
}
write_storage_enrollment() {
  region=$1
  availability_zone=${region}a

  ## Storage kms
  process_roles_secrets -mount=controlplane "${region}/storage/kms/approle" "auth/approle/role/${region}-storage-kms-role"
}
write_bm_instance_operator() {
  availability_zone=$1
  # create kubeconfig
  vault_kv_put -mount=controlplane \
    ${availability_zone}-bm-instance-operator/sdnkubeconfig \
    kubeconfig="@${ENROLLMENT_KUBECONFIG}"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-bm-instance-operator/ssh \
    publickey="@${BM_INSTANCE_OPERATOR_ID_RSA}.pub" \
    privatekey="@${BM_INSTANCE_OPERATOR_ID_RSA}" \
    host_public_key="@${SSH_PROXY_SERVER_HOST_PUBLIC_KEY}"
}

write_bm_validation_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-bm-validation-operator/ssh \
    publickey="@${BM_VALIDATION_OPERATOR_ID_RSA}.pub" \
    privatekey="@${BM_VALIDATION_OPERATOR_ID_RSA}" \
    host_public_key="@${SSH_PROXY_SERVER_HOST_PUBLIC_KEY}"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-bm-validation-operator/aws \
    access_key_id="@${SECRETS_DIR}/VALIDATION_REPORTS_S3_ACCESS_KEY_ID" \
    secret_key="@${SECRETS_DIR}/VALIDATION_REPORTS_S3_SECRET_KEY"
  # Create secret for netbox access, this secret is used only for integration testing
  vault_kv_put -mount=controlplane ${availability_zone}-bm-validation-operator/netbox token=${NETBOX_TOKEN}
  # Create secret for huggingface token, this secret is used only for Gaudi validation test
  vault_kv_put -mount=controlplane \
    ${availability_zone}-bm-validation-operator/huggingface \
    hf_token="@${SECRETS_DIR}/bm-validation-operator/HUGGINGFACE_TOKEN"
}

# Load Harvester/kubevirt kubeconfig secret for vm-instance-operator.
write_vm_instance_operator() {
  local region=$1
  local availability_zone=$2
  write_harvester_kubeconfig ${region} ${availability_zone}
  write_kubevirt_kubeconfig ${region} ${availability_zone}
}

write_kubevirt_kubeconfig() {
  local region=$1
  local availability_zone=$2
  local kubevirt_clusterids=$(get_kubevirt_clusterids_for_region_az ${region} ${availability_zone})
  for kubevirt_clusterid in $kubevirt_clusterids; do
      kubevirt_cluster_kubeconfig=${SECRETS_DIR}/kubevirt-kubeconfig/${kubevirt_clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-vm-instance-operator-${kubevirt_clusterid}/kubeconfig \
        kubeconfig="@${kubevirt_cluster_kubeconfig}"
  done
}

write_harvester_kubeconfig() {
  local region=$1
  local availability_zone=$2
  local clusterids=$(get_harvester_clusterids_for_region_az ${region} ${availability_zone})
  for clusterid in $clusterids; do
      cluster_kubeconfig=${SECRETS_DIR}/harvester-kubeconfig/${clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-vm-instance-operator-${clusterid}/harvester_kubeconfig \
        kubeconfig="@${cluster_kubeconfig}"
  done
}

# Load Harvester and KubeVirt kubeconfigs secret for vm-instance-scheduler.
write_vm_instance_scheduler() {
  local region=$1
  local availability_zone=$2
  local harvester_clusterids=$(get_harvester_clusterids_for_region_az ${region} ${availability_zone})
  for clusterid in $harvester_clusterids; do
      cluster_kubeconfig=${SECRETS_DIR}/harvester-kubeconfig/${clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-vm-instance-scheduler/harvester_kubeconfig_${clusterid} \
        kubeconfig="@${cluster_kubeconfig}"
  done

  local kubevirt_clusterids=$(get_kubevirt_clusterids_for_region_az ${region} ${availability_zone})
  for kubevirt_clusterid in $kubevirt_clusterids; do
      kubevirt_cluster_kubeconfig=${SECRETS_DIR}/kubevirt-kubeconfig/${kubevirt_clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-vm-instance-scheduler/kubeconfig_${kubevirt_clusterid} \
        kubeconfig="@${kubevirt_cluster_kubeconfig}"
  done
}

write_compute_metering_monitor() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-compute-metering-monitor/cognito \
    client_id="@${SECRETS_DIR}/cognito_compute_metering_monitor_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_compute_metering_monitor_client_secret"
}

write_grpc_proxy() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-grpc-proxy-external/cognito \
    client_id="@${SECRETS_DIR}/cognito_region_grpc_proxy_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_region_grpc_proxy_client_secret"
}

write_app_client_grpc_proxy() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-app-client-api-grpc-proxy-external/cognito \
    client_id="@${SECRETS_DIR}/cognito_region_grpc_proxy_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_region_grpc_proxy_client_secret"
}

write_compute_api_server() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-compute-api-server/cognito \
    client_id="@${SECRETS_DIR}/cognito_region_compute_api_server_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_region_compute_api_server_client_secret"
}

write_storage_api_server() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-compute-api-server/cognito \
    client_id="@${SECRETS_DIR}/cognito_region_storage_api_server_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_region_storage_api_server_client_secret"
}

write_training_api_server() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-training-api-server/api \
    cacrt="@${SECRETS_DIR}/training_api_server_cacrt" \
    crt="@${SECRETS_DIR}/training_api_server_crt" \
    key="@${SECRETS_DIR}/training_api_server_key"
  vault_kv_put -mount=controlplane \
    ${region}-training-api-server/cognito \
    client_id="@${SECRETS_DIR}/cognito_region_training_api_server_client_id" \
    client_secret="@${SECRETS_DIR}/cognito_region_training_api_server_client_secret"
}

write_nginx_s3_gateway() {
  region=$1
  vault_kv_put -mount=controlplane \
    ${region}-compute-nginx-s3-gateway/aws \
    access_key_id="@${SECRETS_DIR}/NGINX_S3_GATEWAY_ACCESS_KEY_ID" \
    secret_key="@${SECRETS_DIR}/NGINX_S3_GATEWAY_SECRET_KEY"
}

write_provider_sdn() {
  region=$1

  # Arista Switch eAPI
  vault_kv_put -mount=controlplane \
    ${region}/provider-sdn-controller/eapi \
    username="${EAPI_USERNAME}" \
    password="${EAPI_PASSWD}"

  # Netbox token
  vault_kv_put -mount=controlplane \
    ${region}/provider-sdn-controller/netboxtoken \
    token="${NETBOX_TOKEN}"
}

write_sdn() {
  region=$1
  availability_zone=$2

  # SDN-Controller
  vault_kv_put -mount=controlplane \
    ${region}/${availability_zone}/nw-sdn-controller/raven \
    username="${RAVEN_USERNAME}" \
    password="${RAVEN_PASSWD}"

  # Arista Switch eAPI
  vault_kv_put -mount=controlplane \
    ${region}/${availability_zone}/nw-sdn-controller/eapi \
    username="${EAPI_USERNAME}" \
    password="${EAPI_PASSWD}"

  # copy the az's kubeconfig to network cluster vault folder
  vault_kv_put -mount=controlplane \
    ${region}/${availability_zone}/nw-sdn-controller/bmhkubeconfig \
    kubeconfig="@${ENROLLMENT_KUBECONFIG}"

  # copy the network cluster's restricted kubeconfig to BM controller vault folder
  # NETWORKING_KUBECONFIG is set in the "build/environments/${IDC_ENV}/Makefile.environment" file
  if [ -n "${NETWORKING_KUBECONFIG}" ]; then
      SDN_BMAAS_KUBECONFIG=${SECRETS_DIR}/restricted-kubeconfig/sdn-bmaas-kubeconfig.yaml
      echo "loading SDN Kubeconfig to Vault (for use by bm-instance-operator)"
      vault_kv_put -mount=controlplane \
        ${availability_zone}-bm-instance-operator/sdnkubeconfig \
        kubeconfig="@${SDN_BMAAS_KUBECONFIG}"
  fi

  # Netbox token
  vault_kv_put -mount=controlplane \
    ${region}/${availability_zone}/nw-sdn-controller/netboxtoken \
    token="${NETBOX_TOKEN}"

}

write_iks_kubernetes_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-kubernetes-operator/config \
    config="@${SECRETS_DIR}/kubernetes-operator/config.yaml"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-kubernetes-operator/bootstrap-iks-controlplane \
    bootstrap-iks-controlplane="@${SECRETS_DIR}/kubernetes-operator/bootstrap-iks-controlplane.sh"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-kubernetes-operator/bootstrap-iks-worker \
    bootstrap-iks-worker="@${SECRETS_DIR}/kubernetes-operator/bootstrap-iks-worker.sh"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-kubernetes-operator/bootstrap-rke2 \
    bootstrap-rke2="@${SECRETS_DIR}/kubernetes-operator/bootstrap-rke2.sh"
}

write_iks_kubernetes_reconciler() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-kubernetes-reconciler/config \
    config="@${SECRETS_DIR}/kubernetes-reconciler/config.yaml"
}

write_iks_ilb_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-ilb-operator/config \
    config="@${SECRETS_DIR}/ilb-operator/config.yaml"
}

harvester_k8s_resource_patcher() {
  local region=$1
  local availability_zone=$2

  clusterids=$(get_harvester_clusterids_for_region_az ${region} ${availability_zone})

  for clusterid in $clusterids; do
      cluster_kubeconfig=${SECRETS_DIR}/harvester-kubeconfig/${clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-k8s-resource-patcher-${clusterid}/harvester_kubeconfig \
        kubeconfig="@${cluster_kubeconfig}"
  done
}

kubevirt_k8s_resource_patcher() {
  local region=$1
  local availability_zone=$2

  kubevirt_clusterids=$(get_kubevirt_clusterids_for_region_az ${region} ${availability_zone})

  for clusterid in $kubevirt_clusterids; do
      cluster_kubeconfig=${SECRETS_DIR}/kubevirt-kubeconfig/${clusterid}
      vault_kv_put -mount=controlplane \
        ${availability_zone}-k8s-resource-patcher-${clusterid}/kubeconfig \
        kubeconfig="@${cluster_kubeconfig}"
  done
}

# Load Harvester/KubeVirt kubeconfigs secret for k8s-resource-patcher.
write_k8s_resource_patcher() {
  local region=$1
  local availability_zone=$2
  harvester_k8s_resource_patcher ${region} ${availability_zone}
  kubevirt_k8s_resource_patcher  ${region} ${availability_zone}
}

write_loadbalancer_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-loadbalancer-operator/api \
    api_username="$(cat ${SECRETS_DIR}/${availability_zone}-highwire_api_username)" \
    api_password="$(cat ${SECRETS_DIR}/${availability_zone}-highwire_api_password)" \
    azClusterKubeconfig="$(cat ${SECRETS_DIR}/${availability_zone}-loadbalancer-operator-azkubeconfig)"
}

write_firewall_operator() {
  availability_zone=$1
  vault_kv_put -mount=controlplane \
    ${availability_zone}-firewall-operator/api \
    api_username="$(cat ${SECRETS_DIR}/${availability_zone}-firewall_api_username)" \
    api_password="$(cat ${SECRETS_DIR}/${availability_zone}-firewall_api_password)"
}

write_quick_connect_api_server() {
  availability_zone=$1
  process_roles_secrets -mount=controlplane "${availability_zone}-quick-connect-api-server/approle" "auth/approle/role/${availability_zone}-quick-connect-api-server-role"
  vault_kv_put -mount=controlplane \
    ${availability_zone}-quick-connect-api-server/oauth2_client \
    token="$(cat ${SECRETS_DIR}/${availability_zone}-quick-connect-api-server-oauth-token)" \
    hmac="$(cat ${SECRETS_DIR}/${availability_zone}-quick-connect-api-server-oauth-hmac)"
}

write_kubevirt() {
  local region=$1
  local availability_zone=$2
  vault_kv_put -mount=controlplane \
    ${region}/${availability_zone}/rancher \
    url="@${SECRETS_DIR}/${availability_zone}-rancher-url" \
    access_key="@${SECRETS_DIR}/${availability_zone}-rancher-access_key" \
    secret_key="@${SECRETS_DIR}/${availability_zone}-rancher-secret_key"
}

write_region() {
  region=$1
  basic_postgres ${region}-cloudmonitor-logs ${region}-cloudmonitor-logs-api-server
  basic_postgres ${region}-compute ${region}-compute-api-server
  basic_postgres ${region}-fleet-admin ${region}-fleet-admin-api-server
  basic_postgres ${region}-fleet-admin ${region}-fleet-admin-ui-server
  basic_postgres ${region}-training ${region}-training-api-server
  basic_postgres ${region}-storage ${region}-storage-admin-api-server
  basic_postgres ${region}-storage ${region}-storage-api-server
  basic_postgres ${region}-network ${region}-network-api-server
  basic_postgres ${region}-network ${region}-network-operator
  basic_postgres ${region}-sdn-vn-controller
  basic_postgres ${region}-insights ${region}-security-insights
  basic_postgres ${region}-insights ${region}-security-scanner
  basic_postgres ${region}-insights ${region}-kubescore
  basic_postgres ${region}-quota-management-service
  basic_postgres ${region}-kfaas
  basic_postgres ${region}-dpai
  basic_iks ${region}-iks
  basic_insights ${region}-security-insights
  basic_insights ${region}-security-scanner
  basic_insights ${region}-kubescore
  basic_dpai ${region}-dpai
  write_baremetal_enrollment ${region}
  write_storage_enrollment ${region}
  write_grpc_proxy ${region}
  write_app_client_grpc_proxy ${region}
  write_compute_api_server ${region}
  write_storage_api_server ${region}
  write_training_api_server ${region}
  write_nginx_s3_gateway ${region}
  write_provider_sdn ${region}
  write_rate_limit_redis_region ${region}
}

write_az() {
  region=$1
  availability_zone=$2
  write_ssh_proxy_operator ${availability_zone}
  write_bm_instance_operator ${availability_zone}
  write_bm_validation_operator ${availability_zone}
  write_vm_instance_operator ${region} ${availability_zone}
  write_vm_instance_scheduler ${region} ${availability_zone}
  write_compute_metering_monitor ${availability_zone}
  write_sdn ${region} ${availability_zone}
  write_iks_kubernetes_operator ${availability_zone}
  write_iks_kubernetes_reconciler ${availability_zone}
  write_iks_ilb_operator ${availability_zone}
  write_k8s_resource_patcher ${region} ${availability_zone}
  write_loadbalancer_operator ${availability_zone}
  write_firewall_operator ${availability_zone}
  write_quick_connect_api_server ${availability_zone}
  write_kubevirt ${region} ${availability_zone}
}

list_vault() {
  ${VAULT} kv list public/
  ${VAULT} kv list controlplane/
  ${VAULT} kv get -mount=controlplane ${region}/baremetal/enrollment/approle
  ${VAULT} kv get -mount=controlplane ${region}/storage/kms/approle
  ${VAULT} kv get -mount=controlplane ${availability_zone}-quick-connect-api-server/approle
}

main() {
  wait_for_vault
  write_global

  regions=$(get_regions)
  for region in ${regions}; do
    write_region ${region}
    availability_zones=$(get_availability_zones_for_region ${region})
    for availability_zone in $availability_zones; do
      write_az ${region} ${availability_zone}
    done
  done

  list_vault
  echo load-secrets.sh: Done.
}

main
