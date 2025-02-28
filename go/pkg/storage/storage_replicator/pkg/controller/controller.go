// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
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

var vastEnabled bool

type StorageReplicatorService struct {
	syncTicker       *time.Ticker
	Cfg              *config.Config
	storageAPIClient pb.FilesystemPrivateServiceClient
	k8sclient        dynamic.Interface
	informer         toolscache.SharedIndexInformer
	informerVast     toolscache.SharedIndexInformer
}

func NewStorageReplicatorService(
	ctx context.Context, cfg *config.Config, k8sconfig *rest.Config) (*StorageReplicatorService, error) {
	logger := log.FromContext(ctx).WithName("NewStorageReplicatorService")

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
	storageApiClient := pb.NewFilesystemPrivateServiceClient(storageAPIConn)

	// Ensure that we can ping the filesystem service before starting the manager.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := storageApiClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
		return nil, fmt.Errorf("unable to ping instance service: %w", err)
	}

	resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "storages"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, 4*time.Second, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()

	vastresource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "vaststorages"}
	vastResfactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, 4*time.Second, corev1.NamespaceAll, nil)
	vastinformer := vastResfactory.ForResource(vastresource).Informer()

	return &StorageReplicatorService{
		syncTicker:       time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		Cfg:              cfg,
		storageAPIClient: storageApiClient,
		informer:         informer,
		informerVast:     vastinformer,
		k8sclient:        clusterClient,
	}, nil
}

func (storageSched *StorageReplicatorService) StartStorageReplicationScheduler(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("StorageReplicatorService.StartComlianceScanScheduler")
	logger.Info("start storage replication scheduler")

	// Create an informer for the custom resource weka storage
	_, err := storageSched.informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
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
			err := storageSched.handleResourceUpdate(ctx, oldO, newO)
			if err != nil {
				logger.Info("Update error", logkeys.Error, err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			logger.Info("handle delete event", logkeys.Namespace, u.GetNamespace(), logkeys.ObjectName, u.GetName())
			storageSched.removeFinalizer(ctx, u.GetNamespace(), u.GetName())
		},
	})
	if err != nil {
		logger.Error(err, "error adding event handler")
	}

	// Create an informer for the custom resource VAST storage
	_, err = storageSched.informerVast.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// No need to handle the event
			logger.Info("VAST resource added, ignore", logkeys.Object, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Handle resource updated event
			oldO := oldObj.(*unstructured.Unstructured)
			newO := newObj.(*unstructured.Unstructured)
			if reflect.DeepEqual(oldO, newO) {
				return
			}
			err := storageSched.handleVASTResourceUpdate(ctx, oldO, newO)
			if err != nil {
				logger.Info("Update error", logkeys.Error, err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			logger.Info("VAST resource handle delete event", logkeys.Namespace, u.GetNamespace(), logkeys.ObjectName, u.GetName())
			storageSched.removeFinalizer(ctx, u.GetNamespace(), u.GetName())
		},
	})
	if err != nil {
		logger.Error(err, "error adding vast event handler")
	}
	// Set vast flag
	vastEnabled = storageSched.Cfg.IDCServiceConfig.VASTEnabled

	go storageSched.informer.Run(ctx.Done())
	go storageSched.informerVast.Run(ctx.Done())

	storageSched.ScanReplicationSchedulerLoop(ctx)
}

func (storageSched *StorageReplicatorService) ScanReplicationSchedulerLoop(ctx context.Context) {
	log := log.FromContext(ctx).WithName("StorageReplicatorService.ScanReplicationSchedulerLoop")
	log.Info("storage replication scheduler")
	var latestVersion int64
	for {
		latestVersion = storageSched.Replicate(ctx, latestVersion)
		tm := <-storageSched.syncTicker.C
		if tm.IsZero() {
			return
		}
	}

}

