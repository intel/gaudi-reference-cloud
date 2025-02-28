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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	storageControllerNs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	storageControllerVastApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	idcclientset "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcscheme "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned/scheme"
	idcinformer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions/private.cloud/v1alpha1"
	v1alpha1informers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/vast_storage_operator/util"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Finalizers
const (
	VastStorageFinalizer                = "private.cloud.intel.com/vaststoragefinalizer"
	VastStorageMeteringMonitorFinalizer = "private.cloud.intel.com/vaststoragemeteringmonitorfinalizer"
	addFinalizers                       = "add_finalizers"
	removeFinalizers                    = "remove_finalizers"
)

// StorageReconciler reconciles a Storage object
type StorageReconciler struct {
	// clientset for custom resource
	VastStoragesClientSet idcclientset.Interface
	// cache has synced
	VastStorageSynced cache.InformerSynced
	// lister : will be used to get and list objects avoiding calling api server directly
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	Workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	// configValues defines the config values passed to the the controller
	StorageControllerClient *storagecontroller.StorageControllerClient
	kmsClient               pb.StorageKMSPrivateServiceClient
	vastStorageInformer     v1alpha1informers.VastStorageInformer
	inf                     idcinformerfactory.SharedInformerFactory

	informerFactory idcinformer.VastStorageInformer
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=vaststorages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=vaststorages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=events,verbs=create;get;list;update;patch;delete
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=vaststorages/finalizers,verbs=update
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=namespaces,verbs=create;get;list;update
func NewStorageOperator(ctx context.Context, kubeClient kubernetes.Interface, idcClientSet idcclientset.Interface, informerFactory idcinformer.VastStorageInformer, inf idcinformerfactory.SharedInformerFactory, strClient *storagecontroller.StorageControllerClient, kms pb.StorageKMSPrivateServiceClient, mgr ctrl.Manager) (*StorageReconciler, error) {
	utilruntime.Must(idcscheme.AddToScheme(scheme.Scheme))
	log := log.FromContext(ctx).WithName("vast operator init")
	log.Info("initializing vast storage operator")

	vastStorageInformer := informerFactory

	r := &StorageReconciler{
		VastStoragesClientSet:   idcClientSet,
		VastStorageSynced:       vastStorageInformer.Informer().HasSynced,
		StorageControllerClient: strClient,
		kmsClient:               kms,
		vastStorageInformer:     vastStorageInformer,
		informerFactory:         informerFactory,
		inf:                     inf,
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("vast-storage-controller"),
	}

	_, err := vastStorageInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: r.handleAdd,
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Info("inside update event handler")
			oldResource, okOld := oldObj.(*cloudv1alpha1.VastStorage)
			newResource, okNew := newObj.(*cloudv1alpha1.VastStorage)
			if !okOld || !okNew {
				return
			}
			if !reflect.DeepEqual(oldResource.Spec, newResource.Spec) {
				r.handleUpdate(oldObj, newObj)
			} else if newResource.GetDeletionTimestamp() != nil {
				log.Info("Handle the resource as if it is being deleted")
				// Handle the resource as if it is being deleted
				err := r.reconcileDelete(ctx, newResource)
				if err != nil {
					log.Error(err, "error in reconcile Delete flow")
					return
				}
				return
			}
		},
		DeleteFunc: r.handleDelete,
	})
	if err != nil {
		log.Error(err, "error adding event handler")
	}

	return r, nil
}

func (r *StorageReconciler) Run(ctx context.Context, stopCh <-chan struct{}) error {
	log := log.FromContext(ctx).WithName("vastcontroller.Run").
		WithValues(logkeys.Controller, "vast operator", logkeys.ControllerGroup, "private.cloud.intel.com", logkeys.ControllerKind, "VastStorages")

	// Start the informer factories to begin populating the informer caches
	r.inf.Start(stopCh)

	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, r.VastStorageSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Controller is ready to process events")

	<-stopCh
	log.Info("Controller stopped")

	return nil
}

func (r *StorageReconciler) handleAdd(obj interface{}) {
	storage := obj.(*cloudv1alpha1.VastStorage)
	ctx := context.TODO()
	log := log.FromContext(ctx).WithName("handleAdd")
	log.Info("vast storage added", "name", storage.Name)
	var finalizers []string
	// Add Finalizer (if not present) for the deletion cleanup.
	// This only updates the in-memory Storage.
	finalizers = append(finalizers, VastStorageFinalizer)

	// Add Storage metering monitor Finalizer which can be removed by compute metering monitor only
	finalizers = append(finalizers, VastStorageMeteringMonitorFinalizer)
	storage, err := r.UpdateFinalizers(ctx, storage, addFinalizers, finalizers)
	if err != nil {
		log.Error(err, "failed to append and update vast finalizers", logkeys.Storage, storage)
	}
	err = r.reconcileCreate(ctx, storage)
	if err != nil {
		log.Error(err, "failed to create vast storage :", logkeys.Storage, storage)
	}
}

