#!/bin/bash -e
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# This script triggers clean up on all BMH(BareMetal Hosts) which have failed validation.
# Post the cleanup the BMH will have the label cloud.intel.com/verified=true, which implies the BMH can be provisioned by the scheduler.

## Fetch all the BMH that have failed validation
NAMESPACE=${NAMESPACE:-metal3-1}

usage() { echo "This script triggers cleanup of Baremetal hosts that have failed validation"
          echo "Usage: $0 <BM Host names followed by space> [-n <namespace>]" 
          echo "example: trigger_cleanup.sh host1 host2 -n metal3-1" 1>&2; exit 1; }

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
   hostname=$(kubectl get bmhost --field-selector=metadata.name="$host",metadata.namespace="$NAMESPACE" -Ao jsonpath='{.items..metadata.name}')
    if [[ -z "$hostname" ]]; then
        echo "Skipping validation for $NAMESPACE/$host since host is not present or cloud.intel.com/validation-check-failed label is missing"
        continue
    fi
    echo "Triggering cleanup for $NAMESPACE/$host"
    kubectl --namespace "$NAMESPACE" label --overwrite baremetalhosts.metal3.io "$host" cloud.intel.com/validation-check-failed- cloud.intel.com/verified- cloud.intel.com/validation-checking- cloud.intel.com/group-validation-checking- cloud.intel.com/validation-master-node- cloud.intel.com/validation-instance-completed- cloud.intel.com/validation-id- cloud.intel.com/ready-to-test=true cloud.intel.com/validation-checking-completed=true
done
