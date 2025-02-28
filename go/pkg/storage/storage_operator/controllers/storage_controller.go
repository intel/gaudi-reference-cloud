// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
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

package cloud

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	k8sv1 "k8s.io/api/core/v1"

	"github.com/hashicorp/go-multierror"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/equality"
	apiError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Finalizers
const (
	StorageFinalizer                = "private.cloud.intel.com/storagefinalizer"
	StorageMeteringMonitorFinalizer = "private.cloud.intel.com/storagemeteringmonitorfinalizer"
	addFinalizers                   = "add_finalizers"
	removeFinalizers                = "remove_finalizers"
	RequeueAfterDuration            = 30 * time.Second
)

// StorageReconciler reconciles a Storage object
type StorageReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	StorageControllerClient *storagecontroller.StorageControllerClient
	kmsClient               pb.StorageKMSPrivateServiceClient
}

func NewStorageOperator(ctx context.Context, mgr ctrl.Manager, strClient *storagecontroller.StorageControllerClient, kms pb.StorageKMSPrivateServiceClient) (*StorageReconciler, error) { // Create reconciler
	r := (&StorageReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		StorageControllerClient: strClient,
		kmsClient:               kms,
	})

	// Create controller
	err := ctrl.NewControllerManagedBy(mgr).
		Named("fs_storage_operator").
		For(&cloudv1alpha1.Storage{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
		)).
		Complete(r)
	if err != nil {
		return nil, fmt.Errorf("unable to create storage controller: %w", err)
	}

	return r, nil
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=storages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=storages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=storages/finalizers,verbs=update
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
func (r *StorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Storage.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("inside the reconciling function of storage operator")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the storage FS.
		storage, err := r.getStorage(ctx, req)
		if err != nil {
			if apiError.IsNotFound(err) {
				log.Info("Storage FS not found")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, fmt.Errorf("error encountered in fetching storage context and storage FS")

		}
		if storage == nil {
			log.Info("ignoring reconcile no storage")
			return ctrl.Result{}, nil
		}
		reconcileSkipFlag := shouldReconcileSkip(ctx, storage)
		result, processErr := func() (ctrl.Result, error) {
			if storage.ObjectMeta.DeletionTimestamp.IsZero() {
				if reconcileSkipFlag {
					return ctrl.Result{}, nil
				}
				if err := r.initializeMetadataAndPersist(ctx, req, storage); err != nil {
					return ctrl.Result{}, err
				}
				return r.processStorageCreateUpdate(ctx, storage)
			} else {
				return r.processDeleteStorage(ctx, storage, req)
			}
		}()
		// update status
		// Skip update status if reconcile skip flag is true, as the object is in failed state.
		requeueDeletingFlag := shouldRequeueDeleting(ctx, storage)
		if requeueDeletingFlag {
			return ctrl.Result{RequeueAfter: RequeueAfterDuration}, nil
		}
		if reconcileSkipFlag {
			return ctrl.Result{}, nil
		} else {
			if err := r.updateStatusAndPersist(ctx, storage, processErr); err != nil {
				processErr = multierror.Append(processErr, err)
			}
			if err = r.PersistStatusUpdate(ctx, storage, req, processErr); err != nil {
				processErr = multierror.Append(processErr, err)
			}
		}
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "StorageReconciler.Reconcile: error reconciling Storage")
	}
	log.Info("END", logkeys.Result, result, logkeys.Error, reconcileErr)
	return result, reconcileErr

}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1alpha1.Storage{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func shouldReconcileSkip(ctx context.Context, storage *cloudv1alpha1.Storage) bool {
	log := log.FromContext(ctx).WithName("shouldReconcileSkip")
	log.Info("inside the shouldReconcileSkip function of storage operator")
	emptyStatus := cloudv1alpha1.StorageStatus{}
	if !reflect.DeepEqual(storage.Status, emptyStatus) {
		if storage.Status.Size != "" && storage.Status.Size != storage.Spec.StorageRequest.Size {
			return false
		}
		if storage.Status.Phase == cloudv1alpha1.FilesystemPhaseFailed || storage.Status.Phase == cloudv1alpha1.FilesystemPhaseReady {
			log.Info("ignoring reconcile for failed storage or already ready volume")
			return true
		} else {
			return false
		}
	}
	return false
}

// Add finalizer and storage category label.
// Persists the Storage.
func (r *StorageReconciler) initializeMetadataAndPersist(ctx context.Context, req ctrl.Request, storage *cloudv1alpha1.Storage) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.initializeMetadataAndPersist")
	log.Info("BEGIN")
	var finalizers []string
	// Add Finalizer (if not present) for the deletion cleanup.
	// This only updates the in-memory Storage.
	finalizers = append(finalizers, StorageFinalizer)

	// Add Storage metering monitor Finalizer which can be removed by compute metering monitor only
	finalizers = append(finalizers, StorageMeteringMonitorFinalizer)
	err := r.UpdateFinalizers(ctx, storage, req, addFinalizers, finalizers)
	if err != nil {
		return err
	}

	log.Info("END")
	return nil
}
func (r *StorageReconciler) processStorageCreateUpdate(ctx context.Context, storage *cloudv1alpha1.Storage) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("StorageReconciler.processStorageCreateUpdate")
	log.Info("inside the processStorageCreateUpdate function of storage operator")
	if err := r.validateCreateRequest(ctx, storage); err != nil {
		return ctrl.Result{}, err
	} else {
		log.Info("storage provisioning acceptance validation completed in order")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionAccepted,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            cloudv1alpha1.StorageMessageProvisioningAccepted,
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
	}
	if storage.Status.Size != "" && storage.Status.Size != storage.Spec.StorageRequest.Size && storage.Spec.FilesystemType != cloudv1alpha1.FilesystemTypeComputeKubernetes {
		if err := r.updateFileSystem(ctx, storage); err != nil {
			log.Info("update filesystem request couldn't be completed")
			condition := cloudv1alpha1.StorageCondition{
				Type:               cloudv1alpha1.StorageConditionUpdateFSSuccess,
				Status:             k8sv1.ConditionFalse,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            err.Error(),
			}
			util.SetStatusCondition(&storage.Status.Conditions, condition)
			return ctrl.Result{}, err
		} else {
			log.Info("update FS request completed")
			condition := cloudv1alpha1.StorageCondition{
				Type:               cloudv1alpha1.StorageConditionUpdateFSSuccess,
				Status:             k8sv1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Message:            "update FS request completed",
			}
			util.SetStatusCondition(&storage.Status.Conditions, condition)
			return ctrl.Result{}, nil
		}

	}
	if err := r.createNamespaceIfNotExists(ctx, storage); err != nil {
		log.Info("create namespace request couldn't be completed")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionNamespaceSuccess,
			Status:             k8sv1.ConditionFalse,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
		return ctrl.Result{}, err
	} else {
		log.Info("create namespace request completed")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionNamespaceSuccess,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "Create Namespace request completed",
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
	}
	if storage.Spec.FilesystemType == cloudv1alpha1.FilesystemTypeComputeKubernetes {
		log.Info("k8s compute type flow, No need to create filesystem in operator, Only namespace creation")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionNamespaceK8sSuccess,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "K8s compute Path completed, Namespace created but not FS",
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)

		condition = cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionRunning,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            cloudv1alpha1.StorageMessageRunning,
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
		return ctrl.Result{}, nil
	}
	if err := r.createFileSystem(ctx, storage); err != nil {
		log.Info("create filesystem request couldn't be completed")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionFSSuccess,
			Status:             k8sv1.ConditionFalse,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
		return ctrl.Result{}, err
	} else {
		log.Info("create FS request completed")
		condition := cloudv1alpha1.StorageCondition{
			Type:               cloudv1alpha1.StorageConditionFSSuccess,
			Status:             k8sv1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            "Create FS request completed",
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
	}

	log.Info("namespace, file system provisioning completed in order")
	condition := cloudv1alpha1.StorageCondition{
		Type:               cloudv1alpha1.StorageConditionRunning,
		Status:             k8sv1.ConditionTrue,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Message:            cloudv1alpha1.StorageMessageRunning,
	}
	util.SetStatusCondition(&storage.Status.Conditions, condition)
	return ctrl.Result{}, nil
}