func (r *StorageReconciler) handleUpdate(oldObj, newObj interface{}) {
	storageNew := newObj.(*cloudv1alpha1.VastStorage)
	storageOld := oldObj.(*cloudv1alpha1.VastStorage)

	ctx := context.TODO()
	log := log.FromContext(ctx).WithName("handleUpdate")
	if storageNew.Spec.StorageRequest.Size != "" && storageOld.Spec.StorageRequest.Size != "" && storageOld.Spec.StorageRequest.Size != storageNew.Spec.StorageRequest.Size {
		extraSize, err := calculateExtraSize(storageNew, storageOld)
		if err != nil {
			log.Error(err, "error in converting string to int")
			return
		}
		log.Info("handling size update for : ", "name", storageNew.Name)
		err = r.processVastStorageFileUpdate(ctx, storageNew, strconv.Itoa(extraSize))
		if err != nil {
			log.Error(err, "error in processing vast file update")
			return
		}
	}
	if storageNew.Spec.Networks.SecurityGroups.IPFilters != nil && storageOld.Spec.Networks.SecurityGroups.IPFilters != nil && !reflect.DeepEqual(storageOld.Spec.Networks.SecurityGroups.IPFilters, storageNew.Spec.Networks.SecurityGroups.IPFilters) {
		log.Info("handling IP filters update for : ", "name", storageNew.Name)
		err := r.reconcileIPFilterUpdate(ctx, storageNew)
		if err != nil {
			log.Error(err, "error in updating IP filters")
		}
	}
	r.reconcileUpdate(ctx, storageNew)
}

func calculateExtraSize(storageNew, storageOld *cloudv1alpha1.VastStorage) (int, error) {
	newSize, err := strconv.Atoi(storageNew.Spec.StorageRequest.Size)
	if err != nil {
		return 0, fmt.Errorf("failed to convert new storage size to int: %w", err)
	}

	oldSize, err := strconv.Atoi(storageOld.Spec.StorageRequest.Size)
	if err != nil {
		return 0, fmt.Errorf("failed to convert old storage size to int: %w", err)
	}

	extraSize := newSize - oldSize
	return extraSize, nil
}

func (r *StorageReconciler) handleDelete(obj interface{}) {
	storage := obj.(*cloudv1alpha1.VastStorage)
	ctx := context.TODO()
	log := log.FromContext(ctx).WithName("handleDelete")
	log.Info("vast storage delete event received, ignoring", "name", storage.Name)
	//r.reconcileDelete(ctx, storage)
}

// Add/remove vast finalizers and persist for vast resources
func (r *StorageReconciler) UpdateFinalizers(ctx context.Context, storage *cloudv1alpha1.VastStorage, op string, finalizers []string) (*cloudv1alpha1.VastStorage, error) {
	log := log.FromContext(ctx).WithName("StorageReconciler.UpdateFinalizers")
	log.Info("BEGIN", logkeys.Task, op)
	defer log.Info("END", logkeys.Task, op)

	var updatedStorage *cloudv1alpha1.VastStorage
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest version of the storage resource
		latestStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).Get(ctx, storage.Name, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "failed to get the storage", logkeys.Storage, storage)
			return err
		}

		// Add or remove finalizers
		for _, finalizer := range finalizers {
			if op == addFinalizers {
				controllerutil.AddFinalizer(latestStorage, finalizer)
			} else {
				controllerutil.RemoveFinalizer(latestStorage, finalizer)
			}
		}

		// Check if finalizers have changed
		if !reflect.DeepEqual(storage.GetFinalizers(), latestStorage.GetFinalizers()) {
			log.Info("storage finalizer mismatches", logkeys.CurrentStorageFinalizer, storage.GetFinalizers(), logkeys.LatestStorageFinalizer, latestStorage.GetFinalizers())

			// Update the storage resource with the new finalizers
			updatedStorage, err = r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).Update(ctx, latestStorage, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("UpdateFinalizers: update failed: %w", err)
			}
		} else {
			log.Info("storage finalizer does not need to be changed")
			updatedStorage = latestStorage
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update storage finalizers: %w", err)
	}
	return updatedStorage, nil
}

