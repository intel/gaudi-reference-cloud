#!/bin/bash

# Deploy the Bastion SSH Keys to THIS Deployment Server
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

# Install the Bastion server ssh keys
echo ""
echo "  Installing bmvs-bastion SSH Keys to ${HOME}/.ssh/ ... "
cp -a ${location}/bmvs-bastion* \
    ${HOME}/.ssh/

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
    # Start with our bastion server at the top of file
    cat > ${HOME}/.ssh/edit-copy-config-${timestamp} <<EOF
# Added by Bare Metal Virtual Stack scripting
Host bastion bmvs-bastion 192.168.150.1
    User guest
    HostName 192.168.150.1
    IdentityFile ~/.ssh/bmvs-bastion

# Added by Bare Metal Virtual Stack scripting
Host bastion-admin bmvs-bastion-admin
    User bastion
    HostName 192.168.150.1
    IdentityFile ~/.ssh/bmvs-bastion-admin

EOF
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
