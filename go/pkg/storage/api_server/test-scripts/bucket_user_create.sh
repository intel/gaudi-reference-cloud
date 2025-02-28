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
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/objects/users --data-binary @- \
| jq .
{
    "metadata": {
        "name": "${NAME}"
    },
    "spec": [{
        "bucketId": "714025373525-testb89",
        "permission":["ReadBucket", "WriteBucket", "DeleteBucket"],
        "actions": ["GetBucketLocation", "GetBucketPolicy", "ListBucket", "ListBucketMultipartUploads", "ListMultipartUploadParts", "GetBucketTagging"]
    }]
}
EOF
