<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# IKS Regression Testing

Developer Cloud Services team: IKS Team.

## Contributing

If you are part of the development team, or plan to contribute regularly, the recommendation is to request write access to this repository [here](https://internal-placeholder.com/identityiq/accessRequest/accessRequest.jsf#/accessRequest/manageAccess/add?filterKeyword=DevCloud%20Services%20developer&quickLink=Request%20Access), so you can clone it and push directly.
In AGS you only need to request the "DevCloud Services developer" role. Please don't request the "Owner-DevCloud Services developer" entitlement.

Otherwise, you can fork it and open a pull request with your contribution.

If you have any doubt or concern, feel free to contact rafael.calvo.mora@intel.com, andre.keedy@intel.com or claudio.fahey@intel.com.

## Verify Ginkgo is Installed

Run below command to make sure ginkgo is installed on your VM. 
```bash
ginkgo version
```
If you are getting ginkgo not found error try running below command to install ginkgo
```bash
sudo apt install ginkgo
```

## Libraries needed to run Ginkgo

The following libraries needs to be installed before running the ginkgo test suite
- go get github.com/onsi/ginkgo/ginkgo
- go get github.com/onsi/gomega

## ENV Variables needed to run in staging and production

Make sure you have correct no_proxy and https_proxy in your environment to test staging and production endpoints locally.
  - no_proxy == 'localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24'
  - https_proxy == 'http://internal-placeholder.com:912'


## Detailed Information to run the Regression Test Suite
Before running the run_integration_test.sh script make sure you have correct input request data and flag values are configured properly in config.yaml. Below are the few examples how the config.yaml and request.json files will behave based on the functionality. 

- config/config.yaml:-  check the flag values and enable the flags that are needed in order to run any specific flows.
    Note:- host, global_host, default_account, bearer_token, admin_token values will be automatically populated from the values provided in the run_integration_test.sh script.

  - Flow 1 (New Cluster Flow):- If you want to Create a Cluster, Create a Worker Node Group
    1. In this scenario we need to set create_cluster, create_node_group flags to true to create a cluster and worker node group. 2. Also make sure you are having correct request inputs under request/create_new_cluster_request.json and request/create_nodegroup_request.json

  - Flow 2 (New Cluster Flow):- If you want to Create Cluster, Create a Worker Node Group, Create ILB and Download Kubeconfig
    1. In this scenario we need to set create_cluster, create_node_group flags, create_vip, download_kubeconfig to true.
    2. Also make sure you are having correct request inputs under request/create_new_cluster_request.json, request/create_nodegroup_request.json and request/create_load_balancer_request.json.

  - Flow 3 (Existing Cluster Flow):- If you are working on existing Cluster, Wanted to Create a NodeGroup, Create a ILB
     1. In this scenario we need to set create_cluster flag to false. The create_node_group flags, create_vip needs to be set to true.
     2. Whenever create_Cluster_flag is false it always looks for existing cluster uuid in request/existing_cluster_details.json file. 
     3. If the cluster is empty then it will skip executing all the above scenarios. Only if we have valid cluster and cluster associated to that cloud account then only it will try execute all the above scenario's.

- run_integration_tests.sh :- This is the main file which is used to run the Regression Test Suite. There is no need to change  anything in this file except the TOKEN value. 
    Note:- There is a token.sh file which will generate the token. For any reason if token is not generated we manually need to add the token at line 10 in this file. When you run the script what ever token we have here will be automatically populated for bearer_token and admin_token in the config.yaml files
    
    1. This file consists of host information for different environments. If we wanted to changes any host or vnets information here is the place we are supposed to modify.

## Things to verify before running the script based on the flow
  - New Cluster Creation Flow:- If you are creating a new cluster, make sure create_cluster is set to true. Also verify the create cluster request under requests/create_new_cluster_request.json is having valid data.
  - Existing Cluster Flow:- If you are working on an existing cluster, make sure create_cluster is set to false and this will navigate to existing cluster flow. Also verify the existing cluster request under requests/existing_cluster_details.json is having valid cluster uuid associated to that cloud account.
  - Node Group Creation Flow:- If you are creating a new node group, make sure create_node_group is set to true. Also verify the create node group request under requests/create_nodegroup_request.json is having valid data.
  - Create Loadbalancer Flow:- If you are creating a new load balancer, make sure create_vip is set to true. Also verify the create load balancer request under requests/create_load_balancer_request.json is having valid data.
  - Download Kubeconfig:- If you want to download a kubeconfig, make sure download_kubeconfig is set to true. This will download kubeconfig of cluster to your local folder.
  - Delete Cluster Flow:- If you are working on deleting an existing cluster, make sure delete_cluster is set to true and this will navigate to existing cluster flow and delete the cluster. Also verify the existing cluster request under requests/existing_cluster_details.json is having valid cluster uuid associated to that cloud account.

## Command to run the script
Before running the script make sure you are in correct path.
```bash
cd go/pkg/iks_integration_test
./run_integration_tests.sh
```