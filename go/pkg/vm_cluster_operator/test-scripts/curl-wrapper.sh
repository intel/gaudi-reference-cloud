#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Rancher's system-agent-install.sh does not quote the '*' argument to curl correctly,
# so --noproxy '*' as used by the script does not work. Fix that here by undoing the 
# incorrect quoting.
args=()
for arg in "$@"; do
if [[ "${arg}" == "'*'" ]]; then
        args+=('*')
else
        args+=("${arg}")
fi
done
/usr/bin/curl "${args[@]}"