func (r *StorageReconciler) reconcileCreate(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VastStorage.ReconcileCreate").WithValues(logkeys.ResourceId, storage.Name).Start()
	defer span.End()
	log.Info("inside the reconciling function for create event of vast storage operator")
	// If resource exists, check if it has been fully processed
	if storage.Annotations["processed"] == "true" {
		log.Info("Resource already exists and has been processed, skipping creation")
		return nil
	}

	// Proceed with the creation logic
	err := r.processVastStorageCreateUpdate(ctx, storage)
	if err != nil {
		log.Error(err, "error in processing storage create event")
		return err
	}
	existingStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).Get(ctx, storage.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get latest storage from clientset")
		return err
	}
	// Mark the resource as processed
	if existingStorage.Annotations == nil {
		existingStorage.Annotations = make(map[string]string)
	}
	existingStorage.Annotations["processed"] = "true"
	log.Info("Setting processed annotation", "annotations", existingStorage.Annotations)

	_, err = r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(existingStorage.Namespace).Update(ctx, existingStorage, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "error in updating storage resource with processed annotation")
		return err
	}

	return nil
}

func (r *StorageReconciler) reconcileUpdate(ctx context.Context, storage *cloudv1alpha1.VastStorage) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VastStorage.ReconcileUpdate").WithValues(logkeys.ResourceId, storage.Name).Start()
	defer span.End()
	log.Info("inside the reconciling function for update event of vast storage operator")

	// Check if the deletion timestamp is set
	if storage.GetDeletionTimestamp() != nil {
		// Handle the resource as if it is being deleted
		log.Info("handle delete event of vast storage operator as deletion timestamp is now set")
		err := r.reconcileDelete(ctx, storage)
		if err != nil {
			log.Error(err, "failed to perform reconcile delete operation : ", logkeys.Storage, storage)
		}
		return
	}

	// TODO : Implement the logic to handle update event
	// Update the IP Filters if there's an update in the spec or update the size if there's an update in the spec.
}

func (r *StorageReconciler) reconcileIPFilterUpdate(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {

	log := log.FromContext(ctx).WithName("StorageReconciler.reconcileIPFilterUpdate")
	log.Info("inside the update event for ip filter update", "new ip filters are", storage.Spec.Networks.SecurityGroups.IPFilters)

	// Extract all IP filters
	var ipFilters []storagecontroller.IPFilter
	for _, ipFilter := range storage.Spec.Networks.SecurityGroups.IPFilters {
		ipFilters = append(ipFilters, storagecontroller.IPFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		})
	}

	nsQuery := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
	}

	found, err := r.StorageControllerClient.IsNamespaceExists(ctx, nsQuery)

	if err != nil {
		log.Error(err, "error in namespace lookup")
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : isNamespaceExists, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		return fmt.Errorf("error in namespace lookup")
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : is namespaceExists, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	if found {
		log.Info("namespace already exists and hence updating IP Filters", logkeys.Namespace, storage.Spec.ClusterAssignment.NamespaceName)
		//extending namespace size before we move on to extend file size
		namespaceExisting, err := r.StorageControllerClient.GetVastNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error in namespace fetching from backend")
			return fmt.Errorf("error in namespace fetching from backend")
		}
		nsProps := storagecontroller.NamespaceProperties{
			Quota:     namespaceExisting.Properties.Quota,
			IPFilters: ipFilters,
		}

		nsObj := storagecontroller.Namespace{
			Metadata:   nsQuery,
			Properties: nsProps,
		}

		// Modify namespace request with flag (for extend or shrink), extend in case of create.
		log.Info("updating namespace IP Filters", "ipFilters", ipFilters, "nsObject is ", nsObj)
		err = r.StorageControllerClient.UpdateVastIPFilyers(ctx, nsObj)
		if err != nil {
			log.Error(err, "error updating namespace IP Filters")
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : :updating namespace IP Filters:  %s ", storage.Name),
				fmt.Sprintf("request failed "))
			return fmt.Errorf("error updating namespace IP Filters")
		}
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : updating namespace IP Filters  %s ", storage.Name),
			fmt.Sprintf("request succeeded "))
		return nil
	} else {
		log.Info("namespace not found, ignoring updates", logkeys.Namespace, storage.Spec.ClusterAssignment.NamespaceName)
	}
	return nil
}

