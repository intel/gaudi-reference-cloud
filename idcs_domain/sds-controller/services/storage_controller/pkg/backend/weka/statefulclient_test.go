// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package weka

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultStatefulClient = backend.StatefulClient{
	ID:     "scId",
	Name:   "sc",
	Status: "UP",
	Mode:   "client",
	Cores:  1,
}

func TestBackend_CreateStatefulClient(t *testing.T) {
	opts := backend.CreateStatefulClientOpts{
		Name: "sc",
		// IP address range 192.0.2.0/24, also known as TEST-NET-1,
		// is reserved for use in documentation and example code.
		// More info RFC 5737
		Ip: "192.0.2.0",
	}
	type args struct {
		ctx  context.Context
		opts backend.CreateStatefulClientOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.StatefulClient
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create StatefulClient",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultStatefulClient,
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
				ctx:  context.WithValue(context.Background(), "testing.AddContainerWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AddContainerWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.AddContainerWithResponse.incomplete", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateStatefulClient(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_DeleteStatefulClient(t *testing.T) {
	opts := backend.DeleteStatefulClientOpts{
		StatefulClientID: "scId",
	}
	type args struct {
		ctx  context.Context
		opts backend.DeleteStatefulClientOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete StatefulClient",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			assertion: assert.NoError,
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
				ctx:  context.WithValue(context.Background(), "testing.DeactivateContainerWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.DeactivateContainerWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.RemoveContainerWithResponse.error", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.RemoveContainerWithResponse.empty", true),
				opts: opts,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			tt.assertion(t, b.DeleteStatefulClient(tt.args.ctx, tt.args.opts))
		})
	}
}

func TestBackend_GetStatefulClient(t *testing.T) {
	opts := backend.GetStatefulClientOpts{
		StatefulClientID: "scId",
	}
	type args struct {
		ctx  context.Context
		opts backend.GetStatefulClientOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *backend.StatefulClient
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get StatefulClient",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultStatefulClient,
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
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.incomplete_with_Uid_nil", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.incomplete_with_Hostname_nil", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.incomplete_with_Status_nil", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.incomplete_with_Mode_nil", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.incomplete_with_Cores_nil", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "sc client status DOWN response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.statusdown", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "DOWN", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "non client uid query",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetSingleContainerWithResponse.nonClientUidError", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incomplete_with_Uid_nil", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incomplete_with_Hostname_nil", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incomplete_with_Status_nil", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incomplete_with_Mode_nil", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incomplete_with_Roles_nil", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "incorrect response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.incorrect_with_more_than_one_role", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "statefulclient process not ready",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.notready", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
		{
			name: "statefulclient process DOWN",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetProcessesWithResponse.down", true),
				opts: opts,
			},
			want:      &backend.StatefulClient{ID: "scId", Name: "sc", Status: "PROCESSESNOTUP", Mode: "client", Cores: 1},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.GetStatefulClient(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_ListStatefulClients(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts backend.ListStatefulClientsOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      []*backend.StatefulClient
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list StatefulClients",
			args: args{
				ctx:  context.Background(),
				opts: backend.ListStatefulClientsOpts{},
			},
			want: []*backend.StatefulClient{
				&defaultStatefulClient,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter StatefulClients",
			args: args{
				ctx: context.Background(),
				opts: backend.ListStatefulClientsOpts{
					Names: []string{"sc"},
				},
			},
			want: []*backend.StatefulClient{
				&defaultStatefulClient,
			},
			assertion: assert.NoError,
		},
		{
			name: "filter out StatefulClients",
			args: args{
				ctx: context.Background(),
				opts: backend.ListStatefulClientsOpts{
					Names: []string{"sc1"},
				},
			},
			want:      []*backend.StatefulClient{},
			assertion: assert.NoError,
		},
		{
			name:  "wrong admin creds",
			bOpts: mockBackendOpts{adminCredentials: &backend.AuthCreds{Scheme: backend.Basic, Principal: "username", Credentials: "wrongpassword"}},
			args: args{
				ctx:  context.Background(),
				opts: backend.ListStatefulClientsOpts{},
			},
			assertion: assert.Error,
		},
		{
			name: "api error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetContainersWithResponse.error", true),
				opts: backend.ListStatefulClientsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetContainersWithResponse.empty", true),
				opts: backend.ListStatefulClientsOpts{},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "incomplete response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.GetContainersWithResponse.incomplete", true),
				opts: backend.ListStatefulClientsOpts{},
			},
			want:      []*backend.StatefulClient{},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.ListStatefulClients(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
