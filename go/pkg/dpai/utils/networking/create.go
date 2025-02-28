// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"
	"fmt"
	"log"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/dns"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
	1. Create DNS CNAME record for the service to be added in workspace
	2. Create Istio Gateway Cert for Cert manager self sign PKI to sign
	3. Create Istio Gateway resource for the service
*/

// creates CNAME DNS record in menmice, returns CNAME in DPAI zone, Parent LB A Record in IDC Zone
func (n *Network) CreateDnsCnameRecord(ctx context.Context) (*NetworkRecordsOutput, error) {

	// in DPAI we only care and process creation only for Primary Zone for DNS
	log.Printf("Creating DNS CNAME record for the service %s with service Id %s", n.ServiceType, n.ServiceId)

	lbresource, err := n.SqlModel.GetHwGatewaysByWorkspaceId(context.Background(),
		model.GetHwGatewaysByWorkspaceIdParams{
			WorkspaceID:    n.WorkspaceId,
			CloudAccountID: n.CloudAccountID,
		})

	if err != nil {
		log.Println("Error getting the DPAI LB resource", err)
		return nil, err
	}

	log.Printf("Configuring the Dns Record for the Highwire Vip for the LB A Record %s in IDC Zone.", lbresource.LbFqdn)

	if !lbresource.LbCreated {
		log.Printf("Cannot add new service to the workspace as the LB is not created yet")
		return nil, fmt.Errorf("Load Balancer is not created yet")
	}

	/* the serviceID is itself <svc-type>-<serviceID> which is a sha256 hash unique enough to prevent collusions
	Sample Service ID: af-nb3kgmtp45jx, where first part is the service prefix respective to a service while second is last sum of SHA256 hash
	*/
	lbDnsARecord := lbresource.LbFqdn
	dnsFqdn := fmt.Sprintf("%s", n.ServiceId)

	zoneMap, err := n.DnsService.FetchDNSZoneByName(dns.DPAI_ZONE)
	if err != nil {
		fmt.Printf("Error getting the DNS Zone for the zone %s error: %+v", dns.DPAI_ZONE, err)
		return nil, err
	}

	dnsRecordSuccessResp, err := n.DnsService.CreateDnsRecordInZone(zoneMap,
		dnsFqdn, lbDnsARecord, dns.REALM_INTERNAL)
	if err != nil {
		log.Printf("Error creating the DNS Record in zone %s error: %+v", dns.DPAI_ZONE, err)
		return nil, err
	}

	log.Println("The Dns Record is successfully created and linked via cname to parent LB record created by Highwire ", lbDnsARecord)
	log.Println("tHE dns Response successfull creation response ", dnsRecordSuccessResp)

	// commit and update the information in workspace gateway table holding all informations for the workspace and the associated gateways added into it
	_, err = n.SqlModel.InsertGatewayForWorkspaceService(context.Background(), model.InsertGatewayForWorkspaceServiceParams{
		CloudAccountID:             n.CloudAccountID,
		DnsFqdn:                    dnsFqdn,
		GatewayIstioName:           dnsFqdn,
		GatewayIstioSecretName:     fmt.Sprintf("%s-secret", dnsFqdn),
		GatewaySelectorIstioLabels: dnsFqdn,
		LbFqdn:                     lbDnsARecord,
		IsActive:                   true,
	})

	if err != nil {
		log.Printf("Error adding dns cname record information in database for service addition in workspace %+v", err)
		return nil, err
	}

	return &NetworkRecordsOutput{
		DnsCnameRecord:               dnsFqdn,
		WorkspaceLoadBalancerARecord: lbDnsARecord,
		GatewayIstioName:             dnsFqdn,
		GatewayIstioSecretName:       fmt.Sprintf("%s-secret", dnsFqdn),
		GatewayIstioLabelName:        dnsFqdn,
	}, nil
}

