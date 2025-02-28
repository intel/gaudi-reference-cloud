// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	controllers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	sdn "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups="",resources=secrets,verbs=create;get;list;patch;update;watch;delete
const defaultOSUser string = "sdp"

func NewBmInstanceReconciler(ctx context.Context, mgr ctrl.Manager, vNetPrivateClient pb.VNetPrivateServiceClient, vNetClient pb.VNetServiceClient, cfg *cloudv1alpha1.BmInstanceOperatorConfig) (*controllers.InstanceReconciler, error) {
	instanceBackend, err := NewBmInstanceBackend(ctx, mgr, cfg)
	if err != nil {
		return nil, err
	}
	return controllers.NewInstanceReconciler(ctx, mgr, vNetPrivateClient, vNetClient, instanceBackend, &cfg.InstanceOperator)
}

type BmInstanceBackend struct {
	client.Client
	Cfg          *cloudv1alpha1.BmInstanceOperatorConfig
	bmcSecrets   BMCSecrets
	bmcInterface bmc.Interface
	sdnClient    *sdn.SDNClient
}

func NewBmInstanceBackend(ctx context.Context, mgr ctrl.Manager, cfg *cloudv1alpha1.BmInstanceOperatorConfig) (*BmInstanceBackend, error) {
	var sdnClient *sdn.SDNClient
	var err error
	if cfg.NetworkConfig.NetworkBackEndType == cloudv1alpha1.NetworkBackEndSDN || cfg.NetworkConfig.NetworkBackEndType == cloudv1alpha1.NetworkBackEndTransition {
		sdnClient, err = sdn.NewSDNClient(ctx, sdn.SDNClientConfig{KubeConfig: cfg.NetworkConfig.NetworkKubeConfig})
		if err != nil {
			return nil, err
		}
	}

	return &BmInstanceBackend{
		Client:    mgr.GetClient(),
		Cfg:       cfg,
		sdnClient: sdnClient,
	}, nil
}

func (b *BmInstanceBackend) BuildController(ctx context.Context, ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.
		Watches(
			&baremetalv1alpha1.BareMetalHost{},
			//&handler.EnqueueRequestForObject{},
			handler.EnqueueRequestsFromMapFunc(b.mapBareMetalHostToInstance),
		).WithOptions(controller.Options{MaxConcurrentReconciles: b.Cfg.MaxConcurrentReconciles})
}

// Switch data
type SwitchInformation struct {
	tenantVlanTag  int64
	storageVlanTag int64
	acceVlanTag    int64
	bgpCommunityId int64
}

// BMC Secrets
type BMCSecrets struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URL      string `yaml:"url"`
	KCS      string `yaml:"kcs"`
}

