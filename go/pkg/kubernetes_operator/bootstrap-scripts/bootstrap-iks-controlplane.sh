#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -o pipefail
set -o nounset
set -o errexit

err_report() {
  echo "Exited with error on line $1"
}
trap 'err_report $LINENO' ERR

log_message() {
    command echo $(date) "$@"
}

setup_fluentbit() {
  if [[ "${LOGGING_ENABLED}" == true ]]; then
    while [ ! -f /etc/systemd/system/fluentbit.service ]; do
      FLUENTBIT_CONFIG_FILE="/etc/fluentbit/fluentbit.yaml"
      FLUENTBIT_ENV_FILE="/etc/default/fluentbit"

# write fluentbit env file
cat > $FLUENTBIT_ENV_FILE <<EOF
LOGGING_HOST=${LOGGING_HOST}
LOGGING_USER=${LOGGING_USER}
LOGGING_PASSWORD=${LOGGING_PASSWORD}
$(echo "${LOGGING_ENRICHMENT}" | tr ',' '\n')
EOF
      # Download fluentbit binary
      wget http://135.232.98.200:8080/fluent-bit -O /usr/local/bin/fluentbit && chmod +x /usr/local/bin/fluentbit

      log_message "Configuring FluentBit package"

      dirname $FLUENTBIT_CONFIG_FILE | xargs mkdir -p

      source /etc/default/fluentbit

cat > "$FLUENTBIT_CONFIG_FILE" <<EOF
service:
  log_level: info
  flush: 5
pipeline:
  inputs:
    - name: systemd
      tag: host.*
      db: /var/log/flb_systemd.db
      path: /var/log/journal
      systemd_filter:  _SYSTEMD_UNIT=kube-scheduler.service
      systemd_filter:  _SYSTEMD_UNIT=kube-controller-manager.service
      systemd_filter:  _SYSTEMD_UNIT=kube-apiserver.service
      systemd_filter:  _SYSTEMD_UNIT=etcd.service
      systemd_filter:  _SYSTEMD_UNIT=konnectivity-server.service
      systemd_filter:  _SYSTEMD_UNIT=cadvisor.service
  outputs:
    - name: opensearch
      match: '*'
      processors:
        logs:
          - name: lua
            call: extract_field
            code: |
              function extract_field(tag, timestamp, record)
                local new_record = {}
                new_record["timestamp"] = timestamp
                new_record["cloudaccount_id"] = "${CLOUD_ACCOUNT_ID}"
                new_record["customer_cloudaccount_id"] = "${CUSTOMER_CLOUD_ACCOUNT_ID}"
                new_record["cluster_id"] = "${CLUSTER_ID}"
                new_record["cluster_name"] = "${CLUSTER_NAME}"
                new_record["cluster_region"] = "${CLUSTER_REGION}"
                new_record["host"] = record['_HOSTNAME'] or "UNKNOWN_HOST"
                new_record["system_component"] = record["_COMM"] or "UNKNOWN_SYSTEM_COMPONENT"
                new_record["log"] = record["MESSAGE"] or "EMPTY_MESSAGE"
                return 1, timestamp, new_record
              end
      host: ${LOGGING_HOST}
      port: 443
      index: ${CLOUD_ACCOUNT_ID}
      suppress_type_name: on
      http_user: ${LOGGING_USER}
      http_passwd: ${LOGGING_PASSWORD}
      logstash_format: off
      tls: on
      tls.verify: off
EOF

      log_message "Creating fluentbit systemd service"
cat > /etc/systemd/system/fluentbit.service <<EOF
[Unit]
Description=FluentBit: IKS CP log enricher and shipper
Documentation=https://docs.fluentbit.io/manual/
Requires=network.target
After=network.target

[Service]
Type=simple
EnvironmentFile=-$FLUENTBIT_ENV_FILE
ExecStart=/usr/local/bin/fluentbit -c $FLUENTBIT_CONFIG_FILE
Restart=always

[Install]
WantedBy=multi-user.target

EOF
      log_message "Starting fluentbit service"
      systemctl daemon-reload
      systemctl enable fluentbit
      systemctl start fluentbit
    done
  else
    log_message "FluentBit logging disabled"
  fi
}

IFS=$'\n\t'

function print_help {
  echo "usage: $0 [options]"
  echo "Bootstraps an instance into an IDC Kubernetes cluster"
  echo ""
  echo "-h,--help print this help"
  echo "--ca-cert The Kubernetes CA certificate"
  echo "--ca-key The Kubernetes CA private key"
  echo "--etcd-ca-cert The etcd CA certificate"
  echo "--etcd-ca-key The etcd CA private key"
  echo "--front-proxy-ca-cert The CA certificate for aggregation layer"
  echo "--front-proxy-ca-key The CA private key for aggregation layer"
  echo "--sa-private-key The private key to sign service account tokens"
  echo "--sa-public-key The public key to verify service account tokens"
  echo "--apiserver-lb The Kubernetes apiserver loadbalancer"
  echo "--public-apiserver-lb The Kubernetes apiserver public loadbalancer"
  echo "--etcd-lb The etcd loadbalancer"
  echo "--etcd-initial-cluster The list of etcd members"
  echo "--etcd-cluster-state The state of the etcd cluster"
  echo "--konnectivity-lb The konnectivity loadbalancer"
  echo "--cluster-name The name of the kubernetes cluster"
  echo "--cluster-cidr The IP range used for pods"
  echo "--service-cidr The IP range used for services"
  echo "--cp-cert-expiration-period The Control Plane Expiration Time for services"
  echo "--etcd-encryption-config The Etcd encryption configuration"
  echo "--iptables-enabled Enable iptables change to allow only controlplane CIDR communication"
  echo "--iptables-cidr The CIDR used by controlplane nodes"
  echo "--logging-enabled Enable Fluentbit logging enrichment and shipping"
  echo "--logging-host The URL for Fluentbit log shipping, currently only OpenSearch is supported"
  echo "--logging-user The OpenSearch username"
  echo "--logging-password The OpenSearch password"
  echo "--logging-enrichment Comma separated logging enrichment string K1=V1,K2=V2."
  echo "--system-metrics-enabled Enable Prometheus remote write"
  echo "--system-metrics-prometheus-url The URL for Prometheus remote write"
  echo "--system-metrics-prometheus-username The username for Prometheus remote write"
  echo "--system-metrics-prometheus-password The password for Prometheus remote write"
  echo "--region The region of the cluster"
  echo "--cloudaccount The cloud account of the cluster"
  echo "--end-user-metrics-enabled Enable Prometheus remote write for end users"
  echo "--end-user-metrics-prometheus-url The URL for Prometheus remote write for end users"
  echo "--end-user-metrics-prometheus-bearer-token The bearer token for Prometheus remote write for end users"
}

