// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"os"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const testClusterUuid = "00000000-0000-0000-0000-000000000000"

func TestServer_CreateGrpcServer(t *testing.T) {
	os.Setenv("CREDS_ENV_TEST", "env:1")

	tests := []struct {
		name      string
		s         *Server
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "plaintext weka server",
			s: &Server{
				Config: &conf.Config{
					ListenPort: 500,
					Clusters: []*conf.Cluster{
						{
							Name: "test",
							Type: conf.Weka,
							UUID: testClusterUuid,
							API: &conf.API{
								Type: conf.REST,
								URL:  "localhost",
							},
							Auth: &conf.Auth{
								Scheme: conf.Basic,
								Env:    "CREDS_ENV_TEST",
							},
							WekaConfig: &conf.WekaConfig{
								TenantFsGroupName: "defaultgroup",
								BackendFQDN:       "localhost",
							},
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "tls weka server",
			s: &Server{
				Config: &conf.Config{
					ListenPort: 500,
					GrpcTLS: &conf.GrpcTLS{
						CertFile: "testdata/server-cert.pem",
						KeyFile:  "testdata/server-key.pem",
					},
					Clusters: []*conf.Cluster{
						{
							Name: "test",
							Type: conf.Weka,
							UUID: testClusterUuid,
							API: &conf.API{
								Type: conf.REST,
								URL:  "localhost",
							},
							Auth: &conf.Auth{
								Scheme: conf.Basic,
								Env:    "CREDS_ENV_TEST",
							},
							WekaConfig: &conf.WekaConfig{
								TenantFsGroupName: "defaultgroup",
								BackendFQDN:       "localhost",
							},
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "tls minio server",
			s: &Server{
				Config: &conf.Config{
					ListenPort: 500,
					GrpcTLS: &conf.GrpcTLS{
						CertFile: "testdata/server-cert.pem",
						KeyFile:  "testdata/server-key.pem",
					},
					Clusters: []*conf.Cluster{
						{
							Name: "test",
							Type: conf.MinIO,
							UUID: testClusterUuid,
							API: &conf.API{
								Type: conf.REST,
								URL:  "localhost",
							},
							Auth: &conf.Auth{
								Scheme: conf.Basic,
								Env:    "CREDS_ENV_TEST",
							},
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "tls vast server",
			s: &Server{
				Config: &conf.Config{
					ListenPort: 500,
					GrpcTLS: &conf.GrpcTLS{
						CertFile: "testdata/server-cert.pem",
						KeyFile:  "testdata/server-key.pem",
					},
					Clusters: []*conf.Cluster{
						{
							Name: "test",
							Type: conf.Vast,
							UUID: testClusterUuid,
							API: &conf.API{
								Type: conf.REST,
								URL:  "localhost",
							},
							Auth: &conf.Auth{
								Scheme: conf.Basic,
								Env:    "CREDS_ENV_TEST",
							},
							VastConfig: &conf.VastConfig{
								VipPool:         "pool",
								ProtectedOrgIds: []string{"0"},
							},
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name:      "plaintext empty",
			s:         &Server{},
			assertion: assert.Error,
		},
		{
			name: "tls fail non existent",
			s: &Server{
				Config: &conf.Config{
					GrpcTLS: &conf.GrpcTLS{
						CertFile: "wrong",
						KeyFile:  "wrong",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "tls fail non existent",
			s: &Server{
				Config: &conf.Config{
					GrpcTLS: &conf.GrpcTLS{
						CertFile: "wrong",
						KeyFile:  "wrong",
					},
				},
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.s.CreateGrpcServer()
			tt.assertion(t, err)
		})
	}
}

func TestServer_interceptorLogger(t *testing.T) {

	hook := logHook{}
	logger := zerolog.New(os.Stdout)
	logger = logger.Hook(&hook)

	f := interceptorLogger(logger)

	f.Log(context.Background(), logging.LevelDebug, "debug")
	f.Log(context.Background(), logging.LevelInfo, "info")
	f.Log(context.Background(), logging.LevelWarn, "warn")
	f.Log(context.Background(), logging.LevelError, "error")

	assert.Equal(t, hook.records, []string{"debug", "info", "warn", "error"})
}

type logHook struct {
	records []string
}

func (logHook *logHook) Run(logEvent *zerolog.Event, level zerolog.Level, message string) {
	logHook.records = append(logHook.records, message)
}
