#!/usr/bin/env bash
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
HOSTS_CONTENTS="$(cat ${SCRIPT_DIR}/hosts | tr '\n' ' ')"
grep --fixed-strings --line-regexp "${HOSTS_CONTENTS}" /etc/hosts && exit
echo "${HOSTS_CONTENTS}" | sudo tee -a /etc/hosts