POSITIONAL=()

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -h | --help)
      print_help
      exit 1
      ;;
    --cluster-name)
      CLUSTER_NAME=$2
      shift
      shift
      ;;
    --ca-cert)
      CA_CERT=$2
      shift
      shift
      ;;
    --ca-key)
      CA_KEY=$2
      shift
      shift
      ;;
    --etcd-ca-cert)
      ETCD_CA_CERT=$2
      shift
      shift
      ;;
    --etcd-ca-key)
      ETCD_CA_KEY=$2
      shift
      shift
      ;;
    --front-proxy-ca-cert)
      FRONT_PROXY_CA_CERT=$2
      shift
      shift
      ;;
    --front-proxy-ca-key)
      FRONT_PROXY_CA_KEY=$2
      shift
      shift
      ;;
    --sa-private-key)
      SA_PRIVATE_KEY=$2
      shift
      shift
      ;;
    --sa-public-key)
      SA_PUBLIC_KEY=$2
      shift
      shift
      ;;
    --apiserver-lb)
      APISERVER_LB=$2
      shift
      shift
      ;;
    --apiserver-lb-port)
      APISERVER_LB_PORT=$2
      shift
      shift
      ;;
    --public-apiserver-lb)
      PUBLIC_APISERVER_LB=$2
      shift
      shift
      ;;
    --public-apiserver-lb-port)
      PUBLIC_APISERVER_LB_PORT=$2
      shift
      shift
      ;;
    --etcd-lb)
      ETCD_LB=$2
      shift
      shift
      ;;
    --etcd-lb-port)
      ETCD_LB_PORT=$2
      shift
      shift
      ;;
    --konnectivity-lb)
      KONNECTIVITY_LB=$2
      shift
      shift
      ;;
    --etcd-initial-cluster)
      ETCD_INITIAL_CLUSTER=$2
      shift
      shift
      ;;
    --etcd-cluster-state)
      ETCD_CLUSTER_STATE=$2
      shift
      shift
      ;;
    --etcd-cluster-state)
      ETCD_CLUSTER_STATE=$2
      shift
      shift
      ;;
    --cluster-cidr)
      CLUSTER_CIDR=$2
      shift
      shift
      ;;
    --service-cidr)
      SERVICE_CIDR=$2
      shift
      shift
      ;;
    --cp-cert-expiration-period)
      CP_CERT_EXPIRATION_PERIOD=$2
      shift
      shift
      ;;
    --etcd-encryption-config)
      ETCD_ENCRYPTION_CONFIG=$2
      shift
      shift
      ;;
    --iptables-enabled)
      IPTABLES_ENABLED=$2
      shift
      shift
      ;;
    --iptables-cidr)
      IPTABLES_CIDR=$2
      shift
      shift
      ;;
    --system-metrics-enabled)
      SYSTEM_METRICS_ENABLED=$2
      shift
      shift
      ;;
    --system-metrics-prometheus-url)
      SYSTEM_METRICS_PROMETHEUS_URL=$2
      shift
      shift
      ;;
    --system-metrics-prometheus-username)
      SYSTEM_METRICS_PROMETHEUS_USERNAME=$2
      shift
      shift
      ;;
    --system-metrics-prometheus-password)
      SYSTEM_METRICS_PROMETHEUS_PASSWORD=$2
      shift
      shift
      ;;
    --logging-enabled)
      LOGGING_ENABLED=$2
      shift
      shift
      ;;
    --logging-host)
      LOGGING_HOST=$2
      shift
      shift
      ;;
    --logging-user)
      LOGGING_USER=$2
      shift
      shift
      ;;
    --logging-password)
      LOGGING_PASSWORD=$2
      shift
      shift
      ;;
    --logging-enrichment)
      LOGGING_ENRICHMENT=$2
      shift
      shift
      ;;
    --region)
      REGION=$2
      shift
      shift
      ;;
    --cloudaccount)
      CLOUDACCOUNT=$2
      shift
      shift
      ;;
    --end-user-metrics-enabled)
      END_USER_METRICS_ENABLED=$2
      shift
      shift
      ;;
    --end-user-metrics-prometheus-url)
      END_USER_METRICS_PROMETHEUS_URL=$2
      shift
      shift
      ;;
     --end-user-metrics-prometheus-bearer-token)
      END_USER_METRICS_PROMETHEUS_BEARER_TOKEN=$2
      shift
      shift
      ;;
    *)                   # unknown option
      POSITIONAL+=("$1") # save it in an array for later
      shift              # past argument
      ;;
  esac
done


log_message "Starting the IDC Kubernetes bootstrap script"

