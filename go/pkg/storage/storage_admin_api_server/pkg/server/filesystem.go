// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	qs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	idcutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const clusterTypeMinio = "Minio"

// StorageAdminServiceServer is used to implement pb.UnimplementedStorageAdminServiceServer
type StorageAdminServiceClient struct {
	pb.UnimplementedStorageAdminServiceServer
	pb.UnimplementedQuotaManagementPrivateServiceServer

	session                   *sql.DB
	cloudAccountServiceClient pb.CloudAccountServiceClient
	quotaServiceClient        *qs.QuotaService
	filesystemClient          pb.FilesystemPrivateServiceClient
	strCntClient              *storagecontroller.StorageControllerClient
	customQuotaMaxAllowedInTB int64
	maxVolumesAllowed         int64
	maxBucketsAllowed         int64
	selectedRegion            string
}

type sourceBucketInfo struct {
	Region           string
	AccountType      string
	Email            string
	ClusterScheduled string
}

func NewStorageAdminServiceClient(ctx context.Context, session *sql.DB, cloudAccountSvc pb.CloudAccountServiceClient,
	quotaServiceClient *qs.QuotaService, filesystemSvcClient pb.FilesystemPrivateServiceClient,
	strCntClientSvcClient *storagecontroller.StorageControllerClient, customQuotaMaxAllowedInTB int64, selectedRegion string,
	maxVolumesAllowed int64, maxBucketsAllowed int64) (*StorageAdminServiceClient, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}

	fsSrv := StorageAdminServiceClient{
		session:                   session,
		cloudAccountServiceClient: cloudAccountSvc,
		quotaServiceClient:        quotaServiceClient,
		filesystemClient:          filesystemSvcClient,
		strCntClient:              strCntClientSvcClient,
		customQuotaMaxAllowedInTB: customQuotaMaxAllowedInTB,
		selectedRegion:            selectedRegion,
		maxVolumesAllowed:         maxVolumesAllowed,
		maxBucketsAllowed:         maxBucketsAllowed,
	}

	return &fsSrv, nil
}

// GetResourceUsage implements the GetResourceUsage method of the StorageAdminService.
func (s *StorageAdminServiceClient) GetResourceUsage(ctx context.Context, req *emptypb.Empty) (*pb.StorageGetResourceUsageResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageAdminService.GetResourceUsage").Start()
	defer span.End()
	logger.Info("entering resource usage get")
	defer logger.Info("returning from resource usage get")

	allFilesystemsUsages, cloudAcccloudAccountFilesystemUsages, err := s.getAllFilesystemUsages(ctx)
	if err != nil {
		return nil, err
	}

	allBucketsUsages, err := s.getAllBucketUsages(ctx, cloudAcccloudAccountFilesystemUsages)
	if err != nil {
		return nil, err
	}

	logger.Info("usage response created")
	usageResp := &pb.StorageGetResourceUsageResponse{
		FilesystemUsages: allFilesystemsUsages,
		BucketUsages:     allBucketsUsages,
	}
	log.LogResponseOrError(logger, nil, usageResp, err)
	return usageResp, nil
}

// InsertStorageQuotaByAccount implements the InsertStorageQuotaByAccount method of the StorageAdminService.
func (s *StorageAdminServiceClient) InsertStorageQuotaByAccount(ctx context.Context, req *pb.InsertStorageQuotaByAccountRequest) (*pb.StorageQuotaByAccount, error) {
	// Start a new transaction
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.createFilesystem").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering insert storage quota")
	defer logger.Info("returning from insert quota creation")
	if idcutils.IsValidCloudAccountId(req.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for inserting storage quota")
	}
	if err := s.validateInsertStorageQuotaRequest(req); err != nil {
		return nil, err
	}

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}
	defer tx.Rollback()
	cloudAccountInfo, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudaccount info: %v", err)
	}
	// Call the InsertStorageQuotaByAccount function from the QuotaService
	quota, err := s.quotaServiceClient.InsertStorageQuotaByAccount(ctx, tx, req.CloudAccountId, GetAccountType(cloudAccountInfo.Type), req.Reason, req.FilesizeQuotaInTB, req.FilevolumesQuota, req.BucketsQuota)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}
	logger.Info("quota inserted is : ", "quota in db : ", quota)
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}

	// Convert the quota to the protobuf message type
	pbQuota := &pb.StorageQuotaByAccount{
		CloudAccountType:  quota.CloudAccountType,
		CloudAccountId:    quota.CloudAccountId,
		Reason:            quota.Reason,
		FilesizeQuotaInTB: quota.FilesizeQuotaInTB,
		FilevolumesQuota:  quota.FilevolumesQuota,
		BucketsQuota:      quota.BucketsQuota,
	}
	logger.Info("response quota is ", "response from db : ", pbQuota)

	return pbQuota, nil
}

