// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	schedUtil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"sigs.k8s.io/cluster-api/util/patch"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme                         *runtime.Scheme
	VNetPrivateClient              pb.VNetPrivateServiceClient
	VNetClient                     pb.VNetServiceClient
	InstanceBackend                InstanceBackend
	SshProxyTunnelClient           client.Client
	EnableMeteringMonitorFinalizer bool
}

const (
	InstanceFinalizer                    = "private.cloud.intel.com/instancefinalizer"
	InstanceMeteringMonitorFinalizer     = "private.cloud.intel.com/instancemeteringmonitorfinalizer"
	addFinalizers                        = "add_finalizers"
	removeFinalizers                     = "remove_finalizers"
	requeueAfterNetworkAllocationFailure = 10 * time.Second
)

type FinalizerChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating finalizers change.
func (FinalizerChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	return !slices.Equal(e.ObjectNew.GetFinalizers(), e.ObjectOld.GetFinalizers())
}

type InstanceBackend interface {
	// Create/Update a VM/BM and updates the status of the Instance.
	CreateOrUpdateInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error)
	// Deletes the VM/BM and its associated resources like volumes, secrets.
	DeleteResources(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error)
	// Add watches to controller.
	BuildController(ctx context.Context, ctrlBuilder *builder.Builder) *builder.Builder
}

func NewInstanceReconciler(ctx context.Context, mgr ctrl.Manager, vNetPrivateClient pb.VNetPrivateServiceClient, vNetClient pb.VNetServiceClient, instanceBackend InstanceBackend, cfg *cloudv1alpha1.InstanceOperatorConfig) (*InstanceReconciler, error) {
	sshProxyTunnelClusterConfig, err := util.CreateSshProxyTunnelClusterConfig(ctx, mgr, cfg)
	if err != nil {
		return nil, fmt.Errorf("NewInstanceReconciler: %w", err)
	}
	// Configure connection to SSH Proxy Operator cluster (cache, client).
	sshProxyTunnelCluster, err := cluster.New(sshProxyTunnelClusterConfig, func(o *cluster.Options) {
		o.Scheme = mgr.GetScheme()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get sshProxyTunnelCluster: %w", err)
	}
	if err := mgr.Add(sshProxyTunnelCluster); err != nil {
		return nil, fmt.Errorf("unable to add sshProxyTunnelCluster to manager: %w", err)
	}

	// Create reconciler.
	r := &InstanceReconciler{
		Client:                         mgr.GetClient(),
		Scheme:                         mgr.GetScheme(),
		VNetPrivateClient:              vNetPrivateClient,
		VNetClient:                     vNetClient,
		SshProxyTunnelClient:           sshProxyTunnelCluster.GetClient(),
		InstanceBackend:                instanceBackend,
		EnableMeteringMonitorFinalizer: cfg.EnableMeteringMonitorFinalizer,
	}

	// Create controller.
	builder := ctrl.NewControllerManagedBy(mgr).
		Named("instance_controller").
		For(
			&cloudv1alpha1.Instance{},
			builder.WithPredicates(
				predicate.Or(
					// Reconcile Instance if Instance spec changes.
					predicate.GenerationChangedPredicate{},
					// Reconcile Instance if finalizers change.
					FinalizerChangedPredicate{},
				),
			),
		).
		// Reconcile Instance when sshProxyTunnel with same name changes.
		WatchesRawSource(
			source.Kind(sshProxyTunnelCluster.GetCache(), &cloudv1alpha1.SshProxyTunnel{}),
			&handler.EnqueueRequestForObject{},
		)
	builder = instanceBackend.BuildController(ctx, builder)
	err = builder.Complete(r)
	if err != nil {
		return nil, fmt.Errorf("unable to create instance controller: %w", err)
	}
	return r, nil
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ipaddresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=sshproxytunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=sshproxytunnels/status,verbs=get
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InstanceReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("BEGIN")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the source instance.
		instance, err := r.getInstance(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if instance == nil {
			log.Info("Ignoring reconcile request because source instance was not found in cache")
			return ctrl.Result{}, nil
		}
		result, processErr := func() (ctrl.Result, error) {
			if instance.ObjectMeta.DeletionTimestamp.IsZero() {
				if err := r.initializeMetadataAndPersist(ctx, req, instance); err != nil {
					return ctrl.Result{}, err
				}
				return r.processInstance(ctx, instance)
			} else {
				return r.processDeleteInstance(ctx, instance, req)
			}
		}()
		if err := r.updateStatus(ctx, instance, processErr); err != nil {
			processErr = multierror.Append(processErr, err)
		}
		if err = r.PersistStatusUpdate(ctx, instance, req, processErr); err != nil {
			processErr = multierror.Append(processErr, err)
		}
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "InstanceReconciler.Reconcile: error reconciling Instance")
	}
	log.Info("END", logkeys.Result, result, logkeys.Error, reconcileErr)
	return result, reconcileErr
}

