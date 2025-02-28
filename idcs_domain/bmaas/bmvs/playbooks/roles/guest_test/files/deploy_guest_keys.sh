#!/bin/bash

# Deploy the Guest Test SSH Keys to the Bastion Server
location=$(dirname $(readlink -f $0))
bastion="192.168.150.1"

# We should NOT change roots SSH credentials
if [ "$EUID" -eq 0 ]
  then echo "Please NEVER run as root"
  exit 1
fi

# Get user confirmation
echo ""
echo "!!! This will update your ${HOME}/.ssh/config !!!"
read -p "Okay to update? (yes/no) [n]: " update
case ${update} in
    [Yy]* ) ;;
    * ) echo "No changes made."; exit 0 ;;
esac

# Copy the Guest Test hosts SSH Keys locally
echo ""
echo "  Installing Guest Test SSH keys to ${HOME}/.ssh/ ... "
cp -a ${location}/guest-test-* \
    ${HOME}/.ssh/

# Install the config file
echo "  Installing .ssh/config to ${bastion} ... "
scp -i ${location}/bmvs-bastion-admin \
    ${location}/bastion-config \
    bastion@${bastion}:.ssh/config
if [ $? -ne 0 ]
then
    exit $?
fi

# Copy the Guest Test hosts SSH Keys
echo "  Installing Guest Test SSH keys to ${bastion} ... "
scp -i ${location}/bmvs-bastion-admin \
    ${location}/guest-test-* \
    bastion@${bastion}:.ssh/
if [ $? -ne 0 ]
then
    exit $?
fi

timestamp=$(date +%Y%m%dT%H%M%S)
# Make a backup of the original
echo "  Making a backup of  ${HOME}/.ssh/config as config-${timestamp} ..."
cp -a ${HOME}/.ssh/config \
    ${HOME}/.ssh/config-${timestamp}

# Update the ${HOME}/.ssh/config file safely
if [ $? -eq 0 ]
then
    echo "  Updating ${HOME}/.ssh/config ..."
    # make a second copy to ensure same permissions, etc
    cp -a ${HOME}/.ssh/config \
        ${HOME}/.ssh/edit-copy-config-${timestamp}
    # Start with our deployment-config at the top of file
    cat ${location}/deployment-config >${HOME}/.ssh/edit-copy-config-${timestamp}
    # Append the original file
    cat ${HOME}/.ssh/config-${timestamp} >> \
        ${HOME}/.ssh/edit-copy-config-${timestamp}
    # Atomic move into place
    mv -f \
        ${HOME}/.ssh/edit-copy-config-${timestamp} \
        ${HOME}/.ssh/config
else
    echo "!! Backup failed, no changes to ${HOME}/.ssh/config"
    exit 1
fi

exit 0
