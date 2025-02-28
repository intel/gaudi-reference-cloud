// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"errors"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthorizationService struct {
	casbinEngine                       *casbin.Enforcer
	permissionEngine                   *PermissionEngine
	resourceRepository                 *ResourceRepository
	auditLogging                       *AuditLogging
	pb.UnimplementedAuthzServiceServer // Used for forward compatability
}

func (cs *AuthorizationService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.Ping").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) Check(ctx context.Context, obj *pb.
	AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.Check").WithValues("cloudAccountId", obj.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := obj.Validate(); err != nil {
		logger.Error(err, "validation failed for authorization request")
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters")
	}

	subject, enterpriseId, _, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Error(err, "failed to extract token info from context")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if subject == nil || enterpriseId == nil {
		logger.Error(err, "subject or enterpriseId not provided in token")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	return cs.CheckInternal(ctx, &pb.AuthorizationRequestInternal{CloudAccountId: obj.CloudAccountId, User: &pb.UserIdentification{Email: *subject, EnterpriseId: *enterpriseId}, Path: obj.Path, Verb: obj.Verb, Payload: obj.Payload})
}

func (cs *AuthorizationService) CheckInternal(ctx context.Context, obj *pb.
	AuthorizationRequestInternal) (*pb.AuthorizationResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.CheckInternal").WithValues("cloudAccountId", obj.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	decisionId := uuid.NewString()
	if err := obj.Validate(); err != nil {
		logger.Error(err, "validation failed for authorization request")
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters")
	}

	allowed, permissions, err := cs.casbinEngine.EnforceEx(decisionId, obj.CloudAccountId, obj.User.Email, obj.Path, obj.Verb, obj.Payload.AsMap())

	if err != nil {
		logger.Error(err, "casbin enforcement failed", "enterpriseId", obj.User.EnterpriseId, "path", obj.Path, "verb", obj.Verb)
		return nil, status.Error(codes.Internal, "authorization check failed")
	}

	logger.V(9).Info("enforce result", "cloudAccountId", obj.CloudAccountId, "enterpriseId", obj.User.EnterpriseId, "path", obj.Path, "verb", obj.Verb, "payload", obj.Payload, "result", allowed)

	additionalInfo := structs.Map(obj)
	additionalInfo["Allowed"] = allowed
	additionalInfo["Permissions"] = permissions
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			id:             decisionId,
			cloudAccountId: obj.CloudAccountId,
			eventType:      "check",
			additionalInfo: additionalInfo,
		})
	return &pb.AuthorizationResponse{Allowed: allowed}, nil
}

func (cs *AuthorizationService) Lookup(ctx context.Context, request *pb.LookupRequest) (*pb.LookupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.Lookup").WithValues("cloudAccountId", request.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("request", "request", request)

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for lookup request")
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters")
	}

	subject, enterpriseId, _, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Error(err, "failed to extract token info from context")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if subject == nil || enterpriseId == nil {
		logger.Error(err, "subject or enterpriseId not provided in token")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	return cs.LookupInternal(ctx, &pb.LookupRequestInternal{CloudAccountId: request.CloudAccountId, User: &pb.UserIdentification{Email: *subject, EnterpriseId: *enterpriseId}, ResourceType: request.ResourceType, ResourceIds: request.ResourceIds, Action: request.Action})
}

func (cs *AuthorizationService) LookupInternal(ctx context.Context, request *pb.LookupRequestInternal) (*pb.LookupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.LookupInternal").WithValues("cloudAccountId", request.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for lookup request")
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters")
	}

	err := cs.resourceRepository.validateResource(request.ResourceType, request.Action)

	if err != nil {
		logger.Error(err, "resource validation failed", "resourceType", request.ResourceType, "action", request.Action)
		return nil, status.Error(codes.InvalidArgument, "invalid resource or action specified")
	}

	userSystemRoles := cs.getSystemRolesbyUser(ctx, request.User.Email, request.CloudAccountId)

	isAdmin := IsAdmin(ctx, request.User.Email, request.User.Groups, userSystemRoles)

	if isAdmin {
		return &pb.LookupResponse{ResourceType: request.ResourceType, ResourceIds: request.ResourceIds}, nil
	}

	allowedResourcesIds, err := cs.permissionEngine.LookupPermissionAllowedResources(ctx, request.CloudAccountId, request.User.Email, request.ResourceType, request.ResourceIds, request.Action)

	if err != nil {
		logger.Error(err, "failed to lookup permissions", "enterpriseId", request.User.EnterpriseId, "resourceType", request.ResourceType, "action", request.Action)
		return nil, status.Error(codes.Internal, "permission lookup failed")
	}

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId: request.CloudAccountId,
			eventType:      "lookup",
			additionalInfo: structs.Map(request),
		})

	return &pb.LookupResponse{ResourceType: request.ResourceType, ResourceIds: allowedResourcesIds}, nil
}

func (cs *AuthorizationService) Actions(ctx context.Context, obj *pb.ActionsRequest) (*pb.ActionsResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.Actions").WithValues("cloudAccountId", obj.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := obj.Validate(); err != nil {
		logger.Error(err, "failed to validate actionsRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters")
	}

	err := cs.resourceRepository.validateResource(obj.ResourceType, []string{}...)
	if err != nil {
		logger.Error(err, "failed to validate resource", "resourceType", obj.ResourceType)
		return nil, status.Error(codes.InvalidArgument, "invalid resource specified")

	}

	subject, enterpriseId, groups, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Error(err, "failed to extract token info from context")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if subject == nil || enterpriseId == nil {
		logger.Error(err, "subject or enterpriseId not provided in token")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	logger = logger.WithValues("enterpriseId", enterpriseId)

	userSystemRoles := cs.getSystemRolesbyUser(ctx, *subject, obj.CloudAccountId)

	isAdmin := IsAdmin(ctx, *subject, groups, userSystemRoles)

	allowedActions := []string{}
	if isAdmin {
		resource := cs.resourceRepository.Get(obj.ResourceType)
		if resource != nil {
			allowedActions = resource.GetActionNames()
		}
	} else {
		allowedActions, err = cs.permissionEngine.LookupPermissionAllowedActions(ctx, obj.CloudAccountId, *subject, obj.ResourceType, []string{obj.ResourceId}, "")
		if err != nil {
			logger.Error(err, "failed to get allowed actions", "enterpriseId", enterpriseId, "resourceType", obj.ResourceType, "resourceId", obj.ResourceId)
			return nil, status.Error(codes.Internal, "failed to retrieve allowed actions")
		}
	}

	return &pb.ActionsResponse{Actions: allowedActions}, nil
}

