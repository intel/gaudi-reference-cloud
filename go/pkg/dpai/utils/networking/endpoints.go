// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"
	"log"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/dns"
)

func (n *Network) GetExposedServiceEndpoint(ctx context.Context, networkRecords *NetworkRecordsOutput) string {
	return networkRecords.GatewayIstioName + "." + dns.DPAI_ZONE[:len(dns.DPAI_ZONE)-1]
}

// the serviceID is itself <svc-type>-<serviceID> which is a sha256 hash unique enough to prevent collusions
func getDNSFqdnFromServiceIdentifier(serviceId string) string {
	return serviceId + "." + dns.DPAI_ZONE[:len(dns.DPAI_ZONE)-1]
}

func (n *Network) GetExposedServiceEndpointFromDb(ctx context.Context) (string, error) {
	dns := getDNSFqdnFromServiceIdentifier(n.ServiceId)

	dnsRecrod, err := n.SqlModel.GetServiceGatewayForWorkspaceServiceFromDnsFqdn(ctx, db.GetServiceGatewayForWorkspaceServiceFromDnsFqdnParams{
		DnsFqdn:        dns,
		CloudAccountID: n.CloudAccountID,
	})

	if err != nil {
		log.Println("Error getting the service gateway for this service make sure the  service is first added in the workspace", err)
		return "", err
	}

	return dnsRecrod.DnsFqdn, nil
}
