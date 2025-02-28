package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/bucket_replicator/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	retry "github.com/sethvargo/go-retry"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	toolscache "k8s.io/client-go/tools/cache"
)

type BucketReplicatorService struct {
	syncTicker      *time.Ticker
	Cfg             *config.Config
	bucketAPIClient pb.ObjectStorageServicePrivateClient
	k8sclient       dynamic.Interface
	informer        toolscache.SharedIndexInformer
}

func NewBucketReplicatorService(
	ctx context.Context, cfg *config.Config, k8sconfig *rest.Config) (*BucketReplicatorService, error) {
	logger := log.FromContext(ctx).WithName("NewBucketReplicatorService")

	if k8sconfig == nil {
		return nil, fmt.Errorf("k8s rest config is required")
	}

	if cfg.IDCServiceConfig.StorageAPIGrpcEndpoint == "" {
		logger.Info("config not found: `storageAPIGrpcEndpoint`")
		return nil, fmt.Errorf("error initializing storagr replicator service")
	}
	storageAPIConn, err := grpcutil.NewClient(ctx, cfg.IDCServiceConfig.StorageAPIGrpcEndpoint)
	if err != nil {
		logger.Error(err, "error creating storageServiceClient")
		return nil, fmt.Errorf("error initializing storagr replicator service")
	}

	// Define the API version and resource type of the custom resource

	// Setup client
	clusterClient, err := dynamic.NewForConfig(k8sconfig)
	if err != nil {
		logger.Error(err, "error initializing dynamic kubertetes client")
		return nil, fmt.Errorf("error initializing storagr replicator service")
	}
	bucketApiClient := pb.NewObjectStorageServicePrivateClient(storageAPIConn)

	// Ensure that we can ping the filesystem service before starting the manager.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := bucketApiClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
		return nil, fmt.Errorf("unable to ping bucket service: %w", err)
	}

	resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "objectstores"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()

	return &BucketReplicatorService{
		syncTicker:      time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		Cfg:             cfg,
		bucketAPIClient: bucketApiClient,
		informer:        informer,
		k8sclient:       clusterClient,
	}, nil
}

func (bucketSched *BucketReplicatorService) StartBucketReplicationScheduler(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.StartBucketReplicationScheduler")
	logger.Info("start bucket replication scheduler")

	// Create an informer for the custom resource
	_, err := bucketSched.informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// No need to handle the event
			logger.Info("Resource added, ignore", logkeys.Object, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Handle resource updated event
			oldO := oldObj.(*unstructured.Unstructured)
			newO := newObj.(*unstructured.Unstructured)
			if reflect.DeepEqual(oldO, newO) {
				return
			}
			logger.Info("Resource updated:", logkeys.NewObject, newObj)
			err := bucketSched.handleResourceUpdate(ctx, oldO, newO)
			if err != nil {
				logger.Info("Update error", logkeys.Error, err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			logger.Info("handle delete event", logkeys.Namespace, u.GetNamespace(), logkeys.ObjectName, u.GetName())
			bucketSched.removeFinalizer(ctx, u.GetNamespace(), u.GetName())
		},
	})
	if err != nil {
		logger.Error(err, "error adding event handler")
	}

	go bucketSched.informer.Run(ctx.Done())

	bucketSched.ScanReplicationSchedulerLoop(ctx)
}

func (bucketSched *BucketReplicatorService) ScanReplicationSchedulerLoop(ctx context.Context) {
	log := log.FromContext(ctx).WithName("BucketReplicatorService.ScanReplicationSchedulerLoop")
	log.Info("bucket replication scheduler")
	var latestVersion int64
	for {
		latestVersion = bucketSched.Replicate(ctx, latestVersion)
		tm := <-bucketSched.syncTicker.C
		if tm.IsZero() {
			return
		}
	}

}

