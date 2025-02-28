#!/usr/bin/env bash
# Example: ./upload.sh 10.10.10.1 file.txt

set -e

if [ -z $1 ]; then
  echo "Please specify MinIO URL or IP address"
  exit 1
fi

if [ -z $2 ]; then
  echo "Please specify file path"
  exit 1
fi


URL=$1
FILE=$2

BUCKET=idc-baremetal

RESOURCE="/${BUCKET}/${FILE}"
CONTENT_TYPE="application/octet-stream"
DATE=`date -R`
_signature="PUT\n\n${CONTENT_TYPE}\n${DATE}\n${RESOURCE}"
SIGNATURE=`echo -en ${_signature} | openssl sha1 -hmac minio123 -binary | base64`

curl -vk -X PUT -T "${FILE}" \
          -H "Host: ${URL}:30443" \
          -H "Date: ${DATE}" \
          -H "Content-Type: ${CONTENT_TYPE}" \
          -H "Authorization: AWS minio:${SIGNATURE}" \
          https://${URL}:30443${RESOURCE}