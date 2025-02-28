#!/usr/bin/env bash
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

if [ "${NETBOX_API}" == "" ]; then
    REGION=us-dev-1
    NETBOX_FQDN=dev.netbox.us-dev-1.api.cloud.intel.com.kind.local
    NETBOX_PORT_FILE=${SCRIPT_DIR}/../../../local/${REGION}_host_port_443
    NETBOX_PORT=$(cat ${NETBOX_PORT_FILE})
    NETBOX_API=https://${NETBOX_FQDN}:${NETBOX_PORT}/api
fi

HEADER_AUTH='Authorization: TOKEN 0123456789abcdef0123456789abcdef01234567'
HEADER_CONTENT_TYPE='Content-Type: application/json'
BM_ENROLLMENT_APISERVICE_USERNAME=$(cat ${SECRETS_DIR}/BM_ENROLLMENT_APISERVICE_USERNAME)
BM_ENROLLMENT_APISERVICE_PASSWORD=$(cat ${SECRETS_DIR}/BM_ENROLLMENT_APISERVICE_PASSWORD)
BM_ENROLLMENT_APISERVICE_BASIC_AUTH=$(echo -n ${BM_ENROLLMENT_APISERVICE_USERNAME}:${BM_ENROLLMENT_APISERVICE_PASSWORD} | base64)
BM_ENROLLMENT_ING=$(kubectl get ing ${REGION}-baremetal-enrollment-api -n idcs-enrollment -o json | jq -r .spec.tls[0].hosts[0])
GUEST_HOST_DEPLOYMENTS=${GUEST_HOST_DEPLOYMENTS:-3}

create() {
  local object=$1
  local data=$2
  echo "Creating ${object}..."
  (
    set -x
    curl -X POST \
      --url "${NETBOX_API}/${object}" \
      -k \
      -H "${HEADER_AUTH}" \
      -H "${HEADER_CONTENT_TYPE}" \
      -d "$data"
  )
  echo
}

create extras/custom-field-choice-sets/ '{
  "name": "BM Enrollment",
  "description": "BMaaS enrollment status",
  "extra_choices": [
    ["enroll","Enroll"],
    ["enrolling","Enrollment in Progress"],
    ["enrollment-failed","Enrollment Failed"],
    ["enrolled","Enrollment Complete"],
    ["disenroll","Disenroll"],
    ["disenrolling","Disenrollment in Progress"],
    ["disenrollment-failed","Disenrollment Failed"],
    ["disenrolled","Disenrollment Complete"]
  ]
}'


create extras/custom-fields/ '{
  "name": "bm_enrollment_status",
  "label": "BM Enrollment Status",
  "content_types": [
      "dcim.device"
  ],
  "type": "select",
  "data_type": "string",
  "description": "Select \"Enroll\" or \"Disenroll\" to start the enrollment or disenrollment process",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-write",
  "default": null,
  "weight": 100,
  "choice_set": {
      "name": "BM Enrollment"
  },
  "group_name": "Enrollment"
}'

create extras/custom-fields/ '{
  "name": "bm_enrollment_comment",
  "label": "BM Enrollment Comment",
  "content_types": [
      "dcim.device"
  ],
  "type": "text",
  "data_type": "string",
  "description": "Additional comment from the enrollment service",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-write",
  "default": null,
  "weight": 100,
  "group_name": "Enrollment"
}'

create extras/custom-field-choice-sets/ '{
  "name": "BM Enrollment Namespace",
  "description": "The namespace in which the BM will be enrolled",
  "order_alphabetically": true,
  "extra_choices": [
    ["metal3-1","metal3-1"],
    ["metal3-2","metal3-2"],
    ["metal3-3","metal3-3"]
  ]
}'

create extras/custom-fields/ '{
  "name": "bm_enrollment_namespace",
  "label": "BM Enrollment Namespace",
  "content_types": [
      "dcim.device"
  ],
  "type": "select",
  "data_type": "string",
  "description": "The namespace in which the BM will be enrolled. Leave blank to have the namespace automatically assigned. The device must be re-enrolled for the namespace to take effect.",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-write",
  "default": null,
  "weight": 100,
  "choice_set": {
      "name": "BM Enrollment Namespace"
  },
  "group_name": "Enrollment"
}'

create extras/custom-fields/ '{
  "name": "bm_validation_status",
  "label": "BM Validation Status",
  "content_types": [
      "dcim.device"
  ],
  "type": "text",
  "data_type": "string",
  "description": "Last known validation status",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-only",
  "default": null,
  "weight": 100,
  "group_name": "Validation"
}'

create extras/custom-fields/ '{
  "name": "bm_validation_report_url",
  "label": "BM Validation Report Path",
  "content_types": [
      "dcim.device"
  ],
  "type": "text",
  "data_type": "string",
  "description": "Last known report Path",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-only",
  "default": null,
  "weight": 100,
  "group_name": "Validation"
}'

