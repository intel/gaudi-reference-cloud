// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var user = backend.User{
	ID:   "userId",
	Name: "user",
	Role: backend.Admin,
}

var userResponse = v1.User{
	Id: &v1.UserIdentifier{
		NamespaceId: &v1.NamespaceIdentifier{
			ClusterId: &v1.ClusterIdentifier{
				Uuid: testClusterUuid,
			},
			Id: "nsId",
		},
		Id: user.ID,
	},
	Name: user.Name,
	Role: v1.User_ROLE_ADMIN,
}

type MockUserBackend struct {
	backend.Interface
}

func (*MockUserBackend) CreateUser(ctx context.Context, opts backend.CreateUserOpts) (*backend.User, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	return &user, nil
}

func (*MockUserBackend) DeleteUser(ctx context.Context, opts backend.DeleteUserOpts) error {
	if opts.UserID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockUserBackend) GetUser(ctx context.Context, opts backend.GetUserOpts) (*backend.User, error) {
	if opts.UserID == "error3" {
		return nil, errors.New("something wrong")
	}

	return &user, nil
}

func (*MockUserBackend) ListUsers(ctx context.Context, opts backend.ListUsersOpts) ([]*backend.User, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	return []*backend.User{&user}, nil
}

func (*MockUserBackend) UpdateUser(ctx context.Context, opts backend.UpdateUserOpts) (*backend.User, error) {
	if opts.UserID == "error5" {
		return nil, errors.New("something wrong")
	}

	return &user, nil
}

func (*MockUserBackend) UpdateUserPassword(ctx context.Context, opts backend.UpdateUserPasswordOpts) error {
	if opts.UserID == "error6" {
		return errors.New("something wrong")
	}

	return nil
}

// make sure the backend implements UserOps
var _ backend.UserOps = &MockUserBackend{}

type MockNoUsersBackend struct {
	backend.Interface
}

func TestUserHandler_CreateUser(t *testing.T) {
	createReq := v1.CreateUserRequest{
		NamespaceId:  userResponse.GetId().GetNamespaceId(),
		UserName:     "user",
		UserPassword: "password",
		Role:         v1.User_ROLE_ADMIN,
		AuthCtx: &v1.AuthenticationContext{
			Scheme: &v1.AuthenticationContext_Basic_{
				Basic: &v1.AuthenticationContext_Basic{
					Principal:   "username",
					Credentials: "secretPassword",
				},
			},
		},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.CreateUserRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateUserResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create user",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &v1.CreateUserResponse{
				User: &userResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateUserRequest{
					NamespaceId: &v1.NamespaceIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "userId",
					},
					UserName: "error1",
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.CreateUserRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateUser(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	deleteReq := v1.DeleteUserRequest{
		UserId: userResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.DeleteUserRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.DeleteUserResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete user",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &v1.DeleteUserResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.DeleteUserRequest{
					UserId: &v1.UserIdentifier{
						NamespaceId: &v1.NamespaceIdentifier{
							ClusterId: &v1.ClusterIdentifier{
								Uuid: testClusterUuid,
							},
							Id: "nsId",
						},
						Id: "error2",
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.DeleteUserRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteUser(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	getReq := v1.GetUserRequest{
		UserId: userResponse.GetId(),
	}

	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.GetUserRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.GetUserResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get user",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want: &v1.GetUserResponse{
				User: &userResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetUserRequest{
					UserId: &v1.UserIdentifier{
						NamespaceId: &v1.NamespaceIdentifier{
							ClusterId: &v1.ClusterIdentifier{
								Uuid: testClusterUuid,
							},
							Id: "nsId",
						},
						Id: "error3",
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.GetUserRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetUser(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.ListUsersRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.ListUsersResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list users",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListUsersRequest{
					NamespaceId: userResponse.GetId().GetNamespaceId(),
				},
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{&userResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "apply filter",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListUsersRequest{
					NamespaceId: userResponse.GetId().GetNamespaceId(),
					Filter: &v1.ListUsersRequest_Filter{
						Names: []string{"user"},
					},
				},
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{&userResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListUsersRequest{
					NamespaceId: userResponse.GetId().GetNamespaceId(),
					Filter: &v1.ListUsersRequest_Filter{
						Names: []string{"error4"},
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListUsersRequest{
					NamespaceId: userResponse.GetId().GetNamespaceId(),
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListUsersRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListUsersRequest{
					NamespaceId: userResponse.GetId().GetNamespaceId(),
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListUsers(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	updateReq := v1.UpdateUserRequest{
		UserId: userResponse.GetId(),
		Role:   v1.User_ROLE_REGULAR,
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateUserRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateUserResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update user",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want: &v1.UpdateUserResponse{
				User: &userResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateUserRequest{
					UserId: &v1.UserIdentifier{
						NamespaceId: &v1.NamespaceIdentifier{
							ClusterId: &v1.ClusterIdentifier{
								Uuid: testClusterUuid,
							},
							Id: "nsId",
						},
						Id: "error5",
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateUserRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateUser(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserHandler_UpdateUserPassword(t *testing.T) {
	updatePasswordReq := v1.UpdateUserPasswordRequest{
		UserId:      userResponse.GetId(),
		NewPassword: "newPassword",
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateUserPasswordRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateUserPasswordResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update user password",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updatePasswordReq,
			},
			want:      &v1.UpdateUserPasswordResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateUserPasswordRequest{
					UserId: &v1.UserIdentifier{
						NamespaceId: &v1.NamespaceIdentifier{
							ClusterId: &v1.ClusterIdentifier{
								Uuid: testClusterUuid,
							},
							Id: "nsId",
						},
						Id: "error6",
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updatePasswordReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockUserBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateUserPasswordRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoUsersBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updatePasswordReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UserHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateUserPassword(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
