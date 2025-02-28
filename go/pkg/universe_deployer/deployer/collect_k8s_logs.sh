#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Collect logs from Kubernetes and store in LOGDIR.

KUBECTL=${KUBECTL:-kubectl}
LOGDIR=${LOGDIR:-/tmp/collect_k8s_logs}
: "${KUBECONTEXT:?environment variable is required}"

echo collect_k8s_logs.sh: Collecting logs for ${KUBECONTEXT} to ${LOGDIR}

mkdir -p ${LOGDIR}

${KUBECTL} --context ${KUBECONTEXT} get pods -A >& ${LOGDIR}/get-pods.log
${KUBECTL} --context ${KUBECONTEXT} get deployments -A >& ${LOGDIR}/get-deployments.log

# Get the list of unhealthy pods.
unhealthy_pods=$(${KUBECTL} --context ${KUBECONTEXT} get pods --all-namespaces -o \
    go-template='{{range $index, $element := .items}}{{range .status.containerStatuses}}{{if not .ready}}{{$element.metadata.namespace}}{{" "}}{{$element.metadata.name}}{{"\n"}}{{end}}{{end}}{{end}}')

# Describe and show logs for each unhealthy pod.
if [ ! -z "$unhealthy_pods" ]; then
    echo "$unhealthy_pods" | while read -r namespace pod; do
        ${KUBECTL} --context ${KUBECONTEXT} describe -n "$namespace" "pods/$pod" >& ${LOGDIR}/describe-$namespace-pods-$pod.log
        ${KUBECTL} --context ${KUBECONTEXT} logs --all-containers -n "$namespace" "$pod" >& ${LOGDIR}/logs-$namespace-pods-$pod.log
    done
fi

# Always show logs from these deployments.
DEPLOYMENTS="\
cloudaccount \
cloudaccount-enroll \
user-credentials \
"

for d in ${DEPLOYMENTS}; do
    namespace=idcs-system
    ${KUBECTL} --context ${KUBECONTEXT} describe -n "$namespace" deployment/${d} >& ${LOGDIR}/describe-$namespace-deployment-$d.log
    ${KUBECTL} --context ${KUBECONTEXT} logs --all-containers -n "$namespace" deployment/${d} >& ${LOGDIR}/logs-$namespace-deployment-$d.log
done

true
