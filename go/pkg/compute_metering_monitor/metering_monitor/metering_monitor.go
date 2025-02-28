// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metering_monitor

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/controllers"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MeteringMonitor struct {
	K8sClient                  k8sclient.Client
	MeteringClient             pb.MeteringServiceClient
	MaxUsageRecordSendInterval time.Duration
}

func NewMeteringMonitor(ctx context.Context, mgr ctrl.Manager, grpcClient pb.MeteringServiceClient, maxUsageRecordSendInterval time.Duration) (*MeteringMonitor, error) {
	// Create reconciler
	r := &MeteringMonitor{
		K8sClient:                  mgr.GetClient(),
		MeteringClient:             grpcClient,
		MaxUsageRecordSendInterval: maxUsageRecordSendInterval,
	}

	// Create controller
	err := ctrl.NewControllerManagedBy(mgr).
		Named("metering_monitor").
		For(&cloudv1alpha1.Instance{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			statusChangePredicate())).
		WithEventFilter(ignoreInstanceWithoutFinalizerPredicate()).
		Complete(r)
	if err != nil {
		return nil, fmt.Errorf("unable to create metering monitor controller: %w", err)
	}

	return r, nil
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile

func (r *MeteringMonitor) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringMonitor.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Begin")

	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the source instance
		instance, err := r.getInstance(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if instance == nil {
			log.Info("Ignoring reconcile request because source instance was not found")
			return ctrl.Result{}, nil
		}

		result, processErr := func() (ctrl.Result, error) {
			if instance.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.processInstance(ctx, instance)
			} else {
				return r.processDeleteInstance(ctx, instance)
			}
		}()
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling metering monitor")
		return result, reconcileErr
	}

	// choose a random number of milliseconds between 50% and 100% of MaxUsageRecordSendIntervalMilliseconds so as to space the records evenly
	maxUsageRecordSendIntervalMilliSeconds := r.MaxUsageRecordSendInterval.Milliseconds()
	requeueAfterMilliseconds := rand.Int63n(maxUsageRecordSendIntervalMilliSeconds-maxUsageRecordSendIntervalMilliSeconds/2) + maxUsageRecordSendIntervalMilliSeconds/2

	return ctrl.Result{RequeueAfter: time.Duration(requeueAfterMilliseconds) * time.Millisecond}, nil
}

// Get instance from K8s.
// Returns (nil, nil) if not found.
func (r *MeteringMonitor) getInstance(ctx context.Context, req ctrl.Request) (*cloudv1alpha1.Instance, error) {
	instance := &cloudv1alpha1.Instance{}
	err := r.K8sClient.Get(ctx, req.NamespacedName, instance)
	if errors.IsNotFound(err) || reflect.ValueOf(instance).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getInstance: %w", err)
	}
	return instance, nil
}

func (r *MeteringMonitor) processInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("MeteringMonitor.processInstance")

	startupCompleteCond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
	startupComplete := startupCompleteCond != nil && startupCompleteCond.Status == k8sv1.ConditionTrue

	failedCond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionFailed)
	failed := failedCond != nil && failedCond.Status == k8sv1.ConditionTrue

	var runningTime time.Duration
	if startupComplete {
		log.V(1).Info("Begin creating metering record once the instance starts running and has completed running startup scripts")

		firstReadyTimestamp := startupCompleteCond.LastTransitionTime
		if failed {
			runningTime = failedCond.LastTransitionTime.Sub(firstReadyTimestamp.Time)
		} else {
			runningTime = timestamppb.Now().AsTime().Sub(firstReadyTimestamp.Time)
		}

		// runningTime may be negative in the case when the clock gets adjusted. If runningTime is negative,
		// this should return immediately without an error i.e no need to requeue it right away
		if runningTime < 0 {
			return reconcile.Result{}, nil
		}

		err := r.CreateRecordInDB(ctx, instance, runningTime, firstReadyTimestamp, false)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to create the metering record: %w", err)
		}

		log.V(1).Info("Finished creating metering record after the instance starts running and has completed running startup scripts")
	}

	return reconcile.Result{}, nil
}

