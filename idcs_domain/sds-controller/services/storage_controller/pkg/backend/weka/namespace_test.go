// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultNamespace = backend.Namespace{
	ID:         "nsId",
	Name:       "ns",
	QuotaTotal: 100000000,
}

func TestBackend_CreateNamespace(t *testing.T) {
	opts := backend.CreateNamespaceOpts{
		Name:          "ns",
		Quota:         100000000,
		AdminName:     "username",
		AdminPassword: "password",
	}
	type args struct {
		ctx  context.Context
		opts backend.CreateNamespaceOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.Namespace
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create namespace",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultNamespace,
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateOrganizationWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateOrganizationWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateOrganizationWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateNamespace(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteNamespace(t *testing.T) {
	opts := backend.DeleteNamespaceOpts{
		NamespaceID: "nsId",
	}
	type args struct {
		ctx  context.Context
		opts backend.DeleteNamespaceOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete namespace",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "delete protected",
			args: args{
				ctx: context.Background(),
				opts: backend.DeleteNamespaceOpts{
					NamespaceID: "protected",
				},
			},
			assertion: assert.Error,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteOrganizationWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteOrganizationWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteNamespace(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetNamespace(t *testing.T) {
	opts := backend.GetNamespaceOpts{
		NamespaceID: "nsId",
	}
	type args struct {
		ctx  context.Context
		opts backend.GetNamespaceOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.Namespace
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get namespace",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultNamespace,
			assertion: assert.NoError,
		},
		{
			name: "get protected",
			args: args{
				ctx: context.Background(),
				opts: backend.GetNamespaceOpts{
					NamespaceID: "protected",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetNamespace(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListNamespaces(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts backend.ListNamespacesOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      []*backend.Namespace
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list namespaces",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListNamespacesOpts{},
			},
			want: []*backend.Namespace{
				&defaultNamespace,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter namespaces",
			args: args{
				ctx: context.Background(),
				opts: backend.ListNamespacesOpts{
					Names: []string{"ns"},
				},
			},
			want: []*backend.Namespace{
				&defaultNamespace,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter out namespaces",
			args: args{
				ctx: context.Background(),
				opts: backend.ListNamespacesOpts{
					Names: []string{"ns1"},
				},
			},
			want:      []*backend.Namespace{},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: backend.ListNamespacesOpts{},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationsWithResponse.error", true),
				opts: backend.ListNamespacesOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationsWithResponse.empty", true),
				opts: backend.ListNamespacesOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetOrganizationsWithResponse.incomplete", true),
				opts: backend.ListNamespacesOpts{},
			},
			want:      []*backend.Namespace{},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListNamespaces(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateNamespace(t *testing.T) {
	opts := backend.UpdateNamespaceOpts{
		NamespaceID: "nsId",
		Quota:       5000000,
	}
	type args struct {
		ctx  context.Context
		opts backend.UpdateNamespaceOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.Namespace
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update namespace",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: &backend.Namespace{
				ID:         "nsId",
				Name:       "ns",
				QuotaTotal: 5000000,
			},
			assertion: assert.NoError,
		},
		{
			name: "update protected",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateNamespaceOpts{
					NamespaceID: "protected",
					Quota:       5000000,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "quota not set",
			args: args{
				ctx: context.Background(),
				opts: backend.UpdateNamespaceOpts{
					NamespaceID: "nsId",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetOrganizationLimitResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetOrganizationLimitResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.SetOrganizationLimitResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.UpdateNamespace(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
