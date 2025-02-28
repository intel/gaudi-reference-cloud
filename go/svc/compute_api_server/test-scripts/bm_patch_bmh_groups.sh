#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -e

NAMESPACE=${NAMESPACE:-metal3-1}

usage() { echo "This script is utilized to assign BM groups and IPs for GPU network"
          echo "Usage: $0 [-g <group id>] [-n <node names separated by space>]" 1>&2; exit 1; }

while getopts ":g:n:" o; do
    case "${o}" in
        g)
            g=${OPTARG}
            GROUP_NAME=$g
            ;;
        n)
            n=${OPTARG}
            node_list=($n)
            # size check
            nodes_count=${#node_list[@]}
            ((nodes_count == 3 || nodes_count == 4 || nodes_count == 16)) || usage

            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ -z "${g}" ] || [ -z "${n}" ]; then
    usage
fi
# check if all nodes are valid
for node in ${node_list[@]};
do
     NAMESPACE=$(kubectl get bmh -A -o json | jq -r ".items[].metadata | select(.name==\"${node}\") | .namespace")
     if [ -z "${NAMESPACE}" ] ; then
          echo "$node is not available on this system"
          exit 1
     fi
done

# Annotation starts here
i=0
for node in ${node_list[@]}; 
do
     echo "Set IP for ${node}"
     NAMESPACE=$(kubectl get bmh -A -o json | jq -r ".items[].metadata | select(.name==\"${node}\") | .namespace")
     kubectl -n ${NAMESPACE} label bmh ${node} cloud.intel.com/cluster-size="${nodes_count}" --overwrite

     # set IP 
     kubectl patch bmh ${node}  -n ${NAMESPACE} --type='json'\
     -p="[{\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth0-0\", \"value\":\"192.168.$((i)).101/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth0-1\", \"value\":\"192.168.$((i)).102/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth0-2\", \"value\":\"192.168.$((i)).103/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth1-0\", \"value\":\"192.168.$((i)).104/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth1-1\", \"value\":\"192.168.$((i)).105/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth1-2\", \"value\":\"192.168.$((i)).106/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth2-0\", \"value\":\"192.168.$((i)).107/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth2-1\", \"value\":\"192.168.$((i)).108/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth2-2\", \"value\":\"192.168.$((i)).109/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth3-0\", \"value\":\"192.168.$((i)).110/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth3-1\", \"value\":\"192.168.$((i)).111/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth3-2\", \"value\":\"192.168.$((i)).112/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth4-0\", \"value\":\"192.168.$((i)).113/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth4-1\", \"value\":\"192.168.$((i)).114/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth4-2\", \"value\":\"192.168.$((i)).115/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth5-0\", \"value\":\"192.168.$((i)).116/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth5-1\", \"value\":\"192.168.$((i)).117/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth5-2\", \"value\":\"192.168.$((i)).118/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth6-0\", \"value\":\"192.168.$((i)).119/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth6-1\", \"value\":\"192.168.$((i)).120/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth6-2\", \"value\":\"192.168.$((i)).121/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth7-0\", \"value\":\"192.168.$((i)).122/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth7-1\", \"value\":\"192.168.$((i)).123/16\"},
          {\"op\": \"add\", \"path\": \"/metadata/annotations/gpu.ip.cloud.intel.com~1gpu-eth7-2\", \"value\":\"192.168.$((i++)).124/16\"}
     ]"

     # finally enable group
     kubectl patch bmh ${node}  -n ${NAMESPACE} --type='json'\
     -p="[{\"op\": \"add\", \"path\": \"/metadata/labels/cloud.intel.com~1cluster-size\", \"value\":\"${nodes_count}\"},
          {\"op\": \"add\", \"path\": \"/metadata/labels/cloud.intel.com~1instance-group-id\", \"value\":\"${GROUP_NAME}\"}
     ]"
 done