func (storageSched *StorageReplicatorService) Replicate(ctx context.Context, latestVersion int64) int64 {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.Replicate").Start()
	defer span.End()

	version := strconv.FormatInt(latestVersion, 10)
	params := pb.FilesystemSearchStreamPrivateRequest{
		AvailabilityZone: "az1", // constant for now
		ResourceVersion:  version,
	}

	// get all requests
	respStream, err := storageSched.storageAPIClient.SearchFilesystemRequests(ctx, &params)
	if err != nil {
		logger.Error(err, "error reading requests ")
		return latestVersion
	}

	var fsReq *pb.FilesystemRequestResponse

	for {
		fsReq, err = respStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error reading from stream")
			break
		}
		if fsReq == nil {
			logger.Info("received empty response")
			break
		}
		if fsReq.Filesystem.Metadata.DeletionTimestamp != nil {
			logger.Info("handle filesystem delete request", logkeys.Request, fsReq, logkeys.ResourceVersion, fsReq.Filesystem.Metadata.ResourceVersion)
			if err := storageSched.processFilesystemDelete(ctx, fsReq.Filesystem); err != nil {
				logger.Error(err, "error deletig filesystem private instance")
				// break
				// Do not break, otherwise revision won't be updated
			}

		} else if fsReq.Filesystem.Metadata.UpdateTimestamp != nil {
			logger.Info("handle filesystem update request", logkeys.Request, fsReq, logkeys.ResourceVersion, fsReq.Filesystem.Metadata.ResourceVersion)
			if err := storageSched.processFilesystemUpdate(ctx, fsReq.Filesystem); err != nil {
				logger.Error(err, "error updating filesystem private instance")
				// break
				// Do not break, otherwise revision won't be updated
			}
		} else {
			if err := storageSched.processFilesystemCreate(ctx, fsReq.Filesystem); err != nil {
				logger.Info("handle filesystem create request", logkeys.Request, fsReq, logkeys.ResourceVersion, fsReq.Filesystem.Metadata.ResourceVersion)
				logger.Error(err, "error creating filesystem private instance")
				// break
				// Do not break, otherwise revision won't be updated
			}
		}

		currResVer, err := strconv.ParseInt(fsReq.Filesystem.Metadata.ResourceVersion, 10, 64)
		if err != nil {
			logger.Error(err, "error parsing filesystem response")
			break
		}
		if latestVersion < currResVer {
			latestVersion = currResVer
		}
	}

	return latestVersion
}

func (storageSched *StorageReplicatorService) processFilesystemDelete(ctx context.Context, instance *pb.FilesystemPrivate) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.processFilesystemDelete").Start()
	defer span.End()

	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	logger.Info("deleting filesystem private instance")
	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		var err error
		if instance.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
			// Delete the resource for vast filesystem
			err = storageSched.k8sclient.Resource(
				getFileStorageResourceVAST()).
				Namespace(instance.Metadata.CloudAccountId).
				Delete(ctx,
					instance.Metadata.Name, metav1.DeleteOptions{})
		} else {
			// Delete the resource for weka filesystem
			err = storageSched.k8sclient.Resource(
				getFileStorageResource()).
				Namespace(instance.Metadata.CloudAccountId).
				Delete(ctx,
					instance.Metadata.Name, metav1.DeleteOptions{})
		}
		if err != nil {
			if IsResourceNotFoundError(err) {
				logger.Info("resource not found, remove finalizer and skip")
				storageSched.removeFinalizer(ctx, instance.Metadata.CloudAccountId, instance.Metadata.Name)
				return nil
			}
			logger.Error(err, "error deleting filesystem instance")
			return retry.RetryableError(fmt.Errorf("filesystem state deletion failed, retry again: resourceName:  %s", instance.Metadata.Name))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create after maximum retries")
	}
	logger.Info("filesystem private instance deleted successfully", logkeys.InstanceName, instance.Metadata.Name)
	return nil
}

