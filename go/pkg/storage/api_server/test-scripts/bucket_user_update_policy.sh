#!/usr/bin/env bash
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/objects/users/id/${ID}/policy --data-binary @- \
| jq .
{
    "metadata": {
        "name": "testbuser19"
    },
    "spec": [{
        "bucketId": "986258171890-testb20",
        "permission":["ReadBucket"],
        "actions": ["GetBucketLocation", "GetBucketPolicy", "ListBucket", "ListBucketMultipartUploads", "ListMultipartUploadParts", "GetBucketTagging"]
    }]
}
EOF
