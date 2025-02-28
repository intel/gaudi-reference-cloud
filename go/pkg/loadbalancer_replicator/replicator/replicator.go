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
	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/convert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/atomicduration"
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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// LoadBalancer Replicator reconciles LoadBalancer objects from the Compute API Server to K8s objects of kind loadbalancer.private.cloud.intel.com.
// It copies lb Status and deletion confirmation (removal of finalizer) from K8s to the Compute API Server.
// See https://docs.bitnami.com/tutorials/kubewatch-an-example-of-kubernetes-custom-controller/
// See https://komodor.com/learn/controller-manager/
type LoadBalancerReplicator struct {
	informer                 toolscache.SharedIndexInformer
	cache                    cache.Cache
	k8sClient                k8sclient.Client
	loadbalancerClient       pb.LoadBalancerPrivateServiceClient
	converter                *convert.LoadBalancerConverter
	durationSinceLastSuccess *atomicduration.AtomicDuration
}

func NewLoadBalancerReplicator(ctx context.Context, mgr ctrl.Manager, grpcClient pb.LoadBalancerPrivateServiceClient,
	regionId, availabilityZoneId string) (*LoadBalancerReplicator, error) {

	converter, err := convert.NewLoadBalancerConverter(regionId, availabilityZoneId)
	if err != nil {
		return nil, err
	}

	durationSinceLastSuccess := atomicduration.New()
	// Create source that reads from GRPC LoadBalancerServiceClient.
	lw, err := NewLoadBalancerListerWatcher(grpcClient, 60*time.Second, converter)
	if err != nil {
		return nil, err
	}

	// Whenever the ListerWatcher receives a Watch response, reset durationSinceLastSuccess so that
	// the health check can detect idleness.
	lw.OnWatchSuccess = durationSinceLastSuccess.Reset
	informer := toolscache.NewSharedIndexInformer(lw, &lbv1alpha1.Loadbalancer{}, 0, toolscache.Indexers{})
	cache := &Cache{
		Informer: informer,
	}

	// Create replicator.
	r := &LoadBalancerReplicator{
		informer:                 informer,
		cache:                    cache,
		k8sClient:                mgr.GetClient(),
		loadbalancerClient:       grpcClient,
		converter:                converter,
		durationSinceLastSuccess: durationSinceLastSuccess,
	}
	// Create controller.
	controllerOptions := controller.Options{
		Reconciler: r,
	}
	c, err := controller.New("loadbalancer_replicator", mgr, controllerOptions)
	if err != nil {
		return nil, err
	}
	// Connect sources to manager.
	src := source.Kind(cache, &lbv1alpha1.Loadbalancer{})
	if err := c.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}
	// Connect source for the target resource in Kubernetes.
	target := source.Kind(mgr.GetCache(), &lbv1alpha1.Loadbalancer{})
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

func (c *LoadBalancerReplicator) Run(ctx context.Context) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerReplicator.Run").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	defer utilruntime.HandleCrash()
	log.Info("Service running")
	return c.cache.Start(ctx)
}

// Liveness check. Returns success (nil) if the service recently received a Watch response.
func (r *LoadBalancerReplicator) Healthz(req *http.Request) error {
	ctx := req.Context()
	log := log.FromContext(ctx).WithName("LoadBalancerReplicator.Healthz")
	lastSuccessAge := r.durationSinceLastSuccess.SinceReset()
	log.Info("Checking health", "lastSuccessAge", lastSuccessAge)
	if lastSuccessAge > 10*time.Second {
		return fmt.Errorf("last success was %s ago", lastSuccessAge)
	}
	return nil
}

// Reconcile is called by the controller runtime when an create/update/delete event occurs
// in the Compute API Server or K8s.
// req contains only the namespace (CloudAccountId) and name (ResourceId).
func (r *LoadBalancerReplicator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("LoadBalancerReplicator.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Reconciling loadbalancer")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the source load balancer from the Compute API Server (Postgres).
		loadbalancer, err := r.getSourceLoadBalancer(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if loadbalancer == nil {
			log.Info("Ignoring reconcile request because source load balancer was not found in cache")
			return ctrl.Result{}, nil
		}

		newStatus, result, processErr := func() (*lbv1alpha1.LoadbalancerStatus, ctrl.Result, error) {
			if loadbalancer.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.processLoadBalancer(ctx, loadbalancer)
			} else {
				return r.processDeleteLoadBalancer(ctx, loadbalancer)
			}
		}()
		if err := r.updateStatusAndPersist(ctx, loadbalancer, newStatus); err != nil {
			processErr = multierror.Append(processErr, err)
		}
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling load balancer")
	}
	log.Info("LoadBalancerReconciler.Reconcile: Completed", logkeys.Result, result, logkeys.Error, reconcileErr)
	// If an error occurs, the controller runtime will schedule a retry.
	return result, reconcileErr
}

