#!/bin/bash
# SSH to a VM through an SSH proxy server whose authorized_keys file is maintained by SSH Proxy Operator.
# This uses the default private key in ~/.ssh/id_rsa.
# ssh-agent is not required.
set -ex

ssh -v \
-J guest@internal-placeholder.com:22 \
ubuntu@10.165.62.252 "hostname; echo SUCCESS"