func (cs *AuthorizationService) CreateCloudAccountRole(ctx context.Context, cloudAccountRole *pb.CloudAccountRole) (*pb.CloudAccountRole, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.CreateCloudAccountRole").WithValues("cloudAccountId", cloudAccountRole.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRole.Validate(); err != nil {
		logger.Error(err, "validation failed for cloud account role")
		return nil, status.Error(codes.InvalidArgument, "invalid role data provided")
	}

	// validate permissions
	for _, permission := range cloudAccountRole.Permissions {
		err := cs.resourceRepository.validateResource(permission.ResourceType, permission.Actions...)
		if err != nil {
			logger.Error(err, "resource validation failed", "resourceType", permission.ResourceType, "actions", permission.Actions)
			return nil, status.Error(codes.InvalidArgument, "resource validation failed for permissions")
		}
	}

	// validate usersId has access to cloud account and member systemrole
	for _, userId := range cloudAccountRole.Users {
		roles := cs.casbinEngine.GetRolesForUserInDomain(userId, cloudAccountRole.CloudAccountId)
		if isMember := IsMember(ctx, roles); !isMember {
			logger.Error(errors.New("provided user is not a member of the cloud account"), "provided user is not a member of the cloud account")
			return nil, status.Error(codes.InvalidArgument, "provided user is not a member of the cloud account")
		}
	}

	cloudAccountRole, err := cs.permissionEngine.CreateCloudAccountRole(ctx, cloudAccountRole)
	if err != nil {
		logger.Error(err, "failed to create cloud account role")
		if s, ok := status.FromError(err); ok && s.Code() == codes.AlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "failed to create role alias already exists")
		} else {
			return nil, status.Error(codes.Internal, "failed to create role")
		}
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRole.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountRole.Id},
			eventType:           "create",
			additionalInfo:      structs.Map(cloudAccountRole),
		})
	return cloudAccountRole, nil
}

func (cs *AuthorizationService) QueryCloudAccountRoles(ctx context.Context, cloudAccountRoleQuery *pb.CloudAccountRoleQuery) (*pb.CloudAccountRoles, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.QueryCloudAccountRoles").WithValues("cloudAccountId", cloudAccountRoleQuery.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRoleQuery.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleQuery")
		return nil, status.Error(codes.InvalidArgument, "invalid query parameters")
	}

	subject, enterpriseId, groups, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Error(err, "failed to extract token info from context")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if subject == nil || enterpriseId == nil {
		logger.Error(err, "subject or enterpriseId not provided in token")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	logger = logger.WithValues("enterpriseId", enterpriseId)

	userSystemRoles := cs.casbinEngine.GetRolesForUserInDomain(*subject, cloudAccountRoleQuery.CloudAccountId)

	isAdmin := IsIntelAdmin(ctx, *subject, groups, userSystemRoles)

	if cloudAccountRoleQuery.CloudAccountId == "*" && !isAdmin {
		errMsg := "wildcard cloudAccountId is not allowed for non-admin users"
		logger.Error(errors.New(errMsg), errMsg)
		return nil, status.Errorf(codes.InvalidArgument, errMsg)
	}

	cloudAccountRoles, err := cs.permissionEngine.ListCloudAccountRoles(ctx, cloudAccountRoleQuery)
	if err != nil {
		logger.Error(err, "failed to list cloud account roles")
		return nil, status.Error(codes.Internal, "failed to retrieve cloud account roles")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId: cloudAccountRoleQuery.CloudAccountId,
			eventType:      "query",
			additionalInfo: structs.Map(cloudAccountRoleQuery),
		})
	return &pb.CloudAccountRoles{CloudAccountRoles: cloudAccountRoles}, nil
}

func (cs *AuthorizationService) GetCloudAccountRole(ctx context.Context, cloudAccountRoleId *pb.CloudAccountRoleId) (*pb.CloudAccountRole, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.GetCloudAccountRole").WithValues("cloudAccountId", cloudAccountRoleId.CloudAccountId, "cloudAccountRoleId", cloudAccountRoleId.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRoleId.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleId")
		return nil, status.Error(codes.InvalidArgument, "invalid role id provided")
	}
	cloudAccountRole, err := cs.permissionEngine.GetCloudAccountRole(ctx, cloudAccountRoleId)
	if err != nil {
		logger.Error(err, "failed to get cloudAccountRole", "roleId", cloudAccountRoleId.Id)
		return nil, status.Error(codes.Internal, "failed to retrieve role information")
	}

	permissions, err := cs.permissionEngine.ListPermissions(ctx, cloudAccountRole.Id)
	if err != nil {
		logger.Error(err, "failed to get permissions", "roleId", cloudAccountRoleId.Id)
		return nil, status.Error(codes.Internal, "failed to retrieve role information")
	}
	cloudAccountRole.Permissions = permissions

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRoleId.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountRoleId.Id},
			eventType:           "get",
			additionalInfo:      structs.Map(cloudAccountRoleId),
		})
	return cloudAccountRole, nil
}

