// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

// Delete a specific Service from workspace

import (
	"context"
	"fmt"
	"log"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/dns"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Delete Cname Records fro the specific service
func (n *Network) DeleteServiceCnameRecord(ctx context.Context, dnsCnameRecord string) (*NetworkRecordsOutput, error) {

	zoneMap, err := n.DnsService.FetchDNSZoneByName(dns.DPAI_ZONE)
	if err != nil {
		fmt.Printf("Error fetching the DNS Zone for the DPAI Zone %+v", err)
		return nil, err
	}

	serviceDnsResource, err := n.SqlModel.GatGatewayForWorkspaceService(context.Background(), model.GatGatewayForWorkspaceServiceParams{
		CloudAccountID: n.CloudAccountID,
		DnsFqdn:        dnsCnameRecord,
	})

	if err != nil {
		log.Println("Error getting the service gateway for this service make sure the  service is first added in the workspace", err)
		return nil, err
	}

	err = n.DnsService.DeleteDnsRecordsInZoneByDataRef(zoneMap, dnsCnameRecord)

	if err != nil {
		log.Fatal("Error deleting the DNS Record by Data Ref in the DNS Primary Zone", err)
		return nil, err
	}

	// remove it from the database holding the service gateway

	err = n.SqlModel.DeleteGatewayForWorkspaceService(context.Background(), model.DeleteGatewayForWorkspaceServiceParams{
		DnsFqdn: dnsCnameRecord,
	})

	return &NetworkRecordsOutput{
		GatewayIstioName:       serviceDnsResource.GatewayIstioName,
		GatewayIstioLabelName:  serviceDnsResource.GatewaySelectorIstioLabels,
		GatewayIstioSecretName: serviceDnsResource.GatewayIstioSecretName,
	}, nil
}

func (n *Network) DeleteIstioGatewayCertificate(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {

	n.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    n.IksClusterUUID,
		CloudAccountId: n.CloudAccountID,
	}

	err := n.K8sClient.GetIksClient(n.ServiceConfig)

	if err != nil {
		log.Printf("Error getting the IKS k8s client set %+v", err)
		return nil, err
	}

	client, err := GetCertManagerClient(n.K8sClientConfig)

	// use controller runtime for registering the custom cert manager schema to manager
	if err != nil {
		log.Println("Error getting the cert manager client", err)
		return nil, err
	}

	gatewayCertResource := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkRecords.GatewayIstioSecretName,
			Namespace: deployment.IstioNamespace,
		},
	}

	if err := client.Delete(context.Background(), gatewayCertResource); err != nil {
		log.Printf("Error deleting the cert manager certificate resource %+v", err)
		return nil, err
	}

	log.Println("Deleted the cert manager certificate resource for Istio Gateway ", networkRecords.GatewayIstioSecretName)

	return networkRecords, nil
}

func (n *Network) DeleteServiceIstioGateway(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {
	n.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    n.IksClusterUUID,
		CloudAccountId: n.CloudAccountID,
	}

	err := n.K8sClient.GetIksClient(n.ServiceConfig)

	if err != nil {
		log.Printf("Error getting the IKS k8s client set %+v", err)
		return nil, err
	}

	istioGatewayHostName := networkRecords.GatewayIstioName + "." + dns.DPAI_ZONE[:len(dns.DPAI_ZONE)-1]

	err = n.IstioClientSet.DeleteGateway(&k8s.IstioGatewayConfig{
		Name:      networkRecords.GatewayIstioName,
		Namespace: deployment.IstioNamespace,
		IstioGatewayControllerSelector: map[string]string{
			"istio": "ingressgateway", // will always always point to the underlying istio ingress controller
		},
		IstioCertManagerSecretName: networkRecords.GatewayIstioSecretName,
		GatewayHostName: []string{
			istioGatewayHostName,
		},
		IstioGatewayCustomLabels: map[string]string{},
	})

	if err != nil {
		log.Printf("Error Deleting The Istio Ingress Gateway for the service with name %s", networkRecords.GatewayIstioName)
		return nil, err
	}

	return networkRecords, nil
}

func (n *Network) DeleteIstioVirtualService(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {

	return networkRecords, nil
}

func (n *Network) DeleteIstioDestinationRule(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {
	return networkRecords, nil
}
