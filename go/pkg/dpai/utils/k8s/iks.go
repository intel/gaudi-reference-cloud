// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) GetIksClient(conf *config.Config) error {
	dialOptions := []grpc.DialOption{}
	dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
	// 	InsecureSkipVerify: true,
	// })),
	// grpc.WithBlock(),

	grpcClientConn, err := grpcutil.NewClient(context.Background(), conf.GrpcAPIServerAddr, dialOptions...)
	if err != nil {
		return err
	}
	k.GrpcClientConn = grpcClientConn
	k.IksClient = pb.NewIksClient(grpcClientConn)
	k.SshClient = pb.NewSshPublicKeyServiceClient(grpcClientConn)
	k.VnetClient = pb.NewVNetServiceClient(grpcClientConn)

	if k.ClusterID.Clusteruuid != "" {
		err = k.GetK8sClientSet()
		if err != nil {
			return err
		}
	}

	return nil
}

// IKS Cluster
func (k *K8sClient) CreateIKSCluster(req *pb.ClusterRequest) (*pb.ClusterResponseForm, error) {
	request, err := k.IksClient.CreateNewCluster(context.Background(), req)
	if err != nil {
		log.Printf("Error submitting the request to create IKS Cluster. Error message: %+v", err)
		return nil, err
	}

	// Get cluster status
	inProgress := true
	errorCount := 0
	for inProgress {
		status, err := k.IksClient.GetClusterStatus(context.Background(), &pb.ClusterID{
			Clusteruuid:    request.Uuid,
			CloudAccountId: req.CloudAccountId,
		})
		if err != nil {
			errorCount += 1
			log.Printf("Error getting the cluster status. Error count: %d. Error message: %+v", errorCount, err)
			// }

			// The layout for parsing the time string
			// layout := "2006-01-02 15:04:05 -0700 MST"
			// parsedTime, err := time.Parse(layout, status.Lastupdate)
			// if err != nil {
			// 	log.Println("Error parsing time:", err)
			// 	errorCount += 1
			// }
		} else if status.Errorcode != 0 {
			log.Printf("Failed to create the cluster with the error %+v", status)
			return nil, fmt.Errorf("failed to create the cluster with the error %+v", status)
		} else if status.State == "Active" {
			log.Printf("Successfully created the cluster. Status: %+v", status)
			inProgress = false
			// } else if time.Since(parsedTime).Minutes() > 60 {
			// 	// Parse the string into a time.Time object
			// 	log.Printf("No status update for past 60 mins. Last update ts: %s . Current Status: %+v", parsedTime, status)
			// 	errorCount += 1
		}

		if errorCount > 5 {
			return nil, fmt.Errorf("error fetching the cluster status")
		} else {
			log.Printf("Cluster provisioning is still in progress. Check status after 30 seconds. Current Status: %+v", status)
			time.Sleep(30 * time.Second)
		}

	}

	// Get Cluster
	cluster, err := k.IksClient.GetCluster(context.Background(), &pb.ClusterID{
		Clusteruuid:    request.Uuid,
		CloudAccountId: req.CloudAccountId,
	})
	if err != nil {
		log.Printf("Error getting the cluster. Error message: %+v", err)
	}

	return cluster, nil
}

func (k *K8sClient) DeleteIKSCluster() error {
	log.Printf("deleting the cluster %+v", k.ClusterID)
	_, err := k.IksClient.DeleteCluster(context.Background(), k.ClusterID)

	if err != nil {
		err = fmt.Errorf("error: Failed to delete the cluster %+v. Error message: %+v", k.ClusterID, err)
	}

	return err

}

