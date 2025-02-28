// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	weka "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	weka_backend "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka"

	"github.com/stretchr/testify/assert"
)

type MockFilesystemBackend struct {
	backend.Interface
}

var fs = weka_backend.Filesystem{
	ID:               "fsId",
	Name:             "fsName",
	Encrypted:        true,
	AuthRequired:     true,
	FilesystemStatus: weka_backend.Ready,
	AvailableBytes:   50000000000,
	TotalBytes:       100000000000000,
	BackendFQDN:      "localhost",
}

const testClusterUuid = "00000000-0000-0000-0000-000000000000"

var fsResponse = weka.Filesystem{
	Id: &weka.FilesystemIdentifier{
		NamespaceId: &v1.NamespaceIdentifier{
			ClusterId: &v1.ClusterIdentifier{
				Uuid: testClusterUuid,
			},
			Id: "nsId",
		},
		Id: fs.ID,
	},
	Name:         fs.Name,
	Status:       weka.Filesystem_STATUS_READY,
	IsEncrypted:  fs.Encrypted,
	AuthRequired: fs.AuthRequired,
	Capacity: &weka.Filesystem_Capacity{
		TotalBytes:     fs.TotalBytes,
		AvailableBytes: fs.AvailableBytes,
	},
	Backend: "localhost",
}

func (*MockFilesystemBackend) CreateFilesystem(ctx context.Context, opts weka_backend.CreateFilesystemOpts) (*weka_backend.Filesystem, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	return &fs, nil
}

func (*MockFilesystemBackend) DeleteFilesystem(ctx context.Context, opts weka_backend.DeleteFilesystemOpts) error {
	if opts.FilesystemID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockFilesystemBackend) GetFilesystem(ctx context.Context, opts weka_backend.GetFilesystemOpts) (*weka_backend.Filesystem, error) {
	if opts.FilesystemID == "error3" {
		return nil, errors.New("something wrong")
	}

	return &fs, nil
}

func (*MockFilesystemBackend) ListFilesystems(ctx context.Context, opts weka_backend.ListFilesystemsOpts) ([]*weka_backend.Filesystem, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	return []*weka_backend.Filesystem{&fs}, nil
}

func (*MockFilesystemBackend) UpdateFilesystem(ctx context.Context, opts weka_backend.UpdateFilesystemOpts) (*weka_backend.Filesystem, error) {
	if opts.FilesystemID == "error5" {
		return nil, errors.New("something wrong")
	}

	return &fs, nil
}

func TestFilesystemHandler_CreateFilesystem(t *testing.T) {
	createReq := weka.CreateFilesystemRequest{
		NamespaceId:  fsResponse.GetId().GetNamespaceId(),
		Name:         "fsName",
		TotalBytes:   fs.TotalBytes,
		Encrypted:    fs.Encrypted,
		AuthRequired: fs.AuthRequired,
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
		r   *weka.CreateFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.CreateFilesystemResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create filesystem",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &weka.CreateFilesystemResponse{
				Filesystem: &fsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateFilesystemRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
					Name:        "error1",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.CreateFilesystemRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FilesystemHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateFilesystem(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilesystemHandler_DeleteFilesystem(t *testing.T) {
	deleteReq := weka.DeleteFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.DeleteFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.DeleteFilesystemResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete filesystem",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &weka.DeleteFilesystemResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.DeleteFilesystemRequest{
					FilesystemId: &weka.FilesystemIdentifier{
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
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.DeleteFilesystemRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FilesystemHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteFilesystem(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilesystemHandler_GetFilesystem(t *testing.T) {
	getReq := weka.GetFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.GetFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.GetFilesystemResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get filesystem",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want: &weka.GetFilesystemResponse{
				Filesystem: &fsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.GetFilesystemRequest{
					FilesystemId: &weka.FilesystemIdentifier{
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
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.GetFilesystemRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FilesystemHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetFilesystem(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilesystemHandler_ListFilesystems(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.ListFilesystemsRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.ListFilesystemsResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list filesystems",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
				},
			},
			want: &weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{&fsResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "apply filter",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
					Filter: &weka.ListFilesystemsRequest_Filter{
						Names: []string{
							"fsName",
						},
					},
				},
			},
			want: &weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{&fsResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
					Filter: &weka.ListFilesystemsRequest_Filter{
						Names: []string{
							"error4",
						},
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.ListFilesystemsRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FilesystemHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListFilesystems(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilesystemHandler_UpdateFilesystem(t *testing.T) {
	newName := "newName"
	updateReq := weka.UpdateFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
		NewName:      &newName,
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.UpdateFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.UpdateFilesystemResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update filesystem",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want: &weka.UpdateFilesystemResponse{
				Filesystem: &fsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.UpdateFilesystemRequest{
					FilesystemId: &weka.FilesystemIdentifier{
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
			assertion: assert.Error,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockFilesystemBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.UpdateFilesystemRequest{},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FilesystemHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateFilesystem(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