create extras/webhooks/ "$(cat <<EOF
{
  "content_types": [
    "dcim.device"
  ],
  "name": "bmaas-enrollment",
  "type_create": true,
  "type_update": true,
  "payload_url": "https://${BM_ENROLLMENT_ING}/api/v1/enroll",
  "enabled": true,
  "http_method": "POST",
  "http_content_type": "application/json",
  "additional_headers": "Authorization: Basic ${BM_ENROLLMENT_APISERVICE_BASIC_AUTH}",
  "body_template": "{\"id\": {{ data.id }}, \"name\": \"{{ data.name }}\", \"site\": \"{{ data.site.name }}\", \"rack\": \"{{ data.rack.name }}\", \"bm_enrollment_status\": \"{{ data.custom_fields.bm_enrollment_status }}\", \"cluster\": \"{{ data.cluster.name if data.cluster is not none else 'None' }}\"}",
  "secret": "",
  "conditions": {
    "and": [
      {"attr": "role.slug","value": "bmas"},
      {"attr": "status.value","value": "active"},
      {"or": [
        {"attr": "custom_fields.bm_enrollment_status","value": "enroll"},
        {"attr": "custom_fields.bm_enrollment_status","value": "disenroll"}]}]},
  "ssl_verification": false,
  "ca_file_path": null
}
EOF
)"

create dcim/manufacturers/ '{
  "name": "manufacturer-1",
  "slug": "manufacturer-1"
}'

create dcim/regions/ '{
  "name": "us-dev-1",
  "slug": "us-dev-1"
}'

create dcim/sites/ '{
  "name": "us-dev-1a",
  "slug": "us-dev-1a",
  "region": {
    "slug": "us-dev-1",
    "name": "us-dev-1"
  }
}'

create dcim/sites/ '{
  "name": "us-west-1b",
  "slug": "us-west-1b",
  "region": {
    "slug": "us-dev-1",
    "name": "us-dev-1"
  }
}'

create dcim/sites/ '{
  "name": "us-west-1c",
  "slug": "us-west-1c",
  "region": {
    "slug": "us-dev-1",
    "name": "us-dev-1"
  }
}'

create dcim/device-roles/ '{
  "name": "bmaas",
  "slug": "bmas",
  "color": "ffc107"
}'

create dcim/device-roles/ '{
  "name": "other",
  "slug": "other"
}'

create dcim/device-types/ '{
  "manufacturer": {
    "name": "manufacturer-1",
    "slug": "manufacturer-1"
  },
  "model": "device-type-1",
  "slug": "device-type-1"
}'

create dcim/racks/ '{
  "name": "myrack",
  "site": {
    "slug": "us-dev-1a"
  },
  "slug": "myrack",
  "color": "ffc107"
}'

create extras/custom-field-choice-sets/ '{
  "name": "BM Network Mode",
  "description": "Network configuration of the BM",
  "order_alphabetically": true,
  "extra_choices": [
    ["VVV","VVV"],
    ["VVX-standalone","VVX-standalone"],
    ["XBX","XBX"]
  ]
}'

create extras/custom-fields/ '{
  "name": "bm_network_mode",
  "label": "BM Network Mode",
  "content_types": [
      "virtualization.cluster"
  ],
  "type": "select",
  "data_type": "string",
  "description": "Network configuration of the cluster",
  "required": false,
  "search_weight": 1000,
  "filter_logic": "exact",
  "ui_visibility": "read-write",
  "default": null,
  "weight": 100,
  "choice_set": {
    "name": "BM Network Mode"
  }
}'

create virtualization/cluster-types/ '{
  "name": "2x-cluster",
  "slug": "2x-cluster"
}'

create virtualization/cluster-types/ '{
  "name": "4x-cluster",
  "slug": "4x-cluster"
}'

create virtualization/cluster-types/ '{
  "name": "8x-cluster",
  "slug": "8x-cluster"
}'

create virtualization/clusters/ '{
  "name": "1",
  "type": {
      "name": "2x-cluster",
      "slug": "2x-cluster"
  },
  "site": {
    "name": "us-dev-1a",
    "slug": "us-dev-1a"
  },
  "custom_fields": {
    "bm_network_mode": "VVV"
  }
}'

create virtualization/clusters/ '{
  "name": "2",
  "type": {
      "name": "2x-cluster",
      "slug": "2x-cluster"
  },
  "site": {
    "name": "us-dev-1a",
    "slug": "us-dev-1a"
  },
  "custom_fields": {
    "bm_network_mode": "VVV"
  }
}'

BMC_HOST=$(hostname -I | awk '{print $1;}')

for i in $(seq 1 "$GUEST_HOST_DEPLOYMENTS"); do
create dcim/devices/ "$(cat <<EOF 
{
  "name": "device-${i}",
  "device_type": {
    "manufacturer": {
      "slug": "manufacturer-1"
    },
    "slug": "device-type-1"
  },
  "rack": 1,
  "role": {
    "slug": "other"
  },
  "serial": "1",
  "site": {
    "slug": "us-dev-1a"
  }
}
EOF
)"
done

for i in $(seq 1 "$GUEST_HOST_DEPLOYMENTS"); do
device_id=$(curl -k --url "${NETBOX_API}/dcim/devices/?name=device-${i}" -H "${HEADER_AUTH}" | jq -r .results[].id)
create dcim/interfaces/ "$(cat <<EOF 
{
  "name": "BMC",
  "device": ${device_id},
  "enabled": true,
  "type": "other",
  "label": "http://${BMC_HOST}:$((8000+i))"
}
EOF
)"
done


# Support for addition Netbox data based on Environment

if [ -d ${IDC_ENV_DIR}/netbox ]; then
    pushd ${IDC_ENV_DIR}/netbox
    for order in $(ls -d */); do
        pushd ${order}
        for json in $(find . -type f \( -iname "*.json" \) ); do
            obj=$(dirname ${json})
            obj="${obj#./}/"
            echo "JSON file found: $json"
            echo create_from_file ${obj} ${json}
            create ${obj} "@${json}"
        done
        popd
    done
    popd
fi