// IKS Nodegroup
func (k *K8sClient) CreateNodeGroup(req *pb.CreateNodeGroupRequest) (*pb.NodeGroupResponseForm, error) {

	err := req.ValidateAll()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	vnet, err := k.VnetClient.Search(context.Background(), &pb.VNetSearchRequest{
		Metadata: &pb.VNetSearchRequest_Metadata{
			CloudAccountId: req.CloudAccountId,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error: Failed to get the vnet. Error message: %+v", err)
	}

	req.Vnets = []*pb.Vnet{
		{
			Availabilityzonename:     vnet.Items[0].GetSpec().GetAvailabilityZone(),
			Networkinterfacevnetname: vnet.Items[0].Metadata.GetName(),
		},
	}

	request, err := k.IksClient.CreateNodeGroup(context.Background(), req)
	if err != nil {
		log.Printf("Error submitting the request to create IKS NodeGroup. Error message: %+v", err)
		return nil, err
	}

	// Get cluster status
	inProgress := true
	errorCount := 0
	for inProgress {
		status, err := k.IksClient.GetNodeGroupStatus(context.Background(), &pb.NodeGroupid{
			Clusteruuid:    req.Clusteruuid,
			CloudAccountId: req.CloudAccountId,
			Nodegroupuuid:  request.Nodegroupuuid,
		})
		if err != nil {
			errorCount += 1
			log.Printf("Error getting the NodeGroup status. Error count: %d. Error message: %+v", errorCount, err)
		}

		if status.Errorcode != 0 {
			log.Printf("Failed to create the NodeGroup with the error %+v", status)
			return nil, fmt.Errorf("failed to create the NodeGroup with the error %+v", status)
		} else if status.State == "Active" {
			log.Printf("Successfully created the NodeGroup. Status: %+v", status)
			inProgress = false
		} else if errorCount > 5 {
			return nil, fmt.Errorf("error fetching the NodeGroup status")
		} else {
			log.Printf("NodeGroup provisioning is still in progress. Check status after 30 seconds. Current Status: %+v", status)
			time.Sleep(30 * time.Second)
		}

	}

	// Get NodeGroup
	nodes := true
	nodeGroup, err := k.IksClient.GetNodeGroup(context.Background(), &pb.GetNodeGroupRequest{
		Clusteruuid:    request.Clusteruuid,
		CloudAccountId: req.CloudAccountId,
		Nodegroupuuid:  request.Nodegroupuuid,
		Nodes:          &nodes,
	})
	if err != nil {
		log.Printf("Error getting the Nodegroup. Error message: %+v", err)
	}

	return nodeGroup, nil
}

func (k *K8sClient) GetNodeGroupById(iksClusterUUID string, cloudAccountID string, nodegroupUUID string) (*pb.NodeGroupResponseForm, error) {
	nodeGroups, err := k.IksClient.GetNodeGroups(context.Background(), &pb.GetNodeGroupsRequest{
		Clusteruuid:    iksClusterUUID,
		CloudAccountId: cloudAccountID,
	})

	if err != nil {
		log.Printf("Error fetching nodegroups %+v", err)
		return nil, err
	}

	if nodeGroups == nil {
		return nil, fmt.Errorf("No Matching Nodegroup found for nodegroupId %s in IKS Cluster %s", nodegroupUUID, iksClusterUUID)
	}

	for _, nodeGroup := range nodeGroups.Nodegroups {
		if nodeGroup.Nodegroupuuid == nodegroupUUID {
			return nodeGroup, nil
		}
	}
	msg := fmt.Sprintf("Error No matching nodegroup found with id %s", nodegroupUUID)
	return nil, fmt.Errorf("%s", msg)
}

// Get IKS NodegrupSelctetor for Node affinity or taints deploy services over specifc node
func (k *K8sClient) GetNodeGroupSelectorLabel(nodegroupUUID string, iksClusterUUID string, cloudAccountID string) (*pb.NodeGroupResponseForm, error) {
	nodeGroups, err := k.IksClient.GetNodeGroups(context.Background(), &pb.GetNodeGroupsRequest{
		Clusteruuid:    iksClusterUUID,
		CloudAccountId: cloudAccountID,
	})

	if err != nil {
		log.Printf("Error fetching nodegroups %+v", err)
		return nil, err
	}

	for _, nodeGroup := range nodeGroups.Nodegroups {
		if strings.HasPrefix(nodeGroup.Name, nodegroupUUID) {
			return nodeGroup, nil
		}
	}

	return nil, fmt.Errorf("No Matching Nodegroup Label Nodes found")
}

// Label the IKS Nodes in public tenant namespace, since the GetNodeGroupSelectorLabel only select in control plane subnet
// To deploy services over node with affinity the target ndodes must be labelled
func (k *K8sClient) AddCustomLabelTenantNodegroupNodes(nodegroupUUID string, iksCLusterUUID string,
	clountAccountID string, serviceType string) error {
	nodegroup, err := k.GetNodeGroupSelectorLabel(nodegroupUUID, iksCLusterUUID, clountAccountID)
	if err != nil {
		log.Printf("Error fetching nodegroup %+v", err)
		return err
	}

	nodegorupName := nodegroup.Name

	nodes, err := k.ClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Printf("Error fetching nodes in tenant space %+v", err)
		return err
	}

	for _, node := range nodes.Items {
		for _, nodeLabel := range node.Labels {
			if strings.HasPrefix(nodeLabel, nodegorupName) {
				nodeObj, err := k.ClientSet.CoreV1().Nodes().Get(context.Background(), node.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("failed to get node %s: %+v", node.Name, err)
				}

				nodeObj.Labels["DPAI_SERVICE"] = serviceType

				_, err = k.ClientSet.CoreV1().Nodes().Update(context.Background(), nodeObj, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("failed to update node %s: %+v", node.Name, err)
				}

				log.Printf("Succcessfully updated node %s with label %s", node.Name, nodeLabel)
				break
			}
		}
	}
	return nil
}

func (k *K8sClient) DeleteNodeGroup(id string) error {

	nodeGroupId := &pb.NodeGroupid{
		CloudAccountId: k.ClusterID.CloudAccountId,
		Clusteruuid:    k.ClusterID.Clusteruuid,
		Nodegroupuuid:  id,
	}

	_, err := k.IksClient.DeleteNodeGroup(context.Background(), nodeGroupId)

	if err != nil {
		err = fmt.Errorf("error: Failed to delete the nodegroup %+v. Error message: %+v", nodeGroupId, err)
	}

	return err
}
