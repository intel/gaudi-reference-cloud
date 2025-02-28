package backend

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
)

func CreateHTTPTransport(config *conf.Cluster, gzip bool) (*http.Transport, error) {
	if config.API == nil {
		return nil, fmt.Errorf("api filed in the config cannot be nil for cluster: %s", config.UUID)
	}

	tlsConfig := tls.Config{
		InsecureSkipVerify: config.API.CaCertFile == "",
	}

	if config.API.CaCertFile != "" {
		caCert, err := os.ReadFile(config.API.CaCertFile)
		if err != nil {
			return nil, fmt.Errorf("cannot load CA certificate %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	}

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:       5 * time.Second,
			KeepAlive:     15 * time.Second,
			FallbackDelay: 100 * time.Millisecond,
		}).DialContext,
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   1024,
		ResponseHeaderTimeout: 60 * time.Second,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    !gzip,
		TLSClientConfig:       &tlsConfig,
	}, nil
}
