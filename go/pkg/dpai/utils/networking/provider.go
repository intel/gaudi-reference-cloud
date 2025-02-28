// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"

	"k8s.io/client-go/rest"
)

type NetworkUtilServices interface {
	InitDnsService(context.Context)
	InitIstioClient(context.Context, *rest.Config)
}

type DnsExposedEndpoints interface {
	// returns the single FQDN exposed for the service with a dedicated Istio Gateway, fetched from database
	GetExposedServiceEndpoint(context.Context, *NetworkRecordsOutput) string
	// returns the single FQDN exposed for the service with a dedicated Istio Gateway,
	GetExposedServiceEndpointFromDb(context.Context) (string, error)
}

type FirewallServices interface {
	VerifyIKSFirewallCreation(context.Context, *FirewallVerifyRequest) (bool, error)
}

type NetworkResourceCreate interface {
	CreateDnsCnameRecord(context.Context) (*NetworkRecordsOutput, error)
	CreateIstioGatewayCertificate(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	CreateIstioGateway(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	CreateIstioVirtualService(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	CreateIstioDestinationRule(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
}

type NetworkResourceDelete interface {
	DeleteServiceCnameRecord(context.Context, string) (*NetworkRecordsOutput, error)
	DeleteIstioGatewayCertificate(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	DeleteServiceIstioGateway(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	DeleteIstioVirtualService(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	DeleteIstioDestinationRule(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
	DeleteIKSLoadbalancerResource(context.Context) (*NetworkRecordsOutput, error)
	DeleteNetworkingResourcesInDB(context.Context, *NetworkRecordsOutput) (*NetworkRecordsOutput, error)
}

type NetworkResourceProvider interface {
	NetworkUtilServices
	DnsExposedEndpoints
	NetworkResourceCreate
	NetworkResourceDelete
}
