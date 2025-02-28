// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// StorageSchedulerServiceServer is used to implement pb.UnimplementedFileStorageServiceServer
type StorageSchedulerServiceServer struct {
	pb.UnimplementedFilesystemSchedulerPrivateServiceServer
	pb.UnimplementedWekaStatefulAgentPrivateServiceServer
	kmsClient v1.StorageKMSPrivateServiceClient

	StrCntClient  *storagecontroller.StorageControllerClient
	gpVASTEnabled bool
}

type ClusterAssigner struct {
	shuffledClusters []storagecontroller.ClusterInfo
	currentIndex     int
}

func NewClusterAssigner() *ClusterAssigner {
	return &ClusterAssigner{}
}

func NewStorageSchedulerService(client *storagecontroller.StorageControllerClient, kmsClient v1.StorageKMSPrivateServiceClient, gpVASTEnabled bool) (*StorageSchedulerServiceServer, error) {
	if client == nil {
		return nil, fmt.Errorf("storage client is required")
	}
	return &StorageSchedulerServiceServer{
		StrCntClient:  client,
		kmsClient:     kmsClient,
		gpVASTEnabled: gpVASTEnabled,
	}, nil
}

func (scheduler *StorageSchedulerServiceServer) ScheduleFile(ctx context.Context, in *pb.FilesystemScheduleRequest) (*pb.FilesystemScheduleResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageSchedulerServiceServer.ScheduleFile").WithValues(logkeys.CloudAccountId, in.CloudaccountId).Start()
	defer span.End()
	logger.Info("input filesystem specs", logkeys.RequestSpec, in.RequestSpec, logkeys.StorageClass, in.RequestSpec.StorageClass)
	logger.Info("discovering cluster states")
	reqSize := utils.ParseFileSize(in.RequestSpec.Request.Storage)
	if reqSize == -1 {
		logger.Info("request size could not be parsed", "size", in.RequestSpec.Request.Storage)
		return nil, fmt.Errorf("failed to assign cluster due to incorrect size")
	}

	clusters, err := scheduler.StrCntClient.GetClusters(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error querying cluster info")
	}
	logger.Info("all available cluster info ", logkeys.Clusters, clusters)

	onlineWekaFastClusters, onlineWekaGPClusters, vastClusters := scheduler.filterClusters(clusters, reqSize)

	logger.Info("input args to assign cluster ", logkeys.Input, in)

	//Assigning clusters based on most available capacity and storageClass requested
	if in.RequestSpec.StorageClass == pb.FilesystemStorageClass_AIOptimized {
		logger.Info("online available weka fast cluster ", logkeys.Clusters, onlineWekaFastClusters)
		return scheduler.assignCluster(onlineWekaFastClusters, in)
	} else if in.RequestSpec.StorageClass == pb.FilesystemStorageClass_GeneralPurpose {
		logger.Info("online available weka gp cluster ", logkeys.Clusters, onlineWekaGPClusters)
		return scheduler.assignCluster(onlineWekaGPClusters, in)
	} else if in.RequestSpec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
		// if scheduler.gpVASTEnabled {
		logger.Info("online available vast gp cluster ", logkeys.Clusters, vastClusters)
		return scheduler.assignCluster(vastClusters, in)
		// } else {
		// 	logger.Info("online available weka gp cluster ", logkeys.Clusters, onlineWekaGPClusters)
		// 	return scheduler.assignCluster(onlineWekaGPClusters, in)
		// }
	}
	return nil, fmt.Errorf("no valid storage class found")
}

