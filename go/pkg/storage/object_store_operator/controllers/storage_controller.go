/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package objectStore

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	strcntr "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Finalizers
const (
	ObjectFinalizer                = "private.cloud.intel.com/objectfinalizer"
	addFinalizers                  = "add_finalizers"
	removeFinalizers               = "remove_finalizers"
	BucketMeteringMonitorFinalizer = "private.cloud.intel.com/bucketmeteringmonitorfinalizer"
	RequeueAfterDuration           = 30 * time.Second
)

// ObjectStoreReconciler reconciles a ObjectStore object
type ObjectStoreReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	StorageControllerClient *storagecontroller.StorageControllerClient
}

func NewObjectStoreOperator(ctx context.Context, mgr ctrl.Manager, strClient *storagecontroller.StorageControllerClient) (*ObjectStoreReconciler, error) {
	// Create reconciler
	r := (&ObjectStoreReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		StorageControllerClient: strClient,
	})

	// Create controller
	err := ctrl.NewControllerManagedBy(mgr).
		Named("object_storage_operator").
		For(&cloudv1alpha1.ObjectStore{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
		)).
		Complete(r)
	if err != nil {
		return nil, fmt.Errorf("unable to create object store controller: %w", err)
	}

	return r, nil
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=objectstorages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=objectstorages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=objectstorages/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Storage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *ObjectStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ObjectStoreReconciler.Reconcile").Start()
	defer span.End()

	log.Info("inside the reconciling function of object store operator")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the object_store CR.
		objectStore, err := r.getObject(ctx, req)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("Object Store  not found")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, fmt.Errorf("error encountered in fetching object context and object store custom resource")

		}
		if objectStore == nil {
			log.Info("ignoring reconcile no object")
			return ctrl.Result{}, nil
		}
		//Check if bucket is in ready state
		reconcileSkipFlag := shouldReconcileSkip(ctx, objectStore)
		var skipUpdate bool
		result, processErr := func() (ctrl.Result, error) {
			if objectStore.ObjectMeta.DeletionTimestamp.IsZero() {
				if reconcileSkipFlag {
					skipUpdate = true
					return ctrl.Result{}, nil
				}
				if err := r.initializeMetadataAndPersist(ctx, req, objectStore); err != nil {
					return ctrl.Result{}, err
				}
				log.Info("processCreateObjectStore")
				return r.createBucket(ctx, objectStore, req)
			} else {
				log.Info("processDeleteObjectStore")
				return r.deleteBucket(ctx, objectStore, req)
			}
		}()
		// update status
		if !skipUpdate {
			if err := r.updateStatusAndPersist(ctx, objectStore, processErr); err != nil {
				processErr = multierror.Append(processErr, err)
			}
			if err = r.PersistStatusUpdate(ctx, objectStore, req, processErr); err != nil {
				processErr = multierror.Append(processErr, err)
			}
		}
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "ObjectStoreReconciler.Reconcile: error reconciling object_store")
	}
	log.Info("END", logkeys.Result, result, logkeys.Error, reconcileErr)
	return result, reconcileErr

}