// Copy load balancer Spec from the Compute API Server (source) to K8s (target).
func (r *LoadBalancerReplicator) processLoadBalancer(ctx context.Context, loadbalancer *lbv1alpha1.Loadbalancer) (*lbv1alpha1.LoadbalancerStatus, ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerReplicator.processLoadBalancer")
	var newStatus *lbv1alpha1.LoadbalancerStatus
	if err := r.createNamespaceIfNeeded(ctx, loadbalancer.Namespace); err != nil {
		return newStatus, ctrl.Result{}, err
	}
	isRetryable := func(err error) bool {
		return errors.IsConflict(err) || errors.IsAlreadyExists(err)
	}
	err := retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		targetLoadBalancer, err := r.getTargetLoadBalancer(ctx, loadbalancer)
		if err != nil {
			return err
		}
		if targetLoadBalancer == nil {
			log.Info("Creating new load balancer", logkeys.LoadBalancer, loadbalancer, logkeys.LoadBalancerSpec, loadbalancer.Spec)
			// ResourceVersion must be empty when creating a K8s resource.
			loadbalancer.ObjectMeta.ResourceVersion = ""
			return r.k8sClient.Create(ctx, loadbalancer)
		} else {
			// Target status will be returned so that updateStatusAndPersist can update the source status.
			newStatus = targetLoadBalancer.Status.DeepCopy()
			// Check if the spec has been changed.
			if reflect.DeepEqual(loadbalancer.Spec, targetLoadBalancer.Spec) {
				log.Info("Existing load balancer is already up-to-date", logkeys.TargetLoadBalancer, targetLoadBalancer)
			} else {
				// Copy the spec from the source to the target.
				loadbalancer.Spec.DeepCopyInto(&targetLoadBalancer.Spec)
				log.Info("Updating existing load balancer", logkeys.TargetLoadBalancer, targetLoadBalancer, logkeys.TargetLoadBalancerSpec, targetLoadBalancer.Spec)
				return r.k8sClient.Update(ctx, targetLoadBalancer)
			}
		}
		return nil
	})
	return newStatus, ctrl.Result{}, err
}

