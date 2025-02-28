#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -x
echo $$ > /tmp/validation/validation.pid
# printenv # this is for debugging

cleanup() {
    # Update the exit code only if the file does not exist.
    if [ ! -f /tmp/validation/validation_exitcode ]
    then
        echo "Validation abruptly ended, updating the exit code"
        echo 1 > /tmp/validation/validation_exitcode
    fi

    # archive the logs and upload it to s3 bucket
    echo "Archiving logs and pushing it to s3 bucket"
    mkdir -p /tmp/validation/logs
    cd /tmp/validation || exit
    cp validation.out logs/
    tar -czvf validation_logs.tar.gz logs/*
    file=/tmp/validation/validation_logs.tar.gz
    resource="/${bucket}$uploadPath/validation_logs.tar.gz"
    contentType="application/x-compressed-tar"
    dateValue=$(date -R)
    stringToSign="PUT\n\n${contentType}\n${dateValue}\n${resource}"

    signature=`echo -en ${stringToSign} | openssl sha1 -hmac ${s3Secret} -binary | base64`
    curl -X PUT -T "${file}" \
    -H "Host: ${bucket}.s3.amazonaws.com" \
    -H "Date: ${dateValue}" \
    -H "Content-Type: ${contentType}" \
    -H "Authorization: AWS ${s3Key}:${signature}" \
    https://${bucket}.s3.amazonaws.com$uploadPath/validation_logs.tar.gz
}
trap "cleanup" EXIT
echo "validation task archive url is $1"
# extract the tar.gz
tar -xzvf /tmp/validation/validation.tar.gz --directory /tmp/validation
# capturing the env into the validation_result.meta file. This file can also store test events in the form key=value
printenv | awk '!/s3Secret/ && !/s3Key/ && !/huggingFaceToken/' > /tmp/validation_result.meta
echo "testing"
# Invoke the start script from the task archive with a timeout of 4 hours
# Incase of timeout the exit code will be 143
timeout --preserve-status 4h /tmp/validation/start.sh 2>&1 | tee -a ~/validation_bkup.out

# (exit 1) #Use this to simulate a failed test
# Capture the exit code of the test, this will ensure the reconciler thread does not wait for the command to return.
echo $? > /tmp/validation/validation_exitcode
echo "testing completed"
