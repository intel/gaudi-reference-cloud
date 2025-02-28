package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"google.golang.org/grpc/metadata"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type LifecycleRule struct {
	CloudAccountId string
	ClusterId      string
	BucketId       string
	Predicates     []Predicate
}
type Predicate struct {
	RuleId               string
	Prefix               string
	ExpireDays           uint32
	NoncurrentExpireDays uint32
	DeleteMarker         bool
}

// Create bucket lifecycle policy
func (client *StorageControllerClient) CreateBucketLifecycleRules(ctx context.Context, in LifecycleRule) (*LifecycleRule, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.CreateBucketLifecycleRules")
	lifeCycleRule := LifecycleRule{}
	//validate input
	if in.ClusterId == "" {
		logger.Info("empty cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}
	if in.BucketId == "" {
		logger.Info("empty bucket id")
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	rules := []*stcnt_api.LifecycleRule{}
	for _, p := range in.Predicates {
		lifecyelRule := stcnt_api.LifecycleRule{
			Prefix:               p.Prefix,
			ExpireDays:           p.ExpireDays,
			NoncurrentExpireDays: p.NoncurrentExpireDays,
			DeleteMarker:         p.DeleteMarker,
		}
		rules = append(rules, &lifecyelRule)
	}
	md := metadata.Pairs(
		"X-Cloud-Account", in.CloudAccountId,
	)
	// Create a context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)

	// make request
	logger.Info("CreateLifecycleRules sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.BucketId, in.BucketId, logkeys.Input, fmt.Sprintf("%v", rules))
	resp, err := client.S3ServiceClient.CreateLifecycleRules(ctx, &stcnt_api.CreateLifecycleRulesRequest{
		BucketId: &stcnt_api.BucketIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.BucketId,
		},
		LifecycleRules: rules,
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in creating lifecycle rule in controller")
		return nil, err
	}
	// format output
	lifeCycleRule.ClusterId = in.ClusterId
	lifeCycleRule.BucketId = in.BucketId

	for _, r := range resp.LifecycleRules {
		predicate := Predicate{
			RuleId:               r.Id.Id,
			Prefix:               r.Prefix,
			ExpireDays:           r.ExpireDays,
			NoncurrentExpireDays: r.NoncurrentExpireDays,
			DeleteMarker:         r.DeleteMarker,
		}
		lifeCycleRule.Predicates = append(lifeCycleRule.Predicates, predicate)
	}
	return &lifeCycleRule, nil
}

// helper validator for Lifecycle struct
func validateLifecycleRule(in LifecycleRule) error {
	//validate input
	if in.ClusterId == "" {
		return fmt.Errorf("cluster id cannot be empty")
	}
	if in.BucketId == "" {
		return fmt.Errorf("bucket id cannot be empty")
	}
	return nil
}
