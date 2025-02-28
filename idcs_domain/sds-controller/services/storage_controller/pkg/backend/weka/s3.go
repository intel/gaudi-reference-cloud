// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateBucket(ctx context.Context, opts backend.CreateBucketOpts) (*backend.Bucket, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	policy := intoPolicyName(&opts.AccessPolicy)
	quota := fmt.Sprintf("%sB", strconv.FormatUint(opts.QuotaBytes, 10))

	resp, err := b.client.CreateS3BucketWithResponse(ctx, v4.CreateS3BucketJSONRequestBody{
		BucketName: opts.Name,
		Policy:     &policy,
		HardQuota:  &quota,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not create bucket", resp.StatusCode(), resp.Body)
	}

	return &backend.Bucket{
		ID:             opts.Name,
		Name:           opts.Name,
		Versioned:      false,
		QuotaBytes:     opts.QuotaBytes,
		AvailableBytes: opts.QuotaBytes,
		AccessPolicy:   opts.AccessPolicy,
	}, nil
}

func (b *Backend) DeleteBucket(ctx context.Context, opts backend.DeleteBucketOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	resp, err := b.client.DestroyS3BucketWithResponse(ctx, opts.ID, nil, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete bucket", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) GetBucketPolicy(ctx context.Context, opts backend.GetBucketPolicyOpts) (*backend.AccessPolicy, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetS3BucketPolicyWithResponse(ctx, opts.ID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get bucket policy", resp.StatusCode(), resp.Body)
	}

	if resp.JSON200.Data.Policy == nil {
		return nil, nil
	}

	policy := fromPolicyName(*resp.JSON200.Data.Policy)

	return &policy, nil
}

func (b *Backend) ListBuckets(ctx context.Context, opts backend.ListBucketsOpts) ([]*backend.Bucket, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getBuckets(ctx, opts.Names, token)
}