func (r *StorageReconciler) processVastStorageFileUpdate(ctx context.Context, storage *cloudv1alpha1.VastStorage, extraSize string) error {

	log := log.FromContext(ctx).WithName("StorageReconciler.processVastStorageFileUpdate")
	log.Info("inside the update event for vast storage size", "new size is", extraSize)
	nsQuery := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
		Id:   strconv.FormatInt(storage.Status.VolumeProps.NamespaceId, 10),
	}
	found, err := r.StorageControllerClient.IsNamespaceExists(ctx, nsQuery)

	if err != nil {
		log.Error(err, "error in namespace lookup")
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : isNamespaceExists, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		return fmt.Errorf("error in namespace lookup")
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : is namespaceExists, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	if found {
		log.Info("namespace already exists and hence extending size", logkeys.Namespace, storage.Spec.ClusterAssignment.NamespaceName)
		//extending namespace size before we move on to extend file size
		namespaceExisting, err := r.StorageControllerClient.GetVastNamespace(ctx, nsQuery)
		if err != nil {
			log.Error(err, "error in namespace fetching from backend")
			return fmt.Errorf("error in namespace fetching from backend")
		}
		nsProps := storagecontroller.NamespaceProperties{
			Quota:     namespaceExisting.Properties.Quota,
			IPFilters: namespaceExisting.Properties.IPFilters,
		}

		nsObj := storagecontroller.Namespace{
			Metadata:   nsQuery,
			Properties: nsProps,
		}

		// Modify namespace request with flag (for extend or shrink), extend in case of create.

		err = r.StorageControllerClient.ModifyVastNamespace(ctx, nsObj, true, extraSize)
		if err != nil {
			log.Error(err, "error extending namespace capacity")
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : updateNamespace (expand), filesystem:  %s ", storage.Name),
				fmt.Sprintf("request failed "))
			return fmt.Errorf("error extending namespace capacity")
		}
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : updateNamespace, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request succeeded "))

		// update filesystem capacity
		sizes, err := strconv.ParseUint(storage.Spec.StorageRequest.Size, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse size: %w", err)
		}
		params := &storagecontroller.UpdateFilesystemParams{
			NamespaceID:   strconv.FormatInt(storage.Status.VolumeProps.NamespaceId, 10),
			FilesystemID:  strconv.FormatInt(storage.Status.VolumeProps.FilesystemId, 10),
			NewTotalBytes: sizes,
			ClusterID:     storage.Spec.ClusterAssignment.ClusterUUID,
		}
		_, err = r.StorageControllerClient.UpdateVastFilesystem(ctx, params)
		if err != nil {
			log.Error(err, "failed to update vast filesystem")

			// revert back the size of namespace which was extended.
			modifyErr := r.StorageControllerClient.ModifyVastNamespace(ctx, nsObj, false, extraSize)

			if modifyErr != nil {
				log.Error(modifyErr, "error shrinking namespace capacity due to failure in updating filesystem")
				r.Recorder.Event(storage,
					k8sv1.EventTypeNormal,
					fmt.Sprintf("SDS Controller API : : Modify event updateNamespace (shrink), filesystem:  %s ", storage.Name),
					fmt.Sprintf("request failed "))
				return fmt.Errorf("error shrinking namespace capacity")
			}
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : updateNamespace (shrink post fs extension), filesystem:  %s ", storage.Name),
				fmt.Sprintf("request succeeded "))
			return err

		}
		log.Info("successfully update of vast namespace and filesystem")

	} else {
		log.Info("namespace not found, ignoring updates to filesystem", logkeys.Namespace, storage.Spec.ClusterAssignment.NamespaceName)
	}
	return nil
}

func hasVastStorageFinalizer(storage *cloudv1alpha1.VastStorage) bool {
	for _, finalizer := range storage.GetFinalizers() {
		if finalizer == VastStorageFinalizer {
			return true
		}
	}
	return false
}

func (r *StorageReconciler) reconcileDelete(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VastStorage.ReconcileDelete").WithValues(logkeys.ResourceId, storage.Name).Start()
	defer span.End()
	log.Info("inside the reconciling function for delete event of vast storage operator")
	// check if there's any finalizers.
	finalizerExists := hasVastStorageFinalizer(storage)
	if !finalizerExists {
		log.Info("no vast finalizer found. skipping deletion since object is already cleaned up")
		return nil
	}
	err := r.deleteStorage(ctx, storage)
	if err != nil {
		log.Error(err, "error in processing storage delete event")
		return err

	}
	// Remove storageFinalizer from list and update it
	log.Info("all vast storage resources deleted. removing vast finalizer.")
	finalizers := []string{VastStorageFinalizer}
	_, err = r.UpdateFinalizers(ctx, storage, removeFinalizers, finalizers)
	if err != nil {
		log.Error(err, "failed to remove vast finalizer from storage resource")
		return err
	}
	log.Info("successfully removed vast finalizer from vast storage resource")
	return nil
}