// UpdateStorageQuotaByAccount implements the UpdateStorageQuotaByAccount method of the StorageAdminService.
func (s *StorageAdminServiceClient) UpdateStorageQuotaByAccount(ctx context.Context, req *pb.UpdateStorageQuotaByAccountRequest) (*pb.StorageQuotaByAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.UpdateStorageQuotaByAccount").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering update storage quota")
	defer logger.Info("returning from update quota")

	// Start a new transaction
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}
	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function

	if idcutils.IsValidCloudAccountId(req.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for updating storage quota")
	}
	if err := s.validateUpdateStorageQuotaRequest(req); err != nil {
		return nil, err
	}
	cloudAccountInfo, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cloud account fetch failed: %v", err)
	}
	// Call the UpdateStorageQuotaByAccount function from the QuotaService
	quota, err := s.quotaServiceClient.UpdateStorageQuotaByAccount(ctx, tx, req.CloudAccountId, GetAccountType(cloudAccountInfo.Type), req.Reason, req.FilesizeQuotaInTB, req.FilevolumesQuota, req.BucketsQuota)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)

	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "database commit failed: %v", err)
	}

	// Convert the quota to the protobuf message type
	pbQuota := &pb.StorageQuotaByAccount{
		CloudAccountType:  quota.CloudAccountType,
		CloudAccountId:    quota.CloudAccountId,
		Reason:            quota.Reason,
		FilesizeQuotaInTB: quota.FilesizeQuotaInTB,
		FilevolumesQuota:  quota.FilevolumesQuota,
		BucketsQuota:      quota.BucketsQuota,
	}

	return pbQuota, nil
}

// DeleteStorageQuotaByAccount implements the DeleteStorageQuotaByAccount method of the StorageAdminService.
func (s *StorageAdminServiceClient) DeleteStorageQuotaByAccount(ctx context.Context, req *pb.DeleteStorageQuotaByAccountRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.DeleteStorageQuotaByAccount").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering delete storage quota")
	defer logger.Info("returning from delete quota")
	return nil, status.Errorf(codes.Unimplemented, "method delete not implemented")
	// Start a new transaction
	// tx, err := s.session.BeginTx(ctx, nil)
	// if err != nil {
	// 	return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	// }
	// defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function

	// Call the DeleteStorageQuotaByAccount function from the QuotaService
	// err = s.quotaServiceClient.DeleteStorageQuotaByAccount(ctx, tx, req.CloudAccountId)
	// if err != nil {
	// 	return nil, status.Errorf(codes.Internal, "database delete storage account failed: %v", err)
	// }

	// Commit the transaction
	// if err := tx.Commit(); err != nil {
	// 	return nil, status.Errorf(codes.Internal, "database commit failed: %v", err)
	// }
	// return &emptypb.Empty{}, nil
}

