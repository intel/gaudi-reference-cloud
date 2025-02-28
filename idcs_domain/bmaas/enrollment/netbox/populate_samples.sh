#!/usr/bin/env bash

NETBOX_API='http://localhost:30001/api'
HEADER_AUTH='Authorization: TOKEN 0123456789abcdef0123456789abcdef01234567'
HEADER_CONTENT_TYPE='Content-Type: application/json'

create() {
  local object=$1
  local data=$2

  echo "Creating ${object}..."

  (
    set -x
    curl -X POST \
      --url "${NETBOX_API}/${object}" \
      -H "${HEADER_AUTH}" \
      -H "${HEADER_CONTENT_TYPE}" \
      -d "$data"
  )

  echo
}

create extras/webhooks/ '{
  "content_types": [
    "dcim.device"
  ],
  "name": "bmass-enrollment",
  "type_create": true,
  "type_update": true,
  "payload_url": "http://dev.baremetal-enrollment-api.us-dev-1.api.cloud.intel.com.kind.local:80/api/v1/enroll",
  "enabled": true,
  "http_method": "POST",
  "http_content_type": "application/json",
  "additional_headers": "",
  "body_template": "{ \"id\": {{ data.id }}, \"name\": \"{{ data.name }}\"}",
  "secret": "",
  "ssl_verification": false,
  "ca_file_path": null
}'

create dcim/manufacturers/ '{
  "name": "manufacturer-1",
  "slug": "manufacturer-1"
}'

create dcim/sites/ '{
  "name": "site-1",
  "slug": "site-1"
}'

create dcim/device-roles/ '{
  "name": "device-role-1",
  "slug": "device-role-1",
  "color": "ffc107"
}'

create dcim/device-types/ '{
  "manufacturer": {
    "name": "manufacturer-1",
    "slug": "manufacturer-1"
  },
  "model": "device-type-1",
  "slug": "device-type-1"
}'

create dcim/devices/ '{
  "name": "device-1",
  "device_type": {
    "manufacturer": {
      "slug": "manufacturer-1"
    },
    "slug": "device-type-1"
  },
  "role": {
    "slug": "device-role-1"
  },
  "serial": "1",
  "site": {
    "slug": "site-1"
  }
}'