// Add/remove finalizers and persist
func (r *StorageReconciler) UpdateFinalizers(ctx context.Context, storage *cloudv1alpha1.Storage, req ctrl.Request, op string, finalizers []string) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.UpdateFinalizers")
	log.Info("BEGIN", logkeys.Task, op)
	defer log.Info("END", logkeys.Task, op)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestStorage, err := r.getStorage(ctx, req)
		if latestStorage == nil {
			log.Info("storage not found", logkeys.Storage, storage)
			return nil
		}
		if err != nil {
			log.Error(err, "failed to get the storage", logkeys.Storage, storage)
			return err
		}
		for _, finalizer := range finalizers {
			if op == addFinalizers {
				controllerutil.AddFinalizer(latestStorage, finalizer)
			} else {
				controllerutil.RemoveFinalizer(latestStorage, finalizer)
			}
		}
		if !reflect.DeepEqual(storage.GetFinalizers(), latestStorage.GetFinalizers()) {
			log.Info("storage finalizer mismatches", logkeys.CurrentStorageFinalizer, storage.GetFinalizers(), logkeys.LatestStorageFinalizer, latestStorage.GetFinalizers())

			if err := r.Update(ctx, latestStorage); err != nil {
				return fmt.Errorf("UpdateFinalizers: update failed: %w", err)
			}
		} else {
			log.Info("storage finalizer does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update storage finalizers: %w", err)
	}
	return nil
}

// Get storage from K8s.
// Returns (nil, nil) if not found.
func (r *StorageReconciler) getStorage(ctx context.Context, req ctrl.Request) (*cloudv1alpha1.Storage, error) {

	storage := &cloudv1alpha1.Storage{}
	err := r.Get(ctx, req.NamespacedName, storage)
	if apiError.IsNotFound(err) || reflect.ValueOf(storage).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getStorage: %w", err)
	}
	return storage, nil

}

