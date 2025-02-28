package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type BucketAccessPolicy int32

const (
	BucketAccessPolicyReadWrite   BucketAccessPolicy = 3
	BucketAccessPolicyReadOnly    BucketAccessPolicy = 2
	BucketAccessPolicyUnspecified BucketAccessPolicy = 0
	BucketAccessPolicyNone        BucketAccessPolicy = 1
)

type Bucket struct {
	Metadata BucketMetadata
	Spec     BucketSpec
}

type BucketMetadata struct {
	Name      string
	BucketId  string
	ClusterId string
}

type BucketSpec struct {
	AccessPolicy   BucketAccessPolicy
	Versioned      bool
	Totalbytes     uint64
	AvailableBytes uint64
	EndpointUrl    string
}

type BucketPolicy struct {
	ClusterId string
	BucketId  string
	Policy    BucketAccessPolicy
}

type BucketFilter struct {
	ClusterId string
	BucketId  string
}

// Provision a new s3 bucket
func (client *StorageControllerClient) CreateBucket(ctx context.Context, in Bucket) (*Bucket, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.CreateBucket")
	logger.Info("inside CreateBucket func")

	bucket := &Bucket{}
	// Input validation
	if in.Metadata.Name == "" {
		logger.Info("empty bucket name")
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	if in.Spec.Totalbytes == 0 {
		logger.Info("zero request size")
		return nil, fmt.Errorf("requested bytes cannot be 0")
	}
	// clusterUUID
	if in.Metadata.ClusterId == "" {
		logger.Info("empty cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}

	logger.Info("Preparing to call minio backend")
	// Create new bucket in the `Cluster`
	logger.Info("CreateBucket sds input params", logkeys.ClusterId, in.Metadata.ClusterId, logkeys.BucketName, in.Metadata.Name, logkeys.AccessPolicy, in.Spec.AccessPolicy, logkeys.BucketCapacity, in.Spec.Totalbytes)
	resp, err := client.S3ServiceClient.CreateBucket(ctx, &stcnt_api.CreateBucketRequest{
		ClusterId: &stcnt_api.ClusterIdentifier{
			Uuid: in.Metadata.ClusterId,
		},
		Name:         in.Metadata.Name,
		AccessPolicy: stcnt_api.BucketAccessPolicy(in.Spec.AccessPolicy),
		Versioned:    in.Spec.Versioned,
		QuotaBytes:   in.Spec.Totalbytes,
	})
	if err != nil {
		logger.Error(err, "error in creating bucket in controller")
		//return pointer
		return nil, fmt.Errorf("create bucket error")
	}

	logger.Info("formatting response")
	// format resp
	bucket = &Bucket{
		Metadata: BucketMetadata{
			BucketId:  resp.Bucket.Id.Id,
			ClusterId: resp.Bucket.Id.ClusterId.Uuid,
			Name:      resp.Bucket.Name,
		},
		Spec: BucketSpec{
			Versioned:      resp.Bucket.Versioned,
			AvailableBytes: resp.Bucket.Capacity.AvailableBytes,
			Totalbytes:     resp.Bucket.Capacity.TotalBytes,
			EndpointUrl:    resp.Bucket.EndpointUrl,
		},
	}
	return bucket, nil
}

// Request deletion of a bucket.
func (client *StorageControllerClient) DeleteBucket(ctx context.Context, in BucketMetadata, force bool) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.DeleteBucket")
	logger.Info("inside DeleteBucket func")

	// Verify bucket_id is not nil
	if in.BucketId == "" {
		logger.Info("invalid bucket id")
		return fmt.Errorf("bucket id cannot be empty")
	}
	if in.ClusterId == "" {
		logger.Info("invalid cluster id")
		return fmt.Errorf("cluster id cannot be empty")
	}
	// Contruct request struct
	logger.Info("DeleteBucket sds input params", logkeys.ClusterId, in.ClusterId, logkeys.BucketId, in.BucketId)
	_, err := client.S3ServiceClient.DeleteBucket(ctx, &stcnt_api.DeleteBucketRequest{
		BucketId: &stcnt_api.BucketIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.BucketId,
		},
		Force: force,
	})
	if err != nil {
		logger.Error(err, "error in deleting bucket in controller")
		return err
	}
	return nil
}

// Get policy of a bucket.
func (client *StorageControllerClient) GetBucketPolicy(ctx context.Context, in BucketMetadata) (*BucketPolicy, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetBucketPolicy")
	logger.Info("inside GetBucketPolicy func")
	policy := &BucketPolicy{}
	// validate inputs
	if in.BucketId == "" {
		logger.Info("invalid bucket id")
		return nil, fmt.Errorf("bucket id cannot be empty")
	}
	if in.ClusterId == "" {
		logger.Info("invalid cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}

	// make request call
	resp, err := client.S3ServiceClient.GetBucketPolicy(ctx, &stcnt_api.GetBucketPolicyRequest{
		BucketId: &stcnt_api.BucketIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.BucketId,
		},
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in getting bucket policy in controller")
		return nil, err
	}

	// format output
	policy = &BucketPolicy{
		ClusterId: resp.BucketId.ClusterId.Uuid,
		BucketId:  resp.BucketId.Id,
		Policy:    BucketAccessPolicy(resp.Policy),
	}

	return policy, nil

}

// Update policy of a bucket.
func (client *StorageControllerClient) UpdateBucketPolicy(ctx context.Context, in BucketPolicy) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.UpdateBucketPolicy")
	logger.Info("inside UpdateBucketPolicy func")
	// validate inputs
	if in.BucketId == "" {
		logger.Info("invalid bucket id")
		return fmt.Errorf("bucket id cannot be empty")
	}
	if in.ClusterId == "" {
		logger.Info("invalid cluster id")
		return fmt.Errorf("cluster id cannot be empty")
	}

	// make request call
	_, err := client.S3ServiceClient.UpdateBucketPolicy(ctx, &stcnt_api.UpdateBucketPolicyRequest{
		BucketId: &stcnt_api.BucketIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.BucketId,
		},
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in updating bucket policy in controller")
		return err
	}

	return nil
}

// Retrieval of a bucket.
func (client *StorageControllerClient) GetBucket(ctx context.Context, in BucketFilter) (*Bucket, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetBucket")
	logger.Info("inside GetBucket func")

	// Verify bucket_id is not nil
	if in.BucketId == "" {
		logger.Info("invalid bucket id")
		return nil, fmt.Errorf("bucket id cannot be empty")
	}
	if in.ClusterId == "" {
		logger.Info("invalid cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}
	// Contruct request struct
	logger.Info("ListBuckets sds input params", logkeys.ClusterId, in.ClusterId, logkeys.BucketId, in.BucketId)
	bucketList, err := client.S3ServiceClient.ListBuckets(ctx, &stcnt_api.ListBucketsRequest{
		ClusterId: &stcnt_api.ClusterIdentifier{
			Uuid: in.ClusterId,
		},
		Filter: &stcnt_api.ListBucketsRequest_Filter{
			Names: []string{in.BucketId},
		},
	})
	if err != nil {
		logger.Error(err, "error in searching bucket in controller")
		return nil, err
	}
	if len(bucketList.Buckets) > 1 {
		return nil, fmt.Errorf("error searching bucket, more than 1 bucket found")
	}
	if len(bucketList.Buckets) == 0 {
		return nil, fmt.Errorf("error searching bucket, bucket not found")
	}
	return castToBucket(bucketList.Buckets[0]), nil
}

// Request list of all buckets.
func (client *StorageControllerClient) GetAllBuckets(ctx context.Context, in BucketFilter) ([]*Bucket, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetAllBuckets")
	logger.Info("inside GetBucket func")

	// Verify cluster_id is not nil
	if in.ClusterId == "" {
		logger.Info("invalid cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}
	// Contruct request struct
	logger.Info("ListBuckets sds input params", logkeys.ClusterId, in.ClusterId)
	bucketList, err := client.S3ServiceClient.ListBuckets(ctx, &stcnt_api.ListBucketsRequest{
		ClusterId: &stcnt_api.ClusterIdentifier{
			Uuid: in.ClusterId,
		},
	})
	if err != nil {
		logger.Error(err, "error in searching buckets in controller")
		return nil, err
	}
	resp := []*Bucket{}
	for _, bk := range bucketList.Buckets {
		resp = append(resp, castToBucket(bk))
	}
	return resp, nil
}

func castToBucket(strnBucket *stcnt_api.Bucket) *Bucket {
	bucket := &Bucket{
		Metadata: BucketMetadata{
			BucketId:  strnBucket.Id.Id,
			ClusterId: strnBucket.Id.ClusterId.Uuid,
			Name:      strnBucket.Name,
		},
		Spec: BucketSpec{
			Versioned:      strnBucket.Versioned,
			AvailableBytes: strnBucket.Capacity.AvailableBytes,
			Totalbytes:     strnBucket.Capacity.TotalBytes,
			EndpointUrl:    strnBucket.EndpointUrl,
		},
	}
	return bucket
}
