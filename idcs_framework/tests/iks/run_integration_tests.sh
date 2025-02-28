#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# Empty any existing bearer or admin token values and expiries
sed -i 's/bearer_token:.*/bearer_token: ""/' config/config.yaml
sed -i 's/admin_token:.*/admin_token: ""/' config/config.yaml
sed -i 's/bearer_token_expiry:.*/bearer_token_expiry: 0/' config/config.yaml
sed -i 's/admin_token_expiry:.*/admin_token_expiry: 0/' config/config.yaml

# Function to prompt user to select the environment
select_environment() {
    read -p "select environment (dev/staging/prod): " environment
    echo "Selected environment: $environment"
    case $environment in "dev"|"staging"|"prod")
    return 0
    ;;
    *)
    echo "Invalid environment selection."
    return 1
    ;;
    esac
}



# Function to prompt user to enter his email and password
get_email_password() {
    default_email="iks-user@intel.com"
    default_paswd="test"
    read -p "Do you want to use the default email ($default_email)? (y/n):" response
    case $response in
        [Yy]* )
            email=$default_email
            password=$default_paswd
            ;;
        [Nn]* )
            while true; do
                read -p "Enter your email address: " email
                if [[ -z "$email" ]]; then
                echo "Email Address cannot be empty"
                else
                echo "Entered email address: $email"
                fi

                read -s -p "Enter your password: " password
                if [[ -z "$password" ]]; then
                echo "Password cannot be empty"
                else
                echo "Password successfully recorded."
                break
                fi
            done
            ;;
            * )
        echo "please enter y or n."
        get_email_password
        ;;
    esac
}

# prompt user to select environment
while ! select_environment; do
    continue
done

# Update the environment value
awk -v environmentValue="$environment" '/environment:/{$2="\"" environmentValue "\""} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml

# Define host values for each environment
case $environment in
    "dev")
host_value="https://iks-dev-vm-compute.caas.intel.com"
availabilityzonename="us-dev-1a"
networkinterfacevnetname="us-dev-1a-default"
global_host="https://iks-dev-vm.caas.intel.com"
    ;;
    "staging")
host_value="https://staging-idc-us-1.eglb.intel.com"
availabilityzonename="us-staging-1a"
networkinterfacevnetname="us-staging-1a-default"
global_host="https://staging.api.idcservice.net"
    ;;
    "prod")
    # Prompt user to select region
    read -p "select region (region1/region2/region3): " region
    echo "Selected region: $region"
    case $region in
    "region1")
host_value="https://compute-us-region-1-api.cloud.intel.com"
availabilityzonename="us-region-1a"
networkinterfacevnetname="us-region-1a-default"
global_host="https://api.idcservice.net"
    ;;
    "region2")
host_value="https://compute-us-region-2-api.cloud.intel.com"
availabilityzonename="us-region-2a"
networkinterfacevnetname="us-region-2a-default"
global_host="https://api.idcservice.net"
    ;;
    "region3")
host_value="https://compute-us-region-3-api.cloud.intel.com"
availabilityzonename="us-region-3a"
networkinterfacevnetname="us-region-3a-default"
global_host="https://api.idcservice.net"
    ;;
    *)
    echo "Invalid Region Selection."
    exit 1
    ;;
    esac
    ;;
esac

# prompt user to enter email and password
while ! get_email_password; do
    continue
done

# base64 encode password
base64Passwd=$(echo -ne $password | base64)

# update the email, password and account id in the auth config file to obtain the bearer token
jq --arg emailID "$email" '.authConfig.userAccounts[0].email = $emailID' config/auth_config.json > temp.json && mv temp.json config/auth_config.json
jq --arg passwd "$base64Passwd" '.authConfig.userAccounts[0].password = $passwd' config/auth_config.json > temp.json && mv temp.json config/auth_config.json
jq --arg cloudAccountID "$CloudAccountID" '.authConfig.userAccounts[0].cloudAccountId = $cloudAccountID' config/auth_config.json > temp.json && mv temp.json config/auth_config.json

# update the email, password and account id in the auth config file for prod and staging env
jq --arg emailID "$email" '.env.staging.adminAuthConfig.username = $emailID' config/auth_config.json > temp.json && mv temp.json config/auth_config.json
jq --arg passwd "$base64Passwd" '.env.staging.adminAuthConfig.password = $passwd' config/auth_config.json > temp.json && mv temp.json config/auth_config.json

jq --arg emailID "$email" '.env.prod.adminAuthConfig.username = $emailID' config/auth_config.json > temp.json && mv temp.json config/auth_config.json
jq --arg passwd "$base64Passwd" '.env.prod.adminAuthConfig.password = $passwd' config/auth_config.json > temp.json && mv temp.json config/auth_config.json


# # # Generate the Admin Token
go run ../iks_commons/token/cmd/main.go -a $environment

