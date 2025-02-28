// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"fmt"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_storage"
)

func getSlurmCtlNodeLabels() map[string]string {
	return map[string]string{
		"oneapi-instance-role": "slurm-controller-node",
	}
}

func getSlurmdNodeLabels() map[string]string {
	return map[string]string{
		"oneapi-instance-role": "slurm-compute-node",
	}
}

func getSlurmLoginNodeLabels() map[string]string {
	return map[string]string{
		"oneapi-instance-role": "slurm-login-node",
	}
}

func getSlurmNFSNodeLabels() map[string]string {
	return map[string]string{
		"oneapi-instance-role": "nfs-server-node",
	}
}

func getSlurmJupyterHubNodeLabels() map[string]string {
	return map[string]string{
		"oneapi-instance-role": "slurm-jupyterhub-node",
	}
}

/*******************************************************
* IDC Compute gRPC - Slurm Instance Create Request
*******************************************************/
func getSlurmLoginNodeSpecs(node *v1.ClusterNode, vnet, clusterId, cloudAccountId string, instIdx int, keys []string, loginCloudCfgUserData string) idc_compute.InstanceCreateRequest {
	return idc_compute.InstanceCreateRequest{
		Name:           fmt.Sprintf("%s-login-%d", clusterId, instIdx),
		ClusterId:      clusterId,
		CloudAccountId: cloudAccountId,
		VNet:           vnet,
		MachineType:    MapInstanceTypeFromPbToString(node.MachineType),
		SshKeyNames:    keys,
		Labels:         getSlurmLoginNodeLabels(),
		ImageName:      node.ImageType,
		UserData:       loginCloudCfgUserData,
	}
}

func getSlurmComputeNodeSpecs(node *v1.ClusterNode, vnet, clusterId, cloudAccountId string, instIdx int, keys []string, slurmdCloudCfgUserData string) idc_compute.InstanceCreateRequest {
	return idc_compute.InstanceCreateRequest{
		Name:           fmt.Sprintf("%s-slurmd-%d", clusterId, instIdx),
		ClusterId:      clusterId,
		CloudAccountId: cloudAccountId,
		VNet:           vnet,
		MachineType:    MapInstanceTypeFromPbToString(node.MachineType),
		SshKeyNames:    keys,
		Labels:         getSlurmdNodeLabels(),
		ImageName:      node.ImageType,
		UserData:       slurmdCloudCfgUserData,
	}
}

func getSlurmJupyterhubNodeSpecs(node *v1.ClusterNode, vnet, clusterId, cloudAccountId string, instIdx int, keys []string, jupyterhubCloudCfgUserData string) idc_compute.InstanceCreateRequest {
	return idc_compute.InstanceCreateRequest{
		Name:           fmt.Sprintf("%s-jupyterhub-%d", clusterId, instIdx),
		ClusterId:      clusterId,
		CloudAccountId: cloudAccountId,
		VNet:           vnet,
		MachineType:    MapInstanceTypeFromPbToString(node.MachineType),
		SshKeyNames:    keys,
		Labels:         getSlurmJupyterHubNodeLabels(),
		ImageName:      node.ImageType,
		UserData:       jupyterhubCloudCfgUserData,
	}
}

func getSlurmControllerNodeSpecs(node *v1.ClusterNode, vnet, clusterId, cloudAccountId string, instIdx int, keys []string, slurmctldCloudCfgUserData string) idc_compute.InstanceCreateRequest {
	return idc_compute.InstanceCreateRequest{
		Name:           fmt.Sprintf("%s-slurmctld-%d", clusterId, instIdx),
		ClusterId:      clusterId,
		CloudAccountId: cloudAccountId,
		VNet:           vnet,
		MachineType:    MapInstanceTypeFromPbToString(node.MachineType),
		SshKeyNames:    keys,
		Labels:         getSlurmCtlNodeLabels(),
		ImageName:      node.ImageType,
		UserData:       slurmctldCloudCfgUserData,
	}
}

/*****************************************************
* IDC Storage gRPC - Slurm Storage Filesystem
*****************************************************/
func getSlurmStorageNodeSpecs(storageNode *v1.StorageNode, clusterId string, cloudAccountId, az string) idc_storage.FilesystemCreateRequest {
	return idc_storage.FilesystemCreateRequest{
		Name:             storageNode.GetName(),
		CloudAccountId:   cloudAccountId,
		Description:      fmt.Sprintf("%s %s", clusterId, storageNode.GetDescription()),
		AvailabilityZone: az,
		Capacity:         storageNode.GetCapacity(),
		AccessModes:      MapAccessModeFromTrainingToStorage(storageNode.GetAccessMode()),
		MountProtocol:    MapMountProtocolFromTrainingToStorage(storageNode.GetMount()),
	}
}