// SetupWithManager sets up the ObjectStore controller with the Manager.
func (r *ObjectStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1alpha1.ObjectStore{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

// Get object store from K8s.
// Returns (nil, nil) if not found.
func (r *ObjectStoreReconciler) getObject(ctx context.Context, req ctrl.Request) (*cloudv1alpha1.ObjectStore, error) {
	object_store := &cloudv1alpha1.ObjectStore{}
	err := r.Get(ctx, req.NamespacedName, object_store)
	if errors.IsNotFound(err) || reflect.ValueOf(object_store).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getObject: %w", err)
	}
	return object_store, nil
}

func (r *ObjectStoreReconciler) createBucket(ctx context.Context, bucket *cloudv1alpha1.ObjectStore, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.createBucket")
	log.Info("inside the createBucket function of object store operator")
	totalSize, err := strconv.ParseUint(bucket.Spec.Quota, 10, 64)
	if err != nil {
		fmt.Println("Error converting string to uint:", err)
		return ctrl.Result{}, err
	}

	// make request
	resp, err := r.StorageControllerClient.CreateBucket(ctx, strcntr.Bucket{
		Metadata: strcntr.BucketMetadata{
			Name:      bucket.Name,
			ClusterId: bucket.Spec.ObjectStoreBucketSchedule.ObjectStoreCluster.UUID,
		},
		Spec: strcntr.BucketSpec{
			AccessPolicy: strcntr.BucketAccessPolicy(bucket.Spec.BucketAccessPolicy),
			Versioned:    bucket.Spec.Versioned,
			Totalbytes:   totalSize,
		},
	})

	log.Info("Create bucket response", logkeys.Response, resp)
	// check for error
	if err != nil {
		log.Error(err, "error creating bucket")
		condition := cloudv1alpha1.ObjectStoreCondition{
			Type:               cloudv1alpha1.ObjectStoreConditionFailed,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		}
		SetStatusCondition(&bucket.Status.Conditions, condition)
		return ctrl.Result{RequeueAfter: RequeueAfterDuration}, err
	} else {
		log.Info("Successfully created bucket in minio")
		//TODO: update crd with resp info
		condition := cloudv1alpha1.ObjectStoreCondition{
			Type:               cloudv1alpha1.ObjectStoreConditionAccepted,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "Bucket is ready",
		}
		SetStatusCondition(&bucket.Status.Conditions, condition)
	}

	// set bucketId in object crd status
	log.Info("setting bucketId", logkeys.BucketId, resp.Metadata.BucketId)
	bucket.Status.Bucket.Id = resp.Metadata.BucketId
	// persist the status changes to the Kubernetes API server
	if err := r.PersistStatusUpdate(ctx, bucket, req, nil); err != nil {
		// handle error
		log.Error(err, "unable persist crd status after creation")
	}

	return ctrl.Result{}, nil
}

func (r *ObjectStoreReconciler) deleteBucket(ctx context.Context, bucket *cloudv1alpha1.ObjectStore, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.deleteBucket")
	log.Info("inside the deleteBucket function of object store operator")
	// make request
	err := r.StorageControllerClient.DeleteBucket(ctx, strcntr.BucketMetadata{
		Name:      bucket.Name,
		BucketId:  bucket.Name,
		ClusterId: bucket.Spec.ObjectStoreBucketSchedule.ObjectStoreCluster.UUID,
	}, true)
	if err != nil {
		log.Error(err, "error deleting bucket")
		return ctrl.Result{RequeueAfter: RequeueAfterDuration}, err
	}
	log.Info("Successfully deleted bucket in minio")
	//TODO: perform cleanup procedure as needed

	// remove finalizer from crd
	log.Info("begin removing object store finalizer.")
	finalizers := []string{ObjectFinalizer}
	err = r.UpdateFinalizers(ctx, bucket, req, removeFinalizers, finalizers)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("finished removing object store finalizer.")
	return ctrl.Result{}, nil
}

// Helper function to add a conditions to a given condition list.
func SetStatusCondition(conditions *[]cloudv1alpha1.ObjectStoreCondition, newCondition cloudv1alpha1.ObjectStoreCondition) {
	if conditions == nil {
		conditions = &[]cloudv1alpha1.ObjectStoreCondition{}
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

// Utility to find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.ObjectStoreCondition, conditionType cloudv1alpha1.ObjectStoreConditionType) *cloudv1alpha1.ObjectStoreCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

func (r *ObjectStoreReconciler) initializeMetadataAndPersist(ctx context.Context, req ctrl.Request, object *cloudv1alpha1.ObjectStore) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.initializeMetadataAndPersist")
	log.Info("BEGIN")
	var finalizers []string
	// Add Finalizer (if not present) for the deletion cleanup.
	// This only updates the in-memory Object.
	finalizers = append(finalizers, ObjectFinalizer)
	finalizers = append(finalizers, BucketMeteringMonitorFinalizer)

	// TODO: Add Object metering monitor Finalizer which can be removed by Object metering monitor only
	err := r.UpdateFinalizers(ctx, object, req, addFinalizers, finalizers)
	if err != nil {
		return err
	}
	log.Info("END")

	return nil
}

// Add/remove finalizers and persist
func (r *ObjectStoreReconciler) UpdateFinalizers(ctx context.Context, object *cloudv1alpha1.ObjectStore, req ctrl.Request, op string, finalizers []string) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.UpdateFinalizers")
	log.Info("BEGIN", logkeys.Task, op)
	defer log.Info("END", logkeys.Task, op)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestObject, err := r.getObject(ctx, req)
		if latestObject == nil {
			log.Info("object not found", logkeys.Object, object)
			return nil
		}
		if err != nil {
			log.Error(err, "failed to get the object", logkeys.Object, object)
			return err
		}
		for _, finalizer := range finalizers {
			if op == addFinalizers {
				controllerutil.AddFinalizer(latestObject, finalizer)
			} else {
				controllerutil.RemoveFinalizer(latestObject, finalizer)
			}
		}
		if !reflect.DeepEqual(object.GetFinalizers(), latestObject.GetFinalizers()) {
			log.Info("object finalizer mismatches", logkeys.CurrentObjectFinalizer,
				object.GetFinalizers(), logkeys.LatestObjectFinalizer,
				latestObject.GetFinalizers())

			if err := r.Update(ctx, latestObject); err != nil {
				return fmt.Errorf("UpdateFinalizers: update failed: %w", err)
			}
		} else {
			log.Info("object finalizer does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update object finalizers: %w", err)
	}
	return nil
}

// Update accepted condition, phase, and message.
func (r *ObjectStoreReconciler) updateStatusAndPersist(ctx context.Context, object *cloudv1alpha1.ObjectStore, reconcileErr error) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.updateStatusAndPersist")
	log.Info("inside the updateStatusAndPersist function of object store operator")

	// Logical Conditions, Reason and Message based on Status
	ObjSuccess := FindStatusCondition(object.Status.Conditions, cloudv1alpha1.ObjectStoreConditionAccepted)
	ObjCreationDone := ObjSuccess != nil && ObjSuccess.Status == k8sv1.ConditionTrue

	var condStatus k8sv1.ConditionStatus
	var reason cloudv1alpha1.ObjectStoreConditionReason
	var message string
	if ObjCreationDone {
		condStatus = k8sv1.ConditionTrue
		reason = cloudv1alpha1.ObjectStoreConditionReasonAccepted
		message = "Bucket is ready"
		if err := r.updateStatusCondition(ctx, object, cloudv1alpha1.ObjectStoreConditionAccepted, condStatus, reason, message); err != nil {
			return err
		}
	} else {
		condStatus = k8sv1.ConditionTrue
		reason = cloudv1alpha1.ObjectStoreConditionReasonNotAccepted
		message = "Bucket is unavailable"
		if err := r.updateStatusCondition(ctx, object, cloudv1alpha1.ObjectStoreConditionFailed, condStatus, reason, message); err != nil {
			return err
		}
	}

	if err := r.updateStatusPhaseAndMessage(ctx, object); err != nil {
		return fmt.Errorf("updateStatusPhaseAndMessage: %w", err)
	}

	return nil
}