// Raven Secrets
type RavenSecrets struct {
	Credentials struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"credentials"`
}

// gaudinet
type GaudiNetworkInformation struct {
	NicMac     string `json:"NIC_MAC"`
	NicIP      string `json:"NIC_IP"`
	SubnetMask string `json:"SUBNET_MASK"`
	GatewayMac string `json:"GATEWAY_MAC"`
}

type Gaudinet struct {
	NicNetConfig []GaudiNetworkInformation `json:"NIC_NET_CONFIG"`
}

// Networkdata

type NetworkDataLinkEthernet struct {
	// Type is the type of the ethernet link. It can be one of:
	// bridge, dvs, hw_veb, hyperv, ovs, tap, vhostuser, vif, phy
	Type string `yaml:"type"`

	// Id is the ID of the interface (used for naming)
	ID string `yaml:"id"`

	// MTU is the MTU of the interface
	// +optional
	MTU int `yaml:"mtu,omitempty"`

	// MACAddress is the MAC address of the interface, containing the object
	// used to render it.
	MACAddress string `yaml:"ethernet_mac_address,omitempty"`

	// Name is used for set-name
	Name string `yaml:"name,omitempty"`

	// Bond mode
	BondMode string `yaml:"bond_mode,omitempty"`

	// Bond transmit hash policy
	BondTransmitHashPolicy string `yaml:"bond_xmit_hash_policy,omitempty"`

	// Bond MII monitor interval(verifying if an interface of the bond has carrier)
	BondMiimon int `yaml:"bond_miimon,omitempty"`

	// Bond Links
	BondLinks []string `yaml:"bond_links,omitempty"`
}

type NetworkDataRoutev4 struct {
	// Network is the IPv4 network address
	Network string `yaml:"network"`

	// Gateway is the IPv4 address of the gateway
	Gateway string `yaml:"gateway"`
}

// NetworkDataIPv4 represents an ipv4 static network object.
type NetworkDataIPv4 struct {

	// ID is the network ID (name)
	ID string `yaml:"id"`

	// Link is the link on which the network applies
	Link string `yaml:"link"`

	// IPAddress
	IPAddress string `yaml:"ip_address"`

	// nameservers
	Nameservers []string `yaml:"dns_nameservers,omitempty"`

	// Routes contains a list of IPv4 routes
	Routes []NetworkDataRoutev4 `yaml:"routes,omitempty"`

	// Type
	Type string `yaml:"type"`
}

// NetworkData represents a networkData object.
type NetworkData struct {
	// Links is a structure containing lists of different types objects
	Links []NetworkDataLinkEthernet `yaml:"links,omitempty"`

	// Networks  is a structure containing lists of different types objects
	Networks []NetworkDataIPv4 `yaml:"networks,omitempty"`
}

const (
	HostAnnotation                     = "metal3.io/BareMetalHost"
	metal3SecretType corev1.SecretType = "infrastructure.cloud.intel.com/secret"
	qcow2DiskFormat                    = "qcow2"

	chooseHostRequeueAfter         = 60 * time.Second
	sshAccessRequeueAfter          = 60 * time.Second
	ravenRequeueAfterFailure       = 10 * time.Second
	deleteRequeueAfterFailure      = 20 * time.Second
	associateRequeueAfterFailure   = 30 * time.Second
	requeueAfterKcsChanges         = 5 * time.Second
	requeueAfterHCIChanges         = 5 * time.Second
	requeueAfterNetworkConfChanges = 5 * time.Second
	requeueAfterNetworkConfFailure = 10 * time.Second

	instanceGroupLabel  = "instance-group"
	clusterGroupIdLabel = "cluster-group-id"

	gaudinetPath = "/etc/gaudinet.json"

	rootDiskMaxSize    int64 = 1000000000000 // 1 TB
	tenantInterfaceMTU       = 9000
)

var (
	internalError = fmt.Errorf("failed to associate the instance to a host: Internal Error")
)

func (b BmInstanceBackend) CreateOrUpdateInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BmInstanceBackend.CreateOrUpdateInstance").Start()
	defer span.End()
	log.Info("BEGIN")
	// Set status KCS enable if missing
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionKcsEnabled, corev1.ConditionTrue, cloudv1alpha1.ConditionReasonNone, "")
	// Set status HCI enable if missing
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionHCIEnabled, corev1.ConditionTrue, cloudv1alpha1.ConditionReasonNone, "")

	result, err := b.associate(ctx, instance)
	if err != nil {
		condition := cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionFailed,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            fmt.Sprintf("%v", err),
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
		return result, err
	}
	if !result.IsZero() {
		return result, err
	}
	log.Info("END")
	return ctrl.Result{}, nil
}

// This function deletes the BM, associated volumes and secrets as well.
func (b BmInstanceBackend) DeleteResources(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BmInstanceBackend.DeleteResources").Start()
	defer span.End()
	log.Info("BEGIN")
	// delete BM
	if result, err := b.delete(ctx, instance); err != nil || !result.IsZero() {
		return result, err
	}
	log.Info("END")
	return ctrl.Result{}, nil
}

func (b BmInstanceBackend) setConsumerID(ctx context.Context, instance *cloudv1alpha1.Instance, host *baremetalv1alpha1.BareMetalHost, helper *patch.Helper) error {
	err := b.setHostConsumerRef(ctx, host, instance)
	if err != nil {
		return fmt.Errorf("failed to associate the baremetalhost to the instance: %w", err)
	}

	err = b.setLastAssociatedInstance(ctx, host, instance)
	if err != nil {
		return fmt.Errorf("failed to set last associated instance: %w", err)
	}
	err = helper.Patch(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to update the baremetalhost %w", err)
	}
	helper, err = patch.NewHelper(host, b.Client)
	if err != nil {
		return fmt.Errorf("failed to get updated helper %w", err)
	}
	return nil
}

func (b BmInstanceBackend) associate(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.associate").WithValues(logkeys.InstanceName, instance.Name)
	log.Info("Associating BMH")

	// Check if machine image exists and validate checksum file in repository
	if err := validateMachineImage(b.getImageURL(instance), b.getImageChecksumURL(instance),
		instance.Spec.MachineImageSpec.Md5sum); err != nil {
		log.Error(err, "MachineImage validation failed")
		return ctrl.Result{}, err
	}
	host, helper, err := b.getHost(ctx, instance)
	if err != nil {
		log.Info("Failed to get the BMH for the Instance", logkeys.PatchHelper, helper)
		return ctrl.Result{}, err
	}
	// if host is nill and the instance failed return we should not recover from this
	if host == nil && util.IsInstanceFailed(instance) {
		log.Info("Instance is marked failed")
		return ctrl.Result{}, nil
	}
	// no BMH found, trying to choose from available ones
	if host == nil {
		host, helper, err = b.findHost(ctx, instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		if host == nil {
			log.Info("no available host found. Re-queueing...")
			return ctrl.Result{RequeueAfter: chooseHostRequeueAfter}, nil
		}
		log.Info("Associating instance with host", logkeys.HostName, host.Name)
		result, associateErr := func(ctx context.Context, instance *cloudv1alpha1.Instance, host *baremetalv1alpha1.BareMetalHost) (reconcile.Result, error) {
			err = b.setConsumerID(ctx, instance, host, helper)
			if err != nil {
				return ctrl.Result{}, err
			}

			// setup root device hint and disk partitions
			disks := getHostDisks(ctx, host)
			err = setRootDisk(ctx, host, disks)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set root device hint %w", err)
			}

			// set Userdata
			content, err := b.setUserData(ctx, host, instance, disks)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set userdata for baremetalhost %w", err)
			}
			userDataSecretName := fmt.Sprintf("usr%s-secret", instance.Name)
			err = createSecret(ctx, b.Client,
				userDataSecretName, host.Namespace, map[string][]byte{"userData": content})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create userdata secret %w", err)
			}

			networkData, err := b.setNetworkData(ctx, host, instance)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set network data for baremetalhost %w", err)
			}
			networkDataSecretName := fmt.Sprintf("network-%s-secret", instance.Name)
			err = createSecret(ctx, b.Client, networkDataSecretName, host.Namespace, map[string][]byte{"networkData": networkData})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create network data secret %w", err)
			}

			// ensure that the BMH specs are correctly set.
			err = b.setHostSpec(host, userDataSecretName, networkDataSecretName, instance)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to associate the baremetalhost to the instance %w", err)
			}
			return ctrl.Result{}, nil
		}(ctx, instance, host)
		// remove the hostspec changes if assocaiteErr is not nil
		if associateErr != nil {
			helper, err := patch.NewHelper(host, b.Client)
			if err != nil {
				log.Error(err,
					"failed to associate the instance to host, failed to create the client to patch the host",
					logkeys.HostName, host.Name)
				return result, internalError
			}
			// remove the spec changes to the host
			err = b.removeHostSpec(ctx, host)
			if err != nil {
				associateErr = multierror.Append(associateErr, err)
			}
			err = helper.Patch(ctx, host)
			if err != nil {
				associateErr = multierror.Append(associateErr, err)
			}
			err = fmt.Errorf("%w", associateErr)
			log.Error(err, "failed to associate the instance to a host", logkeys.HostName, host.Name)
			return result, internalError
		}
		// Patch hostspec changes
		err = helper.Patch(ctx, host)
		if err != nil {
			log.Error(err, "failed to update the baremetalhost", logkeys.HostName, host.Name)
			return ctrl.Result{}, internalError
		}
		//patch instance with baremetal host annotation
		err = b.ensureAnnotation(ctx, host, instance)
		if err != nil {
			log.Error(err, "failed to annotate the baremetalhost", logkeys.HostName, host.Name)
			return ctrl.Result{}, internalError
		}
	} else {
		log.Info("instance already associated with host", logkeys.HostName, host.Name)
	}
	// check for host errors and report if any
	var condition cloudv1alpha1.InstanceCondition
	if host.Status.ErrorMessage != "" {
		// Check if the host is pingeable first
		if util.IsInstanceStartupCompleted(instance) {
			// try ssh ping first to verify if the error affected the host availability
			if err := b.verifySSHAccess(ctx, instance); err != nil {
				log.Error(fmt.Errorf("cannot ssh ping the host after baremetahost reported error %v", err), host.Status.ErrorMessage, logkeys.HostName, host.Name, logkeys.ErrorType, host.Status.ErrorType)
				return ctrl.Result{}, fmt.Errorf("errors occured on the reserved host")
			} else {
				//ignore the reported error from baremetalhosts
				log.Error(fmt.Errorf("errors occured on the reserved host (ignored)"), host.Status.ErrorMessage, logkeys.HostName, host.Name, logkeys.ErrorType, host.Status.ErrorType)
				return ctrl.Result{}, nil
			}
		} else {
			log.Error(fmt.Errorf("errors occured on the reserved host"), host.Status.ErrorMessage, logkeys.HostName, host.Name, logkeys.ErrorType, host.Status.ErrorType)
			return ctrl.Result{}, fmt.Errorf("errors occured on the reserved host")
		}
	} else {
		condition = cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionFailed,
			Status:             corev1.ConditionFalse,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "",
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
	}

	//Update based on state
	switch host.Status.Provisioning.State {
	case baremetalv1alpha1.StateProvisioned:
		// create host ref for events
		hostRef := corev1.ObjectReference{
			Namespace: host.Namespace,
			Name:      host.Name,
			Kind:      "BareMetalHost",
		}
		// If the Instance is already running, return success
		// power off the instance using RunStrategy only if it's in 'Ready' state to prevent interruption in provisioning
		if util.IsInstanceStarted(instance) {
			// update instance condition based on RunStrategy
			if instance.Spec.RunStrategy == cloudv1alpha1.RunStrategyHalted {
				err = b.updateHostOnlineStatus(ctx, host, false)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to update host power status %v", err)
				}

				log.Info("baremetalhost is powering off")
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopping)
				log.Info("instance is stopping")
			}
			return ctrl.Result{}, nil
		}

		// if instance is stopping, update instance condition on basis of host.Status.PoweredOn
		if util.IsInstanceStopping(instance) {
			if host.Status.PoweredOn {
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopping)
				log.Info("host is powering off")
				log.Info("instance is in stopping state")
			}

			if !host.Status.PoweredOn {
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopped)
				log.Info("instance has stopped")
			}
			return ctrl.Result{}, nil
		}

		// if instance is stopped, update instance condition based on RunStrategy
		if util.IsInstanceStoppedCompleted(instance) {
			if instance.Spec.RunStrategy == cloudv1alpha1.RunStrategyAlways || instance.Spec.RunStrategy == cloudv1alpha1.RunStrategyRerunOnFailure {
				err = b.updateHostOnlineStatus(ctx, host, true)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to update host power status %v", err)
				}

				log.Info("baremetalhost is powering on")
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarting)

				if host.Status.PoweredOn {
					b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionRunning)
				}
			}
			return ctrl.Result{}, nil
		}

		// if the instance is starting and SSH access is verified
		// Move to Ready state if the host has completed powering on
		// Move to Starting state if the host has not completed powering on
		if util.IsInstanceStarting(instance) && util.IsInstanceVerifiedSshAccess(instance) {
			if host.Status.PoweredOn {
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarted)
			} else {
				b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarting)
			}
			return ctrl.Result{}, nil
		}

		if util.IsInstanceAgentConnected(instance) {
			if b.Cfg.SshConfig.WaitForSSHAccess {
				if util.IsInstanceVerifiedSshAccess(instance) {
					b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarted)
					return ctrl.Result{}, nil
				}
				if err := b.verifySSHAccess(ctx, instance); err != nil {
					log.Info("unable to verify SSH Access. Re-queuing...", logkeys.Error, err)
					event := instance.NewEvent("Verify SSH Access", fmt.Sprintf("failed to ping SSH %v", err), &hostRef)
					b.publishEvent(ctx, event)
					return ctrl.Result{RequeueAfter: sshAccessRequeueAfter}, nil
				} else {
					// publish an event that ssh access is verified
					event := instance.NewEvent("Verify SSH Access", "Completed succesfully", &hostRef)
					b.publishEvent(ctx, event)
					condition = cloudv1alpha1.InstanceCondition{
						Type:               cloudv1alpha1.InstanceConditionVerifiedSshAccess,
						Status:             corev1.ConditionTrue,
						LastProbeTime:      metav1.Now(),
						LastTransitionTime: metav1.Now(),
					}
					util.SetStatusCondition(&instance.Status.Conditions, condition)

					b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarted)
					condition = cloudv1alpha1.InstanceCondition{
						Type:               cloudv1alpha1.InstanceConditionStartupComplete,
						Status:             corev1.ConditionTrue,
						LastProbeTime:      metav1.Now(),
						LastTransitionTime: metav1.Now(),
					}
					util.SetStatusCondition(&instance.Status.Conditions, condition)
				}
			}
		} else if util.IsInstanceHCIEnabled(instance) {
			// disable Redfish host interface
			return b.disableHCIAccess(ctx, host, instance)
		} else if util.IsInstanceKcsEnabled(instance) {
			// disable KCS interface in the host
			return b.disableKCSAccess(ctx, host, instance)
		} else {
			return b.updateVnetAccess(ctx, host, instance, &hostRef)
		}
	case baremetalv1alpha1.StateProvisioning:
		log.Info("baremetalhost provisioning")
		condition = cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionRunning,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
	default:
		log.Info("baremetalhost Accepted")
		condition = cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionAccepted,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
	}
	return ctrl.Result{}, nil
}

func (b BmInstanceBackend) isAllInstancesDeleted(ctx context.Context, instance *cloudv1alpha1.Instance) (error, bool) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.isAllInstancesDeleted")
	ns := instance.Namespace
	instanceLabels := &client.MatchingLabels{
		instanceGroupLabel:  instance.Labels[instanceGroupLabel],
		clusterGroupIdLabel: instance.Labels[clusterGroupIdLabel],
	}
	instanceList := cloudv1alpha1.InstanceList{}
	err := b.Client.List(ctx, &instanceList, client.InNamespace(ns), instanceLabels)
	if err != nil {
		log.Error(err, fmt.Sprintf("unable to list instance in %q namespace", ns))
		return fmt.Errorf("unable to list instance in %q namespace %v", ns, err), false
	}
	if len(instanceList.Items) > 0 {
		currentCount := int32(len(instanceList.Items))
		deleteCount := 0
		for _, instance := range instanceList.Items {
			if instance.DeletionTimestamp != nil {
				deleteCount++
			}
		}
		if currentCount == int32(deleteCount) {
			return nil, true
		}
	}
	return nil, false
}

func (b BmInstanceBackend) updateInstanceConditions(ctx context.Context, instance *cloudv1alpha1.Instance, trueCondition cloudv1alpha1.InstanceConditionType) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.updateInstanceConditions")

	allConditions := []cloudv1alpha1.InstanceConditionType{
		cloudv1alpha1.InstanceConditionStarting,
		cloudv1alpha1.InstanceConditionStarted,
		cloudv1alpha1.InstanceConditionStopping,
		cloudv1alpha1.InstanceConditionStopped,
	}

	for _, condition := range allConditions {
		if condition == trueCondition {
			util.SetInstanceCondition(ctx, instance, condition, corev1.ConditionTrue)
		} else {
			util.SetInstanceCondition(ctx, instance, condition, corev1.ConditionFalse)
		}
	}

	log.Info("instance status condition updated")
}

func (b BmInstanceBackend) delete(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.delete").WithValues(logkeys.InstanceName, instance.Name)
	host, helper, err := b.getHost(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	if host == nil {
		log.Info("host not found for Instance")
		return ctrl.Result{}, nil
	}

	// Update BMC interface
	if err = b.updateBMCInterface(ctx, host); err != nil {
		log.Error(err, "Failed to update BMC interface")
		return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
	}

	// check if host is already deprovisioning
	if !util.IsInstanceKcsEnabled(instance) {

		err = b.bmcInterface.PowerOffBMC(ctx)
		if err != nil {
			log.Error(err, "unable to power off host")
			return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
		}

		// Set KCS to provisioning mode
		err = b.bmcInterface.EnableKCS(ctx)
		if err != nil && !errors.Is(err, bmc.ErrKCSNotSupported) {
			log.Error(err, "Unable to Enable KCS")
			return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
		} else {
			// set status to KCS enabled
			var message = ""
			if errors.Is(err, bmc.ErrKCSNotSupported) {
				message = "KCS interface not supported"
			} else {
				message = "KCS interface enabled"
			}
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionKcsEnabled,
				Status:             corev1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            message,
			}
			util.SetStatusCondition(&instance.Status.Conditions, condition)
			return ctrl.Result{RequeueAfter: requeueAfterKcsChanges}, nil
		}
	}
	// enable HostInterface interface
	if !util.IsInstanceHCIEnabled(instance) {
		log.Info("enabling Host Interface")
		err = b.bmcInterface.EnableHCI(ctx)
		if err != nil && !errors.Is(err, bmc.ErrHCINotSupported) {
			log.Error(err, "failed to enable HCI")
			return ctrl.Result{RequeueAfter: associateRequeueAfterFailure}, nil
		} else {
			// Update Condition message
			var message = ""
			if errors.Is(err, bmc.ErrHCINotSupported) {
				message = "host interface not supported"
			} else {
				message = "host interface enabled"
			}
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionHCIEnabled,
				Status:             corev1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            message,
			}
			util.SetStatusCondition(&instance.Status.Conditions, condition)
			return ctrl.Result{RequeueAfter: requeueAfterHCIChanges}, nil
		}
	}
	if b.Cfg.NetworkConfig.NetworkBackEndType == cloudv1alpha1.NetworkBackEndSDN {
		// Switch network to provisioning vlan
		hostSwitchInformation := b.getSwitchInformation(ctx, host, instance)
		description := fmt.Sprintf("bmaas bmh name %s, namespace %s", host.Name, host.Namespace)

		checkNetworkNodeRequest := sdn.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           host.Name,
			DesiredFrontEndFabricVlan: int64(b.Cfg.NetworkConfig.ProvisioningVlan),
		}

		if hostSwitchInformation.acceVlanTag != 0 {
			checkNetworkNodeRequest.DesiredAcceleratorFabricVlan = int64(b.Cfg.NetworkConfig.AcceleratorNetworkDefaultVlan)
		}

		if hostSwitchInformation.bgpCommunityId != 0 {
			// check if all instances are being deleted
			if err, ok := b.isAllInstancesDeleted(ctx, instance); ok {
				checkNetworkNodeRequest.DesiredAcceleratorFabricBGPCommunityID = int64(b.Cfg.NetworkConfig.AccBGPNetworkDefaultCommunityID)
			} else {
				if err != nil {
					log.Error(err, "failed to read the status of other instances in the same group")
					return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
				}
			}
		}

		if hostSwitchInformation.storageVlanTag != 0 {
			checkNetworkNodeRequest.DesiredStorageFabricVlan = int64(b.Cfg.NetworkConfig.StorageDefaultVlan)
		}

		checkNetworkNodeResponse, err := b.sdnClient.CheckNetworkNodeStatus(ctx, checkNetworkNodeRequest)
		if err != nil {
			log.Error(err, "check NetworkNode status failed")
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
		}

		if checkNetworkNodeResponse.Status == sdn.UpdateNotStarted {
			updateNetworkNodeRequest := sdn.NetworkNodeConfUpdateRequest{
				NetworkNodeName:    host.Name,
				FrontEndFabricVlan: int64(b.Cfg.NetworkConfig.ProvisioningVlan),
				Description:        description,
			}

			if hostSwitchInformation.acceVlanTag != 0 {
				updateNetworkNodeRequest.AcceleratorFabricVlan = int64(b.Cfg.NetworkConfig.AcceleratorNetworkDefaultVlan)
			}

			if hostSwitchInformation.bgpCommunityId != 0 {
				if hostSwitchInformation.bgpCommunityId != 0 {
					if err, ok := b.isAllInstancesDeleted(ctx, instance); ok {
						updateNetworkNodeRequest.AcceleratorFabricBGPCommunityID = int64(b.Cfg.NetworkConfig.AccBGPNetworkDefaultCommunityID)
					} else {
						if err != nil {
							log.Error(err, "failed to read the status of other instances in the same group")
							return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
						}
					}
				}
			}

			if hostSwitchInformation.storageVlanTag != 0 {
				updateNetworkNodeRequest.StorageFabricVlan = int64(b.Cfg.NetworkConfig.StorageDefaultVlan)
			}

			err = b.sdnClient.UpdateNetworkNodeConfig(ctx, updateNetworkNodeRequest)
			if err != nil {
				log.Info("sdnClient failed to update NetworkNode", logkeys.Error, err)
				return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
			}
			log.Info("finished updating SDN NetworkNode CR", logkeys.Request, updateNetworkNodeRequest)
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfChanges}, nil
		} else if checkNetworkNodeResponse.Status == sdn.UpdateInProgress {
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfChanges}, nil
		} else if checkNetworkNodeResponse.Status == sdn.UpdateCompleted {
			log.Info("NetworkNode vlan update completed", logkeys.NetworkNodeName, host.Name)
		}
	} else {
		log.Info("Network backend is None or not supported", logkeys.NetworkBackendType, b.Cfg.NetworkConfig.NetworkBackEndType)
	}
	// Delete host provisioning spec
	err = b.removeHostSpec(ctx, host)
	if err != nil {
		log.Error(err, "failed to start deprovision the baremetalhost associated with to this instance")
		return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
	}
	err = helper.Patch(ctx, host)
	if err != nil {
		log.Error(err, "failed to start deprovision the baremetalhost associated with to this instance")
		return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
	}

	if _, ok := host.Labels[bmenrollment.LastClusterGroup]; ok {
		delete(host.Labels, bmenrollment.LastClusterGroup)
	}

	// Trigger validation
	log.Info("Trigger Validation using ReadyToTestLabel")
	if _, ok := host.Labels[bmenrollment.DeletionLabel]; !ok {
		// validate only if the delete step is via a normal delete.
		delete(host.Labels, bmenrollment.VerifiedLabel)
		host.Labels[bmenrollment.ReadyToTestLabel] = "true"
	}
	delete(host.Labels, bmenrollment.DeletionLabel)

	err = helper.Patch(ctx, host)
	if err != nil {
		log.Error(err, "failed to associate the baremetalhost to the instance")
		return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
	}

	instanceHelper, err := patch.NewHelper(instance, b.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Delete the reference between this instance and the baremetal host
	delete(instance.Annotations, HostAnnotation)
	err = instanceHelper.Patch(ctx, instance)
	if err != nil {
		log.Error(err, "failed to remove associate baremetalhost from instance annotation")
		return ctrl.Result{RequeueAfter: deleteRequeueAfterFailure}, nil
	}
	return ctrl.Result{}, nil
}

func (b BmInstanceBackend) updateHostOnlineStatus(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, onlineStatus bool) error {
	helper, err := patch.NewHelper(host, b.Client)
	if err != nil {
		return fmt.Errorf("failed to get updated helper %w", err)
	}
	host.Spec.Online = onlineStatus
	err = helper.Patch(ctx, host)
	if err != nil {
		return fmt.Errorf("unable to update host.Spec.Online %v", err)
	}

	return nil
}

func (b BmInstanceBackend) setLastAssociatedInstance(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) error {
	host.Labels[bmenrollment.LastAssociatedInstance] = instance.Name
	if instance.Spec.InstanceGroup != "" {
		host.Labels[bmenrollment.LastClusterGroup] = instance.Spec.InstanceGroup
	}
	return nil
}

func (b *BmInstanceBackend) mapBareMetalHostToInstance(ctx context.Context, object client.Object) []reconcile.Request {
	bmh := baremetalv1alpha1.BareMetalHost{}
	key := client.ObjectKey{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
	err := b.Get(ctx, key, &bmh)
	if err != nil {
		return nil
	}
	if bmh.Spec.ConsumerRef == nil {
		return nil
	}
	var instanceList cloudv1alpha1.InstanceList
	if err := b.List(ctx, &instanceList,
		client.InNamespace(bmh.Spec.ConsumerRef.Namespace),
	); err != nil {
		return nil
	}
	requests := []reconcile.Request{}
	for _, item := range instanceList.Items {
		if item.Name == bmh.Spec.ConsumerRef.Name {
			requests = append(requests,
				reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&item),
				},
			)
		}
	}
	return requests
}

func (b BmInstanceBackend) ensureAnnotation(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) error {
	helper, err := patch.NewHelper(instance, b.Client)
	if err != nil {
		return fmt.Errorf("error patch.newhelper: %w", err)
	}
	annotations := instance.ObjectMeta.GetAnnotations()

	if annotations == nil {
		annotations = make(map[string]string)
	}
	hostKey, err := cache.MetaNamespaceKeyFunc(host)
	if err != nil {
		return fmt.Errorf("error parsing annotation value: %w", err)
	}
	existing, ok := annotations[HostAnnotation]
	if ok {
		if existing == hostKey {
			return nil
		}
	}
	annotations[HostAnnotation] = hostKey
	instance.ObjectMeta.SetAnnotations(annotations)
	err = helper.Patch(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to add instance annotation: %w", err)
	}
	return nil
}

func (b BmInstanceBackend) setHostConsumerRef(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) error {
	host.Spec.ConsumerRef = &corev1.ObjectReference{
		Kind:       "Instance",
		Name:       instance.Name,
		Namespace:  instance.Namespace,
		APIVersion: instance.APIVersion,
	}
	return nil
}

func (b BmInstanceBackend) getHost(ctx context.Context, instance *cloudv1alpha1.Instance,
) (*baremetalv1alpha1.BareMetalHost, *patch.Helper, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.getHost")
	annotations := instance.ObjectMeta.GetAnnotations()
	if annotations == nil {
		return nil, nil, nil
	}
	hostKey, ok := annotations[HostAnnotation]
	if !ok {
		return nil, nil, nil
	}
	hostNamespace, hostName, err := cache.SplitMetaNamespaceKey(hostKey)
	if err != nil {
		log.Error(err, "Error parsing annotation value", logkeys.HostKey, hostKey)
		return nil, nil, err
	}

	host := baremetalv1alpha1.BareMetalHost{}
	key := client.ObjectKey{
		Name:      hostName,
		Namespace: hostNamespace,
	}
	err = b.Client.Get(ctx, key, &host)
	if apierrors.IsNotFound(err) {
		log.Info("Annotated host not found", logkeys.HostKey, hostKey)
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	helper, err := patch.NewHelper(&host, b.Client)
	if err != nil {
		return nil, nil, err
	}

	return &host, helper, nil
}

func (b BmInstanceBackend) findHost(ctx context.Context, instance *cloudv1alpha1.Instance) (*baremetalv1alpha1.BareMetalHost, *patch.Helper, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.findHost")

	host := baremetalv1alpha1.BareMetalHost{}
	key := client.ObjectKey{
		Name:      instance.Labels["node-id"],
		Namespace: instance.Labels["cluster-id"],
	}
	err := b.Client.Get(ctx, key, &host)
	if apierrors.IsNotFound(err) {
		log.Info("Annotated host not found", logkeys.HostKey, instance.Labels["node-id"])
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	// Check if host is used(Safety Check)
	if host.Spec.ConsumerRef != nil {
		return nil, nil, fmt.Errorf("system reserved for resource %s; Namespace %s", host.Spec.ConsumerRef.Name, host.Spec.ConsumerRef.Namespace)
	}
	helper, err := patch.NewHelper(&host, b.Client)
	if err != nil {
		return nil, nil, err
	}
	return &host, helper, nil
}

// reference and bare metal instance metadata match.
func consumerRefMatches(consumer *corev1.ObjectReference, instance *cloudv1alpha1.Instance) bool {
	if consumer.Name != instance.Name {
		return false
	}
	if consumer.Namespace != instance.Namespace {
		return false
	}
	if consumer.Kind != instance.Kind {
		return false
	}
	if consumer.GroupVersionKind().Group != instance.GroupVersionKind().Group {
		return false
	}
	return true
}

// Get the Machine Image url
func (b BmInstanceBackend) getImageURL(instance *cloudv1alpha1.Instance) string {
	return fmt.Sprintf("%s/%s.qcow2", b.Cfg.OsHttpServerUrl, instance.Spec.MachineImage)
}

// Get the Machine Image checksum url
func (b BmInstanceBackend) getImageChecksumURL(instance *cloudv1alpha1.Instance) string {
	return fmt.Sprintf("%s/%s.qcow2.md5sum", b.Cfg.OsHttpServerUrl, instance.Spec.MachineImage)
}

// validateMachineImage checks if a file exists at the given URL and validates the checksum file.
func validateMachineImage(imageUrl string, checksumUrl string, specChecksum string) error {
	// specChecksum being empty implies that the compute api server has not been updated or an older crd
	if specChecksum == "" {
		return nil
	}

	// Check if the image exists at the imageUrl.
	client := &http.Client{
		Timeout: 60 * time.Second, // Set a timeout for the request
	}
	// Make a HEAD request
	resp, err := client.Head(imageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("machineImage not found")
		} else {
			return fmt.Errorf("failed to validate MachineImage, unexpected status code: %d", resp.StatusCode)
		}
	}

	// Machine image exists, lets compare checksums
	chksum, err := readChecksumFromURL(checksumUrl)
	if err != nil {
		return fmt.Errorf("not able to read MachineImage checksum file: %w", err)
	}

	if specChecksum == chksum {
		return nil
	} else {
		return fmt.Errorf("check sum mismatch. fromImageSpec: %s, fromFile: %s", specChecksum, chksum)
	}
}

// reads the checksum from the given URL.
func readChecksumFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for URL: %s", resp.StatusCode, url)
	}

	scanner := bufio.NewScanner(resp.Body)
	if scanner.Scan() {
		line := scanner.Text()
		// Extract the checksum from the line
		splitStrings := strings.Fields(line)
		if len(splitStrings) > 0 {
			return splitStrings[0], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read checksum: %w", err)
	}

	return "", nil
}

func (b BmInstanceBackend) setHostSpec(host *baremetalv1alpha1.BareMetalHost,
	userDataSecretName string, networkDataSecretName string, instance *cloudv1alpha1.Instance) error {
	//TODO get image and checksum from the image service
	diskFormat := qcow2DiskFormat
	if host.Spec.Image == nil {
		host.Spec.Image = &baremetalv1alpha1.Image{
			URL:        b.getImageURL(instance),
			Checksum:   b.getImageChecksumURL(instance),
			DiskFormat: &diskFormat,
		}
		host.Spec.Online = true
	}
	if host.Spec.UserData == nil {
		host.Spec.UserData = &corev1.SecretReference{
			Name:      userDataSecretName,
			Namespace: host.Namespace,
		}
	}

	if host.Spec.NetworkData == nil {
		host.Spec.NetworkData = &corev1.SecretReference{
			Name:      networkDataSecretName,
			Namespace: host.Namespace,
		}
	}
	return nil
}

func (b BmInstanceBackend) removeHostSpec(ctx context.Context, host *baremetalv1alpha1.BareMetalHost) error {
	if host.Spec.Image != nil {
		host.Spec.Image = nil
		host.Spec.Online = true
	}
	if host.Spec.UserData != nil {
		userDataSecretName := host.Spec.UserData.Name
		namespace := host.Spec.UserData.Namespace
		host.Spec.UserData = nil
		err := deleteSecret(ctx, b.Client, userDataSecretName, namespace)
		if err != nil {
			return fmt.Errorf("failed to delete the secret: %w", err)
		}
	}
	if host.Spec.NetworkData != nil {
		networkDataSecretName := host.Spec.NetworkData.Name
		namespace := host.Spec.NetworkData.Namespace
		host.Spec.NetworkData = nil
		err := deleteSecret(ctx, b.Client, networkDataSecretName, namespace)
		if err != nil {
			return fmt.Errorf("failed to delete the secret: %w", err)
		}
	}
	if host.Spec.ConsumerRef != nil {
		host.Spec.ConsumerRef = nil
	}
	if host.Spec.RootDeviceHints != nil {
		host.Spec.RootDeviceHints = &baremetalv1alpha1.RootDeviceHints{}
	}
	return nil
}

func (b BmInstanceBackend) setNetworkData(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) ([]byte, error) {
	networkData := NetworkData{}
	intfStatus := instance.Status.Interfaces[0]
	intfSpec := instance.Spec.Interfaces[0]
	address := intfStatus.Addresses[0]
	interfaceName := fmt.Sprintf("%s-tenant", intfSpec.Name)
	networkDataLinkEthernet := NetworkDataLinkEthernet{}
	networkDataLinkEthernet.ID = interfaceName
	networkDataLinkEthernet.Name = interfaceName
	networkDataLinkEthernet.Type = "phy"
	networkDataLinkEthernet.MTU = tenantInterfaceMTU
	networkDataLinkEthernet.MACAddress = host.Spec.BootMACAddress

	// routes
	networkRoute := NetworkDataRoutev4{}
	networkRoute.Gateway = intfStatus.Gateway
	networkRoute.Network = "0.0.0.0/0"
	// ipv4
	networkIp4 := NetworkDataIPv4{}
	networkIp4.ID = fmt.Sprintf("%s-ipv4", interfaceName)
	networkIp4.Link = interfaceName
	networkIp4.Type = "ipv4"
	networkIp4.IPAddress = fmt.Sprintf("%s/%d", address, intfStatus.PrefixLength)
	networkIp4.Nameservers = append(networkIp4.Nameservers, intfSpec.Nameservers...)
	networkIp4.Routes = append(networkIp4.Routes, networkRoute)
	networkData.Links = append(networkData.Links, networkDataLinkEthernet)
	networkData.Networks = append(networkData.Networks, networkIp4)

	// Add storage network
	for _, intfStatus := range instance.Status.Interfaces {
		if intfStatus.Name == util.StorageInterfaceName {
			storageMacAnnotations := b.getStorageMacAnnotations(host)
			switch len(storageMacAnnotations) {
			case 0:
				return nil, fmt.Errorf("storage MAC annotations are missing in the BMH Spec")
			case 1:
				// Storage Link
				storageInterfaceName := fmt.Sprintf("%s-tenant", intfStatus.Name)
				storageNetworkDataLinkEthernet := NetworkDataLinkEthernet{}
				storageNetworkDataLinkEthernet.ID = storageInterfaceName
				storageNetworkDataLinkEthernet.Name = storageInterfaceName
				storageNetworkDataLinkEthernet.Type = "phy"
				storageNetworkDataLinkEthernet.MTU = tenantInterfaceMTU
				intfIndex := strings.Replace(strings.Trim(bmenrollment.StorageInterfaceName1, "net"), "/", "-", -1)
				storageNetworkDataLinkEthernet.MACAddress = storageMacAnnotations[fmt.Sprintf("%s/eth%s", bmenrollment.StorageMACAnnotationPrefix, intfIndex)]
				// Storage ipv4
				storageNetworkIp4 := NetworkDataIPv4{}
				storageNetworkIp4.ID = fmt.Sprintf("%s-ipv4", storageInterfaceName)
				storageNetworkIp4.Link = storageInterfaceName
				storageNetworkIp4.Type = "ipv4"
				storageNetworkIp4.IPAddress = fmt.Sprintf("%s/%d", intfStatus.Addresses[0], intfStatus.PrefixLength)
				// Storage routes
				for _, serverSubnet := range b.Cfg.StorageServerSubnets {
					storageNetworkRoute := NetworkDataRoutev4{}
					storageNetworkRoute.Gateway = intfStatus.Gateway
					storageNetworkRoute.Network = serverSubnet
					storageNetworkIp4.Routes = append(storageNetworkIp4.Routes, storageNetworkRoute)
				}
				networkData.Links = append(networkData.Links, storageNetworkDataLinkEthernet)
				networkData.Networks = append(networkData.Networks, storageNetworkIp4)
			case 2:
				// Bond-individual interfaces
				bondLinks := []string{}
				var storageInterfaceName string
				for _, storageMacAddress := range storageMacAnnotations {
					storageInterfaceName = fmt.Sprintf("%s-%s", intfStatus.Name, storageMacAddress)
					storageNetworkDataLinkEthernet := NetworkDataLinkEthernet{}
					storageNetworkDataLinkEthernet.ID = storageInterfaceName
					storageNetworkDataLinkEthernet.Name = storageInterfaceName
					storageNetworkDataLinkEthernet.Type = "phy"
					storageNetworkDataLinkEthernet.MTU = tenantInterfaceMTU
					storageNetworkDataLinkEthernet.MACAddress = storageMacAddress
					networkData.Links = append(networkData.Links, storageNetworkDataLinkEthernet)
					bondLinks = append(bondLinks, storageInterfaceName)
				}
				// Bond link
				bondInterfaceName := fmt.Sprintf("%s-tenant", intfStatus.Name)
				storageNetworkBond := NetworkDataLinkEthernet{}
				storageNetworkBond.ID = bondInterfaceName
				storageNetworkBond.Name = bondInterfaceName
				storageNetworkBond.BondLinks = bondLinks
				storageNetworkBond.Type = "bond"
				storageNetworkBond.BondMiimon = 100
				storageNetworkBond.BondTransmitHashPolicy = "layer3+4"
				storageNetworkBond.BondMode = "802.3ad"
				networkData.Links = append(networkData.Links, storageNetworkBond)
				// Storage bond ipv4
				storageNetworkIp4 := NetworkDataIPv4{}
				storageNetworkIp4.ID = fmt.Sprintf("%s-ipv4", bondInterfaceName)
				storageNetworkIp4.Link = bondInterfaceName
				storageNetworkIp4.Type = "ipv4"
				storageNetworkIp4.IPAddress = fmt.Sprintf("%s/%d", intfStatus.Addresses[0], intfStatus.PrefixLength)
				// Storage routes
				for _, serverSubnet := range b.Cfg.StorageServerSubnets {
					storageNetworkRoute := NetworkDataRoutev4{}
					storageNetworkRoute.Gateway = intfStatus.Gateway
					storageNetworkRoute.Network = serverSubnet
					storageNetworkIp4.Routes = append(storageNetworkIp4.Routes, storageNetworkRoute)
				}
				networkData.Networks = append(networkData.Networks, storageNetworkIp4)

			default:
				return nil, fmt.Errorf("storage MAC annotations in the BMH Spec is greater than the supported count for storage")
			}

		}
	}
	return yaml.Marshal(networkData)
}

func (b BmInstanceBackend) setExtraNetworkData(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, cloudConfig *util.CloudConfig, instance *cloudv1alpha1.Instance) error {
	// find all annotations with mac GPU
	gpuMacs := b.filterAnnotations(host, bmenrollment.GPUAnnotationPrefix)
	// find all annotation with IPs for GPUs Ethernet
	gpuIPsAssigned := b.filterAnnotations(host, bmenrollment.GPUIPsAnnotationPrefix)
	// gaudinet
	gaudinet := Gaudinet{}
	gaudinet.NicNetConfig = []GaudiNetworkInformation{}

	// this is only required for L3/BGP configuration
	if _, ok := host.Labels[bmenrollment.NetworkModeLabel]; ok {
		if host.Labels[bmenrollment.NetworkModeLabel] == bmenrollment.NetworkModeXBX {
			for key, value := range gpuMacs {
				// Replace prefix
				newKey := strings.Replace(key, bmenrollment.GPUAnnotationPrefix, bmenrollment.GPUIPsAnnotationPrefix, -1)
				// find the IP
				ipAddress, ok := gpuIPsAssigned[newKey]
				if ok {
					// Check for gateway MAC
					gaudinetNetworkEntry := GaudiNetworkInformation{}
					// get switch mac address
					switchMacAddress, err := b.findGatewayMac(value, host)
					if err != nil {
						return fmt.Errorf("failed to retrieve switch mac address %v", err)
					}
					gaudinetNetworkEntry.GatewayMac = switchMacAddress
					subnetMask, err := b.getMaskString(ipAddress)
					if err != nil {
						return fmt.Errorf("failed to retrieve mask string %v", err)
					}
					gaudinetNetworkEntry.SubnetMask = subnetMask
					gaudinetNetworkEntry.NicIP = strings.Split(ipAddress, "/")[0]
					gaudinetNetworkEntry.NicMac = value
					// get ply subnet
					destination := net.ParseIP(strings.Split(ipAddress, "/")[0]).To4()
					if destination == nil {
						return fmt.Errorf("failed to convert ipAddress into 4 bytes ip4")
					}
					//Convert to /16
					destination[2] = 0
					destination[3] = 0
					// create the IP assigned to switch port
					switchPortIP := net.ParseIP(strings.Split(ipAddress, "/")[0]).To4()
					switchPortIP[3]++
					// add it to gaudlet
					gaudinet.NicNetConfig = append(gaudinet.NicNetConfig, gaudinetNetworkEntry)
					ethernetIndex := strings.Replace(key, fmt.Sprintf("%s/", bmenrollment.GPUAnnotationPrefix), "", -1)
					networkLinkFilePath := fmt.Sprintf("/etc/systemd/network/10-%s.link", ethernetIndex)
					networkLinkFileContent := fmt.Sprintf(`[Match]
MACAddress=%s
[Link]
Name=%s
MTUBytes=8000
`, value, ethernetIndex)

					networkFilePath := fmt.Sprintf("/etc/systemd/network/10-%s.network", ethernetIndex)
					networkFileContent := fmt.Sprintf(`[Match]
MACAddress=%s
[Link]
Name=%s
MTUBytes=8000
[Network]
LinkLocalAddressing=no
Address=%s
[Route]
Gateway=%s
Destination=%s/16
`, value, ethernetIndex, ipAddress, switchPortIP.String(), destination.String())
					cloudConfig.AddWriteFile(networkFilePath, networkFileContent)
					cloudConfig.AddWriteFile(networkLinkFilePath, networkLinkFileContent)
				}
			}
		}
	}

	if len(gaudinet.NicNetConfig) > 0 {
		out, err := json.Marshal(gaudinet)
		if err != nil {
			return fmt.Errorf("failed to marshal gaudinetWriteFile %v", err)
		}
		// added to write Files
		cloudConfig.AddWriteFile(gaudinetPath, string(out))
	}
	// enable rc.local
	cloudConfig.AddRunCmd("systemctl enable rc-local")
	cloudConfig.AddRunCmd("systemctl start rc-local")
	// disable ondemand
	cloudConfig.AddRunCmd("systemctl disable ondemand")
	//restart networkd
	cloudConfig.AddRunCmd("systemctl restart systemd-networkd")
	// rclocal script
	var rcLocalContent string
	if instance.Spec.InstanceGroup != "" {
		rcLocalContent = `#!/bin/sh
echo -n "performance" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
sleep 2
systemctl restart systemd-udev-trigger.service
sleep 5
/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
exit 0
`
	} else {
		rcLocalContent = `#!/bin/sh
echo -n "performance" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
exit 0
`
	}

	cloudConfig.AddWriteFileWithPermissions("/etc/rc.local", rcLocalContent, "0755")
	return nil
}

func (b BmInstanceBackend) filterAnnotations(host *baremetalv1alpha1.BareMetalHost, filterString string) map[string]string {
	hostAnnotations := host.GetAnnotations()
	filteredAnnotations := map[string]string{}
	for key, value := range hostAnnotations {
		if strings.Contains(key, filterString) {
			filteredAnnotations[key] = value
		}
	}
	return filteredAnnotations
}

func (b BmInstanceBackend) getMaskString(cidr string) (string, error) {
	_, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	if len(ipv4Net.Mask) != 4 {
		return "", fmt.Errorf("wrong mac address length %d", len(ipv4Net.Mask))
	}
	mask := ipv4Net.Mask
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]), nil
}

func (b BmInstanceBackend) findGatewayMac(hostMacAddress string, host *baremetalv1alpha1.BareMetalHost) (string, error) {
	nicInformation := host.Status.HardwareDetails.NIC
	if len(nicInformation) < 1 {
		return "", fmt.Errorf("failed to find nic information in host %s", host.Name)
	}
	var lldpNic baremetalv1alpha1.NIC
	for i := range nicInformation {
		if nicInformation[i].MAC == hostMacAddress {
			lldpNic = nicInformation[i]
			break
		}
	}
	if reflect.ValueOf(lldpNic.LLDP).IsZero() {
		return "", fmt.Errorf("failed to find lldp information in host %s for mac %s", host.Name, hostMacAddress)
	}
	return lldpNic.LLDP.SwitchChassisId, nil
}

func (b BmInstanceBackend) setUserData(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance,
	disks map[string][]baremetalv1alpha1.Storage) ([]byte, error) {

	// Create new CloudConfig object and load incoming UserData
	newCloudConfig, err := util.NewCloudConfig("ubuntu", instance.Spec.UserData)
	if err != nil {
		return nil, fmt.Errorf("setUserData: error while Unmarshaling requestUserData: %w", err)
	}

	// Add metadata: hostname
	intfSpec := instance.Spec.Interfaces[0]
	newCloudConfig.SetHostName(intfSpec.DnsName)

	newCloudConfig.SetDefaultUserGroup(defaultOSUser, util.GetSshPublicKeys(ctx, instance))

	// Add extra network data
	if _, ok := host.Labels[bmenrollment.ClusterGroupID]; ok {
		err = b.setExtraNetworkData(ctx, host, newCloudConfig, instance)
		if err != nil {
			return nil, err
		}
	}
	// setup and configure data disks
	if disks != nil && len(disks["dataDevices"]) >= 1 {
		err = setDataDisks(ctx, disks, newCloudConfig)
		if err != nil {
			return nil, fmt.Errorf("setUserData: failed to setup data disks")
		}
	}

	err = setMachineId(ctx, newCloudConfig, instance.Name)
	if err != nil {
		return nil, fmt.Errorf("setUserData: failed to set new machine id")
	}

	// initialize jupyterlab for quick connect
	if instance.Spec.QuickConnectEnabled == pb.TriState_True.String() && b.Cfg.InstanceOperator.OperatorFeatureFlags.EnableQuickConnectClientCA {
		rootCAPublicCertificateFile, err := os.Open("/vault/secrets/quick-connect-client-ca.pem")
		if err != nil {
			return nil, fmt.Errorf("error occurred file opening root CA public certificate file: %w", err)
		}
		defer rootCAPublicCertificateFile.Close()

		rootCAPublicCertificateFileContentByte, err := io.ReadAll(rootCAPublicCertificateFile)
		if err != nil {
			return nil, fmt.Errorf("error occurred file reading content of root CA public certificate file: %w", err)
		}

		// Install jupyterlab
		err = newCloudConfig.SetJupyterLab(instance.Namespace, instance.Name, string(rootCAPublicCertificateFileContentByte), defaultOSUser,
			b.Cfg.InstanceOperator.QuickConnectHost)
		if err != nil {
			return nil, fmt.Errorf("error occured setting JupyterLab: %w", err)
		}
	}
	err = newCloudConfig.SetStunnelConf(b.Cfg.InstanceOperator.StorageClusterAddr)
	if err != nil {
		return nil, fmt.Errorf("error occured setting stunnel config: %w", err)
	}
	// set write_files to cloud_int
	newCloudConfig.SetWriteFile()

	// set runcmd to cloud_init
	newCloudConfig.SetRunCmd()

	// set packages to cloud_init
	newCloudConfig.SetPackages()

	// Convert CloudConfig object to yaml
	data, err := newCloudConfig.RenderYAML()
	if err != nil {
		return nil, fmt.Errorf("setUserData: error while Marshaling userData: %w", err)
	}

	return data, nil
}

func getHostDisks(ctx context.Context, host *baremetalv1alpha1.BareMetalHost) map[string][]baremetalv1alpha1.Storage {
	disks := make(map[string][]baremetalv1alpha1.Storage)
	rootDevices := make([]baremetalv1alpha1.Storage, 0)
	dataDevices := make([]baremetalv1alpha1.Storage, 0)
	for _, disk := range host.Status.HardwareDetails.Storage {
		if disk.Type == "NVME" {
			if disk.SizeBytes <= baremetalv1alpha1.Capacity(rootDiskMaxSize) {
				rootDevices = append(rootDevices, disk)
			} else {
				dataDevices = append(dataDevices, disk)
			}
		}
	}
	disks["rootDevices"] = rootDevices
	disks["dataDevices"] = dataDevices
	return disks
}

func setRootDisk(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, disks map[string][]baremetalv1alpha1.Storage) error {
	log := log.FromContext(ctx).WithName("setRootDisk")

	rootDisks := disks["rootDevices"]
	log.Info("root disks list", logkeys.RootDisksList, rootDisks)
	// root device list
	if len(rootDisks) >= 1 {
		// sort disks by name
		sort.Slice(rootDisks[:], func(i, j int) bool {
			return rootDisks[i].Name < rootDisks[j].Name
		})
		// root device hint
		var rootDeviceHint baremetalv1alpha1.RootDeviceHints
		rootDeviceHint.WWN = rootDisks[0].WWN

		// update host spec with root device hint
		log.Info("setting root device hint", logkeys.RootDeviceHint, rootDeviceHint)
		host.Spec.RootDeviceHints = &rootDeviceHint
	}
	return nil
}

func setDataDisks(ctx context.Context, disks map[string][]baremetalv1alpha1.Storage, cloudConfig *util.CloudConfig) error {
	log := log.FromContext(ctx).WithName("setDataDisks")

	var disksWwn strings.Builder
	dataDisks := disks["dataDevices"]
	log.Info("data disks list", logkeys.DataDisksList, dataDisks)

	for _, disk := range dataDisks {
		disksWwn.WriteString(fmt.Sprintf("%q ", disk.WWN))
	}

	dataDisksScriptData := fmt.Sprintf(`#!/usr/bin/env bash
set -ex
declare -a wwns=(%s)
count=0
for wwn in "${wwns[@]}"
do
	echo "wwn: ${wwn}"
	partitions_count=$(lsblk -io wwn,name  | grep $wwn | wc -l)
	echo $partitions_count
	# Skip the disk with partitions
	if [[ $partitions_count -gt 1 ]]
	then
		continue
	else
		count=$((count+1))
		echo "getting disk identifier"
		disk_id=$(lsblk -io wwn,name  | grep ${wwn} | awk '{print $2}')
		echo "disk id: $disk_id"
		echo "partitioning disk /dev/${disk_id}"
		sudo parted --script /dev/${disk_id} mklabel gpt mkpart primary ext4 1MiB 100%%
		sleep 2
		echo "setting filesystem"
		sudo mkfs.ext4 /dev/${disk_id}p1
		sudo mkdir -p /scratch-${count}
		sudo mount /dev/${disk_id}p1 /scratch-${count}
		uuid=$(lsblk -io UUID /dev/${disk_id}p1 --noheadings)
		echo "UUID=${uuid}     /scratch-${count}   ext4    rw,user,auto    0    0" | sudo tee -a /etc/fstab
	fi
done
sudo mount -a
exit 0
`, disksWwn.String())

	//write_file with the script
	cloudConfig.AddWriteFileWithPermissions("/etc/configuredatadisks", dataDisksScriptData, "0755")
	// commands to run the script
	cloudConfig.AddRunCmd("cd /etc && ./configuredatadisks && cd ..")
	return nil
}

func setMachineId(ctx context.Context, cloudConfig *util.CloudConfig, instanceId string) error {
	log := log.FromContext(ctx).WithName("setMachineId")

	log.Info("setting machine ID", logkeys.InstanceName, instanceId)
	if instanceId == "" {
		return errors.New("instanceId cannot be empty")
	}
	// ensure unique machine id
	cloudConfig.AddWriteFile("/etc/machine-id", instanceId)
	// copy machine ID as the Weka machine identifier
	cloudConfig.AddRunCmd("mkdir -p /opt/weka/data/agent")
	cloudConfig.AddRunCmd("cp /etc/machine-id /opt/weka/data/agent/machine-identifier")

	setWekaMachineID := `#!/usr/bin/env bash
set -e
	
MACHINE_IDENTIFIER='/etc/machine-id'
WEKA_MACHINE_IDENTIFIER='/opt/weka/data/agent/machine-identifier'

while true
do
	# check if weka machine identifier is present
	if [ ! -f "${WEKA_MACHINE_IDENTIFIER}" ]; then
		mkdir -p '/opt/weka/data/agent'
		cp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}
	fi
	if  ! cmp -s "${MACHINE_IDENTIFIER}" "${WEKA_MACHINE_IDENTIFIER}"; then
		rm -f ${WEKA_MACHINE_IDENTIFIER}
		mkdir -p '/opt/weka/data/agent'
		cp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}
	fi
	sleep 10