func (r *StorageReconciler) processVastStorageCreateUpdate(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.processVastStorageCreateUpdate")
	log.Info("inside the processVastStorageCreateUpdate function of vast storage operator")
	// Set initial phase for replicator : Provisioning
	storage.Status.Phase = cloudv1alpha1.FilesystemPhaseProvisioning
	storage.Status.Message = cloudv1alpha1.StorageMessageProvisioningAccepted
	newStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update storage status")
		return err
	}
	storage = newStorage
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
		log.Info("Updating status in create ns step", "status", storage.Status)
		newStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
		if err != nil {
			log.Error(err, "failed to update storage status")
		}
		log.Info("Updated status", "status", newStorage.Status)
		if err := r.updateStatusAndPersist(ctx, newStorage); err != nil {
			log.Error(err, "failed to persist storage status")
		}
		return err
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
			Message:            cloudv1alpha1.VastStorageMessageRunning,
		}
		util.SetStatusCondition(&storage.Status.Conditions, condition)
		// Make the API call to update the status
		newStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
		if err != nil {
			log.Error(err, "failed to update storage status")
		}
		if err := r.updateStatusAndPersist(ctx, newStorage); err != nil {
			log.Error(err, "failed to persist storage status")
		}

		return nil
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
		newStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
		if err != nil {
			log.Error(err, "failed to update storage status")
		}

		log.Info("Updated status", "status", newStorage.Status)
		if err := r.updateStatusAndPersist(ctx, newStorage); err != nil {
			log.Error(err, "failed to persist storage status")
		}
		return err
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
		Message:            cloudv1alpha1.VastStorageMessageRunning,
	}
	util.SetStatusCondition(&storage.Status.Conditions, condition)
	// Persist the changes
	log.Info("Updating status", "status", storage.Status)

	// Fetch the latest storage object to prevent race conditions and update the status
	latestStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).Get(ctx, storage.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "failed to get the latest storage before update of phase and status", logkeys.Storage, storage)
		return err
	}
	//Merge the changes
	latestStorage.Status = storage.Status
	// Make the API call to update the status

	newStorage, err = r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(latestStorage.Namespace).UpdateStatus(ctx, latestStorage, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update storage status after filesystem creation")
	}
	log.Info("Updated status after filesystem creation", "status", newStorage.Status)
	if err := r.updateStatusAndPersist(ctx, newStorage); err != nil {
		log.Error(err, "failed to persist storage status")
	}
	return nil
}

func (r *StorageReconciler) createNamespaceIfNotExists(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.createNamespaceIfNotExists")
	log.Info("inside the createNamespaceIfNotExists function of vast storage operator", storage.Spec.StorageRequest.Size, storage.Spec.StorageRequest.Size)
	nsQuery := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
	}
	namespaceExisting := storageControllerNs.Namespace{}
	namespaceExisting, err := r.StorageControllerClient.GetVastNamespace(ctx, nsQuery)

	if err != nil {
		log.Error(err, "error in namespace lookup")
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : isNamespaceExists, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		return fmt.Errorf("error in namespace lookup")
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : is namespaceExists, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	if namespaceExisting.Metadata != (storageControllerNs.NamespaceMetadata{}) && namespaceExisting.Properties.Quota != "" {
		log.Info("namespace already exists, skipping creation", logkeys.Namespace, storage.Spec.ClusterAssignment.NamespaceName)
		//extending namespace size before we move on to creating a new file system or in case of only creating a NS flow, we just extend and leave it up to csi drivers
		nsProps := storagecontroller.NamespaceProperties{
			Quota:     namespaceExisting.Properties.Quota,
			IPFilters: namespaceExisting.Properties.IPFilters,
		}
		nsMetadata := storagecontroller.NamespaceMetadata{
			Name: storage.Spec.ClusterAssignment.NamespaceName,
			UUID: storage.Spec.ClusterAssignment.ClusterUUID,
			Id:   namespaceExisting.Metadata.Id,
		}
		nsObj := storagecontroller.Namespace{
			Metadata:   nsMetadata,
			Properties: nsProps,
		}

		// Modify namespace request with flag (for extend or shrink), extend in case of create.

		err = r.StorageControllerClient.ModifyVastNamespace(ctx, nsObj, true, storage.Spec.StorageRequest.Size)
		if err != nil {
			log.Error(err, "error extending namespace capacity")
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : updateNamespace (expand), filesystem:  %s ", storage.Name),
				fmt.Sprintf("request failed "))
			return fmt.Errorf("error extending namespace capacity")
		}
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : updateNamespace, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request succeeded "))
		return nil
	}

	log.Info("namespace not found, creating a new")
	// Extract all IP filters
	log.Info("IP filters", "ipFilters", storage.Spec.Networks.SecurityGroups.IPFilters)
	var ipFilters []storagecontroller.IPFilter
	for _, ipFilter := range storage.Spec.Networks.SecurityGroups.IPFilters {
		ipFilters = append(ipFilters, storagecontroller.IPFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		})
	}
	nsProps := storagecontroller.NamespaceProperties{
		Quota:     storage.Spec.StorageRequest.Size, //namespace quota size = spec size, and if NS exists we will extend this capacity.
		IPFilters: ipFilters,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}
	err = r.StorageControllerClient.CreateNamespace(ctx, nsObj)
	if err != nil {
		log.Error(err, "error creating namespace")
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : createNamespace, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		return fmt.Errorf("error creating namespace")
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : createNamespace, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	return nil
}

