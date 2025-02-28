package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// BucketsServiceServer is used to implement pb.UnimplementedObjectStorageServiceServer and
// UnimplementedObjectStorageServicePrivateServer
type BucketsServiceServer struct {
	pb.UnimplementedObjectStorageServiceServer
	pb.UnimplementedObjectStorageServicePrivateServer
	session                     *sql.DB
	cloudAccountServiceClient   pb.CloudAccountServiceClient
	productcatalogServiceClient pb.ProductCatalogServiceClient
	authzServiceClient          v1.AuthzServiceClient
	schedulerClient             pb.FilesystemSchedulerPrivateServiceClient
	quotaServiceClient          *QuotaService
	userClient                  pb.FilesystemUserPrivateServiceClient
	bucketUserClient            pb.BucketUserPrivateServiceClient
	lifecycleRuleClient         pb.BucketLifecyclePrivateServiceClient
	instanceServiceClient       pb.InstancePrivateServiceClient
	bucketSizeInGB              string
	objectProduct               objectProductInfo
	cfg                         *Config
}

type objectProductInfo struct {
	availableSizes   []int64
	updatedTimestamp time.Time
}

const (
	storageInterfaceName = "storage"
	defaultName          = "default"
	devVnetName          = "us-dev"
)

func NewObjectService(ctx context.Context, session *sql.DB, bucketSize string, cloudAccountSvc pb.CloudAccountServiceClient,
	productcatalogSvc pb.ProductCatalogServiceClient, authzSvc v1.AuthzServiceClient,
	schedulerClient pb.FilesystemSchedulerPrivateServiceClient,
	quotaServiceClient *QuotaService,
	userClient pb.FilesystemUserPrivateServiceClient,
	bucketUserClient pb.BucketUserPrivateServiceClient,
	lifecycleRuleClient pb.BucketLifecyclePrivateServiceClient,
	instanceServiceClient pb.InstancePrivateServiceClient, cfg *Config) (*BucketsServiceServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}

	if err := utils.ParseBucketSize(bucketSize); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("invalid default bucket size")
	}
	objSrv := BucketsServiceServer{
		session:                     session,
		cloudAccountServiceClient:   cloudAccountSvc,
		schedulerClient:             schedulerClient,
		productcatalogServiceClient: productcatalogSvc,
		authzServiceClient:          authzSvc,
		quotaServiceClient:          quotaServiceClient,
		userClient:                  userClient,
		bucketUserClient:            bucketUserClient,
		lifecycleRuleClient:         lifecycleRuleClient,
		bucketSizeInGB:              bucketSize,
		instanceServiceClient:       instanceServiceClient,
		cfg:                         cfg,
	}

	objSrv.objectProduct = objectProductInfo{}

	return &objSrv, nil
}

// Create s3 bucket.
func (objSrv *BucketsServiceServer) CreateBucket(ctx context.Context, in *pb.ObjectBucketCreateRequest) (*pb.ObjectBucket, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateBucket").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket create", logkeys.Input, in)
	defer logger.Info("returning from bucket create")

	privateRequest := convertBucketCreateReqPublicToPrivate(in)
	bucketPrivate, err := objSrv.createBucket(ctx, privateRequest)
	if err != nil {
		return nil, err
	}

	bucketPrivate.Spec.Request.Size = utils.ProcesSize(bucketPrivate.Spec.Request.Size)
	return convertBucketPrivateToPublic(bucketPrivate, false), nil
}

// Get the status of an s3 bucket.
func (objSrv *BucketsServiceServer) GetBucket(ctx context.Context, in *pb.ObjectBucketGetRequest) (*pb.ObjectBucket, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketServiceServer.Get").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket get", logkeys.Input, in)
	defer logger.Info("returning from bucket get")
	inPrivate := (*pb.ObjectBucketGetPrivateRequest)(in)
	bucketPrivate, err := objSrv.getBucket(ctx, inPrivate)
	if err != nil {
		return nil, err
	}
	bucketPublic := convertBucketPrivateToPublic(bucketPrivate, true)
	bucketPrivate.Spec.Request.Size = utils.ProcesSize(bucketPrivate.Spec.Request.Size)

	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", "objectstorage", []string{bucketPrivate.GetMetadata().ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}
	return bucketPublic, nil
}

