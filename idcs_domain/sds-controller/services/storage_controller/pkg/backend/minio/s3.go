// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package minio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/rs/zerolog/log"
)

type IAMPolicy struct {
	Version   string         `json:"Version"`
	Statement []IAMStatement `json:"Statement"`
}

type IAMStatement struct {
	Action    []string     `json:"Action"`
	Effect    string       `json:"Effect"`
	Principal IAMPrincipal `json:"Principal"`
	Resource  []string     `json:"Resource"`
	Sid       string       `json:"Sid"`
}

type IAMPrincipal struct {
	AWS []string `json:"AWS"`
}

func (b *Backend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	if b.minioClient.IsOffline() {
		return &backend.ClusterStatus{
			HealthStatus: backend.Unhealthy,
		}, nil
	} else {
		return &backend.ClusterStatus{
			HealthStatus: backend.Healthy,
		}, nil
	}
}

func (b *Backend) CreateBucket(ctx context.Context, opts backend.CreateBucketOpts) (*backend.Bucket, error) {
	err := b.minioClient.MakeBucket(ctx, opts.Name, minio.MakeBucketOptions{})

	if err != nil {
		log.Error().Str("bucketName", opts.Name).Ctx(ctx).Err(err).Msg("Error creating minio bucket")
		return nil, err
	}

	// Buckets are NOT versioned by default
	if opts.Versioned {
		err = b.minioClient.EnableVersioning(ctx, opts.Name)
		if err != nil {
			opts.Versioned = false
			log.Warn().Str("bucketName", opts.Name).Ctx(ctx).Err(err).Msg("Cannot set versioning for bucket")
		}
	}

	err = b.adminClient.SetBucketQuota(ctx, opts.Name, &madmin.BucketQuota{
		Size: opts.QuotaBytes,
		Type: madmin.HardQuota,
	})
	if err != nil {
		log.Error().Str("bucketName", opts.Name).Ctx(ctx).Err(err).Msg("Failed to set bucket quota, removing bucket")
		removeErr := b.minioClient.RemoveBucketWithOptions(ctx, opts.Name, minio.RemoveBucketOptions{})
		if removeErr != nil {
			log.Error().Str("bucketName", opts.Name).Ctx(ctx).Err(err).Msg("Failed to remove bucket after quota fail")
		}
		return nil, err
	}

	err = b.UpdateBucketPolicy(ctx, backend.UpdateBucketPolicyOpts{ID: opts.Name, AccessPolicy: opts.AccessPolicy})
	if err != nil {
		log.Error().Err(err).Msg("Could not update bucket policy during bucket creation")
		// Not return error case as can be related to higher level IAM settings, client can repeat settings of policy later
		opts.AccessPolicy = backend.None
	}

	if b.config.MinioConfig != nil && b.config.MinioConfig.KESKey != "" {
		err = b.minioClient.SetBucketEncryption(ctx, opts.Name, sse.NewConfigurationSSEKMS(b.config.MinioConfig.KESKey))

		if err != nil {
			// Not fatal for bucket creation, but should be monitored and mediated by rotation/batch encrypt job later
			log.Error().Err(err).Msg("Could not set encryption policy for bucket")
		}
	}

	return &backend.Bucket{
		ID:             opts.Name,
		Name:           opts.Name,
		AccessPolicy:   opts.AccessPolicy,
		Versioned:      opts.Versioned,
		QuotaBytes:     opts.QuotaBytes,
		AvailableBytes: opts.QuotaBytes,
	}, nil
}

func (b *Backend) DeleteBucket(ctx context.Context, opts backend.DeleteBucketOpts) error {
	err := b.minioClient.RemoveBucketWithOptions(ctx, opts.ID, minio.RemoveBucketOptions{
		ForceDelete: opts.ForceDelete,
	})
	if err != nil {
		log.Error().Str("bucketName", opts.ID).Ctx(ctx).Err(err).Msg("Failed to remove bucket")
		return err
	}
	return nil
}

func (b *Backend) GetBucketPolicy(ctx context.Context, opts backend.GetBucketPolicyOpts) (*backend.AccessPolicy, error) {
	policy, err := b.minioClient.GetBucketPolicy(ctx, opts.ID)

	if err != nil {
		log.Error().Str("bucketID", opts.ID).Ctx(ctx).Err(err).Msg("Cannot obtain bucket policy")
		return nil, err
	}

	return jsonToPolicy(policy)
}

