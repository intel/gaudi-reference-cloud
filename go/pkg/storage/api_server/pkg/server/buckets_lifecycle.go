package server

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type lifecycleRuleAction int

const (
	create lifecycleRuleAction = iota + 1

	update

	delete
)

func (objSrv *BucketsServiceServer) createBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleCreatePrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.createBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("creating new lifecycle rule", logkeys.BucketId, in.Metadata.BucketId)

	if err := isValidLifecycleCreateRequest(in); err != nil {
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
	// test if lifecycle rule with the same name exists
	existingRule, err := query.GetBucketLifecycleRuleByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.RuleName, in.Metadata.BucketId)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			logger.Info("bucket lifecycle rule does not exist, ready to create lifecycle rule ")
		default:
			return nil, status.Errorf(codes.FailedPrecondition, "quering bucket lifecycle rule failed")
		}
	}
	if existingRule != nil {
		return nil, status.Error(codes.AlreadyExists, "bucket lifecycle rule name "+in.Metadata.RuleName+" already exists")
	}

	if err := objSrv.update(ctx, in.Metadata.CloudAccountId, in.Metadata.BucketId, "", int(create), in.Spec); err != nil {
		logger.Error(err, "error updating lifecycle rule")
		return nil, err
	}

	lfrulePrivate := pb.BucketLifecycleRulePrivate{
		Metadata: in.Metadata,
		Spec:     in.Spec,
	}
	lfrulePrivate.Status = &pb.BucketLifecycleRuleStatus{
		Phase: pb.BucketLifecycleRuleSPhase_LFRuleReady,
	}
	lfrulePrivate.Metadata.ResourceId = uuid.NewString()
	lfrulePrivate.Metadata.CreationTimestamp = timestamppb.Now()

	if err := query.InsertBucketLifecycleRequest(ctx, tx, &lfrulePrivate); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == kErrUniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "bucket lifecycle rule name "+lfrulePrivate.Metadata.RuleName+" already exists")
		}
		logger.Error(err, "error storing bucket lifecycle rule request into db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	return &lfrulePrivate, nil
}

func (objSrv *BucketsServiceServer) update(ctx context.Context, cloudAccountId, bucketId, ruleId string, action int, updatedRule *pb.BucketLifecycleRuleSpec) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.getBucketLifecycleRule").WithValues(logkeys.CloudAccountId, cloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket lifecycle rule update", logkeys.BucketId, bucketId)

	if action != int(delete) {
		if err := isValidLifecycleUpdateRequest(ctx, updatedRule); err != nil {
			return err
		}
	}

	// check if the bucket exists for this cloudaccount
	buckeyQueryRequest := pb.ObjectBucketGetPrivateRequest{
		Metadata: &pb.ObjectBucketMetadataRef{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.ObjectBucketMetadataRef_BucketId{
				BucketId: bucketId,
			},
		},
	}
	accBucket, err := objSrv.getBucket(ctx, &buckeyQueryRequest)
	if err != nil {
		return err
	}
	logger.Info("account verification of the bucket completed successfully", logkeys.BucketMetadata, accBucket.Metadata)
	inArgs := pb.CreateOrUpdateLifecycleRuleRequest{
		CloudAccountId: cloudAccountId,
		ClusterId:      accBucket.Spec.Schedule.Cluster.ClusterUUID,
		BucketId:       accBucket.Metadata.Name,
	}
	if action == int(create) {
		for _, rule := range accBucket.Status.Policy.LifecycleRules {
			inArgs.Spec = append(inArgs.Spec, rule.Spec)
		}
		inArgs.Spec = append(inArgs.Spec, updatedRule)
	} else if action == int(delete) {
		for _, rule := range accBucket.Status.Policy.LifecycleRules {
			if rule.Metadata.ResourceId != ruleId {
				inArgs.Spec = append(inArgs.Spec, rule.Spec)
			}
		}
	} else if action == int(update) {
		for _, rule := range accBucket.Status.Policy.LifecycleRules {
			if rule.Metadata.ResourceId == ruleId {
				inArgs.Spec = append(inArgs.Spec, updatedRule)
			} else {
				inArgs.Spec = append(inArgs.Spec, rule.Spec)
			}
		}
	}

	_, err = objSrv.lifecycleRuleClient.CreateOrUpdateLifecycleRule(ctx, &inArgs)
	if err != nil {
		return err
	}

	logger.Info("lifecycle rule updated successfully")
	return nil
}