// List buckets.
func (objSrv *BucketsServiceServer) SearchBucket(ctx context.Context, in *pb.ObjectBucketSearchRequest) (*pb.ObjectBucketSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketServiceServer.Search").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket search", logkeys.Input, in)
	defer logger.Info("returning from bucket search")
	if err := isValidBucketSearchRequest(ctx, in); err != nil {
		return nil, err
	}

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "database transaction failed")
	}
	bucketPrivateList, err := query.GetBucketsByCloudaccountId(ctx, tx,
		in.Metadata.CloudAccountId, timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error reading buckets")
	}
	resp := pb.ObjectBucketSearchResponse{}

	// check with authz service is permitted to access resource
	authzResourceIds := []string{}
	authzResourceIdMap := make(map[string]bool)
	if objSrv.cfg.AuthzEnabled {
		for idx := 0; idx < len(bucketPrivateList); idx++ {
			authzResourceIds = append(authzResourceIds, bucketPrivateList[idx].Metadata.ResourceId)
		}
		authzLookupResponse, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", "objectstorage", authzResourceIds, true)
		if err != nil {
			return nil, err
		}
		for _, resourceId := range authzLookupResponse.ResourceIds {
			authzResourceIdMap[resourceId] = true
		}
	}

	for idx := 0; idx < len(bucketPrivateList); idx++ {
		// Check if bucket is in allowed resource list
		if _, exists := authzResourceIdMap[bucketPrivateList[idx].Metadata.ResourceId]; objSrv.cfg.AuthzEnabled && !exists {
			continue
		}
		bucketPrivate := bucketPrivateList[idx]
		bucketPrivate.Spec.Request.Size = utils.ProcesSize(bucketPrivate.Spec.Request.Size)
		userAccessPolicies, err := getBucketUserAccess(ctx, tx, bucketPrivate.Metadata.CloudAccountId, bucketPrivate.Metadata.Name)
		if err != nil {
			return nil, err
		}
		// check with authz service is permitted to access resource
		authzUserIds := []string{}
		authzUserIdMap := make(map[string]bool)
		if objSrv.cfg.AuthzEnabled {
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

		bucketPublic := convertBucketPrivateToPublic(bucketPrivate, true)
		resp.Items = append(resp.Items, bucketPublic)
	}
	return &resp, nil
}

// Request deletion of a bucket.
func (objSrv *BucketsServiceServer) DeleteBucket(ctx context.Context, in *pb.ObjectBucketDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketServiceServer.DeleteBucket").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket delete ", logkeys.Input, in)
	defer logger.Info("returning from bucket delete")
	inPrivate := (*pb.ObjectBucketGetPrivateRequest)(in)
	bucketPrivate, err := objSrv.getBucket(ctx, inPrivate)
	if err != nil {
		return nil, err
	}
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "delete", "objectstorage", []string{bucketPrivate.Metadata.ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}

	privateReq := (*pb.ObjectBucketDeletePrivateRequest)(in)
	return objSrv.deleteBucket(ctx, privateReq, false)
}

// Request bucket lifecycle policy
func (objSrv *BucketsServiceServer) CreateBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleCreateRequest) (*pb.BucketLifecycleRule, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket create lifecyclerule", logkeys.Input, in)
	defer logger.Info("returning from bucket create lifecyclerule")

	ruleReqPrivate := (*pb.BucketLifecycleRuleCreatePrivateRequest)(in)

	lfRulePrivate, err := objSrv.createBucketLifecycleRule(ctx, ruleReqPrivate)
	if err != nil {
		return nil, err
	}

	return convertLifecycleRulePrivateToPublic(lfRulePrivate), nil
}

