// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vast

import (
	"context"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

var defaultQuota uint64 = 100500

var defaultView = View{
	ID:        "2",
	Name:      "view",
	Path:      "/view",
	Protocols: []Protocol{NFSV3},
	PolicyID:  1,
	TotalBytes: defaultQuota,
}

func TestBackend_CreateView(t *testing.T) {

	opts := CreateViewOpts{
		NamespaceID: defaultNamespace.ID,
		Name:        "view",
		Path:        "/view",
		Protocols:   []Protocol{NFSV3},
		Quota:       defaultQuota,
	}
	type args struct {
		ctx  context.Context
		opts CreateViewOpts
	}
	tests := []struct {
		name      string
		args      args
		bOpts     mockBackendOpts
		want      *View
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create view",
			args: args{
				ctx:  context.Background(),
				opts: opts,
			},
			want:      &defaultView,
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
				ctx:  context.WithValue(context.Background(), "testing.ViewsCreateWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "empty response",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.ViewsCreateWithResponse.empty", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "quota error",
			args: args{
				ctx:  context.WithValue(context.Background(), "testing.QuotasCreateWithResponse.error", true),
				opts: opts,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMockBackend(tt.bOpts)
			got, err := b.CreateView(tt.args.ctx, tt.args.opts)
			tt.assertion(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
// ToDo update view test:
// - update name
//   - view exists
//   - view doesn't exist
// - update quote
//   - view exists, quota exists
//   - view exists, quota doesn't exist
//   - view doesn't exist