admin_token=$(yq eval '.admin_token' config/config.yaml)

# Update the host value
awk -v host="$host_value" '/host:/{$2="\"" host "\""} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml

# Update the cloud account host value
awk -v global_host="$global_host" '/global_host:/{$2="\"" global_host "\""} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml

# Update JSON file
jq --arg availabilityzonename "$availabilityzonename" --arg networkinterfacevnetname "$networkinterfacevnetname" '.vnets[0].availabilityzonename = $availabilityzonename | .vnets[0].networkinterfacevnetname = $networkinterfacevnetname' requests/create_nodegroup_request.json > temp.json && mv temp.json requests/create_nodegroup_request.json

# Print Updated Availability Zone Values in json file
echo "Updated Values in Create Nodegroup JSON file:"
jq '.vnets[0].availabilityzonename, .vnets[0].networkinterfacevnetname' requests/create_nodegroup_request.json

# Update the email value
awk -v default_account="$email" '/default_account:/{$2="\"" default_account "\""} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml

# Make Cloud Account API call
API_URL="${global_host}/v1/cloudaccounts/name/${email}"
RESPONSE=$(curl -H "Authorization: ${admin_token}" ${API_URL})
# Extract Cloud Account ID
CloudAccountID=$(echo "$RESPONSE" | jq -r '.id')

echo "Cloud Account for given $email is: $CloudAccountID"
# Validate Cloud Account
if [ -z $CloudAccountID ]; then
    echo "Error: The Cloud Account ID is empty or null"
    echo "Possible Reasons: Token you added has expired or is not valid."
    exit 1
fi

# generate user bearer token
go run ../iks_commons/token/cmd/main.go -u $email

bearer_token=$(yq eval '.bearer_token' config/config.yaml)

# if the create_cluster flag is false get the exisiting clusters for the given accountID
if [ $(yq eval '.create_cluster' config/config.yaml) = false ]; then
    #Make GET Clusters API Call
    API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/iks/clusters"
    RESPONSE=$(curl -H "Authorization: ${bearer_token}" ${API_URL} --insecure)

    #Extract Clusters
    CLUSTER_NAMES=$(echo "$RESPONSE" | jq -r '.clusters[].name')
    CLUSTER_MAP=$(echo "$RESPONSE" | jq -r '.clusters[] | {"clusterName": .name,"clusterUUID":.uuid}')

    if [ -z "$CLUSTER_NAMES" ]; then
        echo "No Clusters Found. Exiting the test suite."
        exit
    else
        #Display names and ask user to select one
        echo "Please select a cluster name:"
        PS3="Enter the number your choice: "
        select NAME in $CLUSTER_NAMES; do
            if [ -n "$NAME" ]; then
            break
            else
            echo "Invalid option. Please select a valid option."
            fi
        done

        ClusterUUID=$(echo $CLUSTER_MAP | jq -r --arg NAME "$NAME" 'select(.clusterName == $NAME)'| jq -r .clusterUUID)
        echo "{\"Selection Cluster Name\": \"$NAME\"}"
        echo "{\"Selection Cluster UUID\": \"$ClusterUUID\"}"

        # # Update JSON file
        jq --arg uuid "$ClusterUUID" '.clusterUUID = $uuid' requests/existing_cluster_details.json > temp.json && mv temp.json requests/existing_cluster_details.json
        jq --arg name "$NAME" '.clusterName = $name' requests/existing_cluster_details.json > temp.json && mv temp.json requests/existing_cluster_details.json
        echo "Updated the clusterUUID and clusterName field value in the requests/existing_cluster_details.json file."
    fi
fi

#Fetch the Create Cluster Flag Value and ask user to select available k8s version if flag value is true
if [ $(yq eval '.create_cluster' config/config.yaml) = true ]; then
    #Make K8s version API Call
    API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/iks/metadata/k8sversions"
    RESPONSE=$(curl -H "Authorization: ${admin_token}" ${API_URL} --insecure)

    # #Extract K8s Version Namesk8s
    K8sVersions=$(echo "$RESPONSE" | jq -r '.k8sversions[].k8sversionname')
    K8sVersions="1.28"
    #Display names and ask user to select one
    echo "Please select an option:"
    PS3="Enter the number your choice: "
    select NAME in $K8sVersions; do
        if [ -n "$NAME" ]; then
        break
        else
        echo "Invalid option. Please select a valid option."
        fi
    done
    echo "{\"Selection K8s Version Name \": \"$NAME\"}"
    # Update JSON file
    jq --arg name "$NAME" '.k8sversionname = $name' requests/create_new_cluster_request.json > temp.json && mv temp.json requests/create_new_cluster_request.json
    echo "Updated the k8sversionname field value in the requests/create_new_cluster_request.json file."

    ## Check if add_node_to_node_group, delete_node_from_node_group, delete_specific_node_group or delete_specific_vip are set to true
    if  ([ $(yq eval '.delete_specific_node_group' config/config.yaml) = true ] || [ $(yq eval '.delete_node_from_node_group' config/config.yaml) = true ]) || ([ $(yq eval '.add_node_to_node_group' config/config.yaml) = true ] || [ $(yq eval '.delete_specific_vip' config/config.yaml) = true ]); then
            #set the flags to be false
            echo "In create cluster flow so setting all existing cluster based test flags to false."
            awk -v val=false '/delete_specific_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
            awk -v val=false '/delete_node_from_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
            awk -v val=false '/add_node_to_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
            awk -v val=false '/delete_specific_vip:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
    fi
