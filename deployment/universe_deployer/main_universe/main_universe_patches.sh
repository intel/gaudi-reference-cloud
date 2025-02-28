#!/usr/bin/env bash
# This file can be used to override the changes made by Universe Deployer.
# It runs immediately before Universe Deployer Git Pusher commits changes to the local idc-argocd branch.
# It can be used to undo changes using "git checkout".
# Warning: Usage of this file should be avoided because it makes it difficult to control deployed versions.

set -ex

YQ=${YQ:-yq}

# Copy a yaml value from HEAD to the working directory.
revert_yaml() {
    local yaml_path=$1
    local file=$2
    local value="$(git show HEAD:${file} | ${YQ} ${yaml_path} -)"
    if [ "${value}" != "null" ]; then
        ${YQ} --inplace "${yaml_path} = \"${value}\"" ${file}
    fi
}

revert_ironic_passwords() {
    for file in applications/idc-regional/*/*/*-baremetal-operator-metal3-*/values.yaml; do
        revert_yaml .inspector.password ${file}
        revert_yaml .ironic.password ${file}
    done
}

main() {
    revert_ironic_passwords
}

main
