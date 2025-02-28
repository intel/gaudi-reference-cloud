// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package minio

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/stretchr/testify/assert"
)

func newMockBackend(opts mockBackendOpts) Backend {
	var cluster conf.Cluster
	if opts.config != nil {
		cluster = *opts.config
	} else {
		cluster = conf.Cluster{
			Name:     "test",
			UUID:     "00000000-0000-0000-0000-000000000000",
			Type:     conf.MinIO,
			Location: "testing",
			API: &conf.API{
				Type: conf.REST,
				URL:  "localhost",
			},
			MinioConfig: &conf.MinioConfig{
				KESKey: "key-name",
			},
		}
	}

	return Backend{
		config:         &cluster,
		minioClient:    &MockMinioClient{},
		adminClient:    &MockMinioAdminClient{},
		adminPrincipal: "admin",
	}
}

type MockMinioClient struct {
}

type MockMinioAdminClient struct {
}

func (*MockMinioClient) EndpointURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "localhost",
	}
}

func (*MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) (err error) {
	if ctx.Value("testing.MakeBucket.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioClient) EnableVersioning(ctx context.Context, bucketName string) error {
	if ctx.Value("testing.EnableVersioning.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioClient) HealthCheck(duration time.Duration) (context.CancelFunc, error) {
  _, cancelFn := context.WithCancel(context.Background())
	return cancelFn, nil
}


func (*MockMinioClient) IsOffline() (bool) {
	return false
}

func (*MockMinioClient) RemoveBucketWithOptions(ctx context.Context, bucketName string, opts minio.RemoveBucketOptions) error {
	if ctx.Value("testing.RemoveBucket.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioClient) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	if ctx.Value("testing.GetBucketPolicy.error") != nil {
		return "", errors.New("Something wrong")
	}
	return policyToJson(backend.ReadWrite, "bucket")
}

func (*MockMinioClient) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	answer := make([]minio.BucketInfo, 0)
	if ctx.Value("testing.ListBuckets.error") != nil {
		return answer, errors.New("Something wrong")
	}
	answer = append(answer, minio.BucketInfo{
		Name: "bucket",
	})
	return answer, nil
}

func (*MockMinioClient) GetBucketVersioning(ctx context.Context, bucketName string) (minio.BucketVersioningConfiguration, error) {
	if ctx.Value("testing.GetBucketVersioning.error") != nil {
		return minio.BucketVersioningConfiguration{}, errors.New("Something wrong")
	}
	return minio.BucketVersioningConfiguration{
		Status: "Enabled",
	}, nil
}

func (*MockMinioClient) GetBucketLifecycle(ctx context.Context, bucketName string) (*lifecycle.Configuration, error) {
	if ctx.Value("testing.GetBucketLifecycle.error") != nil {
		return lifecycle.NewConfiguration(), errors.New("Something wrong")
	}
	return &lifecycle.Configuration{
		Rules: []lifecycle.Rule{
			{
				ID:     "lrId",
				Prefix: "pre",
				NoncurrentVersionExpiration: lifecycle.NoncurrentVersionExpiration{
					NoncurrentDays: lifecycle.ExpirationDays(60),
				},
				Expiration: lifecycle.Expiration{
					Days:         lifecycle.ExpirationDays(60),
					DeleteMarker: lifecycle.ExpireDeleteMarker(true),
				},
			},
		},
	}, nil
}

func (*MockMinioClient) SetBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error {
	if ctx.Value("testing.SetBucketLifecycle.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioClient) SetBucketPolicy(ctx context.Context, bucketName, policy string) error {
	if ctx.Value("testing.SetBucketPolicy.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioClient) SetBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error {
	if ctx.Value("testing.SetBucketEncryption.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) DataUsageInfo(ctx context.Context) (madmin.DataUsageInfo, error) {
	if ctx.Value("testing.DataUsageInfo.error") != nil {
		return madmin.DataUsageInfo{}, errors.New("Something wrong")
	}
	bucketUsage := make(map[string]madmin.BucketUsageInfo)
	bucketUsage["bucket"] = madmin.BucketUsageInfo{
		Size: 3000000,
	}
	return madmin.DataUsageInfo{
		BucketsUsage: bucketUsage,
	}, nil
}

func (*MockMinioAdminClient) SetBucketQuota(ctx context.Context, bucket string, quota *madmin.BucketQuota) error {
	if ctx.Value("testing.SetBucketQuota.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) GetBucketQuota(ctx context.Context, bucket string) (q madmin.BucketQuota, err error) {
	if ctx.Value("testing.GetBucketQuota.error") != nil {
		return madmin.BucketQuota{}, errors.New("Something wrong")
	}
	return madmin.BucketQuota{
		Size: 5000000,
	}, nil
}

func (*MockMinioAdminClient) AddUser(ctx context.Context, accessKey, secretKey string) error {
	if ctx.Value("testing.AddUser.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) RemoveCannedPolicy(ctx context.Context, policyName string) error {
	if ctx.Value("testing.RemoveCannedPolicy.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) RemoveUser(ctx context.Context, accessKey string) error {
	if ctx.Value("testing.RemoveUser.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) InfoCannedPolicyV2(ctx context.Context, policyName string) (*madmin.PolicyInfo, error) {
	if ctx.Value("testing.InfoCannedPolicyV2.error") != nil {
		return nil, errors.New("Something wrong")
	}
	if ctx.Value("testing.InfoCannedPolicyV2.wrong") != nil {
		return &madmin.PolicyInfo{
			Policy: []byte("ddd"),
		}, nil
	}
	version := backend.S3_POLICY_VERSION
	allow := "Allow"
	sidRead := "read"
	sidWrite := "write"
	sidDelete := "delete"

	policy := &backend.S3IAMPolicy{
		Version: &version,
		Statement: &[]backend.S3IAMStatement{
			{
				Effect: &allow,
				Sid:    &sidRead,
				Resource: &[]string{
					"arn:aws:s3:::bucket/pre*",
				},
				Action: &[]string{
					"s3:GetObject",
				},
			},
			{
				Effect: &allow,
				Sid:    &sidWrite,
				Resource: &[]string{
					"arn:aws:s3:::bucket/pre*",
				},
				Action: &[]string{
					"s3:WriteObject",
				},
			},
			{
				Effect: &allow,
				Sid:    &sidDelete,
				Resource: &[]string{
					"arn:aws:s3:::bucket/pre*",
				},
				Action: &[]string{
					"s3:DeleteObject",
				},
			},
		},
	}

	bytes, _ := json.Marshal(policy)
	return &madmin.PolicyInfo{
		Policy: bytes,
	}, nil
}

func (*MockMinioAdminClient) SetUser(ctx context.Context, accessKey, secretKey string, status madmin.AccountStatus) error {
	if ctx.Value("testing.SetUser.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) AddCannedPolicy(ctx context.Context, policyName string, policy []byte) error {
	if ctx.Value("testing.AddCannedPolicy.error") != nil {
		return errors.New("Something wrong")
	}
	return nil
}

func (*MockMinioAdminClient) AttachPolicy(ctx context.Context, r madmin.PolicyAssociationReq) (madmin.PolicyAssociationResp, error) {
	if ctx.Value("testing.AttachPolicy.error") != nil {
		return madmin.PolicyAssociationResp{}, errors.New("Something wrong")
	}
	return madmin.PolicyAssociationResp{}, nil
}

func (*MockMinioAdminClient) DetachPolicy(ctx context.Context, r madmin.PolicyAssociationReq) (madmin.PolicyAssociationResp, error) {
	if ctx.Value("testing.DetachPolicy.error") != nil {
		return madmin.PolicyAssociationResp{}, errors.New("Something wrong")
	}
	return madmin.PolicyAssociationResp{}, nil
}

func (*MockMinioAdminClient) GetUserInfo(ctx context.Context, name string) (u madmin.UserInfo, err error) {
	if ctx.Value("testing.GetUserInfo.error") != nil {
		return madmin.UserInfo{}, errors.New("Something wrong")
	}
	return madmin.UserInfo{}, nil
}

func (*MockMinioAdminClient) GetPolicyEntities(ctx context.Context, q madmin.PolicyEntitiesQuery) (r madmin.PolicyEntitiesResult, err error) {
	if ctx.Value("testing.GetPolicyEntities.error") != nil {
		return madmin.PolicyEntitiesResult{}, errors.New("Something wrong")
	}
	if ctx.Value("testing.GetPolicyEntities.found") != nil {
		return madmin.PolicyEntitiesResult{PolicyMappings: []madmin.PolicyEntities{
			{
				Policy: "policy",
			}},
		}, nil
	}
	return madmin.PolicyEntitiesResult{}, nil
}

func TestBackend_GetStatus(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.ClusterStatus
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:  "get status",
			bOpts: mockBackendOpts{},
			args: args{
				ctx: context.Background(),
			},
			want: &backend.ClusterStatus{
				HealthStatus: backend.Healthy,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetStatus(tt.args.ctx)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_CreateBucket(t *testing.T) {
	opts := backend.CreateBucketOpts{
		Name:         "bucket",
		AccessPolicy: backend.ReadWrite,
		Versioned:    true,
		QuotaBytes:   5000000,
	}
	type args struct {
		ctx  context.Context
		opts backend.CreateBucketOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      *backend.Bucket
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create bucket",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: &backend.Bucket{
				ID:             "bucket",
				Name:           "bucket",
				AccessPolicy:   backend.ReadWrite,
				Versioned:      true,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name: "create unversioned bucket",
			args: args{
				ctx: context.Background(),
				opts: backend.CreateBucketOpts{
					Name:         "bucket",
					AccessPolicy: backend.ReadWrite,
					Versioned:    false,
					QuotaBytes:   5000000,
				},
			},
			want: &backend.Bucket{
				ID:             "bucket",
				Name:           "bucket",
				AccessPolicy:   backend.ReadWrite,
				Versioned:      false,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.MakeBucket.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "versioned api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.EnableVersioning.error", true),
				opts: opts,
			},
			want: &backend.Bucket{
				ID:             "bucket",
				Name:           "bucket",
				AccessPolicy:   backend.ReadWrite,
				Versioned:      false,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name: "quota api error",
			args: args{
				ctx:  context.WithValue(context.WithValue(context.Background(), "testing.RemoveBucket.error", true), "testing.SetBucketQuota.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "policy api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketPolicy.error", true),
				opts: opts,
			},
			want: &backend.Bucket{
				ID:             "bucket",
				Name:           "bucket",
				AccessPolicy:   backend.None,
				Versioned:      true,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name: "bucket encryption api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketEncryption.error", true),
				opts: opts,
			},
			want: &backend.Bucket{
				ID:             "bucket",
				Name:           "bucket",
				AccessPolicy:   backend.ReadWrite,
				Versioned:      true,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateBucket(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteBucket(t *testing.T) {
	opts := backend.DeleteBucketOpts{
		ID: "id",
	}
	type args struct {
		ctx  context.Context
		opts backend.DeleteBucketOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete bucket",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.RemoveBucket.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteBucket(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetBucketPolicy(t *testing.T) {
	opts := backend.GetBucketPolicyOpts{
		ID: "id",
	}
	public := backend.ReadWrite

	type args struct {
		ctx  context.Context
		opts backend.GetBucketPolicyOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      *backend.AccessPolicy
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get bucket policy",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &public,
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetBucketPolicy.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetBucketPolicy(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListBuckets(t *testing.T) {
	getBucket := backend.Bucket{
		ID:             "bucket",
		Name:           "bucket",
		Versioned:      true,
		QuotaBytes:     5000000,
		AvailableBytes: 2000000,
		AccessPolicy:   backend.ReadWrite,
		EndpointURL:    "https://localhost",
	}

	type args struct {
		ctx  context.Context
		opts backend.ListBucketsOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      []*backend.Bucket
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list buckets",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListBucketsOpts{},
			},
			want:      []*backend.Bucket{&getBucket},
			assertion: assert.NoError,
		},
		{
			name: "list bucket names",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListBucketsOpts{Names: []string{"bucket"}},
			},
			want:      []*backend.Bucket{&getBucket},
			assertion: assert.NoError,
		},
		{
			name: "list bucket names not found",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListBucketsOpts{Names: []string{"bucket1"}},
			},
			want:      []*backend.Bucket{},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.ListBuckets.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "data error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DataUsageInfo.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "policy api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetBucketPolicy.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want: []*backend.Bucket{
				{
					ID:             "bucket",
					Name:           "bucket",
					Versioned:      true,
					QuotaBytes:     5000000,
					AvailableBytes: 2000000,
					AccessPolicy:   backend.None,
					EndpointURL:    "https://localhost",
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "quota api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetBucketQuota.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want: []*backend.Bucket{
				{
					ID:             "bucket",
					Name:           "bucket",
					Versioned:      true,
					QuotaBytes:     0,
					AvailableBytes: 0,
					AccessPolicy:   backend.ReadWrite,
					EndpointURL:    "https://localhost",
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "versioning api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetBucketVersioning.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want: []*backend.Bucket{
				{
					ID:             "bucket",
					Name:           "bucket",
					Versioned:      false,
					QuotaBytes:     5000000,
					AvailableBytes: 2000000,
					AccessPolicy:   backend.ReadWrite,
					EndpointURL:    "https://localhost",
				},
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListBuckets(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateBucketPolicy(t *testing.T) {
	opts := backend.UpdateBucketPolicyOpts{
		ID:           "bucket",
		AccessPolicy: backend.Read,
	}

	type args struct {
		ctx  context.Context
		opts backend.UpdateBucketPolicyOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update bucket policy",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketPolicy.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.UpdateBucketPolicy(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_CreateLifecycleRule(t *testing.T) {
	opts := backend.CreateLifecycleRulesOpts{
		BucketID: "bucket",
		LifecycleRules: []backend.LifecycleRule{
			{
				ID:                   "lrId",
				Prefix:               "pre",
				ExpireDays:           60,
				NoncurrentExpireDays: 60,
				DeleteMarker:         true,
			},
		},
	}

	type args struct {
		ctx  context.Context
		opts backend.CreateLifecycleRulesOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      []*backend.LifecycleRule
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create lifecycle rule",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: []*backend.LifecycleRule{
				{
					ID:                   "lrId",
					Prefix:               "pre",
					ExpireDays:           60,
					NoncurrentExpireDays: 60,
					DeleteMarker:         true,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "create lifecycle rule with uuid",
			args: args{
				ctx: context.Background(),
				opts: backend.CreateLifecycleRulesOpts{
					BucketID: "bucket",
					LifecycleRules: []backend.LifecycleRule{
						{
							Prefix:               "pre",
							ExpireDays:           60,
							NoncurrentExpireDays: 60,
							DeleteMarker:         true,
						},
					},
				},
			},
			want: []*backend.LifecycleRule{
				{
					Prefix:               "pre",
					ExpireDays:           60,
					NoncurrentExpireDays: 60,
					DeleteMarker:         true,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketLifecycle.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateLifecycleRules(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			if tt.args.opts.LifecycleRules[0].ID == "" {
				assert.Equal(t, 36, len(got[0].ID))
				got[0].ID = ""
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteLifecycleRule(t *testing.T) {
	opts := backend.DeleteLifecycleRulesOpts{
		BucketID: "bucket",
	}

	type args struct {
		ctx  context.Context
		opts backend.DeleteLifecycleRulesOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete lifecycle rule",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketLifecycle.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteLifecycleRules(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_ListLifecycleRules(t *testing.T) {
	opts := backend.ListLifecycleRulesOpts{
		BucketID: "bucket",
	}

	type args struct {
		ctx  context.Context
		opts backend.ListLifecycleRulesOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      []*backend.LifecycleRule
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list buckets",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: []*backend.LifecycleRule{
				{
					ID:                   "lrId",
					Prefix:               "pre",
					ExpireDays:           60,
					NoncurrentExpireDays: 60,
					DeleteMarker:         true,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetBucketLifecycle.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListLifecycleRules(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateLifecycleRule(t *testing.T) {
	opts := backend.UpdateLifecycleRulesOpts{
		BucketID: "bucket",
		LifecycleRules: []backend.LifecycleRule{
			{
				ID:                   "lrId",
				Prefix:               "pre",
				ExpireDays:           60,
				NoncurrentExpireDays: 60,
				DeleteMarker:         true,
			},
		},
	}

	type args struct {
		ctx  context.Context
		opts backend.UpdateLifecycleRulesOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      []*backend.LifecycleRule
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update lifecycle rule",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: []*backend.LifecycleRule{
				{
					ID:                   "lrId",
					Prefix:               "pre",
					ExpireDays:           60,
					NoncurrentExpireDays: 60,
					DeleteMarker:         true,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetBucketLifecycle.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.UpdateLifecycleRules(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_CreateS3Principal(t *testing.T) {
	opts := backend.CreateS3PrincipalOpts{
		Name:        "userId",
		Credentials: "password",
	}

	type args struct {
		ctx  context.Context
		opts backend.CreateS3PrincipalOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      *backend.S3Principal
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create s3 user",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: &backend.S3Principal{
				ID:   "userId",
				Name: "userId",
			},
			assertion: assert.NoError,
		},
		{
			name: "admin principal error",
			args: args{
				ctx: context.Background(),
				opts: backend.CreateS3PrincipalOpts{
					Name:        "admin",
					Credentials: "password",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AddUser.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateS3Principal(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteS3Principal(t *testing.T) {
	opts := backend.DeleteS3PrincipalOpts{
		PrincipalID: "userId",
	}

	type args struct {
		ctx  context.Context
		opts backend.DeleteS3PrincipalOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete s3 user",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "admin principal error",
			args: args{
				ctx: context.Background(),
				opts: backend.DeleteS3PrincipalOpts{
					PrincipalID: "admin",
				},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.RemoveUser.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "detach policy error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DetachPolicy.error", true),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "remove canned api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.RemoveCannedPolicy.error", true),
				opts: opts,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteS3Principal(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetS3Principal(t *testing.T) {
	opts := backend.GetS3PrincipalOpts{
		PrincipalID: "userId",
	}

	type args struct {
		ctx  context.Context
		opts backend.GetS3PrincipalOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      *backend.S3Principal
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get s3 user",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: &backend.S3Principal{
				ID:   "userId",
				Name: "userId",
				Policies: []*backend.S3Policy{
					{
						BucketID: "bucket",
						Prefix:   "pre",
						Read:     true,
						Write:    true,
						Delete:   true,
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "admin principal error",
			args: args{
				ctx: context.Background(),
				opts: backend.GetS3PrincipalOpts{
					PrincipalID: "admin",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUserInfo.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "policy api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.InfoCannedPolicyV2.error", true),
				opts: opts,
			},
			want: &backend.S3Principal{
				ID:   "userId",
				Name: "userId",
			},
			assertion: assert.NoError,
		},
		{
			name: "policy api error return type",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.InfoCannedPolicyV2.wrong", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetS3Principal(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateS3PrincipalPassword(t *testing.T) {
	opts := backend.UpdateS3PrincipalPasswordOpts{
		PrincipalID: "userId",
		Credentials: "creds",
	}

	type args struct {
		ctx  context.Context
		opts backend.UpdateS3PrincipalPasswordOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "set s3 user password",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "admin principal error",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateS3PrincipalPasswordOpts{
					PrincipalID: "admin",
					Credentials: "creds",
				},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetUser.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.UpdateS3PrincipalPassword(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_UpdateS3PrincipalPolicies(t *testing.T) {
	opts := backend.UpdateS3PrincipalPoliciesOpts{
		PrincipalID: "userId",
		Policies: []*backend.S3Policy{
			{
				BucketID: "bucket",
				Read:     true,
				Write:    true,
				Delete:   true,
			},
		},
	}

	type args struct {
		ctx  context.Context
		opts backend.UpdateS3PrincipalPoliciesOpts
	}
	tests := []struct {
		name      string
		bOpts     mockBackendOpts
		args      args
		want      *backend.S3Principal
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update s3 user policies",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: &backend.S3Principal{
				ID:   "userId",
				Name: "userId",
				Policies: []*backend.S3Policy{
					{
						BucketID: "bucket",
						Read:     true,
						Write:    true,
						Delete:   true,
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "update s3 user found policy",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetPolicyEntities.found", true),
				opts: opts,
			},
			want: &backend.S3Principal{
				ID:   "userId",
				Name: "userId",
				Policies: []*backend.S3Policy{
					{
						BucketID: "bucket",
						Read:     true,
						Write:    true,
						Delete:   true,
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "admin principal error",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateS3PrincipalPoliciesOpts{
					PrincipalID: "admin",
				},
			},
			assertion: assert.Error,
		},
		{
			name: "get policy api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetPolicyEntities.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AddCannedPolicy.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "attach api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AttachPolicy.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.UpdateS3PrincipalPolicies(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_MarshallUnmarshallPolicyIAM(t *testing.T) {
	tests := []struct {
		name string
		args backend.AccessPolicy
	}{
		{name: "read", args: backend.Read},
		{name: "public", args: backend.ReadWrite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyJson, err := policyToJson(tt.args, "bucket")
			assert.NoError(t, err)
			policy, err := jsonToPolicy(policyJson)
			assert.NoError(t, err)
			assert.Equal(t, tt.args, *policy)
		})
	}
}

func TestBackend_JsonToPolicy(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		want      *backend.AccessPolicy
		assertion assert.ErrorAssertionFunc
	}{
		{name: "none", args: "", want: nil, assertion: assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := jsonToPolicy(tt.args)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, policy)
		})
	}
}

func TestBackend_PolicyToJson(t *testing.T) {
	tests := []struct {
		name      string
		args      backend.AccessPolicy
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{name: "none", args: backend.None, want: "", assertion: assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyJson, err := policyToJson(tt.args, "bucket")
			tt.assertion(t, err)
			assert.Equal(t, tt.want, policyJson)
		})
	}
}
