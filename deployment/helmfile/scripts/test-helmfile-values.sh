#!/usr/bin/env bash
#
# Render Helm values and ArgoCD manifests for all environments.
# Output directories can be compared to find the effective differences between Helmfile code.
#
# Example usage:
#   export DOCKER_TAG=n2076-hece46d73
#   git checkout main
#   deployment/helmfile/scripts/test-helmfile-values.sh /tmp/main
#   git checkout my-branch
#   deployment/helmfile/scripts/test-helmfile-values.sh /tmp/my-branch
#   diff -r /tmp/main /tmp/my-branch
#

set -ex
cd "$(dirname "$0")/../../.."

OUTPUT_DIR=${1:-/tmp}

testenv() {
    export IDC_ENV=$1
    export HELMFILE_ARGOCD_VALUES_DIR=${OUTPUT_DIR}/test-helmfile/${IDC_ENV}/helm-values
    make helmfile-generate-argocd-values |& tee /tmp/test-helmfile-${IDC_ENV}.log &
}

rm -rf ${OUTPUT_DIR}/test-helmfile

IDC_ENV= make helm-push

testenv dev-jf
testenv dev1
testenv dev2
testenv dev3
testenv dev4
testenv dev5
testenv dev6
testenv dev7
testenv dev8
testenv dev9
testenv dev10
testenv kind-jenkins
testenv kind-multicluster
testenv kind-singlecluster
testenv pdx03-c01-rgcp001-vm-1
testenv prod
testenv staging

wait

echo Done. Output directory: ${OUTPUT_DIR}