// Get all storage list from K8s.
func (r *StorageReconciler) getAllNamespaceStorages(ctx context.Context, req ctrl.Request) (cloudv1alpha1.StorageList, error) {
	log := log.FromContext(ctx).WithName("StorageReconciler.getAllNamespaceStorages")
	log.Info("inside the getAllNamespaceStorages function of storage operator")
	storagelist := cloudv1alpha1.StorageList{}

	var resources cloudv1alpha1.StorageList
	if err := r.Client.List(ctx, &resources, client.InNamespace(req.Namespace)); err != nil {
		return storagelist, fmt.Errorf("getStorage: %w", err)
	}
	return resources, nil
}

func (r *StorageReconciler) validateCreateRequest(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.validateCreateRequest")
	log.Info("inside the validateCreateRequest function of storage operator")
	return nil
}

// Delete storage and File system.
func (r *StorageReconciler) processDeleteStorage(ctx context.Context, storage *cloudv1alpha1.Storage, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("StorageReconciler.processDeleteStorage")
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("inside the processDeleteStorage function of storage operator")
	condition := cloudv1alpha1.StorageCondition{
		Type:               cloudv1alpha1.StorageConditionDeleting,
		Status:             k8sv1.ConditionTrue,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Message:            cloudv1alpha1.StorageMessageDeleting,
	}
	util.SetStatusCondition(&storage.Status.Conditions, condition)
	if err := r.updateStatusPhaseAndMessage(ctx, storage); err != nil {
		return ctrl.Result{}, fmt.Errorf("updateStatusPhaseAndMessage: %w", err)
	}
	if err := r.PersistStatusUpdate(ctx, storage, req, nil); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating and persist phase to deleting")
	}
	storage, err := r.getStorage(ctx, req)

	// Checking if Compute general or compute k8s flow in case it's not a failed volume
	if storage.Spec.FilesystemType == cloudv1alpha1.FilesystemTypeComputeKubernetes {
		if err := r.deleteK8sComputeResources(ctx, storage, req); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.deleteStorage(ctx, storage); err != nil {
			return ctrl.Result{}, err
		}
	}

	// remove storageFinalizer from list and update it
	log.Info("all storage resources deleted. removing finalizer.")
	finalizers := []string{StorageFinalizer}
	err = r.UpdateFinalizers(ctx, storage, req, removeFinalizers, finalizers)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Stop reconciliation as the item is good for deletion
	return ctrl.Result{}, nil
}

// Update accepted condition, phase, and message.
func (r *StorageReconciler) updateStatusAndPersist(ctx context.Context, storage *cloudv1alpha1.Storage, reconcileErr error) error {

	log := log.FromContext(ctx).WithName("StorageReconciler.updateStatusAndPersist")
	log.Info("inside the updateStatusAndPersist function of storage operator")

	// Logical Conditions, Reason and Message based on Status
	nsSuccess := util.FindStatusCondition(storage.Status.Conditions, cloudv1alpha1.StorageConditionNamespaceSuccess)
	nsCreationDone := nsSuccess != nil && nsSuccess.Status == k8sv1.ConditionTrue

	fsSuccess := util.FindStatusCondition(storage.Status.Conditions, cloudv1alpha1.StorageConditionFSSuccess)
	fSCreationDone := fsSuccess != nil && fsSuccess.Status == k8sv1.ConditionTrue

	var condStatus k8sv1.ConditionStatus
	var reason cloudv1alpha1.StorageConditionReason
	var message string

	// In this Flow File system is not created, so conditions and status are modified accordingly.

	if storage.Spec.FilesystemType == cloudv1alpha1.FilesystemTypeComputeKubernetes {
		if nsCreationDone {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonAccepted
			message = cloudv1alpha1.StorageMessageRunning
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionRunning, condStatus, reason, message); err != nil {
				return err
			}
		} else {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonNotAccepted
			message = cloudv1alpha1.StorageMessageFailed
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionFailed, condStatus, reason, message); err != nil {
				return err
			}
		}
	} else {
		if nsCreationDone && fSCreationDone {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonAccepted
			message = cloudv1alpha1.StorageMessageRunning
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionRunning, condStatus, reason, message); err != nil {
				return err
			}
		} else {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonNotAccepted
			message = cloudv1alpha1.StorageMessageFailed
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionFailed, condStatus, reason, message); err != nil {
				return err
			}
		}
	}

	if err := r.updateStatusPhaseAndMessage(ctx, storage); err != nil {
		return fmt.Errorf("updateStatusPhaseAndMessage: %w", err)
	}

	return nil
}

// Update a status condition.
func (r *StorageReconciler) updateStatusCondition(ctx context.Context, storage *cloudv1alpha1.Storage,
	storageConditionType cloudv1alpha1.StorageConditionType,
	status k8sv1.ConditionStatus, reason cloudv1alpha1.StorageConditionReason, message string,
) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.updateStatusCondition")
	log.Info("inside the updateStatusCondition function of storage operator")

	storageCondition := cloudv1alpha1.StorageCondition{
		Status:             status,
		Message:            message,
		Type:               storageConditionType,
		LastTransitionTime: metav1.Now(),
		LastProbeTime:      metav1.Now(),
		Reason:             reason,
	}
	util.SetStatusCondition(&storage.Status.Conditions, storageCondition)
	return nil
}

