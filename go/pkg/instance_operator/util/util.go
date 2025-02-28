// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"context"
	"fmt"
	"strings"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"golang.org/x/crypto/ssh"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	restclient "k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	StorageAddressSpace             = "storage"
	AcceleratorClusterInterfaceName = "gpu0"
	TenantInterfaceName             = "eth0"
	StorageInterfaceName            = "storage0"
	BGPClusterInterfaceName         = "bgp0"
	XBXAddressSpace                 = "xbx"
)

func SetManagerOptions(ctx context.Context, options *manager.Options, cfg *cloudv1alpha1.InstanceOperatorConfig) error {
	log := log.FromContext(ctx).WithName("SetManagerOptions")
	if len(cfg.Filter.Labels) > 0 {
		log.Info("Filtering instances by labels", logkeys.Labels, cfg.Filter.Labels)
		options.Cache = cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&cloudv1alpha1.Instance{}: {
					Label: labels.SelectorFromSet(labels.Set(cfg.Filter.Labels)),
				},
			},
		}
	}
	return nil
}

// Helper function to add a conditions to a given condition list.
func SetStatusCondition(conditions *[]cloudv1alpha1.InstanceCondition,
	newCondition cloudv1alpha1.InstanceCondition) {

	if conditions == nil {
		conditions = &[]cloudv1alpha1.InstanceCondition{}
	}
	existingCondition := FindStatusCondition(*conditions, newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		*conditions = append(*conditions, newCondition)
	} else {
		if existingCondition.Status != newCondition.Status {
			existingCondition.Status = newCondition.Status
			if !newCondition.LastTransitionTime.IsZero() {
				existingCondition.LastTransitionTime = newCondition.LastTransitionTime
			} else {
				existingCondition.LastTransitionTime = metav1.Now()
			}
		}
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.LastProbeTime = newCondition.LastProbeTime
	}
}

func SetStatusConditionIfMissing(instance *cloudv1alpha1.Instance,
	conditionType cloudv1alpha1.InstanceConditionType, status k8sv1.ConditionStatus,
	reason cloudv1alpha1.ConditionReason, message string) {
	cond := FindStatusCondition(instance.Status.Conditions, conditionType)
	if cond == nil {
		instanceCondition := cloudv1alpha1.InstanceCondition{
			Status:             status,
			Message:            message,
			Type:               conditionType,
			LastTransitionTime: metav1.Now(),
			LastProbeTime:      metav1.Now(),
			Reason:             reason,
		}
		SetStatusCondition(&instance.Status.Conditions, instanceCondition)
	}
}

// Helper to set a condition of given type to true and others to false
func SetInstanceCondition(ctx context.Context, instance *cloudv1alpha1.Instance,
	conditionType cloudv1alpha1.InstanceConditionType, status k8sv1.ConditionStatus) {
	condition := cloudv1alpha1.InstanceCondition{
		Type:               conditionType,
		Status:             status,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}

	SetStatusCondition(&instance.Status.Conditions, condition)
}

// Utility to find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.InstanceCondition, conditionType cloudv1alpha1.InstanceConditionType) *cloudv1alpha1.InstanceCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

// Utility to find a condition of the given type on Kubevirt VirtualMachine
func FindVirtualMachineStatusCondition(virtualMachineConditions []kubevirtv1.VirtualMachineCondition, conditionType kubevirtv1.VirtualMachineConditionType) *kubevirtv1.VirtualMachineCondition {
	for i := range virtualMachineConditions {
		if virtualMachineConditions[i].Type == conditionType {
			return &virtualMachineConditions[i]
		}
	}
	return nil
}

func CreateSshProxyTunnelClusterConfig(ctx context.Context, mgr ctrl.Manager, cfg *cloudv1alpha1.InstanceOperatorConfig) (*restclient.Config, error) {
	if cfg.SshProxyTunnelCluster.KubeConfigFilePath == "" {
		return mgr.GetConfig(), nil
	}
	return toolsk8s.LoadKubeConfigFile(ctx, cfg.SshProxyTunnelCluster.KubeConfigFilePath)
}

func GetSshPublicKeys(ctx context.Context, instance *cloudv1alpha1.Instance) []string {
	var sshPublicKeys []string
	for _, spec := range instance.Spec.SshPublicKeySpecs {
		sshPublicKeys = append(sshPublicKeys, spec.SshPublicKey)
	}
	return sshPublicKeys
}