// Update accepted condition, phase, and message.
func (r *InstanceReconciler) updateStatus(ctx context.Context, instance *cloudv1alpha1.Instance, reconcileErr error) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.updateStatus")
	log.Info("BEGIN")

	// Add missing conditions.
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionAccepted, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionRunning, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionStartupComplete, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionAgentConnected, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionSshProxyReady, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionStopped, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionStopping, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionStarting, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionStarted, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	util.SetStatusConditionIfMissing(instance, cloudv1alpha1.InstanceConditionVerifiedSshAccess, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")

	// Update Accepted condition.
	var condStatus k8sv1.ConditionStatus
	var reason cloudv1alpha1.ConditionReason
	var message string
	if reconcileErr == nil {
		// Acceptance has completed successfully.
		condStatus = k8sv1.ConditionTrue
		reason = cloudv1alpha1.ConditionReasonAccepted
		message = cloudv1alpha1.InstanceMessageProvisioningAccepted
	} else {
		condStatus = k8sv1.ConditionFalse
		reason = cloudv1alpha1.ConditionReasonNotAccepted
		message = reconcileErr.Error()
	}
	if err := r.updateStatusCondition(ctx, instance, cloudv1alpha1.InstanceConditionAccepted, condStatus, reason, message); err != nil {
		return err
	}
	if err := r.updateStatusPhaseAndMessage(ctx, instance); err != nil {
		return fmt.Errorf("updateStatus: %w", err)
	}

	instance.Status.UserName = instance.Spec.MachineImageSpec.UserName
	log.Info("Calculated instance status", logkeys.StatusConditions, condStatus, logkeys.StatusPhase, instance.Status.Phase, logkeys.StatusMessage, instance.Status.Message)

	log.Info("END")
	return nil
}

// Update instance status.
func (r *InstanceReconciler) PersistStatusUpdate(ctx context.Context, instance *cloudv1alpha1.Instance, req ctrl.Request, reconcileErr error) error {
	logger := log.FromContext(ctx).WithName("InstanceReconciler.PersistStatusUpdate")
	logger.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestInstance, err := r.getInstance(ctx, req)
		// instance is deleted
		if latestInstance == nil {
			logger.Info("instance not found", logkeys.CloudAccountId, CloudAccountId(instance), logkeys.InstanceName, instance.Spec.InstanceName)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the instance: %+v. error:%w", instance, err)
		}
		if !equality.Semantic.DeepEqual(instance.Status, latestInstance.Status) {
			logger.Info("instance status mismatch", logkeys.InstanceStatus, instance.Status, "latestInstanceStatus", latestInstance.Status)
			// update latest instance status
			instance.Status.DeepCopyInto(&latestInstance.Status)
			if err := r.Status().Update(ctx, latestInstance); err != nil {
				return fmt.Errorf("PersistStatusUpdate: %w", err)
			}
		} else {
			logger.Info("instance status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update instance status: %w", err)
	}
	logger.Info("END")
	return nil
}