func (storageSched *StorageReplicatorService) processFilesystemCreate(ctx context.Context, instance *pb.FilesystemPrivate) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.processFilesystemCreate").Start()
	defer span.End()
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	logger.Info("creating filesystem private instance")
	storageObj := getStorageObject(ctx, instance)

	if storageObj == nil {
		logger.Info("error composing filesystem storage object")
		return fmt.Errorf("error creating filesystem resource")
	}
	if err := storageSched.createNamespaceIfNeeded(ctx, instance.Metadata.CloudAccountId); err != nil {
		logger.Error(err, "error creating namespace")
		return fmt.Errorf("error creating filesystem resource")
	}

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		var gvr schema.GroupVersionResource

		if instance.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
			gvr = getFileStorageResourceVAST()
		} else {
			gvr = getFileStorageResource()
		}
		_, err := storageSched.k8sclient.Resource(
			gvr).
			Namespace(instance.Metadata.CloudAccountId).
			Create(ctx,
				storageObj, metav1.CreateOptions{})
		if err != nil {
			if IsResourceAlreadyExistsError(err) {
				logger.Info("namespace already exists, skip")
				return nil
			}
			logger.Error(err, "error creating filesystem instance")
			return retry.RetryableError(fmt.Errorf("filesystem state creation failed, retry again: resourceName: %s", instance.Metadata.Name))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create after maximum retries")
	}
	logger.Info("filesystem private instance created successfully", logkeys.InstanceName, instance.Metadata.Name)
	return nil
}

func (storageSched *StorageReplicatorService) processFilesystemUpdate(ctx context.Context, instance *pb.FilesystemPrivate) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.processFilesystemUpdate").Start()
	defer span.End()
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	logger.Info("updating filesystem private instance")

	if instance.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
		vastStorageObj, err := storageSched.k8sclient.Resource(
			getFileStorageResourceVAST()).
			Namespace(instance.Metadata.CloudAccountId).
			Get(ctx,
				instance.Metadata.Name, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "resource not found")
			return nil
		}
		vastVolRes := &cloudv1alpha1.VastStorage{}
		unObj := vastStorageObj.UnstructuredContent()
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(unObj, &vastVolRes)
		if err != nil {
			logger.Error(err, "error parsing resource response")
			return fmt.Errorf("error reading resource")
		}
		vastVolRes.Spec.StorageRequest.Size = utils.ConvertToBytes(instance.Spec.Request.Storage)
		if instance.Spec.SecurityGroup == nil {
			vastVolRes.Spec.Networks.SecurityGroups = cloudv1alpha1.SecurityGroups{
				IPFilters: []cloudv1alpha1.IPFilter{},
			}
		} else {
			for _, ipfilter := range instance.Spec.SecurityGroup.NetworkFilterAllow {
				dupIpFilter := false
				startIp, endIp, err := getIPRange(ipfilter.Subnet, int(ipfilter.PrefixLength))
				if err != nil {
					logger.Info("error decoding ip range", "subnet", ipfilter.Subnet, "prefix length", ipfilter.PrefixLength)
					// vastVolRes.Spec.Networks.SecurityGroups = cloudv1alpha1.SecurityGroups{
					// 	IPFilters: []cloudv1alpha1.IPFilter{},
					// }
				} else {
					logger.Info("ip range decoded", "subnet", ipfilter.Subnet, "prefix length", ipfilter.PrefixLength, "startIp", startIp, "endIp", endIp)
					for _, existingRule := range vastVolRes.Spec.Networks.SecurityGroups.IPFilters {
						if strings.EqualFold(existingRule.Start, startIp.String()) &&
							strings.EqualFold(existingRule.End, endIp.String()) {
							logger.Info("duplicate ip range filter, ignoring", "subnet", ipfilter.Subnet, "prefix length", ipfilter.PrefixLength, "startIp", startIp, "endIp", endIp)
							dupIpFilter = true
							break
						}
					}
					if !dupIpFilter {
						vastVolRes.Spec.Networks.SecurityGroups.IPFilters = append(vastVolRes.Spec.Networks.SecurityGroups.IPFilters, cloudv1alpha1.IPFilter{
							Start: startIp.String(),
							End:   endIp.String(),
						})
					}
				}
			}
		}
		vastResBytes, err := json.Marshal(vastVolRes)
		if err != nil {
			return nil
		}

		// Create an empty unstructured object
		updatedVastResObj := &unstructured.Unstructured{}
		// Unmarshal the JSON data into the unstructured object
		err = updatedVastResObj.UnmarshalJSON(vastResBytes)
		if err != nil {
			return nil
		}

		// Add update logic for updating network security gruoup
		// vastStorageObj.Object["spec"].(map[string]interface{})["request"].(map[string]interface{})["storage"] =
		// 	utils.ConvertToBytes(instance.Spec.Request.Storage)
		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			_, err := storageSched.k8sclient.Resource(
				getFileStorageResourceVAST()).
				Namespace(instance.Metadata.CloudAccountId).
				Update(ctx,
					updatedVastResObj, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "error updating vast filesystem instance")
				return retry.RetryableError(fmt.Errorf("vast filesystem state update failed, retry again: resourceName: %s", instance.Metadata.Name))
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to update after maximum retries")
		}
	} else {
		storageObj, err := storageSched.k8sclient.Resource(
			getFileStorageResource()).
			Namespace(instance.Metadata.CloudAccountId).
			Get(ctx,
				instance.Metadata.Name, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "resource not found")
			return nil
		}
		storageObj.Object["spec"].(map[string]interface{})["request"].(map[string]interface{})["storage"] = utils.ConvertToBytes(instance.Spec.Request.Storage)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			_, err := storageSched.k8sclient.Resource(
				getFileStorageResource()).
				Namespace(instance.Metadata.CloudAccountId).
				Update(ctx,
					storageObj, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "error updating filesystem instance")
				return retry.RetryableError(fmt.Errorf("filesystem state update failed, retry again: resourceName: %s", instance.Metadata.Name))
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to update after maximum retries")
		}
	}
	logger.Info("filesystem private instance updated successfully", logkeys.InstanceName, instance.Metadata.Name)
	return nil
}