func (bucketSched *BucketReplicatorService) Replicate(ctx context.Context, latestVersion int64) int64 {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.Replicate")

	version := strconv.FormatInt(latestVersion, 10)
	params := pb.ObjectBucketSearchPrivateRequest{
		AvailabilityZone: "az1", // constant for now
		ResourceVersion:  version,
	}

	// get all requests
	respStream, err := bucketSched.bucketAPIClient.SearchBucketPrivate(ctx, &params)
	if err != nil {
		logger.Error(err, "error reading requests ")
		return latestVersion
	}

	var bucketReq *pb.ObjectBucketSearchPrivateResponse

	for {
		bucketReq, err = respStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error reading from stream")
			break
		}
		if bucketReq == nil {
			logger.Info("received empty response")
			break
		}
		if bucketReq.Bucket.Metadata.DeletionTimestamp == nil {
			logger.Info("handle bucket create request ", logkeys.Request, bucketReq, logkeys.ResourceVersion, bucketReq.Bucket.Metadata.ResourceVersion)
			if err := bucketSched.processBucketCreate(ctx, bucketReq.Bucket); err != nil {
				logger.Error(err, "error creating bucket private instance")
				// break
				// Do not break, otherwise revision won't be updated
			}
		} else {
			logger.Info("handle bucket delete request ", logkeys.Request, bucketReq, logkeys.ResourceVersion, bucketReq.Bucket.Metadata.ResourceVersion)
			if err := bucketSched.processBucketDelete(ctx, bucketReq.Bucket); err != nil {
				logger.Error(err, "error deletig bucket private instance")
				// break
				// Do not break, otherwise revision won't be updated
			}
		}

		currResVer, err := strconv.ParseInt(bucketReq.Bucket.Metadata.ResourceVersion, 10, 64)
		if err != nil {
			logger.Error(err, "error parsing bucket response")
			break
		}
		if latestVersion < currResVer {
			latestVersion = currResVer
		}
	}

	return latestVersion
}

func (storageSched *BucketReplicatorService) processBucketDelete(ctx context.Context, instance *pb.ObjectBucketPrivate) error {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.processBucketDelete")
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	logger.Info("deleting bucket private instance")
	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		// Delete the resource
		err := storageSched.k8sclient.Resource(
			getBucketResource()).
			Namespace(instance.Metadata.CloudAccountId).
			Delete(ctx,
				instance.Metadata.Name, metav1.DeleteOptions{})
		if err != nil {
			if IsResourceNotFoundError(err) {
				logger.Info("resource not found, remove finalizer and skip")
				storageSched.removeFinalizer(ctx, instance.Metadata.CloudAccountId, instance.Metadata.Name)
				return nil
			}
			logger.Error(err, "error deleting bucket instance")
			return retry.RetryableError(fmt.Errorf("bucket state deletion failed, retry again: resourceName:  %s", instance.Metadata.Name))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create after maximum retries")
	}
	logger.Info("bucket private instance deleted successfully", logkeys.ResourceName, instance.Metadata.Name)
	return nil
}

func (storageSched *BucketReplicatorService) processBucketCreate(ctx context.Context, instance *pb.ObjectBucketPrivate) error {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.processBucketCreate")
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	logger.Info("creating bucket private instance")
	bucketObj := getBucketObject(instance)
	if bucketObj == nil {
		logger.Info("error composing bucket storage object")
		return fmt.Errorf("error creating bucket resource")
	}
	if err := storageSched.createNamespaceIfNeeded(ctx, instance.Metadata.CloudAccountId); err != nil {
		logger.Error(err, "error creating namespace")
		return fmt.Errorf("error creating bucket resource")
	}

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		_, err := storageSched.k8sclient.Resource(
			getBucketResource()).
			Namespace(instance.Metadata.CloudAccountId).
			Create(ctx,
				bucketObj, metav1.CreateOptions{})
		if err != nil {
			if IsResourceAlreadyExistsError(err) {
				logger.Info("resource already exists, skip")
				return nil
			}
			logger.Error(err, "error creating bucket instance")
			return retry.RetryableError(fmt.Errorf("bucket state creation failed, retry again: resourceName: %s", instance.Metadata.Name))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create after maximum retries")
	}
	logger.Info("bucket private instance created successfully", logkeys.ResourceName, instance.Metadata.Name)
	return nil
}