// Set status phase and message based on conditions.
// The message will come from the condition that is most relevant.
func (r *InstanceReconciler) updateStatusPhaseAndMessage(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.updateStatusPhaseAndMessage")
	log.Info("BEGIN")

	acceptedCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionAccepted)
	hasAcceptedCondition := acceptedCond != nil
	accepted := acceptedCond != nil && acceptedCond.Status == k8sv1.ConditionTrue
	runningCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionRunning)
	running := runningCond != nil && runningCond.Status == k8sv1.ConditionTrue
	failedCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionFailed)
	instanceCondfailed := failedCond != nil && failedCond.Status == k8sv1.ConditionTrue
	startingCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStarting)
	starting := startingCond != nil && startingCond.Status == k8sv1.ConditionTrue
	startedCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStarted)
	started := startedCond != nil && startedCond.Status == k8sv1.ConditionTrue
	startupCompleteCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
	startupComplete := startupCompleteCond != nil && startupCompleteCond.Status == k8sv1.ConditionTrue
	stoppingCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStopping)
	stopping := stoppingCond != nil && stoppingCond.Status == k8sv1.ConditionTrue
	stoppedCond := util.FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStopped)
	stopped := stoppedCond != nil && stoppedCond.Status == k8sv1.ConditionTrue
	terminating := !instance.ObjectMeta.DeletionTimestamp.IsZero()

	var phase cloudv1alpha1.InstancePhase
	var message string

	if terminating {
		// The instance and its associated resources are in the process of being deleted
		phase = cloudv1alpha1.PhaseTerminating
		message = cloudv1alpha1.InstanceMessageTerminating
	} else if instanceCondfailed {
		// The instance crashed, failed, or is otherwise unavailable
		phase = cloudv1alpha1.PhaseFailed
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageFailed, failedCond.Message)
	} else if stopping {
		phase = cloudv1alpha1.PhaseStopping
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageStopping, stoppingCond.Message)
	} else if stopped {
		phase = cloudv1alpha1.PhaseStopped
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageStopped, stoppedCond.Message)
	} else if starting {
		phase = cloudv1alpha1.PhaseStarting
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageStarting, startingCond.Message)
	} else if started {
		phase = cloudv1alpha1.PhaseReady
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageStarted, startedCond.Message)
	} else if startupComplete {
		phase = cloudv1alpha1.PhaseReady
		message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageStartupComplete, startupCompleteCond.Message)
	} else {
		// not ready
		phase = cloudv1alpha1.PhaseProvisioning
		if running {
			message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageRunning, startupCompleteCond.Message)
		} else {
			// not running
			if hasAcceptedCondition {
				if accepted {
					message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageProvisioningAccepted, runningCond.Message)
				} else {
					// not accepted
					message = fmt.Sprintf("%s. %v", cloudv1alpha1.InstanceMessageProvisioningNotAccepted, acceptedCond.Message)
				}
			}
		}
	}

	instance.Status.Phase = phase
	instance.Status.Message = message

	log.Info("updated instance status phase and message", logkeys.StatusPhase, instance.Status.Phase, logkeys.StatusMessage, instance.Status.Message, logkeys.StatusConditions, instance.Status.Conditions)
	log.Info("END")
	return nil
}