// Get bucket lifecycle policy
func (objSrv *BucketsServiceServer) GetBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleGetRequest) (*pb.BucketLifecycleRule, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.GetBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket get lifecyclerule ", logkeys.Input, in)
	defer logger.Info("returning from bucket get lifecyclerule")

	reqPrivate := (*pb.BucketLifecycleRuleGetPrivateRequest)(in)
	lfRulePrivate, err := objSrv.getBucketLifecycleRule(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}
	return convertLifecycleRulePrivateToPublic(lfRulePrivate), nil
}

// List bucket lifecycle policy
func (objSrv *BucketsServiceServer) SearchBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleSearchRequest) (*pb.BucketLifecycleRuleSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.SearchBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket search lifecycle", logkeys.BucketId, in.BucketId)
	defer logger.Info("returning from bucket search lifecycle")

	ruleList, err := objSrv.getAllBucketLifecycleRules(ctx, in)
	if err != nil {
		logger.Error(err, "error searching lifecycle rules")
		return nil, err
	}
	// TODO: Handle the case for private API
	return ruleList, nil
}

// Update bucket lifecycle policy
func (objSrv *BucketsServiceServer) UpdateBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleUpdateRequest) (*pb.BucketLifecycleRule, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update lifecycle", logkeys.Input, in)
	defer logger.Info("returning from bucket update lifecycle")

	reqPrivate := &pb.BucketLifecycleRuleGetPrivateRequest{Metadata: in.Metadata}
	lfRulePrivate, err := objSrv.getBucketLifecycleRule(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "update", "lifecyclerule", []string{lfRulePrivate.Metadata.ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}

	privateReq := (*pb.BucketLifecycleRuleUpdatePrivateRequest)(in)
	lfPrivate, err := objSrv.updateBucketLifecycleRule(ctx, privateReq)
	if err != nil {
		logger.Error(err, "error updating lifecycle rule")
		return nil, err
	}
	lfPublic := convertLifecycleRulePrivateToPublic(lfPrivate)
	return lfPublic, nil
}

// Delete bucket lifecycle policy
func (objSrv *BucketsServiceServer) DeleteBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.DeleteBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket delete lifecyclerule", logkeys.BucketId, in.Metadata.BucketId, logkeys.RuleId, in.Metadata.RuleId)
	defer logger.Info("returning from bucket delete lifecyclerule")

	reqPrivate := &pb.BucketLifecycleRuleGetPrivateRequest{Metadata: in.Metadata}
	lfRulePrivate, err := objSrv.getBucketLifecycleRule(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "delete", "lifecyclerule", []string{lfRulePrivate.Metadata.ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}
	privateReq := (*pb.BucketLifecycleRuleDeletePrivateRequest)(in)
	return objSrv.deleteBucketLifecycleRule(ctx, privateReq)
}

// Create object service user
func (objSrv *BucketsServiceServer) CreateObjectUser(ctx context.Context, in *pb.CreateObjectUserRequest) (*pb.ObjectUser, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateObjectUser").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket create user", logkeys.Input, in)
	defer logger.Info("returning from bucket create user")

	privateReq := (*pb.CreateObjectUserPrivateRequest)(in)

	userPrivate, err := objSrv.createObjectUser(ctx, privateReq)
	if err != nil {
		logger.Error(err, "error creating user")
		return nil, err
	}

	publicUser := convertObjectUserPrivateToPublic(userPrivate)

	return publicUser, nil
}

// Get object service user
func (objSrv *BucketsServiceServer) GetObjectUser(ctx context.Context, in *pb.ObjectUserGetRequest) (*pb.ObjectUser, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.GetObjectUser").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket get user", logkeys.Input, in)
	defer logger.Info("returning from bucket get user")

	reqPrivate := (*pb.ObjectUserGetPrivateRequest)(in)
	userPrivate, err := objSrv.getObjectUser(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}
	userPublic := convertObjectUserPrivateToPublic(userPrivate)

	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", "principal", []string{userPrivate.Metadata.UserId}, false)
		if err != nil {
			return nil, err
		}
	}
	return userPublic, nil
}

