// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ovnclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/configmgr"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/types"
)

var (
	Logger logr.Logger
)

// Modified from https://github.com/ovn-org/ovn-kubernetes/blob/master/go-controller/pkg/libovsdb/libovsdb.go
func createTLSConfig(certFile, privKeyFile, caCertFile string) (*tls.Config, error) {
	logger := Logger.WithName("createTLSConfig")
	cert, err := tls.LoadX509KeyPair(certFile, privKeyFile)
	if err != nil {
		logger.Error(err, "Error loading x509 certs.")
		return nil, err
	}
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		logger.Error(err, "Error loading ca certs.")
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}
	return tlsConfig, nil
}

// NewNBClient creates a new OVN Northbound Database client
func NewNBClient(s string, ctx context.Context, sslConfig *configmgr.SslConfig) (libovsdbclient.Client, error) {
	var err error
	var dbModel model.ClientDBModel
	var nbClient libovsdbclient.Client

	logger := Logger.WithName("NewNBClient")

	dbModel, err = nbdb.FullDatabaseModel()
	if err != nil {
		logger.Error(err, "Cannot create ovsdb model.\n")
		return nil, err
	}

	// define client indexes for objects that are using dbIDs
	dbModel.SetIndexes(map[string][]model.ClientIndex{
		nbdb.ACLTable:           {{Columns: []model.ColumnKey{{Column: "external_ids", Key: types.PrimaryIDKey}}}},
		nbdb.DHCPOptionsTable:   {{Columns: []model.ColumnKey{{Column: "external_ids", Key: types.PrimaryIDKey}}}},
		nbdb.LoadBalancerTable:  {{Columns: []model.ColumnKey{{Column: "name"}}}},
		nbdb.LogicalSwitchTable: {{Columns: []model.ColumnKey{{Column: "name"}}}},
		nbdb.LogicalRouterTable: {{Columns: []model.ColumnKey{{Column: "name"}}}},
	})

	nbClientLogger := log.FromContext(ctx).WithName("OvnClient")
	var sslEnabled bool
	sslEnabled, err = strconv.ParseBool(sslConfig.Enabled)
	if err != nil {
		logger.Error(err, "Cannot parse SSL enabled cfg for OVN.\n")
		return nil, err
	}
	if !sslEnabled {
		nbClient, err = libovsdbclient.NewOVSDBClient(dbModel,
			libovsdbclient.WithEndpoint("tcp:"+s),
			libovsdbclient.WithLogger(&nbClientLogger))
		if err != nil {
			logger.Error(err, "Cannot create insecure ovsdb client.\n")
			return nil, err
		}
		logger.Info("Created ovsdb client without SSL")
	} else {
		tlsConfig, err := createTLSConfig(sslConfig.Cert, sslConfig.Key, sslConfig.Ca)
		if err != nil {
			logger.Error(err, "Cannot create TLS config for OVN.\n")
			return nil, err
		}
		nbClient, err = libovsdbclient.NewOVSDBClient(dbModel,
			libovsdbclient.WithEndpoint("ssl:"+s),
			libovsdbclient.WithLogger(&nbClientLogger),
			libovsdbclient.WithTLSConfig(tlsConfig))
		if err != nil {
			logger.Error(err, "Cannot create secure ovsdb client.\n")
			return nil, err
		}
		logger.Info("Created ovsdb client with SSL")
	}

	err = nbClient.Connect(context.Background())
	if err != nil {
		logger.Error(err, "Cannot connect to ovn-central.\n")
		return nil, err
	}
	logger.Info("Connected to ovn-central")
	nbClient.MonitorAll(context.Background())
	logger.Info("nb client monitor all")

	return nbClient, nil
}
