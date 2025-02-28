#!/bin/bash
# Based on https://github.com/masmovil/bazel-rules/blob/master/helm/helm-chart-package.sh.tpl

set -e
set -o pipefail

TEMP_FILES="$(mktemp -t 2>/dev/null || mktemp -t 'helm_release_files')"

# Export XDG directories to get access to
# helm user defined repos
export XDG_CACHE_HOME={HELM_CACHE_PATH}
export XDG_CONFIG_HOME={HELM_CONFIG_PATH}
export XDG_DATA_HOME={HELM_DATA_PATH}

function read_variables() {
    local file="$1"
    local new_file="$(mktemp -t 2>/dev/null || mktemp -t 'helm_release_new')"
    echo "${new_file}" >> "${TEMP_FILES}"

    # Rewrite the file from Bazel for the form FOO=...
    # to a form suitable for sourcing into bash to expose
    # these variables as substitutions in the tag statements.
    sed -E "s/^([^ ]+) (.*)\$/export \\1='\\2'/g" < ${file} > ${new_file}
    source ${new_file}
}

HELM_CHART_VERSION=$(cat {HELM_CHART_VERSION})
APP_VERSION=$(cat {APP_VERSION})
HELM_CHART_NAME={HELM_CHART_NAME}
DIGEST_PATH={DIGEST_PATH}
IMAGE_REPOSITORY={IMAGE_REPOSITORY}
IMAGE_TAG={IMAGE_TAG}

function helm_package_deterministic() {
    local chart_path="$1"
    local destination="$2"
    local version="$3"
    local app_version="$4"

    local package_temp_dir=$(mktemp -d)
    {HELM_PATH} package ${chart_path} --dependency-update --destination ${package_temp_dir} --app-version ${app_version} --version $version 1>>/dev/null
    local temp_package_tgz=${package_temp_dir}/$HELM_CHART_NAME-${version}.tgz
    local package_tgz=${destination}/$HELM_CHART_NAME-${version}.tgz
    local extracted_package_temp_dir=$(mktemp -d)
    tar -C ${extracted_package_temp_dir} --touch -xzf ${temp_package_tgz}
    rm -rf ${package_temp_dir}
    tar -C ${extracted_package_temp_dir} --sort=name --owner=root:0 --group=root:0 --mtime='@0' -c $HELM_CHART_NAME | \
        gzip -n > ${package_tgz}
    rm -rf ${extracted_package_temp_dir}
}

# The directory CHART_PATH may contain files that should not be included in the chart.
# This may occur if files are deleted from a chart.
# To handle this, we create a temp directory containing only the files in the explicit list HELM_PACKAGE_DIR_FILES.
original_tar=$(mktemp)
tar -C {CHART_PATH} --dereference -cf ${original_tar} {HELM_PACKAGE_DIR_FILES}
chart_path=$(mktemp -d)
tar -C ${chart_path} -xf ${original_tar}
rm -f ${original_tar}

chart_values_path=${chart_path}/values.yaml

# Application docker image is not provided by other docker bazel rule
if  [ -z $DIGEST_PATH ]; then

    # Image repository is provided as a static value
    if [ "$IMAGE_REPOSITORY" != "" ] && [ -n $IMAGE_REPOSITORY ]; then
        {YQ_PATH} w -i ${chart_values_path} {VALUES_REPO_YAML_PATH} $IMAGE_REPOSITORY
        echo "Replaced image repository in chart values.yaml with: $IMAGE_REPOSITORY"
    fi

    # Image tag is provided as a static value
    if [ "$IMAGE_TAG" != "" ] && [ -n $IMAGE_TAG ]; then
        {YQ_PATH} w -i ${chart_values_path} {VALUES_TAG_YAML_PATH} $IMAGE_TAG
        echo "Replaced image tag in chart values.yaml with: $IMAGE_TAG"
    fi

fi

# Application docker image is provided by other docker bazel rule
if [ -n $DIGEST_PATH ] && [ "$DIGEST_PATH" != "" ]; then
    # extracts the digest sha and removes 'sha256' text from it
    DIGEST=$(cat {DIGEST_PATH})
    IFS=':' read -ra digest_split <<< "$DIGEST"
    DIGEST_SHA=${digest_split[1]}

    {YQ_PATH} w -i ${chart_values_path} {VALUES_TAG_YAML_PATH} $DIGEST_SHA

    echo "Replaced image tag in chart values.yaml with: $DIGEST_SHA"

    REPO_SUFIX="@sha256"

    if [ -n $IMAGE_REPOSITORY ] && [ "$IMAGE_REPOSITORY" != "" ]; then
        REPO_URL="{IMAGE_REPOSITORY}"
    else
        # if image_repository attr is not provided, extract it from values.yaml
        REPO_URL=$({YQ_PATH} r ${chart_values_path} {VALUES_REPO_YAML_PATH})
    fi

    # appends @sha256 suffix to image repo url value if the repository value does not already contains it
    if ([ -n $REPO_URL ] || [ -n $REPO_SUFIX ]) && ([[ $REPO_URL != *"$REPO_SUFIX" ]] || [[ -z "$REPO_SUFIX" ]]); then
        {YQ_PATH} w -i ${chart_values_path} {VALUES_REPO_YAML_PATH} ${REPO_URL}${REPO_SUFIX}
    fi
fi

if [ "{APPEND_CHART_HASH_TO_HELM_CHART_VERSION}" == "True" ]; then
    # Package temporary Helm chart using a fixed placeholder chart version.
    unversioned_version=0.0.0
    unversioned_package_temp_dir=$(mktemp -d)
    unversioned_package_tgz=${unversioned_package_temp_dir}/{HELM_CHART_NAME}-${unversioned_version}.tgz
    helm_package_deterministic ${chart_path} ${unversioned_package_temp_dir} ${unversioned_version} $APP_VERSION

    # Calculate hash of temporary Helm chart.
    helm_chart_hash=$(sha256sum -b ${unversioned_package_tgz} | cut -d ' ' -f 1)
    rm -rf ${unversioned_package_temp_dir}

    # Calculate final Helm chart version.
    extended_helm_chart_version=$HELM_CHART_VERSION-${helm_chart_hash}
else
    extended_helm_chart_version=$HELM_CHART_VERSION
fi

echo -n "${extended_helm_chart_version}" > {HELM_CHART_VERSION_FILE}

# Package final Helm chart using the chart version that includes the hash.
versioned_package_temp_dir=$(mktemp -d)
versioned_package_tgz=${versioned_package_temp_dir}/{HELM_CHART_NAME}-${extended_helm_chart_version}.tgz
helm_package_deterministic ${chart_path} ${versioned_package_temp_dir} $extended_helm_chart_version $APP_VERSION

mv ${versioned_package_tgz} {PACKAGE_OUTPUT_PATH}
rm -rf ${versioned_package_temp_dir}
rm -rf ${chart_path}

echo "Successfully packaged chart version ${extended_helm_chart_version} and saved it to: {PACKAGE_OUTPUT_PATH}"