# Checking if OS is supported. Ubuntu is the only supported OS at this moment
DISTRO=$(lsb_release -i | cut -d: -f2 | sed s/'^\t'//)
OS_VERSION=$(lsb_release -sr)
if [[ ! "${DISTRO}" == "Ubuntu" ]]; then
  log_message "${DISTRO} is not supported"
  exit 1
fi

# Check if it's being executed by root
if [[ ! $(id -u) -eq 0 ]]; then
  log_message "The script must be run as root"
  exit 1
fi

# Add Intel certificates.
# TODO: Remove this once added to IMI.
echo LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUcxRENDQkx5Z0F3SUJBZ0lURkFBQUFBZlB4MXVwZWoyR053QUFBQUFBQnpBTkJna3Foa2lHOXcwQkFRc0YKQURCSU1Rc3dDUVlEVlFRR0V3SlZVekVhTUJnR0ExVUVDaE1SU1c1MFpXd2dRMjl5Y0c5eVlYUnBiMjR4SFRBYgpCZ05WQkFNVEZFbHVkR1ZzSUZOSVFUSTFOaUJTYjI5MElFTkJNQjRYRFRJeU1EVXhPREUzTVRnek1sb1hEVEkzCk1EVXhPREUzTWpnek1sb3dVREVMTUFrR0ExVUVCaE1DVlZNeEdqQVlCZ05WQkFvVEVVbHVkR1ZzSUVOdmNuQnYKY21GMGFXOXVNU1V3SXdZRFZRUURFeHhKYm5SbGJDQkpiblJsY201aGJDQkpjM04xYVc1bklFTkJJRFZCTUlJQwpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQ0NnS0NBZ0VBeVN0MmRSOXN5Ymt6dmdsRitjc0RnNGxiCnVHRjc1d1QxOXlyT2gyUGtpbTM5UHhSbDZQRUwzdEV6YkdSZ1BFY0VxUEFGdHA1RjhnMmFuaWR2N01NbW9SZVAKbjJmZnB3RmdNL1I2QWNmWDVGdXVnc2ZjTG1qTUVxNnhsVnNYS08yc2JPalBpT3lOMWxXbGtncENNQ0IyZkZGYgpCSmpMdkNUaFE0b3JRYjY2THF6endJR1Z0MUdGLzMrZFMrQzhQTkRXWFlhSVZZNHUySE1ZMVlCc0lDQ0lwazFrCnBqN3JCRjByYzVkSjhsWTc1cFJOVndaREdoN0Fydk5mVC9QNE5MVGFMd3R5Y2NUbDUyNjA2T2FhMzFJRU9lT04KNVNPcVJUeEtpOXp0NGZBYzBxWXJhRmlFcGZjOFluNHAzbUtGbnF3RitKYXRaQTRJMXdGZCt1aER2YXBZNHE5Uwo0ZGtGUFMzQWc0UGVSMGxsZG5saE9qSWlzbzBRZTdDVUNxVnJjaS92akl1Vnc4aTMyVzBpSG83WUtXQW5BSHhJCnVTOVNQMFdCajNwNlNxVGdEdUQ3aUxZUnBmWXpXV0N0ckxqV2FNL0dyYkVaRUtXdDBZK2YwM25PVTgyOCtNUVYKY3Z0ZkloRjV6UmNMbXB6NU94T1RpaXQyd0g3b0tkemdsV1FzcWV2MkJmL0FSMytEa1VmWU14MDRONTdSdjNzWApTbWZhK0h1WE4wdnVLbGU3ZmNWajNWZWxaYUJBN0Z0QjQ2QTNaaDFJQUc1aExHcExud2FqOTZDdm9xRTNpejdWCjArM1FPV3V1SXFiRHZDUTkzeUU5bExoRmtqZTVodXYveXV6YnBwa0lNeVc3MWozWmk0Q0RSMmYzT3VLQjFJYjcKWkFJQ1VXT2U0S2FJUGt0WUtNc0NBd0VBQWFPQ0FhMHdnZ0dwTUJJR0NTc0dBUVFCZ2pjVkFRUUZBZ01DQUFJdwpJd1lKS3dZQkJBR0NOeFVDQkJZRUZDa2RWNTA1eEJDaWN0dlJXNDV1cVBNRlF0aVVNQjBHQTFVZERnUVdCQlJwCmtKWnB4d0JwbHR0N1hacVA5anJxS0pTdEpqQkVCZ05WSFNBRVBUQTdNRGtHQ3lxR1NJYjRUUUVGQVdVQk1Db3cKS0FZSUt3WUJCUVVIQWdFV0hHaDBkSEE2THk5d2Eya3VhVzUwWld3dVkyOXRMMk53Y3k1d1pHWXdHUVlKS3dZQgpCQUdDTnhRQ0JBd2VDZ0JUQUhVQVlnQkRBRUV3Q3dZRFZSMFBCQVFEQWdHR01BOEdBMVVkRXdFQi93UUZNQU1CCkFmOHdId1lEVlIwakJCZ3dGb0FVaVovUXN1MGdxYVkrU21hcThoNkNzTllzQ1RZd1B3WURWUjBmQkRnd05qQTAKb0RLZ01JWXVhSFIwY0RvdkwzQnJhUzVwYm5SbGJDNWpiMjB2WTNKc0wwbHVkR1ZzVTBoQk1qVTJVbTl2ZEVOQgpMbU55YkRCdUJnZ3JCZ0VGQlFjQkFRUmlNR0F3T2dZSUt3WUJCUVVITUFLR0xtaDBkSEE2THk5d2Eya3VhVzUwClpXd3VZMjl0TDJOeWRDOUpiblJsYkZOSVFUSTFObEp2YjNSRFFTNWpjblF3SWdZSUt3WUJCUVVITUFHR0ZtaDAKZEhBNkx5OVBRMU5RTG1sdWRHVnNMbU52YlM4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQkFNRXI0bFRYQ2k2aApDUWZtbENQMXdyM3o2QmZVcHpmRmZFTXFCMVlBYXViVzBkNm9GMGY4aTVMU0pVeFB6YjE2NURjUFJWajF2eFIxCnZZbWNZdUlFdG9HNVkwT0xvVkk1N1FyYThsMDE5bGIvZWxsNTJDSElXOWJzeTJRYmxHcHVSMzhqeStySlp2MXIKNTIwWVFST01IUkt5TnZyYW16T3pXeElJVmNUdW5LOHhJUGpCWHVoVnJpaFpSS0FRYXUzdDVIT1hLVWlYN2NKMAplcElhVUVZazJqYk1nS0JXUndPZ0pRNDI0d1RCL0VrL3k0VTNLazU5aDZEVzJqUUQ2ZTdVOFRkbm05Rm5mVlRvCjZ6RUphSmlacVFpN2tWTVlxTHZubE02a2NWNmNxZkhMNzlWV1hIYTRQS2dGRnBoRi9pZTBpUVIvZndWcEtmWmYKanI5eGdhcCttb1VzTXFZanlWOG41L0VxeDNhM2s0elNaZVl1bHhxVTAwSzF6RUpERTE1MDF6TnhGbWdhVmQvSwpNWlZYc0tRMjNRWkRtNzJSYjBVWU0wQy9tMnczTENabk1OY2dOWjVwWnYra1VhblJhb09URUJ1OVVxRE5QNEdlCmdkcUptT0NYSGtHWHZhbUJzZnNhM1VrWGpjSm5QbzBiODNYVlhpRjJ2WU5lbzFJa3orQXRtbTRIVi91c0JjTEcKVVFoSXJmMHNCN1dmSjhVUytsTGtjQWVWWXB4WmloUkk5dlZxL1N3eUFKaU5TS0M3TjlzaWlYRVZ0eHpFUlRTYwo1WnBIT2kwRTNhbjRHanBhNm16ekwxcWdtYytib3pzVjQwU1VHUEtocXJJUXRGVlFZOURiekROUFozYUdRZGJOClFIc1FwY3AyS3hvQkFkRFQ3TDl6U09qVXZEV0pDTGxWCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZhekNDQTFPZ0F3SUJBZ0lRVmpFblpxcWoxcVZLVW9hWmdCYzE4REFOQmdrcWhraUc5dzBCQVFzRkFEQkkKTVFzd0NRWURWUVFHRXdKVlV6RWFNQmdHQTFVRUNoTVJTVzUwWld3Z1EyOXljRzl5WVhScGIyNHhIVEFiQmdOVgpCQU1URkVsdWRHVnNJRk5JUVRJMU5pQlNiMjkwSUVOQk1CNFhEVEUxTURreE5qSXdNREF3TlZvWERUTTFNRGt4Ck5qSXdNREF3TlZvd1NERUxNQWtHQTFVRUJoTUNWVk14R2pBWUJnTlZCQW9URVVsdWRHVnNJRU52Y25CdmNtRjAKYVc5dU1SMHdHd1lEVlFRREV4UkpiblJsYkNCVFNFRXlOVFlnVW05dmRDQkRRVENDQWlJd0RRWUpLb1pJaHZjTgpBUUVCQlFBRGdnSVBBRENDQWdvQ2dnSUJBTkVITEQ0eDdvNHhLQ0RZVHdYR01pczFxb2IrdUJsTVIxN1ByMTFpCncvTkVSQXg0bmFoR2d6VFptWDlpM3ZpVHRmbjV1RFg2SUpNUGRxN040bStGbHI4djRWdUF6bnZuQ0JHcWtVRGUKSEFQYlI2clhVNGxNajdsTks3c0FXYk11SjgrMGEwU2M0d0V5Tk9YWWJOelM0cExjcTBlTDhBakZ3U3BGTll2MQpOb3BwQjJ0cG5tVFV1K0tqeW1hWGNYRXViYVd1b25tSHZqcUV2elJqaWNxVTEvUU1ZY29GT3czR1A3Z053ejJXCm5PS2huTVNiczhDZ3VCSzllSUJkQVlPc2V0ZENPRW0yczh5WXJxZXF0QldNVldxOTVxaVJBZG16L3JwV0JsRmwKTndsWHBFR1NsR0lJMzc0K0UyL2dQZU9YM3g2Z1VmcEFpVmVYbmk3cHVONGl4UVd5Q3hFSzBXKzUxeERsdE8xcAowb0ZzbDlQdDlVV3BRcVpvQThzRlhzVGYrSXdvQXpiOUpjR29QbHJ6Zk1mdUwrcWdzc0Y4Nko0aTAxZHJ6Uks4CmQrTmhpbkNzakxwZytRdzF3ZS9SUnl2NkNYdkZZQ1ZLcXRFQ0NZb2NZK050SjNqRER0dWt5V0xVZlhNZEZnVGkKSng2SXZFZGluZlYwUlF6KzhCWlkzc05RUU5DNjYvZFBpWldFUjBUWHhvbmpmclMxemUvNlZuZ1VsYTZMOTVMQwpVd2x4cE9kNWN1M21wYnRZdmlkT3d1ZkVRV3cxSUYreGdWakZ4MGk1VkdnOGpSc00zVFNSL3Yzb1BxWS9RdjNLCjZRRlpwc3BXQUpjR3NJdXF5S2ZoY0dMRE5iZFUvdGlWTlpmZDJqdDBycDQwazEyS3lpZzNTNGhlR01CdmRNMDgKVU9ZTkFnTUJBQUdqVVRCUE1Bc0dBMVVkRHdRRUF3SUJoakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQjBHQTFVZApEZ1FXQkJTSm45Q3k3U0NwcGo1S1pxcnlIb0t3MWl3Sk5qQVFCZ2tyQmdFRUFZSTNGUUVFQXdJQkFEQU5CZ2txCmhraUc5dzBCQVFzRkFBT0NBZ0VBSVVuVnFzbFJXdmwrVEFSYlpIZ1BlKzlYUFVESzI4QmN0dk5CYkNEK0FxeG4KNlB3WGt4eW9DZC9waUxzY1ZwQXpraWpUdXllSWJIK2sySkVrS3VLQ1ZhYnNmME9vUnA1WTRqRXdwWHVQcDRiagpXcEduWlRtM2h3bXhOWmszc21GenNXZ1M3d2lKUlNwU0tYaU8zcDlMRVZzdWtSNVJGcHV5N3VLT2RTN0VyTDNBCnNTT3FSVHVOUkdFN2Q2ams1bWNBNkxETzhnK1lCSS9QWlFlUjBCOGlRR2kzanJKZE1mMkdpZlJWYWx6UzBPV1AKVHBFa2xoMXVRVlA5a0c1dUZnNkhHcUQrY3JjMERwNFk2TGFwYW5aSFdxN3FqaUw2b1J3Ky80VExVek9OWXdjbQpvOHY0TXRnWW1sRTdJbzJTNFdLbFZ3eXlIRUdjNGtpM2xnRVB4Rm0xYVZ5UlA1VytEa3NwWC8reGxnTG1NeGplCmJHdUlPZ2NpZXVUbWl4V2FRSlFnRFp4aktCWU1vdUk3a25uSGlDZ1QxODRFZDNxeU5Sd29PeFJDOGRWeS9yNFUKNkxtTmY3M28yeVRQbGhHaGo5Yi96RUtyS1NDcmtXSzVDTU84QjFpaE9oK2YwbmhGK1V0b1hrY2VVRG1JVnNZVgprN3hZZG9xWFhEYnI1VG1NWDZsY3NDOFhIak9NRk96RmN6VHdWcjluSnc4c3NPSzhzMyt0K0sxNHFmMzNTaWJKCkkyNmhpZUFrUFhpY3ZuR2MrRVZSTFJDbjVHWnJkZnJUQlo4S1ZhazRFOWxYbXErM09oWjMyNnNjaEYwWk5WVFkKZ2VZNGNiaU50NHVhTEtWellsUXUzMjdwc3E1WkViampSL2F1cmhiYUpXQTJZZU1ZMlNCOEkyY2hOSE5oSnA4PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg== | base64 -d > /usr/local/share/ca-certificates/intelca.crt
/usr/sbin/update-ca-certificates

# Configure iptables to only allow controlplane nodes communication in port 2380.
if [[ "${IPTABLES_ENABLED}" == true ]]; then
  echo "Configuring iptables"
  mkdir -p /etc/iptables

  iptables -A INPUT -p tcp --dport 2380 -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -m state --state NEW,ESTABLISHED -j ACCEPT
  iptables -A INPUT -p tcp --sport 2380 -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -m state --state ESTABLISHED -j ACCEPT
  iptables -A INPUT -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -j DROP

  iptables -A OUTPUT -p tcp --dport 2380 -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -m state --state NEW,ESTABLISHED -j ACCEPT
  iptables -A OUTPUT -p tcp --sport 2380 -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -m state --state ESTABLISHED -j ACCEPT
  iptables -A OUTPUT -s "${IPTABLES_CIDR}" -d "${IPTABLES_CIDR}" -j DROP

  iptables-save | tee -a /etc/iptables/rules.v4
fi


# Get current host ip
NODE_IP=$(ip -brief addr | grep -oE '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | grep -vE '^127' | head -n1)
echo "Node ip: ${NODE_IP}"

if [[ "${ETCD_CLUSTER_STATE}" == "new" ]]; then
  echo "creating a new etcd cluster"
  ETCD_INITIAL_CLUSTER="${NODE_IP}=https://${NODE_IP}:2380"
else
  echo "adding a new etcd member"
  ETCD_INITIAL_CLUSTER="${ETCD_INITIAL_CLUSTER},${NODE_IP}=https://${NODE_IP}:2380"
fi

# Create required directories
mkdir -p /etc/kubernetes/pki/etcd /etc/kubernetes/enc/ /var/lib/etcd
chmod 700 /var/lib/etcd
if ! getent group etcd >/dev/null 2>&1; then
    groupadd etcd
fi
if ! id -u etcd >/dev/null 2>&1; then
    useradd -r -g etcd etcd
fi
chown etcd:etcd /var/lib/etcd

# Store ca certs and keys
echo "${CA_CERT}" | base64 --decode > /etc/kubernetes/pki/ca.crt
echo "${CA_KEY}" | base64 --decode > /etc/kubernetes/pki/ca.key
echo "${ETCD_CA_CERT}" | base64 --decode > /etc/kubernetes/pki/etcd/ca.crt
echo "${ETCD_CA_KEY}" | base64 --decode > /etc/kubernetes/pki/etcd/ca.key
echo "${FRONT_PROXY_CA_CERT}" | base64 --decode > /etc/kubernetes/pki/front-proxy-ca.crt
echo "${FRONT_PROXY_CA_KEY}" | base64 --decode > /etc/kubernetes/pki/front-proxy-ca.key
echo "${SA_PUBLIC_KEY}" | base64 --decode > /etc/kubernetes/pki/sa.pub
echo "${SA_PRIVATE_KEY}" | base64 --decode > /etc/kubernetes/pki/sa.key
echo "${ETCD_ENCRYPTION_CONFIG}" | base64 --decode > /etc/kubernetes/enc/enc.yaml

chmod 400 -R /etc/kubernetes/enc/

# Create certificates
# apiserver
FIRST_8BIT_OCTET=$(echo "${SERVICE_CIDR}" | cut -d"." -f1)
SECOND_8BIT_OCTET=$(echo "${SERVICE_CIDR}" | cut -d"." -f2)
KUBERNETES_SERVICE_IP="${FIRST_8BIT_OCTET}.${SECOND_8BIT_OCTET}.0.1"

cat > openssl.cnf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster.local
IP.1 = ${KUBERNETES_SERVICE_IP}
IP.2 = 127.0.0.1
IP.3 = ${APISERVER_LB}
IP.4 = ${PUBLIC_APISERVER_LB}
IP.5 = ${NODE_IP}
IP.6 = ${KONNECTIVITY_LB}
EOF

openssl genrsa -out /etc/kubernetes/pki/apiserver.key 4096
openssl req -new -key /etc/kubernetes/pki/apiserver.key -subj "/CN=kube-apiserver" -out apiserver.csr -config openssl.cnf
openssl x509 -req -in apiserver.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out /etc/kubernetes/pki/apiserver.crt -extensions v3_req -extfile openssl.cnf -days ${CP_CERT_EXPIRATION_PERIOD}
rm -f openssl.cnf apiserver.csr

# apiserver-kubelet-client
openssl genrsa -out /etc/kubernetes/pki/apiserver-kubelet-client.key 4096
openssl req -new -key /etc/kubernetes/pki/apiserver-kubelet-client.key -subj "/CN=kube-apiserver-kubelet-client/O=system:masters" -out apiserver-kubelet-client.csr
openssl x509 -req -in apiserver-kubelet-client.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out /etc/kubernetes/pki/apiserver-kubelet-client.crt -days ${CP_CERT_EXPIRATION_PERIOD}
rm -rf apiserver-kubelet-client.csr

# service-account
#openssl genrsa -out /etc/kubernetes/pki/sa.key 4096
#openssl rsa -in /etc/kubernetes/pki/sa.key -pubout -out /etc/kubernetes/pki/sa.pub

# etcd server
cat > openssl.cnf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
IP.1 = 127.0.0.1
IP.2 = ${ETCD_LB}
IP.3 = ${NODE_IP}
EOF

openssl genrsa -out /etc/kubernetes/pki/etcd/server.key 4096
openssl req -new -key /etc/kubernetes/pki/etcd/server.key -subj "/CN=kube-etcd" -out server.csr -config openssl.cnf
openssl x509 -req -in server.csr -CA /etc/kubernetes/pki/etcd/ca.crt -CAkey /etc/kubernetes/pki/etcd/ca.key -CAcreateserial -out /etc/kubernetes/pki/etcd/server.crt -extensions v3_req -extfile openssl.cnf -days ${CP_CERT_EXPIRATION_PERIOD}
rm -f openssl.cnf server.csr


# etcd peer
cat > openssl.cnf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
IP.1 = 127.0.0.1
IP.2 = ${NODE_IP}
EOF

openssl genrsa -out /etc/kubernetes/pki/etcd/peer.key 4096
openssl req -new -key /etc/kubernetes/pki/etcd/peer.key -subj "/CN=kube-etcd-peer" -out peer.csr -config openssl.cnf
openssl x509 -req -in peer.csr -CA /etc/kubernetes/pki/etcd/ca.crt -CAkey /etc/kubernetes/pki/etcd/ca.key -CAcreateserial -out /etc/kubernetes/pki/etcd/peer.crt -extensions v3_req -extfile openssl.cnf -days ${CP_CERT_EXPIRATION_PERIOD}
rm -f openssl.cnf peer.csr

# apiserver-etcd-client
openssl genrsa -out /etc/kubernetes/pki/etcd/apiserver-etcd-client.key 4096
openssl req -new -key /etc/kubernetes/pki/etcd/apiserver-etcd-client.key -subj "/CN=kube-apiserver-etcd-client/O=system:masters" -out apiserver-etcd-client.csr
openssl x509 -req -in apiserver-etcd-client.csr -CA /etc/kubernetes/pki/etcd/ca.crt -CAkey /etc/kubernetes/pki/etcd/ca.key -CAcreateserial -out /etc/kubernetes/pki/etcd/apiserver-etcd-client.crt -days ${CP_CERT_EXPIRATION_PERIOD}
rm -rf apiserver-etcd-client.csr

# etcd-healthcheck-client
openssl genrsa -out /etc/kubernetes/pki/etcd/etcd-healthcheck-client.key 4096
openssl req -new -key /etc/kubernetes/pki/etcd/etcd-healthcheck-client.key -subj "/CN=kube-etcd-healthcheck-client" -out etcd-healthcheck-client.csr
openssl x509 -req -in etcd-healthcheck-client.csr -CA /etc/kubernetes/pki/etcd/ca.crt -CAkey /etc/kubernetes/pki/etcd/ca.key -CAcreateserial -out /etc/kubernetes/pki/etcd/etcd-healthcheck-client.crt -days ${CP_CERT_EXPIRATION_PERIOD}
rm -rf etcd-healthcheck-client.csr

# Create kubeconfig files
# admin
openssl genrsa -out admin.key 4096
openssl req -new -key admin.key -subj "/CN=kubernetes-admin/O=system:masters" -out admin.csr
openssl x509 -req -in admin.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out admin.crt -days 1000

KUBECONFIG=/etc/kubernetes/admin.conf kubectl config set-cluster "${CLUSTER_NAME}" --server="https://127.0.0.1:${APISERVER_LB_PORT}" --certificate-authority /etc/kubernetes/pki/ca.crt --embed-certs
KUBECONFIG=/etc/kubernetes/admin.conf kubectl config set-credentials kubernetes-admin --client-key admin.key --client-certificate admin.crt --embed-certs
KUBECONFIG=/etc/kubernetes/admin.conf kubectl config set-context default --cluster "${CLUSTER_NAME}"  --user kubernetes-admin
KUBECONFIG=/etc/kubernetes/admin.conf kubectl config use-context default

rm -rf admin.key admin.crt admin.csr

# controller-manager
openssl genrsa -out controller-manager.key 4096
openssl req -new -key controller-manager.key -subj "/CN=system:kube-controller-manager" -out controller-manager.csr
openssl x509 -req -in controller-manager.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out controller-manager.crt -days ${CP_CERT_EXPIRATION_PERIOD}

KUBECONFIG=/etc/kubernetes/controller-manager.conf kubectl config set-cluster "${CLUSTER_NAME}" --server="https://127.0.0.1:${APISERVER_LB_PORT}" --certificate-authority /etc/kubernetes/pki/ca.crt --embed-certs
KUBECONFIG=/etc/kubernetes/controller-manager.conf kubectl config set-credentials default-controller-manager --client-key controller-manager.key --client-certificate controller-manager.crt --embed-certs
KUBECONFIG=/etc/kubernetes/controller-manager.conf kubectl config set-context default --cluster "${CLUSTER_NAME}" --user default-controller-manager
KUBECONFIG=/etc/kubernetes/controller-manager.conf kubectl config use-context default

rm -rf controller-manager.key controller-manager.crt controller-manager.csr

# scheduler
openssl genrsa -out scheduler.key 4096
openssl req -new -key scheduler.key -subj "/CN=system:kube-scheduler" -out scheduler.csr
openssl x509 -req -in scheduler.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out scheduler.crt -days ${CP_CERT_EXPIRATION_PERIOD}

KUBECONFIG=/etc/kubernetes/scheduler.conf kubectl config set-cluster "${CLUSTER_NAME}" --server="https://127.0.0.1:${APISERVER_LB_PORT}" --certificate-authority /etc/kubernetes/pki/ca.crt --embed-certs
KUBECONFIG=/etc/kubernetes/scheduler.conf kubectl config set-credentials default-scheduler --client-key scheduler.key --client-certificate scheduler.crt --embed-certs
KUBECONFIG=/etc/kubernetes/scheduler.conf kubectl config set-context default --cluster "${CLUSTER_NAME}" --user default-scheduler
KUBECONFIG=/etc/kubernetes/scheduler.conf kubectl config use-context default

rm -rf scheduler.key scheduler.crt scheduler.csr

# konnectivity-server
openssl genrsa -out konnectivity-server.key 4096
openssl req -new -key konnectivity-server.key -subj "/CN=system:konnectivity-server" -out konnectivity-server.csr
openssl x509 -req -in konnectivity-server.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out konnectivity-server.crt -days 1000

KUBECONFIG=/etc/kubernetes/konnectivity-server.conf kubectl config set-cluster "${CLUSTER_NAME}" --server="https://127.0.0.1:${APISERVER_LB_PORT}" --certificate-authority /etc/kubernetes/pki/ca.crt --embed-certs
KUBECONFIG=/etc/kubernetes/konnectivity-server.conf kubectl config set-credentials system:konnectivity-server --client-key konnectivity-server.key --client-certificate konnectivity-server.crt --embed-certs
KUBECONFIG=/etc/kubernetes/konnectivity-server.conf kubectl config set-context "system:konnectivity-server@${CLUSTER_NAME}" --cluster "${CLUSTER_NAME}" --user system:konnectivity-server
KUBECONFIG=/etc/kubernetes/konnectivity-server.conf kubectl config use-context "system:konnectivity-server@${CLUSTER_NAME}"

rm -rf konnectivity-server.key konnectivity-server.crt konnectivity-server.csr


# front-proxy-client/aggregation layer
openssl genrsa -out /etc/kubernetes/pki/front-proxy-client.key 4096
openssl req -new -key /etc/kubernetes/pki/front-proxy-client.key -subj "/CN=front-proxy-client" -out /etc/kubernetes/pki/front-proxy-client.csr
openssl x509 -req -in /etc/kubernetes/pki/front-proxy-client.csr -CA /etc/kubernetes/pki/front-proxy-ca.crt -CAkey /etc/kubernetes/pki/front-proxy-ca.key -CAcreateserial -out /etc/kubernetes/pki/front-proxy-client.crt -days ${CP_CERT_EXPIRATION_PERIOD}
rm -rf /etc/kubernetes/pki/front-proxy-client.csr

# Setup fluentbit
setup_fluentbit

# Create systemd services
while [ ! -f /etc/systemd/system/kube-scheduler.service ]; do
cat > /etc/systemd/system/kube-scheduler.service <<END
[Unit]
Description=Kubernetes Scheduler
Documentation=https://github.com/kubernetes/kubernetes
[Service]
ExecStart=/usr/local/bin/kube-scheduler --kubeconfig=/etc/kubernetes/scheduler.conf --leader-elect=true --profiling=false --tls-cipher-suites=TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 --bind-address 127.0.0.1 --v=2
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/
ReadOnlyDirectories=/etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
done

while [ ! -f /etc/systemd/system/kube-controller-manager.service ]; do
cat > /etc/systemd/system/kube-controller-manager.service <<END
[Unit]
Description=Kubernetes Controller Manager
Documentation=https://github.com/kubernetes/kubernetes
[Service]
ExecStart=/usr/local/bin/kube-controller-manager --allocate-node-cidrs=true --authentication-kubeconfig=/etc/kubernetes/controller-manager.conf --authorization-kubeconfig=/etc/kubernetes/controller-manager.conf --bind-address=127.0.0.1 --client-ca-file=/etc/kubernetes/pki/ca.crt --cluster-cidr="${CLUSTER_CIDR}" --cluster-name=kubernetes --cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt --cluster-signing-key-file=/etc/kubernetes/pki/ca.key --kubeconfig=/etc/kubernetes/controller-manager.conf --leader-elect=true --node-cidr-mask-size=24 --profiling=false --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt --root-ca-file=/etc/kubernetes/pki/ca.crt --service-account-private-key-file=/etc/kubernetes/pki/sa.key --service-cluster-ip-range="${SERVICE_CIDR}" --terminated-pod-gc-threshold=10 --tls-cipher-suites=TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 --use-service-account-credentials=true --controllers=*,tokencleaner --v=2
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/
ReadOnlyDirectories=/etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
done

while [ ! -f /etc/systemd/system/kube-apiserver.service ]; do
cat > /etc/systemd/system/kube-apiserver.service <<END
[Unit]
Description=Kubernetes API Server
Documentation=https://github.com/kubernetes/kubernetes
[Service]
ExecStart=/usr/local/bin/kube-apiserver --advertise-address ${APISERVER_LB} --allow-privileged=true --anonymous-auth=false --api-audiences=https://kubernetes.default.svc,system:konnectivity-server --audit-log-maxage=30 --audit-log-maxbackup=10 --audit-log-maxsize=100 --audit-log-path=/var/log/audit.log --authorization-mode=Node,RBAC --bind-address=0.0.0.0 --client-ca-file=/etc/kubernetes/pki/ca.crt --egress-selector-config-file=/etc/kubernetes/konnectivity/egress-selector-configuration.yaml --enable-admission-plugins=CertificateApproval,CertificateSigning,CertificateSubjectRestriction,DefaultIngressClass,DefaultStorageClass,DefaultTolerationSeconds,LimitRanger,MutatingAdmissionWebhook,NamespaceLifecycle,NodeRestriction,PersistentVolumeClaimResize,PodSecurity,Priority,ResourceQuota,RuntimeClass,ServiceAccount,StorageObjectInUseProtection,TaintNodesByCondition,ValidatingAdmissionPolicy,ValidatingAdmissionWebhook --enable-aggregator-routing=true --enable-bootstrap-token-auth=true --encryption-provider-config=/etc/kubernetes/enc/enc.yaml --etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt --etcd-certfile=/etc/kubernetes/pki/etcd/apiserver-etcd-client.crt --etcd-keyfile=/etc/kubernetes/pki/etcd/apiserver-etcd-client.key --etcd-servers="https://${ETCD_LB}:${ETCD_LB_PORT}" --event-ttl=1h --kubelet-certificate-authority=/etc/kubernetes/pki/ca.crt --kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt --kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key --kubelet-preferred-address-types="InternalIP,ExternalIP,Hostname" --profiling=false --proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt --proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port ${APISERVER_LB_PORT} --service-account-issuer=https://kubernetes.default.svc --service-account-key-file=/etc/kubernetes/pki/sa.pub --service-account-lookup=true --service-account-signing-key-file=/etc/kubernetes/pki/sa.key --service-cluster-ip-range="${SERVICE_CIDR}" --service-node-port-range=30000-32767 --tls-cert-file=/etc/kubernetes/pki/apiserver.crt --tls-cipher-suites=TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 --tls-private-key-file=/etc/kubernetes/pki/apiserver.key --v=2
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/
ReadOnlyDirectories=/etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
done

while [ ! -f /etc/systemd/system/etcd.service ]; do
cat > /etc/systemd/system/etcd.service <<END
[Unit]
Description=etcd
Documentation=https://github.com/coreos
[Service]
ExecStart=/usr/local/bin/etcd --advertise-client-urls https://${NODE_IP}:2379 --cert-file=/etc/kubernetes/pki/etcd/server.crt --cipher-suites TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 --client-cert-auth --data-dir=/var/lib/etcd --initial-advertise-peer-urls https://${NODE_IP}:2380 --initial-cluster "${ETCD_INITIAL_CLUSTER}" --initial-cluster-state ${ETCD_CLUSTER_STATE} --initial-cluster-token etcd-cluster-0 --key-file=/etc/kubernetes/pki/etcd/server.key --listen-client-urls https://${NODE_IP}:2379,https://127.0.0.1:2379 --listen-peer-urls https://${NODE_IP}:2380 --name ${NODE_IP} --peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt --peer-client-cert-auth --peer-key-file=/etc/kubernetes/pki/etcd/peer.key --peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt --trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/ /var/lib/etcd/
ReadOnlyDirectories=/etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
done

while [ ! -f /etc/systemd/system/konnectivity-server.service ]; do
cat > /etc/systemd/system/konnectivity-server.service <<END
[Unit]
Description=Konnectivity Server
Documentation=https://kubernetes.io/docs/tasks/extend-kubernetes/setup-konnectivity/
[Service]
ExecStart=/usr/local/bin/proxy-server --admin-port=8133 --agent-namespace=kube-system --agent-port=8132 --agent-service-account=konnectivity-agent --authentication-audience=system:konnectivity-server --cipher-suites=TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 --cluster-cert=/etc/kubernetes/pki/apiserver.crt --cluster-key=/etc/kubernetes/pki/apiserver.key --delete-existing-uds-file=true --enable-profiling=false --health-port=8134 --kubeconfig=/etc/kubernetes/konnectivity-server.conf --logtostderr=true --mode=grpc --proxy-strategies=destHost,default --server-count=3 --server-port=0 --uds-name=/etc/kubernetes/konnectivity/konnectivity-server.socket
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/ /etc/kubernetes/konnectivity/
ReadOnlyDirectories=/etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
done

# Add etcd member
if [[ "${ETCD_CLUSTER_STATE}" == "existing" ]]; then
  echo "adding member to etcd cluster"
  while ! etcdctl --endpoints=${ETCD_LB}:${ETCD_LB_PORT} --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key member add "${NODE_IP}" --peer-urls="https://${NODE_IP}:2380"; do
    echo "adding member to etcd failed. Retrying"
    sleep 5
  done
fi

# Start services
systemctl daemon-reload
systemctl enable etcd.service
systemctl start etcd.service

# Check etcd state
while ! etcdctl --write-out=json --endpoints=127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key  endpoint health | grep '"health":true'; do
  echo "waiting for etcd"
  sleep 5
done

# Start remaining services
systemctl enable kube-apiserver.service kube-scheduler.service kube-controller-manager.service konnectivity-server.service
systemctl start kube-apiserver.service kube-scheduler.service kube-controller-manager.service konnectivity-server.service

# Check apiserver state
while ! kubectl --kubeconfig /etc/kubernetes/admin.conf get --raw='/readyz'; do
  echo "waiting for apiserver"
  sleep 5
done

# Allow apiserver communication to kubelet to get metrics, logs, stats...
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<END
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:kube-apiserver-to-kubelet
rules:
  - apiGroups:
      - ""
    resources:
      - nodes/proxy
      - nodes/stats
      - nodes/log
      - nodes/spec
      - nodes/metrics
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:kube-apiserver
  namespace: ""
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kube-apiserver-to-kubelet
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: kubernetes
END

# Enable TLS bootstrapping
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<END
# Enable bootstrapping nodes to create CSR
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: create-csrs-for-bootstrapping
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: system:node-bootstrapper
  apiGroup: rbac.authorization.k8s.io
---
# Approve all CSRs for the group "system:bootstrappers"
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: auto-approve-csrs-for-group
subjects:
- kind: Group
  name: system:bootstrappers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: system:certificates.k8s.io:certificatesigningrequests:nodeclient
  apiGroup: rbac.authorization.k8s.io
---
# Approve renewal CSRs for the group "system:nodes"
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: auto-approve-renewals-for-nodes
subjects:
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: system:certificates.k8s.io:certificatesigningrequests:selfnodeclient
  apiGroup: rbac.authorization.k8s.io
END

# This is to allow nodegroup controller to manage nodes.
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<END
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    controller: nodegroup
    service: iks
  name: iks:nodegroup-controller
rules:
# Get apiserver health status.
- nonResourceURLs:
  - "/healthz"
  - "/healthz/*"
  - "/livez"
  - "/livez/*"
  - "/readyz"
  - "/readyz/*"
  verbs:
  - get
# Create bootstrap token secret and weka storage secret.
- apiGroups:
  - ""
  resources:
  - secrets
  - namespaces
  verbs:
  - create
  - get
# Manage nodes.
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - pods/eviction
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - "apps"
  resources:
  - "*"
  verbs:
  - get
# Approve kubelet serving certificates.
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - get
  - list
  - create
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests/approval
  verbs:
  - update
- apiGroups:
  - certificates.k8s.io
  resources:
  - signers
  verbs:
  - approve
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    controller: nodegroup
    service: iks
  name: iks:nodegroup-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: iks:nodegroup-controller
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: iks:nodegroup-controller
END

# This is to allow addon controller to manage addons.
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<END
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: iks:addon-controller
rules:
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - kube-proxy
  - tigera-operator
  - konnectivity-agent
  resources:
  - serviceaccounts
  verbs:
  - "*"
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  verbs:
  - create
- apiGroups:
  - rbac.authorization.k8s.io
  resourceNames:
  - system:kube-proxy
  - system:coredns
  - tigera-operator
  - system:konnectivity-server
  resources:
  - clusterrolebindings
  verbs:
  - "*"
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  verbs:
  - create
- apiGroups:
  - rbac.authorization.k8s.io
  resourceNames:
  - system:kube-proxy
  - system:coredns
  - tigera-operator
  resources:
  - clusterroles
  verbs:
  - "*"
- apiGroups:
  - authentication.k8s.io #required for konnectivity agents clusterrole
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io #required for konnectivity agents clusterrole
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - "" #required for kube-proxy clusterrole
  resources:
  - endpoints
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "" #required for kube-proxy clusterrole
  - "events.k8s.io"
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - "discovery.k8s.io" #required for kube-proxy clusterrole
  resources:
  - endpointslices
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - kube-proxy
  - coredns
  resources:
  - configmaps
  verbs:
  - "*"
- apiGroups:
  - "apps"
  resources:
  - daemonsets
  verbs:
  - create
- apiGroups:
  - "apps"
  resourceNames:
  - kube-proxy
  - konnectivity-agent
  resources:
  - daemonsets
  verbs:
  - "*"
- apiGroups:
  - "apps"
  resources:
  - deployments
  verbs:
  - create
- apiGroups:
  - "apps"
  resourceNames:
  - tigera-operator
  - coredns
  resources:
  - deployments
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - create
  - list
  - watch
- apiGroups:
  - ""
  resourceNames:
  - kube-dns
  resources:
  - services
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - tigera-operator
  resources:
  - namespaces
  verbs:
  - "*"
- apiGroups:
  - "apiextensions.k8s.io"
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - update
- apiGroups:
  - "crd.projectcalico.org"
  resources:
  - bgpconfigurations
  - bgpfilters
  - bgppeers
  - blockaffinities
  - caliconodestatuses
  - clusterinformations
  - felixconfigurations
  - globalnetworkpolicies
  - globalnetworksets
  - hostendpoints
  - caliconodestatuses
  - ipamblocks
  - ipamconfigs
  - ipamhandles
  - ippools
  - ipreservations
  - kubecontrollersconfigurations
  - networkpolicies
  - networksets
  verbs:
  - "*"
- apiGroups:
  - "operator.tigera.io"
  resources:
  - "*"
  verbs:
  - "*"
# The following permissions are required for the calico tigera operator clusterrole
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - resourcequotas
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - update
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - update
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - get
  - list
  - delete
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  - namespaces
  - secrets
  - serviceaccounts
  verbs:
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - podtemplates
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - "apiregistration.k8s.io"
  resources:
  - apiservices
  verbs:
  - create
  - list
  - update
  - watch
- apiGroups:
  - "apps"
  resources:
  - daemonsets
  - deployments
  verbs:
  - get
  - list
  - patch
  - update
  - delete
  - watch
- apiGroups:
  - "apps"
  resourceNames:
  - tigera-operator
  resources:
  - deployments/finalizers
  verbs:
  - update
- apiGroups:
  - "apps"
  resources:
  - statefulsets
  verbs:
  - create
  - get
  - list
  - patch
  - update
  - delete
  - watch
- apiGroups:
  - "certificates.k8s.io"
  resources:
  - certificatesigningrequests
  verbs:
  - list
  - watch
- apiGroups:
  - "coordination.k8s.io"
  resources:
  - leases
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - "networking.k8s.io"
  resources:
  - networkpolicies
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - "policy"
  resources:
  - poddisruptionbudgets
  - podsecuritypolicies
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - "policy"
  resourceNames:
  - tigera-operator
  resources:
  - podsecuritypolicies
  verbs:
  - use
- apiGroups:
  - "rbac.authorization.k8s.io"
  resources:
  - clusterrolebindings
  - clusterroles
  verbs:
  - get
  - list
  - update
  - delete
  - watch
  - bind
  - escalate
- apiGroups:
  - "rbac.authorization.k8s.io"
  resources:
  - rolebindings
  - roles
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
  - bind
  - escalate
- apiGroups:
  - "scheduling.k8s.io"
  resources:
  - priorityclasses
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
- apiGroups:
  - "storage.k8s.io"
  resources:
  - csidrivers
  - storageclasses
  verbs:
  - create
  - get
  - list
  - update
  - delete
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: iks:addon-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: iks:addon-controller
subjects:
- kind: Group
  name: iks:addon-controller
END

# Clusterrole binding for default readonly user.
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<END
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-read-only-access
subjects:
- kind: User
  name: readonly
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: view
  apiGroup: rbac.authorization.k8s.io
END

# Replace secrets so that we use new etcd encryption configuration if rotated.
kubectl --kubeconfig /etc/kubernetes/admin.conf get secrets --all-namespaces -o json | kubectl --kubeconfig /etc/kubernetes/admin.conf replace -f -

# Delete sensitive files from node.
rm -f /etc/kubernetes/admin.conf

# Configure monitoring.
if [ -d /etc/prometheus ] && ([[ "${SYSTEM_METRICS_ENABLED:-false}" == true ]] || [[ "${END_USER_METRICS_ENABLED:-false}" == true ]]); then  echo "Configuring Prometheus"
  while [ ! -f /etc/systemd/system/cadvisor.service ]; do
cat > /etc/systemd/system/cadvisor.service <<END
[Unit]
Description=cAdvisor Service
Documentation=https://github.com/google/cadvisor

[Service]
ExecStart=/usr/bin/cadvisor
Restart=on-failure
RestartSec=5

ProtectSystem=strict
ReadWriteDirectories=/var/log/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END
  done

  # TODO: Remove this once added to IMI.
  if [ ! -f /usr/bin/cadvisor ]; then
    cAdvisorVersion="0.47.2"
    curl -L "https://github.com/google/cadvisor/releases/download/v${cAdvisorVersion}/cadvisor-v${cAdvisorVersion}-linux-amd64" --output /usr/bin/cadvisor
  fi
  systemctl enable cadvisor.service
  systemctl restart cadvisor.service

cat > /etc/prometheus/prometheus.yml <<END
global:
  scrape_interval: 15s
  external_labels:
    hostname: "${HOSTNAME}"
    clustername: "${CLUSTER_NAME}"
    clustertype: iks
    cloudaccount: "${CLOUDACCOUNT}"
    region:  "${REGION}"

scrape_configs:
  - job_name: node-exporter
    static_configs:
    - targets: ['localhost:9200']
  - job_name: cadvisor
    static_configs:
    - targets: ['localhost:8080']
  - job_name: etcd
    scheme: https
    tls_config:
      insecure_skip_verify: true
      ca_file: /etc/kubernetes/pki/etcd/ca.crt
      cert_file: /etc/kubernetes/pki/etcd/server.crt
      key_file: /etc/kubernetes/pki/etcd/server.key
    static_configs:
    - targets: ['localhost:2379']
  - job_name: kube-apiserver
    scheme: https
    tls_config:
      insecure_skip_verify: true
      ca_file: /etc/kubernetes/pki/ca.crt
      cert_file: /etc/kubernetes/pki/apiserver-kubelet-client.crt
      key_file: /etc/kubernetes/pki/apiserver-kubelet-client.key
    static_configs:
    - targets: ['localhost:443']
  - job_name: kube-controller
    scheme: https
    tls_config:
      insecure_skip_verify: true
      cert_file: /etc/kubernetes/pki/apiserver-kubelet-client.crt
      key_file: /etc/kubernetes/pki/apiserver-kubelet-client.key
    static_configs:
    - targets: ['localhost:10257']
  - job_name: kube-scheduler
    scheme: https
    tls_config:
      insecure_skip_verify: true
      cert_file: /etc/kubernetes/pki/apiserver-kubelet-client.crt
      key_file: /etc/kubernetes/pki/apiserver-kubelet-client.key
    static_configs:
    - targets: ['localhost:10259']
  - job_name: 'prometheus'
    scrape_interval: 5s
    scrape_timeout: 5s
    static_configs:
      - targets: ['localhost:9090']

remote_write:
END

if [[ "${SYSTEM_METRICS_ENABLED:-false}" == true ]]; then
  SYSTEM_METRICS_PROMETHEUS_URL=${SYSTEM_METRICS_PROMETHEUS_URL:-}
  if [ "$SYSTEM_METRICS_PROMETHEUS_URL" != "" ]; then
    echo "Adding system metrics prometheus config"
    cat >> /etc/prometheus/prometheus.yml <<END
- url: "${SYSTEM_METRICS_PROMETHEUS_URL}"
  basic_auth:
      username: "${SYSTEM_METRICS_PROMETHEUS_USERNAME}"
      password: "${SYSTEM_METRICS_PROMETHEUS_PASSWORD}"
END
  fi
fi

if [[ "${END_USER_METRICS_ENABLED:-false}" == true ]]; then
  END_USER_METRICS_PROMETHEUS_URL=${END_USER_METRICS_PROMETHEUS_URL:-}
  if [ "$END_USER_METRICS_PROMETHEUS_URL" != "" ]; then
    echo "Adding end user metrics prometheus url"
    cat >> /etc/prometheus/prometheus.yml <<END
- url: "${END_USER_METRICS_PROMETHEUS_URL}"
END
  fi

# Ensure END_USER_METRICS_PROMETHEUS_BEARER_TOKEN is defined
END_USER_METRICS_PROMETHEUS_BEARER_TOKEN=${END_USER_METRICS_PROMETHEUS_BEARER_TOKEN:-}

if [ "$END_USER_METRICS_PROMETHEUS_BEARER_TOKEN" != "" ]; then
  echo "Adding end user metrics prometheus bearer token"
  cat >> /etc/prometheus/prometheus.yml <<END
  bearer_token: "${END_USER_METRICS_PROMETHEUS_BEARER_TOKEN}"
END
fi

fi

cat > /etc/systemd/system/prometheus.service <<END
[Unit]
Description=Prometheus
ConditionFileIsExecutable=/usr/local/bin/prometheus
After=network-online.target
Wants=network-online.target

[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart=/usr/local/bin/prometheus --enable-feature=agent --config.file=/etc/prometheus/prometheus.yml
RestartSec=120
Delegate=yes
KillMode=process
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
LimitNOFILE=999999
Restart=always

ProtectSystem=strict
ReadWriteDirectories=/var/log/ /data-agent/
ReadOnlyDirectories=/etc/prometheus/ /etc/kubernetes/
LockPersonality=yes
PrivateTmp=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelTunables=yes
NoNewPrivileges=yes
PrivateDevices=yes
DevicePolicy=closed
ProtectControlGroups=yes
ProtectKernelModules=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ProtectClock=yes
CapabilityBoundingSet=~CAP_LINUX_IMMUTABLE CAP_IPC_LOCK CAP_SYS_CHROOT CAP_BLOCK_SUSPEND CAP_LEASE
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_BOOT CAP_SYS_PACCT CAP_SYS_PTRACE CAP_SYS_RAWIO CAP_SYS_TIME CAP_SYS_TTY_CONFIG
CapabilityBoundingSet=~CAP_WAKE_ALARM  CAP_MAC_ADMIN CAP_MAC_OVERRIDE
CapabilityBoundingSet=~CAP_SETUID CAP_SETGID CAP_SETPCAP
CapabilityBoundingSet=~CAP_CHOWN CAP_FSETID CAP_SETFCAP

[Install]
WantedBy=multi-user.target
END

  systemctl enable prometheus.service
  systemctl restart prometheus.service
fi

#Change the logratation to rorate hourly and rotate logs if they are bigger than 100M
if [[ -f /usr/lib/systemd/system/logrotate.timer && -f /etc/logrotate.d/rsyslog ]]; then
  cp /usr/lib/systemd/system/logrotate.timer /usr/lib/systemd/system/logrotate.timer.bak
  sed -i "s/rotate .*/rotate 10/g" /etc/logrotate.d/rsyslog; sed -i "s/weekly/hourly\n        size 100M/g" /etc/logrotate.d/rsyslog
  sed -i "s/OnCalendar=daily/OnCalendar=hourly/g" /usr/lib/systemd/system/logrotate.timer; sed -i "s/Daily/Hourly/g" /usr/lib/systemd/system/logrotate.timer
  systemctl daemon-reload
  systemctl reenable logrotate.timer
  systemctl restart rsyslog.service
  systemctl restart logrotate.timer
fi

log_message "The IDC Kubernetes bootstrap script has completed."