func (cs *AuthorizationService) UpdateCloudAccountRole(ctx context.Context, cloudAccountRole *pb.CloudAccountRoleUpdate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.UpdateCloudAccountRole").WithValues("cloudAccountId", cloudAccountRole.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRole.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleUpdate")
		return nil, status.Error(codes.InvalidArgument, "invalid role update data provided")
	}

	// validate usersId has access to cloud account and member systemrole
	for _, userId := range cloudAccountRole.Users {
		roles := cs.casbinEngine.GetRolesForUserInDomain(userId, cloudAccountRole.CloudAccountId)
		if isMember := IsMember(ctx, roles); !isMember {
			logger.Error(errors.New("provided user is not a member of the cloud account"), "provided user is not a member of the cloud account")
			return nil, status.Error(codes.InvalidArgument, "provided user is not a member of the cloud account")
		}
	}

	logger.Info("permissions is nil", "permissions", cloudAccountRole.Permissions)

	err := cs.permissionEngine.UpdateCloudAccountRole(ctx, cloudAccountRole.Id, cloudAccountRole.CloudAccountId,
		cloudAccountRole.Alias, cloudAccountRole.Effect.String(), cloudAccountRole.Users, cloudAccountRole.Permissions)
	if err != nil {
		logger.Error(err, "failed to update cloudAccountRole", "roleId", cloudAccountRole.Id)
		return nil, status.Error(codes.Internal, "failed to update role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRole.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountRole.Id},
			eventType:           "update",
			additionalInfo:      structs.Map(cloudAccountRole),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) RemoveCloudAccountRole(ctx context.Context, cloudAccountRoleId *pb.CloudAccountRoleId) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemoveCloudAccountRole").WithValues("cloudAccountId", cloudAccountRoleId.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRoleId.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleId")
		return nil, status.Error(codes.InvalidArgument, "invalid role id provided")

	}
	// todo verify cloudAccount

	err := cs.permissionEngine.RemoveCloudAccountRole(ctx, cloudAccountRoleId.Id, cloudAccountRoleId.CloudAccountId)
	if err != nil {
		logger.Error(err, "failed to remove cloudAccountRole", "roleId", cloudAccountRoleId.Id)
		return nil, status.Error(codes.Internal, "failed to remove role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRoleId.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountRoleId.Id},
			eventType:           "delete",
			additionalInfo:      structs.Map(cloudAccountRoleId),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) AddUserToCloudAccountRole(ctx context.Context, userRequest *pb.CloudAccountRoleUserRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.AddUserToCloudAccountRole").WithValues("cloudAccountId", userRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := userRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleUserRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid user request parameters")

	}

	// validate usersId has access to cloud account and member systemrole
	roles := cs.casbinEngine.GetRolesForUserInDomain(userRequest.UserId, userRequest.CloudAccountId)
	if isMember := IsMember(ctx, roles); !isMember {
		logger.Error(errors.New("provided user is not a member of the cloud account"), "provided user is not a member of the cloud account")
		return nil, status.Error(codes.InvalidArgument, "provided user is not a member of the cloud account")
	}

	err := cs.permissionEngine.AddUserToCloudAccountRole(ctx, userRequest.CloudAccountId, userRequest.Id, userRequest.UserId)
	if err != nil {
		logger.Error(err, "failed to add user to cloudAccountRole", "roleId", userRequest.Id)
		return nil, status.Error(codes.Internal, "failed to add user to role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      userRequest.CloudAccountId,
			cloudAccountRoleIds: []string{userRequest.Id},
			eventType:           "add_user",
			additionalInfo:      structs.Map(userRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) AddCloudAccountRolesToUser(ctx context.Context, cloudAccountRolesUserRequest *pb.CloudAccountRolesUserRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.AddCloudAccountRolesToUser").WithValues("cloudAccountId", cloudAccountRolesUserRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRolesUserRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRolesUserRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid user request parameters")

	}

	roles := cs.casbinEngine.GetRolesForUserInDomain(cloudAccountRolesUserRequest.UserId, cloudAccountRolesUserRequest.CloudAccountId)
	if isMember := IsMember(ctx, roles); !isMember {
		logger.Error(errors.New("provided user is not a member of the cloud account"), "provided user is not a member of the cloud account")
		return nil, status.Error(codes.InvalidArgument, "provided user is not a member of the cloud account")
	}

	for _, cloudAccountRoleId := range cloudAccountRolesUserRequest.CloudAccountRoleIds {
		// validate role exist
		if _, err := cs.permissionEngine.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: cloudAccountRolesUserRequest.CloudAccountId, Id: cloudAccountRoleId}); err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				return nil, status.Errorf(codes.NotFound, "failed to get cloud account role id: %v not found", cloudAccountRoleId)
			} else {
				return nil, status.Errorf(codes.Internal, "failed to get cloud account role id: %v", cloudAccountRoleId)
			}
		}
		err := cs.permissionEngine.AddUserToCloudAccountRole(ctx, cloudAccountRolesUserRequest.CloudAccountId, cloudAccountRoleId, cloudAccountRolesUserRequest.UserId)
		if err != nil {
			logger.Error(err, "failed to add user to cloudAccountRole", "cloudAccountRoleId", cloudAccountRoleId)
			return nil, status.Error(codes.Internal, "failed to add user to role")
		}
	}

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRolesUserRequest.CloudAccountId,
			cloudAccountRoleIds: cloudAccountRolesUserRequest.CloudAccountRoleIds,
			eventType:           "add_user_ca_roles",
			additionalInfo:      structs.Map(cloudAccountRolesUserRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) RemoveCloudAccountRolesFromUser(ctx context.Context, cloudAccountRolesUserRequest *pb.CloudAccountRolesUserRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemoveCloudAccountRolesFromUser").WithValues("cloudAccountId", cloudAccountRolesUserRequest.CloudAccountId, "cloudAccountRoleIds", cloudAccountRolesUserRequest.CloudAccountRoleIds).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRolesUserRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRolesUserRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid user request parameters")
	}

	roles := cs.casbinEngine.GetRolesForUserInDomain(cloudAccountRolesUserRequest.UserId, cloudAccountRolesUserRequest.CloudAccountId)
	if isMember := IsMember(ctx, roles); !isMember {
		logger.Error(errors.New("provided user is not a member of the cloud account"), "provided user is not a member of the cloud account")
		return nil, status.Error(codes.InvalidArgument, "provided user is not a member of the cloud account")
	}

	removedCloudAccountRoleIds := []string{}
	for _, cloudAccountRoleId := range cloudAccountRolesUserRequest.CloudAccountRoleIds {
		// validate role exist
		if _, err := cs.permissionEngine.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: cloudAccountRolesUserRequest.CloudAccountId, Id: cloudAccountRoleId}); err != nil {
			logger.Error(err, "failed to get cloud account role", "cloudAccountRoleId", cloudAccountRoleId)
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				return nil, status.Errorf(codes.NotFound, "failed to get cloud account role id: %v not found", cloudAccountRoleId)
			} else {
				return nil, status.Errorf(codes.Internal, "failed to get cloud account role id: %v", cloudAccountRoleId)
			}
		}
		cloudAccountRoleIds, err := cs.permissionEngine.RemoveUserFromCloudAccountRole(ctx, cloudAccountRolesUserRequest.CloudAccountId, cloudAccountRoleId, cloudAccountRolesUserRequest.UserId)
		if err != nil {
			logger.Error(err, "failed to remove user from cloudAccountRole", "cloudAccountRoleId", cloudAccountRoleId)
			return nil, status.Error(codes.Internal, "failed to remove user from role")
		}
		removedCloudAccountRoleIds = append(removedCloudAccountRoleIds, cloudAccountRoleIds...)
	}

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRolesUserRequest.CloudAccountId,
			cloudAccountRoleIds: removedCloudAccountRoleIds,
			eventType:           "remove_user_ca_roles",
			additionalInfo:      structs.Map(cloudAccountRolesUserRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) AddPermissionToCloudAccountRole(ctx context.Context, cloudAccountPermissionReq *pb.CloudAccountRolePermissionRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.AddPermissionToCloudAccountRole").WithValues("cloudAccountId", cloudAccountPermissionReq.CloudAccountId,
		"cloudAccountRoleId", cloudAccountPermissionReq.CloudAccountRoleId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountPermissionReq.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountPermissionReq")
		return nil, status.Error(codes.InvalidArgument, "invalid permission request parameters")
	}

	if cloudAccountPermissionReq.Permission == nil {
		return nil, status.Error(codes.InvalidArgument, "permission parameter is required")
	}

	err := cs.resourceRepository.validateResource(cloudAccountPermissionReq.Permission.ResourceType, cloudAccountPermissionReq.Permission.Actions...)
	if err != nil {
		logger.Error(err, "resource validation failed", "resourceType", cloudAccountPermissionReq.Permission.ResourceType, "actions", cloudAccountPermissionReq.Permission.Actions)
		return nil, status.Error(codes.InvalidArgument, "resource validation failed for permission")
	}

	// verify cloud account role exist
	cloudAccountRole, err := cs.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: cloudAccountPermissionReq.CloudAccountId, Id: cloudAccountPermissionReq.CloudAccountRoleId})
	if err != nil {
		logger.Error(err, "error getting cloud account role")
		return nil, status.Error(codes.Internal, "error getting cloud account role")
	}
	if cloudAccountRole == nil {
		logger.Error(errors.New("cloud account role not found"), "error getting cloud account role")
		return nil, status.Error(codes.NotFound, "cloud account role not found")
	}

	err = cs.permissionEngine.AddPermissionToCloudAccountRole(ctx, cloudAccountPermissionReq.CloudAccountId, cloudAccountPermissionReq.CloudAccountRoleId, cloudAccountPermissionReq.Permission)
	if err != nil {
		logger.Error(err, "failed to add permission to cloud account role")
		if s, ok := status.FromError(err); ok && s.Code() == codes.AlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "failed to add permission to cloud account role permission with cloudAccountRoleId and type already exist")
		} else {
			return nil, status.Error(codes.Internal, "failed to add permission to cloud account role")
		}
	}

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountPermissionReq.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountPermissionReq.Permission.Id},
			eventType:           "add_permission",
			additionalInfo:      structs.Map(cloudAccountPermissionReq),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) RemovePermissionFromCloudAccountRole(ctx context.Context, cloudAccountRolePermissionId *pb.CloudAccountRolePermissionId) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemovePermissionFromCloudAccountRole").WithValues("cloudAccountId", cloudAccountRolePermissionId.CloudAccountId,
		"cloudAccountRoleId", cloudAccountRolePermissionId.CloudAccountRoleId, "permissionId", cloudAccountRolePermissionId.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRolePermissionId.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRolePermissionId")
		return nil, status.Error(codes.InvalidArgument, "invalid parameters")
	}

	// verify cloud account role exist
	cloudAccountRole, err := cs.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: cloudAccountRolePermissionId.CloudAccountId, Id: cloudAccountRolePermissionId.CloudAccountRoleId})
	if err != nil {
		logger.Error(err, "error getting cloud account role")
		return nil, status.Error(codes.Internal, "error getting cloud account role")
	}
	if cloudAccountRole == nil {
		logger.Error(errors.New("cloud account role not found"), "error getting cloud account role")
		return nil, status.Error(codes.NotFound, "cloud account role not found")
	}

	err = cs.permissionEngine.RemovePermissionFromCloudAccountRole(ctx, cloudAccountRolePermissionId.CloudAccountId, cloudAccountRolePermissionId.CloudAccountRoleId, cloudAccountRolePermissionId.Id)
	if err != nil {
		logger.Error(err, "failed to remove permission from cloud account role", "cloudAccountRoleId", cloudAccountRolePermissionId.CloudAccountRoleId, "permissionId", cloudAccountRolePermissionId.Id)
		return nil, status.Error(codes.Internal, "failed to remove permission from cloud account role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRolePermissionId.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountRolePermissionId.Id},
			eventType:           "remove_permission",
			additionalInfo:      structs.Map(cloudAccountRolePermissionId),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) UpdatePermissionCloudAccountRole(ctx context.Context, cloudAccountPermissionReq *pb.CloudAccountRolePermissionRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.UpdatePermissionCloudAccountRole").WithValues("cloudAccountId", cloudAccountPermissionReq.CloudAccountId,
		"cloudAccountRoleId", cloudAccountPermissionReq.CloudAccountRoleId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountPermissionReq.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRolePermissionId")
		return nil, status.Error(codes.InvalidArgument, "invalid parameters")
	}

	if cloudAccountPermissionReq.Permission == nil {
		return nil, status.Error(codes.InvalidArgument, "permission is required")
	}

	if cloudAccountPermissionReq.Permission.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "permission id is required")
	}

	logger = logger.WithValues("permissionId", cloudAccountPermissionReq.Permission.Id)

	err := cs.resourceRepository.validateResource(cloudAccountPermissionReq.Permission.ResourceType, cloudAccountPermissionReq.Permission.Actions...)
	if err != nil {
		logger.Error(err, "resource validation failed", "resourceType", cloudAccountPermissionReq.Permission.ResourceType, "actions", cloudAccountPermissionReq.Permission.Actions)
		return nil, status.Error(codes.InvalidArgument, "resource validation failed for permission")
	}

	// verify cloud account role exist
	cloudAccountRole, err := cs.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: cloudAccountPermissionReq.CloudAccountId, Id: cloudAccountPermissionReq.CloudAccountRoleId})
	if err != nil {
		logger.Error(err, "error getting cloud account role")
		return nil, status.Error(codes.Internal, "error getting cloud account role")
	}
	if cloudAccountRole == nil {
		logger.Error(errors.New("cloud account role not found"), "error getting cloud account role")
		return nil, status.Error(codes.NotFound, "cloud account role not found")
	}
	err = cs.permissionEngine.UpdatePermissionCloudAccountRole(ctx, cloudAccountPermissionReq.CloudAccountId, cloudAccountPermissionReq.CloudAccountRoleId, cloudAccountPermissionReq.Permission)
	if err != nil {
		logger.Error(err, "failed to update permission", "cloudAccountRoleId", cloudAccountPermissionReq.CloudAccountRoleId, "permissionId", cloudAccountPermissionReq.Permission.Id)
		return nil, status.Error(codes.Internal, "failed to update permission from cloud account role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountPermissionReq.CloudAccountId,
			cloudAccountRoleIds: []string{cloudAccountPermissionReq.CloudAccountRoleId},
			eventType:           "update_permission",
			additionalInfo:      structs.Map(cloudAccountPermissionReq),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) RemoveResourceFromCloudAccountRole(ctx context.Context, resourceRequest *pb.CloudAccountRoleResourceRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemoveResourceFromCloudAccountRole").WithValues("cloudAccountId", resourceRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := resourceRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleResourceRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid resource request parameters")
	}

	cloudAccountRoleIds, err := cs.permissionEngine.RemoveResourceFromCloudAccountRole(ctx, resourceRequest.CloudAccountId, resourceRequest.Id, resourceRequest.ResourceId, resourceRequest.ResourceType)
	if err != nil {
		logger.Error(err, "failed to remove resource from cloudAccountRole", "roleId", resourceRequest.Id, "resourceId", resourceRequest.ResourceId)
		return nil, status.Error(codes.Internal, "failed to remove resource from role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      resourceRequest.CloudAccountId,
			cloudAccountRoleIds: cloudAccountRoleIds,
			eventType:           "remove_resource",
			additionalInfo:      structs.Map(resourceRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) RemoveUserFromCloudAccountRole(ctx context.Context, cloudAccountRoleUserRequest *pb.CloudAccountRoleUserRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemoveUserFromCloudAccountRole").WithValues("cloudAccountId", cloudAccountRoleUserRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := cloudAccountRoleUserRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for cloudAccountRoleUserRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid user request parameters")
	}

	cloudAccountRoleIds, err := cs.permissionEngine.RemoveUserFromCloudAccountRole(ctx, cloudAccountRoleUserRequest.CloudAccountId, cloudAccountRoleUserRequest.Id, cloudAccountRoleUserRequest.UserId)
	if err != nil {
		logger.Error(err, "failed to remove user from cloudAccountRole", "cloudAccountRoleId", cloudAccountRoleUserRequest.Id)
		return nil, status.Error(codes.Internal, "failed to remove user from role")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId:      cloudAccountRoleUserRequest.CloudAccountId,
			cloudAccountRoleIds: cloudAccountRoleIds,
			eventType:           "remove_user",
			additionalInfo:      structs.Map(cloudAccountRoleUserRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) AssignSystemRole(ctx context.Context, roleRequest *pb.RoleRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.AssignSystemRole").WithValues("cloudAccountId", roleRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := roleRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for roleRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid role assignment request")
	}

	if _, ok := pb.SystemRole_value[roleRequest.SystemRole]; !ok {
		logger.Error(errors.New("error systemrole not valid"), "validation failed for roleRequest")
		return nil, status.Error(codes.InvalidArgument, "provided systemrole is not valid")
	}

	isIntelAdmin := roleRequest.SystemRole == pb.SystemRole_intel_admin.String()
	if (roleRequest.CloudAccountId == "*" && !isIntelAdmin) || (roleRequest.CloudAccountId != "*" && isIntelAdmin) {
		errMsg := ""
		if roleRequest.CloudAccountId == "*" {
			errMsg = "cannot assign non-admin role with wildcard cloud account id"
		} else {
			errMsg = "intel admin can only be assign with cloud account '*' "
		}
		logger.Error(errors.New(errMsg), errMsg)
		return nil, status.Errorf(codes.InvalidArgument, errMsg)
	}

	// verify user is not assign to the cloud account
	users, err := cs.getUsersFromCasbin(ctx, roleRequest.CloudAccountId)
	if err != nil {
		logger.Error(err, "failed to list users by cloud account from casbin engine")
		return nil, status.Errorf(codes.Internal, "failed to list users by cloud account from casbin engine: %v", err)
	}
	if Contains(users, roleRequest.Subject) {
		logger.Error(errors.New("provided user already has a role assigned in the given cloud account"), "provided user already has a role assigned in the given cloud account")
		return nil, status.Errorf(codes.AlreadyExists, "provided user already has a role assigned in the given cloud account")
	}

	systemRole := []string{roleRequest.Subject, roleRequest.SystemRole, roleRequest.CloudAccountId}
	result, err := cs.casbinEngine.AddGroupingPolicy(systemRole)
	if err != nil {
		logger.Error(err, "failed to assign system role", "systemRole", roleRequest.SystemRole)
		return nil, status.Error(codes.Internal, "failed to assign system role")
	}

	if result {
		logger.V(9).Info("system role assigned correctly")
	} else {
		logger.V(9).Info("system role was not assigned, because it already exists")
	}
	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId: roleRequest.CloudAccountId,
			eventType:      "assign_sys_role",
			additionalInfo: structs.Map(roleRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) UnassignSystemRole(ctx context.Context, roleRequest *pb.RoleRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.UnassignSystemRole").WithValues("cloudAccountId", roleRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := roleRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for roleRequest")
		return nil, status.Error(codes.InvalidArgument, "invalid role unassignment request")
	}

	systemRole := []string{roleRequest.Subject, roleRequest.SystemRole, roleRequest.CloudAccountId}
	_, err := cs.casbinEngine.RemoveGroupingPolicy(systemRole)
	if err != nil {
		logger.Error(err, "failed to unassign system role", "systemRole", roleRequest.SystemRole)
		return nil, status.Error(codes.Internal, "failed to unassign system role")
	}
	logger.V(9).Info("system role removed correctly")

	cs.auditLogging.Logging(ctx,
		LoggingParams{
			cloudAccountId: roleRequest.CloudAccountId,
			eventType:      "remove_sys_role",
			additionalInfo: structs.Map(roleRequest),
		})
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) SystemRoleExists(ctx context.Context, roleRequest *pb.RoleRequest) (*pb.SystemRoleExistResponse, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.SystemRoleExists").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := roleRequest.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	exist := cs.casbinEngine.HasGroupingPolicy(roleRequest.Subject, roleRequest.SystemRole, roleRequest.CloudAccountId)

	return &pb.SystemRoleExistResponse{Exist: exist}, nil
}