func (r *LoadBalancerReplicator) processDeleteLoadBalancer(ctx context.Context, loadbalancer *lbv1alpha1.Loadbalancer) (*lbv1alpha1.LoadbalancerStatus, ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerReplicator.processDeleteLoadBalancer")
	log.Info("Deleting target loadbalancer")
	var newStatus *lbv1alpha1.LoadbalancerStatus
	targetLoadBalancer, err := r.getTargetLoadBalancer(ctx, loadbalancer)
	if err != nil {
		return newStatus, ctrl.Result{}, err
	} else if targetLoadBalancer != nil {
		// The status will be copied to the source even while waiting for it to be deleted.
		newStatus = targetLoadBalancer.Status.DeepCopy()
		// Delete target.
		err := r.k8sClient.Delete(ctx, loadbalancer)
		if errors.IsNotFound(err) {
			// LoadBalancer was deleted after getTargetLoadBalancer.
			targetLoadBalancer = nil
		} else if err != nil {
			return newStatus, ctrl.Result{}, err
		}
	}
	if targetLoadBalancer == nil {
		log.Info("Target loadbalancer not found. Removing finalizer from source.")
		_, err = r.loadbalancerClient.RemoveFinalizer(ctx, &pb.LoadBalancerRemoveFinalizerRequest{
			Metadata: &pb.LoadBalancerIdReference{
				CloudAccountId: loadbalancer.ObjectMeta.Namespace,
				ResourceId:     loadbalancer.ObjectMeta.Name,
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

// Get LoadBalancer from Compute API Server.
// Returns (nil, nil) if not found.
func (r *LoadBalancerReplicator) getSourceLoadBalancer(ctx context.Context, req ctrl.Request) (*lbv1alpha1.Loadbalancer, error) {

	cachedObject, exists, err := r.informer.GetStore().GetByKey(req.NamespacedName.String())
	if err != nil {
		return nil, fmt.Errorf("getSourceLoadBalancer error: %w", err)
	}
	if !exists {
		return nil, nil
	}
	loadbalancer, ok := cachedObject.(*lbv1alpha1.Loadbalancer)
	if !ok {
		return nil, fmt.Errorf("getSourceLoadBalancer error: unexpected type of cached object")
	}
	return loadbalancer, nil
}

// Get LoadBalancer from K8s.
// Returns (nil, nil) if not found.
func (r *LoadBalancerReplicator) getTargetLoadBalancer(ctx context.Context, loadbalancer *lbv1alpha1.Loadbalancer) (*lbv1alpha1.Loadbalancer, error) {
	targetLoadBalancer := &lbv1alpha1.Loadbalancer{}
	err := r.k8sClient.Get(ctx, types.NamespacedName{Name: loadbalancer.Name, Namespace: loadbalancer.Namespace}, targetLoadBalancer)
	if errors.IsNotFound(err) || reflect.ValueOf(targetLoadBalancer).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getTargeLoadBalancer error: %w", err)
	}
	return targetLoadBalancer, nil
}

func (r *LoadBalancerReplicator) createNamespaceIfNeeded(ctx context.Context, namespace string) error {
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

// Copy load balancer Status from K8s (reconcile target) to the Compute API Server (reconcile source).
func (r *LoadBalancerReplicator) updateStatusAndPersist(
	ctx context.Context,
	loadbalancer *lbv1alpha1.Loadbalancer,
	newStatus *lbv1alpha1.LoadbalancerStatus) error {

	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("LoadBalancerReplicator.updateStatusAndPersist").Start()
	defer span.End()

	oldStatus := &loadbalancer.Status
	if newStatus == nil {
		newStatus = oldStatus.DeepCopy()
	}

	// Check for equality.
	areEqual := areEqual(*oldStatus, *newStatus)

	if areEqual {
		log.V(9).Info("Status unchanged")
	} else {
		log.V(9).Info("Status changed", "oldStatus", oldStatus, "newStatus", newStatus)
		loadbalancerUpdateRequest, err := r.newLoadBalancerUpdateRequest(ctx, loadbalancer, newStatus)
		if err != nil {
			return fmt.Errorf("updateStatusAndPersist: %w", err)
		}
		log.Info("Updating load balancer status in source", logkeys.Request, loadbalancerUpdateRequest)
		_, err = r.loadbalancerClient.UpdateStatus(ctx, loadbalancerUpdateRequest)
		if status.Code(err) == codes.NotFound {
			log.Info("LoadBalancer not found in source. It may have been deleted.")
			// No need to retry reconcile.
			return nil
		}
		if err != nil {
			return fmt.Errorf("updateStatusAndPersist: %w", err)
		}
	}
	return nil
}

func areEqual(oldStatus lbv1alpha1.LoadbalancerStatus, newStatus lbv1alpha1.LoadbalancerStatus) bool {

	// Check if the old status and new status are different
	return reflect.DeepEqual(oldStatus, newStatus)
}

func (r *LoadBalancerReplicator) newLoadBalancerUpdateRequest(ctx context.Context, loadbalancer *lbv1alpha1.Loadbalancer, newStatus *lbv1alpha1.LoadbalancerStatus) (*pb.LoadBalancerUpdateStatusRequest, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerReplicator.newLoadBalancerUpdateRequest")
	newLoadBalancer := &lbv1alpha1.Loadbalancer{
		ObjectMeta: loadbalancer.ObjectMeta,
		Status:     *newStatus,
	}
	pbLoadBalancer, err := r.converter.K8sToPb(newLoadBalancer)
	if err != nil {
		return nil, fmt.Errorf("newLoadBalancerUpdateRequest: %w", err)
	}
	loadbalancerUpdateRequest := &pb.LoadBalancerUpdateStatusRequest{
		Metadata: &pb.LoadBalancerIdReference{
			CloudAccountId: pbLoadBalancer.Metadata.CloudAccountId,
			ResourceId:     pbLoadBalancer.Metadata.ResourceId,
		},
		Status: pbLoadBalancer.Status,
	}
	log.V(9).Info("loadbalancerUpdateRequest", "loadbalancerUpdateRequest", loadbalancerUpdateRequest)
	return loadbalancerUpdateRequest, nil
}
