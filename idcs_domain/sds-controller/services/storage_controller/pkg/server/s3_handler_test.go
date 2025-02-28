// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"net"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

type MockS3Backend struct {
	backend.Interface
}

// Make sure our backend implements s3 operations
var _ backend.S3Ops = &MockS3Backend{}

type MockNoS3Backend struct {
	backend.Interface
}

var bBucket = backend.Bucket{
	ID:             "bucketId",
	Name:           "bucket",
	AccessPolicy:   backend.ReadWrite,
	Versioned:      true,
	QuotaBytes:     5000000,
	AvailableBytes: 5000000,
	EndpointURL:    "endpoint1",
}

var bucketResponse = v1.Bucket{
	Id: &v1.BucketIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bBucket.ID,
	},
	Name:      bBucket.Name,
	Versioned: true,
	Capacity: &v1.Bucket_Capacity{
		TotalBytes:     5000000,
		AvailableBytes: 5000000,
	},
	EndpointUrl: "endpoint1",
}

func (*MockS3Backend) CreateBucket(ctx context.Context, opts backend.CreateBucketOpts) (*backend.Bucket, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	} else if opts.Name == "empty1" {
		return nil, nil
	}

	return &backend.Bucket{
		ID:             "bucketId",
		Name:           opts.Name,
		AccessPolicy:   opts.AccessPolicy,
		Versioned:      opts.Versioned,
		QuotaBytes:     opts.QuotaBytes,
		AvailableBytes: opts.QuotaBytes,
		EndpointURL:    "endpoint1",
	}, nil
}

func (*MockS3Backend) DeleteBucket(ctx context.Context, opts backend.DeleteBucketOpts) error {
	if opts.ID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockS3Backend) GetBucketPolicy(ctx context.Context, opts backend.GetBucketPolicyOpts) (*backend.AccessPolicy, error) {
	if opts.ID == "error3" {
		return nil, errors.New("something wrong")
	} else if opts.ID == "notFound3" {
		return nil, nil
	}

	return &bBucket.AccessPolicy, nil
}

func (*MockS3Backend) ListBuckets(ctx context.Context, opts backend.ListBucketsOpts) ([]*backend.Bucket, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	return []*backend.Bucket{&bBucket}, nil
}

func (*MockS3Backend) UpdateBucketPolicy(ctx context.Context, opts backend.UpdateBucketPolicyOpts) error {
	if opts.ID == "error5" {
		return errors.New("something wrong")
	}

	return nil
}

var bLifecycleRule = backend.LifecycleRule{
	ID:                   "ruleId",
	Prefix:               "pre",
	ExpireDays:           50,
	NoncurrentExpireDays: 20,
	DeleteMarker:         true,
}

var lifecycleRuleResponse = v1.LifecycleRule{
	Id: &v1.LifecycleRuleIdentifier{
		Id: "ruleId",
	},
	Prefix:               bLifecycleRule.Prefix,
	ExpireDays:           bLifecycleRule.ExpireDays,
	NoncurrentExpireDays: bLifecycleRule.NoncurrentExpireDays,
	DeleteMarker:         bLifecycleRule.DeleteMarker,
}

func (*MockS3Backend) CreateLifecycleRules(ctx context.Context, opts backend.CreateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	if opts.BucketID == "error1" {
		return nil, errors.New("something wrong")
	}

	rules := make([]*backend.LifecycleRule, 0)

	for _, rule := range opts.LifecycleRules {
		rules = append(rules, &backend.LifecycleRule{
			ID:                   "ruleId",
			Prefix:               rule.Prefix,
			ExpireDays:           rule.ExpireDays,
			NoncurrentExpireDays: rule.NoncurrentExpireDays,
			DeleteMarker:         rule.DeleteMarker,
		})
	}

	return rules, nil
}

