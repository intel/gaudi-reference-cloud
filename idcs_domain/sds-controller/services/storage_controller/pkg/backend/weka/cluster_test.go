// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

func TestBackend_GetStatus(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.ClusterStatus
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:  "get status",
			bOpts: mockBackendOpts{},
			args: args{
				ctx: context.Background(),
			},
			want: &backend.ClusterStatus{
				AvailableBytes:      10000000,
				TotalBytes:          50000000,
				NamespacesLimit:     256,
				NamespacesAvailable: 254,
				HealthStatus:        backend.Healthy,
				Labels: map[string]string{
					"wekaName": "test",
					"wekaGuid": "efed877e-0fed-4a42-a3b4-864040f19686",
				},
			},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx: context.Background(),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "cannot get cluster",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetClusterStatusWithResponse.error", true),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail on empty response",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetClusterStatusWithResponse.empty", true),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "inject unknown status",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetClusterStatusWithResponse.status", true),
			},
			want: &backend.ClusterStatus{
				AvailableBytes:      10000000,
				TotalBytes:          50000000,
				NamespacesLimit:     256,
				NamespacesAvailable: 254,
				HealthStatus:        backend.Unhealthy,
				Labels: map[string]string{
					"wekaName": "test",
					"wekaGuid": "efed877e-0fed-4a42-a3b4-864040f19686",
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "pass on namespaces fail",
			args: args{
				ctx: context.WithValue(context.Background(), "testing.GetOrganizationsWithResponse.error", true),
			},
			want: &backend.ClusterStatus{
				AvailableBytes:      10000000,
				TotalBytes:          50000000,
				NamespacesLimit:     256,
				NamespacesAvailable: 255,
				HealthStatus:        backend.Healthy,
				Labels: map[string]string{
					"wekaName": "test",
					"wekaGuid": "efed877e-0fed-4a42-a3b4-864040f19686",
				},
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetStatus(tt.args.ctx)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