// create the self sign cert for the respective DPAI service on the istio edge ingress Gateway
func (n *Network) CreateIstioGatewayCertificate(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {
	log.Printf("Creating the Istio Gateway Certificate for the DPAI Service type %s with id %s", n.ServiceId, n.ServiceType)
	n.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    n.IksClusterUUID,
		CloudAccountId: n.CloudAccountID,
	}

	err := n.K8sClient.GetIksClient(n.ServiceConfig)

	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while creating certificate for Istio Gateway")
		return nil, err
	}

	// for testing now assume the lb is ready and firewall rule is added and proceed if firewall rule fails other way asynchronously the user should be infromed

	client, err := GetCertManagerClient(n.K8sClientConfig)

	if err != nil {
		log.Println("Error getting the cert manager client", err)
		return nil, err
	}

	istioGatewayCertName := fmt.Sprintf("%s-cert", networkRecords.GatewayIstioName)
	istioGatewayCertSan := networkRecords.DnsCnameRecord + "." + dns.DPAI_ZONE[:len(dns.DPAI_ZONE)-1]

	dnsHostNames := make([]string, 0)
	dnsHostNames = append(dnsHostNames, istioGatewayCertSan) // append and always keep this as the single SAN

	dpaiServicegatewayCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      istioGatewayCertName,
			Namespace: IstioNamespace,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: networkRecords.GatewayIstioSecretName,
			CommonName: dnsHostNames[0],
			DNSNames:   dnsHostNames,
			IssuerRef: cmmeta.ObjectReference{
				Name:  IstioK8SCSR,
				Kind:  CLUSTER_ISSUER,
				Group: CLUSTER_ISSUER_API_VERSION,
			},
		},
	}

	err = client.Create(ctx, dpaiServicegatewayCert)
	if err != nil {
		log.Println("Error in creating the certificate for the dpai istio gateway", err)
	}

	return networkRecords, nil
}

func (n *Network) CreateIstioGateway(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {
	log.Printf("Creating the Istio Gateway for the workspace with name %s", networkRecords.GatewayIstioName)

	n.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    n.IksClusterUUID,
		CloudAccountId: n.CloudAccountID,
	}

	err := n.K8sClient.GetIksClient(n.ServiceConfig)
	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while installing Istio Gateways")
		return nil, err
	}

	/*  istio envoy edge does a strict host matching for upstream vhost envoy connection transfer this need to always match the host information for tls handshake completion
	trim end due to DDI referes domain names via fqdn with delimeter to term valid end DNS record in zone
	This only assumes for now a single service is exposed out
	*/
	istioGatewayHostName := networkRecords.DnsCnameRecord + "." + dns.DPAI_ZONE[:len(dns.DPAI_ZONE)-1]

	err = n.IstioClientSet.CreateGateway(&k8s.IstioGatewayConfig{
		Name:      networkRecords.DnsCnameRecord,
		Namespace: IstioNamespace,
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
		log.Println("The Istio Ingress Gateway for the workspace with name broke", networkRecords.GatewayIstioName)
		log.Println("Error creating the istio gateway", err)
		return nil, err
	}
	return networkRecords, nil
}

// use coredns for internal kubernetes fqdn dns service resolution
func GetK8sServiceFqdn(serviceName string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
}

func (n *Network) CreateIstioVirtualService(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {

	log.Printf("Creating the Istio Virtual Service for the workspace with name %s", networkRecords.GatewayIstioName)
	for service, port := range n.ExposedK8sServiceNames {
		if err := n.IstioClientSet.CreateVirtualService(ctx, &k8s.VirtualServiceConfig{
			VServiceName: fmt.Sprintf("vs-%s", networkRecords.GatewayIstioName),
			ServiceFqdn:  GetK8sServiceFqdn(service, n.ServiceNamespace),
			ServicePort:  port,
			Namespace:    n.ServiceNamespace,
			GatewayHostName: []string{
				/*
					Point the to the upstream Istio Gateway Name pointing to this virtual service
				*/
				n.GetExposedServiceEndpoint(ctx, networkRecords),
			},
			IstioGateways: []string{
				networkRecords.GatewayIstioName,
			},
		}); err != nil {
			log.Printf("Error creating the Istio Virtual Service for the workspace with name %s", networkRecords.GatewayIstioName)
		}
	}
	return networkRecords, nil
}

func (n *Network) CreateIstioDestinationRule(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {
	log.Printf("Creating Istio Destination Rule for the workspace with name %s", networkRecords.DnsCnameRecord)
	return networkRecords, nil
}
