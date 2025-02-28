// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadStorageConfig(t *testing.T) {
	type args struct {
		configLocation string
	}
	tests := []struct {
		name      string
		args      args
		want      *Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty config",
			args: args{
				configLocation: "testdata/empty.yaml",
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "weka config",
			args: args{
				configLocation: "testdata/weka.yaml",
			},
			want: &Config{
				ListenPort: 5000,
				Clusters: []*Cluster{
					{
						Name:     "test",
						UUID:     "00000000-0000-0000-0000-000000000000",
						Type:     Weka,
						Location: "testing",
						Labels:   map[string]string{"label": "value"},
						API: &API{
							Type: REST,
							URL:  "http://testing.localhost",
						},
						SupportsAPI: []SupportsAPI{
							WekaFilesystem,
							ObjectStore,
						},
						Auth: &Auth{
							Scheme: Basic,
							Env:    "WEKA_CREDS",
						},
						WekaConfig: &WekaConfig{
							ProtectedOrgIds:      []string{"00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001"},
							TenantFsGroupName:    "testfsgroup",
							FileSystemDeleteWait: 5,
							BackendFQDN:          "testing.localhost",
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "minio config",
			args: args{
				configLocation: "testdata/minio.yaml",
			},
			want: &Config{
				ListenPort: 5000,
				Clusters: []*Cluster{
					{
						Name:     "test",
						UUID:     "00000000-0000-0000-0000-000000000000",
						Type:     MinIO,
						Location: "testing",
						Labels:   map[string]string{"label": "value"},
						API: &API{
							Type: REST,
							URL:  "http://testing.localhost",
						},
						Auth: &Auth{
							Scheme: Basic,
							Env:    "WEKA_CREDS",
						},
						MinioConfig: &MinioConfig{
							KESKey: "key-name",
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "vast config",
			args: args{
				configLocation: "testdata/vast.yaml",
			},
			want: &Config{
				ListenPort: 5000,
				Clusters: []*Cluster{
					{
						Name:     "test",
						UUID:     "00000000-0000-0000-0000-000000000000",
						Type:     Vast,
						Location: "testing",
						Labels:   map[string]string{"label": "value"},
						API: &API{
							Type: REST,
							URL:  "https://127.0.0.1:8443/api/v2",
						},
						Auth: &Auth{
							Scheme: Basic,
							Env:    "VAST_CREDS",
						},
						SupportsAPI: []SupportsAPI{
							VastView,
						},
						VastConfig: &VastConfig{
							ProtectedOrgIds: []string{"1"},
							VipPool:         "pool1",
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "multiple config",
			args: args{
				configLocation: "testdata/multiple.yaml",
			},
			want: &Config{
				ListenPort: 5000,
				GrpcTLS: &GrpcTLS{
					CertFile: "cert.cert",
					KeyFile:  "key.pem",
				},
				Clusters: []*Cluster{
					{
						Name:     "test",
						UUID:     "00000000-0000-0000-0000-000000000000",
						Type:     Weka,
						Location: "testing",
						Labels:   map[string]string{"label": "value"},
						API: &API{
							Type:       REST,
							URL:        "http://testing.localhost",
							CaCertFile: "caCert.crt",
						},
						Auth: &Auth{
							Scheme: Basic,
							Env:    "WEKA_CREDS",
						},
						WekaConfig: &WekaConfig{
							ProtectedOrgIds:   []string{"00000000-0000-0000-0000-000000000000"},
							TenantFsGroupName: "testfsgroup",
							BackendFQDN:       "testing.localhost",
						},
					},
					{
						Name:     "test2",
						UUID:     "00000000-0000-0000-0000-000000000002",
						Type:     Weka,
						Location: "testing2",
						Labels:   map[string]string{"label": "value2"},
						API: &API{
							Type: REST,
							URL:  "http://testing.localhost",
						},
						Auth: &Auth{
							Scheme: Bearer,
							File:   "password",
						},
						WekaConfig: &WekaConfig{
							TenantFsGroupName: "testfsgroup",
							BackendFQDN:       "testing.localhost",
						},
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "wrong config",
			args: args{
				configLocation: "testdata/wrong.yaml",
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "non found config",
			args: args{
				configLocation: "testdata/nonexists.yaml",
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadStorageConfig(tt.args.configLocation)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadCredentials(t *testing.T) {
	os.Setenv("TEST_PASSWORD", "user:password")
	os.Setenv("TEST_TOKEN", "tokenvalue")
	type args struct {
		auth Auth
	}
	tests := []struct {
		name      string
		args      args
		want      *AuthCreds
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "basic env",
			args: args{
				auth: Auth{
					Scheme: Basic,
					Env:    "TEST_PASSWORD",
				},
			},
			want: &AuthCreds{
				Scheme:      Basic,
				Principal:   "user",
				Credentials: "password",
			},
			assertion: assert.NoError,
		},
		{
			name: "token env",
			args: args{
				auth: Auth{
					Scheme: Bearer,
					Env:    "TEST_TOKEN",
				},
			},
			want: &AuthCreds{
				Scheme:      Bearer,
				Credentials: "tokenvalue",
			},
			assertion: assert.NoError,
		},
		{
			name: "digest env",
			args: args{
				auth: Auth{
					Scheme: Digest,
					Env:    "TEST_TOKEN",
				},
			},
			want: &AuthCreds{
				Scheme:      Digest,
				Credentials: "tokenvalue",
			},
			assertion: assert.NoError,
		},
		{
			name: "wrong basic env",
			args: args{
				auth: Auth{
					Scheme: Basic,
					Env:    "TEST_TOKEN",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "basic file",
			args: args{
				auth: Auth{
					Scheme: Basic,
					File:   "testdata/password",
				},
			},
			want: &AuthCreds{
				Scheme:      Basic,
				Principal:   "user",
				Credentials: "password",
			},
			assertion: assert.NoError,
		},
		{
			name: "token file",
			args: args{
				auth: Auth{
					Scheme: Bearer,
					File:   "testdata/token",
				},
			},
			want: &AuthCreds{
				Scheme:      Bearer,
				Credentials: "tokenvalue",
			},
			assertion: assert.NoError,
		},
		{
			name: "vault file",
			args: args{
				auth: Auth{
					Scheme:    Basic,
					VaultFile: "testdata/vault.json",
				},
			},
			want: &AuthCreds{
				Scheme:      Basic,
				Principal:   "user",
				Credentials: "password",
			},
			assertion: assert.NoError,
		},
		{
			name: "digest file",
			args: args{
				auth: Auth{
					Scheme: Digest,
					File:   "testdata/token",
				},
			},
			want: &AuthCreds{
				Scheme:      Digest,
				Credentials: "tokenvalue",
			},
			assertion: assert.NoError,
		},
		{
			name: "wrong basic file",
			args: args{
				auth: Auth{
					Scheme: Basic,
					File:   "testdata/token",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail for secret",
			args: args{
				auth: Auth{
					Scheme: Basic,
					Secret: "secret-name",
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail for empty",
			args: args{
				auth: Auth{
					Scheme: Basic,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail for invalid file",
			args: args{
				auth: Auth{
					File:   "testdata/wrong",
					Scheme: Basic,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail for invalid vault file",
			args: args{
				auth: Auth{
					VaultFile: "testdata/wrong",
					Scheme:    Basic,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "fail for invalid env",
			args: args{
				auth: Auth{
					Env:    "TEST_WRONG",
					Scheme: Basic,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadCredentials(tt.args.auth)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
