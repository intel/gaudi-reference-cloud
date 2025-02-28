package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (objSrv *BucketsServiceServer) createObjectUser(ctx context.Context, in *pb.CreateObjectUserPrivateRequest) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.createObjectUser")
	logger.Info("entering bucket create user private for ", logkeys.Input, in)

	if err := isValidBucketUserCreateRequest(ctx, in); err != nil {
		return nil, err
	}

	// store user in the database
	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Rollback()

	// validate the input arguments

	// test if user with the same name exists
	existingUser, err := query.GetBucketUserByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.Name, timestampInfinityStr)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			logger.Info("bucket user does not exist, ready to create user ")
		default:
			return nil, status.Errorf(codes.FailedPrecondition, "quering bucket user failed")
		}
	}

	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "bucket user name "+in.Metadata.Name+" already exists")
	}
	// check if the bucket exists for this cloudaccount
	clusterUUID := ""
	clusterName := ""
	accessEndpoint := ""
	networkGroup := []*pb.BucketNetworkGroup{}
	var bucketIds []string
	for _, policy := range in.Spec {
		bucketId := policy.BucketId
		buckeyQueryRequest := pb.ObjectBucketGetPrivateRequest{
			Metadata: &pb.ObjectBucketMetadataRef{
				CloudAccountId: in.Metadata.CloudAccountId,
				NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
					BucketName: bucketId,
				},
			},
		}
		accBucket, err := objSrv.getBucket(ctx, &buckeyQueryRequest)
		if err != nil {
			logger.Error(err, "error quering bucket ")
			return nil, err
		}
		bucketIds = append(bucketIds, accBucket.Metadata.ResourceId)
		// We are assuming there will be a single clusterUUID for all buckets
		clusterUUID = accBucket.Spec.Schedule.Cluster.ClusterUUID
		clusterName = accBucket.Spec.Schedule.Cluster.ClusterName
		accessEndpoint = accBucket.Spec.Schedule.Cluster.ClusterAddr
		if accBucket.Status.SecurityGroup != nil {
			networkGroup = accBucket.Status.SecurityGroup.NetworkFilterAllow
		}
	}

	// check if member has access to policy buckets
	if objSrv.cfg.AuthzEnabled {
		//check if private request
		_, err := grpcutil.ExtractClaimFromCtx(ctx, true, grpcutil.EmailClaim)
		if err != nil && err == grpcutil.ErrNoJWTToken {
			logger.Info("skipping auth check for private request")
		} else {
			//call lookup on bucketids
			for _, bucketId := range bucketIds {
				_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "createprincipal", "objectstorage", []string{bucketId}, false)
				if err != nil {
					return nil, err
				}
			}
		}

	}

	user := pb.ObjectUserPrivate{}
	user.Status = &pb.ObjectUserStatusPrivate{
		Phase: pb.ObjectUserPhase_ObjectUserProvisioning,
	}
	accessKey := utils.GenerateRandomString(10)
	secretKey, err := utils.GenerateRandomPassword()
	if err != nil {
		logger.Error(err, "error generating password")
		user.Status.Phase = pb.ObjectUserPhase_ObjectUserFailed
		return &user, status.Errorf(codes.Internal, "error creating user")
	}
	userParams := pb.CreateBucketUserParams{
		CloudAccountId: in.Metadata.CloudAccountId,
		CreateParams: &pb.BucketUserParams{
			Name:           in.Metadata.Name,
			ClusterUUID:    clusterUUID,
			ClusterName:    clusterName,
			AccessEndpoint: accessEndpoint,
			UserId:         accessKey,
			Password:       secretKey,
			Spec:           in.Spec,
			SecurityGroup: &pb.BucketSecurityGroup{
				NetworkFilterAllow: networkGroup,
			},
		},
	}

	userCreated, err := objSrv.bucketUserClient.CreateBucketUser(ctx, &userParams)
	if err != nil {
		logger.Error(err, "error creating user")
		user.Status.Phase = pb.ObjectUserPhase_ObjectUserFailed
		return &user, err
	}

	user.Metadata = &pb.ObjectUserMetadataPrivate{
		CloudAccountId:    in.Metadata.CloudAccountId,
		Name:              in.Metadata.Name,
		UserId:            uuid.NewString(),
		CreationTimestamp: timestamppb.Now(),
	}
	user.Spec = in.Spec
	user.Status.Principal = &pb.AccessPrincipalPrivate{
		Cluster: &pb.ObjectCluster{
			ClusterId:      userCreated.ClusterId,
			AccessEndpoint: userCreated.AccessEndpoint,
			ClusterName:    userCreated.ClusterName,
		},
		Credentials: &pb.ObjectAccessCredentials{
			AccessKey: accessKey,
		},
		PrincipalId: userCreated.PrincipalId,
	}
	user.Status.Phase = pb.ObjectUserPhase_ObjectUserReady

	if err := query.StoreBucketUserRequest(ctx, tx, &user); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == kErrUniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "bucket user name "+user.Metadata.Name+" already exists")
		}
		logger.Error(err, "error storing bucket user request into db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	user.Status.Principal.Credentials.SecretKey = secretKey

	logger.Info("bucket user created successfully")
	return &user, nil
}

