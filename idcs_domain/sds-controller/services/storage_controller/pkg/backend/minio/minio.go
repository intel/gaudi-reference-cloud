// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package minio

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/sse"
)

type Backend struct {
	config         *conf.Cluster
	adminClient    MinioAdminClient
	minioClient    MinioClient
	adminPrincipal string
}

type MinioClient interface {
	EndpointURL() *url.URL
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) (err error)
	EnableVersioning(ctx context.Context, bucketName string) error
	RemoveBucketWithOptions(ctx context.Context, bucketName string, opts minio.RemoveBucketOptions) error
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	GetBucketVersioning(ctx context.Context, bucketName string) (minio.BucketVersioningConfiguration, error)
	GetBucketLifecycle(ctx context.Context, bucketName string) (*lifecycle.Configuration, error)
	SetBucketLifecycle(ctx context.Context, bucketName string, config *lifecycle.Configuration) error
	SetBucketPolicy(ctx context.Context, bucketName, policy string) error
	SetBucketEncryption(ctx context.Context, bucketName string, config *sse.Configuration) error
	HealthCheck(hcDuration time.Duration) (context.CancelFunc, error)
	IsOffline() bool
}

type MinioAdminClient interface {
	DataUsageInfo(ctx context.Context) (madmin.DataUsageInfo, error)
	SetBucketQuota(ctx context.Context, bucket string, quota *madmin.BucketQuota) error
	GetBucketQuota(ctx context.Context, bucket string) (q madmin.BucketQuota, err error)
	AddUser(ctx context.Context, accessKey, secretKey string) error
	RemoveCannedPolicy(ctx context.Context, policyName string) error
	RemoveUser(ctx context.Context, accessKey string) error
	InfoCannedPolicyV2(ctx context.Context, policyName string) (*madmin.PolicyInfo, error)
	SetUser(ctx context.Context, accessKey, secretKey string, status madmin.AccountStatus) error
	AddCannedPolicy(ctx context.Context, policyName string, policy []byte) error
	AttachPolicy(ctx context.Context, r madmin.PolicyAssociationReq) (madmin.PolicyAssociationResp, error)
	DetachPolicy(ctx context.Context, r madmin.PolicyAssociationReq) (madmin.PolicyAssociationResp, error)
	GetUserInfo(ctx context.Context, name string) (u madmin.UserInfo, err error)
	GetPolicyEntities(ctx context.Context, q madmin.PolicyEntitiesQuery) (r madmin.PolicyEntitiesResult, err error)
}

func NewBackend(config *conf.Cluster) (*Backend, error) {
	if config.Auth == nil {
		return nil, fmt.Errorf("auth field in the config cannot be nil for cluster: %s", config.UUID)
	}

	if config.Auth.Scheme != conf.Basic {
		return nil, fmt.Errorf("minio support only basic credentials, provided: %v", config.Auth.Scheme)
	}

	creds, err := conf.ReadCredentials(*config.Auth)
	if err != nil {
		return nil, err
	}

	// MinIO disables compression in transport
	gzip := false

	tr, err := backend.CreateHTTPTransport(config, gzip)
	if err != nil {
		return nil, err
	}

	adminClient, err := madmin.New(config.API.URL, creds.Principal, creds.Credentials, true)

	if err != nil {
		return nil, fmt.Errorf("could not create MinIO admin client: %v", err)
	}

	minioClient, err := minio.New(config.API.URL, &minio.Options{
		Creds:     credentials.NewStaticV4(creds.Principal, creds.Credentials, ""),
		Transport: tr,
		Secure:    true,
	})

	if err != nil {
		return nil, fmt.Errorf("could not create MinIO client: %v", err)
	}

	_, err = minioClient.HealthCheck(10 * time.Second)

	if err!=nil {
		return nil, fmt.Errorf("could not start MinIO healthcheck: %v", err)
	}
	adminClient.SetCustomTransport(tr)

	return &Backend{
		config:         config,
		adminClient:    adminClient,
		minioClient:    minioClient,
		adminPrincipal: creds.Principal,
	}, nil
}