fi


#Fetch the NodeGroup Flag Value and ask user to select available instance types if flag value is true
if [ $(yq eval '.create_node_group' config/config.yaml) = true ]; then
    #Make InstanceTypes API Call
    API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/iks/metadata/instancetypes"
    RESPONSE=$(curl -H "Authorization: ${bearer_token}" ${API_URL} --insecure)
    #Extract Instance Type Names
    INSTANCETYPE_NAMES=$(echo "$RESPONSE" | jq -r '.instancetypes[].instancetypename')

    #Display names and ask user to select one
    echo "Please select an option:"
    PS3="Enter the number your choice: "
    select NAME in $INSTANCETYPE_NAMES; do
        if [ -n "$NAME" ]; then
        break
        else
        echo "Invalid option. Please select a valid option."
        fi
    done

    echo "{\"Selection Instance Type Name\": \"$NAME\"}"

    # Update JSON file
    jq --arg name "$NAME" '.instancetypeid = $name' requests/create_nodegroup_request.json > temp.json && mv temp.json requests/create_nodegroup_request.json
    echo "Updated the instancetypeid field value in the requests/create_nodegroup_request.json file."
fi

# Check for available SSH KEYS for the given cloud account ID and if no cloud accounts present create one ssh key
# Create or Check SSH Keys only if node group creation flag is true
if [ $(yq eval '.create_node_group' config/config.yaml) = true ]; then
    # SSH Keys URL
    API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/sshpublickeys"
    RESPONSE=$(curl -s -X GET -H "Authorization: ${bearer_token}" ${API_URL} --insecure)

    # Check the items length
    items_length=$(echo "$RESPONSE" | jq '.items | length')

    if [ "$items_length" -eq 0 ]; then
    # Create Env specific folder
    if [ ! -d ~/.ssh/$environment ]; then
        mkdir -p ~/.ssh/$environment
    fi
    # Generate SSH Key
    ssh-keygen -t rsa -b 4096 -f ~/.ssh/$environment/id_rsa
    # Copy SSH Key
    ssh_public_key_file=~/.ssh/$environment/id_rsa.pub
    # Read SSH Public Key into a variable
    ssh_public_key=$(cat "$ssh_public_key_file")
    # Create a Json Request Body
    json_body=$(jq -n --arg ssh_key "$ssh_public_key" '{"metadata": {"name":"iks-integ-test"}, "spec": {"sshPublicKey":$ssh_key}}')

    RESPONSE=$(curl -X POST -d "$json_body" -H "Authorization: ${bearer_token}" -H "Content-Type: application/json" ${API_URL} --insecure)

    name=$(echo "$RESPONSE" | jq -r '.metadata.name')
    # Update JSON file
    jq --arg name "$name" '.sshkeyname[0].sshkey = $name' requests/create_nodegroup_request.json > temp.json && mv temp.json requests/create_nodegroup_request.json
    echo "Updated the SSH KEY field value in the requests/create_nodegroup_request.json file."
    else
        name=$(echo "$RESPONSE" | jq -r '.items[0].metadata.name')
         # Update JSON file
        jq --arg name "$name" '.sshkeyname[0].sshkey = $name' requests/create_nodegroup_request.json > temp.json && mv temp.json requests/create_nodegroup_request.json
        echo "Updated the SSH KEY field value in the requests/create_nodegroup_request.json file."
    fi

    ## input count of nodes it should be between 1 to 10
    while :; do
    read -p "Enter desired number of nodes in the Node Group between 1 and 10: " count
    [[ $count =~ ^[0-9]+$ ]] || { echo "Invalid node count. Please enter a number between 1 and 10"; continue; }
    if ((count >= 0 && count <= 10)); then
        break
    else
        echo "Invalid node count. Please enter a number between 1 and 10"
    fi
    done

    jq --argjson nodeCount $count '.count = $nodeCount ' requests/create_nodegroup_request.json > temp.json && mv temp.json requests/create_nodegroup_request.json
fi

