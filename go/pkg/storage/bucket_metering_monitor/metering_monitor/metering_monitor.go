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
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/object_store_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	util "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MeteringMonitor struct {
	K8sClient                         k8sclient.Client
	MeteringClient                    pb.MeteringServiceClient
	MaxUsageRecordSendIntervalMinutes int64
	StorageControllerClient           *storagecontroller.StorageControllerClient
	ServiceType                       string
	Region                            string
}

func NewMeteringMonitor(ctx context.Context, mgr ctrl.Manager, grpcClient pb.MeteringServiceClient, cfg *cloudv1alpha1.BucketMeteringMonitorConfig, strcnt *storagecontroller.StorageControllerClient) (*MeteringMonitor, error) {
	// Create reconciler
	r := &MeteringMonitor{
		K8sClient:                         mgr.GetClient(),
		MeteringClient:                    grpcClient,
		MaxUsageRecordSendIntervalMinutes: cfg.MaxUsageRecordSendIntervalMinutes,
		StorageControllerClient:           strcnt,
		ServiceType:                       cfg.ServiceType,
		Region:                            cfg.Region,
	}

	// Create controller
	err := ctrl.NewControllerManagedBy(mgr).
		Named("bk_metering_monitor").
		For(&cloudv1alpha1.ObjectStore{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			statusChangePredicate())).
		WithEventFilter(ignoreStorageWithoutFinalizerPredicate()).
		Complete(r)
	if err != nil {
		return nil, fmt.Errorf("unable to create bucket metering monitor controller: %w", err)
	}

	return r, nil
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=objectstorages,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=objectstorages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile

func (r *MeteringMonitor) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringMonitor.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	result, reconcileErr := func() (ctrl.Result, error) {
		objectStore, err := r.getObjectStorage(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if objectStore == nil {
			log.Info("Ignoring reconcile request because source object store was not found")
			return ctrl.Result{}, nil
		}

		result, processErr := func() (ctrl.Result, error) {
			if objectStore.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.processObjectStorage(ctx, objectStore)
			} else {
				return r.processDeleteObjectStorage(ctx, objectStore)
			}
		}()
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling object store metering monitor")
		return result, reconcileErr
	}

	// choose a random number of milliseconds between 50% and 100% of MaxUsageRecordSendIntervalMilliseconds so as to space the records evenly
	duration := time.Duration(r.MaxUsageRecordSendIntervalMinutes) * time.Minute
	maxUsageRecordSendIntervalMilliSeconds := duration.Milliseconds()
	requeueAfterMilliseconds := rand.Int63n(maxUsageRecordSendIntervalMilliSeconds-maxUsageRecordSendIntervalMilliSeconds/2) + maxUsageRecordSendIntervalMilliSeconds/2

	return ctrl.Result{RequeueAfter: time.Duration(requeueAfterMilliseconds) * time.Millisecond}, nil
}

// Get storage from K8s.
// Returns (nil, nil) if not found.
func (r *MeteringMonitor) getObjectStorage(ctx context.Context, req ctrl.Request) (*cloudv1alpha1.ObjectStore, error) {
	objectStore := &cloudv1alpha1.ObjectStore{}
	err := r.K8sClient.Get(ctx, req.NamespacedName, objectStore)
	if errors.IsNotFound(err) || reflect.ValueOf(objectStore).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getObjectStorage: %w", err)
	}
	return objectStore, nil
}

func (r *MeteringMonitor) processObjectStorage(ctx context.Context, objectStore *cloudv1alpha1.ObjectStore) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("MeteringMonitor.processObjectStorage")

	successCond := FindStatusCondition(objectStore.Status.Conditions, cloudv1alpha1.ObjectStoreConditionType(cloudv1alpha1.ObjectStoreConditionAccepted))
	successRunning := successCond != nil && successCond.Status == k8sv1.ConditionTrue

	failedCond := FindStatusCondition(objectStore.Status.Conditions, cloudv1alpha1.ObjectStoreConditionFailed)
	failed := failedCond != nil && failedCond.Status == k8sv1.ConditionTrue

	var runningTime time.Duration
	if successRunning {
		log.V(1).Info("Begin creating metering record once the object store is ready")
		firstReadyTimestamp := successCond.LastTransitionTime
		if failed {
			runningTime = failedCond.LastTransitionTime.Sub(firstReadyTimestamp.Time)
		} else {
			runningTime = timestamppb.Now().AsTime().Sub(firstReadyTimestamp.Time)
		}

		// runningTime may be negative in the case when the clock gets adjusted. If runningTime is negative,
		// this should return immediately without an error i.e no need to requeue it right away
		// if runningTime is 0 means failure status, do not create metering record
		if runningTime <= 0 {
			return reconcile.Result{}, nil
		}

		err := r.CreateRecordInDB(ctx, objectStore, runningTime, firstReadyTimestamp, false)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to create the metering record: %w", err)
		}

		log.V(1).Info("Finished creating metering record after the object store is ready")
	}

	return reconcile.Result{}, nil
}