func (*MockS3Backend) DeleteLifecycleRules(ctx context.Context, opts backend.DeleteLifecycleRulesOpts) error {
	if opts.BucketID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockS3Backend) ListLifecycleRules(ctx context.Context, opts backend.ListLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	if opts.BucketID == "error3" {
		return nil, errors.New("something wrong")
	}

	return []*backend.LifecycleRule{&bLifecycleRule}, nil
}

func (*MockS3Backend) UpdateLifecycleRules(ctx context.Context, opts backend.UpdateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	if opts.BucketID == "error4" {
		return nil, errors.New("something wrong")
	}

	return []*backend.LifecycleRule{&bLifecycleRule}, nil
}

var allowIPString = "127.0.0.0/24"
var disallowIPString = "127.0.0.0/32"
var _, allowNet, _ = net.ParseCIDR(allowIPString)
var _, disallowNet, _ = net.ParseCIDR(disallowIPString)

var bPrincipal = backend.S3Principal{
	ID:   "principalId",
	Name: "name",
	Policies: []*backend.S3Policy{
		{
			BucketID: bBucket.ID,
			Prefix:   "pre",
			Read:     true,
			Write:    true,
			Delete:   true,
			Actions: []string{
				"s3:GetBucketLocation",
				"s3:GetBucketPolicy",
				"s3:ListBucket",
				"s3:ListBucketMultipartUploads",
				"s3:ListMultipartUploadParts",
				"s3:GetBucketTagging",
				"s3:ListBucketVersions",
			},
			AllowSourceNets: []*net.IPNet{
				allowNet,
			},
			DisallowSourceNets: []*net.IPNet{
				disallowNet,
			},
		},
	},
}

var principalResponse = v1.S3Principal{
	Id: &v1.S3PrincipalIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bPrincipal.ID,
	},
	Name: bPrincipal.Name,
	Policies: []*v1.S3Principal_Policy{
		{
			BucketId: bucketResponse.GetId(),
			Prefix:   "pre",
			Read:     true,
			Write:    true,
			Delete:   true,
			Actions: []v1.S3Principal_Policy_BucketActions{
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION,
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_MULTIPART_UPLOADS,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_MULTIPART_UPLOAD_PARTS,
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_TAGGING,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_VERSIONS,
			},
			SourceIpFilter: &v1.S3Principal_Policy_SourceIpFilter{
				Allow: []string{
					allowIPString,
				},
				Disallow: []string{
					disallowIPString,
				},
			},
		},
	},
}

func (*MockS3Backend) CreateS3Principal(ctx context.Context, opts backend.CreateS3PrincipalOpts) (*backend.S3Principal, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	return &backend.S3Principal{
		ID:       bPrincipal.ID,
		Name:     opts.Name,
		Policies: []*backend.S3Policy{},
	}, nil
}

func (*MockS3Backend) DeleteS3Principal(ctx context.Context, opts backend.DeleteS3PrincipalOpts) error {
	if opts.PrincipalID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockS3Backend) GetS3Principal(ctx context.Context, opts backend.GetS3PrincipalOpts) (*backend.S3Principal, error) {
	if opts.PrincipalID == "error3" {
		return nil, errors.New("something wrong")
	} else if opts.PrincipalID == "empty3" {
		return nil, nil
	}

	return &bPrincipal, nil
}

func (*MockS3Backend) UpdateS3PrincipalPassword(ctx context.Context, opts backend.UpdateS3PrincipalPasswordOpts) error {
	if opts.PrincipalID == "error4" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockS3Backend) UpdateS3PrincipalPolicies(ctx context.Context, opts backend.UpdateS3PrincipalPoliciesOpts) (*backend.S3Principal, error) {
	if opts.PrincipalID == "error5" {
		return nil, errors.New("something wrong")
	}

	return &bPrincipal, nil
}