func (r *StorageReconciler) createFileSystem(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.createFileSystem")
	log.Info("inside the createFileSystem function of vast storage operator")
	nsQuery := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
	}
	namespaceExisting, err := r.StorageControllerClient.GetVastNamespace(ctx, nsQuery)
	nsMetadata := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
		Id:   namespaceExisting.Metadata.Id,
	}

	if err != nil {
		log.Error(err, "error in namespace fetch from backend")
		return fmt.Errorf("error in namespace fetch from backend")
	}
	nsProps := storagecontroller.NamespaceProperties{
		Quota:     namespaceExisting.Properties.Quota,
		IPFilters: namespaceExisting.Properties.IPFilters,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsMetadata,
		Properties: nsProps,
	}
	sizeUint64, err := strconv.ParseUint(storage.Spec.StorageRequest.Size, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse size: %w", err)
	}
	// Create the filesystem
	// Log the values of storage.Namespace and storage.Spec.FilesystemName
	log.Info("Namespace and FilesystemName", "Namespace", storage.Namespace, "FilesystemName", storage.Spec.FilesystemName)

	// Concatenate the Namespace and FilesystemName with a hyphen
	filesystemName := storage.Namespace + "-" + storage.Spec.FilesystemName

	// Log the concatenated filesystem name
	log.Info("Concatenated Filesystem Name", "FilesystemName", filesystemName)

	createParams := &storagecontroller.CreateFilesystemParams{
		NamespaceID: namespaceExisting.Metadata.Id,
		Name:        filesystemName, // to make sure unique name across a vast cluster
		Path:        storage.Spec.MountConfig.VolumePath,
		TotalBytes:  sizeUint64,
		Protocols: []storageControllerVastApi.Filesystem_Protocol{
			convertToFilesystemProtocol(storage.Spec.MountConfig.MountProtocol),
		},
		ClusterID: storage.Spec.ClusterAssignment.ClusterUUID,
	}
	log.Info("creating filesystem", "params", createParams)
	fsresp, err := r.StorageControllerClient.CreateVastFilesystem(ctx, createParams)
	if err != nil {
		log.Error(err, "error creating filesystem in sds controller")
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : createVastFilesystem, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		errModify := r.StorageControllerClient.ModifyVastNamespace(ctx, nsObj, false, storage.Spec.StorageRequest.Size)
		if errModify != nil {
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : updateNamespace (shrink), filesystem:  %s ", storage.Name),
				fmt.Sprintf("request failed "))
			log.Error(errModify, "error shrinking namespace capacity")
			return fmt.Errorf("error shrinking namespace capacity")
		}

		quota, err := strconv.ParseInt(namespaceExisting.Properties.Quota, 10, 64)
		if err != nil {
			return err
		}
		fsSize, err := strconv.ParseInt(storage.Spec.StorageRequest.Size, 10, 64)
		if err != nil {
			return err
		}
		if (quota - fsSize) == 0 {
			err = r.StorageControllerClient.DeleteNamespace(ctx, nsMetadata)
			if err != nil {
				r.Recorder.Event(storage,
					k8sv1.EventTypeNormal,
					fmt.Sprintf("SDS Controller API : : deleteNamespace, filesystem:  %s ", storage.Name),
					fmt.Sprintf("request failed "))
				log.Error(err, "error deleting namespace in create failure case")
				return fmt.Errorf("error deleting namespace")
			}
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : deleteNamespace (in create failure case), filesystem:  %s ", storage.Name),
				fmt.Sprintf("request succeeded "))
		}
		return fmt.Errorf("error creating filesystem: %w", err)
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : createVastFilesystem, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	log.Info("filesystem created successfully", "resp :", fsresp)
	// Marshal fsresp to JSON
	fsrespJSON, err := json.Marshal(fsresp)
	if err != nil {
		log.Error(err, "failed to marshal fsresp to JSON")
		return err
	}
	var data FilesystemResponse
	err = json.Unmarshal([]byte(fsrespJSON), &data)
	if err != nil {
		fmt.Println("Error parsing response JSON:", err)
		return err
	}
	// Convert NamespaceId and FilesystemId from string to int64
	namespaceIDInt, err := strconv.ParseInt(data.Id.NamespaceId.Id, 10, 64)
	if err != nil {
		fmt.Println("Error converting NamespaceId to int64:", err)
		return err
	}

	filesystemIDInt, err := strconv.ParseInt(data.Id.Id, 10, 64)
	if err != nil {
		fmt.Println("Error converting FilesystemId to int64:", err)
		return err
	}
	// Update the status of the storage resource
	storage.Status.VolumeProps = cloudv1alpha1.VolumeProperties{
		NamespaceId:  namespaceIDInt,
		FilesystemId: filesystemIDInt,
	}

	return nil
}

