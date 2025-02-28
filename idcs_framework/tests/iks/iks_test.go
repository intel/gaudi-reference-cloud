// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks_integration_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	auth_admin "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks_commons/auth"

	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/lasso/pkg/log"
)

//TODO
// Update NodeGroup(Add or Delete Nodes)
// Upgrade IMI's or K8's
// Bearer Token Automatic Generation

const defaultPublicVipName = "public-apiserver"

var _ = Describe("Iks Integration Test Suite", func() {
	var clusterUUID, clusterState string
	var data []byte
	var clusterName string
	var nodeGroupName string
	var nodeGroupUUID string
	var vipName string
	var vipUUID int32
	var publicVip01Name string
	var publicVip01Port int32
	var defaultPublicVipPort int32
	var publicVipPrevStepOk bool

	BeforeEach(func() {
	})

	Describe("Iks Test Suite", func() {
		Context("Create Cluster", func() {
			It("Should create a new cluster if create_cluster flag is true", func() {
				if hostConfig.CreateCluster {
					log.Infof("==========Creating Cluster Flow========")

					// check number of clusters for the account ID
					data, err = SendHttpRequest(getClustersUrl, getMethod, nil, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())

					// parse the response
					var getClustersResp pb.ClustersResponse
					json.Unmarshal(data, &getClustersResp)

					if len(getClustersResp.Clusters) >= 3 {
						log.Errorf("Skipping create Cluster as maximum cluster limit of 3 reached.")
						Skip("Maximum cluster limit of 3 reached cannot create more clusters for this account ID.")
					}

					// create io reader for create cluster request
					requestData, err := ReadRequestData("requests/create_new_cluster_request.json")
					// assert error to be nil
					Expect(err).To(BeNil())
					type createClusterType struct {
						Description    string   `json:"description"`
						K8sVersionName string   `json:"k8sversionname"`
						Name           string   `json:"name"`
						RuntimeName    string   `json:"runtimename"`
						InstanceType   string   `json:"instanceType"`
						Tags           []string `json:"tags"`
					}
					var createClusterDetails createClusterType
					json.Unmarshal(requestData, &createClusterDetails)
					// add random number to the cluster name to make it unique
					createClusterDetails.Name = randomize(createClusterDetails.Name)

					// set the global variable clusterName so subsequent tests can use it
					clusterName = createClusterDetails.Name

					req, err := json.Marshal(createClusterDetails)
					Expect(err).To(BeNil())
					body := bytes.NewReader(req)

					// For pipeline runs replace k8s version with user specified version
					if pipelineRun {
						createClusterDetails.K8sVersionName = k8sVersion
					}

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}

					// send http request to create cluster
					data, err = SendHttpRequest(createClusterUrl, postMethod, body, bearerToken)
					// assert error to be nil
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())

					// parse server response
					var createClusterResponse pb.ClusterCreateResponseForm
					err = json.Unmarshal(data, &createClusterResponse)

					// assert api response to return status as Pending
					Expect(err).To(BeNil())
					Expect(createClusterResponse.Clusterstate).To(Equal("Pending"))

					// set the global variable clusterName so subsequent tests can use it
					clusterUUID = createClusterResponse.Uuid
					log.Infof("Creating cluster.  Name: %s,  ID: %s, Current state: %s", clusterName, clusterUUID, createClusterResponse.Clusterstate)

					if userEmail == noAdminAPIUserEmail {
						// Bug: Certain user accounts (ex. sys_idc_storage@intel.com) cannot access admin API v1/iks/admin/clusters/<cluster id>
						// This is a temporary workaround until bug with the user account is fixed.
						// Once the bug is fixed this IF block should be removed.
						log.Infof("User account %s - skipping querying cluster state due to lack of admin API access", userEmail)
						log.Infof("Sleeping for 25 minutes for cluster to get to Active state")
						time.Sleep(25 * time.Minute)
						clusterState = "Active"
					} else {
						log.Infof("==========Querying Cluster State========")
						// now query the server to get cluster status
						getClusterUrlUpdated := fmt.Sprintf(getClusterUrl, clusterUUID)
						getAdminClusterByUUIDURLUpdated := fmt.Sprintf(getAdminClusterByUUIDURL, clusterUUID)

						var getAdminClusterResponse pb.GetClusterAdmin
						var getClusterResponse pb.ClusterResponseForm
						var nodeGroupnameLogged, loggedNodeGroupStatus bool = false, false
						nodeNames := make(map[string]time.Time)
						nodeNamesActiveMap := make(map[string]bool)
						startingTimeForTest := time.Now()
						isCreateClusterFailed := true
						var activeNodesLogged int = 0
						var lastNode *pb.Node

						// wait for 3 control plane nodes to come up
						for {
							allControlPaneNodesActive := true

							// check if Admin Token has expired
							if isTokenExpired(bearerTokenExpiry) {
								// create new Admin Token and store in config file
								log.Infof("Admin Token Expired, generating new Admin Token")
								adminToken, adminTokenExpiry = auth_admin.Get_Azure_Admin_Bearer_Token(environment)
							}

							//get Admin Cluster by UUID Response
							data, err = SendHttpRequest(getAdminClusterByUUIDURLUpdated, getMethod, nil, adminToken)
							Expect(err).To(BeNil())
							Expect(data).ToNot(BeNil())
							err = json.Unmarshal(data, &getAdminClusterResponse)
							Expect(err).To(BeNil())

							if len(getAdminClusterResponse.Nodegroups) < 1 {
								allControlPaneNodesActive = false
								log.Infof("Sleeping for 30 seconds to allow cluster state to update")
								time.Sleep(30 * time.Second)
								continue
							}

							if len(getAdminClusterResponse.Nodegroups[0].Nodes) != 3 {
								allControlPaneNodesActive = false
							}

							if !nodeGroupnameLogged {
								controlPlanNodeGroup := getAdminClusterResponse.Nodegroups[0]
								log.Infof("<======Nodegroup Name==========><====Count=======><=====release version====><=====status======><======instancetype=========>")
								log.Infof("<======%s==========><====%d=======><=====%s====><=====%s======><======%s=========>", controlPlanNodeGroup.Name, int(controlPlanNodeGroup.Count), controlPlanNodeGroup.Releaseversion, controlPlanNodeGroup.Status, controlPlanNodeGroup.Instancetype)
								nodeGroupnameLogged = true
							}

							for i, node := range getAdminClusterResponse.Nodegroups[0].Nodes {
								if _, exists := nodeNames[node.Name]; !exists {
									nodeNames[node.Name] = time.Now()
									nodeNamesActiveMap[node.Name] = false
									log.Infof("<======CP Node Name==========><======IP Address=======><=============IMI============================><========State==========>")
									log.Infof("<======%s==========><====%s=======><=====%s====><=====%s======>", node.Name, node.Ipaddress, node.Imi, node.State)
									allControlPaneNodesActive = false
								} else {
									if node.State != "Active" {
										now := time.Now()
										if now.Sub(nodeNames[node.Name]).Minutes() > float64(hostConfig.CreateClusterTimeOutInMinutes) {
											log.Infof("Control Plane with name: %s because it took more than 10 minutes and current status of control plane is %s", node.Name, node.State)
											log.Errorf("=========Cluster provisioning failed, cleanup initiaited=========")
											isCreateClusterFailed = true
											Fail(fmt.Sprintf("Expected %s to be in Active State with in %d Minutes", node.Name, hostConfig.CreateClusterTimeOutInMinutes)) // this causes termination
										}
										allControlPaneNodesActive = false
									} else if !nodeNamesActiveMap[node.Name] {
										log.Infof("<======CP Node Name==========><======IP Address=======><=============IMI============================><========State==========>")
										log.Infof("<======%s==========><====%s=======><=====%s====><=====%s======>", node.Name, node.Ipaddress, node.Imi, node.State)
										nodeNamesActiveMap[node.Name] = true
										activeNodesLogged++
									}
								}
								if i == 2 {
									lastNode = getAdminClusterResponse.Nodegroups[0].Nodes[2]
								}

								log.Infof("Sleeping for 60 seconds to allow cluster state to update")
								time.Sleep(60 * time.Second)
							}

							if !loggedNodeGroupStatus && allControlPaneNodesActive && getAdminClusterResponse.Nodegroups[0].Status == "Active" {
								log.Infof("Node Group %s provisioned successfully", getAdminClusterResponse.Nodegroups[0].Name)
								loggedNodeGroupStatus = true
							}

							// check if bearerToken has expired
							if isTokenExpired(bearerTokenExpiry) {
								// create new bearer Token and store in config file
								log.Infof("Bearer Token Expired, generating new Bearer Token")
								bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
								Expect(err).To(BeNil())
							}
							data, err = SendHttpRequest(getClusterUrlUpdated, getMethod, nil, bearerToken)
							Expect(err).To(BeNil())
							err = json.Unmarshal(data, &getClusterResponse)
							Expect(err).To(BeNil())

							// if the cluster has already been provisioned, we break the loop
							if getClusterResponse.Clusterstate == "Active" {
								if activeNodesLogged == 2 {
									log.Infof("<======CP Node Name==========><======IP Address=======><=============IMI============================><========State==========>")
									log.Infof("<======%s==========><====%s=======><=====%s====><=====%s======>", lastNode.Name, lastNode.Ipaddress, lastNode.Imi, "Active")
									nodeNamesActiveMap[lastNode.Name] = true
									activeNodesLogged++
								}
								clusterState = getClusterResponse.Clusterstate
								log.Infof("Cluster is active with cluster ID: %s", getClusterResponse.Uuid)
								log.Infof("Cluster provisioned successfully")
								clusterState = "Active"
								isCreateClusterFailed = false
								break
							} else if getClusterResponse.Clusterstatus.Errorcode != 0 {
								log.Errorf("Cluster provisioning ran into error: %s with error code : %d", getClusterResponse.Uuid, getClusterResponse.Clusterstatus.Errorcode)
								log.Infof("Need to Initiate cleanup")
								isCreateClusterFailed = true
								break
							}

							if time.Now().Sub(startingTimeForTest).Minutes() > float64(3*hostConfig.CreateClusterTimeOutInMinutes) {
								log.Errorf("Cluster provisioning ran into error: %s", getClusterResponse.Uuid)
								log.Infof("Cluster took longer than the expected time : %d", 3*hostConfig.CreateClusterTimeOutInMinutes)
								isCreateClusterFailed = true
								break
							}

							log.Infof("Sleeping for 60 seconds to allow cluster state to update")
							time.Sleep(60 * time.Second)
						}
						if isCreateClusterFailed {
							Fail("Create Cluster is Failed")
						}
					}
				} else {
					log.Infof("Skipping Create a New Cluster test because the create_cluster_flag is false.")
					Skip("Skipping Create a New Cluster test because the create_cluster_flag is false.")
				}

			})

			It("Should fetch an existing cluster if create_cluster flag is false", func() {
				type existingClusterType struct {
					ClusterName string `json:"clusterName"`
					ClusterUUID string `json:"clusterUUID"`
				}
				var existingClusterDetails existingClusterType

				if pipelineRun {
					existingClusterDetails.ClusterName = clusterName
					existingClusterDetails.ClusterUUID = clusterUUID
				} else {
					requestData, err := ReadRequestData("requests/existing_cluster_details.json")
					Expect(err).To(BeNil())
					json.Unmarshal(requestData, &existingClusterDetails)
				}

				if !hostConfig.CreateCluster && (existingClusterDetails.ClusterUUID != "" || hostConfig.CreateNodeGroup || hostConfig.CreateVIP || hostConfig.DownloadKubeConfig || hostConfig.DeleteCluster) {
					log.Infof("==========Existing Cluster Flow========")

					type existingClusterType struct {
						ClusterName string `json:"clusterName"`
						ClusterUUID string `json:"clusterUUID"`
					}
					var existingClusterDetails existingClusterType

					if pipelineRun {
						existingClusterDetails.ClusterName = clusterName
						existingClusterDetails.ClusterUUID = clusterUUID

					} else {
						requestData, err := ReadRequestData("requests/existing_cluster_details.json")
						Expect(err).To(BeNil())

						json.Unmarshal(requestData, &existingClusterDetails)
					}

					if existingClusterDetails.ClusterUUID == "" {
						log.Errorf("The cluster details are missing in existing_cluster_details.json. Kindly update the file")
						Fail(fmt.Sprintf("Expecting existing Cluster not to be Empty where current value is : %s", existingClusterDetails.ClusterUUID)) // cause termination
					}

					// Validate Cluster is Available and Existing cluster is tagged to account ID
					log.Infof("==========Validate Cluster Existance========")

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}
					// IKS Get Clusters Endpoint
					data, err = SendHttpRequest(getClustersUrl, getMethod, nil, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					// parse the response
					var getClustersResp pb.ClustersResponse
					json.Unmarshal(data, &getClustersResp)

					isClusterAvailable := false
					for _, v := range getClustersResp.Clusters {
						if existingClusterDetails.ClusterUUID == v.Uuid {
							log.Infof("Cluster Exists to the associated Cloud Account and is Valid")
							isClusterAvailable = true
						}
					}

					if !isClusterAvailable {
						log.Errorf("The Given Cluster in request is not associated to this cloud account or should have deleted or maybe invalid.")
						Fail("Expecting Cluster associated to the given cloud account or Should be available.") //causes termination
					}

					log.Infof("==========Querying Cluster State========")

					// IKS Admin Get Cluster By UUIS Endpoint
					getAdminClusterByUUIDURLUpdated := fmt.Sprintf(getAdminClusterByUUIDURL, existingClusterDetails.ClusterUUID)

					// check if Admin Token has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new admin Token and store in config file
						adminToken, adminTokenExpiry = auth_admin.Get_Azure_Admin_Bearer_Token(environment)
					}

					data, err = SendHttpRequest(getAdminClusterByUUIDURLUpdated, getMethod, nil, adminToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					// parse the response
					var getAdminClusterResp pb.GetClusterAdmin
					json.Unmarshal(data, &getAdminClusterResp)

					Expect(getAdminClusterResp.Name).ToNot(Equal("")) // ensures that cluser details were fetched
					log.Infof("<======Cluster Name==========><====Cluster UUID=======>")
					log.Infof("<======%s==========><====%s=======>", getAdminClusterResp.Name, existingClusterDetails.ClusterUUID)
					for _, nodeGroup := range getAdminClusterResp.Nodegroups {
						log.Infof("<======Node Group Name==========><====Node Group ID=======><====-Version=======><====Status=======><====Node Count=======>")
						log.Infof("<======%s==========><====%s=======><====%s=======><====%s=======><====%d=======>", nodeGroup.Name, nodeGroup.Id,
							nodeGroup.Releaseversion, nodeGroup.Status, len(nodeGroup.Nodes))
						for _, node := range nodeGroup.Nodes {
							log.Infof("<======Node Name==========><====Node IP=======><====IMI=======><====Status=======>")
							log.Infof("<======%s==========><====%s=======><====%s=======><====%s=======>", node.Name, node.Ipaddress, node.Imi, node.State)

						}
					}
					clusterUUID = existingClusterDetails.ClusterUUID
					clusterState = "Active"
				} else {
					log.Infof("Skipping Fetch an Existing Cluster test since all the flags are disabled in config file.")
					Skip("Skipping Fetch an Existing Cluster test since all the flags are disabled in config file.")
				}
			})
		})

		Context("Create Node Group", func() {
			It("Should Create a node group for a cluster and cluster state is active", func() {

				log.Infof("Creating node group for clusterUUID: %s, clusterState: %s", clusterUUID, clusterState)

				if hostConfig.CreateNodeGroup && clusterUUID != "" && clusterState == "Active" {
					log.Infof("==========Creating Node Group===========")
					// If cluster is in Active State, we proceed to create NodeGroup
					// create io reader for create node group request
					requestData, err := ReadRequestData("requests/create_nodegroup_request.json")
					// assert error to be nil
					Expect(err).To(BeNil())

					type vnetDetailsType struct {
						Availabilityzonename     string `json:"availabilityzonename"`
						Networkinterfacevnetname string `json:"networkinterfacevnetname"`
					}
					type sshkeyDetailsType struct {
						Sshkey string `json:"sshkey"`
					}
					type createNodeGroupType struct {
						Count          int32               `json:"count"`
						Vnets          []vnetDetailsType   `json:"vnets"`
						Instancetypeid string              `json:"instancetypeid"`
						InstanceType   string              `json:"instanceType"`
						Name           string              `json:"name"`
						Description    string              `json:"description"`
						Tags           []string            `json:"tags"`
						Sshkeyname     []sshkeyDetailsType `json:"sshkeyname"`
					}
					var createNodeGroupDetails createNodeGroupType
					json.Unmarshal(requestData, &createNodeGroupDetails)

					log.Infof("Payload for creating node group: %+v\n", createNodeGroupDetails)

					// For pipeline runs replace vnet the values based on region and other values based on user input
					if pipelineRun {
						avzone := region + "a"
						vnet := region + "a-default"

						createNodeGroupDetails.Vnets[0].Availabilityzonename = avzone
						createNodeGroupDetails.Vnets[0].Networkinterfacevnetname = vnet

						createNodeGroupDetails.Sshkeyname[0].Sshkey = sshKeyName
					}

					// add random number to the node group name to make it unique
					createNodeGroupDetails.Name = randomize(createNodeGroupDetails.Name)
					nodeGroupName = createNodeGroupDetails.Name
					req, err := json.Marshal(createNodeGroupDetails)
					Expect(err).To(BeNil())
					body := bytes.NewReader(req)

					createNodeGroupUrl = fmt.Sprintf(createNodeGroupUrl, clusterUUID)

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}

					// send request
					data, err = SendHttpRequest(createNodeGroupUrl, postMethod, body, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())

					var createNodeGroupResponse pb.NodeGroupResponseForm
					json.Unmarshal(data, &createNodeGroupResponse)

					// assert if node group creation call was success
					Expect(createNodeGroupResponse.Nodegroupuuid).ToNot(Equal(""))

					nodeGroupUUID = createNodeGroupResponse.Nodegroupuuid

					log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====><=====Node Count======>")
					log.Infof("<======%s==========><====%s=======><=====%s====><=====%d======>",
						createNodeGroupResponse.Nodegroupuuid, createNodeGroupResponse.Name,
						createNodeGroupResponse.Nodegroupstate, createNodeGroupResponse.Count)

					// this create request might take upto x min based on the VM Type.
					// The field is cofigurable in config file in minutes for each individual node and total time is calculated below for all nodes.
					// iterating to check node group creation status
					noOfIterations := createNodeGroupResponse.Count * 2 * int32(hostConfig.CreateNodeGroupTimeOutInMinutes)
					getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, createNodeGroupResponse.Nodegroupuuid)
					var getNodeGroupByIdResponse pb.NodeGroupResponseForm
					log.Infof("==========Getting Node Group Status===========")
					shouldBreakOuterFor := false
					isCreateNodeGroupFailed := true
					nodeStatusMap := make(map[string]map[string]bool)
					for i := 0; i < int(noOfIterations); i++ {
						// send request

						// check if bearerToken has expired
						if isTokenExpired(bearerTokenExpiry) {
							// create new bearer Token and store in config file
							log.Infof("Bearer Token Expired, generating new Bearer Token")
							bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
							Expect(err).To(BeNil())
						}

						data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
						Expect(err).To(BeNil())
						Expect(data).ToNot(BeNil())

						json.Unmarshal(data, &getNodeGroupByIdResponse)

						if getNodeGroupByIdResponse.Nodegroupstatus != nil && getNodeGroupByIdResponse.Nodegroupstatus.Errorcode != 0 {
							log.Errorf("Unable to create node group: %s for Cluster with cluster ID: %s due to %s", createNodeGroupResponse.Name,
								clusterUUID, getNodeGroupByIdResponse.Nodegroupstatus.Message)
							isCreateNodeGroupFailed = true
							break
						}

						for _, node := range getNodeGroupByIdResponse.Nodes {
							if node.Errorcode != 0 {
								log.Errorf("Unable to create node group: %s, Error Node is %s and error is %s", createNodeGroupResponse.Name,
									node.Name, node.Message)
								isCreateNodeGroupFailed = true
								shouldBreakOuterFor = true
								break
							}
						}

						// if any node ran into error
						if shouldBreakOuterFor == true {
							break
						}

						allNodesActive := true
						for _, node := range getNodeGroupByIdResponse.Nodes {
							if nodeStatus, exists := nodeStatusMap[node.Name]; exists {
								// if node already existed, check if current state already present
								if _, stateExists := nodeStatus[node.State]; !stateExists {
									// means node went into active state from pending state
									log.Infof("<======Node Group Name==========><=====Node Name======><=====Node Status====>")
									log.Infof("<======%s==========><====%s=======><=====%s====>", getNodeGroupByIdResponse.Name,
										node.Name, node.State)
									nodeStatus[node.State] = true
									nodeStatusMap[node.Name] = nodeStatus
								}
							} else {
								log.Infof("<======Node Group Name==========><=====Node Name======><=====Node Status====>")
								log.Infof("<======%s==========><====%s=======><=====%s====>", getNodeGroupByIdResponse.Name,
									node.Name, node.State)

								nodeStatusMap[node.Name] = map[string]bool{node.State: true}
							}
							if node.State != "Active" {
								allNodesActive = false
							}
						}

						if allNodesActive && getNodeGroupByIdResponse.Nodegroupstatus.State == "Active" {
							log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====><=====Node Count======>")
							log.Infof("<======%s==========><====%s=======><=====%s====><=====%d======>",
								createNodeGroupResponse.Nodegroupuuid, createNodeGroupResponse.Name,
								getNodeGroupByIdResponse.Nodegroupstatus.State, createNodeGroupResponse.Count)
							isCreateNodeGroupFailed = false
							break
						}

						log.Infof("Sleeping for 30 seconds to allow node group state to update")
						time.Sleep(30 * time.Second)
					}
					if isCreateNodeGroupFailed {
						Fail("Expected the Node Group to be active with all the nodes in active state.")
					}
				}
				if hostConfig.CreateNodeGroup && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("Skipping Create NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState)
					Skip(fmt.Sprintf("Skipping Create NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState))
				}
				if !hostConfig.CreateNodeGroup {
					log.Infof("Skipping Create NodeGroup test since the Create Node Group flag is false.")
					Skip("Skipping Create NodeGroup test since the Create Node Group flag is false.")
				}
			})
		})

		// Add node to an existing NodeGroup

		// obtain status of the nodegroup and obtain the count of nodes
		// send a put request to update the count to count+1
		// monitor to ensure the count has been updated

		Context("Add node to an existing Node Group", func() {
			It("Should add a node to for a given nodegroup and cluster ID", func() {
				if hostConfig.AddNodeToNodeGroup && clusterUUID != "" && clusterState == "Active" {

					log.Infof("======Add Node to NodeGroup Flow==========")
					if clusterUUID == "" {
						log.Errorf("Cluster UUID cannot be empty while adding Node to a NodeGroup for a cluster")
						Expect(true).To(Equal(false))
					}

					log.Infof("=========Initiating Add Node to Node Group for cluster=========")

					type existingNodeGroupType struct {
						NodeGroupName string `json:"nodeGroupName"`
						NodeGroupUUID string `json:"nodeGroupUUID"`
					}
					var existingNodeGroupDetails existingNodeGroupType

					if pipelineRun {
						existingNodeGroupDetails.NodeGroupName = nodeGroupName
						existingNodeGroupDetails.NodeGroupUUID = nodeGroupUUID
					} else {
						// obtain the node group id
						requestData, err := ReadRequestData("requests/existing_node_group_details.json")
						Expect(err).To(BeNil())

						json.Unmarshal(requestData, &existingNodeGroupDetails)
					}

					// check if nodeGroupUUID is empty
					if existingNodeGroupDetails.NodeGroupUUID == "" {
						log.Errorf("Node Group UUID cannot be empty while adding Node to NodeGroup for a cluster")
						Expect(true).To(Equal(false))
					} else {
						// obtain the count of nodes for the nodegroupID
						getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
						data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
						Expect(err).To(BeNil())

						var getNodeGroupByIdResponse pb.NodeGroupResponseForm
						err = json.Unmarshal(data, &getNodeGroupByIdResponse)
						Expect(err).To(BeNil())

						// count of nodes
						nodeCount := getNodeGroupByIdResponse.Count
						log.Infof("Number of nodes in the node group : %d", nodeCount)

						// Check if count is greater than or equal to 10
						if nodeCount >= 10 {
							log.Errorf("Node Count reached maximum count 10, cannot add more nodes")
						} else {
							// Add one more node to the nodegroup
							type addNodeToNodeGroupType struct {
								Count int32 `json:"count"`
							}
							nodeCount++
							updatedCountStruct := addNodeToNodeGroupType{Count: nodeCount}
							jsonBody, err := json.Marshal(updatedCountStruct)
							Expect(err).To(BeNil())

							body := bytes.NewReader(jsonBody)

							putNodeGroupByIdUrlUpdated := fmt.Sprintf(putNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
							data, err = SendHttpRequest(putNodeGroupByIdUrlUpdated, putMethod, body, bearerToken)
							Expect(err).To(BeNil())

							var putNodeGroupByIdResponse pb.NodeGroupResponseForm
							err = json.Unmarshal(data, &putNodeGroupByIdResponse)
							Expect(err).To(BeNil())

							// Wait until the nodegroup comes back to active state

							log.Infof("============Getting Cluster Node Group Status===========")
							isNodeAddedSuccessfully := false
							isUpdating := false
							for i := 0; i < 20; i++ {
								// wait for 15 seconds before querying delete confirmation
								time.Sleep(15 * time.Second)

								getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
								data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
								Expect(err).To(BeNil())

								var getNodeGroupByIdResponse pb.NodeGroupResponseForm
								err = json.Unmarshal(data, &getNodeGroupByIdResponse)
								Expect(err).To(BeNil())

								if isNodeAddedSuccessfully == false {
									if getNodeGroupByIdResponse.Nodegroupstatus.State == "Active" {
										log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
										log.Infof("<======%s==========><====%s=======><=====New Node added to Node Group====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName)
										isNodeAddedSuccessfully = true
									} else if isUpdating == false {
										log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
										log.Infof("<======%s==========><====%s=======><=====%s====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName, getNodeGroupByIdResponse.Nodegroupstatus.State)
										isUpdating = true
									}
								} else {
									break
								}

							}

							if isNodeAddedSuccessfully == false {
								Fail("Expected Node to be added to the Node Group in 5 minutes but took longer than expected. Please check manually.")
							}
						}

					}

				}
				if hostConfig.AddNodeToNodeGroup && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("Skipping add Node to NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState)
					Skip(fmt.Sprintf("Skipping andd Node to NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState))
				}
				if !hostConfig.AddNodeToNodeGroup {
					log.Infof("Skipping Add Node to NodeGroup test since the Add Node to Node Group flag is false.")
					Skip("Skipping Add Node to NodeGroup test since the Add Node to Node Group flag is false.")
				}
			})
		})

		//Delete node from a node group
		Context("Delete node from an existing Node Group", func() {
			It("Should delete a node from a nodegroup with given nodegroup ID and cluster ID", func() {
				if hostConfig.DeleteNodeFromNodeGroup && clusterUUID != "" && clusterState == "Active" {

					log.Infof("======Delete Node From NodeGroup Flow==========")
					if clusterUUID == "" {
						log.Errorf("Cluster UUID cannot be empty while deleting Node from a NodeGroup for a cluster")
						Expect(true).To(Equal(false))
					}

					log.Infof("========= Initiating Delete Node from a Node Group =========")

					type existingNodeGroupType struct {
						NodeGroupName string `json:"nodeGroupName"`
						NodeGroupUUID string `json:"nodeGroupUUID"`
					}
					var existingNodeGroupDetails existingNodeGroupType

					if pipelineRun {
						existingNodeGroupDetails.NodeGroupName = nodeGroupName
						existingNodeGroupDetails.NodeGroupUUID = nodeGroupUUID
					} else {
						// obtain the node group id
						requestData, err := ReadRequestData("requests/existing_node_group_details.json")
						Expect(err).To(BeNil())

						json.Unmarshal(requestData, &existingNodeGroupDetails)
					}

					// check if nodeGroupUUID is empty
					if existingNodeGroupDetails.NodeGroupUUID == "" {
						log.Errorf("Node Group UUID cannot be empty while deleting Node from a NodeGroup")
						Expect(true).To(Equal(false))
					} else {
						// obtain the count of nodes for the nodegroupID
						getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
						data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
						Expect(err).To(BeNil())

						var getNodeGroupByIdResponse pb.NodeGroupResponseForm
						err = json.Unmarshal(data, &getNodeGroupByIdResponse)
						Expect(err).To(BeNil())

						// count of nodes
						nodeCount := getNodeGroupByIdResponse.Count
						log.Infof("Number of nodes in the node group : %d", nodeCount)

						// Check if count is equal to 0
						if nodeCount == 0 {
							log.Errorf("Node Count is zero cannot delete more node")
						} else {
							// Add one more node to the nodegroup
							type addNodeToNodeGroupType struct {
								Count int32 `json:"count"`
							}
							nodeCount--

							updatedCountStruct := addNodeToNodeGroupType{Count: nodeCount}
							jsonBody, err := json.Marshal(updatedCountStruct)
							Expect(err).To(BeNil())

							body := bytes.NewReader(jsonBody)

							putNodeGroupByIdUrlUpdated := fmt.Sprintf(putNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
							data, err = SendHttpRequest(putNodeGroupByIdUrlUpdated, putMethod, body, bearerToken)
							Expect(err).To(BeNil())

							var putNodeGroupByIdResponse pb.NodeGroupResponseForm
							err = json.Unmarshal(data, &putNodeGroupByIdResponse)
							Expect(err).To(BeNil())

							// Wait until the nodegroup comes back to active state

							log.Infof("============Getting Cluster Node Group Status===========")
							isNodeDeletedSuccessfully := false
							isUpdating := false
							for i := 0; i < 20; i++ {
								// wait for 15 seconds before querying delete confirmation
								time.Sleep(15 * time.Second)

								getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
								data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
								Expect(err).To(BeNil())

								var getNodeGroupByIdResponse pb.NodeGroupResponseForm
								err = json.Unmarshal(data, &getNodeGroupByIdResponse)
								Expect(err).To(BeNil())

								if isNodeDeletedSuccessfully == false {
									if getNodeGroupByIdResponse.Nodegroupstatus.State == "Active" {
										log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
										log.Infof("<======%s==========><====%s=======><=====Node deleted from Node Group====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName)
										isNodeDeletedSuccessfully = true
									} else if isUpdating == false {
										log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
										log.Infof("<======%s==========><====%s=======><=====%s====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName, getNodeGroupByIdResponse.Nodegroupstatus.State)
										isUpdating = true
									}
								} else {
									break
								}

							}

							// if the delete was not succesfull within the give time frame log timeout and Fail the test case
							if !isNodeDeletedSuccessfully {
								Fail("Expected Node to be deleted from the Node Group in 5 minutes but took longer than expected. Need to Process Manual Cleanup")
							}
						}

					}

				}
				if hostConfig.DeleteNodeFromNodeGroup && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("Skipping delete Node from NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState)
					Skip(fmt.Sprintf("Skipping delete Node from NodeGroup test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState))
				}
				if !hostConfig.DeleteNodeFromNodeGroup {
					log.Infof("Skipping delete node from NodeGroup test since the Delete Node from Node Group flag is false.")
					Skip("Skipping delete node from NodeGroup test since the Delete Node from Node Group flag is false.")
				}
			})
		})

		Context("Create public load balancer", func() {
			It("Create public load balancer", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				publicVipPrevStepOk = false

				// Set global variables
				publicVip01Name = randomize("publiclb01")
				publicVip01Port = 80

				_, err := CreateVIP(clusterUUID, publicVip01Name, "public", publicVip01Port, "test lb description")
				Expect(err).To(BeNil())

				log.Infof("Sleeping for 30 seconds to allow VIP creation to start...")
				time.Sleep(30 * time.Second)

				_, vipId, _, err := GetSecRulesInfoByName(clusterUUID, publicVip01Name)
				Expect(err).To(BeNil())
				log.Infof("vip Id: %d", vipId)

				activeStateReached, err := WaitForVIPState(clusterUUID, vipId, "Active")
				Expect(err).To(BeNil())
				Expect(activeStateReached).To(BeTrue(), "VIP did not reach active state in time")

				log.Infof("VIP %s reached active state", publicVip01Name)
				publicVipPrevStepOk = true
			})
		})

		Context("Update security rule for publiclb01", func() {
			It("Update security rule for publiclb01", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				if !publicVipPrevStepOk {
					log.Infof("Skipping security rules tests because previous test failed")
					Skip("Skipping security rules tests because previous test failed")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				internalIp, _, _, err := GetSecRulesInfoByName(clusterUUID, publicVip01Name)
				Expect(err).To(BeNil())

				var sourceIps = []string{"10.0.0.110"}
				var protocols = []string{"TCP"}
				err = UpdateSecRule(clusterUUID, internalIp, sourceIps, publicVip01Port, protocols)
				Expect(err).To(BeNil())

				log.Infof("Security rule for VIP %s was created successfully", publicVip01Name)

				stateReached, err := WaitForSecRuleState(clusterUUID, publicVip01Name, internalIp, publicVip01Port, "Active")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())

				log.Infof("Security rule for VIP %s reached Active state", publicVip01Name)

				// Check that security rule was updated
				ruleExists, err := CheckSecRuleExists(clusterUUID, publicVip01Name, internalIp, sourceIps, protocols, publicVip01Port)
				Expect(err).To(BeNil())
				Expect(ruleExists).To(BeTrue())
			})
		})

		Context("Second security rule update for load balancer publiclb01", func() {
			It("Second security rule update for load balancer publiclb01", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				if !publicVipPrevStepOk {
					log.Infof("Skipping security rules tests because previous test failed")
					Skip("Skipping security rules tests because previous test failed")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				internalIp, _, _, err := GetSecRulesInfoByName(clusterUUID, publicVip01Name)
				Expect(err).To(BeNil())

				var sourceIps = []string{"10.0.0.110", "10.0.0.170"}
				var protocols = []string{"TCP"}
				err = UpdateSecRule(clusterUUID, internalIp, sourceIps, publicVip01Port, protocols)
				Expect(err).To(BeNil())

				log.Infof("Security rule for VIP %s was created successfully", publicVip01Name)

				stateReached, err := WaitForSecRuleState(clusterUUID, publicVip01Name, internalIp, publicVip01Port, "Active")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())

				log.Infof("Security rule for VIP %s reached Active state", publicVip01Name)

				// Check that security rule was updated
				ruleExists, err := CheckSecRuleExists(clusterUUID, publicVip01Name, internalIp, sourceIps, protocols, publicVip01Port)
				Expect(err).To(BeNil())
				Expect(ruleExists).To(BeTrue(), "Security rule does not have expected values")
			})
		})

		Context("Delete security rule for publiclb01", func() {
			It("Delete security rule for publiclb01 ", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				if !publicVipPrevStepOk {
					log.Infof("Skipping security rules tests because previous test failed")
					Skip("Skipping security rules tests because previous test failed")
				}

				log.Infof("Deleting security rule for publiclb01")
				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				internalIp, vipId, publicVip01Port, err := GetSecRulesInfoByName(clusterUUID, publicVip01Name)
				Expect(err).To(BeNil())

				_, err = DeleteSecRule(clusterUUID, vipId)
				Expect(err).To(BeNil())

				stateReached, err := WaitForSecRuleState(clusterUUID, publicVip01Name, internalIp, publicVip01Port, "Not Specified")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())
			})
		})

		Context("Delete Load balancer publiclb01", func() {
			It("Delete Load balancer publiclb01", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				log.Infof("Deleting load balancer publiclb01")
				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				_, vipId, _, err := GetSecRulesInfoByName(clusterUUID, publicVip01Name)
				Expect(err).To(BeNil())

				err = DeleteVIP(clusterUUID, vipId)
				Expect(err).To(BeNil())

				// Wait for deletion
				deletedStateReached, err := WaitForVIPState(clusterUUID, vipId, "")
				Expect(err).To(BeNil())
				Expect(deletedStateReached).To(BeTrue(), "VIP did not reach target in time")
			})
		})

		Context("Update security rule for public-apiserver", func() {
			It("Update security rule for public-apiserver", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				publicVipPrevStepOk = false

				internalIp, _, vipPort, err := GetSecRulesInfoByName(clusterUUID, defaultPublicVipName)
				defaultPublicVipPort = vipPort
				Expect(err).To(BeNil())

				var sourceIps = []string{"10.0.0.210"}
				var protocols = []string{"TCP"}
				err = UpdateSecRule(clusterUUID, internalIp, sourceIps, defaultPublicVipPort, protocols)
				Expect(err).To(BeNil())

				log.Infof("Security rule for VIP %s was created successfully", defaultPublicVipName)

				stateReached, err := WaitForSecRuleState(clusterUUID, defaultPublicVipName, internalIp, defaultPublicVipPort, "Active")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())

				log.Infof("Security rule for VIP %s reached Active state", defaultPublicVipName)

				// Check that security rule was updated
				ruleExists, err := CheckSecRuleExists(clusterUUID, defaultPublicVipName, internalIp, sourceIps, protocols, defaultPublicVipPort)
				Expect(err).To(BeNil())
				Expect(ruleExists).To(BeTrue())

				publicVipPrevStepOk = true
			})
		})

		Context("Second security rule update for public-apiserver", func() {
			It("Second security rule update for public-apiserver", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				if !publicVipPrevStepOk {
					log.Infof("Skipping security rules tests because previous test failed")
					Skip("Skipping security rules tests because previous test failed")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				internalIp, _, defaultPublicVipPort, err := GetSecRulesInfoByName(clusterUUID, defaultPublicVipName)
				Expect(err).To(BeNil())

				var sourceIps = []string{"10.0.0.210", "10.0.0.220"}
				var protocols = []string{"TCP"}

				err = UpdateSecRule(clusterUUID, internalIp, sourceIps, defaultPublicVipPort, protocols)
				Expect(err).To(BeNil())

				log.Infof("Security rule for VIP %s was created successfully", defaultPublicVipName)

				stateReached, err := WaitForSecRuleState(clusterUUID, defaultPublicVipName, internalIp, defaultPublicVipPort, "Active")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())

				log.Infof("Security rule for VIP %s reached Active state", defaultPublicVipName)

				// Check that security rule was updated
				ruleExists, err := CheckSecRuleExists(clusterUUID, defaultPublicVipName, internalIp, sourceIps, protocols, defaultPublicVipPort)
				Expect(err).To(BeNil())
				Expect(ruleExists).To(BeTrue(), "Security rule does not have expected values")
			})
		})

		Context("Delete security rule for public-apiserver", func() {
			It("Delete security rule for public-apiserver", func() {
				if !hostConfig.RunSecRulesTests {
					Skip("Skipping security rules tests because runSecRulesTests flag is not set")
				}

				if !publicVipPrevStepOk {
					log.Infof("Skipping security rules tests because previous test failed")
					Skip("Skipping security rules tests because previous test failed")
				}

				log.Infof("Generating new bearer token")
				bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
				Expect(err).To(BeNil())

				internalIp, vipId, defaultPublicVipPort, err := GetSecRulesInfoByName(clusterUUID, defaultPublicVipName)
				Expect(err).To(BeNil())

				_, err = DeleteSecRule(clusterUUID, vipId)
				Expect(err).To(BeNil())

				stateReached, err := WaitForSecRuleState(clusterUUID, defaultPublicVipName, internalIp, defaultPublicVipPort, "Not Specified")
				Expect(err).To(BeNil())
				Expect(stateReached).To(BeTrue())
			})
		})

		Context("Create Private Load Balancer", func() {
			It("Should create a private load balancers to a cluster only if create_vip is true", func() {
				if hostConfig.CreateVIP && clusterUUID != "" && clusterState == "Active" {
					log.Infof("==========Creating Load Balancer===========")
					// If cluster is in Active State, we proceed to create load balancer
					// create io reader for create load balancer request

					// check number of load balancers for for the cluster and account ID
					getVIPSUrlUpdated := fmt.Sprintf(getVIPSUrl, clusterUUID)
					data, err = SendHttpRequest(getVIPSUrlUpdated, getMethod, nil, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					// parse the response
					var getVIPsResp pb.GetVipsResponse
					json.Unmarshal(data, &getVIPsResp)

					if len(getVIPsResp.Response) >= 2 {
						log.Errorf("Skipping create Load Balancer as maximum load balancer limit of 2 has been reached.")
						Skip("Skipping create Load Balancer as maximum load balancer limit of 2 has been reached.")
					}

					requestData, err := ReadRequestData("requests/create_load_balancer_request.json")
					// assert error to be nil
					Expect(err).To(BeNil())

					type createVIPType struct {
						Description string `json:"description"`
						Viptype     string `json:"viptype"`
						Name        string `json:"name"`
						Port        int32  `json:"port"`
					}
					var createVIPDetails createVIPType

					json.Unmarshal(requestData, &createVIPDetails)
					// add random number to the load balancer name to make it unique
					createVIPDetails.Name = randomize(createVIPDetails.Name)
					vipName = createVIPDetails.Name

					req, err := json.Marshal(createVIPDetails)
					Expect(err).To(BeNil())
					body := bytes.NewReader(req)

					createVIPSUrl = fmt.Sprintf(createVIPSUrl, clusterUUID)

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}

					// send request
					data, err = SendHttpRequest(createVIPSUrl, postMethod, body, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())

					var createVipResponse pb.VipResponse
					json.Unmarshal(data, &createVipResponse)
					vipUUID = createVipResponse.Vipid

					log.Infof("VIP will be created with name: %s, ID: %d", vipName, vipUUID)

					// assert if load balancer creation call was success
					Expect(createVipResponse.Name).ToNot(Equal(""))
					Expect(createVipResponse.Vipid).ToNot(Equal(""))

					// If load balancer is not active get latest ILB state and wait for 5 mins to get the Load balancer state to active
					// This 5 minutes value is configured in config yaml
					noOfIterations := hostConfig.CreateILBTimeOutInMinutes * 6
					if createVipResponse.Vipstate != "Active" {
						var vipResponse pb.GetVipResponse
						// Get the VIP status based on cluster UUID and VIP ID

						getVIPByIDurl = fmt.Sprintf(getVIPByIDurl, clusterUUID, strconv.Itoa(int(createVipResponse.Vipid)))
						log.Infof("======Fetching Load Balancer Details==========")
						startTime := time.Now()
						for i := 0; i < noOfIterations; i++ {
							// check if bearerToken has expired
							if isTokenExpired(bearerTokenExpiry) {
								// create new bearer Token and store in config file
								log.Infof("Bearer Token Expired, generating new Bearer Token")
								bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
								Expect(err).To(BeNil())
							}

							// send request
							data, err = SendHttpRequest(getVIPByIDurl, getMethod, nil, bearerToken)
							Expect(err).To(BeNil())
							Expect(data).ToNot(BeNil())
							json.Unmarshal(data, &vipResponse)

							if vipResponse.Vipstate == "Active" {
								log.Infof("Load Balancer details with name %s is currently %s", vipResponse.Name, vipResponse.Vipstate)
								log.Infof("Load Balancer provisioned successfully")
								break
							}

							time.Sleep(10 * time.Second)

							// Forcefully kill the process if load balancer is taking more than 5 minutes to get Active
							endTime := time.Now()
							elapsedTime := endTime.Sub(startTime)
							if elapsedTime.Minutes() > float64(hostConfig.CreateILBTimeOutInMinutes) {
								log.Infof("Load Balancer with ID: %s is currently: %s", &vipResponse.Vipid, vipResponse.Vipstatus)
								log.Errorf("=========Load Balancer provisioning failed=========")
								Fail(fmt.Sprintf("Expected VIP Status to be active in %d but current status is %s", hostConfig.CreateILBTimeOutInMinutes, vipResponse.Vipstatus)) // this causes termination
							}
						}
					}
				}
				if hostConfig.CreateVIP && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("Skipping Create VIP test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState)
					Skip(fmt.Sprintf("Skipping Create VIP test since either cluster uuid : %s or cluster state : %s are empty.", clusterUUID, clusterState))
				}
				if !hostConfig.CreateVIP {
					log.Infof("Skipping Create VIP test since the Create VIP flag is false.")
					Skip("Skipping Create VIP test since the Create VIP flag is false.")
				}
			})
		})

		Context("Get Load Balancer Information", func() {
			It("Should get list of load balancers associated to a cluster", func() {
				if userEmail == noAdminAPIUserEmail {
					Skip("Skipping test because account does not have admin access")
				}

				if clusterUUID != "" {
					//Get Admin URL to get list of all load balancers associated to a cluster
					getAdminVIPSUrl = fmt.Sprintf(getAdminVIPSUrl, clusterUUID)
					var ilbResponse pb.LoadBalancers
					log.Infof("======Get Load Balancer Details Flow==========")

					// check if Admin Token has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new admin Token and store in config file
						log.Infof("Admin Token Expired, generating new Admin Token")
						adminToken, adminTokenExpiry = auth_admin.Get_Azure_Admin_Bearer_Token(environment)
					}

					// send request
					data, err = SendHttpRequest(getAdminVIPSUrl, getMethod, nil, adminToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					json.Unmarshal(data, &ilbResponse)

					if len(ilbResponse.Lbresponses) > 0 {
						log.Infof("======Load Balancer Details==========")
						for _, ilb := range ilbResponse.Lbresponses {
							log.Infof("<======ILB Name==========><=====IP Address========><=====ILB Status====><=====ILB Type====>")
							log.Infof("<======%s==========><=====%s========><=====%s====><=====%s====>", ilb.Lb.Lbname, ilb.Lb.Vip, ilb.Lb.Status, ilb.Lb.Viptype)
						}

					} else {
						log.Errorf("There are no public or private load balancers associated with the given cluster: %s", clusterUUID)
					}
				}
				if clusterUUID == "" {
					log.Errorf("Skipping Get Load Balancer test since either cluster uuid : %s is empty.", clusterUUID)
					Skip(fmt.Sprintf("Skipping Get Load Balancer test since either cluster uuid : %s is empty.", clusterUUID))
				}
			})
		})

		Context("Get KubeConfig", func() {
			It("It should download the kubeconfig if cluster is available and kubeconfig flag in ON", func() {
				if hostConfig.DownloadKubeConfig && clusterUUID != "" && clusterState == "Active" {
					log.Infof("Get Kubeconfig - ClusterName: %s, ClusterUUID: %s, ClusterState: %s", clusterName, clusterUUID, clusterState)

					//Get the Kubeconfig Information
					getKubeConfigUrlUpdated := fmt.Sprintf(getKubeConfigUrl, clusterUUID)
					var kubeResponse pb.GetKubeconfigResponse
					log.Infof("======Get Kubeconfig Details Flow==========")

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}

					// send request
					data, err = SendHttpRequest(getKubeConfigUrlUpdated, getMethod, nil, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					json.Unmarshal(data, &kubeResponse)

					if kubeResponse.Kubeconfig == "" {
						log.Errorf("Failed to fetch Kubeconfig from the Kubeconfig API")
						Fail("Get Kubeconfig failed because kubeconfig is empty")
					}

					if kubeResponse.Kubeconfig != "" {
						kubeConfString := kubeResponse.Kubeconfig

						// Get working directory to store kubeconfig file in
						workingDir, err := os.Getwd()
						Expect(err).To(BeNil())
						kubeDir := filepath.Dir(workingDir)
						log.Infof("Current working directory: %s", kubeDir)

						//Define the local file to store the content
						localFilePath := filepath.Join(kubeDir, clusterUUID+"_config"+".yaml")

						//Write Kubeconfig content to the a file
						log.Infof("Target kubeconfig file path: %s", localFilePath)
						err = ioutil.WriteFile(localFilePath, []byte(kubeConfString), 0644)
						Expect(err).To(BeNil())

						log.Infof("Kubeconfig is loaded into file successfully and is stored inside directory : %s with file name : %s", kubeDir, localFilePath)
					}
				}
				if hostConfig.DownloadKubeConfig && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("The Cluster UUID is empty or Cluster State is not active. Cannot download kubeconfig for the cluster")
					Skip("Skipping the Get Kubeconfig flow since the cluster UUID is empty or cluster state is not active")
				}
				if !hostConfig.DownloadKubeConfig {
					log.Infof("The Download Kubeconfig flag is OFF")
					Skip("Skipping the Get Kubeconfig flow since the DownloadKubeConfig is OFF")
				}
			})
		})

		Context("Run Storage Test", func() {
			It("It should download the kubeconfig if cluster is available, set the KUBECONFIG path and run the storage tests if the RunStorageTests flag is ON.", func() {
				if hostConfig.RunStorageTests && clusterUUID != "" && clusterState == "Active" {
					//Get the Kubeconfig Information
					getKubeConfigUrlUpdated := fmt.Sprintf(getKubeConfigUrl, clusterUUID)
					var kubeResponse pb.GetKubeconfigResponse
					log.Infof("======Get Kubeconfig Details Flow==========")

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}

					// send request
					data, err = SendHttpRequest(getKubeConfigUrlUpdated, getMethod, nil, bearerToken)
					Expect(err).To(BeNil())
					Expect(data).ToNot(BeNil())
					json.Unmarshal(data, &kubeResponse)

					if kubeResponse.Kubeconfig == "" {
						log.Errorf("Failed to fetch Kubeconfig from the Kubeconfig API")
						Fail("Get Kubeconfig failed because kubeconfig is empty")
					}
					var localFilePath string
					if kubeResponse.Kubeconfig != "" {
						kubeConfString := kubeResponse.Kubeconfig

						// Get working directory to store kubeconfig file in
						workingDir, err := os.Getwd()
						Expect(err).To(BeNil())
						kubeDir := filepath.Dir(workingDir)
						log.Infof("Current working directory: %s", kubeDir)

						//Define the local file to store the content
						localFilePath = filepath.Join(kubeDir, clusterUUID+"_config"+".yaml")

						//Write Kubeconfig content to the a file
						err = ioutil.WriteFile(localFilePath, []byte(kubeConfString), 0644)
						Expect(err).To(BeNil())

						log.Infof("Kubeconfig is loaded into file successfully and is stored inside directory : %s with file name : %s", kubeDir, localFilePath)

						log.Infof("======Running Storage Tests==========")
						log.Infof("localFilePath: %s", localFilePath)

						// Define the command to run
						cmd := exec.Command("./run_storage_test.sh", localFilePath)

						if pipelineRun {
							cmd = exec.Command("/tmp/iks_storage_test/storage-test.sh", localFilePath)
							cmd.Dir = ("/tmp/iks_storage_test")
							envvar1 := fmt.Sprintf("KUBECONFIG=%s", localFilePath)
							log.Infof("envvar1: %s", envvar1)
							cmd.Env = append(os.Environ(), envvar1)
						}

						stdout, err := cmd.StdoutPipe()
						if err != nil {
							fmt.Println(err)
						}

						err = cmd.Start()
						fmt.Println("The command is running")
						if err != nil {
							fmt.Println(err)
						}

						// print the output of the subprocess
						testPass := false
						scanner := bufio.NewScanner(stdout)
						for scanner.Scan() {
							m := scanner.Text()
							if strings.Contains(m, "Global Test Status: PASS") {
								testPass = true
							}
							fmt.Println(m)
						}

						cmd.Wait()
						Expect(testPass).To(BeTrue(), "One or more storage tests failed.  See console output for more details.")
					}
				}
				if hostConfig.RunStorageTests && (clusterUUID == "" || clusterState != "Active") {
					log.Errorf("The Cluster UUID is empty or Cluster State is not active. Cannot download kubeconfig for the cluster")
					Skip("Skipping the Get Kubeconfig flow since the cluster UUID is empty or cluster state is not active")
				}
				if !hostConfig.RunStorageTests {
					log.Infof("The Run Storage Test flag is OFF")
					Skip("Skipping the Run Storage Test flow since the RunStorageTests is OFF")
				}

			})
		})

		Context("Delete a specific NodeGroup", func() {
			It("Should delete NodeGroup for the given cluster ID and node Group ID", func() {
				// check if the clusterID is valid
				if hostConfig.DeleteSpecificNodeGroup && clusterUUID != "" {
					log.Infof("======Delete Specific NodeGroup Flow==========")
					if clusterUUID == "" {
						log.Errorf("Cluster UUID cannot be empty while deleting specific NodeGroup for a cluster")
						Expect(true).To(Equal(false))
					}

					log.Infof("=========Initiating Delete Specific Node Group for cluster=========")

					type existingNodeGroupType struct {
						NodeGroupName string `json:"nodeGroupName"`
						NodeGroupUUID string `json:"nodeGroupUUID"`
					}
					var existingNodeGroupDetails existingNodeGroupType

					if pipelineRun {
						log.Infof("Deleting node group - Node group name: %s, Node group UUID: %s", nodeGroupName, nodeGroupUUID)
						existingNodeGroupDetails.NodeGroupName = nodeGroupName
						existingNodeGroupDetails.NodeGroupUUID = nodeGroupUUID

					} else {
						// obtain the node group id to be deleted and make delete request
						requestData, err := ReadRequestData("requests/existing_node_group_details.json")
						Expect(err).To(BeNil())

						json.Unmarshal(requestData, &existingNodeGroupDetails)
					}

					// check if nodeGroupUUID is empty
					if existingNodeGroupDetails.NodeGroupUUID == "" {
						log.Errorf("Node Group UUID cannot be empty while deleting NodeGroup for a cluster")
						Expect(true).To(Equal(false))
					} else {
						deleteNodeGroupByIdUrlUpdated := fmt.Sprintf(deleteNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
						deleteNodeGroupRequest := `{}`

						var deleteRequestBuffer bytes.Buffer
						deleteRequestBuffer.WriteString(deleteNodeGroupRequest)
						_, err = SendHttpRequest(deleteNodeGroupByIdUrlUpdated, deleteMethod, &deleteRequestBuffer, bearerToken)

						Expect(err).To(BeNil())

						log.Infof("============Getting Cluster Node Group Status===========")
						isDeletedSuccessfully := false
						isUpdating := false
						for i := 0; i < 60; i++ {
							// wait for 5 seconds before querying delete confirmation
							time.Sleep(5 * time.Second)
							// again send get request to confirm that cluster got deleted
							getNodeGroupByIdUrlUpdated := fmt.Sprintf(getNodeGroupByIdUrl, clusterUUID, existingNodeGroupDetails.NodeGroupUUID)
							data, err = SendHttpRequest(getNodeGroupByIdUrlUpdated, getMethod, nil, bearerToken)
							Expect(err).To(BeNil())

							var getNodeGroupByIdResponse pb.NodeGroupResponseForm
							err = json.Unmarshal(data, &getNodeGroupByIdResponse)
							Expect(err).To(BeNil())

							if isDeletedSuccessfully == false {
								if getNodeGroupByIdResponse.Nodegroupstatus == nil {
									log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
									log.Infof("<======%s==========><====%s=======><=====Deleted====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName)
									isDeletedSuccessfully = true
								} else if isUpdating == false {
									log.Infof("<======Nodegroup Id==========><====Node Group Name=======><=====Node Group Status====>")
									log.Infof("<======%s==========><====%s=======><=====%s====>", existingNodeGroupDetails.NodeGroupUUID, existingNodeGroupDetails.NodeGroupName, getNodeGroupByIdResponse.Nodegroupstatus)
									isUpdating = true
								}
							} else {
								break
							}

						}

						if !isDeletedSuccessfully {
							Fail("Expected Node Group to be deleted in 5 minutes but took longer than expected. Need to Process Manual Cleanup")
						}
					}

					log.Infof("=============Ending=============")
				}

				if hostConfig.DeleteSpecificNodeGroup && clusterUUID == "" {
					//Todo for existing cluster
					log.Infof("============== Cluster UUID is empty=============")
					log.Infof("==============If cluster already exist, Node Group cannot be deleted due to some technical issue, may need manual cleanup=============")
					log.Infof("=============Ending=============")
					Skip(fmt.Sprintf("Skipping Delete Specific Node Group test since either cluster uuid : %s or cluster state : %s are empty", clusterUUID, clusterState))
				}
				if !hostConfig.DeleteSpecificNodeGroup {
					log.Infof("Skipping delete Node Group test since delete node group flag is false.")
					Skip("Skipping delete Node Group test since delete node group flag is false.")
				}
			})
		})

		Context("Delete a specific VIP", func() {
			It("Should delete VIP for the given cluster ID and vip ID", func() {
				if !hostConfig.CreateVIP {
					Skip("Skipping test because no VIP was created")
				}

				log.Infof("Existing VIP information: vipName: %s, vipUUID: %d", vipName, vipUUID)
				log.Infof("Sleeping for 30 sec to allow cluster state to update...")
				time.Sleep((30 * time.Second))

				// check if the clusterID is valid
				if hostConfig.DeleteSpecificVIP && clusterUUID != "" {
					log.Infof("======Delete Specific VIP Flow==========")
					if clusterUUID == "" {
						log.Errorf("Cluster UUID cannot be empty while deleting specific VIP for a cluster")
						Expect(true).To(Equal(false))
					}

					log.Infof("=========Initiating Delete Specific VIP for cluster=========")

					type existingVIPType struct {
						VIPName string `json:"vipName"`
						VIPUUID string `json:"vipUUID"`
					}
					var existingVIPDetails existingVIPType

					if pipelineRun {
						existingVIPDetails.VIPName = vipName
						existingVIPDetails.VIPUUID = strconv.Itoa(int(vipUUID))
					} else {
						// obtain the vip id to be deleted and make delete request
						requestData, err := ReadRequestData("requests/existing_load_balancer_details.json")
						Expect(err).To(BeNil())

						json.Unmarshal(requestData, &existingVIPDetails)
					}

					log.Infof("Deleting VIP with name: %s, ID: %s", existingVIPDetails.VIPName, existingVIPDetails.VIPUUID)

					// check if vipUUID is empty
					if existingVIPDetails.VIPUUID == "" {
						log.Errorf("VIP UUID cannot be empty while deleting VIP for a cluster")
						Expect(true).To(Equal(false))
					} else {
						deleteVIPByIdUrlUpdated := fmt.Sprintf(deleteVIPByIdUrl, clusterUUID, existingVIPDetails.VIPUUID)
						deleteVIPRequest := `{}`

						var deleteRequestBuffer bytes.Buffer
						deleteRequestBuffer.WriteString(deleteVIPRequest)
						_, err = SendHttpRequest(deleteVIPByIdUrlUpdated, deleteMethod, &deleteRequestBuffer, bearerToken)

						Expect(err).To(BeNil())

						log.Infof("============Getting Cluster VIP Status===========")
						isDeletedSuccessfully := false
						isUpdating := false
						for i := 0; i < 20; i++ {
							// wait for 15 seconds before querying delete confirmation
							time.Sleep(15 * time.Second)
							// again send get request to confirm that VIP got deleted
							// getVIPByIdUrlUpdated := fmt.Sprintf(getVIPByIDurl, clusterUUID, existingVIPDetails.VIPUUID)
							data, err = SendHttpRequest(getVIPByIDurl, getMethod, nil, bearerToken)
							Expect(err).To(BeNil())

							var getVIPByIdResponse pb.GetVipResponse
							err = json.Unmarshal(data, &getVIPByIdResponse)
							Expect(err).To(BeNil())

							if isDeletedSuccessfully == false {
								if getVIPByIdResponse.Vipstatus == nil {
									log.Infof("<======VIP Id==========><====VIP Name=======><=====VIP Status====>")
									log.Infof("<======%s==========><====%s=======><=====Deleted====>", existingVIPDetails.VIPUUID, existingVIPDetails.VIPName)
									isDeletedSuccessfully = true
								} else if isUpdating == false {
									log.Infof("<======VIP Id==========><====VIP Name=======><=====VIP Status====>")
									log.Infof("<======%s==========><====%s=======><=====%s====>", existingVIPDetails.VIPUUID, existingVIPDetails.VIPName, getVIPByIdResponse.Vipstatus)
									isUpdating = true
								}
							} else {
								break
							}

						}

						if !isDeletedSuccessfully {
							Fail("Expected VIP to be deleted in 5 minutes but took longer than expected. Need to Process Manual Cleanup")
						}

					}

					log.Infof("=============Ending=============")
				}

				if hostConfig.DeleteSpecificVIP && clusterUUID == "" {
					//Todo for existing cluster
					log.Infof("============== Cluster UUID are empty=============")
					log.Infof("==============If cluster already exist, VIP cannot be deleted due to some technical issue, may need manual cleanup=============")
					log.Infof("=============Ending=============")
					Skip(fmt.Sprintf("Skipping Delete Specific VIP test since either cluster uuid : %s or cluster state : %s are empty", clusterUUID, clusterState))
				}
				if !hostConfig.DeleteSpecificVIP {
					log.Infof("Skipping delete VIP test since delete VIP flag is false.")
					Skip("Skipping delete VIP test since delete VIP flag is false.")
				}
			})
		})

		Context("Delete Cluster", func() {
			It("Should delete cluster and only if cluster UUID is not empty", func() {
				if hostConfig.DeleteCluster && clusterUUID != "" {
					log.Infof("======Delete Cluster Flow==========")
					if clusterUUID == "" {
						log.Errorf("Cluster UUID cannot be empty while deleting a cluster")
						Expect(true).To(Equal(false))
					}

					// if clusterState != "Active" {
					// 	log.Errorf("Cluster is not in actionable state and cannot perform delete operation")
					// 	Expect("Active").To(Equal(clusterState))
					// }

					log.Infof("=========Initiating Delete Cluster=========")

					getClusterUrlUpdated := fmt.Sprintf(getClusterUrl, clusterUUID)

					// at the end cleanup
					deleteClusterRequest := `{}`
					deleteClusterUrl = fmt.Sprintf(deleteClusterUrl, clusterUUID)
					var deleteRequestBuffer bytes.Buffer
					deleteRequestBuffer.WriteString(deleteClusterRequest)

					// check if bearerToken has expired
					if isTokenExpired(bearerTokenExpiry) {
						// create new bearer Token and store in config file
						log.Infof("Bearer Token Expired, generating new Bearer Token")
						bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
						Expect(err).To(BeNil())
					}
					_, err := SendHttpRequest(deleteClusterUrl, deleteMethod, &deleteRequestBuffer, bearerToken)
					Expect(err).To(BeNil())

					log.Infof("============Getting Cluster Status===========")
					isDeletedSuccessfully := false
					for i := 0; i < 20; i++ {
						// wait for 15 seconds before querying delete confirmation
						time.Sleep(15 * time.Second)

						// check if bearerToken has expired
						if isTokenExpired(bearerTokenExpiry) {
							// create new bearer Token and store in config file
							log.Infof("Bearer Token Expired, generating new Bearer Token")
							bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
							Expect(err).To(BeNil())
						}

						// again send get request to confirm that cluster got deleted
						data, err = SendHttpRequest(getClusterUrlUpdated, getMethod, nil, bearerToken)
						Expect(err).To(BeNil())

						var deleteSuccess ErrorStruct
						err = json.Unmarshal(data, &deleteSuccess)
						Expect(err).To(BeNil())

						if deleteSuccess.Message != "" {
							//already delete complete
							Expect(deleteSuccess.Message).To(Equal(fmt.Sprintf("Cluster not found: %s", clusterUUID)))
							log.Infof("===========Cluster was deleted successfully============")
							isDeletedSuccessfully = true
							break
						}
					}

					// deletion may be in progress
					// again send get request to confirm that cluster got deleted
					if !isDeletedSuccessfully {
						var getClusterResponse pb.ClusterResponseForm

						// check if bearerToken has expired
						if isTokenExpired(bearerTokenExpiry) {
							// create new bearer Token and store in config file
							log.Infof("Bearer Token Expired, generating new Bearer Token")
							bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(accountEmail)
							Expect(err).To(BeNil())
						}

						data, err = SendHttpRequest(getClusterUrlUpdated, getMethod, nil, bearerToken)
						Expect(err).To(BeNil())
						err = json.Unmarshal(data, &getClusterResponse)
						Expect(err).To(BeNil())

						if getClusterResponse.Clusterstatus.State == "DeletePending" {
							log.Infof("==============Cluster still in DeletePending state, may need manual cleanup=============")
						}
						if getClusterResponse.Clusterstatus.State == "Deleting" {
							log.Infof("==============Cluster still in Deleting state. Please wait for while and check the status, may need manual cleanup=============")
						}
						Fail("Expected Cluster to be deleted in 5 minutes but took longer than expected. Need to Process Manual Cleanup")
					}

					log.Infof("=============Ending=============")
				}
				if hostConfig.DeleteCluster && clusterUUID == "" {
					//Todo for existing cluster
					log.Infof("============== Cluster UUID are empty=============")
					log.Infof("==============If cluster already exist, Cluster cannot be deleted due to some technical issue, may need manual cleanup=============")
					log.Infof("=============Ending=============")
					Skip(fmt.Sprintf("Skipping Delete Cluster test since either cluster uuid : %s or cluster state : %s are empty", clusterUUID, clusterState))
				}
				if !hostConfig.DeleteCluster {
					log.Infof("Skipping Delete Cluster test since delete cluster flag is false.")
					Skip("Skipping Delete Cluster test since delete cluster flag is false.")
				}
			})
		})

	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			log.Errorf("==========Test case failed with response %s =============", string(data))
		} else {
			//Todo log the each state of the cluster process
		}
	})
})

type ErrorStruct struct {
	Message string `json:"message,omitempty"`
}
