// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"errors"
	"fmt"

	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/otel/attribute"
)

type PermissionEngine struct {
	cloudAccountRoleRepository *CloudAccountRoleRepository
	resourceRepository         *ResourceRepository
	auditLogging               *AuditLogging
	config                     *config.Config
}

func NewPermissionEngine(cfg *config.Config, cloudAccountRoleRepository *CloudAccountRoleRepository, resourceRepository *ResourceRepository, auditLogging *AuditLogging) (*PermissionEngine, error) {
	if cloudAccountRoleRepository == nil {
		return nil, fmt.Errorf("cloudAccountRoleRepository is required")
	}
	if resourceRepository == nil {
		return nil, fmt.Errorf("resourceRepository is required")
	}
	if auditLogging == nil {
		return nil, fmt.Errorf("auditLogging is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	return &PermissionEngine{
		cloudAccountRoleRepository: cloudAccountRoleRepository,
		resourceRepository:         resourceRepository,
		auditLogging:               auditLogging,
		config:                     cfg,
	}, nil
}

func (permissionEngine *PermissionEngine) CheckPermissions(ctx context.Context, args ...interface{}) (interface{}, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.CheckPermissions").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	decisionId, subject, cloudAccountId, resource, resourceId, action := extractParams(args)
	span.SetAttributes(attribute.String("cloudAccountId", cloudAccountId))

	logger.V(9).Info("checkPermissions invoked with ", "decisionId", decisionId, "cloudAccountId:", cloudAccountId, "resource:", resource, "resourceId:", resourceId, "action:", action)
	result := permissionEngine.Check(ctx, decisionId, cloudAccountId, subject, resource, resourceId, action)
	return result, nil
}

func (s *PermissionEngine) Check(ctx context.Context, decisionId string, cloudAccountId string, subject string, resourceType string, resourceId string, action string) (check bool) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.Check").WithValues("cloudAccountId", cloudAccountId, "resourceType", resourceType, "resourceId", resourceId, "action", action).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	resource := s.resourceRepository.Get(resourceType)

	if resource == nil {
		logger.Error(errors.New("resource type not found"), "passed resource type on function check is not defined", "resourceType", resourceType, "action", action)
		return false
	}

	resourceAction := resource.GetAction(action)

	if resourceAction == nil {
		logger.Error(errors.New("resource action not found"), "passed resource action on function check is not defined", "resourceType", resourceType, "action", action)
		return false
	}

	if resourceId == "*" && resourceAction.Type != string(ACTION_COLLECTION_TYPE) {
		logger.Error(errors.New("resourceId * is allow only with actionType collection"), "resourceId * is allow only with actionType collection", "resourceType", resourceType, "action", action)
		return false
	}

	if resourceAction.Type == string(ACTION_COLLECTION_TYPE) {
		resourceId = "*"
	}

	cloudAccountRoles, err := s.cloudAccountRoleRepository.GetCloudAccountRoles(ctx, decisionId, cloudAccountId, subject, resourceType, resourceId, action)
	if err != nil {
		return false
	}

	cloudAccountRoleIds := []string{}
	for _, cloudAccountRole := range cloudAccountRoles {
		cloudAccountRoleIds = append(cloudAccountRoleIds, cloudAccountRole.Id)
	}

	allowed := len(cloudAccountRoles) > 0
	s.auditLogging.Logging(ctx, LoggingParams{
		id:                  decisionId,
		cloudAccountId:      cloudAccountId,
		eventType:           "check",
		cloudAccountRoleIds: cloudAccountRoleIds,
		additionalInfo: map[string]interface{}{
			"CloudAccountId": cloudAccountId,
			"Subject":        subject,
			"ResourceType":   resourceType,
			"ResourceId":     resourceId,
			"Action":         action,
			"Allowed":        allowed,
		},
	})
	return allowed
}