done
exit 0
`
	//write_file with the wekaID script
	cloudConfig.AddWriteFileWithPermissions("/etc/set_weka_id", setWekaMachineID, "0700")
	// systemctl service to check and set the weka ID

	wekaIdSystemdService := `[Unit]
Description=Weka machine id check
StartLimitIntervalSec=90
StartLimitBurst=10
	
[Service]
WorkingDirectory=/etc
ExecStart=/bin/bash set_weka_id
Restart=always
RestartSec=10s
	
[Install]
WantedBy=multi-user.target
`
	//write_file with the wekaID script
	cloudConfig.AddWriteFileWithPermissions("/etc/systemd/system/weka-machine-id.service", wekaIdSystemdService, "0600")
	// start systemctl service
	cloudConfig.AddRunCmd("systemctl daemon-reload")
	cloudConfig.AddRunCmd("systemctl  enable weka-machine-id")
	cloudConfig.AddRunCmd("systemctl  start weka-machine-id")
	return nil
}

func updateObject(ctx context.Context, cl client.Client, obj client.Object, opts ...client.UpdateOption) error {
	err := cl.Update(ctx, obj.DeepCopyObject().(client.Object), opts...)
	return err
}

func createObject(ctx context.Context, cl client.Client, obj client.Object, opts ...client.CreateOption) error {
	err := cl.Create(ctx, obj.DeepCopyObject().(client.Object), opts...)
	return err
}

func getSecretValueFromKey(ctx context.Context, data map[string][]byte, key string) (string, error) {
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in data map", key)
	}

	return string(value), nil
}

func getBMCSecret(ctx context.Context, cl client.Client, host *baremetalv1alpha1.BareMetalHost, bmcSecrets *BMCSecrets) error {
	log := log.FromContext(ctx).WithName("getBMCSecret")

	bmcDataSecretName := fmt.Sprintf("%s-bmc-secret", host.Name)
	log.Info("Looking up the BMC secret", logkeys.HostName, host.Name, logkeys.HostNamespace, host.Namespace, logkeys.BmcDataSecretName, bmcDataSecretName)
	secret, err := checkSecretExists(ctx, cl, bmcDataSecretName, host.Namespace)
	if err == nil {
		data, err := getSecretValueFromKey(ctx, secret.Data, "username")
		if err != nil {
			log.Info("Fail to access BMC secret Username", logkeys.Error, err)
		} else {
			bmcSecrets.Username = data
		}

		data, err = getSecretValueFromKey(ctx, secret.Data, "password")
		if err != nil {
			log.Info("Fail to access BMC secret Password", logkeys.Error, err)
		} else {
			bmcSecrets.Password = data
		}

		parsedURL, err := url.Parse(host.Spec.BMC.Address)
		if err != nil {
			log.Info("Fail to parse BMC URL", logkeys.Error, err)
		} else {
			// Split the base URL by the plus sign
			splitScheme := strings.SplitN(parsedURL.Scheme, "+", 2)
			if len(splitScheme) != 2 {
				if strings.Contains(host.Spec.BMC.Address, "ipmi://") {
					bmcSecrets.URL = strings.Replace(host.Spec.BMC.Address, "ipmi", "https", -1)
				} else {
					log.Info("BMC scheme parsing error")
				}
			} else {
				bmcSecrets.URL = fmt.Sprintf("%s://%s", splitScheme[1], parsedURL.Host)

				bmcType := strings.ToLower(splitScheme[0])
				kcs := strings.Contains(bmcType, "denali") || strings.Contains(bmcType, "coyote")
				if kcs {
					bmcSecrets.KCS = bmcType
					fmt.Println("Intel KCS")
				} else {
					log.Info("KCS criteria not meet", logkeys.BmcType, bmcType)
				}
			}
		}
	} else if apierrors.IsNotFound(err) {
		log.Info("BMC secret not found", logkeys.Error, err)
	} else {
		log.Info("BMC secret failed", logkeys.Error, err)
	}

	if bmcSecrets.Username == "" || bmcSecrets.Password == "" || bmcSecrets.URL == "" {
		return fmt.Errorf("getBMCSecret secret missing critical secret data")
	}
	return nil
}

func createSecret(ctx context.Context, cl client.Client, name string,
	namespace string, content map[string][]byte,
) error {
	bootstrapSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: content,
		Type: metal3SecretType,
	}

	secret, err := checkSecretExists(ctx, cl, name, namespace)
	if err == nil {
		// Update the secret with user data.
		secret.ObjectMeta.Labels = bootstrapSecret.ObjectMeta.Labels
		bootstrapSecret.ObjectMeta = secret.ObjectMeta
		return updateObject(ctx, cl, bootstrapSecret)
	} else if apierrors.IsNotFound(err) {
		// Create the secret with user data.
		return createObject(ctx, cl, bootstrapSecret)
	}
	return err
}

func checkSecretExists(ctx context.Context, cl client.Client, name string,
	namespace string,
) (corev1.Secret, error) {
	tmpBootstrapSecret := corev1.Secret{}
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}
	err := cl.Get(ctx, key, &tmpBootstrapSecret)
	return tmpBootstrapSecret, err
}

func deleteSecret(ctx context.Context, cl client.Client, name string,
	namespace string,
) error {
	tmpBootstrapSecret, err := checkSecretExists(ctx, cl, name, namespace)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err == nil {
		// unset the finalizers (remove all since we do not expect anything else
		// to control that object).
		tmpBootstrapSecret.Finalizers = []string{}
		err = updateObject(ctx, cl, &tmpBootstrapSecret)
		if err != nil {
			return err
		}
		// Delete the secret with metadata.
		err = cl.Delete(ctx, &tmpBootstrapSecret)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func calculateCPUCount(instance *cloudv1alpha1.Instance) int {
	cpu := instance.Spec.InstanceTypeSpec.Cpu
	return int(cpu.Sockets * cpu.Cores * cpu.Threads)
}

func (b BmInstanceBackend) getSwitchInformation(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) *SwitchInformation {
	var switchInformation SwitchInformation
	// VLan tag only needed for provisioning.
	// should not stop deleting node
	switchInformation.tenantVlanTag = 0
	switchInformation.storageVlanTag = 0
	switchInformation.acceVlanTag = 0
	switchInformation.bgpCommunityId = 0
	if len(instance.Status.Interfaces) > 0 {
		for i, netInterface := range instance.Status.Interfaces {
			intfStatus := instance.Status.Interfaces[i]
			switch netInterface.Name {
			case util.TenantInterfaceName:
				switchInformation.tenantVlanTag = int64(intfStatus.VlanId)
			case util.StorageInterfaceName:
				switchInformation.storageVlanTag = int64(intfStatus.VlanId)
			case util.AcceleratorClusterInterfaceName:
				switchInformation.acceVlanTag = int64(intfStatus.VlanId)
			case util.BGPClusterInterfaceName:
				switchInformation.bgpCommunityId = int64(intfStatus.VlanId)
			}
		}

	}
	return &switchInformation
}

func (b BmInstanceBackend) verifySSHAccess(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.verifySSHAccess")
	log.Info("Verifying SSH Access", logkeys.InstanceName, instance.Name)

	privateKey, err := os.ReadFile(b.Cfg.SshConfig.PrivateKeyFilePath)
	if err != nil {
		return fmt.Errorf("unable to read private key file %v: %w", b.Cfg.SshConfig.PrivateKeyFilePath, err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}

	hostPublicKeyByte, err := os.ReadFile(b.Cfg.SshConfig.HostPublicKeyFilePath)
	if err != nil {
		return fmt.Errorf("unable to read host public key file %v: %w", b.Cfg.SshConfig.HostPublicKeyFilePath, err)
	}

	hostPublicKey, err := util.GetHostPublicKey(string(hostPublicKeyByte))
	if err != nil {
		return err
	}

	jumpHostConfig := &ssh.ClientConfig{
		User: b.Cfg.SshConfig.SshProxyUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyAlgorithms: util.GetSupportedHostKeyAlgorithms(),
		HostKeyCallback:   ssh.FixedHostKey(hostPublicKey),
	}

	sshProxyAddress := b.Cfg.SshConfig.SshProxyAddress
	if sshProxyAddress == "" {
		sshProxyAddress = instance.Status.SshProxy.ProxyAddress
	}
	sshProxyPort := b.Cfg.SshConfig.SshProxyPort
	if sshProxyPort == 0 {
		sshProxyPort = instance.Status.SshProxy.ProxyPort
	}
	jumpHostAddress := fmt.Sprintf("%s:%d", sshProxyAddress, sshProxyPort)
	jumpHostClient, err := ssh.Dial("tcp", jumpHostAddress, jumpHostConfig)
	if err != nil {
		return err
	}
	defer jumpHostClient.Close()
	session, err := jumpHostClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	targetAddress := instance.Status.Interfaces[0].Addresses[0]
	targetPort := 22
	out, err := session.CombinedOutput(fmt.Sprintf("nc -vz %s %d", targetAddress, targetPort))
	if err != nil {
		return err
	}

	log.Info("Verified SSH Access", logkeys.Output, string(out))

	return nil
}

func (b *BmInstanceBackend) publishEvent(ctx context.Context, event corev1.Event) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.publishEvent")
	log.Info("publishing event", logkeys.Reason, event.Reason, logkeys.Message, event.Message)
	err := b.Create(ctx, &event)
	if err != nil {
		log.Info("failed to record event, ignoring",
			logkeys.Reason, event.Reason, logkeys.Message, event.Message, logkeys.Error, err)
	}
}

func (b BmInstanceBackend) getStorageMacAnnotations(host *baremetalv1alpha1.BareMetalHost) map[string]string {
	storageMacAnnotations := make(map[string]string)
	for annotation, macAddress := range host.Annotations {
		if strings.Contains(annotation, bmenrollment.StorageMACAnnotationPrefix) {
			storageMacAnnotations[annotation] = macAddress
		}
	}
	return storageMacAnnotations
}

func (b *BmInstanceBackend) disableHCIAccess(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.disableHCIAccess").WithValues(logkeys.InstanceName, instance.Name)
	// Update BMC interface
	if err := b.updateBMCInterface(ctx, host); err != nil {
		log.Error(err, "Failed to update BMC interface")
		return ctrl.Result{RequeueAfter: associateRequeueAfterFailure}, nil
	}
	// Disable HostInterface interface
	log.Info("disabling Host Interface")
	err := b.bmcInterface.DisableHCI(ctx)
	if err != nil && !errors.Is(err, bmc.ErrHCINotSupported) {
		log.Error(err, "failed to disable HCI")
		return ctrl.Result{RequeueAfter: associateRequeueAfterFailure}, nil
	} else {
		// Update Condition message
		var message = ""
		if errors.Is(err, bmc.ErrHCINotSupported) {
			message = "host interface not supported"
		} else {
			message = "host interface disabled"
		}
		condition := cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionHCIEnabled,
			Status:             corev1.ConditionFalse,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            message,
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)

		return ctrl.Result{RequeueAfter: requeueAfterHCIChanges}, nil
	}
}

func (b *BmInstanceBackend) disableKCSAccess(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.disableKCSAccess")
	// Setting KCS Policy Control Mode to Secured
	// Update BMC interface
	if err := b.updateBMCInterface(ctx, host); err != nil {
		log.Error(err, "Failed to update BMC interface")
		return ctrl.Result{RequeueAfter: associateRequeueAfterFailure}, nil
	}
	err := b.bmcInterface.DisableKCS(ctx)
	if err != nil && !errors.Is(err, bmc.ErrKCSNotSupported) {
		log.Error(err, "failed to Disable KCS")
		return ctrl.Result{RequeueAfter: associateRequeueAfterFailure}, nil
	} else {
		// set status to KCS Disabled and condition to starting
		var message = ""
		if errors.Is(err, bmc.ErrHCINotSupported) {
			message = "KCS interface not supported"
		} else {
			message = "KCS interface disabled"
		}
		condition := cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionKcsEnabled,
			Status:             corev1.ConditionFalse,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            message,
		}
		util.SetStatusCondition(&instance.Status.Conditions, condition)
		b.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarting)

		return ctrl.Result{RequeueAfter: requeueAfterKcsChanges}, nil
	}
}

func (b *BmInstanceBackend) updateVnetAccess(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, instance *cloudv1alpha1.Instance, hostRef *corev1.ObjectReference) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.updateVnetAccess")
	var condition cloudv1alpha1.InstanceCondition
	if b.Cfg.NetworkConfig.NetworkBackEndType == cloudv1alpha1.NetworkBackEndSDN {
		// fetch information from instance interfaces
		hostSwitchInformation := b.getSwitchInformation(ctx, host, instance)
		if hostSwitchInformation.tenantVlanTag == 0 {
			log.Info("failed to retrieve assigned Vlan from instance interfaces")
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
		}
		description := fmt.Sprintf("host %s, namespace %s", host.Name, host.Namespace)
		// first, check if ports are at desired values
		checkNetworkNodeRequest := sdn.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           host.Name,
			DesiredFrontEndFabricVlan: int64(hostSwitchInformation.tenantVlanTag),
		}
		// check if we have storage
		if hostSwitchInformation.storageVlanTag != 0 {
			checkNetworkNodeRequest.DesiredStorageFabricVlan = hostSwitchInformation.storageVlanTag
		}
		// check if we have acc fabric
		if hostSwitchInformation.acceVlanTag != 0 {
			checkNetworkNodeRequest.DesiredAcceleratorFabricVlan = hostSwitchInformation.acceVlanTag
		}
		if hostSwitchInformation.bgpCommunityId != 0 {
			checkNetworkNodeRequest.DesiredAcceleratorFabricBGPCommunityID = hostSwitchInformation.acceVlanTag
		}
		// check if the requested vlan settings is correct
		checkNetworkNodeResponse, err := b.sdnClient.CheckNetworkNodeStatus(ctx, checkNetworkNodeRequest)
		if err != nil {
			log.Error(err, "check NetworkNode status failed")
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
		}

		if checkNetworkNodeResponse.Status == sdn.UpdateNotStarted {
			updateNetworkNodeRequest := sdn.NetworkNodeConfUpdateRequest{
				NetworkNodeName:    host.Name,
				FrontEndFabricVlan: hostSwitchInformation.tenantVlanTag,
				Description:        description,
			}

			if hostSwitchInformation.acceVlanTag != 0 {
				updateNetworkNodeRequest.AcceleratorFabricVlan = hostSwitchInformation.acceVlanTag
			}

			if hostSwitchInformation.bgpCommunityId != 0 {
				updateNetworkNodeRequest.AcceleratorFabricBGPCommunityID = hostSwitchInformation.bgpCommunityId
			}

			if hostSwitchInformation.storageVlanTag != 0 {
				updateNetworkNodeRequest.StorageFabricVlan = hostSwitchInformation.storageVlanTag
			}

			err = b.sdnClient.UpdateNetworkNodeConfig(ctx, updateNetworkNodeRequest)
			if err != nil {
				log.Error(err, "sdnClient failed to update NetworkNode")
				return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfFailure}, nil
			}

			// after successfully UpdateVlan, we need to requeue the event, and we requeue immediately by setting `Requeue: true`
			// requeueing allows the reconciler to work on other events while we are waiting for the vlan update to be done.
			// note: UpdateNetworkNodeConfig() only update the Network cluster's NetworkNode CR. The change to the NetworkNode CR will trigger the
			// SDN-Controller's reconciliation, which makes vlan change on the switch.
			log.Info("finished updating SDN NetworkNode CR", logkeys.NetworkBackendType, updateNetworkNodeRequest)
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfChanges}, nil
		} else if checkNetworkNodeResponse.Status == sdn.UpdateInProgress {
			// vlan update is still in progress, requeue the event.
			log.Info("NetworkNode vlan update in progress, requeueing...", logkeys.NetworkNodeName, host.Name)
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfterNetworkConfChanges}, nil
		} else if checkNetworkNodeResponse.Status == sdn.UpdateCompleted {
			// vlan update is done, move forward to the next step to update the condition.
			log.Info("NetworkNode vlan update completed", logkeys.NetworkNodeName, host.Name)
		}
		// Update events
		for _, intfStatus := range instance.Status.Interfaces {
			sdnVariable := "Vlan ID"
			if intfStatus.Name == util.BGPClusterInterfaceName {
				sdnVariable = "Community ID"
			}
			event := instance.NewEvent("network connected to Vnet", fmt.Sprintf("%s %d, interface Name %s", sdnVariable, intfStatus.VlanId, intfStatus.Name), hostRef)
			b.publishEvent(ctx, event)
		}
		// Set condition
		// set status to agent connected
		condition = cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionAgentConnected,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}
	} else {
		log.Info("Network backend is None or not supported", logkeys.NetworkBackendType, b.Cfg.NetworkConfig.NetworkBackEndType)
		event := instance.NewEvent("Network backend is None or not supported", fmt.Sprintf("skipping managing vnet. NetworkBackEndType is set to %s", b.Cfg.NetworkConfig.NetworkBackEndType), hostRef)
		b.publishEvent(ctx, event)
		// In case NetworkBackEndType, make sure it is reported in the message of the agentconnected status
		condition = cloudv1alpha1.InstanceCondition{
			Type:               cloudv1alpha1.InstanceConditionAgentConnected,
			Status:             corev1.ConditionTrue,
			Message:            fmt.Sprintf("skipping vnet setup. NetworkBackEndType is set to %s", b.Cfg.NetworkConfig.NetworkBackEndType),
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}
	}
	util.SetStatusCondition(&instance.Status.Conditions, condition)
	return ctrl.Result{RequeueAfter: sshAccessRequeueAfter}, nil
}

func (b *BmInstanceBackend) updateBMCInterface(ctx context.Context, host *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("BmInstanceBackend.updateBMCInterface")
	log.Info("updating BMC credentials")
	err := getBMCSecret(ctx, b.Client, host, &b.bmcSecrets)
	if err != nil {
		return fmt.Errorf("failed to get bmc-secret data secret %w", err)
	}
	if b.bmcInterface == nil {
		b.bmcInterface, err = bmc.New(
			&mygofish.MyGoFishManager{},
			&bmc.Config{
				URL:      b.bmcSecrets.URL,
				Username: b.bmcSecrets.Username,
				Password: b.bmcSecrets.Password,
			})
		if err != nil {
			return fmt.Errorf("failed to get and update bmc %w", err)
		}
	}
	return nil
}
