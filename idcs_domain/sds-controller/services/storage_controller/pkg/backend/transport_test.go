package backend

import (
	"testing"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	type args struct {
		config *conf.Cluster
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "read tls cert",
			args: args{
				config: &conf.Cluster{
					Name: "test",
					UUID: "00000000-0000-0000-0000-000000000000",
					Type: conf.Weka,
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "testdata/cert.pem",
					},
				},
			},
			assertion: assert.NoError,
		},
		{
			name: "error on wrong cert",
			args: args{
				config: &conf.Cluster{
					Name: "test",
					UUID: "00000000-0000-0000-0000-000000000000",
					Type: conf.Weka,
					API: &conf.API{
						Type:       conf.REST,
						URL:        "localhost",
						CaCertFile: "wrong_location",
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "error on no API",
			args: args{
				config: &conf.Cluster{
					Name: "test",
					UUID: "00000000-0000-0000-0000-000000000000",
					Type: conf.Weka,
				},
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateHTTPTransport(tt.args.config, true)
			tt.assertion(t, err)
		})
	}
}