// Update storage status.
func (r *StorageReconciler) PersistStatusUpdate(ctx context.Context, storage *cloudv1alpha1.Storage, req ctrl.Request, reconcileErr error) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.PersistStatusUpdate")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestStorage, err := r.getStorage(ctx, req)
		// storage is deleted
		if latestStorage == nil {
			log.Info("storage not found", logkeys.Storage, storage)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the storage: %+v. error:%w", storage, err)
		}
		if !equality.Semantic.DeepEqual(storage.Status, latestStorage.Status) {
			log.Info("storage status mismatch", logkeys.CurrentStorageStatus, storage.Status, logkeys.LatestStorageStatus, latestStorage.Status)
			// update latest storage status
			storage.Status.DeepCopyInto(&latestStorage.Status)
			if err := r.Status().Update(ctx, latestStorage); err != nil {
				return fmt.Errorf("PersistStatusUpdate: %w", err)
			}
		} else {
			log.Info("storage status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update storage status: %w", err)
	}
	log.Info("END")
	return nil
}

// Set status phase and message based on conditions.
// The message will come from the condition that is most relevant.
func (r *StorageReconciler) updateStatusPhaseAndMessage(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO change the logical flow approach
	log := log.FromContext(ctx).WithName("StorageReconciler.updateStatusPhaseAndMessage")
	log.Info("BEGIN")

	runningCond := util.FindStatusCondition(storage.Status.Conditions, cloudv1alpha1.StorageConditionRunning)
	running := runningCond != nil && runningCond.Status == k8sv1.ConditionTrue
	deleting := !storage.ObjectMeta.DeletionTimestamp.IsZero()

	var phase cloudv1alpha1.FilesystemPhase
	var message string
	// Phase and Message is based on the above conditions.
	if deleting {
		// The storage and its associated resources are in the process of being deleted
		phase = cloudv1alpha1.FilesystemPhaseDeleting
		message = cloudv1alpha1.StorageMessageDeleting
	} else {
		// If StaaS is ready and running.
		if running {
			message = fmt.Sprintf(cloudv1alpha1.StorageMessageRunning)
			phase = cloudv1alpha1.FilesystemPhaseReady
		} else {
			// not running
			message = fmt.Sprintf(cloudv1alpha1.StorageMessageFailed)
			phase = cloudv1alpha1.FilesystemPhaseFailed
		}
	}

	storage.Status.Phase = phase
	storage.Status.Message = message

	log.Info("END: storage status details", logkeys.StatusPhase, storage.Status.Phase, logkeys.StatusMessage, storage.Status.Message, logkeys.StatusConditions, storage.Status.Conditions)
	return nil

}

func (r *StorageReconciler) createNamespaceIfNotExists(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.createNamespaceIfNotExists")
	log.Info("inside the createNamespaceIfNotExists function of storage operator")
	// Get Secrets from kms (Path is inside the CRD)
	log.Info("inside the createNamespaceIfNotExists function of storage operator", storage.Status.Size, storage.Spec.StorageRequest.Size)
	if storage.Status.Size != "" && storage.Status.Size != storage.Spec.StorageRequest.Size {
		return r.updateOrg(ctx, storage)
	}
	secretPath := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms with path", logkeys.SecretsPath, secretPath)
	nsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPath)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if nsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if nsCreds["username"] == "" || nsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := nsCreds["username"]
	Password := nsCreds["password"]
	nsQuery := storagecontroller.NamespaceMetadata{
		ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
		Name:      storage.Spec.ProviderSchedule.Namespace.Name,
		User:      Username,
		Password:  Password,
		UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
	}

	found, err := r.StorageControllerClient.IsNamespaceExists(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error in namespace lookup")
		return fmt.Errorf("error in namespace lookup")
	}
	if found {
		log.Info("namespace already exists, skipping creation", logkeys.Namespace, storage.Spec.ProviderSchedule.Namespace.Name)
		//extending namespace size before we move on to creating a new file system or in case of only creating a NS flow, we just extend and leave it up to csi drivers
		namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error in namespace fetching from backend")
			return fmt.Errorf("error in namespace fetching from backend")
		}
		nsProps := storagecontroller.NamespaceProperties{
			Quota: namespaceExisting.Properties.Quota,
		}

		nsObj := storagecontroller.Namespace{
			Metadata:   nsQuery,
			Properties: nsProps,
		}

		// Modify namespace request with flag (for extend or shrink), extend in case of create.

		err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, true, storage.Spec.StorageRequest.Size)
		if err != nil {
			log.Error(err, "error extending namespace capacity")
			return fmt.Errorf("error extending namespace capacity")
		}
		storage.Status.Size = storage.Spec.StorageRequest.Size
		return nil
	}

	log.Info("namespace not found, creating a new")

	nsProps := storagecontroller.NamespaceProperties{
		Quota: storage.Spec.StorageRequest.Size, //namespace quota size = spec size, and if NS exists we will extend this capacity.
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}
	err = r.StorageControllerClient.CreateNamespace(ctx, nsObj)
	if err != nil {
		log.Error(err, "error creating namespace")
		return fmt.Errorf("error creating namespace")
	}
	storage.Status.Size = storage.Spec.StorageRequest.Size
	return nil
}