# if the delete node group or add node to node group flag is true get the exisiting nodegroups for the clusterID and AccountID
if [ $(yq eval '.create_cluster' config/config.yaml) = false ]; then
    if  ([ $(yq eval '.delete_specific_node_group' config/config.yaml) = true ] || [ $(yq eval '.delete_node_from_node_group' config/config.yaml) = true ]) || [ $(yq eval '.add_node_to_node_group' config/config.yaml) = true ]; then
        #Make GET Clusters API Call
        API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/iks/clusters/${ClusterUUID}/nodegroups"
        RESPONSE=$(curl -H "Authorization: ${admin_token}" ${API_URL} --insecure)

        #Extract NodeGroups
        NODE_GROUP_NAMES=$(echo "$RESPONSE" | jq -r '.nodegroups[].name')
        NODE_GROUP_MAP=$(echo "$RESPONSE" | jq -r '.nodegroups[] | {"nodeGroupName": .name,"nodegroupUUID":.nodegroupuuid}')

        if [ -z "$NODE_GROUP_NAMES" ]; then
            echo "No Node groups Found. Skipping Node group based tests in the test suite."
            #set the flags to be false
            awk -v val=false '/delete_specific_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
            awk -v val=false '/delete_node_from_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
            awk -v val=false '/add_node_to_node_group:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
        else
            #Display names and ask user to select one
            echo "Please select a node group name:"
            PS3="Enter the number your choice: "
            select NG_NAME in $NODE_GROUP_NAMES; do
                if [ -n "$NG_NAME" ]; then
                break
                else
                echo "Invalid option. Please select a valid option."
                fi
            done

            NodeGroupUUID=$(echo $NODE_GROUP_MAP | jq -r --arg NAME "$NG_NAME" 'select(.nodeGroupName == $NAME)'| jq -r .nodegroupUUID)
            echo "{\"Selection Node Group Name\": \"$NG_NAME\"}"
            echo "{\"Selection Node Group UUID\": \"$NodeGroupUUID\"}"

            # Update JSON file
            jq --arg uuid "$NodeGroupUUID" '.nodeGroupUUID = $uuid' requests/existing_node_group_details.json > temp.json && mv temp.json requests/existing_node_group_details.json
            jq --arg name "$NG_NAME" '.nodeGroupName = $name' requests/existing_node_group_details.json > temp.json && mv temp.json requests/existing_node_group_details.json
            echo "Updated the nodeGroupUUID and nodeGroupName field value in the requests/existing_node_group_details.json."
        fi
    fi
fi

# if the delete vip flag is true get the exisiting vips(ilbs) for the clusterID and AccountID
if [ $(yq eval '.create_cluster' config/config.yaml) = false ] && [ $(yq eval '.delete_specific_vip' config/config.yaml) = true ]; then
    #Make GET VIPs API Call
    API_URL="${host_value}/v1/cloudaccounts/${CloudAccountID}/iks/clusters/${ClusterUUID}/vips"
    RESPONSE=$(curl -H "Authorization: ${bearer_token}" ${API_URL} --insecure)

    #Extract VIPs
    VIP_NAMES=$(echo "$RESPONSE" | jq -r '.response[].name')
    VIP_MAP=$(echo "$RESPONSE" | jq -r '.response[] | {"vipName": .name,"vipUUID":.vipid}')

    if [ -z "$VIP_NAMES" ]; then
        echo "No VIPs Found. Skipping Delete Specific VIP test in the test suite."
        # set the flag delete_spicific_vip flag to false
        awk -v val=false '/delete_specific_vip:/{$2=val} 1' config/config.yaml > temp.yaml && mv temp.yaml config/config.yaml
    else
        #Display names and ask user to select one
        echo "Please select a VIP name:"
        PS3="Enter the number your choice: "
        select VIP_NAME in $VIP_NAMES; do
            if [ -n "$VIP_NAME" ]; then
            break
            else
            echo "Invalid option. Please select a valid option."
            fi
        done

        VIPUUID=$(echo $VIP_MAP | jq -r --arg NAME "$VIP_NAME" 'select(.vipName == $NAME)'| jq -r .vipUUID)
        echo "{\"Selection Node Group Name\": \"$VIP_NAME\"}"
        echo "{\"Selection Node Group UUID\": \"$VIPUUID\"}"

        # Update JSON file
        jq --arg uuid "$VIPUUID" '.vipUUID = $uuid' requests/existing_load_balancer_details.json > temp.json && mv temp.json requests/existing_load_balancer_details.json
        jq --arg name "$VIP_NAME" '.vipName = $name' requests/existing_load_balancer_details.json > temp.json && mv temp.json requests/existing_load_balancer_details.json
        echo "Updated the vipUUID and vipName field value in the requests/existing_load_balancer_details.json."
    fi
fi
ginkgo

if [ $? -eq 0 ]; then
    echo "Integration test passed!"
    exit 0
else
    echo "Integration tests failed!"
    exit 1
fi