func getStorageObject(ctx context.Context, instance *pb.FilesystemPrivate) *unstructured.Unstructured {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.processFilesystemCreate").Start()
	defer span.End()
	// Create a new Storage object
	payload := []byte{}
	var err error
	// Check the storage class and create the appropriate storage object
	if instance.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
		logger.Info("inside vast get storage object")

		secGroups := cloudv1alpha1.SecurityGroups{
			IPFilters: []cloudv1alpha1.IPFilter{},
		}
		if instance.Spec.SecurityGroup != nil {
			logger.Info("inside valid security group for vast")

			for _, ipfilter := range instance.Spec.SecurityGroup.NetworkFilterAllow {
				startIp, endIp, err := getIPRange(ipfilter.Subnet, int(ipfilter.PrefixLength))
				if err != nil {
					logger.Error(err, "error decoding IP range for subnet")
				}
				secGroups.IPFilters = append(secGroups.IPFilters, cloudv1alpha1.IPFilter{
					Start: startIp.String(),
					End:   endIp.String(),
				})
			}
		}

		storageVAST := &cloudv1alpha1.VastStorage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "private.cloud.intel.com/v1alpha1",
				Kind:       "VastStorage",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: instance.Metadata.Name,
			},
			Spec: cloudv1alpha1.VastStorageSpec{
				AvailabilityZone: instance.Spec.AvailabilityZone,
				StorageClass:     cloudv1alpha1.FilesystemStorageClassGeneralPurpose,
				FilesystemName:   instance.Metadata.Name,
				FilesystemType:   cloudv1alpha1.FilesystemTypeComputeGeneral,
				CSIVolumePrefix:  instance.Spec.Prefix,
				StorageRequest: cloudv1alpha1.VASTFilesystemStorageRequest{
					Size: utils.ConvertToBytes(instance.Spec.Request.Storage),
				},
				ClusterAssignment: cloudv1alpha1.ClusterAssignment{
					ClusterUUID:    instance.Spec.Scheduler.Cluster.ClusterUUID,
					ClusterVersion: *instance.Spec.Scheduler.Cluster.ClusterVersion,
					NamespaceName:  instance.Spec.Scheduler.Namespace.Name,
				},
				MountConfig: cloudv1alpha1.MountConfig{
					VolumePath:    instance.Spec.VolumePath,
					MountProtocol: cloudv1alpha1.FilesystemMountProtocolNFSV4,
				},
				Networks: cloudv1alpha1.Networks{
					SecurityGroups: secGroups,
				},
			},
		}

		// if instance.Spec.FilesystemType == pb.FilesystemType_ComputeGeneral || instance.Spec.FilesystemType == pb.FilesystemType_Unspecified {
		// 	storageVAST.Spec.FilesystemType = cloudv1alpha1.FilesystemTypeComputeGeneral
		// } else {
		// 	storageVAST.Spec.FilesystemType = cloudv1alpha1.FilesystemTypeComputeKubernetes
		// }
		// if instance.Spec.StorageClass == pb.FilesystemStorageClass_AIOptimized {
		// 	storageVAST.Spec.StorageClass = cloudv1alpha1.FilesystemStorageClassAIOptimized
		// } else {
		// 	storageVAST.Spec.StorageClass = cloudv1alpha1.FilesystemStorageClassGeneralPurpose
		// }
		payload, err = json.Marshal(storageVAST)
		if err != nil {
			return nil
		}
	} else {
		storageWeka := &cloudv1alpha1.Storage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "private.cloud.intel.com/v1alpha1",
				Kind:       "Storage",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: instance.Metadata.Name,
			},
			Spec: cloudv1alpha1.StorageSpec{
				AvailabilityZone: instance.Spec.AvailabilityZone,
				StorageRequest: cloudv1alpha1.FilesystemStorageRequest{
					Size: utils.ConvertToBytes(instance.Spec.Request.Storage),
				},
				Encrypted: instance.Spec.Encrypted,
				Prefix:    instance.Spec.Prefix,
				ProviderSchedule: cloudv1alpha1.FilesystemSchedule{
					FilesystemName: instance.Metadata.Name,
					Cluster: cloudv1alpha1.AssignedCluster{
						Name:    instance.Spec.Scheduler.Cluster.ClusterName,
						UUID:    instance.Spec.Scheduler.Cluster.ClusterUUID,
						Addr:    instance.Spec.Scheduler.Cluster.ClusterAddr,
						Version: *instance.Spec.Scheduler.Cluster.ClusterVersion,
					},
					Namespace: cloudv1alpha1.AssignedNamespace{
						Name:            instance.Spec.Scheduler.Namespace.Name,
						CredentialsPath: instance.Spec.Scheduler.Namespace.CredentialsPath,
					},
				},
			},
		}
		if instance.Spec.FilesystemType == pb.FilesystemType_ComputeGeneral || instance.Spec.FilesystemType == pb.FilesystemType_Unspecified {
			storageWeka.Spec.FilesystemType = cloudv1alpha1.FilesystemTypeComputeGeneral
		} else {
			storageWeka.Spec.FilesystemType = cloudv1alpha1.FilesystemTypeComputeKubernetes
		}
		if instance.Spec.StorageClass == pb.FilesystemStorageClass_AIOptimized {
			storageWeka.Spec.StorageClass = cloudv1alpha1.FilesystemStorageClassAIOptimized
		} else {
			storageWeka.Spec.StorageClass = cloudv1alpha1.FilesystemStorageClassGeneralPurpose
		}
		payload, err = json.Marshal(storageWeka)
		if err != nil {
			return nil
		}
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