func (r *StorageReconciler) createFileSystem(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.createFileSystem")
	log.Info("inside the createFileSystem function of storage operator")
	// Get Secrets from kms (Path is inside the CRD)
	secretPath := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms with path", logkeys.SecretsPath, secretPath)
	fsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPath)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if fsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if fsCreds["username"] == "" || fsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := fsCreds["username"]
	Password := fsCreds["password"]
	fsQuery := storagecontroller.FilesystemMetadata{
		NamespaceName:  storage.Spec.ProviderSchedule.Namespace.Name,
		User:           Username,
		Password:       Password,
		FileSystemName: storage.Spec.ProviderSchedule.FilesystemName,
		Encrypted:      storage.Spec.Encrypted,
		AuthRequired:   true, // TODO : Hardcoded as true, may need to provide as spec later.
		UUID:           storage.Spec.ProviderSchedule.Cluster.UUID,
	}

	fsProperties := storagecontroller.FilesystemProperties{
		FileSystemCapacity: storage.Spec.StorageRequest.Size,
	}
	fsObj := storagecontroller.Filesystem{
		Metadata:   fsQuery,
		Properties: fsProperties,
	}
	found, err := r.StorageControllerClient.IsFilesystemExists(ctx, fsQuery, false)
	if err != nil {
		log.Info("filesystem does not exist", logkeys.Error, err)
	}
	if found {
		log.Info("file system with name already exists, skipping creation", logkeys.FilesystemName, storage.Spec.ProviderSchedule.FilesystemName)
		return nil
	}

	log.Info("file system not found, so creating a new file system")
	log.Info("extending namespace size before creating a new file system")

	secretPath = storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms for namespace with path", logkeys.SecretsPath, secretPath)
	nsQuery := storagecontroller.NamespaceMetadata{
		ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
		Name:      storage.Spec.ProviderSchedule.Namespace.Name,
		User:      Username,
		Password:  Password,
		UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
	}
	namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error in namespace fetch from backend")
		return fmt.Errorf("error in namespace fetch from backend")
	}

	nsProps := storagecontroller.NamespaceProperties{
		Quota: namespaceExisting.Properties.Quota,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}

	//Create File System and add cluster addr to CRD
	fsresp := storagecontroller.Filesystem{}
	log.Info("Backend Metadata from fs Response", logkeys.Response, fsresp.Metadata.Backend)
	fsresp, err = r.StorageControllerClient.CreateFilesystem(ctx, fsObj)
	log.Info("create fS response", logkeys.Response, fsresp)

	storage.Status.Mount.ClusterAddr = fsresp.Metadata.Backend
	storage.Status.Size = storage.Spec.StorageRequest.Size

	if err != nil {
		log.Error(err, "error creating fs")
		// Removing the namespace extension size
		err := r.CleanResource(ctx, fsQuery, nsObj, storage.Spec.StorageRequest.Size)
		if err != nil {
			log.Error(err, "error cleaning resources")
			return fmt.Errorf("error cleaning resources")
		}
		storage.Status.Size = ""
		return fmt.Errorf("error creating fs and reverting namespace size")
	}
	return nil
}

// This function deletes the Storage Namespaces and FileSystems .
func (r *StorageReconciler) deleteStorage(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.deleteStorage")
	log.Info("inside the deleteStorage function of storage operator")

	secretPathNamespace := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms with path", logkeys.SecretPathNamespace, secretPathNamespace)
	nsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPathNamespace)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if nsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if nsCreds["username"] == "" || nsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := nsCreds["username"]
	Password := nsCreds["password"]
	fsQuery := storagecontroller.FilesystemMetadata{
		NamespaceName:  storage.Spec.ProviderSchedule.Namespace.Name,
		User:           Username,
		Password:       Password,
		FileSystemName: storage.Spec.ProviderSchedule.FilesystemName,
		Encrypted:      storage.Spec.Encrypted,
		AuthRequired:   true, // TODO : Hardcoded as true, may need to provide as spec later.
		UUID:           storage.Spec.ProviderSchedule.Cluster.UUID,
	}

	found, err := r.StorageControllerClient.IsFilesystemExists(ctx, fsQuery, true)
	if err != nil {
		log.Error(err, "error in filesystem lookup")
		return fmt.Errorf("error looking up File system")
	}
	if !found {
		log.Info("file system with name doesn't exists")
		return nil
	} else {
		condStatus := k8sv1.ConditionTrue
		reason := cloudv1alpha1.StorageConditionReasonNotAccepted
		message := cloudv1alpha1.StorageMessageDeleting
		if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionRunning, condStatus, reason, message); err != nil {
			log.Error(err, "error updating delete status for the filesystem resources cleaning")
			return err
		}
		err = r.StorageControllerClient.DeleteFilesystem(ctx, fsQuery)
		if err != nil {
			log.Error(err, "error deleting FS")
			return fmt.Errorf("error deleting FS")
		}
		nsQuery := storagecontroller.NamespaceMetadata{
			ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
			Name:      storage.Spec.ProviderSchedule.Namespace.Name,
			User:      Username,
			Password:  Password,
			UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
		}
		log.Info("reducing namespace size, now that file system has been removed.")
		namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error getting namespace in delete storage method")
			return fmt.Errorf("error getting namespace details")
		}
		nsProps := storagecontroller.NamespaceProperties{
			Quota: namespaceExisting.Properties.Quota,
		}

		nsObj := storagecontroller.Namespace{
			Metadata:   nsQuery,
			Properties: nsProps,
		}
		err = r.CleanResource(ctx, fsQuery, nsObj, storage.Spec.StorageRequest.Size)
		if err != nil {
			log.Error(err, "error cleaning resources")
			return fmt.Errorf("error cleaning resources")
		}
		log.Info("storage resources cleaned")
	}
	return nil
}