func (cs *AuthorizationService) RemovePolicy(ctx context.Context, policyRequest *pb.PolicyRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.RemovePolicy").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := policyRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for policyRequest")
		return nil, status.Errorf(codes.InvalidArgument, "validation failed for roleRequest: %v", err)
	}

	policy := []string{policyRequest.Subject, policyRequest.Object, policyRequest.Action, policyRequest.Expression}
	removed, err := cs.casbinEngine.RemovePolicy(policy)
	if err != nil {
		logger.Error(err, "failed to remove policy", "policy", policy)
		return nil, status.Errorf(codes.Internal, "failed to remove policy: %v", err)
	}
	if removed {
		logger.Info("policy removed correctly", "policy", policy)
	} else {
		logger.Info("policy was not removed. check if the policy actually exists", "policy", policy)
	}
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) CreatePolicy(ctx context.Context, policyRequest *pb.PolicyRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.CreatePolicy").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := policyRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for policyRequest")
		return nil, status.Errorf(codes.InvalidArgument, "validation failed for roleRequest: %v", err)
	}

	policy := []string{policyRequest.Subject, policyRequest.Object, policyRequest.Action, policyRequest.Expression}
	added, err := cs.casbinEngine.AddPolicy(policy)
	if err != nil {
		logger.Error(err, "failed to add policy", "policy", policy)
		return nil, status.Errorf(codes.Internal, "failed to add policy")
	}
	if added {
		logger.Info("policy added correctly", "policy", policy)
	} else {
		logger.Info("policy was not added. it might already exist", "policy", policy)
	}
	return &emptypb.Empty{}, nil
}