func (s *PermissionEngine) CreateCloudAccountRole(ctx context.Context, cloudAccountRole *pb.CloudAccountRole) (*pb.CloudAccountRole, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.CreateCloudAccountRole").WithValues("cloudAccountId", cloudAccountRole.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	amountCloudAccountRole, err := s.cloudAccountRoleRepository.CountCloudAccountRole(ctx, cloudAccountRole.CloudAccountId)

	if err != nil {
		return nil, err
	}

	if amountCloudAccountRole >= uint32(s.config.Limits.MaxCloudAccountRoles) {
		return nil, fmt.Errorf("maximum amount of cloud account roles reached")
	}

	if len(cloudAccountRole.Permissions) >= s.config.Limits.MaxPermissions {
		return nil, fmt.Errorf("maximum amount of permissions reached for cloud account role")
	}

	return s.cloudAccountRoleRepository.CreateCloudAccountRole(ctx, cloudAccountRole)
}

func (s *PermissionEngine) ListCloudAccountRoles(ctx context.Context, cloudAccountRoleQuery *pb.CloudAccountRoleQuery) (cloudAccountRoles []*pb.CloudAccountRole, err error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.ListCloudAccountRoles").WithValues("cloudAccountId", cloudAccountRoleQuery.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var defaultSize uint32 = 0 // Default page size
	var defaultPage uint32 = 1 // Default page number

	size := defaultSize
	if cloudAccountRoleQuery.Size != nil {
		size = *cloudAccountRoleQuery.Size
	}

	page := defaultPage
	if cloudAccountRoleQuery.Page != nil {
		page = *cloudAccountRoleQuery.Page
	}

	offset := (page - 1) * size

	return s.cloudAccountRoleRepository.ListCloudAccountRoles(ctx, cloudAccountRoleQuery, size, offset)
}

func (s *PermissionEngine) GetCloudAccountRole(ctx context.Context, cloudAccountRoleId *pb.CloudAccountRoleId) (*pb.CloudAccountRole, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.GetCloudAccountRole").WithValues("cloudAccountId", cloudAccountRoleId.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.GetCloudAccountRole(ctx, cloudAccountRoleId)
}

func (s *PermissionEngine) LookupPermissionAllowedResources(ctx context.Context, cloudAccountId string, subject string, resourceType string, resourcesId []string, action string) (allowedResourceIds []string, err error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.LookupPermission").WithValues("cloudAccountId", cloudAccountId, "resourceType", resourceType, "resourcesId", resourcesId, "action", action).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	allowedPermissionsMap, err := s.cloudAccountRoleRepository.LookupPermission(ctx, cloudAccountId, subject, resourceType, resourcesId, action, "resources")
	if err != nil {
		return nil, err
	}

	// Check for deny all first, if present return empty
	if effect, ok := allowedPermissionsMap["*"]; ok && effect == pb.CloudAccountRole_deny.Enum().String() {
		return []string{}, nil
	}

	for _, resourceId := range resourcesId {
		// Resource is denied will not be in allowedResourceIds
		if effect, ok := allowedPermissionsMap[resourceId]; ok && effect == pb.CloudAccountRole_deny.Enum().String() {
			continue
		}
		// Resource is allowed so will be in allowedResourceIds
		if effect, ok := allowedPermissionsMap[resourceId]; ok && effect == pb.CloudAccountRole_allow.Enum().String() {
			allowedResourceIds = append(allowedResourceIds, resourceId)
			continue
		}
		// Check for allow all, if present we add the resources
		if effect, ok := allowedPermissionsMap["*"]; ok && effect == pb.CloudAccountRole_allow.Enum().String() {
			allowedResourceIds = append(allowedResourceIds, resourceId)
			continue
		}
	}

	return allowedResourceIds, nil
}

func (s *PermissionEngine) LookupPermissionAllowedActions(ctx context.Context, cloudAccountId string, subject string, resourceType string, resourcesId []string, action string) (allowedActions []string, err error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.LookupPermission").WithValues("cloudAccountId", cloudAccountId, "resourceType", resourceType, "resourcesId", resourcesId, "action", action).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	allowedActionsMap, err := s.cloudAccountRoleRepository.LookupPermission(ctx, cloudAccountId, subject, resourceType, resourcesId, action, "actions")
	if err != nil {
		return nil, err
	}

	for action := range allowedActionsMap {
		if effect, ok := allowedActionsMap[action]; ok && effect == pb.CloudAccountRole_deny.Enum().String() {
			continue
		}
		if effect, ok := allowedActionsMap[action]; ok && effect == pb.CloudAccountRole_allow.Enum().String() {
			allowedActions = append(allowedActions, action)
		}
	}

	return allowedActions, nil
}

