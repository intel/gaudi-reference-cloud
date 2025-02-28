// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package compute

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	nodeprovider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/node_provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pborman/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	networkInterfaceName    = "eth0"
	clusterNameLabelKey     = "clusterName"
	nodegroupNameLabelKey   = "nodegroupName"
	nodegroupTypeLabelKey   = "nodegroupType"
	superComputeClusterType = "supercompute"
	storageBackendInterface = "storage0"
)

type ComputeProvider struct {
	ComputeClient                     pb.InstanceServiceClient
	ComputePrivateClient              pb.InstancePrivateServiceClient
	InstanceGroupClient               pb.InstanceGroupServiceClient
	ComputeInstanceGroupPrivateClient pb.InstanceGroupPrivateServiceClient
}

func NewComputeProvider(ctx context.Context, client pb.InstanceServiceClient, privateClient pb.InstancePrivateServiceClient, instanceGroupClient pb.InstanceGroupServiceClient, computeInstanceGroupPrivateClient pb.InstanceGroupPrivateServiceClient) (*ComputeProvider, error) {
	r := &ComputeProvider{
		ComputeClient:                     client,
		ComputePrivateClient:              privateClient,
		InstanceGroupClient:               instanceGroupClient,
		ComputeInstanceGroupPrivateClient: computeInstanceGroupPrivateClient,
	}
	return r, nil
}