func (cs *AuthorizationService) ListResourceDefinition(ctx context.Context, req *emptypb.Empty) (*pb.ResourceDefinitions, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.ListResourceDefinition").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	resourceDefinitions := []*pb.ResourceDefinition{}

	if cs.resourceRepository == nil {
		logger.Error(errors.New("resource repository is required"), "resource repository is required")
		return nil, status.Errorf(codes.Internal, "failed to get resource repository")
	}

	for _, resource := range cs.resourceRepository.resources {
		resourceDefinitionActions := []*pb.ResourceDefinition_Action{}
		for _, actionResource := range resource.AllowedActions {
			resourceDefinitionActions = append(resourceDefinitionActions, &pb.ResourceDefinition_Action{
				Name:        actionResource.Name,
				Type:        actionResource.Type,
				Description: actionResource.Description,
			})
		}
		resourceDefinitions = append(resourceDefinitions, &pb.ResourceDefinition{Type: resource.Type, Description: resource.Description, Actions: resourceDefinitionActions})
	}
	return &pb.ResourceDefinitions{Resources: resourceDefinitions}, nil
}

func (cs *AuthorizationService) ListUsersByCloudAccount(ctx context.Context, request *pb.ListUsersByCloudAccountRequest) (*pb.ListUsersByCloudAccountResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.ListUsersByCloudAccount").WithValues("cloudAccountId", request.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for request")
		return nil, status.Errorf(codes.InvalidArgument, "validation failed for request: %v", err)
	}
	if request.CloudAccountId == "" {
		logger.Error(nil, "cloudAccountId is required")
		return nil, status.Errorf(codes.InvalidArgument, "cloudAccountId is required")
	}
	var users []string
	var err error
	if request.CloudAccountId == "*" {
		subject, _, groups, errg := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
		if errg != nil {
			logger.Error(errg, "failed to extract token info from context")
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		if subject == nil {
			logger.Error(errors.New("failed to get subject"), "failed to get subject")
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		users, err = cs.getUsersWithSystemWideRoles(ctx, *subject, groups)
	} else {
		users, err = cs.getUsersFromCasbin(ctx, request.CloudAccountId)
	}
	if err != nil {
		logger.Error(err, "failed to list users by cloud account from casbin engine")
		return nil, status.Errorf(codes.Internal, "failed to list users by cloud account from casbin engine: %v", err)
	}
	usersResponse := []*pb.User{}
	for _, user := range users {
		userResponse, err := cs.GetUser(ctx, &pb.GetUserRequest{CloudAccountId: request.CloudAccountId, UserId: user})
		if err != nil {
			logger.Error(err, "failed to get user roles info")
			return nil, status.Errorf(codes.Internal, "failed to get user roles info: %v", err)
		}

		userCloudAccountRoles := []*pb.User_CloudAccountRole{}
		for _, userCloudAccountRole := range userResponse.CloudAccountRoles {
			userCloudAccountRoles = append(userCloudAccountRoles, &pb.User_CloudAccountRole{
				CloudAccountRoleId: userCloudAccountRole.CloudAccountRoleId,
				Alias:              userCloudAccountRole.Alias,
			})
		}

		usersResponse = append(usersResponse, &pb.User{
			Id:                userResponse.Id,
			CloudAccountId:    userResponse.CloudAccountId,
			SystemRoles:       userResponse.SystemRoles,
			CloudAccountRoles: userCloudAccountRoles,
		})
	}

	return &pb.ListUsersByCloudAccountResponse{Users: usersResponse}, nil
}

func (cs *AuthorizationService) DefaultCloudAccountRoleAssigned(ctx context.Context, request *pb.DefaultCloudAccountRoleAssignedRequest) (*pb.DefaultCloudAccountRoleAssignedResponse, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.DefaultCloudAccountRoleAssigned").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	query := "SELECT * FROM default_role_assignments WHERE cloud_account_id = $1"

	rows, err := db.QueryContext(ctx, query, request.CloudAccountId)
	if err != nil {
		logger.Error(err, "failed to query default_role_assigned table")
		return nil, status.Errorf(codes.Internal, "failed to query default_role_assigned table: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		return &pb.DefaultCloudAccountRoleAssignedResponse{Assigned: true}, nil
	}

	return &pb.DefaultCloudAccountRoleAssignedResponse{Assigned: false}, nil
}

func (cs *AuthorizationService) AssignDefaultCloudAccountRole(ctx context.Context, request *pb.AssignDefaultCloudAccountRoleRequest) (*pb.CloudAccountRole, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.AssignDefaultCloudAccountRole").WithValues("cloudAccountId", request.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	assigned, err := cs.DefaultCloudAccountRoleAssigned(ctx, &pb.DefaultCloudAccountRoleAssignedRequest{CloudAccountId: request.CloudAccountId})

	if err != nil {
		logger.Error(err, "failed to check if default cloud account role is assigned.")
		return nil, status.Errorf(codes.Internal, "failed to check if default cloud account role is assigned.")
	}

	if assigned.Assigned {
		logger.Info("default cloud account role already assigned")
		return &pb.CloudAccountRole{}, nil
	}

	for _, admin := range request.Admins {
		_, err = cs.casbinEngine.AddGroupingPolicy([]string{admin, "cloud_account_admin", request.CloudAccountId})
		if err != nil {
			logger.Error(err, "failed to add admin to cloud_account_admin system role.")
			return nil, status.Errorf(codes.Internal, "failed to add admin to cloud_account_admin system role.")
		}
	}

	for _, member := range request.Members {
		_, err = cs.casbinEngine.AddGroupingPolicy([]string{member, "cloud_account_member", request.CloudAccountId})
		if err != nil {
			logger.Error(err, "failed to add member to cloud_account_member system role.")
			return nil, status.Errorf(codes.Internal, "failed to add member to cloud_account_member system role.")
		}
	}

	defaultPermissions := make([]*pb.CloudAccountRole_Permission, 0)

	for _, resource := range cs.resourceRepository.resources {
		actionNames := make([]string, 0)

		for _, action := range resource.AllowedActions {
			actionNames = append(actionNames, action.Name)
		}

		defaultPermissions = append(defaultPermissions, &pb.CloudAccountRole_Permission{
			ResourceId:   "*",
			ResourceType: resource.Type,
			Actions:      actionNames,
		})
	}

	_, err = cs.permissionEngine.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{
		CloudAccountId: request.CloudAccountId,
		Alias:          "default",
		Effect:         pb.CloudAccountRole_allow,
		Users:          request.Members,
		Permissions:    defaultPermissions,
	})

	if err != nil {
		logger.Error(err, "failed to create default cloud account role")
		return nil, status.Errorf(codes.Internal, "failed to create default cloud account role.")
	}

	query := "INSERT INTO default_role_assignments (cloud_account_id, admins, members) VALUES ($1, $2, $3)"

	if request.Members == nil {
		request.Members = make([]string, 0)
	}

	_, err = db.ExecContext(ctx, query, request.CloudAccountId, request.Admins, request.Members)

	if err != nil {
		logger.Error(err, "failed to insert into default_role_assignments table")
		return nil, status.Errorf(codes.Internal, "failed to create default role asssigment record.")
	}

	return &pb.CloudAccountRole{}, nil
}

func (cs *AuthorizationService) getUsersFromCasbin(ctx context.Context, cloudAccountId string) ([]string, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.getUsersFromCasbin").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	userSet := make(map[string]struct{})
	groupingPolicies := cs.casbinEngine.GetFilteredGroupingPolicy(2, cloudAccountId)
	for _, policy := range groupingPolicies {
		if len(policy) > 1 {
			user := policy[0]
			userSet[user] = struct{}{}
		}
	}

	users := make([]string, 0, len(userSet))
	for user := range userSet {
		users = append(users, user)
	}
	logger.Info("number of users found", "count", len(users))
	return users, nil
}

func (cs *AuthorizationService) getUsersWithSystemWideRoles(ctx context.Context, email string, groups []string) ([]string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.getUsersWithSystemWideRoles").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	var systemWideAdmins []string
	adminSet := make(map[string]struct{})
	groupingPolicies := cs.casbinEngine.GetFilteredGroupingPolicy(2, "*")

	for _, policy := range groupingPolicies {
		user := policy[0]
		roles := []string{policy[1]}

		if IsAdmin(ctx, email, groups, roles) {
			if _, exists := adminSet[user]; !exists {
				systemWideAdmins = append(systemWideAdmins, user)
				adminSet[user] = struct{}{}
			}
		}
	}
	logger.Info("retrieved users with system-wide admin roles", "admins", systemWideAdmins)
	return systemWideAdmins, nil
}

func (cs *AuthorizationService) getSystemRolesbyUser(ctx context.Context, user string, cloudAccountId string) []string {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.getSystemRolesbyUser").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	systemRoles := []string{}
	domains := []string{cloudAccountId}
	if cloudAccountId != "*" {
		domains = append(domains, "*")
	}
	for _, domain := range domains {
		roles := cs.casbinEngine.GetRolesForUserInDomain(user, domain)
		if len(roles) > 0 {
			systemRoles = append(systemRoles, roles...)
		}
	}
	logger.V(9).Info("retrieved system roles", "systemRoles", systemRoles)
	return systemRoles
}

func (cs *AuthorizationService) GetUser(ctx context.Context, userRolesRequest *pb.GetUserRequest) (*pb.UserDetailed, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuthorizationService.GetUser").WithValues("cloudAccountId", userRolesRequest.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := userRolesRequest.Validate(); err != nil {
		logger.Error(err, "validation failed for userRolesRequest")
		return nil, status.Errorf(codes.InvalidArgument, "validation failed for userRolesRequest")
	}
	systemRoles, err := cs.casbinEngine.GetRolesForUser(userRolesRequest.UserId, userRolesRequest.CloudAccountId)

	if err != nil {
		logger.Error(err, "failed to get roles for user")
		return nil, status.Errorf(codes.InvalidArgument, "failed to get roles for user")
	}

	cloudAccountRoles, err := cs.permissionEngine.ListCloudAccountRoles(ctx, &pb.CloudAccountRoleQuery{
		CloudAccountId: userRolesRequest.CloudAccountId,
		UserId:         &userRolesRequest.UserId,
	})

	if err != nil {
		logger.Error(err, "failed to list cloud account roles")
		return nil, status.Error(codes.Internal, "failed to retrieve cloud account roles")
	}

	userCloudAccountRoles := []*pb.UserDetailed_CloudAccountRoleDetailed{}

	for _, cloudAccountRole := range cloudAccountRoles {
		usercloudAccountRole := cs.fillUserCloudAccountRoleWithDetails(cloudAccountRole)
		userCloudAccountRoles = append(userCloudAccountRoles, usercloudAccountRole)
	}

	resp := &pb.UserDetailed{
		Id:                userRolesRequest.UserId,
		CloudAccountId:    userRolesRequest.CloudAccountId,
		SystemRoles:       systemRoles,
		CloudAccountRoles: userCloudAccountRoles,
	}

	logger.V(9).Info("retrieved roles for user", "systemRoles", systemRoles)

	return resp, nil
}