func (r *StorageReconciler) CleanResource(ctx context.Context, fsQuery storagecontroller.FilesystemMetadata, nsObj storagecontroller.Namespace, size string) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.CleanResource")
	nsQuery := nsObj.Metadata
	sizeToBeReduced, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		log.Error(err, "error parse spec size in k8s flow")
		return fmt.Errorf("error parse spec size in k8s flow")
	}

	namespaceAvailable, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
	log.Info("namesapce available", "namespaceAvailable object is : ", namespaceAvailable)

	if err != nil {
		log.Error(err, "error getting namespace in delete storage flow")
		return fmt.Errorf("error get namespace in in delete storage flow")
	}
	quotaAvailable, err := strconv.ParseInt(namespaceAvailable.Properties.Quota, 10, 64)
	if err != nil {
		log.Error(err, "error converting datatypes")
		return fmt.Errorf("error converting datatypes")
	}
	if quotaAvailable-sizeToBeReduced > 0 {
		log.Info("modifying namespace size by", "decrement quota : ", quotaAvailable-sizeToBeReduced)
		err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, false, size)
		if err != nil {
			log.Error(err, "error reclaiming namespace in modify namespace")
			return fmt.Errorf("error reclaiming namespace size")
		}
	}
	quotaAvailable = quotaAvailable - sizeToBeReduced
	_, fsExists, err := r.StorageControllerClient.GetAllFileSystems(ctx, fsQuery)
	if err != nil {
		log.Error(err, "error in filesystem lookup")
		return fmt.Errorf("error looking up File system")
	}
	if fsExists {
		log.Info("dont delete namespace, skipping as filesystems exist")
		return nil
	} else {
		if quotaAvailable > 1000000000 {
			log.Info("namespace allocated capacity is more than 1GB , ignoring deletion of namespace ")
			return nil
		}

		log.Info("delete namespace, now that no file system exists size < 1GB.")
		err = r.StorageControllerClient.DeleteNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error deleting namespace")
			return fmt.Errorf("error deleting namespace")
		}
		log.Info("storage resources cleaned in storage FS Flow")
	}
	log.Info("storage resources cleaned")
	return nil
}

// This function deletes the Storage Namespaces for K8sComputeResources .
func (r *StorageReconciler) deleteK8sComputeResources(ctx context.Context, storage *cloudv1alpha1.Storage, req ctrl.Request) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("deleteK8sComputeResources.Reconcile")
	log.Info("inside the delete namespace Storage function of storage operator")

	secretPathNamespace := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching Secrets from vault with path", logkeys.SecretPathNamespace, secretPathNamespace)

	nsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPathNamespace)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if nsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if nsCreds["username"] == "" || nsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := nsCreds["username"]
	Password := nsCreds["password"]
	nsQuery := storagecontroller.NamespaceMetadata{
		ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
		Name:      storage.Spec.ProviderSchedule.Namespace.Name,
		User:      Username,
		Password:  Password,
		UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
	}

	found, err := r.StorageControllerClient.IsNamespaceExists(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error in namespace lookup")
		return fmt.Errorf("error looking up namespace flow")
	}
	if !found {
		log.Info("namespace already does not exists, ignoring deletion")
		return nil
	}

	fsQuery := storagecontroller.FilesystemMetadata{
		NamespaceName: storage.Spec.ProviderSchedule.Namespace.Name,
		User:          Username,
		Password:      Password,
		Encrypted:     storage.Spec.Encrypted,
		AuthRequired:  true,
		UUID:          storage.Spec.ProviderSchedule.Cluster.UUID,
	}
	// a flag to check if any filesystem exists, to avoid repeated API calls
	isRegularFilesystemExists := false
	// get all FS from weka for the namespace.
	allFilesystems, exists, err := r.StorageControllerClient.GetAllFileSystems(ctx, fsQuery)

	if err != nil {
		log.Error(err, "error in get all filesystem lookup ")
		return fmt.Errorf("error get all File system in k8s flow")
	}
	if exists {
		// For all the FS with a given prefix delete the fs that start with the prefix and reclaim the space.
		totalDeletedQuota := int64(0)
		metadata := storagecontroller.FilesystemMetadata{
			NamespaceName:  storage.Spec.ProviderSchedule.Namespace.Name,
			User:           Username,
			Password:       Password,
			Encrypted:      true,
			FileSystemName: "",
			AuthRequired:   true,
			UUID:           storage.Spec.ProviderSchedule.Cluster.UUID,
		}
		for _, filesystem := range allFilesystems {
			if strings.HasPrefix(filesystem.Metadata.FileSystemName, storage.Spec.Prefix) {
				metadata.Encrypted = filesystem.Metadata.Encrypted
				metadata.FileSystemName = filesystem.Metadata.FileSystemName
				log.Info("filesystem metadata", logkeys.FilesystemMetadata, metadata)
				err = r.StorageControllerClient.DeleteFilesystem(ctx, metadata)
				if err != nil {
					log.Error(err, "error deleting fs")
					return fmt.Errorf("error deleting fs")
				}
				log.Info("filesystem capacity", logkeys.FilesystemCapacity, filesystem.Properties.FileSystemCapacity)
				capacity, err := strconv.ParseInt(filesystem.Properties.FileSystemCapacity, 10, 64)
				if err != nil {
					log.Error(err, "cannot parse capacity string")
					return fmt.Errorf("cannot parse capacity string in the k8s compute flow")
				}
				totalDeletedQuota = totalDeletedQuota + capacity
			} else {
				isRegularFilesystemExists = true
			}
		}
		log.Info("reclaiming namespace size, now that file systems have been removed.", logkeys.TotalDeletedQuota, totalDeletedQuota)
	}

	namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error getting Namespace from backend")
		return fmt.Errorf("error getting Namespace from backend")
	}
	quotaAvailable, err := strconv.ParseInt(namespaceExisting.Properties.Quota, 10, 64)
	if err != nil {
		log.Error(err, "error converting datatypes in k8s flow")
		return fmt.Errorf("error converting datatypes in k8s flow")
	}

	nsProps := storagecontroller.NamespaceProperties{
		Quota: namespaceExisting.Properties.Quota,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}

	specSize, err := strconv.ParseInt(storage.Spec.StorageRequest.Size, 10, 64)
	if err != nil {
		log.Error(err, "error parse spec size in k8s flow")
		return fmt.Errorf("error parse spec size in k8s flow")
	}
	quotaAvailable = quotaAvailable - specSize
	// Delete Namspace finally if no filesystems exist in the namespace at all (either namespace k8s compute flow or regular flow)
	if !isRegularFilesystemExists && quotaAvailable < 1000000000 {
		log.Info("namespace allocated capacity is less than 1GB , deleting the namespace org")
		err = r.StorageControllerClient.DeleteNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error deleting namespace in k8s compute as no filesystems exist")
			return fmt.Errorf("error deleting namespace in k8s compute as no filesystems exist")
		}
		return nil
	} else {
		err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, false, storage.Spec.StorageRequest.Size)
		if err != nil {
			log.Error(err, "error in reclaim namespace size in kubernetes compute flow")
			return fmt.Errorf("error in reclaim namespace size in k8s flow")
		}
	}
	return nil
}