func TestS3Handler_CreateBucket(t *testing.T) {
	createReq := v1.CreateBucketRequest{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Name:         bBucket.Name,
		AccessPolicy: v1.BucketAccessPolicy_BUCKET_ACCESS_POLICY_READ_WRITE,
		Versioned:    true,
		QuotaBytes:   bBucket.QuotaBytes,
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.CreateBucketRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateBucketResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create bucket",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &v1.CreateBucketResponse{
				Bucket: &bucketResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateBucketRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: "error1",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "nil on invalid conversion",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateBucketRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: "empty1",
				},
			},
			want:      &v1.CreateBucketResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.CreateBucketRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateBucket(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_DeleteBucket(t *testing.T) {
	deleteReqUnforced := v1.DeleteBucketRequest{
		BucketId: bucketResponse.GetId(),
		Force:    false,
	}

	deleteReqForced := v1.DeleteBucketRequest{
		BucketId: bucketResponse.GetId(),
		Force:    true,
	}

	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.DeleteBucketRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.DeleteBucketResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete bucket unforced",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReqUnforced,
			},
			want:      &v1.DeleteBucketResponse{},
			assertion: assert.NoError,
		},
		{
			name: "delete bucket forced",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReqForced,
			},
			want:      &v1.DeleteBucketResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.DeleteBucketRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error2",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReqUnforced,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.DeleteBucketRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReqUnforced,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteBucket(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_GetBucketPolicy(t *testing.T) {
	getReq := v1.GetBucketPolicyRequest{
		BucketId: bucketResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.GetBucketPolicyRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.GetBucketPolicyResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get bucket policy",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want: &v1.GetBucketPolicyResponse{
				BucketId: bucketResponse.GetId(),
				Policy:   v1.BucketAccessPolicy_BUCKET_ACCESS_POLICY_READ_WRITE,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetBucketPolicyRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error3",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no policy",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetBucketPolicyRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "notFound3",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.GetBucketPolicyRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetBucketPolicy(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_ListBuckets(t *testing.T) {
	listReq := v1.ListBucketsRequest{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.ListBucketsRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.ListBucketsResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list buckets",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want: &v1.ListBucketsResponse{
				Buckets: []*v1.Bucket{&bucketResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListBucketsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Filter: &v1.ListBucketsRequest_Filter{
						Names: []string{"error4"},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListBucketsRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListBuckets(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_UpdateBucketPolicy(t *testing.T) {
	updateReq := v1.UpdateBucketPolicyRequest{
		BucketId:     bucketResponse.GetId(),
		AccessPolicy: v1.BucketAccessPolicy_BUCKET_ACCESS_POLICY_READ_WRITE,
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateBucketPolicyRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateBucketPolicyResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update bucket policy",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      &v1.UpdateBucketPolicyResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateBucketPolicyRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error5",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateBucketPolicyRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateBucketPolicy(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_CreateLifecycleRule(t *testing.T) {
	createReq := v1.CreateLifecycleRulesRequest{
		BucketId: bucketResponse.GetId(),
		LifecycleRules: []*v1.LifecycleRule{
			{
				Prefix:               bLifecycleRule.Prefix,
				ExpireDays:           bLifecycleRule.ExpireDays,
				NoncurrentExpireDays: bLifecycleRule.NoncurrentExpireDays,
				DeleteMarker:         bLifecycleRule.DeleteMarker,
			},
		},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.CreateLifecycleRulesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateLifecycleRulesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create lifecycle rule",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &v1.CreateLifecycleRulesResponse{
				LifecycleRules: []*v1.LifecycleRule{&lifecycleRuleResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateLifecycleRulesRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error1",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.CreateLifecycleRulesRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateLifecycleRules(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_DeleteLifecycleRule(t *testing.T) {
	deleteReq := v1.DeleteLifecycleRulesRequest{
		BucketId: bucketResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.DeleteLifecycleRulesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.DeleteLifecycleRulesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete lifecycle rule",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &v1.DeleteLifecycleRulesResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.DeleteLifecycleRulesRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: clusterResponse.GetId(),
						Id:        "error2",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.DeleteLifecycleRulesRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteLifecycleRules(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_ListLifecycleRules(t *testing.T) {
	listReq := v1.ListLifecycleRulesRequest{
		BucketId: bucketResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.ListLifecycleRulesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.ListLifecycleRulesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list lifecycle rules",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want: &v1.ListLifecycleRulesResponse{
				LifecycleRules: []*v1.LifecycleRule{&lifecycleRuleResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListLifecycleRulesRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error3",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListLifecycleRulesRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &listReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListLifecycleRules(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_UpdateLifecycleRule(t *testing.T) {
	updateReq := v1.UpdateLifecycleRulesRequest{
		BucketId:       bucketResponse.GetId(),
		LifecycleRules: []*v1.LifecycleRule{&lifecycleRuleResponse},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateLifecycleRulesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateLifecycleRulesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update lifecycle rule",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want: &v1.UpdateLifecycleRulesResponse{
				LifecycleRules: []*v1.LifecycleRule{&lifecycleRuleResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateLifecycleRulesRequest{
					BucketId: &v1.BucketIdentifier{
						ClusterId: clusterResponse.GetId(),
						Id:        "error4",
					},
					LifecycleRules: []*v1.LifecycleRule{
						{
							Id: &v1.LifecycleRuleIdentifier{
								Id: "id",
							},
							Prefix:     "pre",
							ExpireDays: 50,
						},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateLifecycleRulesRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateLifecycleRules(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_CreateS3Principal(t *testing.T) {
	createReq := v1.CreateS3PrincipalRequest{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Name:        bPrincipal.Name,
		Credentials: "creds",
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.CreateS3PrincipalRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateS3PrincipalResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create principal",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &v1.CreateS3PrincipalResponse{
				S3Principal: &v1.S3Principal{
					Id:       principalResponse.GetId(),
					Name:     principalResponse.GetName(),
					Policies: []*v1.S3Principal_Policy{},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateS3PrincipalRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: "error1",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.CreateS3PrincipalRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateS3Principal(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_DeleteS3Principal(t *testing.T) {
	deleteReq := v1.DeleteS3PrincipalRequest{
		PrincipalId: principalResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.DeleteS3PrincipalRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.DeleteS3PrincipalResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete principal",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &v1.DeleteS3PrincipalResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.DeleteS3PrincipalRequest{
					PrincipalId: &v1.S3PrincipalIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error2",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.DeleteS3PrincipalRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteS3Principal(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_GetS3Principal(t *testing.T) {
	getReq := v1.GetS3PrincipalRequest{
		PrincipalId: principalResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.GetS3PrincipalRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.GetS3PrincipalResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get principal",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want: &v1.GetS3PrincipalResponse{
				S3Principal: &principalResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetS3PrincipalRequest{
					PrincipalId: &v1.S3PrincipalIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error3",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty on invalid conversion",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetS3PrincipalRequest{
					PrincipalId: &v1.S3PrincipalIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "empty3",
					},
				},
			},
			want:      &v1.GetS3PrincipalResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.GetS3PrincipalRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetS3Principal(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_SetS3PrincipalCredentials(t *testing.T) {
	setReq := v1.SetS3PrincipalCredentialsRequest{
		PrincipalId: principalResponse.GetId(),
		Credentials: "creds",
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.SetS3PrincipalCredentialsRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.SetS3PrincipalCredentialsResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "set principal credentials",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &setReq,
			},
			want:      &v1.SetS3PrincipalCredentialsResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.SetS3PrincipalCredentialsRequest{
					PrincipalId: &v1.S3PrincipalIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error4",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &setReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.SetS3PrincipalCredentialsRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &setReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.SetS3PrincipalCredentials(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestS3Handler_UpdateS3PrincipalPolicies(t *testing.T) {
	updateReq := v1.UpdateS3PrincipalPoliciesRequest{
		PrincipalId: principalResponse.GetId(),
		Policies:    principalResponse.GetPolicies(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateS3PrincipalPoliciesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateS3PrincipalPoliciesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update principal policies",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want: &v1.UpdateS3PrincipalPoliciesResponse{
				S3Principal: &principalResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "allow error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateS3PrincipalPoliciesRequest{
					PrincipalId: principalResponse.GetId(),
					Policies: []*v1.S3Principal_Policy{
						{
							BucketId: bucketResponse.GetId(),
							SourceIpFilter: &v1.S3Principal_Policy_SourceIpFilter{
								Allow: []string{
									"127.0.0.0/44",
								},
								Disallow: []string{
									disallowIPString,
								},
							},
						},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "disallow error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateS3PrincipalPoliciesRequest{
					PrincipalId: principalResponse.GetId(),
					Policies: []*v1.S3Principal_Policy{
						{
							BucketId: bucketResponse.GetId(),
							SourceIpFilter: &v1.S3Principal_Policy_SourceIpFilter{
								Allow: []string{
									allowIPString,
								},
								Disallow: []string{
									"127.0.1.20/44",
								},
							},
						},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateS3PrincipalPoliciesRequest{
					PrincipalId: &v1.S3PrincipalIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error5",
					},
					Policies: principalResponse.GetPolicies(),
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateS3PrincipalPoliciesRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoS3Backend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &S3Handler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateS3PrincipalPolicies(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_TestEnumActionsIntoArray(t *testing.T) {
	type args struct {
		value []v1.S3Principal_Policy_BucketActions
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty behavior",
			args: args{value: make([]v1.S3Principal_Policy_BucketActions, 0)},
			want: []string{
				"s3:GetBucketLocation",
				"s3:GetBucketPolicy",
				"s3:ListBucket",
				"s3:ListBucketMultipartUploads",
				"s3:ListMultipartUploadParts",
				"s3:GetBucketTagging",
				"s3:ListBucketVersions",
			},
		},
		{
			name: "single entry",
			args: args{value: []v1.S3Principal_Policy_BucketActions{v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION}},
			want: []string{
				"s3:GetBucketLocation",
			},
		},
		{
			name: "invalid entry",
			args: args{value: []v1.S3Principal_Policy_BucketActions{v1.S3Principal_Policy_BUCKET_ACTIONS_UNSPECIFIED}},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, enumActionsIntoArray(tt.args.value))
		})
	}
}

func TestBackend_TestArrayIntoActionsEnum(t *testing.T) {
	type args struct {
		value []string
	}
	tests := []struct {
		name string
		args args
		want []v1.S3Principal_Policy_BucketActions
	}{
		{
			name: "empty behavior",
			args: args{value: []string{}},
			want: []v1.S3Principal_Policy_BucketActions{},
		},
		{
			name: "multiple variants",
			args: args{value: []string{
				"s3:GetBucketLocation",
				"s3:GetBucketPolicy",
				"s3:ListBucket",
				"s3:ListBucketMultipartUploads",
				"s3:ListMultipartUploadParts",
				"s3:GetBucketTagging",
				"s3:ListBucketVersions",
			}},
			want: []v1.S3Principal_Policy_BucketActions{
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION,
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_MULTIPART_UPLOADS,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_MULTIPART_UPLOAD_PARTS,
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_TAGGING,
				v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_VERSIONS,
			},
		},
		{
			name: "single entry",
			args: args{value: []string{
				"s3:GetBucketPolicy",
			}},
			want: []v1.S3Principal_Policy_BucketActions{
				v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY,
			},
		},
		{
			name: "invalid entry",
			args: args{value: []string{
				"invalid",
			}},
			want: []v1.S3Principal_Policy_BucketActions{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, arrayIntoActionsEnum(tt.args.value))
		})
	}
}
