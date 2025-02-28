// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"
	"errors"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	vast "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/vast"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	vast_backend "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast"

	"github.com/stretchr/testify/assert"
)

type MockFilesystemBackend struct {
	backend.Interface
}

var view = vast_backend.View{
	ID:             "viewId",
	Name:           "viewName",
	Path:           "path",
	Protocols:      []vast_backend.Protocol{vast_backend.NFSV4},
	PolicyID:       5,
	AvailableBytes: 5,
	TotalBytes:     10,
}

const testClusterUuid = "00000000-0000-0000-0000-000000000000"

var fsResponse = vast.Filesystem{
	Id: &vast.FilesystemIdentifier{
		NamespaceId: &v1.NamespaceIdentifier{
			ClusterId: &v1.ClusterIdentifier{
				Uuid: testClusterUuid,
			},
			Id: "nsId",
		},
		Id: view.ID,
	},
	Name: view.Name,
	Capacity: &vast.Filesystem_Capacity{
		TotalBytes:     view.TotalBytes,
		AvailableBytes: view.AvailableBytes,
	},
	Path:      view.Path,
	Protocols: []vast.Filesystem_Protocol{vast.Filesystem_PROTOCOL_NFS_V4},
}

func (*MockFilesystemBackend) CreateView(ctx context.Context, opts vast_backend.CreateViewOpts) (*vast_backend.View, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	return &view, nil
}

func (*MockFilesystemBackend) DeleteView(ctx context.Context, opts vast_backend.DeleteViewOpts) error {
	if opts.ViewID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockFilesystemBackend) GetView(ctx context.Context, opts vast_backend.GetViewOpts) (*vast_backend.View, error) {
	if opts.ViewID == "error3" {
		return nil, errors.New("something wrong")
	}

	return &view, nil
}

func (*MockFilesystemBackend) ListViews(ctx context.Context, opts vast_backend.ListViewsOpts) ([]*vast_backend.View, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	return []*vast_backend.View{&view}, nil
}

func (*MockFilesystemBackend) UpdateView(ctx context.Context, opts vast_backend.UpdateViewOpts) (*vast_backend.View, error) {
	if opts.ViewID == "error5" {
		return nil, errors.New("something wrong")
	}

	return &view, nil
}

func TestFilesystemHandler_CreateFilesystem(t *testing.T) {
	createReq := vast.CreateFilesystemRequest{
		NamespaceId: fsResponse.GetId().GetNamespaceId(),
		Name:        "viewName",
		Path:        view.Path,
		TotalBytes:  view.TotalBytes,
		Protocols:   []vast.Filesystem_Protocol{vast.Filesystem_PROTOCOL_NFS_V4},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *vast.CreateFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *vast.CreateFilesystemResponse
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
			want: &vast.CreateFilesystemResponse{
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
				r: &vast.CreateFilesystemRequest{
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
				r:   &vast.CreateFilesystemRequest{},
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
	deleteReq := vast.DeleteFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *vast.DeleteFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *vast.DeleteFilesystemResponse
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
			want:      &vast.DeleteFilesystemResponse{},
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
				r: &vast.DeleteFilesystemRequest{
					FilesystemId: &vast.FilesystemIdentifier{
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
				r:   &vast.DeleteFilesystemRequest{},
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
	getReq := vast.GetFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *vast.GetFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *vast.GetFilesystemResponse
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
			want: &vast.GetFilesystemResponse{
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
				r: &vast.GetFilesystemRequest{
					FilesystemId: &vast.FilesystemIdentifier{
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
				r:   &vast.GetFilesystemRequest{},
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
		r   *vast.ListFilesystemsRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *vast.ListFilesystemsResponse
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
				r: &vast.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
				},
			},
			want: &vast.ListFilesystemsResponse{
				Filesystems: []*vast.Filesystem{&fsResponse},
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
				r: &vast.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
					Filter: &vast.ListFilesystemsRequest_Filter{
						Names: []string{
							"fsName",
						},
					},
				},
			},
			want: &vast.ListFilesystemsResponse{
				Filesystems: []*vast.Filesystem{&fsResponse},
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
				r: &vast.ListFilesystemsRequest{
					NamespaceId: fsResponse.GetId().GetNamespaceId(),
					Filter: &vast.ListFilesystemsRequest_Filter{
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
				r: &vast.ListFilesystemsRequest{
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
				r:   &vast.ListFilesystemsRequest{},
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
	updateReq := vast.UpdateFilesystemRequest{
		FilesystemId: fsResponse.GetId(),
		NewName:      &newName,
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *vast.UpdateFilesystemRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *vast.UpdateFilesystemResponse
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
			want: &vast.UpdateFilesystemResponse{
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
				r: &vast.UpdateFilesystemRequest{
					FilesystemId: &vast.FilesystemIdentifier{
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
				r:   &vast.UpdateFilesystemRequest{},
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
