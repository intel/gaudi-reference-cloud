#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -e

[ "$1" = "-v" ] && verbose=-v

URL_PREFIX=${URL_PREFIX:-https://localhost}

run_curl() {
    path=$1
    post=$2

    url=${URL_PREFIX}$path
    cmd=(curl -sk)
    [ -n "${verbose}" ] && cmd+=(${verbose})
    [ -n "${post}" ] && cmd+=("-d" \'${post}\')
    cmd+=(${url})

    echo "${cmd[*]}"
    eval "${cmd[*]}" 2>&1
    echo
}

run_curl \
    /v1/billing/cloudaccounts \
    '{"name":"rest@example.com","owner":"rest@example.com","type":"ACCOUNT_TYPE_STANDARD"}'
id=$(curl -sk ${URL_PREFIX}/v1/cloudaccounts/name/rest@example.com | jq -r .id)

run_curl /v1/billing/accounts "{\"cloudAccountId\":\"${id}\"}"
run_curl "/v1/billing/options?cloudAccountId=${id}"
run_curl "/v1/billing/rates?cloudAccountId=${id}"

run_curl "/v1/billing/credit?cloudAccountId=${id}"
run_curl "/v1/billing/credit/unapplied?cloudAccountId=${id}"
run_curl /v1/billing/credit "{\"cloudAccountId\":\"${id}\"}"

run_curl /v1/billing/coupons
run_curl /v1/billing/coupons {}
run_curl /v1/billing/coupons/redeem "{\"cloudAccountId\":\"${id}\"}"

run_curl "/v1/billing/invoices?cloudAccountId=${id}"
run_curl "/v1/billing/invoices/detail?cloudAccountId=${id}"
run_curl "/v1/billing/invoices/statement?cloudAccountId=${id}"
