// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vast

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultUser = backend.User{
	ID:   "101",
	Name: "username",
	Role: backend.CSI,
}

func TestBackend_CreateUser(t *testing.T) {
	opts := backend.CreateUserOpts{
		Name:        "username",
		NamespaceID: "100",
		Password:    "password",
		Role:        backend.CSI,
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
				ctx: context.Background(),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "100", Password: "password", Role: backend.Admin,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "create user csi",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.CreateUserWithResponse.csi", true),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "100", Password: "password", Role: backend.CSI,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			want: &backend.User{
				ID:   "101",
				Name: "username",
				Role: backend.CSI,
			},
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx:  context.Background(),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "100", Password: "password", Role: backend.CSI},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{},
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx:  context.Background(),
				opts: backend.CreateUserOpts{Name: "username", NamespaceID: "100", Password: "password", Role: backend.CSI},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
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
		NamespaceID: "100",
		UserID:      "101",
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
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{},
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.DeleteUserOpts{UserID: "101", NamespaceID: "100",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
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
		NamespaceID: "100",
		UserID:      "101",
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
			name: "failed login",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "101", NamespaceID: "100",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "error"},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong id",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "102", NamespaceID: "100",
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
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{}},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.GetUserOpts{UserID: "101", NamespaceID: "100",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetUser(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListUsers(t *testing.T) {
	opts := backend.ListUsersOpts{
		NamespaceID: "100",
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
					NamespaceID: "100",
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
					NamespaceID: "100",
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
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.ListUsersOpts{NamespaceID: "100",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
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

// Vast UpdateUser is currently a no-op (it returns nil, nill for all calls).
// The UpdateUser tests are here just for compcompleteness.
// If we update VAST to support more than the single CSI user type, then we cna revisit these tests.
func TestBackend_UpdateUser(t *testing.T) {
	opts := backend.UpdateUserOpts{
		NamespaceID: "100",
		UserID:      "101",
		Role:        backend.CSI,
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
			name: "update use csi",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.UpdateUserWithResponse.regular", true),
				opts: backend.UpdateUserOpts{
					NamespaceID: "100",
					UserID:      "101",
					Role:        backend.CSI,
					AuthCreds:   correctCreds,
				},
			},
			want:      nil,
			assertion: assert.NoError,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.Background(),
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{},
			},
			want:      nil,
			assertion: assert.NoError,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateUserOpts{NamespaceID: "nsId", UserID: "userId", Role: backend.Admin,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
			},
			want:      nil,
			assertion: assert.NoError,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateUserWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.NoError,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateUserWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.NoError,
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
		NamespaceID: "100",
		UserID:      "101",
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
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{},
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateUserPasswordOpts{NamespaceID: "100", UserID: "101", Password: "password",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "password"}},
			},
			bOpts: mockBackendOpts{
				adminCredentials: &backend.AuthCreds{
					Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"},
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
