// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package replicator

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/convert"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/atomicduration"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Instance Replicator reconciles instance objects from the Compute API Server to K8s objects of kind Instance.private.cloud.intel.com.
// It copies instance Status and deletion confirmation (removal of finalizer) from K8s to the Compute API Server.
// See https://docs.bitnami.com/tutorials/kubewatch-an-example-of-kubernetes-custom-controller/
// See https://komodor.com/learn/controller-manager/
type Replicator struct {
	informer                 toolscache.SharedIndexInformer
	cache                    cache.Cache
	k8sClient                k8sclient.Client
	instanceClient           pb.InstancePrivateServiceClient
	converter                *convert.InstanceConverter
	durationSinceLastSuccess *atomicduration.AtomicDuration
}

func NewReplicator(ctx context.Context, mgr ctrl.Manager, grpcClient pb.InstancePrivateServiceClient) (*Replicator, error) {
	durationSinceLastSuccess := atomicduration.New()
	// Create source that reads from GRPC InstanceServiceClient.
	lw := NewListerWatcher(grpcClient, 60*time.Second)
	// Whenever the ListerWatcher receives a Watch response, reset durationSinceLastSuccess so that
	// the health check can detect idleness.
	lw.OnWatchSuccess = durationSinceLastSuccess.Reset
	informer := toolscache.NewSharedIndexInformer(lw, &privatecloudv1alpha1.Instance{}, 0, toolscache.Indexers{})
	cache := &Cache{
		Informer: informer,
	}
	// Create replicator.
	r := &Replicator{
		informer:                 informer,
		cache:                    cache,
		k8sClient:                mgr.GetClient(),
		instanceClient:           grpcClient,
		converter:                convert.NewInstanceConverter(),
		durationSinceLastSuccess: durationSinceLastSuccess,
	}
	// Create controller.
	controllerOptions := controller.Options{
		Reconciler: r,
	}
	c, err := controller.New("instance_replicator", mgr, controllerOptions)
	if err != nil {
		return nil, err
	}
	// Connect sources to manager.
	src := source.Kind(cache, &privatecloudv1alpha1.Instance{})
	if err := c.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}
	// Connect source for the target resource in Kubernetes.
	target := source.Kind(mgr.GetCache(), &privatecloudv1alpha1.Instance{})
	if err := c.Watch(target, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}
	// Ensure manager runs informer.
	err = mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		return r.Run(ctx)
	}))
	if err != nil {
		return nil, err
	}
	if err := mgr.AddHealthzCheck("healthz", r.Healthz); err != nil {
		return r, fmt.Errorf("unable to set up health check: %w", err)
	}
	return r, nil
}

func (c *Replicator) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Replicator.Run")

	log.Info("BEGIN")
	defer log.Info("END")
	defer utilruntime.HandleCrash()
	log.Info("Service running")
	return c.cache.Start(ctx)
}

// Liveness check. Returns success (nil) if the service recently received a Watch response.
func (r *Replicator) Healthz(req *http.Request) error {
	log := log.FromContext(req.Context()).WithName("Replicator.Healthz")
	lastSuccessAge := r.durationSinceLastSuccess.SinceReset()
	log.Info("Checking health", logkeys.LastSuccessAge, lastSuccessAge)
	if lastSuccessAge > 10*time.Second {
		return fmt.Errorf("last success was %s ago", lastSuccessAge)
	}
	return nil
}

