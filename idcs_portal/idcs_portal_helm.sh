#!/bin/bash

set -ex

CHART="0.1.0"
GIT_SHORT_VERSION=$1
REGISTRY=$2
PROJECT=$3
HARBOR_USERNAME=$4
HARBOR_PASSWORD=$5
HELM_CHART_VERSION="${CHART}-${GIT_SHORT_VERSION}"

echo ${PWD}
# Update chart version
sed -i "/version:/ s/$/-${GIT_SHORT_VERSION}/" charts/idcs-portal/Chart.yaml

# package idcs helm chart
helm package charts/idcs-portal || { echo 'helm package failed' ; exit 1; }

# login to the harbor repo
echo "${HARBOR_PASSWORD}" | helm registry login -u ${HARBOR_USERNAME} --password-stdin ${REGISTRY}/${PROJECT} || { echo 'login to harbor repo failed' ; exit 1; }

# push chart to repo
helm push idcs-portal-${HELM_CHART_VERSION}.tgz oci://${REGISTRY}/${PROJECT} || { echo 'helm push to harbor repo failed' ; exit 1; }

# logout from the harbor repo
helm registry logout ${REGISTRY}/${PROJECT} || { echo 'logout from harbor repo failed' ; exit 1; }