func (b *Backend) ListBuckets(ctx context.Context, opts backend.ListBucketsOpts) ([]*backend.Bucket, error) {
	buckets, err := b.minioClient.ListBuckets(ctx)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Could not list buckets")
		return nil, err
	}

	result := make([]*backend.Bucket, 0)

	dataInfo, err := b.adminClient.DataUsageInfo(ctx)
	if err != nil {
		log.Error().Err(err).Ctx(ctx).Msg("Could not fetch data usage from MinIO server")
		return nil, err
	}

	for _, bucket := range buckets {
		if len(opts.Names) > 0 && !slices.Contains(opts.Names, bucket.Name) {
			continue
		}
		policy := backend.None
		p, err := b.GetBucketPolicy(ctx, backend.GetBucketPolicyOpts{ID: bucket.Name})
		if err != nil {
			log.Error().Err(err).Ctx(ctx).Msg("Could not obtain access policy during list buckets operation")
		} else if p != nil {
			policy = *p
		}
		quota, err := b.adminClient.GetBucketQuota(ctx, bucket.Name)
		if err != nil {
			log.Error().Err(err).Ctx(ctx).Msg("Could not obtain quota during list buckets operation")
		}
		versioned, err := b.minioClient.GetBucketVersioning(ctx, bucket.Name)
		if err != nil {
			log.Error().Err(err).Ctx(ctx).Msg("Could not obtain versioned bucket status")
		}

		var bucketSize uint64 = 0
		info, exists := dataInfo.BucketsUsage[bucket.Name]
		if exists {
			bucketSize = info.Size
		}

		var available uint64 = 0

		if quota.Size != 0 && bucketSize != 0 {
			available = quota.Size - bucketSize
		}

		result = append(result, &backend.Bucket{
			ID:             bucket.Name,
			Name:           bucket.Name,
			AccessPolicy:   policy,
			QuotaBytes:     quota.Size,
			EndpointURL:    b.minioClient.EndpointURL().String(),
			Versioned:      versioned.Enabled(),
			AvailableBytes: available,
		})
	}

	return result, nil
}

func (b *Backend) UpdateBucketPolicy(ctx context.Context, opts backend.UpdateBucketPolicyOpts) error {
	policyString, err := policyToJson(opts.AccessPolicy, opts.ID)
	if err != nil {
		log.Fatal().Str("bucketName", opts.ID).Ctx(ctx).Err(err).Msg("Could not marshall policy to string")
	}

	return b.minioClient.SetBucketPolicy(ctx, opts.ID, policyString)
}

func (b *Backend) CreateLifecycleRules(ctx context.Context, opts backend.CreateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	return b.setLifecycleRules(ctx, opts.BucketID, opts.LifecycleRules)
}

func (b *Backend) DeleteLifecycleRules(ctx context.Context, opts backend.DeleteLifecycleRulesOpts) error {
	return b.minioClient.SetBucketLifecycle(ctx, opts.BucketID, lifecycle.NewConfiguration())
}

func (b *Backend) ListLifecycleRules(ctx context.Context, opts backend.ListLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	rules, err := b.minioClient.GetBucketLifecycle(ctx, opts.BucketID)
	if err != nil {
		log.Error().Err(err).Str("bucketId", opts.BucketID).Msg("Could not list lifecycle rules")
		return nil, err
	}

	var answer []*backend.LifecycleRule
	for _, rule := range rules.Rules {
		r := &backend.LifecycleRule{
			ID:                   rule.ID,
			Prefix:               rule.Prefix,
			ExpireDays:           uint32(rule.Expiration.Days),
			NoncurrentExpireDays: uint32(rule.NoncurrentVersionExpiration.NoncurrentDays),
			DeleteMarker:         rule.Expiration.DeleteMarker.IsEnabled(),
		}
		answer = append(answer, r)
	}

	return answer, nil
}

func (b *Backend) UpdateLifecycleRules(ctx context.Context, opts backend.UpdateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	return b.setLifecycleRules(ctx, opts.BucketID, opts.LifecycleRules)
}

