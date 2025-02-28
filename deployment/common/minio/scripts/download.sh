#!/usr/bin/env bash
# Example: ./download.sh 10.10.10.1 /file.txt /tmp/file.txt

set -e

if [ -z $1 ]; then
  echo "Please specify MinIO URL or IP address"
  exit 1
fi

if [ -z $2 ]; then
  echo "Please specify MinIO file path"
  exit 1
fi

if [ -z $3 ]; then
  echo "Please specify local file path"
  exit 1
fi

URL=$1
MINIO_PATH=$2
OUT_FILE=$3

BUCKET=idc-baremetal

DATE=$(date -R --utc)
CONTENT_TYPE='application/zstd'
SIG_STRING="GET\n\n${CONTENT_TYPE}\n${DATE}\n${MINIO_PATH}"
SIGNATURE=`echo -en ${SIG_STRING} | openssl sha1 -hmac minio123 -binary | base64`


curl -vk -o "${OUT_FILE}" \
    -H "Host: $URL:30443" \
    -H "Date: ${DATE}" \
    -H "Content-Type: ${CONTENT_TYPE}" \
    -H "Authorization: AWS minio:${SIGNATURE}" \
    https://${URL}:30443/${BUCKET}${MINIO_PATH}