// List all object service user
func (objSrv *BucketsServiceServer) SearchObjectUser(ctx context.Context, in *pb.ObjectUserSearchRequest) (*pb.ObjectUserSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.SearchObjectUser").WithValues(logkeys.CloudAccountId, in.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket search user")
	defer logger.Info("returning from bucket search user")

	if err := isValidBucketUserSearchRequest(ctx, in); err != nil {
		return nil, err
	}
	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	bucketUsersPrivateList, err := query.GetBucketUsersByCloudaccountId(ctx, tx,
		in.CloudAccountId, timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error reading filesystems")
	}
	resp := pb.ObjectUserSearchResponse{}

	// check with authz service is permitted to access resource
	authzResourceIds := []string{}
	authzResourceIdMap := make(map[string]bool)
	if objSrv.cfg.AuthzEnabled {
		for idx := 0; idx < len(bucketUsersPrivateList); idx++ {
			authzResourceIds = append(authzResourceIds, bucketUsersPrivateList[idx].Metadata.UserId)
		}
		authzLookupResponse, err := objSrv.lookupResourcePermission(ctx, in.CloudAccountId, "get", "principal", authzResourceIds, true)
		if err != nil {
			return nil, err
		}
		for _, resourceId := range authzLookupResponse.ResourceIds {
			authzResourceIdMap[resourceId] = true
		}
	}
	for idx := 0; idx < len(bucketUsersPrivateList); idx++ {
		// Check if bucket is in allowed resource list
		if _, exists := authzResourceIdMap[bucketUsersPrivateList[idx].Metadata.UserId]; objSrv.cfg.AuthzEnabled && !exists {
			continue
		}
		bucketUserPublic := convertObjectUserPrivateToPublic(bucketUsersPrivateList[idx])
		resp.Users = append(resp.Users, bucketUserPublic)
	}
	return &resp, nil
}

// Delete object service user
func (objSrv *BucketsServiceServer) DeleteObjectUser(ctx context.Context, in *pb.ObjectUserDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.DeleteObjectUser").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket delete user", logkeys.Input, in)
	defer logger.Info("returning from bucket delete user")

	reqPrivate := (*pb.ObjectUserDeletePrivateRequest)(in)
	return objSrv.deleteBucketUser(ctx, reqPrivate)
}

// Update policies for object service user
func (objSrv *BucketsServiceServer) UpdateObjectUserPolicy(ctx context.Context, in *pb.ObjectUserUpdateRequest) (*pb.ObjectUser, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateObjectUserPolicy").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update user policy", logkeys.Input, in)
	defer logger.Info("returning from bucket update user policy")
	// Check if principal exists
	reqPrivate := &pb.ObjectUserGetPrivateRequest{Metadata: in.Metadata}
	userPrivate, err := objSrv.getObjectUser(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}

	privateReq := (*pb.ObjectUserUpdatePrivateRequest)(in)
	userPrivate, err = objSrv.updateObjectUserPolicies(ctx, privateReq)
	if err != nil {
		logger.Error(err, "error updating user policies")
		return nil, err
	}
	publicUser := convertObjectUserPrivateToPublic(userPrivate)
	return publicUser, nil
}