func (objSrv *BucketsServiceServer) getBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleGetPrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.getBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering bucket get lifecycle rule private for", logkeys.Input, in)
	defer logger.Info("returning from get lifecycle rule private for")

	if err := isValidBucketLifecycleGetRequest(ctx, in.Metadata); err != nil {
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

	lfPrivate, err := query.GetBucketLifecycleRuleById(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.RuleId, in.Metadata.BucketId)
	if err != nil {
		return nil, err
	}

	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", "lifecyclerule", []string{lfPrivate.GetMetadata().ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}
	return lfPrivate, nil
}

func (objSrv *BucketsServiceServer) deleteBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleDeletePrivateRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.deleteBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("deleting lifecycle rule for", logkeys.BucketId, in.Metadata.BucketId)

	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db txn")
	}
	defer tx.Rollback()

	// bucketId := in.Metadata.BucketId
	// buckeyQueryRequest := pb.ObjectBucketGetPrivateRequest{
	// 	Metadata: &pb.ObjectBucketMetadataRef{
	// 		CloudAccountId: in.Metadata.CloudAccountId,
	// 		NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
	// 			BucketName: bucketId,
	// 		},
	// 	},
	// }
	// accBucket, err := objSrv.getBucket(ctx, &buckeyQueryRequest)
	// if err != nil {
	// 	logger.Error(err, "error quering bucket "+in.Metadata.BucketId)
	// 	return nil, err
	// }
	// logger.Info("account verification of the bucket completed successfully", "bucketMetadata", accBucket.Metadata)

	// deleteParams := pb.DeleteLifecycleRuleRequest{
	// 	ClusterId: accBucket.Spec.Schedule.Cluster.ClusterUUID,
	// 	BucketId:  accBucket.Metadata.BucketId,
	// 	RuleId:    in.Metadata.RuleId,
	// }

	if err := objSrv.update(ctx, in.Metadata.CloudAccountId, in.Metadata.BucketId, in.Metadata.RuleId, int(delete), nil); err != nil {
		logger.Error(err, "error updating lifecycle rule")
		return nil, status.Errorf(codes.Internal, "error deleting bucket lifecycle rule")
	}

	err = query.UpdateBucketLifecycleRulesForDeletion(ctx, tx, in.Metadata)
	if err != nil {
		logger.Error(err, "error updating bucket lifecycle rule for deletion in DB")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "error updating bucket lifecycle rule for deletion in DB")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	logger.Info("bucket lifecycle rule deleted successfully")
	if objSrv.cfg.AuthzEnabled {
		_, err := objSrv.authzServiceClient.RemoveResourceFromCloudAccountRole(ctx,
			&v1.CloudAccountRoleResourceRequest{
				CloudAccountId: in.Metadata.CloudAccountId,
				ResourceId:     in.Metadata.GetRuleId(),
				ResourceType:   "lifecyclerule"})
		if err != nil {
			logger.Error(err, "error authz when removing resource from cloudAccountRole", "resourceId", in.Metadata.GetRuleId(), "resourceType", "lifecyclerule")
			return nil, status.Errorf(codes.Internal, "lifecyclerule %v couldn't be removed from cloudAccountRole resourceType %v",
				in.Metadata.GetRuleId(), "lifecyclerule")
		}
	}
	return &emptypb.Empty{}, nil
}

