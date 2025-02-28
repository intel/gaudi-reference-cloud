#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

CLOUDACCOUNTNAME=john.doe@intel.com
# Retrieve the token
TOKEN=$(curl -s "http://dev.oidc.cloud.intel.com.kind.local/token?email=${CLOUDACCOUNTNAME}")

#TOKEN=...
#URL=https://dev28-compute-us-dev28-1-api.cloud.intel.com/v1/infaas/generatestream
#URL=https://internal-placeholder.com/v1/maas/generatestream
URL=https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local/v1/maas/generatestream
PROMPT='Tell me a random joke about software engineers'
#PROMPT='What is INTEL Developer Cloud?'
#PROMPT='How do you write a loop in Go?'
#PROMPT="How do you hack an ATM?"

#--socks5-hostname 127.0.0.1:1080 \
curl -v --no-buffer --location $URL \
--header 'Content-Type: application/json' \
--insecure \
--header 'Authorization: Bearer '"${TOKEN}" \
--max-time 30 \
--http2 \
--data @- << EOF
{
    "model": "model1",
    "request": {
        "prompt": "${PROMPT}",
        "params": {
            "max_new_tokens": 150
        }
    },
    "cloudAccountId": "513861623936"
}
EOF