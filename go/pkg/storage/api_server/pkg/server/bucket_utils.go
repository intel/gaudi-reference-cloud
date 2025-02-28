package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func isValidBucketCreateRequest(ctx context.Context,
	request *pb.ObjectBucketCreatePrivateRequest,
	storageProduct objectProductInfo,
	skipProductCheck bool) error {
	logger := log.FromContext(ctx).WithName("isValidBucketCreateRequest")
	logger.Info("validating bucket create request")

	// Validate input.
	if request.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if request.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}

	// Calculate name if not provided.
	if request.Metadata.Name == "" {
		return status.Error(codes.InvalidArgument, "missing name")
	}
	name := request.Metadata.Name

	if err := utils.ValidateResourceName(name); err != nil {
		return err
	}
	if err := cloudaccount.CheckValidId(request.Metadata.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketGetRequest(ctx context.Context, md *pb.ObjectBucketMetadataRef) error {
	logger := log.FromContext(ctx).WithName("isValidBucketGetRequest")
	logger.Info("validating bucket get request")

	if md.CloudAccountId == "" || (md.GetBucketName() == "" && md.GetBucketId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate resourceId.
	if md.GetBucketId() != "" {
		if _, err := uuid.Parse(md.GetBucketId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	if md.GetBucketName() != "" {
		if err := utils.ValidateResourceName(md.GetBucketName()); err != nil {
			return err
		}
	}
	if err := cloudaccount.CheckValidId(md.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketSearchRequest(ctx context.Context, request *pb.ObjectBucketSearchRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketSearchRequest")
	logger.Info("validating bucket search request")

	if request.Metadata.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing cloudaccount")
	}
	if err := cloudaccount.CheckValidId(request.Metadata.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketDeleteRequest(ctx context.Context, md *pb.ObjectBucketMetadataRef) error {
	logger := log.FromContext(ctx).WithName("isValidBucketDeleteRequest")
	logger.Info("validating bucket delete request")

	if md.CloudAccountId == "" || (md.GetBucketName() == "" && md.GetBucketId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate resourceId.
	if md.GetBucketId() != "" {
		if _, err := uuid.Parse(md.GetBucketId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	if md.GetBucketName() != "" {
		if err := utils.ValidateResourceName(md.GetBucketName()); err != nil {
			return err
		}
	}
	if err := cloudaccount.CheckValidId(md.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketUserCreateRequest(ctx context.Context, user *pb.CreateObjectUserPrivateRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketUserCreateRequest")
	logger.Info("validating bucket user create request")
	// Validate input.
	if user.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if user.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}
	if user.Metadata.Name == "" {
		return status.Error(codes.InvalidArgument, "missing name")
	}
	if user.Metadata.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing cloudaccount")
	}
	if err := utils.ValidateResourceName(user.Metadata.Name); err != nil {
		return err
	}
	if err := cloudaccount.CheckValidId(user.Metadata.CloudAccountId); err != nil {
		return err
	}
	allowedPermissions := map[pb.BucketPermission]struct{}{
		pb.BucketPermission_DeleteBucket: struct{}{},
		pb.BucketPermission_ReadBucket:   struct{}{},
		pb.BucketPermission_WriteBucket:  struct{}{},
	}
	allowedActions := map[pb.ObjectBucketActions]struct{}{
		pb.ObjectBucketActions_GetBucketLocation:          struct{}{},
		pb.ObjectBucketActions_GetBucketPolicy:            struct{}{},
		pb.ObjectBucketActions_GetBucketTagging:           struct{}{},
		pb.ObjectBucketActions_ListBucket:                 struct{}{},
		pb.ObjectBucketActions_ListBucketMultipartUploads: struct{}{},
		pb.ObjectBucketActions_ListMultipartUploadParts:   struct{}{},
	}

	for _, spec := range user.Spec {
		for _, action := range spec.Actions {
			if _, found := allowedActions[action]; !found {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid action %s", action))
			}
		}
	}
	for _, spec := range user.Spec {
		for _, perm := range spec.Permission {
			if _, found := allowedPermissions[perm]; !found {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid permission %s", perm))
			}
		}
	}
	// validate prefix
	for _, policy := range user.Spec {
		if err := utils.ValidatePrefix(policy.Prefix); err != nil {
			return err
		}
	}
	return nil
}

func isValidBucketUserUpdateRequest(ctx context.Context, user *pb.ObjectUserUpdatePrivateRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketUserCreateRequest")
	logger.Info("validating bucket user create request")
	// Validate input.
	if user.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if user.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}
	if user.Metadata.GetUserName() != "" {
		if err := utils.ValidateResourceName(user.Metadata.GetUserName()); err != nil {
			return err
		}
	}
	if user.Metadata.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing cloudaccount")
	}
	if err := cloudaccount.CheckValidId(user.Metadata.CloudAccountId); err != nil {
		return err
	}
	allowedPermissions := map[pb.BucketPermission]struct{}{
		pb.BucketPermission_DeleteBucket: struct{}{},
		pb.BucketPermission_ReadBucket:   struct{}{},
		pb.BucketPermission_WriteBucket:  struct{}{},
	}
	allowedActions := map[pb.ObjectBucketActions]struct{}{
		pb.ObjectBucketActions_GetBucketLocation:          struct{}{},
		pb.ObjectBucketActions_GetBucketPolicy:            struct{}{},
		pb.ObjectBucketActions_GetBucketTagging:           struct{}{},
		pb.ObjectBucketActions_ListBucket:                 struct{}{},
		pb.ObjectBucketActions_ListBucketMultipartUploads: struct{}{},
		pb.ObjectBucketActions_ListMultipartUploadParts:   struct{}{},
	}

	for _, spec := range user.Spec {
		for _, action := range spec.Actions {
			if _, found := allowedActions[action]; !found {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid action %s", action))
			}
		}
	}
	for _, spec := range user.Spec {
		for _, perm := range spec.Permission {
			if _, found := allowedPermissions[perm]; !found {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid permission %s", perm))
			}
		}
	}
	// validate prefix (optional)
	for _, policy := range user.Spec {
		if err := utils.ValidatePrefix(policy.Prefix); err != nil {
			return err
		}
	}
	return nil
}

func isValidBucketUserGetRequest(ctx context.Context, user *pb.ObjectUserGetPrivateRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketUserGetRequest")
	logger.Info("validating bucket user get request")

	if user.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}

	if user.Metadata.CloudAccountId == "" || (user.Metadata.GetUserName() == "" && user.Metadata.GetUserId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate userId.
	if user.Metadata.GetUserId() != "" {
		if _, err := uuid.Parse(user.Metadata.GetUserId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid userId")
		}
	}
	if user.Metadata.GetUserName() != "" {
		if err := utils.ValidateResourceName(user.Metadata.GetUserName()); err != nil {
			return err
		}
	}
	if err := cloudaccount.CheckValidId(user.Metadata.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketUserSearchRequest(ctx context.Context, user *pb.ObjectUserSearchRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketUserSearchRequest")
	logger.Info("validating bucket user search request")

	if user.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing cloudaccount")
	}
	if err := cloudaccount.CheckValidId(user.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidBucketUserDeleteRequest(ctx context.Context, user *pb.ObjectUserDeletePrivateRequest) error {
	logger := log.FromContext(ctx).WithName("isValidBucketUserDeleteRequest")
	logger.Info("validating bucket user delete request")

	if user.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}

	if user.Metadata.CloudAccountId == "" || (user.Metadata.GetUserName() == "" && user.Metadata.GetUserId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate userId.
	if user.Metadata.GetUserId() != "" {
		if _, err := uuid.Parse(user.Metadata.GetUserId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid userId")
		}
	}
	if user.Metadata.GetUserName() != "" {
		if err := utils.ValidateResourceName(user.Metadata.GetUserName()); err != nil {
			return err
		}
	}
	if err := cloudaccount.CheckValidId(user.Metadata.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func isValidLifecycleCreateRequest(in *pb.BucketLifecycleRuleCreatePrivateRequest) error {

	if in.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if in.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}
	if in.Metadata.BucketId == "" {
		return status.Error(codes.InvalidArgument, "missing bucketId in metadata")
	}
	if in.Metadata.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing cloudaccountId in metadata")
	}
	if in.Metadata.RuleName == "" {
		return status.Error(codes.InvalidArgument, "missing ruleName in metadata")
	}
	if err := utils.ValidateResourceName(in.Metadata.RuleName); err != nil {
		return err
	}
	if err := cloudaccount.CheckValidId(in.Metadata.CloudAccountId); err != nil {
		return err
	}
	if in.Spec.DeleteMarker && in.Spec.ExpireDays > 0 {
		return status.Error(codes.InvalidArgument, "delete marker and expiry days cannot both be set")
	}
	if in.Spec.NoncurrentExpireDays > 3650 || in.Spec.ExpireDays > 3650 {
		return status.Error(codes.InvalidArgument, "maximum expiry and noncurrent expiry days allowed is 3650")
	}
	if err := utils.ValidatePrefix(in.Spec.Prefix); err != nil {
		return err
	}
	return nil
}

func isValidLifecycleUpdateRequest(ctx context.Context, ruleSpec *pb.BucketLifecycleRuleSpec) error {
	logger := log.FromContext(ctx).WithName("isValidLifecycleUpdateRequest")
	logger.Info("validating bucket lifecycle update request")

	if ruleSpec.DeleteMarker && ruleSpec.ExpireDays > 0 {
		return status.Error(codes.InvalidArgument, "delete marker and expiry days cannot both be set")
	}
	if ruleSpec.NoncurrentExpireDays > 3650 || ruleSpec.ExpireDays > 3650 {
		return status.Error(codes.InvalidArgument, "maximum expiry and noncurrent expiry days allowed is 3650")
	}
	if err := utils.ValidatePrefix(ruleSpec.Prefix); err != nil {
		return err
	}
	return nil
}

func isValidBucketLifecycleGetRequest(ctx context.Context, md *pb.BucketLifecycleRuleMetadataRef) error {
	logger := log.FromContext(ctx).WithName("isValidBucketLifecycleGetRequest")
	logger.Info("validating bucket lifecycle get request")

	if md.CloudAccountId == "" || (md.RuleId == "" || md.BucketId == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}

	// Validate bucketId.
	if md.BucketId != "" {
		if _, err := uuid.Parse(md.BucketId); err != nil {
			return status.Error(codes.InvalidArgument, "invalid bucketId")
		}
	}
	// Validate ruleId.
	if md.RuleId != "" {
		if _, err := uuid.Parse(md.RuleId); err != nil {
			return status.Error(codes.InvalidArgument, "invalid ruleId")
		}
	}
	if err := cloudaccount.CheckValidId(md.CloudAccountId); err != nil {
		return err
	}
	return nil
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func removeNetworkFilter(slice []*pb.BucketNetworkGroup, idx int) []*pb.BucketNetworkGroup {
	return append(slice[:idx], slice[idx+1:]...)
}

func (bk *BucketsServiceServer) lookupResourcePermission(ctx context.Context, cloudaccount string, action string, resource string, resourceIds []string, search bool) (*v1.LookupResponse, error) {
	logger := log.FromContext(ctx).WithName("lookupResourcePermission")
	logger.Info("check resource permission with Authz")

	email, enterpriseId, groups, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Error(err, "failed to extract token info from context")
	}
	if email == nil || enterpriseId == nil {
		cloudAccountInfo, err := bk.cloudAccountServiceClient.GetById(ctx, &v1.CloudAccountId{Id: cloudaccount})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "required cloud account info could not be fetched")
		}
		email = &cloudAccountInfo.Name
		enterpriseId = &cloudAccountInfo.Oid
	}

	authzLookupResponse, err := bk.authzServiceClient.LookupInternal(ctx, &v1.LookupRequestInternal{
		CloudAccountId: cloudaccount,
		ResourceType:   resource,
		Action:         action,
		ResourceIds:    resourceIds,
		User: &v1.UserIdentification{
			Email:        *email,
			EnterpriseId: *enterpriseId,
			Groups:       groups,
		},
	})

	if err != nil {
		logger.Error(err, "error authz lookup")
		return nil, status.Errorf(codes.Internal, "error authz lookup")
	}
	logger.Info("lookup", "action", action, "resource", resource)
	logger.Info("request resourceIds", "sent", resourceIds)
	logger.Info("Response resourceIds", "allowed", authzLookupResponse.ResourceIds)
	if len(authzLookupResponse.ResourceIds) == 0 && !search {
		logger.Error(err, "error permission deny for this resource")
		return nil, status.Errorf(codes.PermissionDenied, "permission denied for this resource")
	}
	return authzLookupResponse, nil
}