func (b *Backend) CreateS3Principal(ctx context.Context, opts backend.CreateS3PrincipalOpts) (*backend.S3Principal, error) {
	if opts.Name == b.adminPrincipal {
		log.Error().Bool("security", true).Msg("Attempt to create admin principal")
		return nil, errors.New("principal is incorrect")
	}
	err := b.adminClient.AddUser(ctx, opts.Name, opts.Credentials)
	if err != nil {
		log.Error().Err(err).Str("principalName", opts.Name).Msg("Could not create principal")
		return nil, err
	}
	return &backend.S3Principal{
		ID:   opts.Name,
		Name: opts.Name,
	}, nil
}

func (b *Backend) DeleteS3Principal(ctx context.Context, opts backend.DeleteS3PrincipalOpts) error {
	if opts.PrincipalID == b.adminPrincipal {
		log.Error().Bool("security", true).Msg("Attempt to delete admin principal")
		return errors.New("principal is incorrect")
	}
	resp, err := b.adminClient.DetachPolicy(ctx, madmin.PolicyAssociationReq{
		Policies: []string{opts.PrincipalID},
		User:     opts.PrincipalID,
	})

	if err != nil {
		// Info because not this is not error
		log.Info().Err(err).Msg("Could not detach canned policy, most likely it never existed")
	}
	log.Info().Any("policies", resp).Msg("Detached policies from user")

	err = b.adminClient.RemoveCannedPolicy(ctx, opts.PrincipalID)
	if err != nil {
		// Info because not this is not error
		log.Info().Err(err).Msg("Policy for existing user not created, and was not deleted")
	}
	return b.adminClient.RemoveUser(ctx, opts.PrincipalID)
}

func (b *Backend) GetS3Principal(ctx context.Context, opts backend.GetS3PrincipalOpts) (*backend.S3Principal, error) {
	if opts.PrincipalID == b.adminPrincipal {
		log.Error().Bool("security", true).Msg("Attempt to get admin principal")
		return nil, errors.New("principal is incorrect")
	}

	// Required to test if user is exists in the MinIO, otherwise function will not fail
	_, err := b.adminClient.GetUserInfo(ctx, opts.PrincipalID)
	if err != nil {
		log.Error().Err(err).Msg("Could not get user")
		return nil, err
	}

	var policies []*backend.S3Policy
	policy, err := b.adminClient.InfoCannedPolicyV2(ctx, opts.PrincipalID)

	if err != nil {
		log.Debug().Str("principalId", opts.PrincipalID).Ctx(ctx).Msg("Principal does not have policy attached")
	}

	if err == nil && policy != nil && string(policy.Policy) != "" {
		var p backend.S3IAMPolicy
		err := json.Unmarshal(policy.Policy, &p)
		if err != nil {
			return nil, err
		}

		policies = backend.IntoIAMPolicies(p)
	}

	return &backend.S3Principal{
		ID:       opts.PrincipalID,
		Name:     opts.PrincipalID,
		Policies: policies,
	}, nil
}

func (b *Backend) UpdateS3PrincipalPassword(ctx context.Context, opts backend.UpdateS3PrincipalPasswordOpts) error {
	if opts.PrincipalID == b.adminPrincipal {
		log.Error().Bool("security", true).Msg("Attempt to update admin principal")
		return errors.New("principal is incorrect")
	}

	return b.adminClient.SetUser(ctx, opts.PrincipalID, opts.Credentials, madmin.AccountEnabled)
}

