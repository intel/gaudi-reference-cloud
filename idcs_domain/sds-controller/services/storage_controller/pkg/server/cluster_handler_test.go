// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/stretchr/testify/assert"
)

type MockClusterBackend struct {
	backend.Interface
}

type ErrorMockClusterBackend struct {
	backend.Interface
}

func (*MockClusterBackend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	return &backend.ClusterStatus{
		AvailableBytes:      50000,
		TotalBytes:          10000000,
		NamespacesLimit:     256,
		NamespacesAvailable: 200,
		HealthStatus:        backend.Healthy,
		Labels: map[string]string{
			"statusLabel": "value",
		},
	}, nil
}

func (*ErrorMockClusterBackend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	return nil, errors.New("something wrong")
}

var cluster = conf.Cluster{
	Name: "test",
	UUID: testClusterUuid,
	Type: conf.Weka,
	SupportsAPI: []conf.SupportsAPI{
		conf.WekaFilesystem,
		conf.ObjectStore,
	},
	Location: "testing",
	Labels: map[string]string{
		"key": "value",
	},
}

var clusterResponse = v1.Cluster{
	Id: &v1.ClusterIdentifier{
		Uuid: testClusterUuid,
	},
	Name: "test",
	Type: v1.Cluster_TYPE_WEKA,
	Capacity: &v1.Cluster_Capacity{
		Storage: &v1.Cluster_Capacity_Storage{
			AvailableBytes: 50000,
			TotalBytes:     10000000,
		},
		Namespaces: &v1.Cluster_Capacity_Namespaces{
			TotalCount:     256,
			AvailableCount: 200,
		},
	},
	SupportsApi: []v1.Cluster_ApiType{
		v1.Cluster_API_TYPE_WEKA_FILESYSTEM,
		v1.Cluster_API_TYPE_OBJECT_STORE,
	},
	Location: "testing",
	Labels: map[string]string{
		"key":         "value",
		"statusLabel": "value",
	},
	Health: &v1.Cluster_Health{
		Status: v1.Cluster_Health_STATUS_HEALTHY,
	},
}