func (storageSched *StorageReplicatorService) removeFinalizer(ctx context.Context, cloudaccountId, resourceName string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.removeFinalizer").Start()
	defer span.End()
	logger.Info("removing finalizer for resource", logkeys.ResourceId, resourceName)

	_, err := storageSched.storageAPIClient.RemoveFinalizer(ctx, &pb.FilesystemRemoveFinalizerRequest{
		Metadata: &pb.FilesystemIdReference{
			CloudAccountId: cloudaccountId,
			ResourceId:     resourceName,
		},
	})
	if err != nil {
		logger.Error(err, "error removing finalizer from api service")
	}
	logger.Info("filesystem finalizer removed successfully", logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceName)
}

func getFileStorageResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "private.cloud.intel.com",
		Version:  "v1alpha1",
		Resource: "storages"}
}

func getFileStorageResourceVAST() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "private.cloud.intel.com",
		Version:  "v1alpha1",
		Resource: "vaststorages"}
}

func (storageSched *StorageReplicatorService) createNamespaceIfNeeded(ctx context.Context, namespace string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.createNamespaceIfNeeded").Start()
	defer span.End()
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

func (storageSched *StorageReplicatorService) handleResourceUpdate(ctx context.Context, oldObj, newObj *unstructured.Unstructured) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.handleResourceUpdate").Start()
	defer span.End()
	newObjName := newObj.GetName()
	newObjNamespace := newObj.GetNamespace()

	newK8sRes, err := storageSched.getK8sFileResource(ctx, newObjNamespace, newObjName, getFileStorageResource())
	if err != nil {
		logger.Error(err, "error reading file resource from k8s", logkeys.Namespace, newObjNamespace, logkeys.Name, newObjName)
		return fmt.Errorf("resource update handler failed")
	}

	// if newK8sRes.Status.Phase == oldK8sRes.Status.Phase {
	// 	logger.Info("no phase change observed in the status, skip resource update", "namespace", oldObjNamespace, "name", oldObjName)
	// 	return nil
	// }
	_, err = storageSched.storageAPIClient.UpdateStatus(ctx, getFilesystemUpdateStatusRequest(newK8sRes))
	if err != nil {
		logger.Error(err, "error updating resource state to api server")
		return fmt.Errorf("resource update handler failed")
	}
	logger.Info("resource state updated to api server successfully")
	return nil
}