func (r *InstanceReconciler) processInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	logger := log.FromContext(ctx).WithName("InstanceReconciler.processInstance")
	if err := r.validateCreateRequest(instance); err != nil {
		return ctrl.Result{}, err
	}
	if networErr := r.setNetworkConfiguration(ctx, instance); networErr != nil {
		timeOfexpiredReservation := instance.CreationTimestamp.Add(schedUtil.DurationToExpireAssumedPod)
		if timeOfexpiredReservation.Before(time.Now()) {
			// if this network IP allocation fails:
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionFailed,
				Status:             k8sv1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            fmt.Sprintf("%v", networErr),
			}
			// Mark the instance as failed
			util.SetStatusCondition(&instance.Status.Conditions, condition)
			// Delete the reference to cluster Id and node ID only if this is a bmaas instance
			if instance.Spec.InstanceTypeSpec.InstanceCategory == cloudv1alpha1.InstanceCategoryBareMetalHost {
				instanceHelper, err := patch.NewHelper(instance, r.Client)
				if err != nil {
					logger.Error(err, "failed to create instance helper, retrying")
					return ctrl.Result{RequeueAfter: requeueAfterNetworkAllocationFailure}, nil
				}
				delete(instance.Labels, "node-id")
				delete(instance.Labels, "cluster-id")
				err = instanceHelper.Patch(ctx, instance)
				if err != nil {
					logger.Error(err, "failed to remove associated host from instance labels, retrying")
					return ctrl.Result{RequeueAfter: requeueAfterNetworkAllocationFailure}, nil
				}
			}
			if err := r.unallocateIpAddresses(ctx, instance); err != nil {
				logger.Error(err, "failed to unallocate IP adresses")
				return ctrl.Result{RequeueAfter: requeueAfterNetworkAllocationFailure}, nil
			}
			logger.Error(networErr, "timed out while trying to set network configuration, instance is marked failed")
			return ctrl.Result{}, nil
		}
		logger.Error(networErr, "failed to setup network configuration, retrying")
		return ctrl.Result{RequeueAfter: requeueAfterNetworkAllocationFailure}, nil
	}
	if result, err := r.InstanceBackend.CreateOrUpdateInstance(ctx, instance); err != nil || !result.IsZero() {
		return result, err
	}
	if result, err := r.createOrUpdateSshProxyTunnel(ctx, instance); err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

// Get instance from K8s.
// Returns (nil, nil) if not found.
func (r *InstanceReconciler) getInstance(ctx context.Context, req ctrl.Request) (*cloudv1alpha1.Instance, error) {
	instance := &cloudv1alpha1.Instance{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if errors.IsNotFound(err) || reflect.ValueOf(instance).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getInstance: %w", err)
	}
	return instance, nil
}

func (r *InstanceReconciler) validateCreateRequest(instance *cloudv1alpha1.Instance) error {
	if len(instance.Spec.Interfaces) == 0 {
		return fmt.Errorf("validateCreateRequest: no interfaces")
	}
	if len(instance.Spec.Interfaces) == 0 {
		return fmt.Errorf("validateCreateRequest: empty Spec.Interfaces")
	}
	return nil
}

// Add/remove finalizers and persist
func (r *InstanceReconciler) UpdateFinalizers(ctx context.Context, instance *cloudv1alpha1.Instance, req ctrl.Request, op string, finalizers []string) error {
	logger := log.FromContext(ctx).WithName("InstanceReconciler.UpdateFinalizers")
	logger.Info("BEGIN", logkeys.Task, op)
	defer logger.Info("END", logkeys.Task, op)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestInstance, err := r.getInstance(ctx, req)
		if latestInstance == nil {
			logger.Info("instance not found", logkeys.CloudAccountId, CloudAccountId(instance), logkeys.InstanceName, instance.Spec.InstanceName)
			return nil
		}
		if err != nil {
			logger.Info("instance not found", logkeys.CloudAccountId, CloudAccountId(instance), logkeys.InstanceName, instance.Spec.InstanceName)
			return err
		}
		for _, finalizer := range finalizers {
			if op == addFinalizers {
				controllerutil.AddFinalizer(latestInstance, finalizer)
			} else {
				controllerutil.RemoveFinalizer(latestInstance, finalizer)
			}
		}
		if !reflect.DeepEqual(instance.GetFinalizers(), latestInstance.GetFinalizers()) {
			logger.Info("instance finalizer mismatches", logkeys.CurrentInstanceFinalizers, instance.GetFinalizers(),
				logkeys.LatestInstanceFinalizers, latestInstance.GetFinalizers())

			if err := r.Update(ctx, latestInstance); err != nil {
				return fmt.Errorf("UpdateFinalizers: update failed: %w", err)
			}
		} else {
			logger.Info("instance finalizer doesn't need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update instance finalizers: %w", err)
	}
	return nil
}