func (s *StorageAdminServiceClient) GetStorageQuotaByAccount(ctx context.Context, req *pb.GetStorageQuotaByAccountRequest) (*pb.StorageQuotasByAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetStorageQuotaByAccount").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering get storage quota")
	defer logger.Info("returning from get quota")

	if idcutils.IsValidCloudAccountId(req.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for getting storage quota")
	}
	// Start a new transaction
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}
	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function
	cloudAccountInfo, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get cloud account by ID failed: %v", err)
	}
	accountType := GetAccountType(cloudAccountInfo.GetType())
	defaultQuota, updatedQuota, err := s.quotaServiceClient.GetStorageQuotaByAccount(ctx, tx, req.CloudAccountId, accountType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get storage quota by account failed: %v", err)
	}
	var pbUpdatedQuota *pb.StorageQuotaByAccount

	pbDefaultQuota := &pb.StorageQuotaByAccount{
		CloudAccountId:    defaultQuota.CloudAccountId,
		CloudAccountType:  defaultQuota.CloudAccountType,
		FilesizeQuotaInTB: defaultQuota.FilesizeQuotaInTB,
		FilevolumesQuota:  defaultQuota.FilevolumesQuota,
		BucketsQuota:      defaultQuota.BucketsQuota,
		IsDefault:         true,
	}
	if updatedQuota != nil {
		pbUpdatedQuota = &pb.StorageQuotaByAccount{
			CloudAccountId:    updatedQuota.CloudAccountId,
			CloudAccountType:  updatedQuota.CloudAccountType,
			Reason:            updatedQuota.Reason,
			FilesizeQuotaInTB: updatedQuota.FilesizeQuotaInTB,
			FilevolumesQuota:  updatedQuota.FilevolumesQuota,
			BucketsQuota:      updatedQuota.BucketsQuota,
			IsDefault:         false,
		}
	}
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "database commit failed: %v", err)
	}

	return &pb.StorageQuotasByAccount{
		DefaultQuota: pbDefaultQuota,
		UpdatedQuota: pbUpdatedQuota,
	}, nil
}

func GetAccountType(typ pb.AccountType) string {
	switch typ {
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return "STANDARD"
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		return "PREMIUM"
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return "ENTERPRISE"
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING:
		return "ENTERPRISE_PENDING"
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return "INTEL"
	default:
		return "STANDARD"
	}
}

func (s *StorageAdminServiceClient) GetAllStorageQuota(ctx context.Context, req *emptypb.Empty) (*pb.StorageQuotas, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetAllStorageQuota").Start()
	defer span.End()
	logger.Info("entering get all storage quota")
	defer logger.Info("returning from get all quota")

	// Start a new transaction
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}
	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function

	// Get all storage quotas
	quotas, err := s.quotaServiceClient.GetAllStorageQuota(ctx, tx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get all storage quota failed: %v", err)
	}

	// Convert quotas to []*pb.StorageQuotaByAccount
	pbQuotas := make([]*pb.StorageQuotaByAccount, len(quotas))
	for i, quota := range quotas {
		pbQuotas[i] = &pb.StorageQuotaByAccount{
			CloudAccountId:    quota.CloudAccountId,
			CloudAccountType:  quota.CloudAccountType,
			Reason:            quota.Reason,
			FilesizeQuotaInTB: quota.FilesizeQuotaInTB,
			FilevolumesQuota:  quota.FilevolumesQuota,
			BucketsQuota:      quota.BucketsQuota,
			IsDefault:         false,
		}
	}

	// Prepare the response
	response := &pb.StorageQuotas{
		StorageQuotaByAccount: pbQuotas,
	}

	// Get the default quota for each account type
	accountTypes := []pb.AccountType{
		pb.AccountType_ACCOUNT_TYPE_STANDARD,
		pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
		pb.AccountType_ACCOUNT_TYPE_INTEL,
	}

	for _, accountType := range accountTypes {
		defaultQuota, err := s.quotaServiceClient.GetStorageQuotaByType(ctx, tx, accountType)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "quota service failed to get quota by type: %v", err)
		}

		// Convert defaultQuota to *pb.DefaultQuotaSection
		pbDefaultQuota := &pb.DefaultQuotaSection{
			CloudAccountType:  defaultQuota.CloudAccountType,
			FilesizeQuotaInTB: defaultQuota.FilesizeQuotaInTB,
			FilevolumesQuota:  defaultQuota.FilevolumesQuota,
			BucketsQuota:      defaultQuota.BucketsQuota,
		}

		// Add the default quota to the response
		response.DefaultQuotaSection = append(response.DefaultQuotaSection, pbDefaultQuota)
	}
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed: %v", err)
	}

	return response, nil
}