type FilesystemResponse struct {
	Id struct {
		NamespaceId struct {
			ClusterId struct {
				Uuid string `json:"uuid"`
			} `json:"cluster_id"`
			Id string `json:"id"`
		} `json:"namespace_id"`
		Id string `json:"id"`
	} `json:"id"`
	Name     string   `json:"name"`
	Capacity struct{} `json:"capacity"`
	Path     string   `json:"path"`
}

// Mapping function
func convertToFilesystemProtocol(protocol cloudv1alpha1.VastFilesystemMountProtocol) storageControllerVastApi.Filesystem_Protocol {
	switch protocol {
	case cloudv1alpha1.FilesystemMountProtocolNFSV3:
		return storageControllerVastApi.Filesystem_PROTOCOL_NFS_V3
	case cloudv1alpha1.FilesystemMountProtocolNFSV4:
		return storageControllerVastApi.Filesystem_PROTOCOL_NFS_V4
	case cloudv1alpha1.FilesystemMountProtocolSMB:
		return storageControllerVastApi.Filesystem_PROTOCOL_SMB
	default:
		return storageControllerVastApi.Filesystem_PROTOCOL_UNSPECIFIED
	}
}

// This function deletes the Storage Namespaces and FileSystems .
func (r *StorageReconciler) deleteStorage(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
	// TODO (Implement validation for the requests)
	log := log.FromContext(ctx).WithName("StorageReconciler.deleteStorage")
	log.Info("inside the deleteStorage function of vast storage operator")
	nsMetadata := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
	}
	namespaceExisting, err := r.StorageControllerClient.GetVastNamespace(ctx, nsMetadata)
	nsQuery := storagecontroller.NamespaceMetadata{
		Name: storage.Spec.ClusterAssignment.NamespaceName,
		UUID: storage.Spec.ClusterAssignment.ClusterUUID,
		Id:   namespaceExisting.Metadata.Id,
	}
	if err != nil {
		log.Error(err, "error in namespace fetch from backend")
		return fmt.Errorf("error in namespace fetch from backend")
	}

	// List filesystems
	listParams := &storagecontroller.ListFilesystemsParams{
		NamespaceID: namespaceExisting.Metadata.Id,
		ClusterID:   storage.Spec.ClusterAssignment.ClusterUUID,
	}
	log.Info("listing filesystems", "params", listParams)
	log.Info("existsing ns", "ns", namespaceExisting)
	log.Info("spec ns", "spec", storage.Spec)
	if namespaceExisting.Metadata.Id == "" || namespaceExisting.Metadata.Name == "" {
		log.Info("namespace not found, skipping deletion")
		return nil
	}

	filesystems, err := r.StorageControllerClient.ListVastFilesystems(ctx, listParams)
	if err != nil {
		return fmt.Errorf("failed to list fs in vast: %v", err)
	}

	var filesystemID string
	// We don't have an isExists implementation from the sds side, so using this loop to find the fs.
	// TODO : Optimize this lookup once we have method available.
	searchFilesystemName := storage.Namespace + "-" + storage.Spec.FilesystemName

	for _, fs := range filesystems {
		if fs.Name == searchFilesystemName {
			filesystemID = fs.Id.Id
			break
		}
	}

	if filesystemID == "" {
		log.Info("no fs in vast backend")
		return nil
	}

	deleteFilesystemParams := &storagecontroller.DeleteFilesystemParams{
		NamespaceID:  namespaceExisting.Metadata.Id,
		FilesystemID: filesystemID,
		ClusterID:    storage.Spec.ClusterAssignment.ClusterUUID,
	}
	err = r.StorageControllerClient.DeleteVastFilesystem(ctx, deleteFilesystemParams)

	if err != nil {
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : deleteVastFilesystem, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		log.Error(err, "error deleting FS")
		return fmt.Errorf("error deleting FS")
	}
	log.Info("delete fs in vast completed")

	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : deleteVastFilesystem, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	nsProps := storagecontroller.NamespaceProperties{
		Quota:     namespaceExisting.Properties.Quota,
		IPFilters: namespaceExisting.Properties.IPFilters,
	}

	nsObj := storagecontroller.Namespace{
		Metadata:   nsQuery,
		Properties: nsProps,
	}

	// Modify namespace request with flag to shrink size for delete.
	log.Info("shrinking namespace capacity", "size is : ", storage.Spec.StorageRequest.Size)
	log.Info("nsObject is ", "nsObj is : ", nsObj)

	err = r.StorageControllerClient.ModifyVastNamespace(ctx, nsObj, false, storage.Spec.StorageRequest.Size)
	if err != nil {
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : updateNamespace (shrink) in delete, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request failed "))
		log.Error(err, "error shrinking namespace capacity in delete")
		return fmt.Errorf("error shrinking namespace capacity in delete flow")
	}
	r.Recorder.Event(storage,
		k8sv1.EventTypeNormal,
		fmt.Sprintf("SDS Controller API : : updateNamespace, filesystem:  %s ", storage.Name),
		fmt.Sprintf("request succeeded "))
	quota, err := strconv.ParseInt(namespaceExisting.Properties.Quota, 10, 64)
	if err != nil {
		return err
	}
	fsSize, err := strconv.ParseInt(storage.Spec.StorageRequest.Size, 10, 64)
	if err != nil {
		return err
	}
	if (quota - fsSize) == 0 {
		err = r.StorageControllerClient.DeleteNamespace(ctx, nsQuery)
		if err != nil {
			r.Recorder.Event(storage,
				k8sv1.EventTypeNormal,
				fmt.Sprintf("SDS Controller API : : deleteNamespace, filesystem:  %s ", storage.Name),
				fmt.Sprintf("request failed "))
			log.Error(err, "error deleting namespace")
			return fmt.Errorf("error deleting namespace")
		}
		r.Recorder.Event(storage,
			k8sv1.EventTypeNormal,
			fmt.Sprintf("SDS Controller API : : deleteNamespace, filesystem:  %s ", storage.Name),
			fmt.Sprintf("request succeeded "))
	}

	log.Info("storage resources cleaned in vast FS Flow")
	return nil
}

