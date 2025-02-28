// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"slices"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateUser(ctx context.Context, opts backend.CreateUserOpts) (*backend.User, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	var role v4.CreateUserJSONBodyRole
	switch opts.Role {
	case backend.Admin:
		role = v4.CreateUserJSONBodyRoleOrgAdmin
	default:
		role = v4.CreateUserJSONBodyRoleRegular
	}

	resp, err := b.client.CreateUserWithResponse(ctx, v4.CreateUserJSONRequestBody{
		Username: &opts.Name,
		Password: &opts.Password,
		Role:     &role,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not create user", resp.StatusCode(), resp.Body)
	}

	return intoUser(*resp.JSON200.Data)
}

func (b *Backend) DeleteUser(ctx context.Context, opts backend.DeleteUserOpts) error {
	if opts.AuthCreds == nil {
		return errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return err
	}

	resp, err := b.client.DeleteUserWithResponse(ctx, opts.UserID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete user", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) GetUser(ctx context.Context, opts backend.GetUserOpts) (*backend.User, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	resp, err := b.getUsers(ctx, make([]string, 0), token)

	if err != nil {
		return nil, err
	}

	for _, u := range resp {
		if u.ID == opts.UserID {
			return u, nil
		}
	}

	return nil, errors.New("could not find user by id")
}

func (b *Backend) ListUsers(ctx context.Context, opts backend.ListUsersOpts) ([]*backend.User, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	return b.getUsers(ctx, opts.Names, token)
}

func (b *Backend) UpdateUser(ctx context.Context, opts backend.UpdateUserOpts) (*backend.User, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	var role v4.UpdateUserJSONBodyRole
	switch opts.Role {
	case backend.Admin:
		role = v4.UpdateUserJSONBodyRoleOrgAdmin
	default:
		role = v4.UpdateUserJSONBodyRoleRegular
	}

	resp, err := b.client.UpdateUserWithResponse(ctx, opts.UserID, v4.UpdateUserJSONRequestBody{
		Role: &role,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not update user", resp.StatusCode(), resp.Body)
	}

	return intoUser(*resp.JSON200.Data)
}

func (b *Backend) UpdateUserPassword(ctx context.Context, opts backend.UpdateUserPasswordOpts) error {
	if opts.AuthCreds == nil {
		return errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return err
	}

	resp, err := b.client.SetUserPasswordWithResponse(ctx, opts.UserID, v4.SetUserPasswordJSONRequestBody{
		Password: opts.Password,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not update user password", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) getUsers(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.User, error) {
	resp, err := b.client.GetUsersWithResponse(ctx, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get namespaces", resp.StatusCode(), resp.Body)
	}

	users := make([]*backend.User, 0)

	for _, user := range *resp.JSON200.Data {
		user, err := intoUser(user)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse user")
			continue
		}
		if len(namesToFilter) > 0 && !slices.Contains(namesToFilter, user.Name) {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}
