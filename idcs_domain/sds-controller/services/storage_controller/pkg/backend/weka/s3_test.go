// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"net"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

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
				Versioned:      false,
				QuotaBytes:     5000000,
				AvailableBytes: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateS3BucketWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateS3BucketWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DestroyS3BucketWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DestroyS3BucketWithResponse.empty", true),
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketPolicyWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketPolicyWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketPolicyWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.NoError,
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
		ID:             "bucketId",
		Name:           "bucketId",
		Versioned:      false,
		QuotaBytes:     5000000,
		AvailableBytes: 2000000,
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
				opts: backend.ListBucketsOpts{Names: []string{"bucketId"}},
			},
			want:      []*backend.Bucket{&getBucket},
			assertion: assert.NoError,
		},
		{
			name: "list bucket names not found",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListBucketsOpts{Names: []string{"bucketId1"}},
			},
			want:      []*backend.Bucket{},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: backend.ListBucketsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketsWithResponse.error", true),
				opts: backend.ListBucketsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketsWithResponse.empty", true),
				opts: backend.ListBucketsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "list bucket no name",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetS3BucketsWithResponse.noName", true),
				opts: backend.ListBucketsOpts{},
			},
			want:      []*backend.Bucket{},
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetS3BucketPolicyWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetS3BucketPolicyWithResponse.empty", true),
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
					ID:         "lfId",
					Prefix:     "pre",
					ExpireDays: 60,
				},
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3CreateLifecycleRuleWithResponse.error", true),
				opts: opts,
			},
			want:      []*backend.LifecycleRule{},
			assertion: assert.NoError,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3CreateLifecycleRuleWithResponse.empty", true),
				opts: opts,
			},
			want:      []*backend.LifecycleRule{},
			assertion: assert.NoError,
		},
		{
			name: "api error delete",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3DeleteLifecycleRuleWithResponse.error", true),
				opts: opts,
			},
			want: []*backend.LifecycleRule{
				{
					ID:         "lfId",
					Prefix:     "pre",
					ExpireDays: 60,
				},
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateLifecycleRules(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3DeleteLifecycleRuleWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3DeleteLifecycleRuleWithResponse.empty", true),
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
					ID:         "lfId",
					Prefix:     "pre",
					ExpireDays: 60,
				},
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3ListAllLifecycleRulesWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3ListAllLifecycleRulesWithResponse.empty", true),
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
					ID:         "lfId",
					Prefix:     "pre",
					ExpireDays: 60,
				},
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error delete",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3DeleteLifecycleRuleWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response delete",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3DeleteLifecycleRuleWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error create",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3CreateLifecycleRuleWithResponse.error", true),
				opts: opts,
			},
			want:      []*backend.LifecycleRule{},
			assertion: assert.NoError,
		},
		{
			name: "empty response create",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.S3CreateLifecycleRuleWithResponse.empty", true),
				opts: opts,
			},
			want:      []*backend.LifecycleRule{},
			assertion: assert.NoError,
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
		Name:        "username",
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
				Name: "username",
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateUserWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateUserWithResponse.empty", true),
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteUserWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "ignore policy delete error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteS3PolicyWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteUserWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
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
				Name: "username",
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.incomplete", true),
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
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetUserPasswordWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetUserPasswordWithResponse.empty", true),
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
	ipNetAllow := "127.0.0.0/28"
	_, netAllow, _ := net.ParseCIDR(ipNetAllow)

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
				Name: "username",
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
			name: "fail on ip filter",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateS3PrincipalPoliciesOpts{
					PrincipalID: "userId",
					Policies: []*backend.S3Policy{
						{
							BucketID: "bucket",
							Read:     true,
							Write:    true,
							Delete:   true,
							AllowSourceNets: []*net.IPNet{
								netAllow,
							},
						},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateS3PolicyWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateS3PolicyWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "get principal error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "get principal empty",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "attach error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AttachS3PolicyWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "attach empty",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AttachS3PolicyWithResponse.empty", true),
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

func TestBackend_TestIntoPolicyName(t *testing.T) {
	type args struct {
		value backend.AccessPolicy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "read", args: args{value: backend.Read}, want: "download"},
		{name: "public", args: args{value: backend.ReadWrite}, want: "public"},
		{name: "none", args: args{value: backend.None}, want: "none"},
		{name: "empty", args: args{}, want: "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, intoPolicyName(&tt.args.value))
		})
	}
}

func TestBackend_TestFromPolicyName(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want backend.AccessPolicy
	}{
		{name: "read", args: args{value: "download"}, want: backend.Read},
		{name: "public", args: args{value: "upload"}, want: backend.None},
		{name: "none", args: args{value: "none"}, want: backend.None},
		{name: "none", args: args{value: "public"}, want: backend.ReadWrite},
		{name: "empty", args: args{}, want: backend.None},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fromPolicyName(tt.args.value))
		})
	}
}
