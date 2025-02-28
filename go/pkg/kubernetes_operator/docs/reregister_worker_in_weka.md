# Reregister Worker in Weka

1. Add label to Cluster CR to skip weka registration. Label is added to the specific nodegroup's labels field.
    ```
    # Add label
    "iks.cloud.intel.com/storage-register": "false"

    # Check if label is added to nodegroup CR
    kubectl get nodegroup "${NODEGROUP_ID}" -o yaml
    ```

2. Stop and remove weka CLI from worker
    ```
    sudo weka local stop client
    sudo weka local rm client
    sudo weka local ps
    ```

3. Deregister worker from weka

    ```
    # Copy vault token into local/secrets/prod/VAULT_TOKEN
    
    # Generate client certificates
    export IDC_ENV=prod
    export COMMON_NAME=${USER}
    export CREATE_ROLE=1
    export VAULT_ADDR=https://internal-placeholder.com/
    make generate-vault-pki-cert

    # To get clusterId, get cluster CR and check Storage status
    kubectl get cluster "${CLUSTER_ID}" -o yaml

    # List registered agent and get clientId of worker to be deregistered
    grpcurl --cacert local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/ca.pem --cert local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/cert.pem --key local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/cert.key -d '{"clusterId": ""}' internal-placeholder.com:443 proto.FilesystemStorageClusterPrivateService/ListRegisteredAgents

    # Run deregister commmand. Ser correct clusterId and clientId
    grpcurl --cacert local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/ca.pem --cert local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/cert.pem --key local/secrets/${IDC_ENV}/pki/${COMMON_NAME}/cert.key   -d '{"clusterId": "", "clientId":""}' internal-placeholder.com:443  proto.FilesystemStorageClusterPrivateService/DeregisterAgent

    # Wait for worker to show empty weka status
    kubectl get cluster "${CLUSTER_ID}" -o yaml
    ```

4. Run weka CLI command
    ```
    STORAGE_WEKA_CONTAINER_NAME="client"
    STORAGE_WEKA_NUM_CORES=4

    # Get network information.
    INTERFACE_NAME=$(ip -o -4 route show to default | awk '{print $5}')

    # Some nodes will have a dedicated interface for storage configuration named "storage0-tenant".
    # If it is found, it will be used instead of default interface.
    if ip addr | grep "storage0-tenant"; then
    log_message "Dedicated storage interface found"
    INTERFACE_NAME="storage0-tenant"
    fi

    IP_ADDR=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f1 | head -n 1)
    NETMASK=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f2 | head -n 1)
    GATEWAY=$(ip -o -4 route show to default | awk '{print $3}')

    echo "Network information: $INTERFACE_NAME $IP_ADDR $NETMASK $GATEWAY"

    weka local setup container --name $STORAGE_WEKA_CONTAINER_NAME --cores $STORAGE_WEKA_NUM_CORES --only-frontend-cores --net "$INTERFACE_NAME/$IP_ADDR/$NETMASK/$GATEWAY"
    ```

5. Remove label from Cluster CR and wait for registration
    ```
    kubectl get cluster "${CLUSTER_ID}" -o yaml
    ```