#!/usr/bin/env bash
#
# Render Helm values and Argo CD manifests for all environments.
# Output directories can be compared to find the effective differences between Manifests Generator and Helmfile code.
#
# Example usage:
#   git checkout main
#   deployment/helmfile/scripts/test-helmfile-values.sh /tmp/main
#   git checkout my-branch
#   deployment/helmfile/scripts/test-helmfile-values.sh /tmp/my-branch
#   diff -r /tmp/main /tmp/my-branch
#

set -ex
set -o pipefail
cd "$(dirname "$0")/../../.."

ROOT_DIR=$(pwd)
OUTPUT_DIR=${1:-${ROOT_DIR}/local/manifests-generator_v1}
GIT_COMMIT=${GIT_COMMIT:-e0299ea1f6b09b7d4f9f5e44478a3d56193d24be}
SECRETS_BASE_DIR=${SECRETS_BASE_DIR:-/tmp/test-manifests-generator/secrets}

pids=()

generate_universe_config() {
	local IDC_ENV=$1
	local INPUT=$2
	local OUTPUT=$3
	sed "s/{IDC_ENV}/${IDC_ENV}/" ${INPUT} > ${OUTPUT}
}

testenv_background() {
    IDC_ENV=$1
	SECRETS_DIR=${SECRETS_BASE_DIR}/${IDC_ENV}
	OUTPUT_DIR_LOGS=${OUTPUT_DIR}/logs/${IDC_ENV}
	OUTPUT_DIR_MANIFESTS=${OUTPUT_DIR}/manifests/${IDC_ENV}
	OUTPUT_DIR_MANIFESTS_TARS=${OUTPUT_DIR}/manifests-tars/${IDC_ENV}
	OUTPUT_DIR_UNIVERSE_CONFIGS=${OUTPUT_DIR}/universe-configs/${IDC_ENV}
	UNIVERSE_CONFIG=${ROOT_DIR}/universe_deployer/environments/${IDC_ENV}.json

	mkdir -p ${OUTPUT_DIR_LOGS}
	mkdir -p ${OUTPUT_DIR_MANIFESTS}
	mkdir -p ${OUTPUT_DIR_MANIFESTS_TARS}
	mkdir -p ${OUTPUT_DIR_UNIVERSE_CONFIGS}

	if [ ! -f ${UNIVERSE_CONFIG} ]; then
		UNIVERSE_CONFIG=${OUTPUT_DIR_UNIVERSE_CONFIGS}/${IDC_ENV}.json
		generate_universe_config \
			${IDC_ENV} \
			${ROOT_DIR}/deployment/helmfile/scripts/test-manifests-generator.json \
			${UNIVERSE_CONFIG}
	fi

	IDC_ENV=$1 SECRETS_DIR=${SECRETS_DIR} make secrets \
	|& tee ${OUTPUT_DIR_LOGS}/make_secrets.log

	(bazel-bin/go/pkg/universe_deployer/cmd/manifests_generator/manifests_generator_/manifests_generator \
		--commit ${GIT_COMMIT} \
		--commit-dir ${ROOT_DIR}/local/manifests-generator/commit \
		--default-chart-registry localhost:5001 \
		--output ${OUTPUT_DIR_MANIFESTS_TARS}/manifests.tar \
		--secrets-dir ${SECRETS_DIR} \
		--snapshot \
		--universe-config ${UNIVERSE_CONFIG} && \
	tar -C ${OUTPUT_DIR_MANIFESTS} -x -f ${OUTPUT_DIR_MANIFESTS_TARS}/manifests.tar) \
	|& tee ${OUTPUT_DIR_LOGS}/manifests_generator.log
}

testenv() {
	testenv_background $1 &
	pids+=($!)
}

rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}/manifests
mkdir -p ${OUTPUT_DIR}/manifests-tars
mkdir -p ${OUTPUT_DIR}/logs

make build-manifests-generator

testenv dev27
testenv kind-singlecluster
testenv pdx03-c01-rgcp001-vm-4
SSH_PROXY_IP=10.12.34.56 testenv kind-jenkins
TEST_ENVIRONMENT_ID=6435a32c DOCKER_REGISTRY=localhost:5678 SSH_PROXY_IP=10.12.34.56 SSH_PROXY_USER=guest789 testenv test-e2e-compute-vm
TEST_ENVIRONMENT_ID=6435a32c DOCKER_REGISTRY=localhost:5678 SSH_PROXY_IP=10.12.34.56 SSH_PROXY_USER=guest789 testenv test-e2e-compute-bm
testenv prod
testenv staging

# Wait for all processes to complete in the background.
# Terminate all processes if any fail.
for pid in "${pids[@]}"; do
  wait "$pid" || (echo ERROR ; kill ${pids[*]} ; sleep 5s ; echo ERROR ; exit 1)
done

echo Done. Output directory: ${OUTPUT_DIR}