func (storageSched *StorageReplicatorService) handleVASTResourceUpdate(ctx context.Context, oldObj, newObj *unstructured.Unstructured) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.handleVASTResourceUpdate").Start()
	defer span.End()

	var newVastRes, oldVastRes cloudv1alpha1.VastStorage

	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(newObj.Object, &newVastRes)
	if err != nil {
		logger.Error(err, "error parsing new vast resource response")
		return fmt.Errorf("error reading resource")
	}

	err = runtime.DefaultUnstructuredConverter.
		FromUnstructured(oldObj.Object, &oldVastRes)
	if err != nil {
		logger.Error(err, "error parsing old vast resource response")
		return fmt.Errorf("error reading resource")
	}

	logger.Info("debug", "old phase ", oldVastRes.Status.Phase, "new phase", newVastRes.Status.Phase)
	if oldVastRes.Status.Phase == newVastRes.Status.Phase {
		logger.Info("no phase change observed in the status, skip resource update")
		return nil
	}

	_, err = storageSched.storageAPIClient.UpdateStatus(ctx, getVastFilesystemUpdateStatusRequest(&newVastRes))
	if err != nil {
		logger.Error(err, "error updating vast resource state to api server")
		return fmt.Errorf("resource update handler failed")
	}

	logger.Info("vast resource state updated to api server successfully")
	return nil
}

