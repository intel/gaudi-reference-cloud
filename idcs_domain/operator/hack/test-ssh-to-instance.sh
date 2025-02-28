#!/bin/bash
# SSH directly to the instance.
set -ex

chmod go-rwx config/samples/sshkeys/testuser@example.com_id_rsa

ssh \
-i config/samples/sshkeys/testuser@example.com_id_rsa \
ubuntu@10.45.190.26
