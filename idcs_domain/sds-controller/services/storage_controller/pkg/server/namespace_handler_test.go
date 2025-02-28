// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockNamespaceBackend struct {
	backend.Interface
}

var bNs = backend.Namespace{
	ID:         "nsId",
	Name:       "ns",
	QuotaTotal: 5000000,
	IPRanges:   [][]string{{"10.0.0.1", "10.0.0.2"}},
}

var nsResponse = v1.Namespace{
	Id: &v1.NamespaceIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bNs.ID,
	},
	Name: bNs.Name,
	Quota: &v1.Namespace_Quota{
		TotalBytes: bNs.QuotaTotal,
	},
	IpFilters: []*v1.Namespace_IpFilter{
		&v1.Namespace_IpFilter{
			Start: "10.0.0.1",
			End:   "10.0.0.2",
		},
	},
}

func (*MockNamespaceBackend) CreateNamespace(ctx context.Context, opts backend.CreateNamespaceOpts) (*backend.Namespace, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	return &backend.Namespace{
		ID:         "nsId",
		Name:       opts.Name,
		QuotaTotal: opts.Quota,
		IPRanges:   [][]string{{"10.0.0.1", "10.0.0.2"}},
	}, nil
}

func (*MockNamespaceBackend) DeleteNamespace(ctx context.Context, opts backend.DeleteNamespaceOpts) error {
	if opts.NamespaceID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockNamespaceBackend) GetNamespace(ctx context.Context, opts backend.GetNamespaceOpts) (*backend.Namespace, error) {
	if opts.NamespaceID == "error3" {
		return nil, errors.New("something wrong")
	}

	return &bNs, nil
}

func (*MockNamespaceBackend) ListNamespaces(ctx context.Context, opts backend.ListNamespacesOpts) ([]*backend.Namespace, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	return []*backend.Namespace{&bNs}, nil
}

func (*MockNamespaceBackend) UpdateNamespace(ctx context.Context, opts backend.UpdateNamespaceOpts) (*backend.Namespace, error) {
	if opts.NamespaceID == "error5" {
		return nil, errors.New("something wrong")
	}

	return &bNs, nil
}

// Make sure our backend implements namespace operations
var _ backend.NamespaceOps = &MockNamespaceBackend{}

type MockNoNamespacesBackend struct {
	backend.Interface
}

func TestNamespaceHandler_CreateNamespace(t *testing.T) {
	createReq := v1.CreateNamespaceRequest{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Name: "ns",
		Quota: &v1.Namespace_Quota{
			TotalBytes: 5000000,
		},
		IpFilters: []*v1.Namespace_IpFilter{
			&v1.Namespace_IpFilter{
				Start: "10.0.0.1",
				End:   "10.0.0.2",
			},
		},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.CreateNamespaceRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateNamespaceResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create ns",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want: &v1.CreateNamespaceResponse{
				Namespace: &nsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.CreateNamespaceRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: "error1",
					Quota: &v1.Namespace_Quota{
						TotalBytes: 5000000,
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
					"00000000-0000-0000-0000-000000000001": &MockNamespaceBackend{},
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
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.CreateNamespaceRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoNamespacesBackend{},
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
			h := &NamespaceHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateNamespace(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespaceHandler_DeleteNamespace(t *testing.T) {
	deleteReq := v1.DeleteNamespaceRequest{
		NamespaceId: nsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.DeleteNamespaceRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.DeleteNamespaceResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete ns",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &v1.DeleteNamespaceResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.DeleteNamespaceRequest{
					NamespaceId: &v1.NamespaceIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
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
					"00000000-0000-0000-0000-000000000001": &MockNamespaceBackend{},
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
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.DeleteNamespaceRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoNamespacesBackend{},
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
			h := &NamespaceHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.DeleteNamespace(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespaceHandler_GetNamespace(t *testing.T) {
	getReq := v1.GetNamespaceRequest{
		NamespaceId: nsResponse.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.GetNamespaceRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.GetNamespaceResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get ns",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReq,
			},
			want: &v1.GetNamespaceResponse{
				Namespace: &nsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetNamespaceRequest{
					NamespaceId: &v1.NamespaceIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
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
					"00000000-0000-0000-0000-000000000001": &MockNamespaceBackend{},
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
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.GetNamespaceRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoNamespacesBackend{},
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
			h := &NamespaceHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetNamespace(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespaceHandler_ListNamespaces(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.ListNamespacesRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.ListNamespacesResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list ns",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListNamespacesRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want: &v1.ListNamespacesResponse{
				Namespaces: []*v1.Namespace{&nsResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "apply filter",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListNamespacesRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Filter: &v1.ListNamespacesRequest_Filter{
						Names: []string{"ns"},
					},
				},
			},
			want: &v1.ListNamespacesResponse{
				Namespaces: []*v1.Namespace{&nsResponse},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListNamespacesRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Filter: &v1.ListNamespacesRequest_Filter{
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
					"00000000-0000-0000-0000-000000000001": &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListNamespacesRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListNamespacesRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoNamespacesBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListNamespacesRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &NamespaceHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListNamespaces(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespaceHandler_UpdateNamespace(t *testing.T) {
	updateReq := v1.UpdateNamespaceRequest{
		NamespaceId: nsResponse.GetId(),
		Quota: &v1.Namespace_Quota{
			TotalBytes: 50000,
		},
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *v1.UpdateNamespaceRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.UpdateNamespaceResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "update ns",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &updateReq,
			},
			want: &v1.UpdateNamespaceResponse{
				Namespace: &nsResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.UpdateNamespaceRequest{
					NamespaceId: &v1.NamespaceIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
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
					"00000000-0000-0000-0000-000000000001": &MockNamespaceBackend{},
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
					testClusterUuid: &MockNamespaceBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.UpdateNamespaceRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported operation",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoNamespacesBackend{},
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
			h := &NamespaceHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.UpdateNamespace(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func AssertStatus(c codes.Code) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		status, _ := status.FromError(err)
		if c != status.Code() {
			return assert.Fail(t, fmt.Sprintf("Unexpected gRPC status: \n"+
				"expected: %s\n"+
				"actual  : %s", c.String(), status.Code().String()), msgAndArgs...)
		}
		return true
	}
}
