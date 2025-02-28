// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package harvester

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/builder"
	harvesterKubevirtv1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/clientset/versioned/typed/kubevirt.io/v1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	nodeprovider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/node_provider"
	"github.com/pborman/uuid"
	"github.com/rancher/norman/clientbase"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

const (
	namespace                               = "default"
	harvesterBaseAPIVersion                 = "/v1"
	harvesterVolumeClaimTemplatesAnnotation = "harvesterhci.io/volumeClaimTemplates"
	harvesterImageIdAnnotation              = "harvesterhci.io/imageId"
)

type HarvesterProvider struct {
	KubevirtClient *harvesterKubevirtv1.KubevirtV1Client
	BaseClient     *clientbase.APIBaseClient
}

func NewHarvesterProvider(url string, accessKey string, secretKey string) (*HarvesterProvider, error) {
	// Kubevirt client
	rc := rest.Config{
		Host:     url,
		Username: accessKey,
		Password: secretKey,
	}

	kubevirtClient, err := harvesterKubevirtv1.NewForConfig(&rc)
	if err != nil {
		return nil, err
	}

	// Base client
	baseClient, err := clientbase.NewAPIClient(&clientbase.ClientOpts{
		URL:       url + "/k8s/clusters/local" + harvesterBaseAPIVersion,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return nil, err
	}

	return &HarvesterProvider{
		KubevirtClient: kubevirtClient,
		BaseClient:     &baseClient,
	}, nil
}

func (p *HarvesterProvider) CreateNode(ctx context.Context, ipAddress string, nameserver string, gateway string, registrationCmd string, bootstrapScript string, nodegroup privatecloudv1alpha1.Nodegroup) (privatecloudv1alpha1.NodeStatus, error) {
	var nodeStatus privatecloudv1alpha1.NodeStatus

	nodeName := nodegroup.Name + "-" + uuid.New()[:5]

	nodeLabels := make(map[string]string, 3)
	nodeLabels["clusterName"] = nodegroup.Spec.ClusterName
	nodeLabels["nodegroupType"] = string(nodegroup.Spec.NodegroupType)
	nodeLabels["nodegroupName"] = nodegroup.Name
	nodeLabels["ipAddress"] = ipAddress

	nd := &bytes.Buffer{}
	if err := nodeprovider.NetworkDataTemplate.Execute(nd, nodeprovider.NetworkData{
		IPAddressSubnetMask: ipAddress + "/24",
		Gateway:             gateway,
		Nameserver:          nameserver,
	}); err != nil {
		return nodeStatus, err
	}

	ud := &bytes.Buffer{}
	if err := nodeprovider.UserDataTemplate.Execute(ud, nodeprovider.UserData{
		RegistrationCmd: registrationCmd,
		BootstrapScript: strings.Split(bootstrapScript, "\n"),
	}); err != nil {
		return nodeStatus, err
	}

	nodeBuilder := builder.NewVMBuilder("iks").
		Namespace(namespace).
		Name(nodeName).
		EvictionStrategy(true).
		CPU(2).
		Memory(ctx, "4096Mi", 0).
		Run(true).
		HostName(nodeName).
		NetworkInterface("default", "virtio", "", builder.NetworkInterfaceTypeBridge, "default/intel-amr-vms").
		Disk("disk-0", builder.DiskBusVirtio, false, 1).
		PVCVolume("disk-0", "40Gi", nodeName+"-disk-0", false, &builder.PersistentVolumeClaimOption{
			VolumeMode:       v1.PersistentVolumeBlock,
			AccessMode:       v1.ReadWriteMany,
			ImageID:          "default/" + nodegroup.Spec.InstanceIMI,
			StorageClassName: pointer.String("longhorn-" + nodegroup.Spec.InstanceIMI),
		}).
		Disk("cloudinitdisk", builder.DiskBusVirtio, false, 0).
		CloudInit("cloudinitdisk", builder.CloudInitSource{
			CloudInitType:      "noCloud",
			UserDataSecretName: nodeName,
			NetworkData:        nd.String(),
		}).Labels(nodeLabels)

	// Create secret that contains userdata
	userDataSecret := v1.Secret{}
	userDataSecret.ObjectMeta.Name = nodeName
	userDataSecret.ObjectMeta.Namespace = namespace
	userDataSecret.Type = v1.SecretTypeOpaque
	userDataSecret.Data = map[string][]byte{
		"userdata": ud.Bytes(),
	}

	respSecret := &v1.Secret{}
	err := p.BaseClient.Create("secret", userDataSecret, respSecret)
	if err != nil {
		return nodeStatus, err
	}

	// Create the node in harvester
	vm, err := p.KubevirtClient.VirtualMachines(namespace).Create(ctx, nodeBuilder.VirtualMachine, metav1.CreateOptions{})
	if err != nil {
		return nodeStatus, err
	}

	// Get node status
	nodeStatus.Name = vm.Name

	if a, ok := vm.Annotations[harvesterVolumeClaimTemplatesAnnotation]; ok {
		var pvcTemplateList []v1.PersistentVolumeClaimTemplate
		if err := json.Unmarshal([]byte(a), &pvcTemplateList); err == nil {
			for _, pvcTemplate := range pvcTemplateList {
				if imageId, ok := pvcTemplate.Annotations[harvesterImageIdAnnotation]; ok {
					instanceIMI := imageId
					instanceIMISplit := strings.Split(imageId, "/")
					if len(instanceIMISplit) == 2 {
						instanceIMI = instanceIMISplit[1]
					}
					nodeStatus.InstanceIMI = instanceIMI
				}
			}
		}
	}

	state, lastUpdate, reason, message := getState(vm.Status.Conditions)
	nodeStatus.State = state
	nodeStatus.LastUpdate = lastUpdate
	nodeStatus.Reason = reason
	nodeStatus.Message = message
	nodeStatus.CreationTime = vm.ObjectMeta.CreationTimestamp

	if ip, ok := vm.GetLabels()["ipAddress"]; ok {
		nodeStatus.IpAddress = ip
	}

	return nodeStatus, nil
}

func (p *HarvesterProvider) GetNodes(ctx context.Context, selector string, cloudaccountid string) ([]privatecloudv1alpha1.NodeStatus, error) {
	nodesStatus := make([]privatecloudv1alpha1.NodeStatus, 0)

	harvesterNodes, err := p.KubevirtClient.VirtualMachines(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	for _, node := range harvesterNodes.Items {
		status := privatecloudv1alpha1.NodeStatus{
			Name: node.Name,
		}

		status.InstanceIMI = ""
		if a, ok := node.Annotations[harvesterVolumeClaimTemplatesAnnotation]; ok {
			var pvcTemplateList []v1.PersistentVolumeClaimTemplate
			if err := json.Unmarshal([]byte(a), &pvcTemplateList); err == nil {
				for _, pvcTemplate := range pvcTemplateList {
					if imageId, ok := pvcTemplate.Annotations[harvesterImageIdAnnotation]; ok {
						instanceIMI := imageId
						instanceIMISplit := strings.Split(imageId, "/")
						if len(instanceIMISplit) == 2 {
							instanceIMI = instanceIMISplit[1]
						}
						status.InstanceIMI = instanceIMI
					}
				}
			}
		}

		state, lastUpdate, reason, message := getState(node.Status.Conditions)
		status.State = state
		status.LastUpdate = lastUpdate
		status.Reason = reason
		status.Message = message
		status.CreationTime = node.ObjectMeta.CreationTimestamp

		if ip, ok := node.GetLabels()["ipAddress"]; ok {
			status.IpAddress = ip
		}

		nodesStatus = append(nodesStatus, status)
	}

	return nodesStatus, nil
}

// GetNode is a compute provider interface implementing Instance Service Get Method
func (p *HarvesterProvider) GetNode(ctx context.Context, nodeName string, cloudaccountid string) (privatecloudv1alpha1.NodeStatus, error) {

	harvesterNodes, err := p.KubevirtClient.VirtualMachines(namespace).Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return privatecloudv1alpha1.NodeStatus{}, err
	}

	status := privatecloudv1alpha1.NodeStatus{
		Name: harvesterNodes.Name,
	}

	status.InstanceIMI = ""
	if a, ok := harvesterNodes.Annotations[harvesterVolumeClaimTemplatesAnnotation]; ok {
		var pvcTemplateList []v1.PersistentVolumeClaimTemplate
		if err := json.Unmarshal([]byte(a), &pvcTemplateList); err == nil {
			for _, pvcTemplate := range pvcTemplateList {
				if imageId, ok := pvcTemplate.Annotations[harvesterImageIdAnnotation]; ok {
					instanceIMI := imageId
					instanceIMISplit := strings.Split(imageId, "/")
					if len(instanceIMISplit) == 2 {
						instanceIMI = instanceIMISplit[1]
					}
					status.InstanceIMI = instanceIMI
				}
			}
		}
	}

	state, lastUpdate, reason, message := getState(harvesterNodes.Status.Conditions)
	status.State = state
	status.LastUpdate = lastUpdate
	status.Reason = reason
	status.Message = message
	status.CreationTime = harvesterNodes.ObjectMeta.CreationTimestamp

	if ip, ok := harvesterNodes.GetLabels()["ipAddress"]; ok {
		status.IpAddress = ip
	}

	return status, nil
}

func (p *HarvesterProvider) DeleteNode(ctx context.Context, nodeName string, cloudaccountid string) error {
	err := p.KubevirtClient.VirtualMachines(namespace).Delete(ctx, nodeName, *metav1.NewDeleteOptions(0))
	if !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}

func getState(conditions []kubevirtapiv1.VirtualMachineCondition) (privatecloudv1alpha1.NodegroupState, metav1.Time, string, string) {
	var lastUpdate metav1.Time
	var reason string
	var message string

	for _, condition := range conditions {
		if condition.Type == kubevirtapiv1.VirtualMachineFailure && condition.Status == v1.ConditionTrue {
			return privatecloudv1alpha1.ErrorNodegroupState,
				condition.LastTransitionTime,
				condition.Reason,
				condition.Message
		}

		if condition.Type == kubevirtapiv1.VirtualMachineReady && condition.Status == v1.ConditionTrue {
			return privatecloudv1alpha1.ActiveNodegroupState,
				condition.LastTransitionTime,
				condition.Reason,
				condition.Message
		}

		if condition.Type == kubevirtapiv1.VirtualMachineReady && condition.Status == v1.ConditionFalse {
			lastUpdate = condition.LastTransitionTime
			reason = condition.Reason
			message = condition.Message
			break
		}
	}

	return privatecloudv1alpha1.UpdatingNodegroupState,
		lastUpdate,
		reason,
		message
}

func (p *HarvesterProvider) CreateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return nil, "", errors.New("not supported")
}

func (p *HarvesterProvider) CreatePrivateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return nil, "", errors.New("not supported")
}

func (p *HarvesterProvider) DeleteInstanceGroupMember(ctx context.Context, nodeName string, cloudaccountid string, instanceGroup string) error {
	return errors.New("not supported")
}

func (p *HarvesterProvider) ScaleUpInstanceGroup(context.Context, string, string, string, int, privatecloudv1alpha1.Nodegroup, string) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return nil, "", errors.New("not supported")
}

func (p *HarvesterProvider) SearchInstanceGroup(ctx context.Context, cloudaccountid string, instanceGroup string) (bool, error) {
	return false, errors.New("not supported")
}