// Update policies for object service user
func (objSrv *BucketsServiceServer) UpdateObjectUserCredentials(ctx context.Context, in *pb.ObjectUserUpdateCredsRequest) (*pb.ObjectUser, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateObjectUserCredentials").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update user credentials", logkeys.Input, in)
	defer logger.Info("returning from bucket update user credentials")
	// Check if principal exists
	reqPrivate := &pb.ObjectUserGetPrivateRequest{Metadata: in.Metadata}
	userPrivate, err := objSrv.getObjectUser(ctx, reqPrivate)
	if err != nil {
		return nil, err
	}
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "updatecreds", "principal", []string{userPrivate.Metadata.UserId}, false)
		if err != nil {
			return nil, err
		}
	}
	userPrivate, err = objSrv.updateObjectUserCredentials(ctx, in)
	if err != nil {
		logger.Error(err, "error updating user credentials")
		return nil, status.Errorf(codes.Internal, "error updating user credentials")
	}
	return convertObjectUserPrivateToPublic(userPrivate), nil

}

func (objSrv *BucketsServiceServer) CreateBucketPrivate(ctx context.Context, in *pb.ObjectBucketCreatePrivateRequest) (*pb.ObjectBucketPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateBucketPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket create private", logkeys.Input, in)
	defer logger.Info("returning from bucket create private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}

func (objSrv *BucketsServiceServer) GetBucketPrivate(ctx context.Context, in *pb.ObjectBucketGetPrivateRequest) (*pb.ObjectBucketPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.GetBucketPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket get private", logkeys.Input, in)
	defer logger.Info("returning from bucket get private")
	bucketPrivate, err := objSrv.getBucket(ctx, in)
	if err != nil {
		return nil, err
	}
	return bucketPrivate, nil
}

// List Filesystem as a stream.
// This returns all bucket requests that are pending.
func (objSrv *BucketsServiceServer) SearchBucketPrivate(in *pb.ObjectBucketSearchPrivateRequest, rs pb.ObjectStorageServicePrivate_SearchBucketPrivateServer) error {
	dbSession := objSrv.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(rs.Context(), nil)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}
	defer tx.Rollback()

	return query.GetBucketsRequests(tx, in.ResourceVersion, timestampInfinityStr, rs)
}
func (objSrv *BucketsServiceServer) DeleteBucketPrivate(ctx context.Context, in *pb.ObjectBucketDeletePrivateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.DeleteBucketPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket delete private", logkeys.Input, in)
	defer logger.Info("returning from bucket delete private")

	return objSrv.deleteBucket(ctx, in, true)
}

func (objSrv *BucketsServiceServer) CreateBucketLifecycleRulePrivate(ctx context.Context, in *pb.BucketLifecycleRuleCreatePrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateBucketLifecycleRulePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket create lifecyclerule private", logkeys.Input, in)
	defer logger.Info("returning from bucket create lifecyclerule private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")

}
func (objSrv *BucketsServiceServer) GetBucketLifecycleRulePrivate(ctx context.Context, in *pb.BucketLifecycleRuleGetPrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.GetBucketLifecycleRulePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket get lifecyclerule private", logkeys.Input, in)
	defer logger.Info("returning from bucket get lifecyclerule private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) SearchBucketLifecycleRulePrivate(ctx context.Context, in *pb.BucketLifecycleRuleSearchPrivateRequest) (*pb.BucketLifecycleRuleSearchPrivateResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.SearchBucketLifecycleRulePrivate").WithValues(logkeys.CloudAccountId, in.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket search lifecycle private", logkeys.Input, in)
	defer logger.Info("returning from bucket search lifecycle private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) UpdateBucketLifecycleRulePrivate(ctx context.Context, in *pb.BucketLifecycleRuleUpdatePrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateBucketLifecycleRulePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update lifecycle private", logkeys.Input, in)
	defer logger.Info("returning from bucket update lifecycle private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) DeleteBucketLifecycleRulePrivate(ctx context.Context, in *pb.BucketLifecycleRuleDeletePrivateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.DeleteBucketLifecycleRulePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket delete lifecyclerule private", logkeys.Input, in)
	defer logger.Info("returning from bucket delete lifecyclerule private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) CreateObjectUserPrivate(ctx context.Context, in *pb.CreateObjectUserPrivateRequest) (*pb.ObjectUserPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.CreateObjectUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket create user private", logkeys.Input, in)
	defer logger.Info("returning from bucket create user private")

	userPrivate, err := objSrv.createObjectUser(ctx, in)
	if err != nil {
		logger.Error(err, "error creating user")
		return nil, err
	}
	return userPrivate, nil
}
func (objSrv *BucketsServiceServer) GetObjectUserPrivate(ctx context.Context, in *pb.ObjectUserGetPrivateRequest) (*pb.ObjectUserPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.GetObjectUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket get user private", logkeys.Input, in)
	defer logger.Info("returning from bucket get user private")

	userPrivate, err := objSrv.getObjectUser(ctx, in)
	if err != nil {
		return nil, err
	}

	return userPrivate, nil
}
func (objSrv *BucketsServiceServer) SearchObjectUserPrivate(in *pb.ObjectUserSearchPrivateRequest, rs pb.ObjectStorageServicePrivate_SearchObjectUserPrivateServer) error {
	logger := log.FromContext(rs.Context()).WithName("BucketsServiceServer.SearchObjectUserPrivate")

	logger.Info("entering bucket search user private", logkeys.Input, in)
	defer logger.Info("returning from bucket search user private")

	return status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) DeleteObjectUserPrivate(ctx context.Context, in *pb.ObjectUserDeletePrivateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.DeleteObjectUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket delete user private", logkeys.Input, in)
	defer logger.Info("returning from bucket delete user private")

	return objSrv.deleteBucketUser(ctx, in)
}
func (objSrv *BucketsServiceServer) UpdateObjectUserPrivate(ctx context.Context, in *pb.ObjectUserUpdatePrivateRequest) (*pb.ObjectUserPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateObjectUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update user private", logkeys.Input, in)
	defer logger.Info("returning from bucket update user private")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}

// APIs not mapped to public gRPC
func (objSrv *BucketsServiceServer) UpdateBucketStatus(ctx context.Context, in *pb.ObjectBucketStatusUpdateRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateBucketStatus").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering update bucket status", logkeys.Input, in)
	defer logger.Info("returning from update bucket status")
	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}
	defer tx.Rollback()

	if err := query.UpdateBucketState(ctx, tx, in); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &emptypb.Empty{}, nil
}

