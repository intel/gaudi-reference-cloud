#!/bin/bash -e
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# This script triggers validation on all the BareMetal hosts that are available

NAMESPACE=${NAMESPACE:-metal3-1}

usage() { echo "This script triggers validation on Baremetal hosts"
          echo "Usage: $0 <BM Host names followed by space> [-n <namespace>]" 
          echo "example: trigger_validation.sh host1 host2 -n metal3-2" 1>&2; exit 1; }

for host in "$@"; do
    if [[ $host = "-"* ]]; then
        break
    else
        host_list+=("$1")
        shift
    fi
done

while getopts ":n:h:" o; do
    case "${o}" in
        n)
            n=${OPTARG}
            NAMESPACE=$n
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

for host in "${host_list[@]}"
do
   state=$(kubectl get bmhost --field-selector=metadata.name="$host",metadata.namespace="$NAMESPACE" -l cloud.intel.com/verified -Ao jsonpath='{.items..status.provisioning.state}')
    if [[ -z "$state" ]]; then
        echo "Skipping validation for $NAMESPACE/$host since host is not present or cloud.intel.com/verified label is missing"
        continue
    fi
    if [ "$state" == "available" ]; then
        echo "Triggering validation for $NAMESPACE/$host"
        kubectl --namespace "$NAMESPACE" label --overwrite baremetalhosts.metal3.io "$host" cloud.intel.com/validation-check-failed- cloud.intel.com/verified- cloud.intel.com/validation-checking- cloud.intel.com/ready-to-test=true cloud.intel.com/validation-checking-completed- cloud.intel.com/validation-gating=true
    else
        echo "Skipping validation for $NAMESPACE/$host since host status is not available"
    fi
done

#Clean up the Gating flag so that all the nodes are part of the same group validation.
for host in "${host_list[@]}"
do
    kubectl --namespace "$NAMESPACE" label --overwrite baremetalhosts.metal3.io "$host" cloud.intel.com/validation-gating-
done