// helper function to get filesystem usages for all cloudaccounts
func (s *StorageAdminServiceClient) getAllFilesystemUsages(ctx context.Context) ([]*pb.StorageFilesystemUsageResponse, map[string]*pb.StorageFilesystemUsageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageAdminService.getAllFilesystemUsages").Start()
	defer span.End()

	var allFilesystemsUsages []*pb.StorageFilesystemUsageResponse
	cloudAccountFilesystemUsages := make(map[string]*pb.StorageFilesystemUsageResponse)

	searchReq := pb.FilesystemSearchStreamPrivateRequest{ResourceVersion: "0", AvailabilityZone: "az1"}
	fsStream, err := s.filesystemClient.SearchFilesystemRequests(ctx, &searchReq)
	if err != nil {
		return nil, nil, err
	}
	var fsReq *pb.FilesystemRequestResponse
	for {
		fsReq, err = fsStream.Recv()
		if err == io.EOF {
			logger.Info("stream EOF")
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
		logger.Info("handle filesystem request", logkeys.Filesystem, fsReq, logkeys.Filesystem, fsReq.Filesystem.Metadata.ResourceVersion)
		currCloudAccountId := fsReq.Filesystem.Metadata.CloudAccountId
		// Cloud account already seen, fetching info from the map
		if fsUsage, ok := cloudAccountFilesystemUsages[currCloudAccountId]; ok {
			numFs, err := strconv.Atoi(fsUsage.NumFilesystems)
			if err != nil {
				logger.Error(err, "failed to convert number of filesystem to int", "cloudaccount %s", currCloudAccountId)
			}
			cloudAccountFilesystemUsages[currCloudAccountId].NumFilesystems = strconv.Itoa(numFs + 1)
			newProvisioned, addError := utils.AddSizes(fsUsage.TotalProvisioned, fsReq.Filesystem.Spec.Request.Storage)
			if addError != nil {
				logger.Error(err, "failed to add file sizes")
			} else {
				cloudAccountFilesystemUsages[currCloudAccountId].TotalProvisioned = newProvisioned

				// override only if no IKS volumes have been found so far, for this cloud account
				if cloudAccountFilesystemUsages[currCloudAccountId].HasIksVolumes == "No" {
					cloudAccountFilesystemUsages[currCloudAccountId].HasIksVolumes = checkForIksVolumes(fsReq)
				}
			}
		} else {
			// Cloud account seen for the first time, fetching cloudaccount info
			cloudAccountInfo, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: currCloudAccountId})
			if err != nil {
				logger.Error(err, "failed to get cloudaccount, skipping this file", "filename:", fsReq.Filesystem.Metadata.Name, "cloudaccountid:", currCloudAccountId)
			} else {
				newFsUsage := &pb.StorageFilesystemUsageResponse{
					CloudAccountId:   currCloudAccountId,
					Region:           s.selectedRegion,
					AccountType:      cloudAccountInfo.Type.String(),
					Email:            cloudAccountInfo.Name,
					OrgId:            fsReq.Filesystem.Spec.Scheduler.Namespace.Name,
					NumFilesystems:   "1",
					TotalProvisioned: fsReq.Filesystem.Spec.Request.Storage,
					ClusterScheduled: fsReq.Filesystem.Spec.Scheduler.Cluster.ClusterAddr,
					HasIksVolumes:    checkForIksVolumes(fsReq),
				}
				if fsReq.Filesystem.Spec.FilesystemType == pb.FilesystemType_ComputeKubernetes { //HasIKS == ComputeKubernetes
					newFsUsage.HasIksVolumes = "Yes"
				}
				cloudAccountFilesystemUsages[currCloudAccountId] = newFsUsage
			}
		}
	}
	for _, usage := range cloudAccountFilesystemUsages {
		allFilesystemsUsages = append(allFilesystemsUsages, usage)
	}
	logger.Info("completed getting filesystem usages", logkeys.FilesystemCount, len(allFilesystemsUsages))
	return allFilesystemsUsages, cloudAccountFilesystemUsages, nil
}

