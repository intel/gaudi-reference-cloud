// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package minio

import (
	"os"
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
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

func TestNewBackend(t *testing.T) {
	os.Setenv("MINIO_ENV_TEST", "env:1")

	type args struct {
		config *conf.Cluster
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create minio backend",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "MINIO_ENV_TEST",
					},
					MinioConfig: &conf.MinioConfig{
						KESKey: "key-name",
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
					Type:     conf.MinIO,
					Location: "testing",
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
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
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
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Bearer,
						Env:    "MINIO_ENV_TEST",
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
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "localhost",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "MINIO_ENV_TEST1",
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
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "wrong_location",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "MINIO_ENV_TEST",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on invalid endpoint",
			args: args{
				config: &conf.Cluster{
					Name:     "test",
					UUID:     "00000000-0000-0000-0000-000000000000",
					Type:     conf.MinIO,
					Location: "testing",
					API: &conf.API{
						Type: conf.REST,
						URL:  "---",
					},
					Auth: &conf.Auth{
						Scheme: conf.Basic,
						Env:    "MINIO_ENV_TEST",
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
