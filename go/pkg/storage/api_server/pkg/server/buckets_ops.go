package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (objSrv *BucketsServiceServer) createBucket(ctx context.Context, in *pb.ObjectBucketCreatePrivateRequest) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.createBucket")
	logger.Info("entering bucket creation for", logkeys.CloudAccountId, in.Metadata.CloudAccountId)
	defer logger.Info("returning from bucket creation")

	if err := isValidBucketCreateRequest(ctx, in, objSrv.objectProduct, in.Metadata.SkipProductCheck); err != nil {
		return nil, err
	}
	// Check quota for this cloudaccount
	cloudAccount, err := objSrv.cloudAccountServiceClient.GetById(ctx, &v1.CloudAccountId{Id: in.Metadata.CloudAccountId})
	if err != nil {
		logger.Error(err, "error querying cloudaccount")
		return nil, status.Errorf(codes.FailedPrecondition, "error querying cloudaccount")
	}

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Rollback()

	if !objSrv.quotaServiceClient.checkAndUpdateBucketQuota(ctx, cloudAccount) {
		return nil, status.Errorf(codes.FailedPrecondition, "quota check failed")
	}

	bucketPrivate := pb.ObjectBucketPrivate{
		Metadata: in.Metadata,
		Spec:     in.Spec,
	}

	vnetPrivate, err := query.GetBucketSubnetByAccount(ctx, tx, in.Metadata.CloudAccountId)
	if err != nil {
		// soft handling of the error here
		logger.Info("error reading bucket subnet from database")
	}
	if vnetPrivate == nil {
		// this is for backward compatibility only.
		// This would account for enabling access to compute instances, that were
		// already created
		vnet, err := objSrv.readVNetForCloudaccount(ctx, in.Metadata.CloudAccountId)
		if err != nil {
			// soft error handling
			logger.Info("error reading vnet info for cloudaccount")
		}
		vnetPrivate = &v1.VNetPrivate{
			Spec: &v1.VNetSpecPrivate{
				Subnet:       vnet.Subnet,
				PrefixLength: vnet.PrefixLength,
				Gateway:      vnet.Gateway,
			},
		}
	}
	if objSrv.quotaServiceClient.qmsClient != nil {
		// Retrieve cloudaccount bucket size from QMS
		res, err := objSrv.quotaServiceClient.qmsClient.GetResourceQuotaPrivate(ctx, &v1.ServiceQuotaResourceRequestPrivate{
			ServiceName:    storageServiceName,
			ResourceType:   bucketSizeQuota,
			CloudAccountId: in.Metadata.CloudAccountId,
		})
		if err != nil {
			logger.Error(err, "failed to retrieve bucket size quota for", logkeys.CloudAccount, in.Metadata.CloudAccountId)
			return nil, status.Errorf(codes.Internal, "failed to retrieve bucket size from quota management service")
		}
		var bkSize string
		if len(res.CustomQuota.ServiceResources) != 0 {
			bkSize = strconv.FormatInt(res.CustomQuota.ServiceResources[0].QuotaConfig.Limits*1000, 10)
		} else {
			bkSize = strconv.FormatInt(res.DefaultQuota.ServiceResources[0].QuotaConfig.Limits*1000, 10)
		}

		logger.Info("bucket size in TB", logkeys.Size, bkSize)
		bucketPrivate.Spec.Request = &pb.StorageCapacityRequest{
			Size: fmt.Sprintf("%sGB", bkSize),
		}
	} else {
		// Retrieve cloudaccount bucket size from config
		bucketPrivate.Spec.Request = &pb.StorageCapacityRequest{
			Size: fmt.Sprintf("%sGB", objSrv.bucketSizeInGB),
		}
	}

	logger.Info("bucket create request size", logkeys.Size, bucketPrivate.Spec.Request.Size)
	schedule := &pb.BucketSchedule{}
	logger.Info("fetching cluster info from active buckets for cloud account")
	bucketClusterInfo, err := query.SearchBucketClusterInfo(ctx, tx, in.Metadata.CloudAccountId, timestampInfinityStr)
	if err != nil {
		logger.Error(err, "error fetching active buckets for cloud account")
		return nil, status.Errorf(codes.Internal, "bucket scheduling failed due to fetch active buckets")
	}
	if bucketClusterInfo != nil {
		logger.Info("Setting information from bucket cluster", logkeys.ClusterName, bucketClusterInfo.ClusterName)
		schedule.Cluster = bucketClusterInfo
	}

	if bucketClusterInfo == nil {
		logger.Info("fetching cluster details from active principals for cloud account")
		userClusterInfo, err := query.SearchPrincipalClusterInfo(ctx, tx, in.Metadata.CloudAccountId, timestampInfinityStr)
		if err != nil {
			logger.Error(err, "error fetching active users for cloud account")
			return nil, status.Errorf(codes.Internal, "bucket scheduling failed due to fetch active users")
		}
		if userClusterInfo != nil {
			logger.Info("Setting from user cluster fetch info", logkeys.ClusterName, userClusterInfo.ClusterName)
			schedule = &pb.BucketSchedule{
				Cluster: &pb.AssignedCluster{
					ClusterName: userClusterInfo.ClusterName,
					ClusterAddr: userClusterInfo.AccessEndpoint,
					ClusterUUID: userClusterInfo.ClusterId,
				},
			}
		}
	}
	// schedule bucket if cluster info is not already available
	if reflect.DeepEqual(schedule, &pb.BucketSchedule{}) {
		logger.Info("scheduling bucket from scheduler")
		scheduler, err := objSrv.scheduleBucket(ctx, in.Spec)
		if err != nil {
			logger.Error(err, "error scheduling bucket to cluster")
			objSrv.quotaServiceClient.decBucketQuota(ctx, in.Metadata.CloudAccountId)
			return nil, status.Errorf(codes.Internal, "bucket scheduling failed")
		}
		if scheduler.Schedule == nil {
			logger.Info("no clusters scheduled")
			return nil, status.Errorf(codes.Internal, "no clusters scheduled")
		}
		schedule = scheduler.Schedule
	}
	bucketPrivate.Spec.Schedule = schedule
	bucketPrivate.Status = &pb.ObjectBucketStatus{
		Cluster: &pb.ObjectCluster{
			ClusterId:      bucketPrivate.Spec.Schedule.Cluster.ClusterName,
			AccessEndpoint: bucketPrivate.Spec.Schedule.Cluster.ClusterAddr,
			ClusterName:    bucketPrivate.Spec.Schedule.Cluster.ClusterName,
		},
		SecurityGroup: &v1.BucketSecurityGroup{
			NetworkFilterAllow: []*v1.BucketNetworkGroup{
				&v1.BucketNetworkGroup{
					Subnet:       vnetPrivate.Spec.Subnet,
					PrefixLength: vnetPrivate.Spec.PrefixLength,
					Gateway:      vnetPrivate.Spec.Gateway,
				},
			},
		},
	}

	bucketPrivate.Metadata.ResourceId = uuid.NewString()
	bucketPrivate.Metadata.CreationTimestamp = timestamppb.Now()
	// update bucket-id to include cloudaccount
	in.Metadata.Name = fmt.Sprintf("%s-%s", in.Metadata.CloudAccountId, in.Metadata.Name)
	if err := query.StoreBucketRequest(ctx, tx, &bucketPrivate); err != nil {
		objSrv.quotaServiceClient.decBucketQuota(ctx, in.Metadata.CloudAccountId)
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == kErrUniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "bucket name "+bucketPrivate.Metadata.Name+" already exists")
		}
		logger.Error(err, "error storing bucket request into db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	if vnetPrivate.Metadata != nil {
		// update the vnet bucket status if vnet already exists
		updateSubnetParam := pb.BucketSubnetStatusUpdateRequest{
			ResourceId:      vnetPrivate.Metadata.ResourceId,
			CloudacccountId: vnetPrivate.Metadata.CloudAccountId,
			VNetName:        vnetPrivate.Metadata.Name,
			Status:          pb.BucketSubnetEventStatus_E_ADDED,
		}
		if err := query.UpdateStatusForSubnet(ctx, tx, &updateSubnetParam); err != nil {
			logger.Error(err, "error updating status timestamp for bucket subnet")
			return nil, fmt.Errorf("error updating bucket status")
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &bucketPrivate, nil
}

func (objSrv *BucketsServiceServer) scheduleBucket(ctx context.Context,
	requestSpec *pb.ObjectBucketSpecPrivate) (*pb.BucketScheduleResponse, error) {

	logger := log.FromContext(ctx).WithName("BucketsServiceServer.scheduleBucket")
	logger.Info("discovering bucket target")
	request := pb.BucketScheduleRequest{
		RequestSpec: requestSpec,
	}
	schedResp, err := objSrv.schedulerClient.ScheduleBucket(ctx, &request)
	if err != nil {
		logger.Error(err, "error scheduling request")
		return nil, fmt.Errorf("error scheduling request")
	}
	logger.Info("scheduler response", logkeys.Response, schedResp)
	return schedResp, nil
}

func (objSrv *BucketsServiceServer) getBucket(ctx context.Context,
	in *pb.ObjectBucketGetPrivateRequest) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.getBucket")

	logger.Info("entering bucket get", logkeys.CloudAccountId, in.Metadata.CloudAccountId)
	defer logger.Info("returning from bucket get")

	if err := isValidBucketGetRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}
	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Commit()

	bucketPrivate := &pb.ObjectBucketPrivate{}
	if in.Metadata.GetBucketId() != "" {
		bucketPrivate, err = query.GetBucketById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetBucketId(), timestampInfinityStr)
	} else if in.Metadata.GetBucketName() != "" {
		bucketPrivate, err = query.GetBucketByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetBucketName(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}
	// update bucket user policies
	userAccessPolicies, err := getBucketUserAccess(ctx, tx, bucketPrivate.Metadata.CloudAccountId, bucketPrivate.Metadata.Name)
	if err != nil {
		return nil, err
	}
	// check with authz service is permitted to access resource
	authzUserIds := []string{}
	authzUserIdMap := make(map[string]bool)
	if objSrv.cfg.AuthzEnabled {
		//check if private request
		_, err := grpcutil.ExtractClaimFromCtx(ctx, true, grpcutil.EmailClaim)
		if err != nil && err == grpcutil.ErrNoJWTToken {
			logger.Info("skipping auth check for private request")
		} else {
			logger.Info("successfully extracted email from context")
			for idx := 0; idx < len(userAccessPolicies); idx++ {
				authzUserIds = append(authzUserIds, userAccessPolicies[idx].Metadata.UserId)
			}
			authzLookupResponse, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", "principal", authzUserIds, true)
			if err != nil {
				return nil, err
			}
			for _, resourceId := range authzLookupResponse.ResourceIds {
				authzUserIdMap[resourceId] = true
			}

			allowedUserAccessPolicies := []*v1.BucketUserAccess{}
			for idx := 0; idx < len(userAccessPolicies); idx++ {
				// Check if bucket is in allowed resource list
				if _, exists := authzUserIdMap[userAccessPolicies[idx].Metadata.UserId]; !exists {
					continue
				}
				allowedUserAccessPolicies = append(allowedUserAccessPolicies, userAccessPolicies[idx])
			}
			userAccessPolicies = allowedUserAccessPolicies
		}
	}

	lfRequest := pb.BucketLifecycleRuleSearchRequest{
		CloudAccountId: bucketPrivate.Metadata.CloudAccountId,
		BucketId:       bucketPrivate.Metadata.ResourceId,
	}
	lifecyclePolicies, err := objSrv.getAllBucketLifecycleRules(ctx, &lfRequest)
	if err != nil {
		return nil, err
	}
	bucketPrivate.Status.Policy = &pb.BucketPolicyStatus{
		UserAccessPolicies: userAccessPolicies,
		LifecycleRules:     lifecyclePolicies.Rules,
	}

	return bucketPrivate, nil
}

