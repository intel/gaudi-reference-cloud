// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultUser = backend.User{
	ID:   "userId",
	Name: "username",
	Role: backend.Admin,
}

func TestBackend_CreateUser(t *testing.T) {
	opts := backend.CreateUserOpts{
		Name:        "username",
		NamespaceID: "nsId",
		Password:    "password",
		Role:        backend.Admin,
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts backend.CreateUserOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create user admin",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultUser,
			assertion: assert.NoError,
		},
		{
			name: "create user regular",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.CreateUserWithResponse.regular", true),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "nsId", Password: "password", Role: backend.Regular,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			want: &backend.User{
				ID:   "userId",
				Name: "username",
				Role: backend.Regular,
			},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "nsId", Password: "password", Role: backend.Admin,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateUserWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateUserWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateUser(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteUser(t *testing.T) {
	opts := backend.DeleteUserOpts{
		NamespaceID: "nsId",
		UserID:      "userId",
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts backend.DeleteUserOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete user",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.DeleteUserOpts{UserID: "userId", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteUserWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteUserWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteUser(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetUser(t *testing.T) {
	opts := backend.GetUserOpts{
		NamespaceID: "nsId",
		UserID:      "userId",
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts backend.GetUserOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get user",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultUser,
			assertion: assert.NoError,
		},
		{
			name: "wrong ns",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{
					NamespaceID: "nsId1",
					UserID:      "userId",
					AuthCreds:   correctCreds,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "failed login",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "userId1", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "error"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong id",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "userId1", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "userId", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(mockBackendOpts{})
			got, err := b.GetUser(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListUsers(t *testing.T) {
	opts := backend.ListUsersOpts{
		NamespaceID: "nsId",
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts backend.ListUsersOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      []*backend.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list users",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: []*backend.User{
				&defaultUser,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter users",
			args: args{
				ctx: context.Background(),
				opts: backend.ListUsersOpts{
					NamespaceID: "nsId",
					AuthCreds:   correctCreds,
					Names:       []string{"username"},
				},
			},
			want: []*backend.User{
				&defaultUser,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter out users",
			args: args{
				ctx: context.Background(),
				opts: backend.ListUsersOpts{
					NamespaceID: "nsId",
					AuthCreds:   correctCreds,
					Names:       []string{"username1"},
				},
			},
			want:      []*backend.User{},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.ListUsersOpts{NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetUsersWithResponse.incomplete", true),
				opts: opts,
			},
			want:      []*backend.User{},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListUsers(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateUser(t *testing.T) {
	opts := backend.UpdateUserOpts{
		NamespaceID: "nsId",
		UserID:      "userId",
		Role:        backend.Admin,
		AuthCreds:   correctCreds,
	}

	type args struct {
		ctx  context.Context
		opts backend.UpdateUserOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.User
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update user admin",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultUser,
			assertion: assert.NoError,
		},
		{
			name: "update use regularr",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.UpdateUserWithResponse.regular", true),
				opts: backend.UpdateUserOpts{
					NamespaceID: "nsId",
					UserID:      "userId",
					Role:        backend.Regular,
					AuthCreds:   correctCreds,
				},
			},
			want: &backend.User{
				ID:   "userId",
				Name: "username",
				Role: backend.Regular,
			},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateUserOpts{NamespaceID: "nsId", UserID: "userId", Role: backend.Admin,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateUserWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateUserWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateUserWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.UpdateUser(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateUserPassword(t *testing.T) {
	opts := backend.UpdateUserPasswordOpts{
		NamespaceID: "nsId",
		UserID:      "userId",
		Password:    "password",
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts backend.UpdateUserPasswordOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update password",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateUserPasswordOpts{NamespaceID: "nsId", UserID: "userId", Password: "password",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetUserPasswordWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetUserPasswordWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.UpdateUserPassword(tt.args.ctx, tt.args.opts))
		})
	}
}