// Process delete instance request
func (r *InstanceReconciler) processDeleteInstance(ctx context.Context, instance *cloudv1alpha1.Instance, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("InstanceReconciler.processDeleteInstance")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := r.deleteSshProxyTunnel(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("processDeleteInstance: %w", err)
	}

	// Attempt to delete instance resources.
	if result, err := r.InstanceBackend.DeleteResources(ctx, instance); err != nil || !result.IsZero() {
		return result, err
	}

	// Verify if a LoadBalancer finalizer exists, if it does, reconcile the object. Once the LB operator removes
	// the Instance from the Load Balancer pool members and the Finalizer from the Instance, the update will
	// trigger another reconcile loop.
	if controllerutil.ContainsFinalizer(instance, loadbalancer.LoadbalancerFinalizer) {
		log.Info("requeuing instance termination due to presence of LoadBalancer finalizer")
		return ctrl.Result{}, nil
	}

	if err := r.unallocateIpAddresses(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("processDeleteInstance: %w", err)
	}

	// remove our finalizer from the list and update it.
	log.Info("All resources deleted. Removing finalizer.")
	finalizers := []string{InstanceFinalizer}
	err := r.UpdateFinalizers(ctx, instance, req, removeFinalizers, finalizers)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Stop reconciliation as the item is good for deletion
	return ctrl.Result{}, nil
}

// Persists the Instance finalizers update.
func (r *InstanceReconciler) initializeMetadataAndPersist(ctx context.Context, req ctrl.Request, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.initializeMetadataAndPersist")
	log.Info("BEGIN")
	var finalizers []string
	// Add Finalizer (if not present) for the deletion cleanup.
	// This only updates the in-memory Instance.
	finalizers = append(finalizers, InstanceFinalizer)

	// Add Instance metering monitor Finalizer which can be removed by compute metering monitor only
	if r.EnableMeteringMonitorFinalizer {
		finalizers = append(finalizers, InstanceMeteringMonitorFinalizer)
	}

	err := r.UpdateFinalizers(ctx, instance, req, addFinalizers, finalizers)
	if err != nil {
		return err
	}
	log.Info("END")
	return nil
}

func (r *InstanceReconciler) AddressConsumerId(fqdn string) string {
	return fqdn
}

func CloudAccountId(instance *cloudv1alpha1.Instance) string {
	return instance.ObjectMeta.Namespace
}

func (r *InstanceReconciler) allocateSubnets(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.allocateSubnets")
	for i := range instance.Spec.Interfaces {
		spec := &instance.Spec.Interfaces[i]
		out := &instance.Status.Interfaces[i]
		vlanDomain := ""
		addressSpace := ""
		if spec.Name == util.AcceleratorClusterInterfaceName {
			vlanDomain = instance.Spec.ClusterGroupId
			addressSpace = instance.Spec.ClusterGroupId
		} else if spec.Name == util.BGPClusterInterfaceName {
			vlanDomain = util.XBXAddressSpace
			addressSpace = util.XBXAddressSpace
		} else if spec.Name == util.StorageInterfaceName {
			addressSpace = util.StorageAddressSpace
		}
		if out.Subnet == "" {
			// Reserve subnet and get subnet parameters
			subnetResp, err := r.VNetPrivateClient.ReserveSubnet(ctx, &pb.VNetReserveSubnetRequest{
				VNetReference: &pb.VNetReference{
					CloudAccountId: CloudAccountId(instance),
					Name:           spec.VNet,
				},
				VlanDomain:          vlanDomain,
				AddressSpace:        addressSpace,
				MaximumPrefixLength: utils.GetMaximumPrefixLength(instance.Spec.InstanceGroupSize),
			})
			if err != nil {
				return fmt.Errorf("IPAM: ReserveSubnet for %s: %w", spec.Name, err)
			}
			out.Subnet = subnetResp.Spec.Subnet
			out.Gateway = subnetResp.Spec.Gateway
			out.PrefixLength = int(subnetResp.Spec.PrefixLength)
			out.VlanId = int(subnetResp.Spec.VlanId)
			log.Info("IPAM: Reserved subnet", logkeys.SubnetResp, subnetResp)
		}
	}
	return nil
}