// Remove finalizer from an bucket that was previously requested to be deleted.
// After this returns, the record will no longer be visible to users or controllers.
// Used by object bucket Replicator.
func (objSrv *BucketsServiceServer) RemoveBucketFinalizer(ctx context.Context, in *pb.ObjectBucketRemoveFinalizerRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.RemoveBucketFinalizer").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering bucket update bucket remove finalizer for", logkeys.Input, in)
	defer logger.Info("returning from bucket remove finalizer status")

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}
	defer tx.Rollback()

	if err := query.UpdateBucketDeletionTime(ctx, tx, in.Metadata.CloudAccountId,
		in.Metadata.ResourceId); err != nil {
		logger.Error(err, "error updating deletion timestamp")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &emptypb.Empty{}, nil
}

func (objSrv *BucketsServiceServer) UpdateObjectUserStatus(ctx context.Context, in *pb.ObjectUserStatusUpdateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateObjectUserStatus").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket update user status", logkeys.Input, in)
	defer logger.Info("returning from bucket update user status")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}

// Remove finalizer from an object user that was previously requested to be deleted.
// After this returns, the record will no longer be visible to users or controllers.
// Used by storage user service.
func (objSrv *BucketsServiceServer) RemoveObjectUserFinalizer(ctx context.Context, in *pb.ObjectUserRemoveFinalizerRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.RemoveObjectUserFinalizer").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering user remove finalizer", logkeys.Input, in)
	defer logger.Info("returning from user remove finalizer")

	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (objSrv *BucketsServiceServer) PingPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.PingPrivate")

	logger.Info("entering filesystem private Ping")
	defer logger.Info("returning from filesystem private Ping")

	return &emptypb.Empty{}, nil
}