// CreateNode is a compute provider interface implementing Instance Service Create Method
func (p *ComputeProvider) CreateNode(ctx context.Context, _ string, nameserver string, gateway string, registrationCmd string, bootstrapScript string, nodegroup privatecloudv1alpha1.Nodegroup) (privatecloudv1alpha1.NodeStatus, error) {
	log := log.FromContext(ctx).WithName("ComputeProvider.CreateNode")

	var nodeStatus privatecloudv1alpha1.NodeStatus

	if len(nodegroup.Spec.VNETS) == 0 {
		return nodeStatus, fmt.Errorf("no vnets found in nodegroup spec")
	}

	log.V(0).Info("Creating node", "Checking Cluster Type", nodegroup.Spec.ClusterType)
	var serviceType pb.InstanceServiceType
	if nodegroup.Spec.ClusterType == superComputeClusterType {
		serviceType = pb.InstanceServiceType_SuperComputingAsAService
	} else {
		serviceType = pb.InstanceServiceType_KubernetesAsAService
	}

	nodeName := nodegroup.Name + "-" + uuid.New()[:5]
	nodeLabels := make(map[string]string, 3)
	nodeLabels[clusterNameLabelKey] = nodegroup.Spec.ClusterName
	nodeLabels[nodegroupTypeLabelKey] = string(nodegroup.Spec.NodegroupType)
	nodeLabels[nodegroupNameLabelKey] = nodegroup.Name

	cloudInit := &bytes.Buffer{}

	userData := nodeprovider.UserData{
		RegistrationCmd: registrationCmd,
		BootstrapScript: strings.Split(bootstrapScript, "\n"),
	}

	if len(nodegroup.Spec.UserDataURL) > 0 {
		userData.DownloadCustomBashScript = fmt.Sprintf("- curl --retry 5 --retry-connrefused --connect-timeout 10 %s -o /usr/local/bin/user-script.sh", nodegroup.Spec.UserDataURL)
		userData.RunCustomBashScript = "- bash /usr/local/bin/user-script.sh"
	}

	if err := nodeprovider.UserDataTemplate.Execute(cloudInit, userData); err != nil {
		return nodeStatus, err
	}

	skipQuotaCheck := false
	spreadConstrains := make([]*pb.TopologySpreadConstraints, 0)
	// If controlplane nodes we want to create them if possible on different availability zones,
	// to provide high availability of the cluster.
	// We also want to skip quota check so that we avoid quota issues for controlplane nodes.
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		skipQuotaCheck = true

		spreadConstrains = append(spreadConstrains, &pb.TopologySpreadConstraints{
			LabelSelector: &pb.LabelSelector{
				MatchLabels: map[string]string{
					clusterNameLabelKey:   nodegroup.Spec.ClusterName,
					nodegroupNameLabelKey: nodegroup.Name,
					nodegroupTypeLabelKey: string(nodegroup.Spec.NodegroupType),
				},
			},
		})
	}

	log.V(0).Info("Creating node", logkeys.NodeName, nodeName)
	instance, err := p.ComputePrivateClient.CreatePrivate(ctx, &pb.InstanceCreatePrivateRequest{
		Metadata: &pb.InstanceMetadataCreatePrivate{
			CloudAccountId: nodegroup.Spec.CloudAccountId,
			Name:           nodeName,
			Labels:         nodeLabels,
			ResourceId:     uuid.NewRandom().String(),
			SkipQuotaCheck: skipQuotaCheck,
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone:  nodegroup.Spec.VNETS[0].AvailabilityZone,
			InstanceType:      nodegroup.Spec.InstanceType,
			MachineImage:      nodegroup.Spec.InstanceIMI,
			SshPublicKeyNames: nodegroup.Spec.SSHKey,
			Interfaces: []*pb.NetworkInterfacePrivate{{
				Name: networkInterfaceName,
				VNet: nodegroup.Spec.VNETS[0].NetworkInterfaceVnet,
			}},
			UserData:                  cloudInit.String(),
			TopologySpreadConstraints: spreadConstrains,
			ServiceType:               serviceType,
		},
	})

	if err != nil {
		return nodeStatus, err
	}

	nodeStatus.Name = instance.Metadata.Name
	nodeStatus.InstanceIMI = instance.Spec.MachineImage
	nodeStatus.CreationTime = metav1.Time{Time: instance.Metadata.CreationTimestamp.AsTime()}
	nodeStatus.LastUpdate = metav1.Time{Time: time.Now()}
	nodeStatus.State = privatecloudv1alpha1.UpdatingNodegroupState

	if instance.Status == nil {
		return nodeStatus, nil
	}

	var ipAddress string
	if len(instance.Status.Interfaces) > 0 {
		if len(instance.Status.Interfaces[0].Addresses) > 0 {
			ipAddress = instance.Status.Interfaces[0].Addresses[0]
		}
	}
	nodeStatus.IpAddress = ipAddress

	// Nodegroup : active, updating, error, deleting
	// Phase: PROVISIONING ,READY,STOPPING, STOPPED, TERMINATING, FAILED

	nodeStatus.Message = instance.Status.Message

	switch {
	case instance.Status.Phase == pb.InstancePhase_Provisioning:
		nodeStatus.State = "Updating"
		nodeStatus.Message = "Provisioning node"
	case instance.Status.Phase == pb.InstancePhase_Ready:
		nodeStatus.State = "Active"
		nodeStatus.Message = "Configuring node"
	case instance.Status.Phase == pb.InstancePhase_Terminating:
		nodeStatus.State = "Deleting"
		nodeStatus.Message = "Deleting node"
	case instance.Status.Phase == pb.InstancePhase_Failed:
		nodeStatus.State = "Error"
	case instance.Status.Phase == pb.InstancePhase_Stopping:
		nodeStatus.State = "Updating"
	case instance.Status.Phase == pb.InstancePhase_Stopped:
		nodeStatus.State = "Updating"
	default:
		nodeStatus.State = "Updating"
		nodeStatus.Message = "Provisioning node"
	}

	return nodeStatus, nil
}

