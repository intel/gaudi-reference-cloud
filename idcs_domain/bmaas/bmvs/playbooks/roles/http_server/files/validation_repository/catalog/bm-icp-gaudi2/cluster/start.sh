#!/bin/bash -e

echo "starting bm-icp-gaudi2 cluster validation"
export MASTER_IP=$(printenv | awk -F '=' '/MASTER_IP/{print $2}')
export MEMBER_IPS=$(printenv | awk -F '=' '/MEMBER_IPS/{print $2}')
export TEST_CONFIGURATION=$(printenv | awk -F '=' '/TEST_CONFIGURATION/{print $2}')

# Here, FW_VERSION refers to Gaudi BUILD_VERSION (1.17.0-495)
export FW_VERSION=$(printenv | awk -F '=' '/BUILD_VERSION/{print $2}')

echo "MASTER_IP: ${MASTER_IP}"
echo "MASTER_NAME: ${MASTER_NAME}"
echo "MEMBER_IPS: ${MEMBER_IPS}"
echo "MEMBER_NAMES: ${MEMBER_NAMES}"
echo "TEST_CONFIGURATION: ${TEST_CONFIGURATION}"
echo "FW_VERSION: ${FW_VERSION}"

# Check if the TEST_CONFIGURATION environment variable is set
if [[ -z "$TEST_CONFIGURATION" ]]; then
    echo "TEST_CONFIGURATION is empty. Deepspeed Test for $FW_VERSION will be initiated"
    echo "starting deepseed cluster validation test"

    # Set the maximum number of forks in the ansible.cfg
    # If the number of nodes greater than 32, set the max forks to 32. Otherwise, set it to the number of nodes
    MEMBER_NODE_COUNT=$(echo $MEMBER_IPS | awk -F',' '{print NF}')
    TOTAL_NODE_COUNT=$((${MEMBER_NODE_COUNT} + 1))

    sed -i "/forks =/d" /tmp/validation/testsuite/deepspeed/playbooks/ansible.cfg

    if [ ${TOTAL_NODE_COUNT} -gt 32 ]
    then
        echo "forks = 32"  | tee -a /tmp/validation/testsuite/deepspeed/playbooks/ansible.cfg
    else
        echo "forks = ${TOTAL_NODE_COUNT}"  | tee -a /tmp/validation/testsuite/deepspeed/playbooks/ansible.cfg
    fi

    # Set environmental variable L3_ENABLED to true when /etc/gaudinet.json file is present
    sed -i "/export L3_ENABLED=/d" /tmp/validation/testsuite/deepspeed/.env
    if [ -f /etc/gaudinet.json ]; then
        echo "export L3_ENABLED=true" | tee -a /tmp/validation/testsuite/deepspeed/.env
    else
        echo "export L3_ENABLED=false" | tee -a /tmp/validation/testsuite/deepspeed/.env
    fi

    export SKIP_NETWORK_SETUP=${SKIP_NETWORK_SETUP:-true}

    cd /tmp/validation/testsuite/deepspeed
    source .env
    make validate-cluster-e2e SKIP_NETWORK_SETUP=$SKIP_NETWORK_SETUP
    echo "end of deepspeed cluster validation test"
elif [[ "$TEST_CONFIGURATION" == *"HCCL"* ]]; then
    echo "TEST_CONFIGURATION is set to HCCL."
    echo "starting hccl-demo cluster validation test"
    cd /tmp/validation/testsuite/hccl-demo
    make run-hccl-demo
    cd /tmp
    sudo rm -rf ssh-*
    sudo rm -rf ighs
    echo "end of hccl-demo cluster validation test"
else
    echo "TEST_CONFIGURATION is set to invalid test type. Exiting..."
fi

echo "end of bm-icp-gaudi2 cluster validation"
