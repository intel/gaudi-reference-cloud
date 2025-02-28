// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"os"
	"sync"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client/mocks"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/stretchr/testify/assert"
)

type mockBackendOpts struct {
	adminCredentials *backend.AuthCreds
	config           *conf.Cluster
}

var correctCreds = &backend.AuthCreds{
	Scheme:      backend.Basic,
	Principal:   "username",
	Credentials: "password",
}

func newMockBackend(opts mockBackendOpts) Backend {
	var cluster conf.Cluster
	if opts.config != nil {
		cluster = *opts.config
	} else {
		cluster = conf.Cluster{
			Name:     "test",
			UUID:     "00000000-0000-0000-0000-000000000000",
			Type:     conf.Vast,
			Location: "testing",
			API: &conf.API{
				Type: conf.REST,
				URL:  "localhost",
			},
			VastConfig: &conf.VastConfig{
				ProtectedOrgIds: []string{"protected"},
				VipPool:         "1",
			},
		}
	}

	var adminCredentials backend.AuthCreds
	if opts.adminCredentials != nil {
		adminCredentials = *opts.adminCredentials
	} else {
		adminCredentials = backend.AuthCreds{
			Scheme:      backend.Basic,
			Principal:   "username",
			Credentials: "password",
		}
	}

	return Backend{
		config:           &cluster,
		adminCredentials: adminCredentials,
		client:           &mocks.MockVastClient{},
		lock:             sync.Mutex{},
	}
}

// make sure the vast backend implements all the required ops
var _ backend.NamespaceOps = &Backend{}
var _ backend.UserOps = &Backend{}

func TestNewBackend(t *testing.T) {
	os.Setenv("VAST_ENV_TEST", "env:1")

	type args struct {
		config *conf.Cluster
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create vast backend",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "VAST_ENV_TEST",
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "error on no vippool",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on no api",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on no auth",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on wrong auth",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
					Auth: &conf.Auth{
						Scheme: conf.Bearer,
						Env:    "VAST_ENV_TEST",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on parse auth",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "VAST_ENV_TEST1",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on wrong cert",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Vast,
					Location: "testing",
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "wrong_location",
					},
					VastConfig: &conf.VastConfig{
						ProtectedOrgIds: []string{"protected"},
						VipPool:         "1",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "VAST_ENV_TEST",
					},
				},
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBackend(tt.args.config)
			tt.assertion(t, err)
		})
	}
}
