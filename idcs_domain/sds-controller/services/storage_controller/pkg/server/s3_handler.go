// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type S3Handler struct {
	Backends map[string]backend.Interface
}

// CreateBucket implements v1.S3ServiceServer.
func (h *S3Handler) CreateBucket(ctx context.Context, r *v1.CreateBucketRequest) (*v1.CreateBucketResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("bucket", r.GetName()).Msg("Creating bucket")

	bucket, err := s.CreateBucket(ctx, backend.CreateBucketOpts{
		Name:         r.GetName(),
		AccessPolicy: backend.AccessPolicy(r.GetAccessPolicy()),
		Versioned:    r.GetVersioned(),
		QuotaBytes:   r.GetQuotaBytes(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("bucket", r.GetName()).Msg("Created bucket")

	return &v1.CreateBucketResponse{
		Bucket: intoBucket(r.GetClusterId(), bucket),
	}, nil
}

// DeleteBucket implements v1.S3ServiceServer.
func (h *S3Handler) DeleteBucket(ctx context.Context, r *v1.DeleteBucketRequest) (*v1.DeleteBucketResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Bool("Forced", r.GetForce()).Msg("Deleting bucket")

	err := s.DeleteBucket(ctx, backend.DeleteBucketOpts{
		ID:          r.GetBucketId().GetId(),
		ForceDelete: r.GetForce(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Bool("Forced", r.GetForce()).Msg("Deleted bucket")

	return &v1.DeleteBucketResponse{}, nil
}

// GetBucketPolicy implements v1.S3ServiceServer.
func (h *S3Handler) GetBucketPolicy(ctx context.Context, r *v1.GetBucketPolicyRequest) (*v1.GetBucketPolicyResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Msg("Getting bucket policy")

	policy, err := s.GetBucketPolicy(ctx, backend.GetBucketPolicyOpts{
		ID: r.GetBucketId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	if policy == nil {
		return nil, status.Error(codes.NotFound, "policy does not exists")
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Msg("Bucket policy found")

	return &v1.GetBucketPolicyResponse{
		BucketId: r.BucketId,
		Policy:   v1.BucketAccessPolicy(*policy),
	}, nil
}

// ListBuckets implements v1.S3ServiceServer.
func (h *S3Handler) ListBuckets(ctx context.Context, r *v1.ListBucketsRequest) (*v1.ListBucketsResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Listing buckets")

	bucketsResponse, err := s.ListBuckets(ctx, backend.ListBucketsOpts{
		Names: r.GetFilter().GetNames(),
	})

	if err != nil {
		return nil, err
	}

	buckets := make([]*v1.Bucket, 0)

	for _, bucket := range bucketsResponse {
		buckets = append(buckets, intoBucket(r.GetClusterId(), bucket))
	}

	return &v1.ListBucketsResponse{
		Buckets: buckets,
	}, nil
}

// UpdateBucketPolicy implements v1.S3ServiceServer.
func (h *S3Handler) UpdateBucketPolicy(ctx context.Context, r *v1.UpdateBucketPolicyRequest) (*v1.UpdateBucketPolicyResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Msg("Updating bucket policy")

	err := s.UpdateBucketPolicy(ctx, backend.UpdateBucketPolicyOpts{
		ID:           r.GetBucketId().GetId(),
		AccessPolicy: backend.AccessPolicy(r.AccessPolicy),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetBucketId().GetId()).Msg("Bucket policy updated")

	return &v1.UpdateBucketPolicyResponse{}, nil
}

// CreateLifecycleRule implements v1.S3ServiceServer.
func (h *S3Handler) CreateLifecycleRules(ctx context.Context, r *v1.CreateLifecycleRulesRequest) (*v1.CreateLifecycleRulesResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("Creating lifecycle rule")

	rules := make([]backend.LifecycleRule, 0)

	for _, rule := range r.LifecycleRules {
		rules = append(rules, backend.LifecycleRule{
			Prefix:               rule.GetPrefix(),
			ExpireDays:           rule.GetExpireDays(),
			NoncurrentExpireDays: rule.GetNoncurrentExpireDays(),
			DeleteMarker:         rule.GetDeleteMarker(),
		})
	}

	lr, err := s.CreateLifecycleRules(ctx, backend.CreateLifecycleRulesOpts{
		BucketID:       r.GetBucketId().GetId(),
		LifecycleRules: rules,
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("Lifecycle rule created")

	return &v1.CreateLifecycleRulesResponse{
		LifecycleRules: intoLifecycleRules(lr),
	}, nil
}

// DeleteLifecycleRule implements v1.S3ServiceServer.
func (h *S3Handler) DeleteLifecycleRules(ctx context.Context, r *v1.DeleteLifecycleRulesRequest) (*v1.DeleteLifecycleRulesResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("Delete lifecycle rules")

	err := s.DeleteLifecycleRules(ctx, backend.DeleteLifecycleRulesOpts{
		BucketID: r.GetBucketId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("Lifecycle rule deleted")

	return &v1.DeleteLifecycleRulesResponse{}, nil
}

// ListLifecycleRules implements v1.S3ServiceServer.
func (h *S3Handler) ListLifecycleRules(ctx context.Context, r *v1.ListLifecycleRulesRequest) (*v1.ListLifecycleRulesResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("List lifecycle rules")

	lrs, err := s.ListLifecycleRules(ctx, backend.ListLifecycleRulesOpts{
		BucketID: r.GetBucketId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &v1.ListLifecycleRulesResponse{
		LifecycleRules: intoLifecycleRules(lrs),
	}, nil
}

// UpdateLifecycleRule implements v1.S3ServiceServer.
func (h *S3Handler) UpdateLifecycleRules(ctx context.Context, r *v1.UpdateLifecycleRulesRequest) (*v1.UpdateLifecycleRulesResponse, error) {
	b := h.Backends[r.GetBucketId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetBucketId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("Update lifecycle rules")

	rules := make([]backend.LifecycleRule, 0)

	for _, rule := range r.LifecycleRules {
		rules = append(rules, backend.LifecycleRule{
			Prefix:               rule.GetPrefix(),
			ExpireDays:           rule.GetExpireDays(),
			NoncurrentExpireDays: rule.GetNoncurrentExpireDays(),
			DeleteMarker:         rule.GetDeleteMarker(),
		})
	}

	lr, err := s.UpdateLifecycleRules(ctx, backend.UpdateLifecycleRulesOpts{
		BucketID:       r.GetBucketId().GetId(),
		LifecycleRules: rules,
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("bucketId", r.GetBucketId().GetId()).Msg("List lifecycle updated")

	return &v1.UpdateLifecycleRulesResponse{
		LifecycleRules: intoLifecycleRules(lr),
	}, nil
}

// CreateS3Principal implements v1.S3ServiceServer.
func (h *S3Handler) CreateS3Principal(ctx context.Context, r *v1.CreateS3PrincipalRequest) (*v1.CreateS3PrincipalResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("username", r.GetName()).Msg("Creating S3 Principal")

	principal, err := s.CreateS3Principal(ctx, backend.CreateS3PrincipalOpts{
		Name:        r.GetName(),
		Credentials: r.GetCredentials(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("username", r.GetName()).Msg("Created S3 Principal")

	return &v1.CreateS3PrincipalResponse{
		S3Principal: intoS3Principal(r.GetClusterId(), principal),
	}, nil
}

// DeleteS3Principal implements v1.S3ServiceServer.
func (h *S3Handler) DeleteS3Principal(ctx context.Context, r *v1.DeleteS3PrincipalRequest) (*v1.DeleteS3PrincipalResponse, error) {
	b := h.Backends[r.GetPrincipalId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("Deleting S3 Principal")

	err := s.DeleteS3Principal(ctx, backend.DeleteS3PrincipalOpts{
		PrincipalID: r.GetPrincipalId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("Deleted S3 Principal")

	return &v1.DeleteS3PrincipalResponse{}, nil
}

// GetS3Principal implements v1.S3ServiceServer.
func (h *S3Handler) GetS3Principal(ctx context.Context, r *v1.GetS3PrincipalRequest) (*v1.GetS3PrincipalResponse, error) {
	b := h.Backends[r.GetPrincipalId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("Get S3 Principal")

	principal, err := s.GetS3Principal(ctx, backend.GetS3PrincipalOpts{
		PrincipalID: r.GetPrincipalId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &v1.GetS3PrincipalResponse{
		S3Principal: intoS3Principal(r.GetPrincipalId().GetClusterId(), principal),
	}, nil
}

// SetS3PrincipalCredentials implements v1.S3ServiceServer.
func (h *S3Handler) SetS3PrincipalCredentials(ctx context.Context, r *v1.SetS3PrincipalCredentialsRequest) (*v1.SetS3PrincipalCredentialsResponse, error) {
	b := h.Backends[r.GetPrincipalId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("Updating S3 Principal credentials")

	err := s.UpdateS3PrincipalPassword(ctx, backend.UpdateS3PrincipalPasswordOpts{
		PrincipalID: r.GetPrincipalId().GetId(),
		Credentials: r.GetCredentials(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("S3 Principal credentials set")

	return &v1.SetS3PrincipalCredentialsResponse{}, nil
}

// UpdateS3PrincipalPolicies implements v1.S3ServiceServer.
func (h *S3Handler) UpdateS3PrincipalPolicies(ctx context.Context, r *v1.UpdateS3PrincipalPoliciesRequest) (*v1.UpdateS3PrincipalPoliciesResponse, error) {
	b := h.Backends[r.GetPrincipalId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	s, ok := b.(backend.S3Ops)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetPrincipalId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support S3 operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support S3 operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("Updating S3 Policies")

	policies := make([]*backend.S3Policy, 0)

	for _, policy := range r.GetPolicies() {
		if policy != nil {
			allowNets, errs := backend.ParseNets(policy.GetSourceIpFilter().GetAllow())
			if len(errs) > 0 {
				log.Error().Errs("parseErrors", errs).Msg("Allow nets parse error")
				return nil, errors.New("Could not parse Allow Subnets")
			}
			disallowNets, errs := backend.ParseNets(policy.GetSourceIpFilter().GetDisallow())
			if len(errs) > 0 {
				log.Error().Errs("parseErrors", errs).Msg("Disallow nets parse error")
				return nil, errors.New("Could not parse Disallow Subnets")
			}

			policies = append(policies,
				&backend.S3Policy{
					BucketID:           policy.GetBucketId().GetId(),
					Prefix:             policy.GetPrefix(),
					Read:               policy.GetRead(),
					Write:              policy.GetWrite(),
					Delete:             policy.GetDelete(),
					Actions:            enumActionsIntoArray(policy.Actions),
					AllowSourceNets:    allowNets,
					DisallowSourceNets: disallowNets,
				})
		}
	}
	principal, err := s.UpdateS3PrincipalPolicies(ctx, backend.UpdateS3PrincipalPoliciesOpts{
		PrincipalID: r.GetPrincipalId().GetId(),
		Policies:    policies,
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetPrincipalId().GetId()).Msg("S3 Principal Policies updated")

	return &v1.UpdateS3PrincipalPoliciesResponse{
		S3Principal: intoS3Principal(r.GetPrincipalId().GetClusterId(), principal),
	}, nil
}

func intoBucket(clusterId *v1.ClusterIdentifier, bucket *backend.Bucket) *v1.Bucket {
	if bucket == nil || clusterId == nil {
		return nil
	}

	return &v1.Bucket{
		Id: &v1.BucketIdentifier{
			ClusterId: clusterId,
			Id:        bucket.ID,
		},
		Name:      bucket.Name,
		Versioned: bucket.Versioned,
		Capacity: &v1.Bucket_Capacity{
			TotalBytes:     bucket.QuotaBytes,
			AvailableBytes: bucket.AvailableBytes,
		},
		EndpointUrl: bucket.EndpointURL,
	}
}

func intoLifecycleRules(rules []*backend.LifecycleRule) []*v1.LifecycleRule {
	answer := make([]*v1.LifecycleRule, 0)
	for _, rule := range rules {
		if rule != nil {
			answer = append(answer, &v1.LifecycleRule{
				Id: &v1.LifecycleRuleIdentifier{
					Id: rule.ID,
				},
				Prefix:               rule.Prefix,
				ExpireDays:           rule.ExpireDays,
				NoncurrentExpireDays: rule.NoncurrentExpireDays,
				DeleteMarker:         rule.DeleteMarker,
			})
		}
	}

	return answer
}

func intoS3Principal(clusterId *v1.ClusterIdentifier, principal *backend.S3Principal) *v1.S3Principal {
	if principal == nil || clusterId == nil {
		return nil
	}

	policies := make([]*v1.S3Principal_Policy, 0)

	for _, policy := range principal.Policies {
		var allowNets []string
		var disallowNets []string
		for _, cidr := range policy.AllowSourceNets {
			allowNets = append(allowNets, cidr.String())
		}
		for _, cidr := range policy.DisallowSourceNets {
			disallowNets = append(disallowNets, cidr.String())
		}
		policies = append(policies,
			&v1.S3Principal_Policy{
				BucketId: &v1.BucketIdentifier{
					ClusterId: clusterId,
					Id:        policy.BucketID,
				},
				Prefix:  policy.Prefix,
				Read:    policy.Read,
				Write:   policy.Write,
				Delete:  policy.Delete,
				Actions: arrayIntoActionsEnum(policy.Actions),
				SourceIpFilter: &v1.S3Principal_Policy_SourceIpFilter{
					Allow:    allowNets,
					Disallow: disallowNets,
				},
			})
	}

	return &v1.S3Principal{
		Id: &v1.S3PrincipalIdentifier{
			ClusterId: clusterId,
			Id:        principal.ID,
		},
		Name:     principal.Name,
		Policies: policies,
	}
}

func enumActionsIntoArray(actions []v1.S3Principal_Policy_BucketActions) []string {
	if len(actions) == 0 {
		return []string{
			"s3:GetBucketLocation",
			"s3:GetBucketPolicy",
			"s3:ListBucket",
			"s3:ListBucketMultipartUploads",
			"s3:ListMultipartUploadParts",
			"s3:GetBucketTagging",
			"s3:ListBucketVersions",
		}
	}
	answer := make([]string, 0)
	for _, action := range actions {
		switch action {
		case v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION:
			answer = append(answer, "s3:GetBucketLocation")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY:
			answer = append(answer, "s3:GetBucketPolicy")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET:
			answer = append(answer, "s3:ListBucket")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_MULTIPART_UPLOADS:
			answer = append(answer, "s3:ListBucketMultipartUploads")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_MULTIPART_UPLOAD_PARTS:
			answer = append(answer, "s3:ListMultipartUploadParts")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_TAGGING:
			answer = append(answer, "s3:GetBucketTagging")
		case v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_VERSIONS:
			answer = append(answer, "s3:ListBucketVersions")
		}
	}

	return answer
}

func arrayIntoActionsEnum(actions []string) []v1.S3Principal_Policy_BucketActions {
	answer := make([]v1.S3Principal_Policy_BucketActions, 0)
	for _, action := range actions {
		switch action {
		case "s3:GetBucketLocation":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION)
		case "s3:GetBucketPolicy":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY)
		case "s3:ListBucket":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET)
		case "s3:ListBucketMultipartUploads":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_MULTIPART_UPLOADS)
		case "s3:ListMultipartUploadParts":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_MULTIPART_UPLOAD_PARTS)
		case "s3:GetBucketTagging":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_TAGGING)
		case "s3:ListBucketVersions":
			answer = append(answer, v1.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_VERSIONS)
		}
	}

	return answer
}
