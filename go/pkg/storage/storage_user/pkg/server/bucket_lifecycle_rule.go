package server

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// BucketLifecycleServiceServer is used to implement pb.UnimplementedFileStorageServiceServer
type BucketLifecycleServiceServer struct {
	pb.UnimplementedBucketLifecyclePrivateServiceServer

	strCntClient *storagecontroller.StorageControllerClient
}

func NewBucketLifecycleServiceServer(strCntCli *storagecontroller.StorageControllerClient) (*BucketLifecycleServiceServer, error) {
	if strCntCli == nil {
		return nil, fmt.Errorf("storage sds client is required")
	}
	return &BucketLifecycleServiceServer{
		strCntClient: strCntCli,
	}, nil
}

func (bucketRule *BucketLifecycleServiceServer) CreateOrUpdateLifecycleRule(ctx context.Context, in *pb.CreateOrUpdateLifecycleRuleRequest) (*pb.LifecycleRulePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BucketLifecycleServiceServer.CreateOrUpdateLifecycleRule").WithValues(logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.BucketId, in.BucketId).Start()
	defer span.End()

	logger.Info("create or update bucket rule", logkeys.BucketSpec, in.Spec)

	params := storagecontroller.LifecycleRule{
		CloudAccountId: in.CloudAccountId,
		ClusterId:      in.ClusterId,
		BucketId:       in.BucketId,
	}
	for _, rule := range in.Spec {
		params.Predicates = append(params.Predicates, storagecontroller.Predicate{
			Prefix:               rule.Prefix,
			ExpireDays:           rule.ExpireDays,
			NoncurrentExpireDays: rule.NoncurrentExpireDays,
			DeleteMarker:         rule.DeleteMarker,
		})
	}

	createRule := &storagecontroller.LifecycleRule{}
	var err error
	//create request
	logger.Info("processing as bucket create or update lifecycle rule")
	createRule, err = bucketRule.strCntClient.CreateBucketLifecycleRules(ctx, params)
	if err != nil {
		logger.Error(err, "error creating/updating bucket lifecycle rule")
		return nil, status.Errorf(codes.Internal, "error creating/updating bucket lifecycle rule")
	}
	logger.Info("bucket rule created/updated successfully")

	lfPrivate := pb.LifecycleRulePrivate{
		ClusterId: createRule.ClusterId,
		BucketId:  createRule.BucketId,
		Spec:      in.Spec,
	}
	return &lfPrivate, nil
}

func (bucketRule *BucketLifecycleServiceServer) PingBucketLifecyclePrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("BucketLifecycleServiceServer.PingBucketLifecyclePrivate")
	logger.Info("ping private bucket lifecycle")

	return &emptypb.Empty{}, nil
}