func TestClusterHandler_GetCluster(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
		Clusters []*conf.Cluster
	}
	type args struct {
		ctx context.Context
		r   *v1.GetClusterRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.GetClusterResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get cluster",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetClusterRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want: &v1.GetClusterResponse{
				Cluster: &clusterResponse,
			},
			assertion: assert.NoError,
		},
		{
			name: "errors",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &ErrorMockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetClusterRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "error on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &ErrorMockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.GetClusterRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: "00000000-0000-0000-0000-000000000001",
					},
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ClusterHandler{
				Backends: tt.fields.Backends,
				Clusters: tt.fields.Clusters,
			}
			got, err := h.GetCluster(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClusterHandler_ListClusters(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
		Clusters []*conf.Cluster
	}
	type args struct {
		ctx context.Context
		r   *v1.ListClustersRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.ListClustersResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list all",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListClustersRequest{},
			},
			want: &v1.ListClustersResponse{
				Clusters: []*v1.Cluster{
					&clusterResponse,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "filter in",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
					{
						Name:     "test2",
						UUID:     "0",
						Type:     conf.Weka,
						Location: "testing",
						Labels:   map[string]string{"key": "value1"},
					},
					{
						Name:     "test3",
						UUID:     "0",
						Type:     conf.Weka,
						Location: "testing",
						Labels:   map[string]string{"key": "value2"},
					},
					{
						Name:     "test",
						UUID:     "0",
						Type:     conf.Weka,
						Location: "testing1",
						Labels:   map[string]string{"key": "value3"},
					},
					{
						Name:     "test",
						UUID:     "0",
						Type:     conf.Weka,
						Location: "testing",
						Labels:   map[string]string{"key": "value4"},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListClustersRequest{
					Filter: &v1.ListClustersRequest_Filter{
						Names:     []string{"test"},
						Locations: []string{"testing"},
						Types:     []v1.Cluster_Type{v1.Cluster_TYPE_WEKA},
						Labels:    map[string]string{"key": "value"},
					},
				},
			},
			want: &v1.ListClustersResponse{
				Clusters: []*v1.Cluster{
					&clusterResponse,
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "filter out",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListClustersRequest{
					Filter: &v1.ListClustersRequest_Filter{
						Names:     []string{"test1"},
						Locations: []string{"testing"},
						Types:     []v1.Cluster_Type{v1.Cluster_TYPE_WEKA},
						Labels:    map[string]string{"key": "value"},
					},
				},
			},
			want:      &v1.ListClustersResponse{Clusters: make([]*v1.Cluster, 0)},
			assertion: assert.NoError,
		},
		{
			name: "filter out type",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListClustersRequest{
					Filter: &v1.ListClustersRequest_Filter{
						Names:     []string{"test"},
						Locations: []string{"testing"},
						Types:     []v1.Cluster_Type{v1.Cluster_TYPE_UNSPECIFIED},
						Labels:    map[string]string{"key": "value"},
					},
				},
			},
			want:      &v1.ListClustersResponse{Clusters: make([]*v1.Cluster, 0)},
			assertion: assert.NoError,
		},
		{
			name: "filter in type",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					{
						Name:     "test",
						UUID:     testClusterUuid,
						Type:     "unknown",
						Location: "testing",
						Labels:   map[string]string{"key": "value"},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListClustersRequest{
					Filter: &v1.ListClustersRequest_Filter{
						Names:     []string{"test"},
						Locations: []string{"testing"},
						Types:     []v1.Cluster_Type{v1.Cluster_TYPE_UNSPECIFIED},
						Labels:    map[string]string{"key": "value"},
					},
				},
			},
			want: &v1.ListClustersResponse{
				Clusters: []*v1.Cluster{
					{
						Id: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Name: "test",
						Type: v1.Cluster_TYPE_UNSPECIFIED,
						Capacity: &v1.Cluster_Capacity{
							Storage: &v1.Cluster_Capacity_Storage{
								AvailableBytes: 50000,
								TotalBytes:     10000000,
							},
							Namespaces: &v1.Cluster_Capacity_Namespaces{
								TotalCount:     256,
								AvailableCount: 200,
							},
						},
						SupportsApi: []v1.Cluster_ApiType{},
						Location:    "testing",
						Labels: map[string]string{
							"key":         "value",
							"statusLabel": "value",
						},
						Health: &v1.Cluster_Health{
							Status: v1.Cluster_Health_STATUS_HEALTHY,
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "get minio cluster",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					{
						Name:     "test",
						UUID:     testClusterUuid,
						Type:     "MinIO",
						Location: "testing",
						Labels:   map[string]string{"key": "value"},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &v1.ListClustersRequest{
					Filter: &v1.ListClustersRequest_Filter{
						Names:     []string{"test"},
						Locations: []string{"testing"},
						Types:     []v1.Cluster_Type{v1.Cluster_TYPE_MINIO},
						Labels:    map[string]string{"key": "value"},
					},
				},
			},
			want: &v1.ListClustersResponse{
				Clusters: []*v1.Cluster{
					{
						Id: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Name: "test",
						Type: v1.Cluster_TYPE_MINIO,
						Capacity: &v1.Cluster_Capacity{
							Storage: &v1.Cluster_Capacity_Storage{
								AvailableBytes: 50000,
								TotalBytes:     10000000,
							},
							Namespaces: &v1.Cluster_Capacity_Namespaces{
								TotalCount:     256,
								AvailableCount: 200,
							},
						},
						SupportsApi: []v1.Cluster_ApiType{},
						Location:    "testing",
						Labels: map[string]string{
							"key":         "value",
							"statusLabel": "value",
						},
						Health: &v1.Cluster_Health{
							Status: v1.Cluster_Health_STATUS_HEALTHY,
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "error out",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &ErrorMockClusterBackend{},
				},
				Clusters: []*conf.Cluster{
					&cluster,
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &v1.ListClustersRequest{},
			},
			want:      &v1.ListClustersResponse{Clusters: make([]*v1.Cluster, 0)},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ClusterHandler{
				Backends: tt.fields.Backends,
				Clusters: tt.fields.Clusters,
			}
			got, err := h.ListClusters(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
