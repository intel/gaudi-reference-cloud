#!/bin/bash
# SSH directly to the proxy server. This should connect but immediately disconect.
# Expected output:
#   PTY allocation request failed on channel 0
#   Connection to internal-placeholder.com closed.
set -ex

chmod go-rwx config/samples/sshkeys/testuser@example.com_id_rsa

ssh \
-i config/samples/sshkeys/testuser@example.com_id_rsa \
guest@internal-placeholder.com
