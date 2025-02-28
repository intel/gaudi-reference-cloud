#!/usr/bin/env bash
# Get public keys of Kubernetes clusters.
# If the public keys are already in build/environments/${IDC_ENV}, then use those.
# Otherwise, if the directory ${SECRETS_DIR}/kubeconfig exists, then process each kubeconfig file in it.
# Otherwise, process using the KUBECONFIG environment variable (or the default $HOME/.kube/config).

set -ex

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SECRETS_DIR=${SECRETS_DIR:-/tmp}
JQ=${JQ:-jq}
KUBECTL=${KUBECTL:-kubectl}

JWK_SOURCE_DIR=${JWK_SOURCE_DIR:-${SCRIPT_DIR}/../../../build/environments/${IDC_ENV}/vault-jwk-validation-public-keys}
JWKDIR=${SECRETS_DIR}/vault-jwk-validation-public-keys

get_kubernetes_public_keys() {
  echo Getting public key from default Kubernetes cluster with KUBECONFIG=${KUBECONFIG}
  jwkpath=${JWKDIR}/$(basename ${KUBECONFIG}).jwk
  ${KUBECTL} get --raw \
    "$(${KUBECTL} get --raw /.well-known/openid-configuration | ${JQ} -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')" > ${jwkpath}
  cat ${jwkpath}
  echo
}

rm -rf ${JWKDIR}
mkdir -p ${JWKDIR}

if [ -d ${JWK_SOURCE_DIR} ] ; then
  cp ${JWK_SOURCE_DIR}/* ${JWKDIR}/
elif [ -d ${SECRETS_DIR}/kubeconfig ] ; then
  for f in ${SECRETS_DIR}/kubeconfig/* ; do
    export KUBECONFIG=$f
    get_kubernetes_public_keys
  done
else
  export KUBECONFIG=${KUBECONFIG:-${HOME}/.kube/config}
  get_kubernetes_public_keys
fi