func (objSrv *BucketsServiceServer) deleteBucket(ctx context.Context,
	in *pb.ObjectBucketDeletePrivateRequest, force bool) (*emptypb.Empty, error) {

	logger := log.FromContext(ctx).WithName("BucketsServiceServer.deleteBucket")
	logger.Info("entering bucket delete", logkeys.CloudAccountId, in.Metadata.CloudAccountId)
	defer logger.Info("returning from bucket delete")
	if err := isValidBucketDeleteRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Rollback()

	bucketPrivate := &pb.ObjectBucketPrivate{}
	if in.Metadata.GetBucketId() != "" {
		bucketPrivate, err = query.GetBucketById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetBucketId(), timestampInfinityStr)
	} else if in.Metadata.GetBucketName() != "" {
		bucketPrivate, err = query.GetBucketByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetBucketName(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	// check if bucket is empty
	if !force {
		bucketCap, err := objSrv.bucketUserClient.GetBucketCapacity(ctx, &v1.BucketFilter{
			ClusterId: bucketPrivate.Spec.Schedule.Cluster.ClusterUUID,
			BucketId:  bucketPrivate.Metadata.Name,
		})
		if err != nil {
			logger.Error(err, "error checking bucket capacity")
			return &emptypb.Empty{}, err
		}
		if bucketCap.Capacity.AvailableBytes != 0 {
			return &emptypb.Empty{}, status.Errorf(codes.FailedPrecondition, "ensure the bucket is empty")
		}
	}
	reclaimedSize, err := query.UpdateBucketForDeletion(ctx, tx, in.Metadata)
	if err != nil {
		logger.Error(err, "error updating bucket for deletion")
		return &emptypb.Empty{}, status.Errorf(codes.FailedPrecondition, "bucket precondition failed")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	objSrv.quotaServiceClient.decBucketQuota(ctx, in.Metadata.CloudAccountId)
	logger.Info("bucket deleted successfully, Size reclaimed", logkeys.Size, reclaimedSize)

	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.authzServiceClient.RemoveResourceFromCloudAccountRole(ctx,
			&v1.CloudAccountRoleResourceRequest{
				CloudAccountId: in.Metadata.CloudAccountId,
				ResourceId:     bucketPrivate.Metadata.ResourceId,
				ResourceType:   "objectstorage"})
		if err != nil {
			logger.Error(err, "error authz when removing resource from cloudAccountRole", "resourceId", bucketPrivate.Metadata.ResourceId, "resourceType", "objectstorage")
			return nil, status.Errorf(codes.Internal, "bucket %v couldn't be removed from cloudAccountRole resourceType %v",
				bucketPrivate.Metadata.ResourceId, "objectstorage")
		}
	}
	return &emptypb.Empty{}, nil
}

// reads atleast one instance and atleast one interface of that instance
// to determine the tenant subnet assigned to that cloudaccount
// For any error, it always returns a default subnet
func (objSrv *BucketsServiceServer) readVNetForCloudaccount(ctx context.Context,
	cloudaccount string) (*pb.VNetSpecPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.readVNetForCloudaccount")

	logger.Info("entering vnet query", logkeys.CloudAccountId, cloudaccount)
	defer logger.Info("returning from vnet query")
	vnetPrivate := pb.VNetSpecPrivate{
		Subnet:       "0.0.0.0",
		PrefixLength: 27,
		Gateway:      "0.0.0.0",
	}

	instances, err := objSrv.instanceServiceClient.SearchPrivate(ctx, &v1.InstanceSearchPrivateRequest{
		Metadata: &v1.InstanceMetadataSearch{
			CloudAccountId: cloudaccount,
		},
	})

	if err != nil {
		return &vnetPrivate, err
	}

	for _, instance := range instances.Items {
		for _, inf := range instance.Status.Interfaces {
			vnetPrivate.Subnet = inf.Subnet
			vnetPrivate.PrefixLength = inf.PrefixLength
			vnetPrivate.Gateway = inf.Gateway
			// break as we need info from atleast one interface
			break
		}
		// break as we need info from atleast one instance
		break
	}
	return &vnetPrivate, nil
}