// helper function to get filesystem usages for all cloudaccounts
func (s *StorageAdminServiceClient) getAllBucketUsages(ctx context.Context, cloudAccountFilesystemUsages map[string]*pb.StorageFilesystemUsageResponse) ([]*pb.StorageBucketUsageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageAdminService.getAllBucketUsages").Start()
	defer span.End()
	var allBucketsUsages []*pb.StorageBucketUsageResponse
	cloudAccountBucketUsages := make(map[string]*pb.StorageBucketUsageResponse)

	minioClusters, err := s.filterClusters(ctx, clusterTypeMinio)
	if err != nil {
		logger.Error(err, "error getting minio cluster info from sds")
	}
	for _, cluster := range minioClusters {
		//get all buckets for cluster
		buckets, err := s.strCntClient.GetAllBuckets(ctx, storagecontroller.BucketFilter{ClusterId: cluster.UUID})
		if err != nil {
			logger.Error(err, "failed to get buckets for cluster", "cluster id:", cluster.UUID)
		}
		logger.Info("Successfully retrieved buckets from minio", logkeys.ClusterName, cluster.Name, logkeys.BucketCount, len(buckets))
		for _, bucket := range buckets {
			name := bucket.Metadata.BucketId
			// Check if bucket is prefixed by cloudaccount
			if len(name) > 13 {
				cloudaccount := name[:12]
				_, err := strconv.Atoi(cloudaccount)
				if err != nil {
					logger.Error(err, "failed to convert cloud account from bucket")
				}
				var bucketInfo *sourceBucketInfo
				if _, ok := cloudAccountFilesystemUsages[cloudaccount]; ok {
					// cloud account already found during filesystem usage so reuse it
					logger.Info("cloud account retrieved", logkeys.CloudAccountId, cloudaccount)
					bucketInfo = &sourceBucketInfo{
						Region:           cloudAccountFilesystemUsages[cloudaccount].Region,
						AccountType:      cloudAccountFilesystemUsages[cloudaccount].AccountType,
						Email:            cloudAccountFilesystemUsages[cloudaccount].Email,
						ClusterScheduled: cloudAccountFilesystemUsages[cloudaccount].ClusterScheduled,
					}
				} else {
					// cloud account found for the first time so retrieve it
					cloudAccountBucket, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: cloudaccount})
					if err != nil {
						logger.Error(err, "failed to get cloudaccount", "cloudaccount", cloudaccount)
					} else {
						bucketInfo = &sourceBucketInfo{
							Region:           s.selectedRegion,
							AccountType:      cloudAccountBucket.Type.String(),
							Email:            cloudAccountBucket.Name,
							ClusterScheduled: removeHTTPSAndPort(cluster.S3Endpoint),
						}
					}
				}

				// if bucket info was found then process it
				if bucketInfo != nil {
					logger.Info("cloud account retrieved", logkeys.CloudAccountId, cloudaccount)
					if bucketUsage, ok := cloudAccountBucketUsages[cloudaccount]; ok {
						numBk, err := strconv.Atoi(bucketUsage.Buckets)
						if err != nil {
							logger.Error(err, "failed to convert number of buckets to int", "cloudaccount", cloudaccount)
						}
						cloudAccountBucketUsages[cloudaccount].Buckets = strconv.Itoa(numBk + 1)

						// add to bucket size
						currTotalBucketSize := cloudAccountBucketUsages[cloudaccount].BucketSize
						newTotalBucketSize, addError := utils.AddSizes(currTotalBucketSize, utils.ConvertBytesToGBOrTB(bucket.Spec.Totalbytes))
						if addError != nil {
							// soft error continue to next bucket
							logger.Error(addError, "failed to add bucket sizes")
						} else {
							cloudAccountBucketUsages[cloudaccount].BucketSize = newTotalBucketSize

							// add to bucket used capacity
							currTotalUsedBucket := cloudAccountBucketUsages[cloudaccount].UsedCapacity
							newTotalUsedBucket, addError := utils.AddSizes(currTotalUsedBucket, utils.ConvertBytesToGBOrTB(getUsedCapacity(bucket)))
							if addError != nil {
								// soft error continue to next bucket
								logger.Error(addError, "failed to add used bucket capacities")
							} else {
								cloudAccountBucketUsages[cloudaccount].UsedCapacity = newTotalUsedBucket
							}
						}
					} else {
						newBucketUsage := &pb.StorageBucketUsageResponse{
							Region:           bucketInfo.Region,
							CloudAccountId:   cloudaccount,
							AccountType:      bucketInfo.AccountType,
							Email:            bucketInfo.Email,
							Buckets:          "1",
							UsedCapacity:     utils.ConvertBytesToGBOrTB(getUsedCapacity(bucket)),
							BucketSize:       utils.ConvertBytesToGBOrTB(bucket.Spec.Totalbytes),
							ClusterScheduled: bucketInfo.ClusterScheduled,
						}
						cloudAccountBucketUsages[cloudaccount] = newBucketUsage
					}
				}
			}
		}
	}
	for _, usage := range cloudAccountBucketUsages {
		allBucketsUsages = append(allBucketsUsages, usage)
	}
	logger.Info("completed getting all bucket usages", logkeys.BucketCount, len(allBucketsUsages))
	return allBucketsUsages, nil
}