func (b *Backend) UpdateBucketPolicy(ctx context.Context, opts backend.UpdateBucketPolicyOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	policy := intoPolicyName(&opts.AccessPolicy)

	resp, err := b.client.SetS3BucketPolicyWithResponse(ctx, opts.ID, v4.SetS3BucketPolicyJSONRequestBody{
		BucketPolicy: policy,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not set bucket policy", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) CreateLifecycleRules(ctx context.Context, opts backend.CreateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	mutex := b.acquireLrMutex(opts.BucketID)
	mutex.Lock()
	defer mutex.Unlock()

	err = b.deleteLifecycleRules(ctx, opts.BucketID, token)
	if err != nil {
		// Info as error is normal operation
		log.Info().Err(err).Ctx(ctx).Msg("Lifecycle rules was not cleared from bucket, as they not exists")
	}

	return b.createLifecycleRules(ctx, opts.BucketID, opts.LifecycleRules, token)

}

func (b *Backend) DeleteLifecycleRules(ctx context.Context, opts backend.DeleteLifecycleRulesOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	mutex := b.acquireLrMutex(opts.BucketID)
	mutex.Lock()
	defer mutex.Unlock()

	return b.deleteLifecycleRules(ctx, opts.BucketID, token)
}

func (b *Backend) ListLifecycleRules(ctx context.Context, opts backend.ListLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getLifecycleRules(ctx, opts.BucketID, token)
}

func (b *Backend) UpdateLifecycleRules(ctx context.Context, opts backend.UpdateLifecycleRulesOpts) ([]*backend.LifecycleRule, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	mutex := b.acquireLrMutex(opts.BucketID)
	mutex.Lock()
	defer mutex.Unlock()

	err = b.deleteLifecycleRules(ctx, opts.BucketID, token)

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Could not update lifecycle rule")
		return nil, err
	}

	return b.createLifecycleRules(ctx, opts.BucketID, opts.LifecycleRules, token)
}

func (b *Backend) CreateS3Principal(ctx context.Context, opts backend.CreateS3PrincipalOpts) (*backend.S3Principal, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	role := v4.CreateUserJSONBodyRoleS3
	resp, err := b.client.CreateUserWithResponse(ctx, v4.CreateUserJSONRequestBody{
		Username: &opts.Name,
		Password: &opts.Credentials,
		Role:     &role,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not create S3 Principal", resp.StatusCode(), resp.Body)
	}

	return intoS3Principal(*resp.JSON200.Data)
}

func (b *Backend) DeleteS3Principal(ctx context.Context, opts backend.DeleteS3PrincipalOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	_, err = b.client.DeleteS3PolicyWithResponse(ctx, opts.PrincipalID, v4.RequestEditorFn(token.Intercept))
	if err == nil {
		log.Info().Ctx(ctx).Str("principalId", opts.PrincipalID).Msg("S3 IAM Policy deleted")
	}
	resp, err := b.client.DeleteUserWithResponse(ctx, opts.PrincipalID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete S3 Principal", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) GetS3Principal(ctx context.Context, opts backend.GetS3PrincipalOpts) (*backend.S3Principal, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getS3Principal(ctx, opts.PrincipalID, token)
}

func (b *Backend) UpdateS3PrincipalPassword(ctx context.Context, opts backend.UpdateS3PrincipalPasswordOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	resp, err := b.client.SetUserPasswordWithResponse(ctx, opts.PrincipalID, v4.SetUserPasswordJSONRequestBody{
		Password: opts.Credentials,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not set S3 Principal credentials", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) UpdateS3PrincipalPolicies(ctx context.Context, opts backend.UpdateS3PrincipalPoliciesOpts) (*backend.S3Principal, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	principal, err := b.getS3Principal(ctx, opts.PrincipalID, token)

	if err != nil {
		return nil, err
	}

	if len(principal.Policies) != 0 {
		_, err = b.client.DeleteS3PolicyWithResponse(ctx, opts.PrincipalID, v4.RequestEditorFn(token.Intercept))
		if err != nil {
			log.Error().Err(err).Msg("Could not delete principal IAM policy")
		}
	}

	policy := backend.FromIAMPolicy(opts.Policies)
	policyContent, err := toIAMPolicy(policy)

	if policyContent == nil || err != nil {
		log.Error().Err(err).Msg("Could not prepare policy for IAM")
		return nil, err
	}

	resp, err := b.client.CreateS3PolicyWithResponse(ctx, v4.CreateS3PolicyJSONRequestBody{
		PolicyName:        opts.PrincipalID,
		PolicyFileContent: *policyContent,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 {
		log.Error().Any("policy", policy).Ctx(ctx).Err(err).Msg("Could not apply policy")
		return nil, backend.ResponseAsErr("could not set S3 Principal Policies", resp.StatusCode(), resp.Body)
	}

	attachResp, err := b.client.AttachS3PolicyWithResponse(ctx, v4.AttachS3PolicyJSONRequestBody{
		PolicyName: opts.PrincipalID,
		UserName:   principal.Name,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if attachResp.StatusCode() != 200 {
		return nil, backend.ResponseAsErr("could not set attach policy to principal", resp.StatusCode(), resp.Body)
	}

	return &backend.S3Principal{
		ID:       principal.ID,
		Name:     principal.Name,
		Policies: opts.Policies,
	}, nil
}

func (b *Backend) getBuckets(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.Bucket, error) {
	resp, err := b.client.GetS3BucketsWithResponse(ctx, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil || resp.JSON200.Data.Buckets == nil {
		return nil, backend.ResponseAsErr("could not list buckets", resp.StatusCode(), resp.Body)
	}

	buckets := make([]*backend.Bucket, 0)

	for _, bucket := range *resp.JSON200.Data.Buckets {
		if bucket.Name == nil {
			log.Warn().Ctx(ctx).Msg("Bucket without name returned")
			continue
		}
		if len(namesToFilter) > 0 && !slices.Contains(namesToFilter, *bucket.Name) {
			continue
		}
		quotaBytes := helpers.ValueOrNil(bucket.HardLimitBytes)
		usedBytes := helpers.ValueOrNil(bucket.UsedBytes)

		bucket := backend.Bucket{
			ID:             *bucket.Name,
			Name:           *bucket.Name,
			Versioned:      false,
			QuotaBytes:     quotaBytes,
			AvailableBytes: quotaBytes - usedBytes,
		}

		buckets = append(buckets, &bucket)
	}

	return buckets, nil
}

func (b *Backend) getS3Principal(ctx context.Context, principalID string, token *securityprovider.SecurityProviderBearerToken) (*backend.S3Principal, error) {
	resp, err := b.getUsers(ctx, make([]string, 0), token)

	if err != nil {
		return nil, err
	}

	for _, u := range resp {
		if u.ID == principalID {
			policies := make([]*backend.S3Policy, 0)
			policyJson, err := b.client.GetS3PolicyWithResponse(ctx, u.ID, v4.RequestEditorFn(token.Intercept))
			if err == nil && policyJson.JSON200 != nil && policyJson.JSON200.Data != nil &&
				policyJson.JSON200.Data.Policy != nil && policyJson.JSON200.Data.Policy.Content != nil {
				policy := fromIAMPolicy(*policyJson.JSON200.Data.Policy.Content)
				policies = backend.IntoIAMPolicies(policy)
			}
			return &backend.S3Principal{
				ID:       u.ID,
				Name:     u.Name,
				Policies: policies,
			}, nil
		}
	}

	return nil, errors.New("could not find S3 Principal by id")
}

func (b *Backend) getLifecycleRules(ctx context.Context, bucketID string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.LifecycleRule, error) {
	resp, err := b.client.S3ListAllLifecycleRulesWithResponse(ctx, bucketID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get lifecycle rules", resp.StatusCode(), resp.Body)
	}

	lrs := make([]*backend.LifecycleRule, 0)

	for _, rule := range *resp.JSON200.Data.Rules {
		var expireDays uint64
		if rule.ExpiryDays != nil {
			expireDays, _ = strconv.ParseUint(*rule.ExpiryDays, 10, 32)
		}
		lr := backend.LifecycleRule{
			ID:                   helpers.ValueOrNil(rule.Id),
			Prefix:               helpers.ValueOrNil(rule.Prefix),
			ExpireDays:           uint32(expireDays),
			NoncurrentExpireDays: 0,
			DeleteMarker:         false,
		}

		lrs = append(lrs, &lr)
	}

	return lrs, nil
}

func (b *Backend) createLifecycleRules(ctx context.Context, bucketID string, rules []backend.LifecycleRule,
	token *securityprovider.SecurityProviderBearerToken) ([]*backend.LifecycleRule, error) {
	answer := make([]*backend.LifecycleRule, 0)
	for _, rule := range rules {
		resp, err := b.client.S3CreateLifecycleRuleWithResponse(ctx, bucketID, v4.S3CreateLifecycleRuleJSONRequestBody{
			ExpiryDays: strconv.FormatUint(uint64(rule.ExpireDays),  10),
			Prefix:     &rule.Prefix,
		}, v4.RequestEditorFn(token.Intercept))

		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
			continue
		}

		if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil || resp.JSON200.Data.Id == nil {
			log.Error().Ctx(ctx).Msgf("could not create lifecycle rule, status: %d, body: %v", resp.StatusCode(), string(resp.Body))
			continue
		}

		answer = append(answer, &backend.LifecycleRule{
			ID:                   *resp.JSON200.Data.Id,
			Prefix:               rule.Prefix,
			ExpireDays:           rule.ExpireDays,
			DeleteMarker:         false,
			NoncurrentExpireDays: 0,
		})
	}

	return answer, nil
}

func (b *Backend) deleteLifecycleRules(ctx context.Context, bucketID string,
	token *securityprovider.SecurityProviderBearerToken) error {
	resp, err := b.client.S3DeleteAllLifecycleRulesWithResponse(ctx, bucketID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete lifecycle rules", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) acquireLrMutex(bucketId string) *sync.Mutex {
	b.lock.Lock()
	defer b.lock.Unlock()
	mutex, exists := b.lrLocks[bucketId]
	if !exists {
		mutex = &sync.Mutex{}
		b.lrLocks[bucketId] = mutex
	}
	return mutex
}

func intoPolicyName(policy *backend.AccessPolicy) string {
	switch *policy {
	case backend.None:
		return "none"
	case backend.Read:
		return "download"
	case backend.ReadWrite:
		return "public"
	default:
		return "none"
	}
}

func fromPolicyName(policyName string) backend.AccessPolicy {
	switch policyName {
	case "none":
		return backend.None
	case "download":
		return backend.Read
	case "public":
		return backend.ReadWrite
	default:
		return backend.None
	}
}

func fromIAMPolicy(policy v4.S3IAMPolicy) backend.S3IAMPolicy {
	statements := make([]backend.S3IAMStatement, 0)
	for _, statement := range *policy.Statement {
		bstatement := backend.S3IAMStatement{
			Action:   statement.Action,
			Effect:   statement.Effect,
			Resource: statement.Resource,
			Sid:      statement.Sid,
		}
		statements = append(statements, bstatement)
	}
	return backend.S3IAMPolicy{
		Statement: &statements,
		Version:   policy.Version,
	}
}

func toIAMPolicy(policy backend.S3IAMPolicy) (*v4.S3IAMPolicy, error) {
	statements := make([]v4.S3IAMStatement, 0)
	for _, statement := range *policy.Statement {
		if statement.Condition != nil {
			return nil, errors.New("weka backend does not support policy conditions")
		}
		wstatement := v4.S3IAMStatement{
			Action:   statement.Action,
			Effect:   statement.Effect,
			Resource: statement.Resource,
			Sid:      statement.Sid,
		}
		statements = append(statements, wstatement)
	}
	return &v4.S3IAMPolicy{
		Statement: &statements,
		Version:   policy.Version,
	}, nil
}