func (s *PermissionEngine) UpdateCloudAccountRole(ctx context.Context, cloudAccountRoleID string, cloudAccountId string, alias string, effect string, users []string, permissions []*pb.CloudAccountRoleUpdate_Permission) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.UpdateCloudAccountRole").WithValues("cloudAccountRoleID", cloudAccountRoleID, "cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if len(permissions) >= s.config.Limits.MaxPermissions {
		return fmt.Errorf("maximum amount of permissions reached for cloud account role")
	}

	return s.cloudAccountRoleRepository.UpdateCloudAccountRole(ctx, cloudAccountRoleID, cloudAccountId, alias, effect, users, permissions)
}

func (s *PermissionEngine) RemoveCloudAccountRole(ctx context.Context, cloudAccountRoleID string, cloudAccountId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.RemoveCloudAccountRole").WithValues("cloudAccountRoleID", cloudAccountRoleID, "cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.RemoveCloudAccountRole(ctx, cloudAccountRoleID, cloudAccountId)
}

func (s *PermissionEngine) AddUserToCloudAccountRole(ctx context.Context, cloudAccountId string, id string, userId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.AddUserToCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "id", id).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.AddUserToCloudAccountRole(ctx, cloudAccountId, id, userId)
}

func (s *PermissionEngine) AddPermissionToCloudAccountRole(ctx context.Context, cloudAccountId string, cloudAccountRoleId string, permission *pb.CloudAccountRole_Permission) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.AddPermissionToCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	amountPermission, err := s.cloudAccountRoleRepository.CountPermission(ctx, cloudAccountId, cloudAccountRoleId)

	if err != nil {
		return err
	}

	if amountPermission >= uint32(s.config.Limits.MaxPermissions) {
		return fmt.Errorf("maximum amount of permissions reached for cloud account role")
	}

	return s.cloudAccountRoleRepository.AddPermissionToCloudAccountRole(ctx, nil, cloudAccountId, cloudAccountRoleId, permission)
}

func (s *PermissionEngine) RemovePermissionFromCloudAccountRole(ctx context.Context, cloudAccountId string, cloudAccountRoleId, permissionId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.RemovePermissionFromCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleId, "permissionId", permissionId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.RemovePermissionFromCloudAccountRole(ctx, nil, cloudAccountId, cloudAccountRoleId, permissionId)
}

func (s *PermissionEngine) UpdatePermissionCloudAccountRole(ctx context.Context, cloudAccountId string, cloudAccountRoleId string, permission *pb.CloudAccountRole_Permission) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.UpdatePermissionCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleId, "permissionId", permission.Id).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	return s.cloudAccountRoleRepository.UpdatePermissionCloudAccountRole(ctx, nil, cloudAccountId, cloudAccountRoleId, permission)
}

func (s *PermissionEngine) RemoveResourceFromCloudAccountRole(ctx context.Context, cloudAccountId string, id string, resourceId string, resourceType string) ([]string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.RemoveResourceFromCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "id", id, "resourceId", resourceId, "resourceType", resourceType).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.RemoveResourceFromCloudAccountRole(ctx, cloudAccountId, id, resourceId, resourceType)
}

func (s *PermissionEngine) RemoveUserFromCloudAccountRole(ctx context.Context, cloudAccountId string, id string, userId string) ([]string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.RemoveUserFromCloudAccountRole").WithValues("cloudAccountId", cloudAccountId, "id", id).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.RemoveUserFromCloudAccountRole(ctx, cloudAccountId, id, userId)
}

func (s *PermissionEngine) ListPermissions(ctx context.Context, cloudAccountRoleId string) ([]*pb.CloudAccountRole_Permission, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PermissionEngine.ListPermissions").WithValues("cloudAccountRoleId", cloudAccountRoleId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	return s.cloudAccountRoleRepository.ListPermissions(ctx, cloudAccountRoleId)
}

func extractParams(args []interface{}) (string, string, string, string, string, string) {
	decisionId := args[0].(string)
	subject := args[1].(string)
	cloudAccountId := args[2].(string)
	resource := args[3].(string)
	resourceId := args[4].(string)
	action := args[5].(string)
	return decisionId, subject, cloudAccountId, resource, resourceId, action
}