func (scheduler *StorageSchedulerServiceServer) filterClusters(clusters []storagecontroller.ClusterInfo, reqSize int64) ([]storagecontroller.ClusterInfo, []storagecontroller.ClusterInfo, []storagecontroller.ClusterInfo) {
	onlineWekaFastClusters := []storagecontroller.ClusterInfo{}
	onlineWekaGeneralPurposeClusters := []storagecontroller.ClusterInfo{}
	onlineVASTGeneralPurposeClusters := []storagecontroller.ClusterInfo{}
	for _, cl := range clusters {
		if cl.Type == "Weka" && cl.Status == "Online" && cl.AvailableCapacity > reqSize {
			if cl.Category == "ai-storage" || cl.Category == "development" {
				if cl.ClusterLabelType == "fast" {
					onlineWekaFastClusters = append(onlineWekaFastClusters, cl)
				} else {
					onlineWekaGeneralPurposeClusters = append(onlineWekaGeneralPurposeClusters, cl)
				}
			}
		} else if cl.Type == "VAST" && cl.Status == "Online" && cl.AvailableCapacity > reqSize {
			onlineVASTGeneralPurposeClusters = append(onlineVASTGeneralPurposeClusters, cl)
		}
	}
	return onlineWekaFastClusters, onlineWekaGeneralPurposeClusters, onlineVASTGeneralPurposeClusters
}

func (scheduler *StorageSchedulerServiceServer) findExistingSchedule(in *pb.FilesystemScheduleRequest, clusters []storagecontroller.ClusterInfo) *pb.ResourceSchedule {
	for _, assignment := range in.Assignments {
		for _, cl := range clusters {
			if assignment.ClusterUUID == cl.UUID {
				return &pb.ResourceSchedule{
					ClusterName:    assignment.ClusterName,
					ClusterAddr:    assignment.ClusterAddr,
					ClusterUUID:    assignment.ClusterUUID,
					ClusterVersion: "4.2.2", //FIXME: Update cluster versioning
					Namespace:      assignment.Namespace,
				}
			}
		}
	}
	return nil // return nil if no matching existing cluster is found
}

func (scheduler *StorageSchedulerServiceServer) assignCluster(clusters []storagecontroller.ClusterInfo, in *pb.FilesystemScheduleRequest) (*pb.FilesystemScheduleResponse, error) {
	resp := pb.FilesystemScheduleResponse{}
	if len(clusters) > 0 {
		existingSchedule := scheduler.findExistingSchedule(in, clusters)
		if existingSchedule != nil && existingSchedule.ClusterUUID != "" {
			resp.Schedule = existingSchedule
			return &resp, nil
		} else {
			sort.Slice(clusters, func(i, j int) bool {
				return clusters[i].AvailableCapacity > clusters[j].AvailableCapacity
			})
			assignedCluster := clusters[0]
			namespace := fmt.Sprintf("ns-%s", in.CloudaccountId)
			resp.Schedule = &pb.ResourceSchedule{
				ClusterName:    assignedCluster.Name,
				ClusterAddr:    assignedCluster.Addr,
				ClusterUUID:    assignedCluster.UUID,
				Namespace:      namespace,
				ClusterVersion: "", //FIXME: Update cluster versioning
			}
			resp.NewSchedule = true
			return &resp, nil
		}
	} else {
		return nil, fmt.Errorf("failed to assign cluster")
	}
}

func (scheduler *StorageSchedulerServiceServer) PingPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.PingPrivate")

	logger.Info("entering storage scheduler private Ping  ")
	defer logger.Info("returning from storage scheduler private Ping ")

	return &emptypb.Empty{}, nil
}
func (ca *ClusterAssigner) assignClusterForBucket(ctx context.Context, clusters []storagecontroller.ClusterInfo, reqSize int64) (*storagecontroller.ClusterInfo, error) {
	logger := log.FromContext(ctx).WithName("ClusterAssigner.assignClusterForBucket")
	logger.Info("assigning available cluster for Buckets ")

	// If shuffledClusters is empty or clusters have changed, shuffle the clusters
	if len(ca.shuffledClusters) == 0 || !areSlicesEqual(clusters, ca.shuffledClusters) {
		// Filter clusters of type "Minio" and status "Online"
		onlineMinioClusters := []storagecontroller.ClusterInfo{}
		onlineMinioNonGPClusters := []storagecontroller.ClusterInfo{}
		for _, cl := range clusters {
			if cl.Type == "Minio" && cl.Status == "Online" {
				if cl.ClusterLabelType == "gp" {
					onlineMinioClusters = append(onlineMinioClusters, cl)
				} else {
					onlineMinioNonGPClusters = append(onlineMinioNonGPClusters, cl)
				}
			}
		}

		if len(onlineMinioClusters) == 0 && len(onlineMinioNonGPClusters) == 0 {
			return nil, fmt.Errorf("no cluster available for scheduling")
		} else if len(onlineMinioClusters) == 0 {
			onlineMinioClusters = onlineMinioNonGPClusters
		}
		rand.Shuffle(len(onlineMinioClusters), func(i, j int) {
			onlineMinioClusters[i], onlineMinioClusters[j] = onlineMinioClusters[j], onlineMinioClusters[i]
		})

		// Update shuffledClusters and reset currentIndex
		ca.shuffledClusters = onlineMinioClusters
		ca.currentIndex = 0
	}

	// Select the current cluster from the shuffled list
	selectedCluster := &ca.shuffledClusters[ca.currentIndex]
	ca.currentIndex = (ca.currentIndex + 1) % len(ca.shuffledClusters)

	return selectedCluster, nil
}