// GetNodes is a compute provider interface implementing Instance Service Search Method
func (p *ComputeProvider) GetNodes(ctx context.Context, selector string, cloudaccountid string) ([]privatecloudv1alpha1.NodeStatus, error) {
	nodesStatus := make([]privatecloudv1alpha1.NodeStatus, 0)

	// TODO: currently compute api doesn't suppor filtering instances by labels, once supported
	// we can stop iterating over the list of instances to find those that belong to the nodegroup.
	instances, err := p.ComputeClient.Search(ctx, &pb.InstanceSearchRequest{
		Metadata: &pb.InstanceMetadataSearch{
			CloudAccountId: cloudaccountid,
			Labels: map[string]string{
				nodegroupNameLabelKey: selector,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, instance := range instances.Items {
		if value, found := instance.Metadata.Labels[nodegroupNameLabelKey]; found {
			if value == selector {
				status := privatecloudv1alpha1.NodeStatus{
					Name: instance.Metadata.Name,
				}

				var ipAddress string
				if len(instance.Status.Interfaces) > 0 {
					if len(instance.Status.Interfaces[0].Addresses) > 0 {
						ipAddress = instance.Status.Interfaces[0].Addresses[0]
					}
				}

				status.IpAddress = ipAddress
				status.InstanceIMI = instance.Spec.MachineImage
				status.CreationTime = metav1.Time{Time: instance.Metadata.CreationTimestamp.AsTime()}
				status.LastUpdate = metav1.Time{Time: time.Now()}

				// Nodegroup : active, updating, error, deleting
				// Phase: PROVISIONING ,READY,STOPPING, STOPPED, TERMINATING, FAILED
				status.Message = instance.Status.Message
				switch {
				case instance.Status.Phase == pb.InstancePhase_Provisioning:
					status.State = "Updating"
					status.Message = "Provisioning node"
				case instance.Status.Phase == pb.InstancePhase_Ready:
					status.State = "Active"
					status.Message = "Configuring node"
				case instance.Status.Phase == pb.InstancePhase_Terminating:
					status.State = "Deleting"
					status.Message = "Deleting node"
				case instance.Status.Phase == pb.InstancePhase_Failed:
					status.State = "Error"
				case instance.Status.Phase == pb.InstancePhase_Stopping:
					status.State = "Updating"
				case instance.Status.Phase == pb.InstancePhase_Stopped:
					status.State = "Updating"
				default:
					status.State = "Updating"
					status.Message = "Provisioning node"
				}
				nodesStatus = append(nodesStatus, status)
			}
		}
	}

	return nodesStatus, nil
}

// GetNode is a compute provider interface implementing Instance Service Get Method
func (p *ComputeProvider) GetNode(ctx context.Context, nodeName string, cloudaccountid string) (privatecloudv1alpha1.NodeStatus, error) {
	log := log.FromContext(ctx).WithName("ComputeProvider.GetNode")

	instance, err := p.ComputeClient.Get(ctx, &pb.InstanceGetRequest{
		Metadata: &pb.InstanceMetadataReference{
			CloudAccountId: cloudaccountid,
			NameOrId: &pb.InstanceMetadataReference_Name{
				Name: nodeName,
			},
		},
	})
	if err != nil {
		return privatecloudv1alpha1.NodeStatus{}, err
	}
	log.V(0).Info("Node status from compute provider", logkeys.NodeName, instance.Metadata.Name, logkeys.NodeStatus, instance.GetStatus())

	if instance.Status == nil {
		return privatecloudv1alpha1.NodeStatus{}, errors.New("instance status is nil")
	}

	status := privatecloudv1alpha1.NodeStatus{
		Name: instance.Metadata.Name,
	}

	var ipAddress string
	if len(instance.Status.Interfaces) > 0 {
		if len(instance.Status.Interfaces[0].Addresses) > 0 {
			ipAddress = instance.Status.Interfaces[0].Addresses[0]
		}

		status.Gateway = instance.Status.Interfaces[0].Gateway
		status.Subnet = instance.Status.Interfaces[0].Subnet
		status.Netmask = instance.Status.Interfaces[0].PrefixLength

		for _, iface := range instance.Status.Interfaces {
			if iface.Name == storageBackendInterface {
				if len(iface.Addresses) > 0 {
					status.StorageBackendIP = iface.Addresses[0]
				}

				status.StorageBackendGateway = iface.Gateway
				status.StorageBackendSubnet = iface.Subnet
			}
		}
	}

	status.IpAddress = ipAddress
	status.InstanceIMI = instance.Spec.MachineImage
	status.CreationTime = metav1.Time{Time: instance.Metadata.CreationTimestamp.AsTime()}
	status.LastUpdate = metav1.Time{Time: time.Now()}

	// Nodegroup : active, updating, error, deleting
	// Phase: PROVISIONING ,READY,STOPPING, STOPPED, TERMINATING, FAILED
	status.Message = instance.Status.Message
	switch {
	case instance.Status.Phase == pb.InstancePhase_Provisioning:
		status.State = "Updating"
		status.Message = "Provisioning node"
	case instance.Status.Phase == pb.InstancePhase_Ready:
		status.State = "Active"
		status.Message = "Configuring node"
	case instance.Status.Phase == pb.InstancePhase_Terminating:
		status.State = "Deleting"
		status.Message = "Deleting node"
	case instance.Status.Phase == pb.InstancePhase_Failed:
		status.State = "Error"
	case instance.Status.Phase == pb.InstancePhase_Stopping:
		status.State = "Updating"
	case instance.Status.Phase == pb.InstancePhase_Stopped:
		status.State = "Updating"
	default:
		status.State = "Updating"
		status.Message = "Provisioning node"
	}

	return status, nil
}

// DeleteNode is compute provider interface implementing Instance Service Delete Method
func (p *ComputeProvider) DeleteNode(ctx context.Context, nodeName string, cloudaccountid string) error {
	_, err := p.ComputePrivateClient.DeletePrivate(ctx, &pb.InstanceDeletePrivateRequest{
		Metadata: &pb.InstanceMetadataReference{
			CloudAccountId: cloudaccountid,
			NameOrId: &pb.InstanceMetadataReference_Name{
				Name: nodeName,
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *ComputeProvider) CreateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	log := log.FromContext(ctx).WithName("ComputeProvider.CreateInstanceGroup")
	if len(nodegroup.Spec.VNETS) == 0 {
		return nil, "", errors.New("no vnets found in nodegroup spec")
	}

	instanceGroup := nodegroup.Name + "-ig-" + uuid.New()[:5]

	log.V(0).Info("Creating Instance Group", "Checking Cluster Type", nodegroup.Spec.ClusterType)
	var serviceType pb.InstanceServiceType
	if nodegroup.Spec.ClusterType == superComputeClusterType {
		serviceType = pb.InstanceServiceType_SuperComputingAsAService
	} else {
		serviceType = pb.InstanceServiceType_KubernetesAsAService
	}

	// This should be true only for controlplane nodes and instance groups only support
	// workers.
	skipQuotaCheck := false

	cloudInit := &bytes.Buffer{}

	userData := nodeprovider.UserData{
		RegistrationCmd: registrationCmd,
		BootstrapScript: strings.Split(bootstrapScript, "\n"),
	}

	if len(nodegroup.Spec.UserDataURL) > 0 {
		userData.DownloadCustomBashScript = fmt.Sprintf("- curl --retry 5 --retry-connrefused --connect-timeout 10 %s -o /usr/local/bin/user-script.sh", nodegroup.Spec.UserDataURL)
		userData.RunCustomBashScript = "- bash /usr/local/bin/user-script.sh"
	}

	if err := nodeprovider.UserDataTemplate.Execute(cloudInit, userData); err != nil {
		return nil, "", err
	}

	instances := make([]*pb.InstanceCreatePrivateRequest, 0, instanceCount)
	for i := 0; i < instanceCount; i++ {
		instance := &pb.InstanceCreatePrivateRequest{
			Metadata: &pb.InstanceMetadataCreatePrivate{
				CloudAccountId: nodegroup.Spec.CloudAccountId,
				Name:           fmt.Sprintf("%s-%d", instanceGroup, i),
				Labels: map[string]string{
					clusterNameLabelKey:   nodegroup.Spec.ClusterName,
					nodegroupTypeLabelKey: string(nodegroup.Spec.NodegroupType),
					nodegroupNameLabelKey: nodegroup.Name,
				},
				ResourceId:     uuid.NewRandom().String(),
				SkipQuotaCheck: skipQuotaCheck,
			},
			Spec: &pb.InstanceSpecPrivate{
				AvailabilityZone:  nodegroup.Spec.VNETS[0].AvailabilityZone,
				InstanceType:      instanceType,
				MachineImage:      nodegroup.Spec.InstanceIMI,
				SshPublicKeyNames: nodegroup.Spec.SSHKey,
				Interfaces: []*pb.NetworkInterfacePrivate{{
					Name: networkInterfaceName,
					VNet: nodegroup.Spec.VNETS[0].NetworkInterfaceVnet,
				}},
				UserData:      cloudInit.String(),
				ServiceType:   serviceType,
				InstanceGroup: instanceGroup,
			},
		}

		if nodegroup.Spec.ClusterType == superComputeClusterType {
			log.V(0).Info("Creating Instance Group Request", logkeys.Request, instance)
		}
		instances = append(instances, instance)
	}

	req := pb.InstanceCreateMultiplePrivateRequest{Instances: instances}

	resp, err := p.ComputePrivateClient.CreateMultiplePrivate(ctx, &req)
	if err != nil {
		return nil, "", err
	}

	log.V(0).Info("Creating Instance Group Response", "Compute Instance Response Count", len(resp.Instances))
	nodeStatusList := make([]privatecloudv1alpha1.NodeStatus, 0, nodegroup.Spec.Count)
	for _, instance := range resp.Instances {
		if nodegroup.Spec.ClusterType == superComputeClusterType {
			log.V(0).Info("Creating Instance Group Response", "Compute Instance Response", instance)
		}
		nodeStatus := populateNodeStatus(instance)
		nodeStatusList = append(nodeStatusList, nodeStatus)
	}

	return nodeStatusList, instanceGroup, nil
}

func (p *ComputeProvider) CreatePrivateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	log := log.FromContext(ctx).WithName("ComputeProvider.CreatePrivateInstanceGroup")
	if len(nodegroup.Spec.VNETS) == 0 {
		return nil, "", errors.New("no vnets found in nodegroup spec")
	}

	instanceGroup := nodegroup.Name + "-ig-" + uuid.New()[:5]

	log.V(0).Info("Creating Private Instance Group", "Checking Cluster Type", nodegroup.Spec.ClusterType)
	var serviceType pb.InstanceServiceType
	if nodegroup.Spec.ClusterType == superComputeClusterType {
		serviceType = pb.InstanceServiceType_SuperComputingAsAService
	} else {
		serviceType = pb.InstanceServiceType_KubernetesAsAService
	}

	// This should be true only for controlplane nodes and instance groups only support
	// workers.
	skipQuotaCheck := false

	cloudInit := &bytes.Buffer{}

	userData := nodeprovider.UserData{
		RegistrationCmd: registrationCmd,
		BootstrapScript: strings.Split(bootstrapScript, "\n"),
	}

	if len(nodegroup.Spec.UserDataURL) > 0 {
		userData.DownloadCustomBashScript = fmt.Sprintf("- curl --retry 5 --retry-connrefused --connect-timeout 10 %s -o /usr/local/bin/user-script.sh", nodegroup.Spec.UserDataURL)
		userData.RunCustomBashScript = "- bash /usr/local/bin/user-script.sh"
	}

	if err := nodeprovider.UserDataTemplate.Execute(cloudInit, userData); err != nil {
		return nil, "", err
	}

	instanceReq := pb.InstanceGroupCreatePrivateRequest{
		Metadata: &pb.InstanceGroupMetadataCreatePrivate{
			CloudAccountId: nodegroup.Spec.CloudAccountId,
			Labels: map[string]string{
				clusterNameLabelKey:   nodegroup.Spec.ClusterName,
				nodegroupTypeLabelKey: string(nodegroup.Spec.NodegroupType),
				nodegroupNameLabelKey: nodegroup.Name,
			},
			SkipQuotaCheck: skipQuotaCheck,
			Name:           instanceGroup,
		},
		Spec: &pb.InstanceGroupSpecPrivate{
			InstanceCount: int32(instanceCount),
			InstanceSpecPrivate: &pb.InstanceSpecPrivate{
				AvailabilityZone:  nodegroup.Spec.VNETS[0].AvailabilityZone,
				InstanceType:      instanceType,
				MachineImage:      nodegroup.Spec.InstanceIMI,
				SshPublicKeyNames: nodegroup.Spec.SSHKey,
				Interfaces: []*pb.NetworkInterfacePrivate{{
					Name: networkInterfaceName,
					VNet: nodegroup.Spec.VNETS[0].NetworkInterfaceVnet,
				}},
				UserData:      cloudInit.String(),
				ServiceType:   serviceType,
				InstanceGroup: instanceGroup,
			},
			Placement: &pb.InstanceGroupPlacement{},
		},
		DryRun: false,
	}

	if nodegroup.Spec.ClusterType == superComputeClusterType {
		log.V(0).Info("Creating Instance Group Request", logkeys.Request, &instanceReq)
	}

	if p.ComputeInstanceGroupPrivateClient == nil {
		return nil, "", errors.New("compute private client is nil and not initialized")
	}

	resp, err := p.ComputeInstanceGroupPrivateClient.CreatePrivate(ctx, &instanceReq)
	if err != nil {
		return nil, "", err
	}

	log.V(0).Info("Creating Private Instance Group Client Response", "Compute Instance Response Count", len(resp.Instances))
	log.V(0).Info("Creating Private Instance Group Client Response", "Compute Instance SC Placement Group IDs", len(resp.Placement.SuperComputeGroupIds))

	nodeStatusList := make([]privatecloudv1alpha1.NodeStatus, 0, nodegroup.Spec.Count)
	for _, instance := range resp.Instances {
		if nodegroup.Spec.ClusterType == superComputeClusterType {
			log.V(0).Info("Creating Private Instance Group Response", "Compute Instance Response", instance)
		}
		nodeStatus := populateNodeStatus(instance)
		nodeStatusList = append(nodeStatusList, nodeStatus)
	}

	return nodeStatusList, instanceGroup, nil
}

// ScaleUpInstanceGroup is used to add any missing nodes for an instance group based on instance type, missing number of nodes in instance group and instance group name
func (p *ComputeProvider) ScaleUpInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup, instanceGroup string) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	log := log.FromContext(ctx).WithName("ComputeProvider.ScaleUpInstanceGroup")

	if len(nodegroup.Spec.VNETS) == 0 {
		return nil, "", errors.New("no vnets found in nodegroup spec")
	}

	cloudInit := &bytes.Buffer{}

	userData := nodeprovider.UserData{
		RegistrationCmd: registrationCmd,
		BootstrapScript: strings.Split(bootstrapScript, "\n"),
	}

	if len(nodegroup.Spec.UserDataURL) > 0 {
		userData.DownloadCustomBashScript = fmt.Sprintf("- curl --retry 5 --retry-connrefused --connect-timeout 10 %s -o /usr/local/bin/user-script.sh", nodegroup.Spec.UserDataURL)
		userData.RunCustomBashScript = "- bash /usr/local/bin/user-script.sh"
	}

	if err := nodeprovider.UserDataTemplate.Execute(cloudInit, userData); err != nil {
		return nil, "", err
	}

	instanceReq := &pb.InstanceGroupScaleRequest{
		Metadata: &pb.InstanceGroupMetadata{
			CloudAccountId: nodegroup.Spec.CloudAccountId,
			Name:           instanceGroup,
		},
		Spec: &pb.InstanceGroupSpec{
			InstanceSpec: &pb.InstanceSpec{
				AvailabilityZone:  nodegroup.Spec.VNETS[0].AvailabilityZone,
				InstanceType:      instanceType,
				MachineImage:      nodegroup.Spec.InstanceIMI,
				SshPublicKeyNames: nodegroup.Spec.SSHKey,
				Interfaces: []*pb.NetworkInterface{{
					Name: networkInterfaceName,
					VNet: nodegroup.Spec.VNETS[0].NetworkInterfaceVnet,
				}},
				UserData:      cloudInit.String(),
				InstanceGroup: instanceGroup,
			},
			InstanceCount: int32(instanceCount),
		},
	}

	resp, err := p.InstanceGroupClient.ScaleUp(ctx, instanceReq)
	if err != nil {
		return nil, "", err
	}

	log.V(0).Info("Scale up response from compute provider", "current members", strings.Join(resp.Status.CurrentMembers, ""), "ready members", strings.Join(resp.Status.ReadyMembers, ""), "new members", strings.Join(resp.Status.NewMembers, ""))

	nodeStatusList := make([]privatecloudv1alpha1.NodeStatus, 0, nodegroup.Spec.Count)
	// for _, instance := range resp.Status.ReadyMembers {
	// 	nodeStatus := populateInstanceGroupNodeStatus(instance, "Ready")
	// 	nodeStatusList = append(nodeStatusList, nodeStatus)
	// }

	for _, instance := range resp.Status.NewMembers {
		nodeStatus := privatecloudv1alpha1.NodeStatus{
			Name: instance,
		}
		nodeStatus.CreationTime = metav1.Time{Time: time.Now()}
		nodeStatus.LastUpdate = metav1.Time{Time: time.Now()}
		nodeStatus.State = privatecloudv1alpha1.UpdatingNodegroupState
		nodeStatus.State = "Updating"
		nodeStatus.Message = "Provisioning node"

		nodeStatusList = append(nodeStatusList, nodeStatus)
	}

	return nodeStatusList, instanceGroup, nil
}

// SearchInstanceGroup looks for all the instance group members available for a cloud account from compute and validates the give instance group is still available
func (p *ComputeProvider) SearchInstanceGroup(ctx context.Context, cloudaccountid string, instanceGroup string) (bool, error) {

	instanceReq := &pb.InstanceGroupSearchRequest{
		Metadata: &pb.InstanceGroupMetadataSearch{
			CloudAccountId: cloudaccountid,
		},
	}

	resp, err := p.InstanceGroupClient.Search(ctx, instanceReq)
	if err != nil {
		return false, err
	}

	for _, instanceMembers := range resp.Items {
		if instanceMembers.Metadata.Name == instanceGroup {
			return true, nil
		}
	}

	return false, nil
}

// DeleteInstanceGroupMember is compute provider interface Instance Group Delete Method based on node name and instance group
func (p *ComputeProvider) DeleteInstanceGroupMember(ctx context.Context, nodeName string, cloudaccountid string, instanceGroup string) error {
	_, err := p.InstanceGroupClient.DeleteMember(ctx, &pb.InstanceGroupMemberDeleteRequest{
		Metadata: &pb.InstanceGroupMetadata{
			CloudAccountId: cloudaccountid,
			Name:           instanceGroup,
		},
		InstanceNameOrId: &pb.InstanceGroupMemberDeleteRequest_InstanceName{
			InstanceName: nodeName,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// populateNodeStatus reads the information from the given compute instance and creates
// a node status.
func populateNodeStatus(instance *pb.InstancePrivate) privatecloudv1alpha1.NodeStatus {
	var nodeStatus privatecloudv1alpha1.NodeStatus

	nodeStatus.Name = instance.Metadata.Name
	nodeStatus.InstanceIMI = instance.Spec.MachineImage
	nodeStatus.CreationTime = metav1.Time{Time: instance.Metadata.CreationTimestamp.AsTime()}
	nodeStatus.LastUpdate = metav1.Time{Time: time.Now()}
	nodeStatus.State = privatecloudv1alpha1.UpdatingNodegroupState

	if instance.Status == nil {
		return nodeStatus
	}

	var ipAddress string
	if len(instance.Status.Interfaces) > 0 {
		if len(instance.Status.Interfaces[0].Addresses) > 0 {
			ipAddress = instance.Status.Interfaces[0].Addresses[0]
		}
	}
	nodeStatus.IpAddress = ipAddress

	nodeStatus.Message = instance.Status.Message
	switch {
	case instance.Status.Phase == pb.InstancePhase_Provisioning:
		nodeStatus.State = "Updating"
		nodeStatus.Message = "Provisioning node"
	case instance.Status.Phase == pb.InstancePhase_Ready:
		nodeStatus.State = "Active"
		nodeStatus.Message = "Configuring node"
	case instance.Status.Phase == pb.InstancePhase_Terminating:
		nodeStatus.State = "Deleting"
		nodeStatus.Message = "Deleting node"
	case instance.Status.Phase == pb.InstancePhase_Failed:
		nodeStatus.State = "Error"
	case instance.Status.Phase == pb.InstancePhase_Stopping:
		nodeStatus.State = "Updating"
	case instance.Status.Phase == pb.InstancePhase_Stopped:
		nodeStatus.State = "Updating"
	default:
		nodeStatus.State = "Updating"
		nodeStatus.Message = "Provisioning node"
	}

	return nodeStatus
}