func (objSrv *BucketsServiceServer) AddBucketSubnet(ctx context.Context, in *pb.VNetPrivate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.AddSubnet").
		WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId, logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("entering bucket private service for add subnet")
	defer logger.Info("returning from bucket private service for add subnet")

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}
	defer tx.Rollback()
	// skip if storage interface
	if strings.HasSuffix(in.Metadata.Name, storageInterfaceName) {
		logger.Info("ignore subnet insertion for", logkeys.SubnetName, in.Metadata.Name)
		return &emptypb.Empty{}, nil
	}
	// skip if not default or dev subnet
	if !strings.HasPrefix(in.Metadata.Name, devVnetName) && !strings.HasSuffix(in.Metadata.Name, defaultName) {
		logger.Info("ignore non-default subnet insertion for", logkeys.Subnet, in.Metadata.Name)
		return &emptypb.Empty{}, nil
	}
	exists, err := query.CheckSubnetExists(ctx, tx, in)
	if err != nil {
		logger.Error(err, "error checking bucket subnet in db")
		return &emptypb.Empty{}, err
	}
	logger.Info("db entry", "found?", exists)
	if !exists {
		// store bucket subnet in the database
		if err := query.StoreBucketSubnet(ctx, tx, in); err != nil {
			logger.Error(err, "error storing bucket subnet request")
			return &emptypb.Empty{}, err
		}
	} else {
		// store bucket subnet in the database
		if err := query.UpdateBucketSubnet(ctx, tx, in); err != nil {
			logger.Error(err, "error updating bucket subnet request")
			return &emptypb.Empty{}, err
		}
	}

	bucketPrivateList, err := query.GetBucketsByCloudaccountId(ctx, tx,
		in.Metadata.CloudAccountId, timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error reading buckets")
	}

	for idx := range bucketPrivateList {
		isDuplicate := false
		bucketPrivate := bucketPrivateList[idx]
		// Check if SecurityGroup is nil
		if bucketPrivate.Status.SecurityGroup == nil {
			bucketPrivate.Status.SecurityGroup = &v1.BucketSecurityGroup{
				NetworkFilterAllow: []*v1.BucketNetworkGroup{},
			}
		}
		// Now append to the existing NetworkFilterAllow
		for idx, rule := range bucketPrivate.Status.SecurityGroup.NetworkFilterAllow {
			if strings.EqualFold(rule.Subnet, "0.0.0.0") {
				bucketPrivate.Status.SecurityGroup.NetworkFilterAllow = removeNetworkFilter(bucketPrivate.Status.SecurityGroup.NetworkFilterAllow, idx)
			} else if strings.EqualFold(rule.Subnet, in.Spec.Subnet) {
				isDuplicate = true
			}
		}
		if !isDuplicate {
			bucketPrivate.Status.SecurityGroup.NetworkFilterAllow = append(
				bucketPrivate.Status.SecurityGroup.NetworkFilterAllow,
				&v1.BucketNetworkGroup{
					Subnet:       in.Spec.Subnet,
					PrefixLength: in.Spec.PrefixLength,
					Gateway:      in.Spec.Gateway,
				},
			)
		}
		logger.Info("bucket information added with subnet is ", logkeys.BucketInfo, bucketPrivate)
		if err := query.UpdateSubnetBucketRequest(ctx, tx, bucketPrivate); err != nil {
			logger.Error(err, "error storing bucket request into db")
			return nil, status.Errorf(codes.Internal, "database transaction failed")
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	return &emptypb.Empty{}, nil
}

func (objSrv *BucketsServiceServer) RemoveBucketSubnet(ctx context.Context, in *pb.VNetReleaseSubnetRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.RemoveSubnet").WithValues(logkeys.CloudAccountId, in.VNetReference.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket private service for remove subnet")
	defer logger.Info("returning from bucket private service for remove subnet")

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}
	defer tx.Rollback()

	// skip if storage interface
	if strings.HasSuffix(in.VNetReference.Name, storageInterfaceName) {
		logger.Info("ignore subnet removal for", logkeys.Subnet, in.VNetReference.Name)
		return &emptypb.Empty{}, nil
	}
	// skip removal if not default or dev subnet
	if !strings.HasPrefix(in.VNetReference.Name, devVnetName) && !strings.HasSuffix(in.VNetReference.Name, defaultName) {
		logger.Info("ignore non-default subnet removal for", logkeys.Subnet, in.VNetReference.Name)
		return &emptypb.Empty{}, nil
	}
	bucketPrivateList, err := query.GetBucketsByCloudaccountId(ctx, tx,
		in.VNetReference.CloudAccountId, timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error reading buckets")
	}

	for idx := range bucketPrivateList {
		bucketPrivate := bucketPrivateList[idx]
		bucketPrivate.Status.SecurityGroup = &v1.BucketSecurityGroup{
			NetworkFilterAllow: []*v1.BucketNetworkGroup{
				&v1.BucketNetworkGroup{
					Subnet:       "0.0.0.0",
					PrefixLength: 27,
					Gateway:      "0.0.0.0",
				},
			},
		}

		if err := query.UpdateSubnetBucketRequest(ctx, tx, bucketPrivate); err != nil {
			logger.Error(err, "error storing bucket request into db")
			return nil, status.Errorf(codes.Internal, "database transaction failed")
		}
	}

	// delete bucket subnet from the database
	if err := query.DeleteBucketSubnet(ctx, tx, in); err != nil {
		return &emptypb.Empty{}, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &emptypb.Empty{}, nil
}