func areSlicesEqual(a, b []storagecontroller.ClusterInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (scheduler *StorageSchedulerServiceServer) ScheduleBucket(ctx context.Context, in *pb.BucketScheduleRequest) (*pb.BucketScheduleResponse, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.ScheduleBucket")
	resp := pb.BucketScheduleResponse{}
	logger.Info("input bucket specs", logkeys.Input, in.RequestSpec.Request.Size)

	logger.Info("discovering cluster states  ")
	reqSize := utils.ParseFileSize(in.RequestSpec.Request.Size)
	if reqSize == -1 {
		logger.Info("request size could not be parsed", logkeys.Size, in.RequestSpec.Request.Size)
		return nil, fmt.Errorf("failed to assign cluster due to incorrect size")
	}
	clusters, err := scheduler.StrCntClient.GetClusters(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error querying cluster info")
	}
	logger.Info("all available cluster info ", logkeys.Clusters, clusters)
	assigner := NewClusterAssigner()
	assignedCluster, err := assigner.assignClusterForBucket(ctx, clusters, reqSize)
	if err != nil {
		return nil, fmt.Errorf("failed to assign cluster for bucket")
	}

	logger.Info("assigned cluster info ", logkeys.Cluster, assignedCluster)
	resp.AvailabilityZone = in.RequestSpec.AvailabilityZone
	resp.Schedule = &pb.BucketSchedule{
		Cluster: &pb.AssignedCluster{
			ClusterName: assignedCluster.Name,
			ClusterAddr: assignedCluster.S3Endpoint,
			ClusterUUID: assignedCluster.UUID,
		},
	}
	return &resp, nil
}

func (scheduler *StorageSchedulerServiceServer) ListClusters(in *pb.ListClusterRequest, rs pb.FilesystemSchedulerPrivateService_ListClustersServer) error {
	logger := log.FromContext(rs.Context()).WithName("StorageSchedulerServiceServer.ListClusters")
	logger.Info("input cluster filters", logkeys.Input, in)

	clusters, err := scheduler.StrCntClient.GetClusters(rs.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "error querying cluster info")
	}
	logger.Info("all available cluster info ", logkeys.Clusters, clusters)

	for _, cl := range clusters {
		labels := map[string]string{
			"VASTBackend":     cl.VASTBackend,
			"VastVMSEndpoint": cl.VastVMSEndpoint,
		}
		resp := pb.FilesystemStorageClusters{
			ClusterId:  cl.UUID,
			Name:       cl.Name,
			Location:   cl.DCLocation,
			VendorType: cl.Type,
			Capacity: &pb.StorageClusterCapacity{
				TotalBytes:     uint64(cl.AvailableCapacity),
				AvailableBytes: uint64(cl.AvailableCapacity),
			},
			Labels: labels,
			Health: cl.Status,
		}
		if err := rs.Send(&resp); err != nil {
			logger.Error(err, "error sending cluster info")
			return err
		}
	}
	return nil
}