// Update accepted condition, phase, and message.
func (r *StorageReconciler) updateStatusAndPersist(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {

	log := log.FromContext(ctx).WithName("StorageReconciler.updateStatusAndPersist")
	log.Info("inside the updateStatusAndPersist function of vast storage operator")

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
			message = cloudv1alpha1.VastStorageMessageRunning
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionRunning, condStatus, reason, message); err != nil {
				return err
			}
		} else {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonNotAccepted
			message = cloudv1alpha1.VastStorageMessageFailed
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionFailed, condStatus, reason, message); err != nil {
				return err
			}
		}
	} else {
		if nsCreationDone && fSCreationDone {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonAccepted
			message = cloudv1alpha1.VastStorageMessageRunning
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionRunning, condStatus, reason, message); err != nil {
				return err
			}
		} else {
			condStatus = k8sv1.ConditionTrue
			reason = cloudv1alpha1.StorageConditionReasonNotAccepted
			message = cloudv1alpha1.VastStorageMessageFailed
			if err := r.updateStatusCondition(ctx, storage, cloudv1alpha1.StorageConditionFailed, condStatus, reason, message); err != nil {
				return err
			}
		}
	}
	updatedStorage, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update storage status")
	}
	if err := r.updateStatusPhaseAndMessage(ctx, updatedStorage); err != nil {
		return fmt.Errorf("updateStatusPhaseAndMessage: %w", err)
	}

	return nil
}

// Update a status condition.
func (r *StorageReconciler) updateStatusCondition(ctx context.Context, storage *cloudv1alpha1.VastStorage,
	storageConditionType cloudv1alpha1.StorageConditionType,
	status k8sv1.ConditionStatus, reason cloudv1alpha1.StorageConditionReason, message string,
) error {
	log := log.FromContext(ctx).WithName("StorageReconciler.updateStatusCondition")
	log.Info("inside the updateStatusCondition function of vast storage operator")

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

// Set status phase and message based on conditions.
// The message will come from the condition that is most relevant.
func (r *StorageReconciler) updateStatusPhaseAndMessage(ctx context.Context, storage *cloudv1alpha1.VastStorage) error {
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
		message = cloudv1alpha1.VastStorageMessageDeleting
	} else {
		// If StaaS is ready and running.
		if running {
			message = fmt.Sprintf(cloudv1alpha1.VastStorageMessageRunning)
			phase = cloudv1alpha1.FilesystemPhaseReady
		} else {
			// not running
			message = fmt.Sprintf(cloudv1alpha1.VastStorageMessageFailed)
			phase = cloudv1alpha1.FilesystemPhaseFailed
		}
	}

	storage.Status.Phase = phase
	storage.Status.Message = message
	_, err := r.VastStoragesClientSet.PrivateV1alpha1().VastStorages(storage.Namespace).UpdateStatus(ctx, storage, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update storage status")
	}
	log.Info("END: storage status details", logkeys.StatusPhase, storage.Status.Phase, logkeys.StatusMessage, storage.Status.Message, logkeys.StatusConditions, storage.Status.Conditions)
	return nil

}