func getBucketObject(instance *pb.ObjectBucketPrivate) *unstructured.Unstructured {
	storage := &cloudv1alpha1.ObjectStore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "ObjectStore",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Metadata.Name,
		},
		Spec: cloudv1alpha1.ObjectStoreSpec{
			AvailabilityZone:   instance.Spec.AvailabilityZone,
			Versioned:          instance.Spec.Versioned,
			Quota:              utils.ConvertToBytes(instance.Spec.Request.Size),
			BucketAccessPolicy: mapBucketAccessPolicyFromPBToK8s(instance.Spec.AccessPolicy),
			ObjectStoreBucketSchedule: cloudv1alpha1.ObjectStoreBucketSchedule{
				ObjectStoreCluster: cloudv1alpha1.ObjectStoreAssignedCluster{
					Name: instance.Spec.Schedule.Cluster.ClusterName,
					UUID: instance.Spec.Schedule.Cluster.ClusterUUID,
					Addr: instance.Spec.Schedule.Cluster.ClusterAddr,
				},
			},
		},
	}
	payload, err := json.Marshal(storage)
	if err != nil {
		return nil
	}
	// Create an empty unstructured object
	obj := &unstructured.Unstructured{}
	// Unmarshal the JSON data into the unstructured object
	err = obj.UnmarshalJSON(payload)
	if err != nil {
		return nil
	}
	return obj
}

func IsResourceAlreadyExistsError(err error) bool {
	if err != nil {
		statusErr, isStatus := err.(*apierr.StatusError)
		if isStatus && statusErr.Status().Reason == metav1.StatusReasonAlreadyExists {
			return true
		}
	}
	return false
}

func IsResourceNotFoundError(err error) bool {
	if err != nil {
		statusErr, isStatus := err.(*apierr.StatusError)
		if isStatus && statusErr.Status().Reason == metav1.StatusReasonNotFound {
			return true
		}
	}
	return false
}

func (bucketSched *BucketReplicatorService) removeFinalizer(ctx context.Context, cloudaccountId, resourceName string) {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.removeFinalizer")
	logger.Info("removing finalizer for resource", logkeys.ResourceName, resourceName)

	_, err := bucketSched.bucketAPIClient.RemoveBucketFinalizer(ctx, &pb.ObjectBucketRemoveFinalizerRequest{
		Metadata: &pb.ObjectBucketIdReference{
			CloudAccountId: cloudaccountId,
			ResourceId:     resourceName,
		},
	})
	if err != nil {
		logger.Error(err, "error removing finalizer from api service")
	}
	logger.Info("bucket finalizer removed successfully", logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceName)
}

func getBucketResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "private.cloud.intel.com",
		Version:  "v1alpha1",
		Resource: "objectstores"}
}

func (storageSched *BucketReplicatorService) createNamespaceIfNeeded(ctx context.Context, namespace string) error {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.createNamespaceIfNeeded")
	// Define the Namespace object
	namespaceObj := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	payload, err := json.Marshal(namespaceObj)
	if err != nil {
		return err
	}
	// Create an empty unstructured object
	obj := &unstructured.Unstructured{}
	// Unmarshal the JSON data into the unstructured object
	err = obj.UnmarshalJSON(payload)
	if err != nil {
		return err
	}

	// Create the Namespace
	_, err = storageSched.k8sclient.Resource(
		schema.GroupVersionResource{
			Version:  "v1",
			Resource: "namespaces",
		},
	).Create(ctx, obj, metav1.CreateOptions{})

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	logger.Info("namespace created successfully (or already exists)", logkeys.Namespace, namespace)
	return nil
}

