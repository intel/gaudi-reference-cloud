#!/bin/bash -l

# Need this to run has a Bash shell to ensure we have
# the no_proxy_append function
#

START_DIR=$(pwd)

echo "Setting proxy..."
no_proxy_append $(virsh net-dumpxml --network data | grep 'ip address' | awk -F\' '{print $2}')
printenv | grep -i no_proxy

if [ ! -d bifrost ]
then
    echo "Cloning bifrost..."
    git clone https://opendev.org/openstack/bifrost
    cd bifrost

    # git checkout -b stable/zed
    # Consider using master was it supports ubuntu 22.04 and 'zed' does not yet

    echo "Patching bifrost..."
    # Apply the patch due to a bug in the repo
    git am ${START_DIR}/0001-Drop-rootwrap.d-ironic-lib.filters-file.patch
else
    echo "SKPPING: bifrost directory already exists!"
    cd bifrost
fi

# Install Bifrost
## ./bifrost-cli --debug install \
##     --dhcp-pool 192.168.150.100-192.168.150.200 \
##     --develop \
##     --network-interface databr0 \
##     -e noauth_mode=true
# Use playbooks installation method
# https://docs.openstack.org/bifrost/latest/install/playbooks.html

echo "Installing bifrost requiremnts..."
bash ./scripts/env-setup.sh
source /opt/stack/bifrost/bin/activate

echo "Installing bifrost ..."
# echo "Enter to install with playbooks ..."
# read a
cd playbooks
ansible-playbook -vvvv \
    -i inventory/target \
    install.yaml \
    --extra-vars "@${START_DIR}/install_bifrost.yaml"

echo ""
echo "Verify that ironic and ironic inspector are reachable outside your node"
echo "curl http://$(hostname):6385/v1/ | jq ."
echo "curl http://$(hostname):5050/v1/ | jq ."
echo ""

exit 0