func (r *InstanceReconciler) allocateIpAddresses(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.allocateIpAddresses")
	for i := range instance.Spec.Interfaces {
		spec := &instance.Spec.Interfaces[i]
		out := &instance.Status.Interfaces[i]

		if len(out.Addresses) == 0 || out.Addresses[0] == "" {
			// Reserve address.
			addressConsumerId := r.AddressConsumerId(spec.DnsName)
			addressResp, err := r.VNetPrivateClient.ReserveAddress(ctx, &pb.VNetReserveAddressRequest{
				VNetReference: &pb.VNetReference{
					CloudAccountId: CloudAccountId(instance),
					Name:           spec.VNet,
				},
				AddressReference: &pb.VNetAddressReference{
					AddressConsumerId: addressConsumerId,
				},
			})
			if err != nil {
				return fmt.Errorf("allocateIpAddresses: ReserveAddress for %s: %w", spec.Name, err)
			}
			out.Addresses = []string{addressResp.Address}
			log.Info("IPAM: Reserved address", logkeys.AddressConsumerId, addressConsumerId, logkeys.Address, addressResp.Address)
		}
	}
	return nil
}

// To ensure that allocated IP addresses are unallocated when the instance is deleted, unallocateIpAddresses
// must be able to unallocate IP addresses without knowing the IP addresses.
func (r *InstanceReconciler) unallocateIpAddresses(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.unallocateIpAddresses")
	for i := range instance.Spec.Interfaces {
		spec := &instance.Spec.Interfaces[i]

		// Release address.
		addressConsumerId := r.AddressConsumerId(spec.DnsName)
		_, err := r.VNetPrivateClient.ReleaseAddress(ctx, &pb.VNetReleaseAddressRequest{
			VNetReference: &pb.VNetReference{
				CloudAccountId: CloudAccountId(instance),
				Name:           spec.VNet,
			},
			AddressReference: &pb.VNetAddressReference{
				AddressConsumerId: addressConsumerId,
			},
		})
		if status.Code(err) == codes.NotFound {
			log.Info("IPAM: Address already released", logkeys.AddressConsumerId, addressConsumerId)
		} else if err != nil {
			return fmt.Errorf("unallocateIpAddresses: ReleaseAddress: %w", err)
		} else {
			log.Info("IPAM: Released address", logkeys.AddressConsumerId, addressConsumerId)
		}
		// Release subnet (if unused).
		_, err = r.VNetPrivateClient.ReleaseSubnet(ctx, &pb.VNetReleaseSubnetRequest{
			VNetReference: &pb.VNetReference{
				CloudAccountId: CloudAccountId(instance),
				Name:           spec.VNet,
			},
		})
		if status.Code(err) == codes.NotFound {
			log.Info("IPAM: Subnet already released")
		} else if status.Code(err) == codes.FailedPrecondition {
			log.Info("IPAM: Subnet could not be released because it is in use")
		} else if err != nil {
			return fmt.Errorf("unallocateIpAddresses: ReleaseSubnet: %w", err)
		} else {
			log.Info("IPAM: Released subnet")
		}
		// Delete accelerator vNet (if unused).
		// Accelerator cluster vNet is deleted once all the consumed addresses are released.
		// When an instance group is deleted, it makes sure the accelerator vNet is deleted only,
		// after deleting the last instance from group
		if spec.Name == util.AcceleratorClusterInterfaceName {
			err = r.deleteAcceleratorClustervNet(ctx, instance)
			if err != nil {
				return err
			}
		} else if spec.Name == util.BGPClusterInterfaceName {
			err = r.deleteBGPClustervNet(ctx, instance)
			if err != nil {
				return err
			}
		} else if spec.Name == util.StorageInterfaceName {
			err = r.deleteStoragevNet(ctx, instance)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *InstanceReconciler) deleteAcceleratorClustervNet(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	// accelerator network vNet name
	acceleratorvNetName := utils.GenerateAcceleratorVNetName(instance.Spec.InstanceName, instance.Spec.InstanceGroup, instance.Spec.ClusterGroupId)
	return r.deletevNet(ctx, instance, acceleratorvNetName)
}

func (r *InstanceReconciler) deleteBGPClustervNet(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	acceleratorvNetName := utils.GenerateBGPClusterVNetName(instance.Spec.SuperComputeGroupId, instance.Spec.InstanceGroup)
	return r.deletevNet(ctx, instance, acceleratorvNetName)
}

func (r *InstanceReconciler) deleteStoragevNet(ctx context.Context, instance *cloudv1alpha1.Instance) error {

	// storage vNet name
	storagevNetName := fmt.Sprintf("%s-storage", instance.Spec.AvailabilityZone)
	return r.deletevNet(ctx, instance, storagevNetName)
}

func (r *InstanceReconciler) deletevNet(ctx context.Context, instance *cloudv1alpha1.Instance, vNetName string) error {
	// delete vNet
	log := log.FromContext(ctx).WithName("InstanceReconciler.deletevNet").WithValues(logkeys.Instance, instance)
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("deleting vNet", logkeys.VNetName, vNetName)
	_, err := r.VNetClient.Delete(ctx, &pb.VNetDeleteRequest{Metadata: &pb.VNetDeleteRequest_Metadata{CloudAccountId: CloudAccountId(instance),
		NameOrId: &pb.VNetDeleteRequest_Metadata_Name{Name: vNetName}}})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Info("deletevNet: vNet not found", logkeys.VNetName, vNetName)
		} else if status.Code(err) == codes.FailedPrecondition {
			log.Info("deletevNet: vNet could not be deleted because subnet associated with it is in use", logkeys.VNetName, vNetName)
		} else {
			return fmt.Errorf("deletevNet: failed to delete vNet: %s. error: %w", vNetName, err)
		}
	} else {
		log.Info("deleted vNet", logkeys.VNetName, vNetName)
	}

	return nil
}