// helper function to filter clusters based on type
func (s *StorageAdminServiceClient) filterClusters(ctx context.Context, clusterType string) ([]storagecontroller.ClusterInfo, error) {
	logger := log.FromContext(ctx).WithName("StorageAdminServiceClient.filterClusters")
	// get all clusters
	var clusterList []storagecontroller.ClusterInfo
	clusters, err := s.strCntClient.GetClusters(ctx)
	if err != nil {
		return clusterList, err
	}
	// filter clusters for specified type
	for _, cluster := range clusters {
		if cluster.Type == clusterType {
			logger.Info("cluster type", logkeys.ClusterType, cluster.Type)
			clusterList = append(clusterList, cluster)
		}
	}
	return clusterList, nil
}

// helper function to get used capacity for a bucket
func getUsedCapacity(bucket *storagecontroller.Bucket) uint64 {
	currentUsage := uint64(0)
	if bucket.Spec.AvailableBytes != 0 {
		currentUsage = bucket.Spec.Totalbytes - bucket.Spec.AvailableBytes
	}
	return currentUsage
}

// helper function that returns "Yes" if IKS volumes are present else returns "No"
func checkForIksVolumes(fsReq *pb.FilesystemRequestResponse) string {
	if fsReq.Filesystem.Spec.FilesystemType == pb.FilesystemType_ComputeKubernetes {
		return "Yes"
	}
	return "No"
}

// helper function get name cluster scheduled name from the endpoint
func removeHTTPSAndPort(rawURL string) string {
	// Parse the URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Remove the scheme (https)
	u.Scheme = ""

	// Split host into hostname and port
	host := u.Hostname()

	return host
}

func (s *StorageAdminServiceClient) validateInsertStorageQuotaRequest(req *pb.InsertStorageQuotaByAccountRequest) error {
	if req.FilesizeQuotaInTB > s.customQuotaMaxAllowedInTB {
		return status.Errorf(codes.InvalidArgument, "quota being added for file size is greater than allowed quota")
	}
	if req.FilevolumesQuota > s.maxVolumesAllowed {
		return status.Errorf(codes.InvalidArgument, "quota being added for volumes is greater than allowed quota")
	}
	if req.BucketsQuota > s.maxBucketsAllowed {
		return status.Errorf(codes.InvalidArgument, "quota being added for buckets is greater than allowed quota")
	}
	return nil
}

func (s *StorageAdminServiceClient) validateUpdateStorageQuotaRequest(req *pb.UpdateStorageQuotaByAccountRequest) error {
	if req.FilesizeQuotaInTB > s.customQuotaMaxAllowedInTB {
		return status.Errorf(codes.InvalidArgument, "quota being updated for file size is greater than allowed quota")
	}
	if req.FilevolumesQuota > s.maxVolumesAllowed {
		return status.Errorf(codes.InvalidArgument, "quota being updated for volumes is greater than allowed quota")
	}
	if req.BucketsQuota > s.maxBucketsAllowed {
		return status.Errorf(codes.InvalidArgument, "quota being updated for buckets is greater than allowed quota")
	}
	return nil
}
