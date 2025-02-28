#!/usr/bin/env bash
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/objects/buckets/id/${ID}/lifecyclerule --data-binary @- \
| jq .
{
    "metadata": {
        "ruleName": "lfrule-2"
    },
    "spec": {
        "prefix": "/tmp",
        "expireDays":10,
        "noncurrentExpireDays": 5,
        "deleteMarker": false
    }
}
EOF