// Update a status condition.
func (r *ObjectStoreReconciler) updateStatusCondition(ctx context.Context, object *cloudv1alpha1.ObjectStore,
	objectConditionType cloudv1alpha1.ObjectStoreConditionType,
	status k8sv1.ConditionStatus, reason cloudv1alpha1.ObjectStoreConditionReason, message string,
) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.updateStatusCondition")
	log.Info("inside the updateStatusCondition function of object operator")

	objectCondition := cloudv1alpha1.ObjectStoreCondition{
		Status:             status,
		Message:            message,
		Type:               objectConditionType,
		LastTransitionTime: metav1.Now(),
		LastProbeTime:      metav1.Now(),
		Reason:             reason,
	}
	SetStatusCondition(&object.Status.Conditions, objectCondition)
	return nil
}

// Update object status.
func (r *ObjectStoreReconciler) PersistStatusUpdate(ctx context.Context, object *cloudv1alpha1.ObjectStore, req ctrl.Request, reconcileErr error) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.PersistStatusUpdate")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestObject, err := r.getObject(ctx, req)
		// object is deleted
		if latestObject == nil {
			log.Info("object not found", logkeys.Object, object)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the object: %+v. error:%w", object, err)
		}
		if !equality.Semantic.DeepEqual(object.Status, latestObject.Status) {
			log.Info("object status mismatch", logkeys.ObjectStatus, object.Status, logkeys.LatestObjectStatus, latestObject.Status)
			// update latest object status
			object.Status.DeepCopyInto(&latestObject.Status)
			if err := r.Status().Update(ctx, latestObject); err != nil {
				return fmt.Errorf("PersistStatusUpdate: %w", err)
			}
		} else {
			log.Info("object status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update object status: %w", err)
	}
	log.Info("END")
	return nil
}

// Set status phase and message based on conditions.
// The message will come from the condition that is most relevant.
func (r *ObjectStoreReconciler) updateStatusPhaseAndMessage(ctx context.Context, object *cloudv1alpha1.ObjectStore) error {
	log := log.FromContext(ctx).WithName("ObjectStoreReconciler.updateStatusPhaseAndMessage")
	log.Info("BEGIN")

	acceptedCond := FindStatusCondition(object.Status.Conditions, cloudv1alpha1.ObjectStoreConditionAccepted)
	accepted := acceptedCond != nil
	terminating := !object.ObjectMeta.DeletionTimestamp.IsZero()

	var phase cloudv1alpha1.ObjectStorePhase
	var message string
	// Phase and Message is based on the above conditions.
	if terminating {
		// The bucket and its associated resources are in the process of being deleted
		phase = cloudv1alpha1.ObjectStorePhasePhaseTerminating
		message = "Bucket is in the process of being deleted"
	} else {
		// If Bucket is accepted and ready
		if accepted {
			message = "Bucket is ready"
			phase = cloudv1alpha1.ObjectStorePhasePhaseReady
		} else {
			message = "Bucket is unavailable"
			phase = cloudv1alpha1.ObjectStorePhasePhaseFailed
		}
	}

	object.Status.Phase = phase
	object.Status.Message = message

	log.Info("END: Object status details", logkeys.StatusPhase, object.Status.Phase, logkeys.StatusMessage, object.Status.Message, logkeys.StatusConditions, object.Status.Conditions)
	return nil

}

func shouldReconcileSkip(ctx context.Context, object *cloudv1alpha1.ObjectStore) bool {
	log := log.FromContext(ctx).WithName("shouldReconcileSkip")
	log.Info("inside the shouldReconcileSkip function of object store operator")
	emptyStatus := cloudv1alpha1.ObjectStoreStatus{}
	if !reflect.DeepEqual(object.Status, emptyStatus) {
		if object.Status.Phase == cloudv1alpha1.ObjectStorePhasePhaseReady {
			log.Info("ignoring reconcile for ready objects")
			return true
		} else {
			return false
		}
	}
	return false
}
