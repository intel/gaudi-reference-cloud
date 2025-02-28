#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -e -o pipefail

secrets_dir=../../../local/secrets
secrets_file_suffix=_db_user_password

usage() {
    [ -n "$1" ] && echo $* 1>&2
    svcs=$((cd $secrets_dir; ls -1 *${secrets_file_suffix}) | sed -e "s/$secrets_file_suffix$//")
    echo "usage: manage <service> <directory>"
    echo "  service is one of:" ${svcs}
}

svc=$1
dir=$2
if [ -z "$svc" -o -z "$dir" ]; then
    usage
    exit 1
fi

secrets_file=${secrets_dir}/${svc}${secrets_file_suffix}
if [ \! -f ${secrets_file} ]; then
    usage "${svc} is not a valid service"
    exit 1
fi

if [ \! -d ${dir} ]; then
    usage "${dir} is not a directory"
    exit 1
fi
    
export PGPASSWORD=$(cat ${secrets_file})
DBNAME=${DBNAME:-postgres}
export DATABASE_URL="postgres://dbuser:${PGPASSWORD}@127.0.0.1:5432/${DBNAME}?sslmode=disable"
migrate --source file://$dir --database ${DATABASE_URL} up