// helper function to read secrets using storage kms
func (r *StorageReconciler) readSecretsFromStorageKMS(ctx context.Context, secretKeyPath string) (map[string]string, error) {
	logger := log.FromContext(ctx).WithName("StorageReconciler.readSecretsFromStorageKMS")
	logger.Info("calling storage kms service for get", logkeys.SecretsPath, secretKeyPath)

	request := pb.GetSecretRequest{
		KeyPath: secretKeyPath,
	}
	tries := 3
	for tries > 0 {
		secretResp, err := r.kmsClient.Get(ctx, &request)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				logger.Error(err, "ContextDeadline Exceeded or cancelled: true")
				tries -= 1
				time.Sleep(2000 * time.Millisecond)
				continue
			} else {
				break
			}
		}
		return secretResp.Secrets, nil
	}
	return nil, fmt.Errorf("error reading secrets from storage kms")
}

func (r *StorageReconciler) updateFileSystem(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.updateFileSystem")
	log.Info("inside the update filesystem function of storage operator")
	// Get Secrets from kms (Path is inside the CRD)
	secretPath := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms with path", logkeys.SecretsPath, secretPath)
	fsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPath)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if fsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if fsCreds["username"] == "" || fsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := fsCreds["username"]
	Password := fsCreds["password"]
	fsQuery := storagecontroller.FilesystemMetadata{
		NamespaceName:  storage.Spec.ProviderSchedule.Namespace.Name,
		User:           Username,
		Password:       Password,
		FileSystemName: storage.Spec.ProviderSchedule.FilesystemName,
		Encrypted:      storage.Spec.Encrypted,
		AuthRequired:   true,
		UUID:           storage.Spec.ProviderSchedule.Cluster.UUID,
	}

	fsProperties := storagecontroller.FilesystemProperties{
		FileSystemCapacity: storage.Spec.StorageRequest.Size,
	}
	fsObj := storagecontroller.Filesystem{
		Metadata:   fsQuery,
		Properties: fsProperties,
	}
	isExists, err := r.StorageControllerClient.IsFilesystemExists(ctx, fsQuery, false)
	if err != nil {
		log.Error(err, "filesystem does not exist")
		return err
	}
	if !isExists {
		log.Info("filesystem does not exist")
		return fmt.Errorf("error finding filesystem")
	}

	secretPath = storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms for namespace with path", logkeys.SecretsPath, secretPath)
	nsQuery := storagecontroller.NamespaceMetadata{
		ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
		Name:      storage.Spec.ProviderSchedule.Namespace.Name,
		User:      Username,
		Password:  Password,
		UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
	}
	namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error getting ns")
		return fmt.Errorf("error getting namespace details")
	}
	nsProps := storagecontroller.NamespaceProperties{
		Quota: namespaceExisting.Properties.Quota,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}

	specSize, err := strconv.ParseUint(storage.Spec.StorageRequest.Size, 10, 64)
	if err != nil {
		log.Error(err, "error parse spec size")
		return fmt.Errorf("error parse spec size")
	}
	statusSize, err := strconv.ParseUint(storage.Status.Size, 10, 64)
	if err != nil {
		log.Error(err, "error parse status size")
		return fmt.Errorf("error parse status size")
	}

	namespaceIncreaseFlag := true
	netSize := specSize - statusSize
	if statusSize > specSize {
		namespaceIncreaseFlag = false
		netSize = statusSize - specSize
	}
	log.Info("modifying namespace by netSize", logkeys.NetSize, netSize)

	// We need to know if the update is an increase or decrease, so based on that we need to expand or shrink
	// If its an increase first, we expand namespace and increase filesystem size, else we first decrease size and reduce namespace size.
	if namespaceIncreaseFlag {
		err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, namespaceIncreaseFlag, strconv.FormatUint(netSize, 10))
		if err != nil {
			log.Error(err, "error expanding namespace")
			return fmt.Errorf("error expanding namespace size")
		}
	}
	fsresp := storagecontroller.Filesystem{}
	fsresp, err = r.StorageControllerClient.UpdateFilesystem(ctx, fsObj)
	log.Info("update filesystem response", logkeys.Response, fsresp)

	if !namespaceIncreaseFlag {
		err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, namespaceIncreaseFlag, strconv.FormatUint(netSize, 10))
		if err != nil {
			log.Error(err, "error shrink ns")
			return fmt.Errorf("error shrinking namespace size")
		}
	}
	storage.Status.Mount.ClusterAddr = fsresp.Metadata.Backend
	storage.Status.Size = storage.Spec.StorageRequest.Size

	if err != nil {
		log.Error(err, "error updating fs")
		// Removing the namespace extension size
		error := r.StorageControllerClient.ModifyNamespace(ctx, nsObj, namespaceIncreaseFlag, strconv.FormatUint(netSize, 10))
		if error != nil {
			log.Error(err, "error modifying fs")
			return fmt.Errorf("error modify namespace size")
		}
		return fmt.Errorf("error updating fs and reverting namespace size")
	}
	return nil
}