func (r *MeteringMonitor) processDeleteInstance(ctx context.Context, instance *cloudv1alpha1.Instance) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MeteringMonitor.processDeleteInstance")
	log.V(1).Info("Begin creating metering record before instance deletion")

	var runningTime time.Duration
	startupCompleteCond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
	failedCond := FindStatusCondition(instance.Status.Conditions, cloudv1alpha1.InstanceConditionFailed)
	failed := failedCond != nil && failedCond.Status == k8sv1.ConditionTrue
	firstReadyTimestamp := startupCompleteCond.LastTransitionTime
	if failed {
		runningTime = failedCond.LastTransitionTime.Sub(firstReadyTimestamp.Time)
	} else {
		runningTime = instance.DeletionTimestamp.Time.Sub(firstReadyTimestamp.Time)
	}

	// runningTime may be negative in the case when the clock gets adjusted. If runningTime is negative,
	// this should not create a record
	if runningTime >= 0 {
		err := r.CreateRecordInDB(ctx, instance, runningTime, firstReadyTimestamp, true)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to create the metering record before deletion: %w", err)
		}
	}

	// Remove Instance Metering Monitor finalizer and update it
	controllerutil.RemoveFinalizer(instance, privatecloud.InstanceMeteringMonitorFinalizer)
	if err := r.K8sClient.Update(ctx, instance); err != nil {
		return reconcile.Result{}, fmt.Errorf("processDeleteInstance: removing finalizer: %w", err)
	}

	log.V(1).Info("Finished creating metering record before instance deletion")
	return reconcile.Result{}, nil
}

func (r *MeteringMonitor) CreateRecordInDB(ctx context.Context, instance *cloudv1alpha1.Instance, runningTime time.Duration, firstReadyTimestamp metav1.Time, isDeleted bool) error {
	log := log.FromContext(ctx).WithName("MeteringMonitor.createRecordinDB")
	log.V(1).Info("Begin creating metering record in DB")

	// Since these usage records use cumulative counters, deduplication is not needed. It is effectively disabled by using a random transactionId.
	transactionId := uuid.NewString()
	runningTimeInSeconds := fmt.Sprintf("%v", runningTime.Seconds())
	properties := map[string]string{
		"instanceName":        instance.Spec.InstanceName,
		"instanceType":        instance.Spec.InstanceTypeSpec.Name,
		"region":              instance.Spec.Region,
		"availabilityZone":    instance.Spec.AvailabilityZone,
		"clusterId":           instance.Spec.ClusterId,
		"serviceType":         instance.Spec.ServiceType,
		"firstReadyTimestamp": firstReadyTimestamp.Format(time.RFC3339),
		"runningSeconds":      runningTimeInSeconds,
		"deleted":             strconv.FormatBool(isDeleted),
		"instanceGroup":       instance.Spec.InstanceGroup,
		"instanceGroupSize":   strconv.FormatInt(int64(instance.Spec.InstanceGroupSize), 10),
	}

	meteringRecord := &pb.UsageCreate{
		TransactionId:  transactionId,
		ResourceId:     instance.ObjectMeta.Name,
		CloudAccountId: instance.ObjectMeta.Namespace,
		Timestamp:      timestamppb.Now(),
		Properties:     properties,
	}
	_, err := r.MeteringClient.Create(ctx, meteringRecord)
	if err != nil {
		return err
	}

	log.Info("Created metering record", logkeys.MeteringRecord, meteringRecord)
	return nil
}

// Predicate that returns true when Status.Phase changes.
// This ignores updates to Status.Conditions.
func statusChangePredicate() predicate.Predicate {
	// To enable dump in the diff between old and new instance objects
	// diff := cmp.Diff(e.ObjectOld.(*cloudv1alpha1.Instance), e.ObjectNew.(*cloudv1alpha1.Instance))
	// log.Info("diff", "diff", diff)
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.(*cloudv1alpha1.Instance).Status.Phase != e.ObjectNew.(*cloudv1alpha1.Instance).Status.Phase
		},
	}
}

func ignoreInstanceWithoutFinalizerPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Filtering all the events which does not contain Instance Metering Monitor Finalizer
			// Without this whenever the metering reconciler removes the finalizer, another update
			// event is triggered resulting in duplication of record in metering db
			return slices.Contains(e.ObjectNew.GetFinalizers(), privatecloud.InstanceMeteringMonitorFinalizer)
		},
	}
}

// To find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.InstanceCondition, conditionType cloudv1alpha1.InstanceConditionType) *cloudv1alpha1.InstanceCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