func (objSrv *BucketsServiceServer) GetBucketSubnetEvent(in *pb.SubnetEventRequest, rs pb.ObjectStorageServicePrivate_GetBucketSubnetEventServer) error {
	ctx, logger, span := obs.LogAndSpanFromContext(rs.Context()).WithName("BucketsServiceServer.GetBucketSubnetEvent").Start()
	defer span.End()

	logger.Info("entering bucket private service get subnet events")
	defer logger.Info("returning from bucket private service get subnet events")

	dbSession := objSrv.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	defer tx.Rollback()

	subnetEvents, err := query.GetAllBucketSubnetEvents(ctx, tx)
	if err != nil {
		return status.Errorf(codes.Internal, "error reading bucket subnet events")
	}

	logger.Info("total number of subnet events", logkeys.NumSubnetEvents, len(subnetEvents))
	for _, subnet := range subnetEvents {
		if err := rs.Send(subnet); err != nil {
			logger.Error(err, "error sending subnet event")
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return status.Errorf(codes.Internal, "database transaction failed")
	}

	return nil
}

func (objSrv *BucketsServiceServer) UpdateBucketSubnetStatus(ctx context.Context, in *pb.BucketSubnetStatusUpdateRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.UpdateBucketSubnetStatus").WithValues(logkeys.CloudAccountId, in.CloudacccountId).Start()
	defer span.End()

	logger.Info("entering bucket private service for update status subnet")
	defer logger.Info("returning from bucket private service for update status subnet")

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	defer tx.Rollback()

	logger.Info("Update params", logkeys.Input, in)

	if in.Status == pb.BucketSubnetEventStatus_E_DELETED {
		if err := query.DeleteBucketSubnetFromDB(ctx, tx, in); err != nil {
			logger.Error(err, "error deleting subnet from bucket")
			return nil, fmt.Errorf("error updating bucket status for delete")
		}
	} else {
		if err := query.UpdateStatusForSubnet(ctx, tx, in); err != nil {
			logger.Error(err, "error updating delete timestamp for bucket subnet")
			return nil, fmt.Errorf("error updating bucket status for update")
		}
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &emptypb.Empty{}, nil
}
