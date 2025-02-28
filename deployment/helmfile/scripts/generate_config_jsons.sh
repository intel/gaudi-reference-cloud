#! /usr/bin/env bash

ENVIRONMENT_NAME=${ENVIRONMENT_NAME:-staging}
ENVIRONMENT_TYPE=${ENVIRONMENT_TYPE:-global}
RELEASE_LIST_JSON=${RELEASE_LIST_JSON:-list.json}
OUTPUT_DIR=${OUTPUT_DIR:-/tmp/helm-values}

echo "Generating application jsons for $ENVIRONMENT_TYPE environment type."
jq -c '.[]' ${RELEASE_LIST_JSON} | while read i; do
    RELEASENAME=$(echo $i | jq -r ".name")
    NAMESPACE=$(echo $i | jq -r ".namespace")
    CHART=$(echo $i | jq -r ".chart")
    CHART_NAME=${CHART#*/} # removes prefix with registry name from chart name
    CHART_VERSION=$(echo $i | jq -r ".version")
    LABELS=$(echo $i | jq -r ".labels")
    IFS="," read -a LABELS_ARRAY <<< $LABELS

    # Parsing array items to environmental variables
    for pair in "${LABELS_ARRAY[@]}"
    do
        key=$(echo $pair | cut -d ':' -f 1)
        value=$(echo $pair | cut -d ':' -f 2)
        export $key=$value
    done
    case $ENVIRONMENT_TYPE in
        global)
            JSON_OUTPUT_PATH=${OUTPUT_DIR}/idc-global-services/${environmentName}/${kubeContext}/${RELEASENAME}/config.json
            ;;
        regional)
            JSON_OUTPUT_PATH=${OUTPUT_DIR}/idc-regional/${region}/${kubeContext}/${RELEASENAME}/config.json
            ;;
        az)
            JSON_OUTPUT_PATH=${OUTPUT_DIR}/idc-regional/${region}/${kubeContext}/${RELEASENAME}/config.json
            ;;
        az-network)
            JSON_OUTPUT_PATH=${OUTPUT_DIR}/idc-network/${region}/${kubeContext}/${RELEASENAME}/config.json
            ;;
    esac

    case $CHART in
        idc*)
            if [[ "$ENVIRONMENT_NAME" == "prod" ]]; then
                CHART_REGISTRY=amr-idc-registry.infra-host.com
            else
                CHART_REGISTRY=amr-idc-registry-pre.infra-host.com
            fi
            CHART_NAME=intelcloud/${CHART_NAME}
            ;;
        argo*)
            CHART_REGISTRY=https://argoproj.github.io/argo-helm
            ;;
        bitnami*)
            CHART_REGISTRY=https://charts.bitnami.com/bitnami
            ;;
        bootc*)
            CHART_REGISTRY=https://charts.boo.tc
            ;;
        external-secrets*)
            CHART_REGISTRY=https://charts.external-secrets.io
            ;;
        gitea-charts*)
            CHART_REGISTRY=https://dl.gitea.com/charts/
            ;;
        hashicorp*)
            CHART_REGISTRY=https://helm.releases.hashicorp.com
            ;;
        istio*)
            CHART_REGISTRY=https://istio-release.storage.googleapis.com/charts
            ;;
        jetstack*)
            CHART_REGISTRY=https://charts.jetstack.io
            ;;
        ingress-nginx*)
            CHART_REGISTRY=https://kubernetes.github.io/ingress-nginx
            ;;
        opal*)
            CHART_REGISTRY=https://permitio.github.io/opal-helm-chart
            ;;
        minio*)
            CHART_REGISTRY=https://charts.min.io
            ;;
        *)
            echo "Unknown registry $CHART is set in helmfile, automatic conversion to ArgoCD format is not possible."
            exit 1
            ;;
    esac
    echo "Writing $JSON_OUTPUT_PATH file"
    jq -n --arg releaseName "${RELEASENAME}" --arg chartVersion "${CHART_VERSION}" --arg namespace "${NAMESPACE}" \
    --arg chartName "${CHART_NAME}" --arg chartRegistry "${CHART_REGISTRY}" \
        '{"envconfig":{
            "releaseName":"\($releaseName)",
            "chartName":"\($chartName)",
            "chartVersion":"\($chartVersion)",
            "chartRegistry": "\($chartRegistry)",
            "namespace":"\($namespace)"
        }}' \
        > $JSON_OUTPUT_PATH
done
