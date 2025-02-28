#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# This script is executed immediately before stopping the Kind cluster.
# It can be used for collecting logs.
set -x

KUBECTL=${KUBECTL:-kubectl}

${KUBECTL} get pods -A
${KUBECTL} get deployments -A

telemetry_pods=$(${KUBECTL} get pods -n idcs-observability -l app.kubernetes.io/name=opentelemetry-collector-agent -o \
    go-template='{{range $index, $element := .items}}{{$element.metadata.namespace}}{{" "}}{{$element.metadata.name}}{{"\n"}}{{end}}')
if [ ! -z "$telemetry_pods" ]; then
    echo "$telemetry_pods" | while read -r namespace pod; do
        echo "Describe pod: $pod in namespace: $namespace"
        echo ------------------------------------------------------------------
        ${KUBECTL} describe -n "$namespace" "pods/$pod"
        echo ------------------------------------------------------------------
        echo "Logs for pod: $pod in namespace: $namespace"
        echo ------------------------------------------------------------------
        ${KUBECTL} logs -n "$namespace" "$pod"
        echo ------------------------------------------------------------------
    done
fi

# Get the list of unhealthy pods.
unhealthy_pods=$(${KUBECTL} get pods --all-namespaces -o \
    go-template='{{range $index, $element := .items}}{{range .status.containerStatuses}}{{if not .ready}}{{$element.metadata.namespace}}{{" "}}{{$element.metadata.name}}{{"\n"}}{{end}}{{end}}{{end}}')

# Describe and show logs for each unhealthy pod.
if [ ! -z "$unhealthy_pods" ]; then
    echo "$unhealthy_pods" | while read -r namespace pod; do
        echo "Describe unhealthy pod: $pod in namespace: $namespace"
        echo ------------------------------------------------------------------
        ${KUBECTL} describe -n "$namespace" "pods/$pod"
        echo ------------------------------------------------------------------
        echo "Logs for unhealthy pod: $pod in namespace: $namespace"
        echo ------------------------------------------------------------------
        ${KUBECTL} logs -n "$namespace" "$pod"
        echo ------------------------------------------------------------------
    done
fi

# Always show logs from these deployments.
DEPLOYMENTS="\
cloudaccount \
cloudaccount-enroll \
us-dev-1-compute-api-server \
us-dev-1a-ssh-proxy-operator \
us-dev-1a-vm-instance-scheduler \
"

for d in ${DEPLOYMENTS}; do
    echo ------------------------------------------------------------------
    ${KUBECTL} describe -n idcs-system deployment/${d}
    echo ------------------------------------------------------------------
    ${KUBECTL} logs -n idcs-system deployment/${d}
    echo ------------------------------------------------------------------
done

true
