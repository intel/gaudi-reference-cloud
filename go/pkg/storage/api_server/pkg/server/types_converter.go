// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
)

func convertFilesystemPrivateToPublic(fsPrivate *pb.FilesystemPrivate, updateStatus bool) *pb.Filesystem {

	fsPublic := pb.Filesystem{
		Metadata: &pb.FilesystemMetadata{
			CloudAccountId:    fsPrivate.Metadata.CloudAccountId,
			Name:              fsPrivate.Metadata.Name,
			ResourceId:        fsPrivate.Metadata.ResourceId,
			ResourceVersion:   fsPrivate.Metadata.ResourceVersion,
			Description:       fsPrivate.Metadata.Description,
			Labels:            fsPrivate.Metadata.Labels,
			CreationTimestamp: fsPrivate.Metadata.CreationTimestamp,
			DeletionTimestamp: fsPrivate.Metadata.DeletionTimestamp,
		},
		Spec: &pb.FilesystemSpec{
			AvailabilityZone: fsPrivate.Spec.AvailabilityZone,
			Request:          fsPrivate.Spec.Request,
			StorageClass:     fsPrivate.Spec.StorageClass,
			AccessModes:      pb.FilesystemAccessModes_ReadWrite,
			MountProtocol:    pb.FilesystemMountProtocols_Weka,
			Encrypted:        fsPrivate.Spec.Encrypted,
			FilesystemType:   fsPrivate.Spec.FilesystemType,
		},
		Status: &pb.FilesystemStatus{
			Phase:   fsPrivate.Status.Phase,
			Message: fsPrivate.Status.Message,
			// Mount:   (*pb.FilesystemMountStatus)(fsPrivate.Status.Mount),
			//TODO: Fix me
		},
	}
	if updateStatus {
		if fsPrivate.Status.Phase == pb.FilesystemPhase_FSReady {
			if fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_AIOptimized ||
				fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurpose {
				fsPublic.Status.Mount = &pb.FilesystemMountStatus{
					ClusterName:    fsPrivate.Status.Mount.ClusterAddr,
					ClusterAddr:    fsPrivate.Status.Mount.ClusterAddr,
					ClusterVersion: fsPrivate.Spec.Scheduler.Cluster.ClusterVersion,
					Namespace:      fsPrivate.Spec.Scheduler.Namespace.Name,
					FilesystemName: fsPrivate.Metadata.Name,
					Username:       utils.GenerateFilesystemUser(fsPrivate.Metadata.CloudAccountId),
				}
			} else if fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
				fsPublic.Status.Mount = &pb.FilesystemMountStatus{
					ClusterName:    fsPrivate.Spec.Scheduler.Cluster.ClusterAddr,
					ClusterAddr:    fsPrivate.Spec.Scheduler.Cluster.ClusterAddr,
					ClusterVersion: fsPrivate.Spec.Scheduler.Cluster.ClusterVersion,
					Namespace:      fsPrivate.Spec.Scheduler.Namespace.Name,
					FilesystemName: fsPrivate.Metadata.Name,
					VolumePath:     fsPrivate.Spec.VolumePath,
				}
				// Ensure fsPublic.Status.SecurityGroup is initialized
				if fsPublic.Status.SecurityGroup == nil {
					fsPublic.Status.SecurityGroup = &pb.VolumeSecurityGroup{
						NetworkFilterAllow: []*pb.VolumeNetworkGroup{},
					}
				}
				// Ensure fsPrivate.Spec.SecurityGroup is initialized
				if fsPrivate.Spec.SecurityGroup == nil {
					fsPrivate.Spec.SecurityGroup = &pb.VolumeSecurityGroup{
						NetworkFilterAllow: []*pb.VolumeNetworkGroup{},
					}
				}
				fsPublic.Status.SecurityGroup.NetworkFilterAllow = append(fsPublic.Status.SecurityGroup.NetworkFilterAllow, fsPrivate.Spec.SecurityGroup.NetworkFilterAllow...)
			}
		}
	}
	return &fsPublic
}

func convertFilesystemCreatePublicToPrivate(fsCreatePublic *pb.FilesystemCreateRequest) *pb.FilesystemCreateRequestPrivate {
	fsCreatePrivate := &pb.FilesystemCreateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			CloudAccountId:   fsCreatePublic.Metadata.CloudAccountId,
			Name:             fsCreatePublic.Metadata.Name,
			Description:      fsCreatePublic.Metadata.Description,
			Labels:           fsCreatePublic.Metadata.Labels,
			SkipQuotaCheck:   false,
			SkipProductCheck: false,
			ClientType:       pb.APIClientTypes_Public,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: fsCreatePublic.Spec.AvailabilityZone,
			Request:          fsCreatePublic.Spec.Request,
			StorageClass:     fsCreatePublic.Spec.StorageClass,
			AccessModes:      pb.FilesystemAccessModes_ReadWrite,
			MountProtocol:    pb.FilesystemMountProtocols_NFS,
			Encrypted:        fsCreatePublic.Spec.Encrypted,
			FilesystemType:   fsCreatePublic.Spec.FilesystemType,
			VolumePath:       utils.GenerateVASTVolumePath(fsCreatePublic.Metadata.CloudAccountId, fsCreatePublic.Metadata.Name),
		},
	}

	return fsCreatePrivate
}

