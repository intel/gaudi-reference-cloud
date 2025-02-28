// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"

	controllers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/controllers"
	util "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	vmbuilder "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/builder"
	vmutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	kubevirtclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/clientset/versioned"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	nadclient "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned"
	"gopkg.in/yaml.v2"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	defaultInterfaceName = "default"
	defaultVolumeDisk    = "disk-0"
	storageNetworkMTU    = 9000
	defaultNetworkMTU    = 1500
	storageInterfaceName = "storage0"
	defaultOSUser        = "ubuntu"
)

var (
	// Regular expression pattern to match an valid IPv4 address
	ipPattern = regexp.MustCompile(`\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
)

func NewVmInstanceReconciler(ctx context.Context, mgr ctrl.Manager, vNetPrivateClient pb.VNetPrivateServiceClient, vNetClient pb.VNetServiceClient, cfg *cloudv1alpha1.VmInstanceOperatorConfig) (*controllers.InstanceReconciler, error) {
	instanceBackend, err := NewVmInstanceBackend(ctx, mgr, cfg)
	if err != nil {
		return nil, err
	}
	return controllers.NewInstanceReconciler(ctx, mgr, vNetPrivateClient, vNetClient, instanceBackend, &cfg.InstanceOperator)
}

type VmInstanceBackend struct {
	InstanceClient    client.Client
	VmProviderClient  client.Client
	VmProviderCluster cluster.Cluster
	KubevirtClient    *kubevirtclient.Clientset
	K8Client          *kubernetes.Clientset
	NadClient         *nadclient.Clientset
	Cfg               *cloudv1alpha1.VmInstanceOperatorConfig
	VmProviderHelper  vmProviderInterface
}

type vmProviderInterface interface {
	createRoleBindingIfNeeded(ctx context.Context, namespace string, k8Client *kubernetes.Clientset) error
	getVmBuilderCreator() string
	getVmSecretLabels() map[string]string
	buildVolume(instance *cloudv1alpha1.Instance, vmBuilder *vmbuilder.VMBuilder) *vmbuilder.VMBuilder
}

type kubeVirtProvider struct {
	vmBuilderCreator string
}

type harvesterProvider struct {
	vmBuilderCreator string
}

func NewVmInstanceBackend(ctx context.Context, mgr ctrl.Manager, cfg *cloudv1alpha1.VmInstanceOperatorConfig) (*VmInstanceBackend, error) {
	if err := ValidateVmConfig(cfg); err != nil {
		return nil, err
	}
	vmInfraConfig, err := CreateVmProvisionClusterConfig(ctx, mgr, cfg)
	if err != nil {
		return nil, fmt.Errorf("NewVmInstanceBackend: %w", err)
	}
	vmInfraCluster, err := cluster.New(vmInfraConfig, func(o *cluster.Options) {
		o.Scheme = mgr.GetScheme()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get vmProviderCluster: %w", err)
	}
	vmInfraKubevirtClient, err := kubevirtclient.NewForConfig(vmInfraConfig)
	if err != nil {
		return nil, fmt.Errorf("GetVmProviderClients: error in kubevirtclient: %w", err)
	}
	vmInfraK8Client, err := kubernetes.NewForConfig(vmInfraConfig)
	if err != nil {
		return nil, fmt.Errorf("GetVmProviderClients: error in k8client: %w", err)
	}
	nadClient, err := nadclient.NewForConfig(vmInfraConfig)
	if err != nil {
		return nil, fmt.Errorf("GetVmProviderClients: error in nadclient: %w", err)
	}
	if err := mgr.Add(vmInfraCluster); err != nil {
		return nil, fmt.Errorf("unable to add vmProviderCluster to manager: %w", err)
	}

	var vmProviderHelper vmProviderInterface
	if cfg.InstanceOperator.OperatorFeatureFlags.UseKubeVirtCluster {
		vmProviderHelper = &kubeVirtProvider{
			vmBuilderCreator: "kubevirt-vmaas",
		}
	} else {
		vmProviderHelper = &harvesterProvider{
			vmBuilderCreator: "harvester-vmaas",
		}
	}

	return &VmInstanceBackend{
		InstanceClient:    mgr.GetClient(),
		VmProviderClient:  vmInfraCluster.GetClient(),
		VmProviderCluster: vmInfraCluster,
		KubevirtClient:    vmInfraKubevirtClient,
		K8Client:          vmInfraK8Client,
		NadClient:         nadClient,
		Cfg:               cfg,
		VmProviderHelper:  vmProviderHelper,
	}, nil
}

func ValidateVmConfig(cfg *cloudv1alpha1.VmInstanceOperatorConfig) error {
	networkData := cfg.VmProvisionConfig.CloudInitData.NetworkData
	if len(networkData.Network.Config) == 0 {
		return fmt.Errorf("empty VmProvisionConfig.cloudInitData.networkData.network.config")
	}
	if len(networkData.Network.Config[0].Subnets) == 0 {
		return fmt.Errorf("empty VmProvisionConfig.cloudInitData.networkData.network.config[0].subnets")
	}
	return nil
}

func CreateVmProvisionClusterConfig(ctx context.Context, mgr ctrl.Manager, cfg *cloudv1alpha1.VmInstanceOperatorConfig) (*restclient.Config, error) {
	if cfg.VmProvisionConfig.KubeConfigFilePath == "" {
		return mgr.GetConfig(), nil
	}
	return toolsk8s.LoadKubeConfigFile(ctx, cfg.VmProvisionConfig.KubeConfigFilePath)
}

// Reconcile Instance when kubevirt VirtualMachine with same name changes.
func (b *VmInstanceBackend) BuildController(ctx context.Context, ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.
		WatchesRawSource(
			source.Kind(b.VmProviderCluster.GetCache(), &kubevirtv1.VirtualMachine{}),
			&handler.EnqueueRequestForObject{},
		)
}

// Check if the error message contain a possible IP address.
// In some cases, this may return true even if the string contains an invalid IP address.
func (op VmInstanceBackend) errorContainsIp(err error) bool {
	if err == nil {
		return false
	}
	return ipPattern.MatchString(err.Error())
}

func (op VmInstanceBackend) CreateOrUpdateInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.CreateOrUpdateInstance").Start()
	defer span.End()
	logger.Info("BEGIN")

	// Try to get Kubevirt VirtualMachine. If not found, continue to reconcile.
	virtualMachine, err := op.getKubevirtVirtualMachine(ctx, instance)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance failed: %w", err)
	}
	if err := op.updateStatusFromVirtualMachine(ctx, instance, virtualMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance failed: %w", err)
	}

	if util.IsInstanceFailed(instance) {
		logger.Info("Skipping creation of VM because instance has failed")
	} else if virtualMachine != nil {
		logger.Info("Skipping creation of VM because it already exists")
	} else {
		logger.Info("Creating VM", logkeys.Instance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(instance)))
		// Creating VM Artifacts in KubeVirt/Harvester (Idempotent Harvester APIs)
		if err := op.CreateVmArtifacts(ctx, instance); err != nil {

			if op.errorContainsIp(err) {
				logger.Error(err, "CreateOrUpdateInstance: CreateVmArtifacts failed")
				return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance: CreateVmArtifacts failed")
			} else {
				return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance: CreateVmArtifacts failed: %w", err)
			}
		}
		// Get the created Kubevirt Virtual Machine and update the Instance status.
		virtualMachine, err = op.getKubevirtVirtualMachine(ctx, instance)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance failed: %w", err)
		}
		if err := op.updateStatusFromVirtualMachine(ctx, instance, virtualMachine); err != nil {
			return ctrl.Result{}, fmt.Errorf("CreateOrUpdateInstance failed: %w", err)
		}
	}

	logger.Info("END")
	return ctrl.Result{}, nil
}

// Returns (nil, NotFound) if the VirtualMachine is not found.
func (op VmInstanceBackend) getKubevirtVirtualMachine(ctx context.Context, instance *cloudv1alpha1.Instance) (*kubevirtv1.VirtualMachine, error) {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.getKubevirtVirtualMachine")
	log.Info("BEGIN")
	virtualMachine := &kubevirtv1.VirtualMachine{}
	err := op.VmProviderClient.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, virtualMachine)
	if err != nil {
		return nil, fmt.Errorf("getKubevirtVirtualMachine: %w", err)
	}
	log.Info("END", logkeys.VirtualMachine, virtualMachine, logkeys.VmPrintableStatus, virtualMachine.Status.PrintableStatus)
	return virtualMachine, nil
}

func (op VmInstanceBackend) updateStatusFromVirtualMachine(ctx context.Context, instance *cloudv1alpha1.Instance, virtualMachine *kubevirtv1.VirtualMachine) error {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.updateStatusFromVirtualMachine")
	log.Info("BEGIN")
	if err := op.updateInstanceConditions(ctx, instance, virtualMachine); err != nil {
		return fmt.Errorf("updateStatusFromVirtualMachine failed: %w", err)
	}
	log.Info("END")
	return nil
}

func (op VmInstanceBackend) updateInstanceConditions(ctx context.Context, instance *cloudv1alpha1.Instance, virtualMachine *kubevirtv1.VirtualMachine) error {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.updateInstanceConditions")
	log.Info("BEGIN")
	defer log.Info("END")

	startupComplete := util.IsInstanceStartupCompleted(instance)

	if virtualMachine == nil {
		// VirtualMachine does not exist.
		instanceFailed := util.IsInstanceFailed(instance)
		if startupComplete && !instanceFailed {
			log.Info("Instance previously started but VirtualMachine no longer exists. Marking instance as permanently failed.")
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionFailed,
				Status:             k8sv1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            "Virtual machine not found",
			}
			util.SetStatusCondition(&instance.Status.Conditions, condition)
		} else {
			log.Info("VirtualMachine does not exist")
		}
		return nil
	}

	log.Info("VirtualMachine PrintableStatus", logkeys.VmPrintableStatus, virtualMachine.Status.PrintableStatus)

	// Copy conditions from VirtualMachine to Instance.
	for _, cond := range virtualMachine.Status.Conditions {
		var newConditionType cloudv1alpha1.InstanceConditionType
		var newConditionStatus k8sv1.ConditionStatus

		switch cond.Type {
		case kubevirtv1.VirtualMachineReady:
			newConditionType = cloudv1alpha1.InstanceConditionRunning
			newConditionStatus = cond.Status
		case kubevirtv1.VirtualMachineConditionType("PodScheduled"):
			// This condition is reported as false but disappears before turning true. Do not copy it.
			continue
		default:
			newConditionType = cloudv1alpha1.InstanceConditionType(cond.Type)
			newConditionStatus = cond.Status
		}
		condition := cloudv1alpha1.InstanceCondition{
			Type:               newConditionType,
			Status:             newConditionStatus,
			LastProbeTime:      cond.LastProbeTime,
			LastTransitionTime: cond.LastTransitionTime,
			Reason:             cloudv1alpha1.ConditionReason(cond.Reason),
			Message:            cond.Message,
		}
		log.V(9).Info("Instance Condition", logkeys.InstanceCondition, condition)
		util.SetStatusCondition(&instance.Status.Conditions, condition)
	}

	// Set StartupComplete condition to true when Running becomes true. StartupComplete never changes after becoming true.
	if !startupComplete {
		agentConnectedCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionAgentConnected)
		if agentConnectedCond != nil && agentConnectedCond.Status == k8sv1.ConditionTrue {
			log.Info("Startup complete")
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionStartupComplete,
				Status:             k8sv1.ConditionTrue,
				LastProbeTime:      agentConnectedCond.LastProbeTime,
				LastTransitionTime: agentConnectedCond.LastTransitionTime,
				Reason:             cloudv1alpha1.ConditionReason(agentConnectedCond.Reason),
				Message:            agentConnectedCond.Message,
			}
			util.SetStatusCondition(&instance.Status.Conditions, condition)
		}
	}

	// When the virt-laucher pod gets terminated for a VirtualMachine, the Kubervirt VirtualMachine status for condition 'running' changes to 'False'
	// and instance is marked as Failed
	isKubevirtVirtualMachineReady := util.IsKubevirtVirtualMachineReady(virtualMachine)
	if startupComplete && !isKubevirtVirtualMachineReady {
		condition := cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionFailed,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "virt-launcher pod does not exists",
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
	}
	return nil
}

// This function deletes the VM, associated volumes and secrets as well.
func (op VmInstanceBackend) DeleteResources(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.DeleteResources").Start()
	defer span.End()
	log.Info("BEGIN")
	if result, err := op.DeleteVM(ctx, instance); err != nil || result.Requeue {
		return result, err
	}
	log.Info("Deletion of VM completed")
	if err := op.DeletePVC(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("deleteVMResources: failed to delete Volume : %w", err)
	}
	log.Info("Deletion of volume completed")
	if err := op.DeleteSecret(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("deleteVMResources: failed to delete secret : %w", err)
	}
	log.Info("Deletion of Secret completed")
	log.Info("END")
	return ctrl.Result{}, nil
}

// Function to create Virtual Machine
func (op VmInstanceBackend) CreateVmArtifacts(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.CreateVmArtifacts").Start()
	defer span.End()
	logger.Info("BEGIN", logkeys.Instance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(instance)))
	if err := util.CreateNamespaceIfNeeded(ctx, instance.ObjectMeta.Namespace, op.VmProviderClient); err != nil {
		return err
	}
	if err := op.VmProviderHelper.createRoleBindingIfNeeded(ctx, instance.ObjectMeta.Namespace, op.K8Client); err != nil {
		return err
	}
	if err := op.CreateVmSecrets(ctx, instance); err != nil {
		return err
	}
	if err := op.CreateVlans(ctx, instance); err != nil {
		return err
	}
	if _, err := op.CreateVM(ctx, instance); err != nil {
		return err
	}
	logger.Info("END", logkeys.Instance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(instance)))
	return nil
}

// Function to create VLANs (NetworkAttachmentDefinition).
func (op VmInstanceBackend) CreateVlans(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.CreateVlans")
	log.Info("BEGIN")
	for _, intfStatus := range instance.Status.Interfaces {
		err := op.CreateVlan(ctx, instance.ObjectMeta.Namespace, intfStatus)
		if err != nil {
			return err
		}
	}
	return nil
}

// Function to create a VLAN (NetworkAttachmentDefinition).
func (op VmInstanceBackend) CreateVlan(ctx context.Context, namespace string, intfStatus cloudv1alpha1.InstanceInterfaceStatus) error {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.CreateVlan")
	log.Info("BEGIN")
	clusterNetwork := op.Cfg.VmProvisionConfig.VMclusterNetwork
	mtu := defaultNetworkMTU
	if intfStatus.Name == storageInterfaceName {
		clusterNetwork = op.Cfg.VmProvisionConfig.StorageClusterNetwork
		mtu = storageNetworkMTU
	}
	networkConfig := vmbuilder.CreateNetworkAttachmentDefinitionSpecStr(intfStatus.VlanId, clusterNetwork, mtu)
	log.Info("networkConfig", logkeys.Configuration, networkConfig)

	// Check whether VLAN already exists
	_, err := op.NadClient.K8sCniCncfIoV1().
		NetworkAttachmentDefinitions(namespace).
		Get(ctx, vmbuilder.GetNetworkAttachmentDefinitionName(intfStatus.VlanId), metav1.GetOptions{})
	if err == nil {
		log.V(9).Info("CreateVlan: VLAN already exists")
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("CreateVlan: error while checking for existing vlan: %w", err)
	}

	nadSpec := &networkv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: vmbuilder.GetNetworkAttachmentDefinitionName(intfStatus.VlanId),
			Labels: map[string]string{
				vmbuilder.LabelKeyNetworkType:    vmbuilder.NetworkTypeVLAN,
				vmbuilder.LabelKeyClusterNetwork: clusterNetwork,
			},
		},
		Spec: networkv1.NetworkAttachmentDefinitionSpec{Config: networkConfig},
	}

	_, err = op.NadClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).
		Create(ctx, nadSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("CreateVlan: error while creating the vlan: %w", err)
	}
	log.Info("END")
	return nil
}

// Function to create Secret
func (op VmInstanceBackend) CreateVmSecrets(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.CreateVmSecrets")
	log.Info("BEGIN")
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name
	intfStatus := instance.Status.Interfaces[0]
	intfSpec := instance.Spec.Interfaces[0]
	address := intfStatus.Addresses[0]

	// Create new CloudConfig object and load incoming UserData
	newCloudConfig, err := util.NewCloudConfig("ubuntu", instance.Spec.UserData)
	if err != nil {
		return fmt.Errorf("CreateVmSecrets: error while Unmarshaling requestUserData: %w", err)
	}

	// Override VMaaS required UserData fields
	hostName, _, _ := strings.Cut(intfStatus.DnsName, ".")
	newCloudConfig.SetManageEtcHosts("localhost")
	newCloudConfig.SetHostName(hostName)
	newCloudConfig.SetFqdn(intfStatus.DnsName)

	newCloudConfig.SetDefaultUserGroup(defaultOSUser, util.GetSshPublicKeys(ctx, instance))
	if instance.Spec.QuickConnectEnabled == pb.TriState_True.String() && op.Cfg.InstanceOperator.OperatorFeatureFlags.EnableQuickConnectClientCA {
		rootCAPublicCertificateFile, err := os.Open("/vault/secrets/quick-connect-client-ca.pem")
		if err != nil {
			return fmt.Errorf("error occurred file opening root CA public certificate file: %w", err)
		}
		defer rootCAPublicCertificateFile.Close()

		rootCAPublicCertificateFileContentByte, err := io.ReadAll(rootCAPublicCertificateFile)
		if err != nil {
			return fmt.Errorf("error occurred file reading content of root CA public certificate file: %w", err)
		}
		// Install jupyterlab
		err = newCloudConfig.SetJupyterLab(namespace, name, string(rootCAPublicCertificateFileContentByte), defaultOSUser,
			op.Cfg.InstanceOperator.QuickConnectHost)
		if err != nil {
			return fmt.Errorf("error occured setting JupyterLab: %w", err)
		}
	}
	err = newCloudConfig.SetStunnelConf(op.Cfg.InstanceOperator.StorageClusterAddr)
	if err != nil {
		return fmt.Errorf("error occured setting stunnel config: %w", err)
	}
	// set write_files to cloud_int
	newCloudConfig.SetWriteFile()
	// set runcmd to cloud_init
	newCloudConfig.SetRunCmd()
	// set packages to cloud_init
	newCloudConfig.SetPackages()
	// Convert CloudConfig object to yaml
	userDataSerialized, err := newCloudConfig.RenderYAML()
	if err != nil {
		return fmt.Errorf("CreateVmSecrets: error while Marshaling userData: %w", err)
	}

	networkData := op.Cfg.VmProvisionConfig.CloudInitData.NetworkData
	networkData.Network.Config[0].Subnets[0].Address = fmt.Sprintf("%s/%d", address, intfStatus.PrefixLength)
	networkData.Network.Config[0].Subnets[0].Gateway = intfStatus.Gateway
	networkData.Network.Config[0].Subnets[0].Dns_nameservers = intfSpec.Nameservers

	needStorageConfig := len(instance.Spec.Interfaces) > 1 &&
		len(instance.Status.Interfaces) > 1 &&
		instance.Status.Interfaces[1].Name == storageInterfaceName &&
		len(op.Cfg.StorageServerSubnets) > 0

	if needStorageConfig {
		storageConfig, err := op.BuildStorageConfig(ctx, instance.Status.Interfaces[1], instance.Spec.Interfaces[1].Nameservers)
		if err != nil {
			return err
		}
		networkData.Network.Config = append(networkData.Network.Config, storageConfig)
	}

	networkDataSerialized, err := yaml.Marshal(networkData)
	log.Info("Network Data", logkeys.NetworkDataSerialized, networkDataSerialized)
	if err != nil {
		return fmt.Errorf("CreateVmSecrets: error while Marshaling networkData: %w", err)
	}

	vmSecret := &k8sv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    op.VmProviderHelper.getVmSecretLabels(),
		},
		Data: map[string][]byte{
			"userdata":    userDataSerialized,
			"networkdata": networkDataSerialized,
		},
		Type: "secret",
	}

	// Check whether secret already exists
	_, err = op.K8Client.CoreV1().Secrets(instance.ObjectMeta.Namespace).Get(ctx, instance.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil {
		log.Info("CreateVmSecrets: Secret Already exists. So updating")
		_, err = op.K8Client.CoreV1().Secrets(instance.ObjectMeta.Namespace).
			Update(ctx, vmSecret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("CreateVmSecrets: error while updating the secret: %w", err)
		}
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("CreateVmSecrets: error while checking for existing secret: %w", err)
	}

	_, err = op.K8Client.CoreV1().Secrets(instance.ObjectMeta.Namespace).
		Create(ctx, vmSecret, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("CreateVmSecrets: error while creating the secret: %w", err)
	}
	log.Info("END")
	return nil
}

func (op VmInstanceBackend) BuildStorageConfig(ctx context.Context, intfs cloudv1alpha1.InstanceInterfaceStatus, dns_nameservers []string) (cloudv1alpha1.CloudInitSubnetConfig, error) {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.BuildStorageConfig")
	log.Info("BEGIN")
	defer log.Info("END")
	storageSubnet := cloudv1alpha1.CloudInitSubnet{
		Address:         fmt.Sprintf("%s/%d", intfs.Addresses[0], intfs.PrefixLength),
		Dns_nameservers: dns_nameservers,
		Type:            "static",
	}
	for _, storageServerSubnet := range op.Cfg.StorageServerSubnets {
		destination, netmask, err := cidrToDestinationAndNetmask(storageServerSubnet)
		if err != nil {
			return cloudv1alpha1.CloudInitSubnetConfig{}, fmt.Errorf("BuildStorageConfig: error while computing destination, netmask: %w", err)
		}
		storageSubnet.Routes = append(
			storageSubnet.Routes, cloudv1alpha1.Route{
				Gateway: intfs.Gateway, Destination: destination, Netmask: netmask,
			},
		)
	}

	// Create a new config entry with the storage subnet
	storageConfig := cloudv1alpha1.CloudInitSubnetConfig{
		Name:    "enp2s0",
		Subnets: []cloudv1alpha1.CloudInitSubnet{storageSubnet},
		Type:    "physical",
	}
	return storageConfig, nil
}

func cidrToDestinationAndNetmask(cidr string) (string, string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	destination := ip.Mask(ipnet.Mask).String() // Convert destination IP to string
	netmask := fmt.Sprintf("%d.%d.%d.%d", ipnet.Mask[0], ipnet.Mask[1], ipnet.Mask[2], ipnet.Mask[3])
	return destination, netmask, nil
}

func (op VmInstanceBackend) CreateVM(ctx context.Context, instance *cloudv1alpha1.Instance) (*kubevirtv1.VirtualMachine, error) {
	log := log.FromContext(ctx).WithName("VmInstanceBackend.CreateVM")
	log.Info("BEGIN")
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name

	// Check VM already exists
	_, err := op.KubevirtClient.KubevirtV1().
		VirtualMachineInstances(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		log.Info("CreateVM: VM Already exists")
		return &kubevirtv1.VirtualMachine{}, nil
	} else if !errors.IsNotFound(err) {
		return &kubevirtv1.VirtualMachine{}, fmt.Errorf("CreateVM: error while checking for existing VM: %w", err)
	}

	// build VM artifacts
	vmToCreate, err := op.buildVirtualMachine(ctx, instance)
	if err != nil {
		return &kubevirtv1.VirtualMachine{}, fmt.Errorf("CreateVM: error while Building VirtualMachine: %w", err)
	}
	log.Info("CreateVM", logkeys.VirtualMachine, &vmToCreate)

	// Create VM
	createdVM, err := op.KubevirtClient.KubevirtV1().
		VirtualMachines(namespace).
		Create(ctx, vmToCreate, metav1.CreateOptions{})
	if err != nil {
		return &kubevirtv1.VirtualMachine{}, fmt.Errorf("CreateVM: error while creating VirtualMachine: %w", err)
	}
	log.Info("END")
	return createdVM, nil
}

func (op VmInstanceBackend) buildVirtualMachine(ctx context.Context, instance *cloudv1alpha1.Instance) (*kubevirtv1.VirtualMachine, error) {
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name
	instanceTypeSpec := instance.Spec.InstanceTypeSpec
	affinity, err := op.getAffinityForInstance(instance)
	if err != nil {
		return nil, err
	}
	topologySpreadConstraint, err := op.getTopologySpreadConstraintsForInstance(instance)
	if err != nil {
		return nil, err
	}

	vmBuilder := vmbuilder.NewVMBuilder(op.VmProviderHelper.getVmBuilderCreator()).
		Name(name).
		Namespace(namespace).
		Labels(instance.Spec.Labels).
		Labels(map[string]string{util.LabelKeyForInstanceType(instance.Spec.InstanceTypeSpec.Name): "true"})
	// Iterate over each interface to configure network interfaces.
	for _, intfStatus := range instance.Status.Interfaces {
		networkName := fmt.Sprintf("%s/%s", namespace, vmbuilder.GetNetworkAttachmentDefinitionName(intfStatus.VlanId))
		interfaceName := defaultInterfaceName
		if intfStatus.Name == storageInterfaceName {
			interfaceName = storageInterfaceName
		}
		vmBuilder = vmBuilder.NetworkInterface(interfaceName, string(kubevirtv1.DiskBusVirtio), "", vmbuilder.NetworkInterfaceTypeBridge, networkName)
	}

	// Continue building VM after configuring network interfaces.
	vmBuilder = op.VmProviderHelper.buildVolume(instance, vmBuilder).
		Memory(ctx, instanceTypeSpec.Memory.Size, int32(instanceTypeSpec.Gpu.Count)).
		CPU(int(instanceTypeSpec.Cpu.Cores)).
		CloudInitDisk("cloudinitdisk", "virtio", false, 0, vmbuilder.CloudInitSource{
			CloudInitType:         vmbuilder.CloudInitTypeNoCloud,
			UserDataSecretName:    name,
			NetworkDataSecretName: name,
		}).
		Run(true).
		Affinity(affinity).
		TopologySpreadConstraints(topologySpreadConstraint)

	// add GPU devices
	if instanceTypeSpec.Gpu.Count >= 1 {
		vmBuilder.HostDevices(ctx, instanceTypeSpec)
	}

	vmToCreate, err := vmBuilder.Build()
	if err != nil {
		return &kubevirtv1.VirtualMachine{}, err
	}
	vmToCreate.Spec.Template.Spec.Domain.CPU = &kubevirtv1.CPU{
		Cores:   instanceTypeSpec.Cpu.Cores,
		Sockets: instanceTypeSpec.Cpu.Sockets,
		Threads: instanceTypeSpec.Cpu.Threads,
		Model:   "host-passthrough",
	}

	return vmToCreate, nil
}

func (op VmInstanceBackend) getAffinityForInstance(instance *cloudv1alpha1.Instance) (*k8sv1.Affinity, error) {
	return vmutil.AffinityForInstance(instance)
}

func (op VmInstanceBackend) getTopologySpreadConstraintsForInstance(instance *cloudv1alpha1.Instance) ([]k8sv1.TopologySpreadConstraint, error) {
	return vmutil.TopologySpreadConstraintsForInstance(instance)
}

func (op VmInstanceBackend) DeleteVM(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.DeleteVM").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name

	// Helper function that indicates that an error other than IsNotFound occurred.
	isUnexpectedError := func(err error) bool {
		return err != nil && !errors.IsNotFound(err)
	}

	// Delete Kubevirt VirtualMachine.
	err := op.KubevirtClient.KubevirtV1().
		VirtualMachines(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
	if isUnexpectedError(err) {
		return reconcile.Result{}, err
	}

	// Determine whether Kubevirt VirtualMachineInstance has been deleted.
	// This should get deleted by Kubevirt after the VM stops.
	_, err = op.KubevirtClient.KubevirtV1().
		VirtualMachineInstances(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if isUnexpectedError(err) {
		return reconcile.Result{}, err
	}
	if err == nil {
		log.Info("Kubevirt VirtualMachineInstances still exists. VM is still running. Requeuing.")
		return reconcile.Result{Requeue: true}, nil
	}
	log.V(9).Info("Kubevirt VirtualMachineInstances not found. VM has stopped.")
	return reconcile.Result{}, nil
}

func (op VmInstanceBackend) DeleteSecret(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.DeleteSecret").Start()
	defer span.End()
	log.Info("BEGIN")
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name

	// Delete Secret
	err := op.K8Client.CoreV1().Secrets(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})

	// Already deleted
	if errors.IsNotFound(err) {
		log.V(9).Info("DeleteSecret: Secret not found", logkeys.Name, name)
		err = nil
	}
	log.Info("END")
	return err
}

func (op VmInstanceBackend) DeletePVC(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VmInstanceBackend.DeletePVC").Start()
	defer span.End()
	namespace := instance.ObjectMeta.Namespace
	name := instance.ObjectMeta.Name
	pvcName := name + "-" + defaultVolumeDisk
	log.Info("BEGIN", logkeys.PvcName, pvcName)

	// Delete PVC
	err := op.K8Client.CoreV1().PersistentVolumeClaims(namespace).
		Delete(ctx, pvcName, metav1.DeleteOptions{})

	// Already deleted
	if errors.IsNotFound(err) {
		log.V(9).Info("DeletePVC: PVC not found")
		err = nil
	}
	log.Info("END")
	return err
}

func (kv kubeVirtProvider) buildVolume(instance *cloudv1alpha1.Instance, vmBuilder *vmbuilder.VMBuilder) *vmbuilder.VMBuilder {
	name := instance.ObjectMeta.Name
	vmImageKey := instance.Spec.MachineImage
	instanceTypeSpec := instance.Spec.InstanceTypeSpec
	dvName := name + "-" + defaultVolumeDisk
	dvOpt := vmbuilder.DataVolumeTemplateOption{
		ImageID:    vmImageKey,
		VolumeMode: vmbuilder.PersistentVolumeModeBlock,
		AccessMode: vmbuilder.PersistentVolumeAccessModeReadWriteMany,
	}
	return vmBuilder.DataVolumeDisk(defaultVolumeDisk, string(kubevirtv1.DiskBusVirtio), false, 1, instanceTypeSpec.Disks[0].Size, dvName, &dvOpt)

}

func (hp harvesterProvider) buildVolume(instance *cloudv1alpha1.Instance, vmBuilder *vmbuilder.VMBuilder) *vmbuilder.VMBuilder {
	name := instance.ObjectMeta.Name
	vmImageKey := instance.Spec.MachineImage
	storageClass := vmbuilder.BuildImageStorageClassName("", vmImageKey)
	instanceTypeSpec := instance.Spec.InstanceTypeSpec
	pvcName := name + "-" + defaultVolumeDisk
	pvcimageId := vmbuilder.DefaultVMNamespace + "/" + vmImageKey
	pvcOpt := vmbuilder.PersistentVolumeClaimOption{
		ImageID:          pvcimageId,
		VolumeMode:       vmbuilder.PersistentVolumeModeBlock,
		AccessMode:       vmbuilder.PersistentVolumeAccessModeReadWriteMany,
		StorageClassName: &storageClass,
	}
	return vmBuilder.
		PVCDisk(defaultVolumeDisk, string(kubevirtv1.DiskBusVirtio), false, false, 1, instanceTypeSpec.Disks[0].Size, pvcName, &pvcOpt)
}

func (kv kubeVirtProvider) getVmBuilderCreator() string {
	return kv.vmBuilderCreator
}

func (hp harvesterProvider) getVmBuilderCreator() string {
	return hp.vmBuilderCreator
}

// Ensures the necessary RoleBinding for PVCs used by DataVolumes is created if missing.
// This RoleBinding is not deleted with a VM DeleteResources because:
// 1. PVCs for DataVolumes rely on it, and deleting it would cause permission issues.
// 2. Recreating or checking its usage on every VM deletion adds unnecessary complexity.
func (kv kubeVirtProvider) createRoleBindingIfNeeded(ctx context.Context, namespace string, k8Client *kubernetes.Clientset) error {
	roleBindingResource := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-cdi-cloner", namespace),
			Namespace: "default",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cdi-cloner",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "default",
				Namespace: namespace,
			},
		},
	}
	roleBindingClient := k8Client.RbacV1().RoleBindings("default")
	if _, err := roleBindingClient.Create(ctx, roleBindingResource, metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// return nil as we are not using rolebinding in harvester
func (hp harvesterProvider) createRoleBindingIfNeeded(ctx context.Context, namespace string, k8Client *kubernetes.Clientset) error {
	return nil
}

func (kv kubeVirtProvider) getVmSecretLabels() map[string]string {
	return map[string]string{}
}

func (hp harvesterProvider) getVmSecretLabels() map[string]string {
	return map[string]string{"harvesterhci.io/cloud-init-template": "harvester"}
}
