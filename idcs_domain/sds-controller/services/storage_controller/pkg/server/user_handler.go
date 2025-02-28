// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
)

type UserHandler struct {
	Backends map[string]backend.Interface
}

// CreateUser implements v1.UserServiceServer.
func (h *UserHandler) CreateUser(ctx context.Context, r *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}

	log.Info().Ctx(ctx).Str("name", r.GetUserName()).Msg("Creating user")

	user, err := u.CreateUser(ctx, backend.CreateUserOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Name:        r.GetUserName(),
		Password:    r.GetUserPassword(),
		Role:        backend.UserRole(r.GetRole()),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", user.ID).Msg("Created user")

	return &v1.CreateUserResponse{
		User: intoUser(r.GetNamespaceId(), user),
	}, nil
}

// DeleteUser implements v1.UserServiceServer.
func (h *UserHandler) DeleteUser(ctx context.Context, r *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	b := h.Backends[r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}
	log.Info().Ctx(ctx).Str("id", r.GetUserId().GetId()).Msg("Deleting user")

	err := u.DeleteUser(ctx, backend.DeleteUserOpts{
		NamespaceID: r.GetUserId().GetNamespaceId().GetId(),
		UserID:      r.GetUserId().GetId(),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserResponse{}, nil
}

// GetUser implements v1.UserServiceServer.
func (h *UserHandler) GetUser(ctx context.Context, r *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	b := h.Backends[r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}

	user, err := u.GetUser(ctx, backend.GetUserOpts{
		NamespaceID: r.GetUserId().GetNamespaceId().GetId(),
		UserID:      r.GetUserId().GetId(),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &v1.GetUserResponse{
		User: intoUser(r.GetUserId().GetNamespaceId(), user),
	}, nil
}

// ListUsers implements v1.UserServiceServer.
func (h *UserHandler) ListUsers(ctx context.Context, r *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}

	log.Info().Ctx(ctx).Strs("names", r.GetFilter().GetNames()).Msg("Listing users")

	usrs, err := u.ListUsers(ctx, backend.ListUsersOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Names:       r.GetFilter().GetNames(),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	users := make([]*v1.User, 0)

	for _, user := range usrs {
		users = append(users, intoUser(r.GetNamespaceId(), user))
	}

	return &v1.ListUsersResponse{
		Users: users,
	}, nil
}

// UpdateUser implements v1.UserServiceServer.
func (h *UserHandler) UpdateUser(ctx context.Context, r *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	b := h.Backends[r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}
	log.Info().Ctx(ctx).Str("id", r.GetUserId().GetId()).Str("role", r.GetRole().String()).Msg("Updating user")

	user, err := u.UpdateUser(ctx, backend.UpdateUserOpts{
		NamespaceID: r.GetUserId().GetNamespaceId().GetId(),
		UserID:      r.GetUserId().GetId(),
		Role:        backend.UserRole(r.GetRole()),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserResponse{
		User: intoUser(r.GetUserId().GetNamespaceId(), user),
	}, nil
}

// UpdateUserPassword implements v1.UserServiceServer.
func (h *UserHandler) UpdateUserPassword(ctx context.Context, r *v1.UpdateUserPasswordRequest) (*v1.UpdateUserPasswordResponse, error) {
	b := h.Backends[r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	u, ok := b.(backend.UserOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetUserId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support user operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support user operations")
	}
	log.Info().Ctx(ctx).Str("id", r.GetUserId().GetId()).Msg("Chaning user password")

	err := u.UpdateUserPassword(ctx, backend.UpdateUserPasswordOpts{
		NamespaceID: r.GetUserId().GetNamespaceId().GetId(),
		UserID:      r.GetUserId().GetId(),
		Password:    r.GetNewPassword(),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserPasswordResponse{}, nil
}

func intoUser(namespaceID *v1.NamespaceIdentifier, user *backend.User) *v1.User {
	if user == nil || namespaceID == nil {
		return nil
	}

	return &v1.User{
		Id: &v1.UserIdentifier{
			NamespaceId: namespaceID,
			Id:          user.ID,
		},
		Name: user.Name,
		Role: v1.User_Role(user.Role), // Such conversation can be dangerous if enums description change
	}
}