func (r *StorageReconciler) updateOrg(ctx context.Context, storage *cloudv1alpha1.Storage) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.updateOrg")
	log.Info("inside the update org function of storage operator")
	secretPath := storage.Spec.ProviderSchedule.Namespace.CredentialsPath
	log.Info("fetching secrets from storage kms with path", logkeys.SecretsPath, secretPath)
	nsCreds, err := r.readSecretsFromStorageKMS(ctx, secretPath)
	if err != nil {
		return status.Error(codes.Internal, "error reading from storage kms")
	}
	if nsCreds == nil {
		log.Error(err, "storageKms returned nil for secret")
		return fmt.Errorf("storageKms returned nil for secret")
	}
	if nsCreds["username"] == "" || nsCreds["password"] == "" {
		log.Error(err, "credentials cannot be empty")
		return fmt.Errorf("empty credentials in secret")
	}
	Username := nsCreds["username"]
	Password := nsCreds["password"]

	nsQuery := storagecontroller.NamespaceMetadata{
		ClusterId: storage.Spec.ProviderSchedule.Cluster.Name,
		Name:      storage.Spec.ProviderSchedule.Namespace.Name,
		User:      Username,
		Password:  Password,
		UUID:      storage.Spec.ProviderSchedule.Cluster.UUID,
	}
	namespaceExisting, err := r.StorageControllerClient.GetNamespace(ctx, nsQuery)
	if err != nil {
		log.Error(err, "error getting ns")
		return fmt.Errorf("error getting namespace details")
	}
	nsProps := storagecontroller.NamespaceProperties{
		Quota: namespaceExisting.Properties.Quota,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}

	specSize, err := strconv.ParseUint(storage.Spec.StorageRequest.Size, 10, 64)
	if err != nil {
		log.Error(err, "error parse spec size")
		return fmt.Errorf("error parse spec size")
	}
	statusSize, err := strconv.ParseUint(storage.Status.Size, 10, 64)
	if err != nil {
		log.Error(err, "error parse status size")
		return fmt.Errorf("error parse status size")
	}
	netSize := specSize - statusSize

	log.Info("modifying namespace org by netSize", logkeys.NetSize, netSize)

	err = r.StorageControllerClient.ModifyNamespace(ctx, nsObj, true, strconv.FormatUint(netSize, 10))
	if err != nil {
		log.Error(err, "error expanding org size")
		return fmt.Errorf("error expanding namespace org size")
	}
	storage.Status.Size = storage.Spec.StorageRequest.Size
	return nil
}

func shouldRequeueDeleting(ctx context.Context, storage *cloudv1alpha1.Storage) bool {
	log := log.FromContext(ctx).WithName("shouldRequeueDeleting")
	log.Info("inside the shouldRequeueDeleting function of storage operator")
	emptyStatus := cloudv1alpha1.StorageStatus{}
	if !reflect.DeepEqual(storage.Status, emptyStatus) {
		if storage.Status.Phase == cloudv1alpha1.FilesystemPhaseDeleting {
			log.Info("requeing reconcile for pending deleting storage ")
			return true
		} else {
			return false
		}
	}
	return false
}