func (objSrv *BucketsServiceServer) getObjectUser(ctx context.Context, in *pb.ObjectUserGetPrivateRequest) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.getObjectUser")

	logger.Info("entering bucket get user private for", logkeys.Input, in)
	defer logger.Info("returning from get user private")

	if err := isValidBucketUserGetRequest(ctx, in); err != nil {
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

	userPrivate := &pb.ObjectUserPrivate{}
	if in.Metadata.GetUserName() != "" {
		userPrivate, err = query.GetBucketUserByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserName(), timestampInfinityStr)
	} else if in.Metadata.GetUserId() != "" {
		userPrivate, err = query.GetBucketUserById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserId(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}
	return userPrivate, nil
}

func (objSrv *BucketsServiceServer) deleteBucketUser(ctx context.Context,
	in *pb.ObjectUserDeletePrivateRequest) (*emptypb.Empty, error) {

	logger := log.FromContext(ctx).WithName("BucketsServiceServer.deleteBucketUser")
	logger.Info("entering bucket user delete for ", logkeys.CloudAccountId, in.Metadata.CloudAccountId)
	defer logger.Info("returning from bucket user delete")
	if err := isValidBucketUserDeleteRequest(ctx, in); err != nil {
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

	userPrivate := &pb.ObjectUserPrivate{}
	if in.Metadata.GetUserName() != "" {
		userPrivate, err = query.GetBucketUserByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserName(), timestampInfinityStr)
	} else if in.Metadata.GetUserId() != "" {
		userPrivate, err = query.GetBucketUserById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserId(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	if userPrivate.Status == nil ||
		userPrivate.Status.Principal == nil ||
		userPrivate.Status.Principal.Cluster == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to query user data")
	}
	// check if member has access to policy buckets
	if objSrv.cfg.AuthzEnabled {
		//check if private request
		var isPrivate bool
		_, err := grpcutil.ExtractClaimFromCtx(ctx, true, grpcutil.EmailClaim)
		if err != nil && err == grpcutil.ErrNoJWTToken {
			logger.Info("skipping auth check for private request")
			isPrivate = true
		}
		//retrieve all bucket resource ids from db
		for _, policy := range userPrivate.Spec {
			bucketPrivate, err := query.GetBucketByName(ctx, tx, userPrivate.Metadata.CloudAccountId, policy.BucketId, timestampInfinityStr)
			if err != nil {
				//continue if bucket not found, bucket has been deleted
				if status.Code(err) == codes.NotFound {
					continue
				}
				return nil, err
			}
			//call lookup on bucketids
			if !isPrivate {
				_, err = objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "deleteprincipal", "objectstorage", []string{bucketPrivate.Metadata.ResourceId}, false)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	deleteParams := pb.DeleteBucketUserParams{
		CloudAccountId: in.Metadata.CloudAccountId,
		ClusterId:      userPrivate.Status.Principal.Cluster.ClusterId,
		PrincipalId:    userPrivate.Status.Principal.PrincipalId,
	}
	if _, err = objSrv.bucketUserClient.DeleteBucketUser(ctx, &deleteParams); err != nil {
		return nil, err
	}

	err = query.UpdateBucketUserForDeletion(ctx, tx, in.Metadata)
	if err != nil {
		logger.Error(err, "error updating bucket for deletion")
		return &emptypb.Empty{}, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	logger.Info("bucket user deleted successfully")
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.authzServiceClient.RemoveResourceFromCloudAccountRole(ctx,
			&v1.CloudAccountRoleResourceRequest{
				CloudAccountId: in.Metadata.CloudAccountId,
				ResourceId:     userPrivate.Metadata.GetUserId(),
				ResourceType:   "principal"})
		if err != nil {
			logger.Error(err, "error authz when removing resource from cloudAccountRole", "resourceId", userPrivate.Metadata.GetUserId(), "resourceType", "principal")
			return nil, status.Errorf(codes.Internal, "principal %v couldn't be removed from cloudAccountRole resourceType %v",
				userPrivate.Metadata.GetUserId(), "principal")
		}
	}
	return &emptypb.Empty{}, nil
}

func getBucketUserAccess(ctx context.Context, tx *sql.Tx, cloudaccount, bucketName string) ([]*pb.BucketUserAccess, error) {
	logger := log.FromContext(ctx).WithName("getBucketUserAccess")
	accountUsers, err := query.GetBucketUsersByCloudaccountId(ctx, tx, cloudaccount, timestampInfinityStr)
	if err != nil {
		logger.Error(err, "error reading user access policies")
		return nil, status.Errorf(codes.Internal, "error reading bucket access users")
	}

	userAccessPolicies := []*pb.BucketUserAccess{}
	for _, accUser := range accountUsers {
		for _, spec := range accUser.Spec {
			if strings.EqualFold(spec.BucketId, bucketName) {
				userAccess := pb.BucketUserAccess{
					Metadata: (*pb.ObjectUserMetadata)(accUser.Metadata),
					Spec:     []*pb.ObjectUserPermissionSpec{spec},
				}
				userAccessPolicies = append(userAccessPolicies, &userAccess)
			}
		}
	}
	return userAccessPolicies, nil
}

func (objSrv *BucketsServiceServer) updateObjectUserPolicies(ctx context.Context, in *pb.ObjectUserUpdatePrivateRequest) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.updateObjectUserPolicies")
	logger.Info("entering bucket update user private for ", logkeys.Input, in)
	defer logger.Info("returning from update create user private")

	if err := isValidBucketUserUpdateRequest(ctx, in); err != nil {
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
	userPrivate := &pb.ObjectUserPrivate{}
	if in.Metadata.GetUserName() != "" {
		userPrivate, err = query.GetBucketUserByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserName(), timestampInfinityStr)
	} else if in.Metadata.GetUserId() != "" {
		userPrivate, err = query.GetBucketUserById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserId(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	// check if policy list is empty
	if len(in.Spec) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no new policies supplied")
	}

	// check if member has access to policy buckets
	if objSrv.cfg.AuthzEnabled {
		//retrieve all bucket resource ids from db
		for _, policy := range userPrivate.Spec {
			bucketPrivate, err := query.GetBucketByName(ctx, tx, userPrivate.Metadata.CloudAccountId, policy.BucketId, timestampInfinityStr)
			if err != nil {
				return nil, err
			}
			//call lookup on bucketids
			_, err = objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "updateprincipal", "objectstorage", []string{bucketPrivate.Metadata.ResourceId}, false)
			if err != nil {
				return nil, err
			}
		}
	}

	// check if the bucket exists for this cloudaccount
	networkGroup := []*pb.BucketNetworkGroup{}
	for _, policy := range in.Spec {
		bucketId := policy.BucketId
		buckeyQueryRequest := pb.ObjectBucketGetPrivateRequest{
			Metadata: &pb.ObjectBucketMetadataRef{
				CloudAccountId: in.Metadata.CloudAccountId,
				NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
					BucketName: bucketId,
				},
			},
		}
		accBucket, err := objSrv.getBucket(ctx, &buckeyQueryRequest)
		if err != nil {
			logger.Error(err, "error quering bucket ")
			return nil, err
		}
		// We are assuming there will be a single clusterUUID for all buckets
		if accBucket == nil {
			logger.Error(err, "error quering bucket, bucket not found")
			return nil, status.Errorf(codes.FailedPrecondition, "bucket not found")
		}
		if accBucket.Status.SecurityGroup != nil {
			networkGroup = accBucket.Status.SecurityGroup.NetworkFilterAllow
		}
	}

	params := pb.UpdateBucketUserPolicyParams{
		CloudAccountId: in.Metadata.CloudAccountId,
		UpdateParams: &pb.BucketUpdateUserPolicyParams{
			PrincipalId: userPrivate.Status.Principal.PrincipalId,
			ClusterUUID: userPrivate.Status.Principal.Cluster.ClusterId,
			Spec:        in.Spec,
			SecurityGroup: &pb.BucketSecurityGroup{
				NetworkFilterAllow: networkGroup,
			},
			ClusterName:    userPrivate.Status.Principal.Cluster.ClusterName,
			AccessEndpoint: userPrivate.Status.Principal.Cluster.AccessEndpoint,
		},
	}

	_, err = objSrv.bucketUserClient.UpdateBucketUserPolicy(ctx, &params)
	if err != nil {
		logger.Error(err, "error updating user policies")
		userPrivate.Status.Phase = pb.ObjectUserPhase_ObjectUserFailed
		return userPrivate, status.Errorf(codes.Internal, "error updating user policies")
	}

	userPrivate.Spec = in.Spec
	userPrivate.Metadata.UpdateTimestamp = timestamppb.Now()
	if err := query.UpdateBucketUserPolicy(ctx, tx, userPrivate, timestampInfinityStr); err != nil {
		logger.Error(err, "error updating bucket user request into db")
		// DO NOT Return error here, since user policy is already updated
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		// DO NOT Return error here, since user policy is already updated
	}

	return userPrivate, nil
}

func (objSrv *BucketsServiceServer) updateObjectUserCredentials(ctx context.Context, in *pb.ObjectUserUpdateCredsRequest) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("BucketsServiceServer.updateObjectUserCredentials")

	logger.Info("entering bucket update user crdentials for", logkeys.Input, in)
	defer logger.Info("returning from update user crdentials")

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Rollback()
	userPrivate := &pb.ObjectUserPrivate{}
	if in.Metadata.GetUserName() != "" {
		userPrivate, err = query.GetBucketUserByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserName(), timestampInfinityStr)
	} else if in.Metadata.GetUserId() != "" {
		userPrivate, err = query.GetBucketUserById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetUserId(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	accessKey := userPrivate.Status.Principal.Credentials.AccessKey
	secretKey, err := utils.GenerateRandomPassword()
	if err != nil {
		logger.Error(err, "error generating password")
		return nil, status.Errorf(codes.Internal, "error generating access credentials")
	}

	params := pb.UpdateBucketUserCredsParams{
		CloudAccountId: in.Metadata.CloudAccountId,
		UpdateParams: &pb.BucketUpdateUserCredsParams{
			PrincipalId: userPrivate.Status.Principal.PrincipalId,
			ClusterUUID: userPrivate.Status.Principal.Cluster.ClusterId,
			UserId:      accessKey,
			Password:    secretKey,
		},
	}

	userUpdated, err := objSrv.bucketUserClient.UpdateBucketUserCredentials(ctx, &params)
	if err != nil {
		logger.Error(err, "error updating user credentials")
		userPrivate.Status.Phase = pb.ObjectUserPhase_ObjectUserFailed
		return userPrivate, status.Errorf(codes.Internal, "error updating user credentials")
	}

	logger.Info("updated user credentials", logkeys.PrincipalId, userUpdated.PrincipalId)
	userPrivate.Metadata.UpdateTimestamp = timestamppb.Now()
	if err := query.UpdateBucketUserPolicy(ctx, tx, userPrivate, timestampInfinityStr); err != nil {
		logger.Error(err, "error updating bucket user request into db")
		// DO NOT Return error here, since user policy is already updated
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		// DO NOT Return error here, since user policy is already updated
	}
	userPrivate.Status.Principal.Credentials.AccessKey = accessKey
	userPrivate.Status.Principal.Credentials.SecretKey = secretKey

	return userPrivate, nil
}