func (r *InstanceReconciler) setNetworkConfiguration(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("InstanceReconciler.setNetworkConfiguration")
	log.Info("BEGIN")

	numIntf := len(instance.Spec.Interfaces)

	// Extend Status.Interfaces if needed.
	for i := range instance.Spec.Interfaces {
		if len(instance.Status.Interfaces) <= i {
			instance.Status.Interfaces = append(instance.Status.Interfaces, cloudv1alpha1.InstanceInterfaceStatus{
				Name:    instance.Spec.Interfaces[i].Name,
				VNet:    instance.Spec.Interfaces[i].VNet,
				DnsName: instance.Spec.Interfaces[i].DnsName,
			})
		}
	}

	// Truncate Status.Interfaces if needed.
	instance.Status.Interfaces = instance.Status.Interfaces[:numIntf]

	if err := r.allocateSubnets(ctx, instance); err != nil {
		return fmt.Errorf("SetNetworkConfiguration: %w", err)
	}
	if err := r.allocateIpAddresses(ctx, instance); err != nil {
		return fmt.Errorf("SetNetworkConfiguration: %w", err)
	}

	// Check that we now have required fields.
	if len(instance.Status.Interfaces) == 0 {
		return fmt.Errorf("SetNetworkConfiguration: empty Status.Interfaces")
	}
	intfStatus := instance.Status.Interfaces[0]
	if len(intfStatus.Addresses) == 0 {
		return fmt.Errorf("SetNetworkConfiguration: empty Status.Interfaces[0].Addresses")
	}
	address := intfStatus.Addresses[0]
	if address == "" {
		return fmt.Errorf("SetNetworkConfiguration: empty Status.Interfaces[0].Addresses[0]")
	}
	log.Info("END", logkeys.InstanceSpec, instance.Spec.Interfaces, logkeys.InstanceStatus, instance.Status.Interfaces)
	return nil
}

// Update a status condition.
func (r *InstanceReconciler) updateStatusCondition(ctx context.Context, instance *cloudv1alpha1.Instance,
	conditionType cloudv1alpha1.InstanceConditionType,
	status k8sv1.ConditionStatus, reason cloudv1alpha1.ConditionReason, message string,
) error {
	logger := log.FromContext(ctx).WithName("InstanceReconciler.updateStatusCondition")
	logger.Info("BEGIN", logkeys.InstanceConditionType, conditionType, logkeys.Message, message)
	instanceCondition := cloudv1alpha1.InstanceCondition{
		Status:             status,
		Message:            message,
		Type:               conditionType,
		LastTransitionTime: metav1.Now(),
		LastProbeTime:      metav1.Now(),
		Reason:             reason,
	}
	util.SetStatusCondition(&instance.Status.Conditions, instanceCondition)
	logger.Info("END", logkeys.Instance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(instance)))
	return nil
}

