// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package k8s

import (
	"context"
	"log"

	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IstioClientSet struct {
	ClientSet *versionedclient.Clientset
	Ctx       context.Context
}

type IstioGatewayConfig struct {
	Name                           string
	Namespace                      string
	IstioGatewayControllerSelector map[string]string
	IstioCertManagerSecretName     string // the cert manager secret name for the self-signed elliptic curve secret
	GatewayHostName                []string
	IstioGatewayCustomLabels       map[string]string
}

func (istio *IstioClientSet) CreateGateway(istioGwConfig *IstioGatewayConfig) error {
	// the gateway should run always in https with tls 1.2+ expected standard way to have elliptical curve
	gateway := &networkingv1alpha3.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      istioGwConfig.Name,
			Namespace: istioGwConfig.Namespace,
			Labels:    istioGwConfig.IstioGatewayCustomLabels,
		},
		Spec: v1alpha3.Gateway{
			Selector: istioGwConfig.IstioGatewayControllerSelector,
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Number:   443,
						Name:     "https",
						Protocol: "HTTPS",
					},
					Hosts: istioGwConfig.GatewayHostName,
					Tls: &v1alpha3.ServerTLSSettings{
						Mode:               v1alpha3.ServerTLSSettings_SIMPLE,
						CredentialName:     istioGwConfig.IstioCertManagerSecretName,
						MinProtocolVersion: v1alpha3.ServerTLSSettings_TLSV1_2,
						MaxProtocolVersion: v1alpha3.ServerTLSSettings_TLSV1_3,
					},
				},
			},
		},
	}
	_, err := istio.ClientSet.NetworkingV1alpha3().Gateways(gateway.Namespace).Create(istio.Ctx, gateway, v1.CreateOptions{})
	if err != nil {
		log.Printf("Error creating gateway %s", err)
		return err
	}

	log.Printf("Gateway created successfully with Name %s", istioGwConfig.Name)
	return nil
}

func (istio *IstioClientSet) DeleteGateway(istioGwConfig *IstioGatewayConfig) error {
	err := istio.ClientSet.NetworkingV1alpha3().Gateways(istioGwConfig.Namespace).Delete(istio.Ctx, istioGwConfig.Name,
		v1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting Istio gateway %s", err)
		return err
	}

	return nil
}