// Reconcile is called by the controller runtime when an create/update/delete event occurs
// in the Compute API Server or K8s.
// req contains only the namespace (CloudAccountId) and name (ResourceId).
func (r *Replicator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Replicator.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Reconciling instance")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the source instance from the Compute API Server (Postgres).
		instance, err := r.getSourceInstance(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if instance == nil {
			log.Info("Ignoring reconcile request because source instance was not found in cache")
			return ctrl.Result{}, nil
		}
		newStatus, result, processErr := func() (*privatecloudv1alpha1.InstanceStatus, ctrl.Result, error) {
			if instance.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.processInstance(ctx, instance)
			} else {
				return r.processDeleteInstance(ctx, instance)
			}
		}()
		if err := r.updateStatusAndPersist(ctx, instance, newStatus, processErr); err != nil {
			processErr = multierror.Append(processErr, err)
		}
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling instance")
	}
	log.Info("InstanceReconciler.Reconcile: Completed", logkeys.Result, result, logkeys.Error, reconcileErr)
	// If an error occurs, the controller runtime will schedule a retry.
	return result, reconcileErr
}

// Copy instance Spec from the Compute API Server (source) to K8s (target).
func (r *Replicator) processInstance(ctx context.Context, instance *privatecloudv1alpha1.Instance) (*privatecloudv1alpha1.InstanceStatus, ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Replicator.processInstance")
	var newStatus *privatecloudv1alpha1.InstanceStatus
	if err := r.createNamespaceIfNeeded(ctx, instance.Namespace); err != nil {
		return newStatus, ctrl.Result{}, err
	}
	isRetryable := func(err error) bool {
		return errors.IsConflict(err) || errors.IsAlreadyExists(err)
	}
	err := retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		targetInstance, err := r.getTargetInstance(ctx, instance)
		if err != nil {
			return err
		}
		if targetInstance == nil {
			logger.Info("Creating new instance", logkeys.Instance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(instance)))
			// ResourceVersion must be empty when creating a K8s resource.
			instance.ObjectMeta.ResourceVersion = ""
			return r.k8sClient.Create(ctx, instance)
		} else {
			// Target status will be returned so that updateStatusAndPersist can update the source status.
			newStatus = targetInstance.Status.DeepCopy()
			// Check if the spec has been changed.
			if reflect.DeepEqual(instance.Spec, targetInstance.Spec) {
				logger.Info("Existing instance is already up-to-date", logkeys.TargetInstance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(targetInstance)))
			} else {
				// Copy the spec from the source to the target.
				instance.Spec.DeepCopyInto(&targetInstance.Spec)
				logger.Info("Updating existing instance", logkeys.TargetInstance, fmt.Sprintf("%v", utils.TrimInstanceCloneForLogs(targetInstance)))
				return r.k8sClient.Update(ctx, targetInstance)
			}
		}
		return nil
	})
	return newStatus, ctrl.Result{}, err
}

func (r *Replicator) processDeleteInstance(ctx context.Context, instance *privatecloudv1alpha1.Instance) (*privatecloudv1alpha1.InstanceStatus, ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("Replicator.processDeleteInstance")
	log.Info("Deleting target instance")
	var newStatus *privatecloudv1alpha1.InstanceStatus
	targetInstance, err := r.getTargetInstance(ctx, instance)
	if err != nil {
		return newStatus, ctrl.Result{}, err
	} else if targetInstance != nil {
		// The status will be copied to the source even while waiting for it to be deleted.
		newStatus = targetInstance.Status.DeepCopy()
		// Delete target.
		err := r.k8sClient.Delete(ctx, instance)
		if errors.IsNotFound(err) {
			// Instance was deleted after getTargetInstance.
			targetInstance = nil
		} else if err != nil {
			return newStatus, ctrl.Result{}, err
		}
	}
	if targetInstance == nil {
		log.Info("Target instance not found. Removing finalizer from source.")
		_, err = r.instanceClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: instance.ObjectMeta.Namespace,
				ResourceId:     instance.ObjectMeta.Name,
			},
		})
		if err != nil {
			return newStatus, ctrl.Result{}, err
		}
		// Set newStatus to nil so that updateStatusAndPersist doesn't attempt to update the status.
		newStatus = nil
	}
	return newStatus, ctrl.Result{}, nil
}