func CreateNamespaceIfNeeded(ctx context.Context, namespace string, client client.Client) error {
	namespaceResource := &k8sv1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := client.Create(ctx, namespaceResource); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func IsInstanceStartupCompleted(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceAgentConnected(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionAgentConnected)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceKcsEnabled(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionKcsEnabled)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceHCIEnabled(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionHCIEnabled)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceFailed(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionFailed)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceStarting(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStarting)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceStarted(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStarted)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceStopping(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStopping)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceStoppedCompleted(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStopped)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsInstanceVerifiedSshAccess(instance *cloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionVerifiedSshAccess)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func IsKubevirtVirtualMachineReady(virtualMachine *kubevirtv1.VirtualMachine) bool {
	cond := FindVirtualMachineStatusCondition(virtualMachine.Status.Conditions, kubevirtv1.VirtualMachineReady)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

func GetCondition(vm *kubevirtv1.VirtualMachine, cond kubevirtv1.VirtualMachineConditionType) *kubevirtv1.VirtualMachineCondition {
	if vm == nil {
		return nil
	}
	for _, c := range vm.Status.Conditions {
		if c.Type == cond {
			return &c
		}
	}
	return nil
}

func GetSupportedHostKeyAlgorithms() []string {
	supportedHostKeyAlgorithms := []string{
		ssh.KeyAlgoRSASHA512,
		ssh.KeyAlgoRSASHA256,
		ssh.KeyAlgoRSA,
		ssh.KeyAlgoED25519,
		ssh.KeyAlgoECDSA521,
		ssh.KeyAlgoECDSA384,
		ssh.KeyAlgoECDSA256,
	}
	return supportedHostKeyAlgorithms
}

func GetHostPublicKey(publicKey string) (ssh.PublicKey, error) {

	// Remove line breaks from the public key string
	hostPublicKeyString := strings.ReplaceAll(publicKey, "\n", "")

	// Remove semi colon from the public key string. A semicolon seems to get added while pasring the contents of host_public_key file
	// followed by subsequent conversion to string especially in ssh key verification of bm instance controller while using localhost
	// as bastion server for testing.
	// Ex: ssh-keyscan -t rsa localhost | awk '{print $2, $3}'> local/secrets/ssh-proxy-operator/host_public_key
	hostPublicKeyString = strings.ReplaceAll(hostPublicKeyString, ";", "")

	// We have used ParseAuthorizedKey instead of ParsePublicKey. This is because
	// There's a difference between the wire and disk format for the key. From the doc
	// ParsePublicKey parses an SSH public key formatted for use in the SSH wire protocol according to RFC 4253 whereas
	// ParseAuthorizedKey parses a public key from an authorized_keys file used in OpenSSH according to the sshd(8) manual page.
	hostPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostPublicKeyString))
	if err != nil {
		return nil, fmt.Errorf("error encountered while parsing host public key %v :%w", hostPublicKeyString, err)
	}

	return hostPublicKey, nil
}

type InstanceNetworkInfo struct {
	// NetworkMode is the network configuration of the cluster.
	NetworkMode string
	// SuperComputeGroupIDs specifies the location of the supercompute clusters.
	SuperComputeGroupIDs []string
	// ClusterGroupIDs specifies the locations of the node groups.
	ClusterGroupIDs []string
}

func GetVmOverheadMemory(ctx context.Context, memoryQty resource.Quantity, gpuCount int32) resource.Quantity {
	log := log.FromContext(ctx).WithName("Util.GetVmOverheadMemory")
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("Inputs", "memoryQty", memoryQty.Value(), "GpuCount", gpuCount)

	var reservedMemoryBytes int64
	if gpuCount > 0 {
		// Each GPU card requires 2147483648 bytes (2048 MiB)
		reservedMemoryBytes = 2048 * 1024 * 1024 * int64(gpuCount)
	} else {
		// Calculate 3% of the memory size as reserved memory
		reservedMemoryBytes = (memoryQty.Value() * 3) / 100
	}

	log.Info("VmOverheadMemory", "ReservedMemoryBytes", reservedMemoryBytes)

	// Convert reservedMemoryBytes to resource.Quantity
	reservedMemoryQty := *resource.NewQuantity(reservedMemoryBytes, resource.BinarySI)
	return reservedMemoryQty
}