func (bucketSched *BucketReplicatorService) handleResourceUpdate(ctx context.Context, oldObj, newObj *unstructured.Unstructured) error {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.handleResourceUpdate")
	newObjName := newObj.GetName()
	newObjNamespace := newObj.GetNamespace()

	newK8sRes, err := bucketSched.getK8sBucketResource(ctx, newObjNamespace, newObjName)
	if err != nil {
		logger.Error(err, "error reading file resource from k8s", logkeys.Namespace, newObjNamespace, logkeys.Name, newObjName)
		return fmt.Errorf("resource update handler failed")
	}

	// if newK8sRes.Status.Phase == oldK8sRes.Status.Phase {
	// 	logger.Info("no phase change observed in the status, skip resource update", "namespace", oldObjNamespace, "name", oldObjName)
	// 	return nil
	// }

	_, err = bucketSched.bucketAPIClient.UpdateBucketStatus(ctx, getObjectUpdateStatusRequest(newK8sRes))
	if err != nil {
		logger.Error(err, "error updating resource state to api server")
		return fmt.Errorf("resource update handler failed")
	}
	logger.Info("resource state updated to api server successfully")
	return nil
}

func (bucketSched *BucketReplicatorService) getK8sBucketResource(ctx context.Context, namespace, name string) (*cloudv1alpha1.ObjectStore, error) {
	logger := log.FromContext(ctx).WithName("BucketReplicatorService.getK8sBucketResource")
	res := cloudv1alpha1.ObjectStore{}
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		obj, err := bucketSched.k8sclient.Resource(
			getBucketResource()).
			Namespace(namespace).
			Get(ctx,
				name, metav1.GetOptions{})
		if err != nil {
			if IsResourceNotFoundError(err) {
				logger.Info("resource not found, skip")
				return nil
			}
			logger.Error(err, "error reading filesystem instance")
			return retry.RetryableError(fmt.Errorf("filesystem read failed, retry again: resourceName: %s", name))
		}
		unstructured := obj.UnstructuredContent()
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(unstructured, &res)
		if err != nil {
			logger.Error(err, "error parsing resource response")
			return fmt.Errorf("error reading resource")
		}
		return nil
	}); err != nil {
		return &res, fmt.Errorf("failed to get after maximum retries")
	}
	logger.Info("resource to be returned ", logkeys.Resource, res)
	return &res, nil
}

func getObjectUpdateStatusRequest(bucket *cloudv1alpha1.ObjectStore) *pb.ObjectBucketStatusUpdateRequest {

	res := pb.ObjectBucketStatusUpdateRequest{
		Metadata: &pb.ObjectBucketIdReference{
			CloudAccountId:  bucket.ObjectMeta.Namespace,
			ResourceId:      bucket.ObjectMeta.Name,
			ResourceVersion: bucket.ObjectMeta.ResourceVersion,
		},
		Status: &pb.ObjectBucketStatus{
			Phase: mapResPhaseFromK8sToPB(bucket.Status.Phase),
		},
	}
	return &res
}

func mapResPhaseFromK8sToPB(phase cloudv1alpha1.ObjectStorePhase) pb.BucketPhase {
	switch phase {
	case cloudv1alpha1.ObjectStorePhasePhaseProvisioning:
		return pb.BucketPhase_BucketProvisioning
	case cloudv1alpha1.ObjectStorePhasePhaseFailed:
		return pb.BucketPhase_BucketFailed
	case cloudv1alpha1.ObjectStorePhasePhaseReady:
		return pb.BucketPhase_BucketReady
	case cloudv1alpha1.ObjectStorePhasePhaseTerminating:
		return pb.BucketPhase_BucketDeleting
	default:
		return pb.BucketPhase_BucketProvisioning
	}
}

func mapBucketAccessPolicyFromPBToK8s(policy pb.BucketAccessPolicy) cloudv1alpha1.BucketAccessPolicy {
	switch policy {
	case pb.BucketAccessPolicy_NONE:
		return cloudv1alpha1.BucketAccessPolicyNone
	case pb.BucketAccessPolicy_READ:
		return cloudv1alpha1.BucketAccessPolicyReadOnly
	case pb.BucketAccessPolicy_READ_WRITE:
		return cloudv1alpha1.BucketAccessPolicyReadWrite
	case pb.BucketAccessPolicy_UNSPECIFIED:
		return cloudv1alpha1.BucketAccessPolicyUnspecified
	default:
		return cloudv1alpha1.BucketAccessPolicyUnspecified
	}
}
