#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

REQUEST=$(cat << 'EOF'
{
    "model": "mistralai/Mistral-7B-Instruct-v0.1",
    "request": {
        "prompt": "what is AI?",
        "params": {
            "max_new_tokens": 150
        }
    },
    "cloudAccountId": "320247762680",
    "productName": "maas-model-mistral-7b-v0.1",
    "productId": "269c3034-e6c7-4359-9e77-c3efedfaa778"
}
EOF
)

grpcurl -plaintext \
-d "${REQUEST}" \
localhost:8443 \
proto.MaasGateway/GenerateStream