func (scheduler *StorageSchedulerServiceServer) ListFilesystemInOrgs(ctx context.Context, in *pb.FilesystemInOrgGetRequestPrivate) (*pb.FilesystemsInOrgListResponsePrivate, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.ListFilesystemInOrgs")
	logger.Info("input list filesystems in org filters", logkeys.Input, in)
	request := pb.GetSecretRequest{
		KeyPath: in.NamespaceCredsPath,
	}
	secretResp, err := scheduler.kmsClient.Get(ctx, &request)
	if err != nil {
		logger.Error(err, "error fetch credentials from kms")
		return nil, status.Errorf(codes.Internal, "kms fetch transaction failed")
	}
	nsCreds := secretResp.Secrets
	if nsCreds == nil {
		logger.Error(err, "storageKms returned nil for secret")
		return nil, fmt.Errorf("storageKms returned nil for secret")
	}
	if nsCreds["username"] == "" || nsCreds["password"] == "" {
		logger.Error(err, "credentials cannot be empty")
		return nil, fmt.Errorf("empty credentials in vault secret")
	}
	Username := nsCreds["username"]
	Password := nsCreds["password"]
	queryParams := storagecontroller.FilesystemMetadata{
		NamespaceName: in.Name,
		User:          Username,
		Password:      Password,
		AuthRequired:  true,
		UUID:          in.ClusterId,
	}

	fileSystemList, exists, err := scheduler.StrCntClient.GetAllFileSystems(ctx, queryParams)
	if err != nil {
		logger.Error(err, "error fetching filesystem list")
		return nil, status.Errorf(codes.Internal, "error fetching filesystem list")
	}
	if !exists {
		return nil, nil
	}
	// Create a response with the retrieved filesystems.
	response := &pb.FilesystemsInOrgListResponsePrivate{
		Items: make([]*pb.FilesystemPrivate, 0, len(fileSystemList)),
	}

	for _, fs := range fileSystemList {
		// Convert each filesystem to FilesystemPrivate format.
		fsPrivate := &pb.FilesystemPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				Name:           fs.Metadata.FileSystemName,
				CloudAccountId: in.CloudAccountId,
			},
			Spec: &pb.FilesystemSpecPrivate{
				Request: &pb.FilesystemCapacity{
					Storage: fs.Properties.FileSystemCapacity,
				},
			},
		}
		response.Items = append(response.Items, fsPrivate)
	}

	return response, nil
}

func (scheduler *StorageSchedulerServiceServer) ListFilesystemOrgs(ctx context.Context, in *pb.FilesystemOrgsGetRequestPrivate) (*pb.FilesystemOrgsResponsePrivate, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.ListFilesystemOrgs")
	logger.Info("input list org filters", logkeys.Input, in)

	orgList, exists, err := scheduler.StrCntClient.GetAllFileSystemOrgs(ctx, in.ClusterId)
	if err != nil {
		logger.Error(err, "error fetching org list")
		return nil, status.Errorf(codes.Internal, "error fetching org list")
	}
	if !exists {
		return nil, nil
	}

	// Create a response with the retrieved org.
	response := &pb.FilesystemOrgsResponsePrivate{
		Org: make([]*pb.FilesystemOrgsPrivate, 0, len(orgList)),
	}

	for _, ns := range orgList {
		nsPrivate := &pb.FilesystemOrgsPrivate{
			Name: ns.Metadata.Name,
		}
		response.Org = append(response.Org, nsPrivate)
	}

	return response, nil
}

func (scheduler *StorageSchedulerServiceServer) IsOrgExists(ctx context.Context, in *pb.FilesystemOrgsIsExistsRequestPrivate) (*pb.FilesystemOrgsIsExistsResponsePrivate, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.IsOrgExists")
	logger.Info("input is org exists filters", logkeys.Input, in)

	nsQuery := storagecontroller.NamespaceMetadata{
		Name: in.Name,
		UUID: in.ClusterId,
	}
	exists, err := scheduler.StrCntClient.IsNamespaceExists(ctx, nsQuery)
	if err != nil {
		logger.Error(err, "error fetching org ")
		return &pb.FilesystemOrgsIsExistsResponsePrivate{Exists: true}, status.Errorf(codes.Internal, "error fetching org")
		// return true to avoid breaking existing clients (since this is used to delete the last existing record from accounts table)
	}
	return &pb.FilesystemOrgsIsExistsResponsePrivate{Exists: exists}, nil
}