func (storageSched *StorageReplicatorService) getK8sFileResource(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource) (*cloudv1alpha1.Storage, error) {
	logger := log.FromContext(ctx).WithName("StorageReplicatorService.getK8sFileResource")
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.getK8sFileResource").Start()
	defer span.End()
	res := cloudv1alpha1.Storage{}
	backoffTimer := retry.NewConstant(2 * time.Second)
	backoffTimer = retry.WithMaxDuration(10*time.Second, backoffTimer)
	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		obj, err := storageSched.k8sclient.Resource(
			gvr).
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

func getFilesystemUpdateStatusRequest(filesystem *cloudv1alpha1.Storage) *pb.FilesystemUpdateStatusRequest {

	res := pb.FilesystemUpdateStatusRequest{
		Metadata: &pb.FilesystemIdReference{
			CloudAccountId:  filesystem.ObjectMeta.Namespace,
			ResourceId:      filesystem.ObjectMeta.Name,
			ResourceVersion: filesystem.ObjectMeta.ResourceVersion,
		},
		Status: &pb.FilesystemStatusPrivate{
			Phase: mapResPhaseFromK8sToPB(filesystem.Status.Phase),
			Mount: &pb.FilesystemMountStatusPrivate{
				ClusterAddr: filesystem.Status.Mount.ClusterAddr,
			},
		},
	}
	return &res
}

func getVastFilesystemUpdateStatusRequest(filesystem *cloudv1alpha1.VastStorage) *pb.FilesystemUpdateStatusRequest {

	res := pb.FilesystemUpdateStatusRequest{
		Metadata: &pb.FilesystemIdReference{
			CloudAccountId:  filesystem.ObjectMeta.Namespace,
			ResourceId:      filesystem.ObjectMeta.Name,
			ResourceVersion: filesystem.ObjectMeta.ResourceVersion,
		},
		Status: &pb.FilesystemStatusPrivate{
			Phase: mapResPhaseFromK8sToPB(filesystem.Status.Phase),
			Mount: &pb.FilesystemMountStatusPrivate{
				// ClusterAddr: filesystem.Status.Mount.ClusterAddr,
			},
			VolumeIdentifiers: &pb.VolumeIdentifiers{
				Size:         filesystem.Status.VolumeProps.Size,
				TenantId:     filesystem.Status.VolumeProps.NamespaceId,
				FilesystemId: filesystem.Status.VolumeProps.FilesystemId,
			},
		},
	}
	return &res
}

func mapResPhaseFromK8sToPB(phase cloudv1alpha1.FilesystemPhase) pb.FilesystemPhase {
	switch phase {
	case cloudv1alpha1.FilesystemPhaseProvisioning:
		return pb.FilesystemPhase_FSProvisioning
	case cloudv1alpha1.FilesystemPhaseReady:
		return pb.FilesystemPhase_FSReady
	case cloudv1alpha1.FilesystemPhaseFailed:
		return pb.FilesystemPhase_FSFailed
	case cloudv1alpha1.FilesystemPhaseDeleting:
		return pb.FilesystemPhase_FSDeleting
	default:
		return pb.FilesystemPhase_FSProvisioning
	}
}

func getIPRange(subnet string, prefix int) (net.IP, net.IP, error) {
	// Parse the subnet
	ip, ipNet, err := net.ParseCIDR(subnet + "/" + strconv.Itoa(prefix))
	if err != nil {
		return nil, nil, err
	}

	// Get the IP address as a uint32
	ipUint32 := binary.BigEndian.Uint32(ip.To4())

	// Calculate the mask
	mask := binary.BigEndian.Uint32(ipNet.Mask)

	// Calculate start and end of range
	startIP := ipUint32 & mask
	endIP := startIP | ^mask

	// Convert back to net.IP
	start := make(net.IP, 4)
	end := make(net.IP, 4)
	binary.BigEndian.PutUint32(start, startIP)
	binary.BigEndian.PutUint32(end, endIP)

	return start, end, nil
}