// Get instance from Compute API Server.
// Returns (nil, nil) if not found.
func (r *Replicator) getSourceInstance(ctx context.Context, req ctrl.Request) (*privatecloudv1alpha1.Instance, error) {
	cachedObject, exists, err := r.informer.GetStore().GetByKey(req.NamespacedName.String())
	if err != nil {
		return nil, fmt.Errorf("getSourceInstance error: %w", err)
	}
	if !exists {
		return nil, nil
	}
	instance, ok := cachedObject.(*privatecloudv1alpha1.Instance)
	if !ok {
		return nil, fmt.Errorf("getSourceInstance error: unexpected type of cached object")
	}
	return instance, nil
}

// Get instance from K8s.
// Returns (nil, nil) if not found.
func (r *Replicator) getTargetInstance(ctx context.Context, instance *privatecloudv1alpha1.Instance) (*privatecloudv1alpha1.Instance, error) {
	targetInstance := &privatecloudv1alpha1.Instance{}
	err := r.k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, targetInstance)
	if errors.IsNotFound(err) || reflect.ValueOf(targetInstance).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getTargetInstance error: %w", err)
	}
	return targetInstance, nil
}

func (r *Replicator) createNamespaceIfNeeded(ctx context.Context, namespace string) error {
	namespaceResource := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := r.k8sClient.Create(ctx, namespaceResource); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// Copy instance Status from K8s (reconcile target) to the Compute API Server (reconcile source).
func (r *Replicator) updateStatusAndPersist(
	ctx context.Context,
	instance *privatecloudv1alpha1.Instance,
	newStatus *privatecloudv1alpha1.InstanceStatus,
	reconcileErr error) error {
	log := log.FromContext(ctx).WithName("Replicator.updateStatusAndPersist")

	oldStatus := &instance.Status
	if newStatus == nil {
		newStatus = oldStatus.DeepCopy()
	}
	// TODO: update newStatus if any errors occurred

	// For DeepEqual, do not consider conditions because they are not stored in source.
	newStatus.Conditions = oldStatus.Conditions

	if reflect.DeepEqual(oldStatus, newStatus) {
		log.V(9).Info("Status unchanged")
	} else {
		log.V(9).Info("Status changed", logkeys.OldStatus, oldStatus, logkeys.NewStatus, newStatus)
		instanceUpdateRequest, err := r.newInstanceUpdateRequest(ctx, instance, newStatus)
		if err != nil {
			return fmt.Errorf("updateStatusAndPersist: %w", err)
		}
		log.Info("Updating instance status in source", logkeys.Request, instanceUpdateRequest)
		_, err = r.instanceClient.UpdateStatus(ctx, instanceUpdateRequest)
		if status.Code(err) == codes.NotFound {
			log.Info("Instance not found in source. It may have been deleted.")
			// No need to retry reconcile.
			return nil
		}
		if err != nil {
			return fmt.Errorf("updateStatusAndPersist: %w", err)
		}
	}
	return nil
}

func (r *Replicator) newInstanceUpdateRequest(ctx context.Context, instance *privatecloudv1alpha1.Instance, newStatus *privatecloudv1alpha1.InstanceStatus) (*pb.InstanceUpdateStatusRequest, error) {
	log := log.FromContext(ctx).WithName("Replicator.newInstanceUpdateRequest")
	newInstance := &privatecloudv1alpha1.Instance{
		ObjectMeta: instance.ObjectMeta,
		Status:     *newStatus,
	}
	pbInstance, err := r.converter.K8sToPb(newInstance)
	if err != nil {
		return nil, fmt.Errorf("newInstanceUpdateRequest: %w", err)
	}
	instanceUpdateRequest := &pb.InstanceUpdateStatusRequest{
		Metadata: &pb.InstanceIdReference{
			CloudAccountId: pbInstance.Metadata.CloudAccountId,
			ResourceId:     pbInstance.Metadata.ResourceId,
		},
		Status: pbInstance.Status,
	}
	log.V(9).Info("instanceUpdateRequest", logkeys.Request, instanceUpdateRequest)
	return instanceUpdateRequest, nil
}
