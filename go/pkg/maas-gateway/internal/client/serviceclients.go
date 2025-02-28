// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
)

type ServiceClients struct {
	config               *config.Config
	usageRecordClient    *ServiceClient[pb.UsageRecordServiceClient]
	productCatalogClient *ServiceClient[pb.ProductCatalogServiceClient]
	dispatcherClient     *ServiceClient[pb.DispatcherClient]
	logger               logr.Logger
	connector            Connector
}

func (sc *ServiceClients) UsageRecordClient() *ServiceClient[pb.UsageRecordServiceClient] {
	return sc.usageRecordClient
}

func (sc *ServiceClients) ProductCatalogClient() *ServiceClient[pb.ProductCatalogServiceClient] {
	return sc.productCatalogClient
}

func (sc *ServiceClients) DispatcherClient() *ServiceClient[pb.DispatcherClient] {
	return sc.dispatcherClient
}

func SetupServiceClients(config *config.Config, logger logr.Logger, connector Connector) *ServiceClients {

	usageRecordClient := NewServiceClient[pb.UsageRecordServiceClient](pb.NewUsageRecordServiceClient)
	productCatalogClient := NewServiceClient[pb.ProductCatalogServiceClient](pb.NewProductCatalogServiceClient)
	dispatcherClient := NewServiceClient[pb.DispatcherClient](pb.NewDispatcherClient)

	return &ServiceClients{
		usageRecordClient:    usageRecordClient,
		productCatalogClient: productCatalogClient,
		dispatcherClient:     dispatcherClient,
		logger:               logger.WithName("SetupServiceClients"),
		config:               config,
		connector:            connector,
	}
}

func (sc *ServiceClients) Connect(ctx context.Context) error {
	usageRecordClientConn, err := sc.connector.GetIdcConnection(ctx, sc.config.UsageServerAddr)
	if err != nil {
		return errors.Wrap(err, "couldn't create metering client connection")
	}
	sc.usageRecordClient.Connect(usageRecordClientConn)

	productCatalogConn, err := sc.connector.GetIdcConnection(ctx, sc.config.ProductCatalogServerAddr)
	if err != nil {
		return errors.Wrap(err, "couldn't create metering client connection")
	}
	sc.productCatalogClient.Connect(productCatalogConn)

	dispatcherConn, err := sc.connector.GetIksConnection(ctx, sc.config.DispatcherServerAddr)
	if err != nil {
		return errors.Wrap(err, "couldn't create dispatcher client connection")
	}

	sc.dispatcherClient.Connect(dispatcherConn)

	return nil
}

func (sc *ServiceClients) Close() {

	if sc.usageRecordClient.GrpcConnection() != nil {
		if err := sc.usageRecordClient.Close(); err != nil {
			sc.logger.Error(err, "error closing metering client")
		}
	}

	if sc.productCatalogClient.GrpcConnection() != nil {
		if err := sc.productCatalogClient.Close(); err != nil {
			sc.logger.Error(err, "error closing product access client")
		}
	}

	if sc.dispatcherClient.GrpcConnection() != nil {
		if err := sc.dispatcherClient.Close(); err != nil {
			sc.logger.Error(err, "error closing dispatcher client")
		}
	}
}