func (r *MeteringMonitor) processDeleteObjectStorage(ctx context.Context, objectStore *cloudv1alpha1.ObjectStore) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MeteringMonitor.processDeleteObjectStorage")
	log.V(1).Info("Begin creating metering record before object store deletion")

	var runningTime time.Duration
	successCond := FindStatusCondition(objectStore.Status.Conditions, cloudv1alpha1.ObjectStoreConditionType(cloudv1alpha1.ObjectStoreConditionAccepted))

	failedCond := FindStatusCondition(objectStore.Status.Conditions, cloudv1alpha1.ObjectStoreConditionFailed)
	failed := failedCond != nil && failedCond.Status == k8sv1.ConditionTrue

	var firstReadyTimestamp metav1.Time
	// startup Failed: skip creating metering record
	// Running -> Fail: charge based on time up til failure

	// Check if successCond is not nil and true
	if successCond != nil {
		// Running or was Running
		firstReadyTimestamp = successCond.LastTransitionTime
	} else {
		if failedCond != nil {
			//successCond is nil and startup Failed never reach Running
			firstReadyTimestamp = failedCond.LastTransitionTime
		}

	}

	if failed {
		runningTime = failedCond.LastTransitionTime.Sub(firstReadyTimestamp.Time) // runningTime = 0
	} else {
		runningTime = objectStore.DeletionTimestamp.Time.Sub(firstReadyTimestamp.Time) //runningTime > 0
	}

	// runningTime may be negative in the case when the clock gets adjusted. If runningTime is negative,
	// this should not create a record, only create record when runningTime is greater than 0
	if runningTime > 0 {
		err := r.CreateRecordInDB(ctx, objectStore, runningTime, firstReadyTimestamp, true)
		if err != nil {
			log.Error(err, "failed to create the metering record before deletion")
		}
	}

	// Remove object sore Metering Monitor finalizer and update it
	controllerutil.RemoveFinalizer(objectStore, privatecloud.BucketMeteringMonitorFinalizer)

	// Update object storage resource
	if err := r.K8sClient.Update(ctx, objectStore); err != nil {
		return reconcile.Result{}, fmt.Errorf("processDeleteObjectStorage: removing finalizer: %w", err)
	}

	log.V(1).Info("Finished creating metering record before object store deletion")
	return reconcile.Result{}, nil
}

func (r *MeteringMonitor) CreateRecordInDB(ctx context.Context, objectStore *cloudv1alpha1.ObjectStore, runningTime time.Duration, firstReadyTimestamp metav1.Time, isDeleted bool) error {
	log := log.FromContext(ctx).WithName("MeteringMonitor.CreateRecordInDB")
	log.V(1).Info("Begin creating object store metering record in DB")
	//make call to backend controller
	// getBucketParams := storagecontroller.BucketFilter{
	// 	ClusterId: objectStore.Spec.ObjectStoreBucketSchedule.ObjectStoreCluster.UUID,
	// 	BucketId:  objectStore.Name,
	// }
	totalStorageSize := "0"
	if !isDeleted {
		// resp, err := r.StorageControllerClient.GetBucket(ctx, getBucketParams)
		// log.Info("Fetch bucket response", logkeys.Response, resp)
		// // check for error
		// if err != nil {
		// 	log.Error(err, "error fetch object store bucket")
		// 	return fmt.Errorf("failed to create the metering record as bucket fetch failed: %w", err)
		// }
		// if resp != nil {
		// 	totalStorageSize = strconv.FormatUint(resp.Spec.Totalbytes-resp.Spec.AvailableBytes, 10)
		// }
		totalStorageSize = objectStore.Spec.Quota
	}

	// Since these usage records use cumulative counters, deduplication is not needed. It is effectively disabled by using a random transactionId.
	transactionId := uuid.NewString()
	cloudAccountId := string(objectStore.ObjectMeta.Namespace)
	runningTimeInHours := "1" // runningTime minimum 1 hour
	if runningTime.Hours() > 1 {
		runningTimeInHours = fmt.Sprintf("%v", runningTime.Hours())
	}
	bytes, err := strconv.ParseInt(totalStorageSize, 10, 64)
	if err != nil {
		log.Error(err, "error parsing byte string")
	}
	usage := util.BytesToTerabytes(bytes)
	properties := map[string]string{
		"availabilityZone":    objectStore.Spec.AvailabilityZone,
		"bucketName":          objectStore.Status.Bucket.Id,
		"bucketConditionType": string(objectStore.Status.Conditions[0].Type),
		"firstReadyTimestamp": firstReadyTimestamp.Format(time.RFC3339),
		"deleted":             strconv.FormatBool(isDeleted),
		"serviceType":         r.ServiceType,
		"hour":                runningTimeInHours,
		"TB":                  strconv.FormatFloat(usage, 'f', 3, 64),
		"region":              r.Region,
	}
	meteringRecord := &pb.UsageCreate{
		TransactionId:  transactionId,
		ResourceId:     string(objectStore.ObjectMeta.UID),
		CloudAccountId: cloudAccountId,
		Timestamp:      timestamppb.Now(),
		Properties:     properties,
	}

	_, err = r.MeteringClient.Create(ctx, meteringRecord)
	if err != nil {
		return err
	}

	log.Info("Created metering record", logkeys.MeteringRecord, meteringRecord)
	return nil
}

// Predicate that returns true when Status.Phase changes.
// This ignores updates to Status.Conditions.
func statusChangePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.(*cloudv1alpha1.ObjectStore).Status.Phase != e.ObjectNew.(*cloudv1alpha1.ObjectStore).Status.Phase
		},
	}
}

func ignoreStorageWithoutFinalizerPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Filtering all the events which does not contain Bucket Metering Monitor Finalizer
			// Without this whenever the metering reconciler removes the finalizer, another update
			// event is triggered resulting in duplication of record in metering db
			return slices.Contains(e.ObjectNew.GetFinalizers(), privatecloud.BucketMeteringMonitorFinalizer)
		},
	}
}

// To find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.ObjectStoreCondition, conditionType cloudv1alpha1.ObjectStoreConditionType) *cloudv1alpha1.ObjectStoreCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