func (cs *AuthorizationService) fillUserCloudAccountRoleWithDetails(cloudAccountRole *pb.CloudAccountRole) *pb.UserDetailed_CloudAccountRoleDetailed {
	effect := cloudAccountRole.Effect.String()
	return &pb.UserDetailed_CloudAccountRoleDetailed{
		CloudAccountRoleId: cloudAccountRole.Id,
		Alias:              cloudAccountRole.Alias,
		CreatedAt:          cloudAccountRole.CreatedAt,
		DeletedAt:          cloudAccountRole.DeletedAt,
		UpdatedAt:          cloudAccountRole.UpdatedAt,
		CloudAccountId:     &cloudAccountRole.CloudAccountId,
		Effect:             &effect,
		Permissions:        cloudAccountRole.Permissions,
	}
}

func IsAdmin(ctx context.Context, subject string, groups []string, systemRoles []string) bool {
	if IsIntelAdmin(ctx, subject, groups, systemRoles) {
		return true
	}
	return Contains(systemRoles, pb.SystemRole_cloud_account_admin.String())
}

func IsMember(ctx context.Context, systemRoles []string) bool {
	return Contains(systemRoles, pb.SystemRole_cloud_account_member.String())
}

func IsIntelAdmin(ctx context.Context, subject string, groups []string, systemRoles []string) bool {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IsIntelAdmin").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for _, group := range groups {
		if group == "IDC.Admin" && strings.HasSuffix(subject, "@intel.com") {
			return true
		}
	}

	// step 2 check for the systemroles passed by casbin
	return Contains(systemRoles, pb.SystemRole_intel_admin.String())
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
