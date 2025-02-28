#!/usr/bin/env bash

# This script is used to powercycle the BM node

# The general command to run the script: ./instance_update_power_status <power_state> <run_strategy>
# Run the following command to run the script with parameters:
# To power 'off' the instance (RunStrategy 'Halted'): ./instance_update_power_status off
# To power 'on' the instance with default RunStrategy 'Always': ./instance_update_power_status on
# To power 'on' the instance with RunStrategy: ./instance_update_power_status on <run_strategy> (Always/RerunOnFailure)
  
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

# default runStrategy is 'Always'
DEFAULT_RUN_STRATEGY="Always"
RUN_STRATEGY=$DEFAULT_RUN_STRATEGY

if [ $# -eq 1 ]; then
    if [ "$1" == "off" ]; then
        RUN_STRATEGY="Halted"
    elif [ "$1" == "on" ]; then
        RUN_STRATEGY=$DEFAULT_RUN_STRATEGY
    else
        echo "Invalid parameter. Use 'on' or 'off'."
        exit 1
    fi
elif [ $# -eq 2 ]; then
    if [ "$1" != "on" ]; then
        echo "Invalid parameter. Use 'off' OR 'on' with RunStrategy 'Always' or 'RerunOnFailure'."
        exit 1
    fi

    RUN_STRATEGY="$2"
else
    echo "Invalid parameters"
    exit 1
fi

cat <<EOF | \
curl -vk \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X PUT \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances/name/${NAME} --data-binary @- \
| jq .
{
  "spec": {
    "runStrategy": "${RUN_STRATEGY}",
    "sshPublicKeyNames": [
      "${KEYNAME}"
    ]
  }
}
EOF