func (objSrv *BucketsServiceServer) getAllBucketLifecycleRules(ctx context.Context, in *pb.BucketLifecycleRuleSearchRequest) (*pb.BucketLifecycleRuleSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.deleteBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.CloudAccountId).Start()
	defer span.End()
	logger.Info("get all lifecycle rules", logkeys.BucketId, in.BucketId)
	// validate input
	if in.BucketId != "" {
		if _, err := uuid.Parse(in.BucketId); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid bucketId")
		}
	}
	dbSession := objSrv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "database transaction failed")
	}
	defer tx.Rollback()
	lfrulePrivateList, err := query.GetBucketLifecycleRulesByCloudaccountId(ctx, tx,
		in.CloudAccountId, in.BucketId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error reading bucket lifecycle rules")
	}
	// commit transaction and release resource back to pool
	if err := tx.Commit(); err != nil {
		logger.Error(err, "Failed to commit transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	// check with authz service is permitted to access resource
	authzResourceIds := []string{}
	authzResourceIdMap := make(map[string]bool)
	if objSrv.cfg.AuthzEnabled {
		//check if private request
		_, err := grpcutil.ExtractClaimFromCtx(ctx, true, grpcutil.EmailClaim)
		if err != nil && err == grpcutil.ErrNoJWTToken {
			logger.Info("skipping auth check for private request")
		} else {
			for idx := 0; idx < len(lfrulePrivateList); idx++ {
				authzResourceIds = append(authzResourceIds, lfrulePrivateList[idx].Metadata.ResourceId)
			}
			authzLookupResponse, err := objSrv.lookupResourcePermission(ctx, in.CloudAccountId, "get", "lifecyclerule", authzResourceIds, true)
			if err != nil {
				return nil, err
			}
			for _, resourceId := range authzLookupResponse.ResourceIds {
				authzResourceIdMap[resourceId] = true
			}
		}
	}
	resp := pb.BucketLifecycleRuleSearchResponse{}

	for idx := 0; idx < len(lfrulePrivateList); idx++ {
		// this means that authz didn't find permission for this resourceId so will skip it in the response
		if _, exists := authzResourceIdMap[lfrulePrivateList[idx].Metadata.ResourceId]; objSrv.cfg.AuthzEnabled && !exists {
			continue
		}
		lfPrivate := lfrulePrivateList[idx]
		resp.Rules = append(resp.Rules, convertLifecycleRulePrivateToPublic(lfPrivate))
	}
	return &resp, nil
}

func (objSrv *BucketsServiceServer) updateBucketLifecycleRule(ctx context.Context, in *pb.BucketLifecycleRuleUpdatePrivateRequest) (*pb.BucketLifecycleRulePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketsServiceServer.updateBucketLifecycleRule").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("update bucket lifecycle rules for", logkeys.BucketId, in.Metadata.BucketId)

	if err := isValidLifecycleUpdateRequest(ctx, in.Spec); err != nil {
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

	// // test if lifecycle rule with the name exists
	existingRule, err := query.GetBucketLifecycleRuleById(ctx, tx,
		in.Metadata.CloudAccountId, in.Metadata.RuleId,
		in.Metadata.BucketId)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return nil, status.Errorf(codes.FailedPrecondition, "bucket rule not found")
		default:
			return nil, status.Errorf(codes.FailedPrecondition, "quering bucket lifecycle rule failed")
		}
	}

	if err := objSrv.update(ctx, in.Metadata.CloudAccountId, in.Metadata.BucketId, in.Metadata.RuleId, int(update), in.Spec); err != nil {
		logger.Error(err, "error updating lifecycle rule")
		return nil, err
	}

	// // check if the bucket exists for this cloudaccount
	// clusterUUID := ""

	// bucketId := in.Metadata.BucketId
	// buckeyQueryRequest := pb.ObjectBucketGetPrivateRequest{
	// 	Metadata: &pb.ObjectBucketMetadataRef{
	// 		CloudAccountId: in.Metadata.CloudAccountId,
	// 		NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
	// 			BucketName: bucketId,
	// 		},
	// 	},
	// }
	// accBucket, err := objSrv.getBucket(ctx, &buckeyQueryRequest)
	// if err != nil {
	// 	logger.Error(err, "error quering bucket "+in.Metadata.BucketId)
	// 	return nil, err
	// }
	// logger.Info("account verification of the bucket completed successfully", "bucketMetadata", accBucket.Metadata)

	// clusterUUID = accBucket.Spec.Schedule.Cluster.ClusterUUID
	// inArgs := pb.CreateOrUpdateLifecycleRuleRequest{
	// 	ClusterId:            clusterUUID,
	// 	BucketId:             in.Metadata.BucketId,
	// 	RuleId:               in.Metadata.RuleId,
	// 	Prefix:               in.Spec.Prefix,
	// 	ExpireDays:           in.Spec.ExpireDays,
	// 	NoncurrentExpireDays: in.Spec.NoncurrentExpireDays,
	// 	DeleteMarker:         in.Spec.DeleteMarker,
	// }

	// _, err = objSrv.lifecycleRuleClient.CreateOrUpdateLifecycleRule(ctx, &inArgs)
	// if err != nil {
	// 	logger.Error(err, "error creating bucket lifecycle rule")
	// 	return nil, status.Errorf(codes.Internal, "error creating bucket lifecycle rule")
	// }

	existingRule.Spec = in.Spec
	existingRule.Metadata.UpdateTimestamp = timestamppb.Now()

	if err := query.UpdateBucketLifecycleRequest(ctx, tx, existingRule); err != nil {
		logger.Error(err, "error persisting updated lifecycle rule")
		//do not return error
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		//do not return error
	}
	logger.Info("returning from bucket lifecycle rule update", logkeys.BucketId, in.Metadata.BucketId)
	return existingRule, nil
}
