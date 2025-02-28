#!/bin/bash
# SSH to a VM through an SSH proxy server whose authorized_keys file is maintained by SSH Proxy Operator.
# We must run ssh-agent so that non-default private key can be passed to both hosts.
set -ex

pkill ssh-agent || true
eval `ssh-agent`
ssh-add config/samples/sshkeys/testuser@example.com_id_rsa

ssh -v \
-J guest@internal-placeholder.com:22 \
ubuntu@10.45.190.26 "hostname; echo SUCCESS"

ssh-agent -k