func (r *InstanceReconciler) createOrUpdateSshProxyTunnel(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InstanceReconciler.createOrUpdateSshProxyTunnel").Start()
	defer span.End()
	log.Info("BEGIN")

	if len(instance.Status.Interfaces) == 0 || len(instance.Status.Interfaces[0].Addresses) == 0 {
		return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: Instance does not have an IP address; instance.Status.Interfaces=%v", instance.Status.Interfaces)
	}

	sshProxyTunnelInstance := &cloudv1alpha1.SshProxyTunnel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name,
			Namespace: instance.ObjectMeta.Namespace,
		},
	}

	err := r.SshProxyTunnelClient.Get(ctx, types.NamespacedName{
		Name: instance.ObjectMeta.Name, Namespace: instance.ObjectMeta.Namespace}, sshProxyTunnelInstance)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: failed to get sshproxytunnel: %w", err)
	}

	sshPublicKeys := util.GetSshPublicKeys(ctx, instance)

	sshProxyTunnelInstance.Spec = cloudv1alpha1.SshProxyTunnelSpec{
		// Only allow SSH to the first IP address of the instance
		TargetAddresses: []string{instance.Status.Interfaces[0].Addresses[0]},
		SshPublicKeys:   sshPublicKeys,
		TargetPorts:     []int{22},
	}

	// Update Instance.Status.SshProxy.
	// These will be blank if SshProxyTunnel does not exist.
	instance.Status.SshProxy.ProxyAddress = sshProxyTunnelInstance.Status.ProxyAddress
	instance.Status.SshProxy.ProxyPort = sshProxyTunnelInstance.Status.ProxyPort
	instance.Status.SshProxy.ProxyUser = sshProxyTunnelInstance.Status.ProxyUser
	log.Info("Copied status from SshProxyTunnel", logkeys.SshProxyTunnelStatus, instance.Status.SshProxy)

	// Set InstanceConditionSshProxyReady condition.
	var condStatus k8sv1.ConditionStatus
	reason := cloudv1alpha1.ConditionReasonNone
	message := ""
	if instance.Status.SshProxy.ProxyAddress != "" && instance.Status.SshProxy.ProxyPort != 0 && instance.Status.SshProxy.ProxyUser != "" {
		condStatus = k8sv1.ConditionTrue
	} else {
		condStatus = k8sv1.ConditionFalse
	}
	if err := r.updateStatusCondition(ctx, instance, cloudv1alpha1.InstanceConditionSshProxyReady, condStatus, reason, message); err != nil {
		return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: %w", err)
	}

	if errors.IsNotFound(err) {
		// Create
		log.Info("SshProxyTunnel object not found. So creating new SshProxyTunnel")
		if err := util.CreateNamespaceIfNeeded(ctx, instance.ObjectMeta.Namespace, r.SshProxyTunnelClient); err != nil {
			return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: failed to create namespace for sshproxytunnel: %w", err)
		}
		if err := r.SshProxyTunnelClient.Create(ctx, sshProxyTunnelInstance); err != nil {
			return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: failed to create sshproxytunnel: %w", err)
		}
	} else {
		// Update
		log.Info("SshProxyTunnel object already exist. So updating it", logkeys.SshProxyTunnelInstance, sshProxyTunnelInstance)
		if err := r.SshProxyTunnelClient.Update(ctx, sshProxyTunnelInstance); err != nil {
			return ctrl.Result{}, fmt.Errorf("createOrUpdateSshProxyTunnel: failed to update sshproxytunnel: %w", err)
		}
	}

	log.Info("END", logkeys.SshProxyTunnelInstance, sshProxyTunnelInstance)
	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) deleteSshProxyTunnel(ctx context.Context, instance *cloudv1alpha1.Instance) error {
	sshProxyTunnel := &cloudv1alpha1.SshProxyTunnel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	if err := r.SshProxyTunnelClient.Delete(ctx, sshProxyTunnel); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("deleteSshProxyTunnel: %w", err)
		}
	}
	return nil
}
