// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vast

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateUser(ctx context.Context, opts backend.CreateUserOpts) (*backend.User, error) {
	if opts.Role != backend.CSI {
		return nil, fmt.Errorf("Only CSI role is supported in VAST")
	}

	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	tenantID, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		errMsg := fmt.Errorf("NamespaceID has invalid value: %s", opts.NamespaceID)
		log.Error().Ctx(ctx).Err(err).Msg(errMsg.Error())
		return nil, err
	}

	getTenantsResp, err := b.client.TenantsReadWithResponse(ctx, tenantID, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return nil, err
	}

	if getTenantsResp.StatusCode() != 200 || getTenantsResp.JSON200 == nil || getTenantsResp.JSON200.Id == nil {
		return nil, backend.ResponseAsErr("could not get the vast tenant data", getTenantsResp.StatusCode(), getTenantsResp.Body)
	}

	nc, err := intoNamespace(*getTenantsResp.JSON200, 0)
	if err != nil {
		return nil, err
	}

	rolename := fmt.Sprintf("%s_%s_csi_role", nc.Name, opts.Name)
	isDefault := false
	permissionsList := []string{"delete_logical", "view_logical", "edit_logical", "create_logical"}

	createRoleResp, err := b.client.RolesCreateWithResponse(ctx, client.Role{
		Name:            rolename,
		IsDefault:       &isDefault,
		PermissionsList: &permissionsList,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return nil, err
	}

	if createRoleResp.StatusCode() != 201 || createRoleResp.JSON201 == nil || createRoleResp.JSON201.Id == nil {
		return nil, backend.ResponseAsErr("could not create user role", createRoleResp.StatusCode(), createRoleResp.Body)
	}

	tenantIDs := []int{tenantID}

	updateRoleResp, err := b.client.RolesPartialUpdateWithResponse(ctx, *createRoleResp.JSON201.Id, client.RolesPartialUpdateJSONRequestBody{
		Name:      rolename,
		TenantIds: &tenantIDs,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return nil, err
	}

	if updateRoleResp.StatusCode() != 200 || updateRoleResp.JSON200 == nil || updateRoleResp.JSON200.Id == nil {
		// Delete Role
		var deleteRoleResp *client.RolesDeleteResponse
		deleteRoleResp, err = b.client.RolesDeleteWithResponse(ctx, *createRoleResp.JSON201.Id)
		if err != nil {
			return nil, backend.ResponseAsErr("Could not delete user role due to error in updating tenant id in the role", deleteRoleResp.StatusCode(), deleteRoleResp.Body)
		}
		return nil, backend.ResponseAsErr("could not update user role to include tenant", updateRoleResp.StatusCode(), updateRoleResp.Body)
	}

	tenantRoles := []int{*updateRoleResp.JSON200.Id}

	createManagerResp, err := b.client.ManagersCreateWithResponse(ctx, client.ManagerCreate{
		Username: opts.Name,
		Password: &opts.Password,
		Roles:    &tenantRoles,
	}, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Warn().Ctx(ctx).Err(err).Msg("Error calling Vast API with scheme error in VAST API")
		return nil, err //Todo: Fix the vast scheme issue
	}
	//Todo: Fix the vast scheme
	if createManagerResp != nil && (createManagerResp.StatusCode() != 201 || createManagerResp.JSON201 == nil) {
		// Delete Role
		deleteRoleResp, err := b.client.RolesDeleteWithResponse(ctx, *updateRoleResp.JSON200.Id, client.RequestEditorFn(token.Intercept))
		if err != nil {
			return nil, backend.ResponseAsErr("Could not delete user role due to user's vast manager configuration not created", deleteRoleResp.StatusCode(), deleteRoleResp.Body)
		}
		return nil, backend.ResponseAsErr("could not create manager configuration for the user", createManagerResp.StatusCode(), createManagerResp.Body)
	}

	return intoUser(*createManagerResp)
}

func (b *Backend) DeleteUser(ctx context.Context, opts backend.DeleteUserOpts) error {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	userID, err := strconv.Atoi(opts.UserID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return err
	}

	// Retrieve the manager
	managerResp, err := b.client.ManagersReadWithResponse(ctx, userID, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if managerResp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not retrieve user's vast manager config", managerResp.StatusCode(), managerResp.Body)
	}

	deleteManagerResp, err := b.client.ManagersDeleteWithResponse(ctx, userID, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if deleteManagerResp.StatusCode() != 204 {
		return backend.ResponseAsErr("could not delete user's vast manager config", deleteManagerResp.StatusCode(), deleteManagerResp.Body)
	}

	tenantID, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return err
	}

	getTenantsResp, err := b.client.TenantsReadWithResponse(ctx, tenantID, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if getTenantsResp.StatusCode() != 200 || getTenantsResp.JSON200 == nil || getTenantsResp.JSON200.Id == nil {
		return backend.ResponseAsErr("could not get the vast tenant data", getTenantsResp.StatusCode(), getTenantsResp.Body)
	}

	nc, err := intoNamespace(*getTenantsResp.JSON200, 0)
	if err != nil {
		return err
	}

	rolename := fmt.Sprintf("%s_%s_csi_role", nc.Name, managerResp.JSON200.Username)

	getRoleResp, err := b.client.RolesListWithResponse(ctx, &client.RolesListParams{
		Name: &rolename,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if getRoleResp.StatusCode() != 200 || getRoleResp.JSON200 == nil || len(*getRoleResp.JSON200) == 0 {
		return backend.ResponseAsErr("could not list user role", getRoleResp.StatusCode(), getRoleResp.Body)
	}

	roles := *getRoleResp.JSON200

	// We expect the number of roles to always be 1
	const expectedNumRoles int = 1
	if len(roles) != expectedNumRoles {
		return fmt.Errorf("unexpected number of roles, expected 1 but got: %s", strconv.Itoa(len(roles)))
	}

	// Access the single role's ID
	if roles[0].Id == nil {
		return fmt.Errorf("role ID is nil for the first role")
	}
	roleID := *roles[0].Id

	deleteRoleResp, err := b.client.RolesDeleteWithResponse(ctx, roleID, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if deleteRoleResp.StatusCode() != 204 {
		return backend.ResponseAsErr("could not create user role", deleteRoleResp.StatusCode(), deleteRoleResp.Body)
	}

	return nil
}

func (b *Backend) GetUser(ctx context.Context, opts backend.GetUserOpts) (*backend.User, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	userID, err := strconv.Atoi(opts.UserID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return nil, err
	}

	managerResp, err := b.client.ManagersReadWithResponse(ctx, userID, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return nil, err
	}

	if managerResp.StatusCode() != 200 {
		return nil, backend.ResponseAsErr("could not retrieve user's vast manager config", managerResp.StatusCode(), managerResp.Body)
	}

	return &backend.User{
		ID:   opts.UserID,
		Name: managerResp.JSON200.Username,
		Role: backend.CSI,
	}, nil
}

func (b *Backend) ListUsers(ctx context.Context, opts backend.ListUsersOpts) ([]*backend.User, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	tenantID, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		errMsg := fmt.Errorf("NamespaceID has invalid value: %s", opts.NamespaceID)
		log.Error().Ctx(ctx).Err(err).Msg(errMsg.Error())
		return nil, err
	}

	params := &client.ManagersListParams{}
	namesToFilter := opts.Names

	if len(namesToFilter) > 0 {
		if len(namesToFilter) == 1 {
			log.Info().Ctx(ctx).Str("name", namesToFilter[0]).Msg("Issuing ListUsers API with username param")
			params.Username = &namesToFilter[0]
		} else {
			allNames := strings.Join(namesToFilter, ",")
			log.Info().Ctx(ctx).Str("username__in", allNames).Msg("Issuing ListUsers API with username__in param")
			params.UsernameIn = &allNames
		}
	}

	managersResp, err := b.client.ManagersListWithResponse(ctx, params, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return nil, err
	}

	if managersResp.StatusCode() != 200 || managersResp.JSON200 == nil {
		return nil, backend.ResponseAsErr("Could not retrieve user's vast manager config", managersResp.StatusCode(), managersResp.Body)
	}

	results := make([]*backend.User, 0)

	// Managers -> Roles -> Tenants
	for _, manager := range *managersResp.JSON200 {
		if len(namesToFilter) == 0 || slices.Contains(namesToFilter, manager.Username) {
			if manager.Roles == nil {
				log.Error().Ctx(ctx).Err(err).Msg("Unexpected nil roles for manager")
				continue
			}

			for _, role := range *manager.Roles {
				rolesResp, err := b.client.RolesReadWithResponse(ctx, *role.Id,
					client.RequestEditorFn(token.Intercept))

				if err != nil {
					log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API RolesReadWithResponse()")
					return nil, err
				}

				if rolesResp.StatusCode() != 200 || rolesResp.JSON200 == nil {
					return nil, backend.ResponseAsErr("could not retrieve user's vast roles", rolesResp.StatusCode(), rolesResp.Body)
				}

				if rolesResp.JSON200.Tenants != nil && slices.Contains(*rolesResp.JSON200.Tenants, tenantID) {

					user := backend.User{
						ID:   strconv.Itoa(*manager.Id),
						Name: manager.Username,
						Role: backend.CSI,
					}

					results = append(results, &user)
				}
			}
		}
	}

	return results, nil
}

func (b *Backend) UpdateUserPassword(ctx context.Context, opts backend.UpdateUserPasswordOpts) error {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	userID, err := strconv.Atoi(opts.UserID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return err
	}

	managerResp, err := b.client.ManagersPartialUpdateWithResponse(ctx, userID, client.ManagersPartialUpdateJSONRequestBody{
		Password: &opts.Password,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API")
		return err
	}

	if managerResp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not update user's password", managerResp.StatusCode(), managerResp.Body)
	}

	return nil
}

// Todo: remove it
var user = backend.User{}

func (b *Backend) UpdateUser(ctx context.Context, opts backend.UpdateUserOpts) (*backend.User, error) {
	return nil, nil
}