func convertFilesystemUpdatePublicToPrivate(fsUpdatePublic *pb.FilesystemUpdateRequest) *pb.FilesystemUpdateRequestPrivate {
	fsCreatePrivate := &pb.FilesystemUpdateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			CloudAccountId:   fsUpdatePublic.Metadata.CloudAccountId,
			Name:             fsUpdatePublic.Metadata.GetName(),
			ResourceId:       fsUpdatePublic.Metadata.GetResourceId(),
			Labels:           fsUpdatePublic.Metadata.Labels,
			SkipQuotaCheck:   false,
			SkipProductCheck: false,
			ClientType:       pb.APIClientTypes_Public,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: fsUpdatePublic.Spec.AvailabilityZone,
			Request:          fsUpdatePublic.Spec.Request,
			StorageClass:     fsUpdatePublic.Spec.StorageClass,
			AccessModes:      pb.FilesystemAccessModes_ReadWrite,
			MountProtocol:    pb.FilesystemMountProtocols_Weka,
			Encrypted:        fsUpdatePublic.Spec.Encrypted,
			FilesystemType:   fsUpdatePublic.Spec.FilesystemType,
		},
	}

	return fsCreatePrivate
}

func convertBucketPrivateToPublic(bucketPrivate *pb.ObjectBucketPrivate, updateStatus bool) *pb.ObjectBucket {
	bucketPublic := pb.ObjectBucket{
		Metadata: &pb.ObjectBucketMetadata{
			CloudAccountId:    bucketPrivate.Metadata.CloudAccountId,
			Name:              bucketPrivate.Metadata.Name,
			ResourceId:        bucketPrivate.Metadata.ResourceId,
			ResourceVersion:   bucketPrivate.Metadata.ResourceVersion,
			Description:       bucketPrivate.Metadata.Description,
			Labels:            bucketPrivate.Metadata.Labels,
			CreationTimestamp: bucketPrivate.Metadata.CreationTimestamp,
			DeletionTimestamp: bucketPrivate.Metadata.DeletionTimestamp,
		},
		Spec: &pb.ObjectBucketSpec{
			AvailabilityZone: bucketPrivate.Spec.AvailabilityZone,
			Request:          bucketPrivate.Spec.Request,
			AccessPolicy:     bucketPrivate.Spec.AccessPolicy,
			Versioned:        bucketPrivate.Spec.Versioned,
		},
		Status: &pb.ObjectBucketStatus{
			Phase:         bucketPrivate.Status.Phase,
			Message:       bucketPrivate.Status.Message,
			Policy:        bucketPrivate.Status.Policy,
			SecurityGroup: bucketPrivate.Status.SecurityGroup,
		},
	}
	if updateStatus {
		// TODO: We need to update this status from private instance status
		// This needs change in operator to update the same in the object store object
		// until that fix is ready, use this option
		bucketPublic.Status.Cluster = &pb.ObjectCluster{
			ClusterId:      bucketPrivate.Spec.Schedule.Cluster.ClusterUUID,
			AccessEndpoint: bucketPrivate.Spec.Schedule.Cluster.ClusterAddr,
			ClusterName:    bucketPrivate.Spec.Schedule.Cluster.ClusterName,
		}
	}
	return &bucketPublic
}

func convertBucketCreateReqPublicToPrivate(publicReq *pb.ObjectBucketCreateRequest) *pb.ObjectBucketCreatePrivateRequest {
	bucketCreatePrivate := &pb.ObjectBucketCreatePrivateRequest{
		Metadata: &pb.ObjectBucketMetadataPrivate{
			CloudAccountId: publicReq.Metadata.CloudAccountId,
			Name:           publicReq.Metadata.Name,
			Description:    publicReq.Metadata.Description,
			Labels:         publicReq.Metadata.Labels,
		},
		Spec: &pb.ObjectBucketSpecPrivate{
			AvailabilityZone: publicReq.Spec.AvailabilityZone,
			Request:          publicReq.Spec.Request,
			Versioned:        publicReq.Spec.Versioned,
			AccessPolicy:     publicReq.Spec.AccessPolicy,
		},
	}

	return bucketCreatePrivate
}

func convertObjectUserPrivateToPublic(userPrivate *pb.ObjectUserPrivate) *pb.ObjectUser {
	publicUser := pb.ObjectUser{
		Metadata: (*pb.ObjectUserMetadata)(userPrivate.Metadata),
		Spec:     userPrivate.Spec,
		Status: &pb.ObjectUserStatus{
			Phase: userPrivate.Status.Phase,
			Principal: &pb.AccessPrincipal{
				Cluster:     (*pb.ObjectCluster)(userPrivate.Status.Principal.Cluster),
				Credentials: (*pb.ObjectAccessCredentials)(userPrivate.Status.Principal.Credentials),
			},
		},
	}
	return &publicUser
}

func convertLifecycleRulePrivateToPublic(lfrulePrivate *pb.BucketLifecycleRulePrivate) *pb.BucketLifecycleRule {
	publicRule := pb.BucketLifecycleRule{
		Metadata: (*pb.BucketLifecycleRuleMetadata)(lfrulePrivate.Metadata),
		Spec:     (*pb.BucketLifecycleRuleSpec)(lfrulePrivate.Spec),
		Status:   lfrulePrivate.Status,
	}

	return &publicRule
}
