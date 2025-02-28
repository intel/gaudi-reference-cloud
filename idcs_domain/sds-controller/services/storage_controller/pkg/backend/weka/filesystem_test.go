// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultFs = Filesystem{
	ID:               "fsId",
	Name:             "fsName",
	Encrypted:        true,
	AuthRequired:     true,
	FilesystemStatus: Ready,
	AvailableBytes:   500000,
	TotalBytes:       1000000,
	BackendFQDN:      "localhost",
}

func TestBackend_CreateFilesystem(t *testing.T) {
	opts := CreateFilesystemOpts{
		NamespaceID:  "nsId",
		Name:         "fsName",
		Quota:        1000000,
		Encrypted:    true,
		AuthRequired: true,
		AuthCreds:    correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts CreateFilesystemOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *Filesystem
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create fs",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultFs,
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
				opts: CreateFilesystemOpts{Name: "fsName", NamespaceID: "nsId", Quota: 1000000,
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateFileSystemResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateFileSystemResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.CreateFileSystemResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateFilesystem(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteFilesystem(t *testing.T) {
	opts := DeleteFilesystemOpts{
		NamespaceID:  "nsId",
		FilesystemID: "fsId",
		AuthCreds:    correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts DeleteFilesystemOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete fs",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.404", true),
				opts: opts,
			},
			assertion: assert.NoError,
		},
		{
			name: "timeout delete",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "no creds",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.404", true),
			},
			assertion: assert.Error,
		},
		{
			name: "wrong creds",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.404", true),
				opts: DeleteFilesystemOpts{FilesystemID: "fsId", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			assertion: assert.Error,
		},
		{
			name: "error get",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteFileSystemWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeleteFileSystemWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteFilesystem(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetFilesystem(t *testing.T) {
	opts := GetFilesystemOpts{
		NamespaceID:  "nsId",
		FilesystemID: "fsId",
		AuthCreds:    correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts GetFilesystemOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *Filesystem
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get fs",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultFs,
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
				opts: GetFilesystemOpts{FilesystemID: "fsId", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetFilesystem(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListFilesystems(t *testing.T) {
	opts := ListFilesystemsOpts{
		NamespaceID: "nsId",
		AuthCreds:   correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts ListFilesystemsOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      []*Filesystem
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list fs",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want: []*Filesystem{
				&defaultFs,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter fs",
			args: args{
				ctx: context.Background(),
				opts: ListFilesystemsOpts{
					NamespaceID: "nsId",
					Names:       []string{"fsName"},
					AuthCreds:   correctCreds,
				},
			},
			want: []*Filesystem{
				&defaultFs,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter out fs",
			args: args{
				ctx: context.Background(),
				opts: ListFilesystemsOpts{
					NamespaceID: "nsId",
					Names:       []string{"fsName2"},
					AuthCreds:   correctCreds,
				},
			},
			want:      []*Filesystem{},
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
				opts: ListFilesystemsOpts{NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemsWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemsWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetFileSystemsWithResponse.incomplete", true),
				opts: opts,
			},
			want:      []*Filesystem{},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListFilesystems(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_UpdateFilesystem(t *testing.T) {
	opts := UpdateFilesystemOpts{
		NamespaceID:  "nsId",
		FilesystemID: "fsId",
		AuthCreds:    correctCreds,
	}
	type args struct {
		ctx  context.Context
		opts UpdateFilesystemOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *Filesystem
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update fs",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultFs,
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
				opts: UpdateFilesystemOpts{FilesystemID: "fsId", NamespaceID: "nsId",
					AuthCreds: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateFileSystemWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateFileSystemWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.UpdateFileSystemWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.UpdateFilesystem(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
