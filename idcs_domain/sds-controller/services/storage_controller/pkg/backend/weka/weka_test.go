// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"os"
	"sync"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4/mocks"
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
			Type:     conf.Weka,
			Location: "testing",
			API: &conf.API{
				Type: conf.REST,
				URL:  "localhost",
			},
			WekaConfig: &conf.WekaConfig{
				ProtectedOrgIds:      []string{"protected"},
				TenantFsGroupName:    "testfsgroup",
				FileSystemDeleteWait: 1,
				BackendFQDN:          "localhost",
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
		client:           &mocks.MockWekaClient{},
		lock:             sync.Mutex{},
		orgNames:         make(map[string]string),
		lrLocks:          make(map[string]*sync.Mutex),
	}
}

// make sure the weka backend implements all the required ops
var _ backend.StatefulClientOps = &Backend{}
var _ backend.UserOps = &Backend{}
var _ backend.NamespaceOps = &Backend{}

func TestNewBackend(t *testing.T) {
	os.Setenv("WEKA_ENV_TEST", "env:1")

	type args struct {
		config *conf.Cluster
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create weka backend",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						BackendFQDN:          "localhost",
						FileSystemDeleteWait: 1,
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "WEKA_ENV_TEST",
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "error on no fsgroup",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "default delete time time ",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:   []string{"protected"},
						TenantFsGroupName: "testfsgroup",
						BackendFQDN:       "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "WEKA_ENV_TEST",
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "error on no api",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Weka,
					Location: "testing",
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
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
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
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
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Bearer,
						Env:    "WEKA_ENV_TEST",
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
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "WEKA_ENV_TEST1",
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
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "wrong_location",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						FileSystemDeleteWait: 1,
						BackendFQDN:          "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "WEKA_ENV_TEST",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on no backendFqdn",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.Weka,
					Location: "testing",
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "testdata/cert.pem",
					},
					WekaConfig: &conf.WekaConfig{
						ProtectedOrgIds:      []string{"protected"},
						TenantFsGroupName:    "testfsgroup",
						FileSystemDeleteWait: 1,
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "WEKA_ENV_TEST",
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