func (b *Backend) UpdateS3PrincipalPolicies(ctx context.Context, opts backend.UpdateS3PrincipalPoliciesOpts) (*backend.S3Principal, error) {
	if opts.PrincipalID == b.adminPrincipal {
		log.Error().Bool("security", true).Msg("Attempt to update admin principal")
		return nil, errors.New("principal is incorrect")
	}

	policyJson, err := json.Marshal(backend.FromIAMPolicy(opts.Policies))
	if err != nil {
		return nil, err
	}

	err = b.adminClient.AddCannedPolicy(ctx, opts.PrincipalID, policyJson)
	if err != nil {
		log.Error().Err(err).Msg("Could not add new canned policy")
		return nil, err
	}

	policies, err := b.adminClient.GetPolicyEntities(ctx, madmin.PolicyEntitiesQuery{
		Users:  []string{opts.PrincipalID},
		Policy: []string{opts.PrincipalID},
	})

	if err != nil {
		log.Info().Err(err).Msg("Could not get principal policies info")
		return nil, err
	}

	if len(policies.PolicyMappings) == 0 {
		resp, err := b.adminClient.AttachPolicy(ctx, madmin.PolicyAssociationReq{
			Policies: []string{opts.PrincipalID},
			User:     opts.PrincipalID,
		})

		if err != nil {
			log.Error().Err(err).Msg("Could not attach canned policy")
			return nil, err
		}
		log.Info().Any("policies", resp).Msg("Attached policies to user")
	}

	return &backend.S3Principal{
		ID:       opts.PrincipalID,
		Name:     opts.PrincipalID,
		Policies: opts.Policies,
	}, nil
}

func (b *Backend) setLifecycleRules(ctx context.Context, bucketId string, lifecycleRules []backend.LifecycleRule) ([]*backend.LifecycleRule, error) {
	rules := make([]lifecycle.Rule, 0)
	answer := make([]*backend.LifecycleRule, 0)

	for _, rule := range lifecycleRules {
		if rule.ID == "" {
			uid, err := uuid.NewUUID()
			if err != nil {
				return nil, err
			}
			rule.ID = uid.String()
		}

		rules = append(rules, lifecycle.Rule{
			ID: rule.ID,
			RuleFilter: lifecycle.Filter{
				Prefix: rule.Prefix,
			},
			NoncurrentVersionExpiration: lifecycle.NoncurrentVersionExpiration{
				NoncurrentDays: lifecycle.ExpirationDays(rule.ExpireDays),
			},
			Expiration: lifecycle.Expiration{
				Days:         lifecycle.ExpirationDays(rule.ExpireDays),
				DeleteMarker: lifecycle.ExpireDeleteMarker(rule.DeleteMarker),
			},
			Status: "Enabled",
		})
		answer = append(answer, &rule)
	}

	err := b.minioClient.SetBucketLifecycle(ctx, bucketId, &lifecycle.Configuration{
		Rules: rules,
	})

	if err != nil {
		log.Error().Str("bucketName", bucketId).Ctx(ctx).Err(err).Msg("Failed to create lifecycle rule")
		return nil, err
	}

	return answer, nil
}

func policyToJson(policy backend.AccessPolicy, bucketName string) (string, error) {
	var iamPolicy *IAMPolicy

	switch policy {
	case backend.Read:
		iamPolicy = &IAMPolicy{
			Version: backend.S3_POLICY_VERSION,
			Statement: []IAMStatement{
				{
					Action: []string{"s3:GetObject"},
					Effect: "Allow",
					Principal: IAMPrincipal{
						AWS: []string{"*"},
					},
					Resource: []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)},
					Sid:      "ReadOnly",
				},
			},
		}
	case backend.ReadWrite:
		iamPolicy = &IAMPolicy{
			Version: backend.S3_POLICY_VERSION,
			Statement: []IAMStatement{
				{
					Action: []string{
						"s3:GetObject",
						"s3:WriteObject",
					},
					Effect: "Allow",
					Principal: IAMPrincipal{
						AWS: []string{"*"},
					},
					Resource: []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)},
					Sid:      "ReadWrite",
				},
			},
		}
	default:
		return "", nil
	}

	policyBytes, err := json.Marshal(iamPolicy)
	if err != nil {
		return "", err
	}

	return string(policyBytes), nil
}

func jsonToPolicy(policy string) (*backend.AccessPolicy, error) {
	var iamPolicy IAMPolicy

	err := json.Unmarshal([]byte(policy), &iamPolicy)
	if err != nil {
		return nil, err
	}

	accessPolicy := backend.None

	if len(iamPolicy.Statement) == 1 {
		switch iamPolicy.Statement[0].Sid {
		case "ReadWrite":
			accessPolicy = backend.ReadWrite
		case "ReadOnly":
			accessPolicy = backend.Read
		}
	}

	return &accessPolicy, nil
}
