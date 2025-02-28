// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/protobuf/types/known/durationpb"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// outlier, tls, weight routing, canaries add if requried
type VirtualServiceAddonConfig struct {
	Attempts   int
	TryTimeout int
}

// Virtual Service Configuration for Istio K8s DNS Vhost traffic routing
type VirtualServiceConfig struct {
	Namespace       string
	VServiceName    string
	GatewayHostName []string
	IstioGateways   []string
	/* Expect the K8s Service FQDN  explorable from KUBE coredns for service discovery internally used by Istio */
	ServiceFqdn string
	ServicePort int
}

/*
The Deployment always runs gateways in root istio-system namespace in same ns as istio ingress gateway controller
All vs and routing traffic services must be in their respective namespace
*/
func getGatewaySelectorSvc(istioGateways []string) []string {
	istioGatewayVsPrefixSelector := []string{}
	for _, gateway := range istioGateways {
		istioGatewayVsPrefixSelector = append(istioGatewayVsPrefixSelector, fmt.Sprintf("istio-system/%s", gateway))
	}
	return istioGatewayVsPrefixSelector
}

func (istio *IstioClientSet) CreateVirtualService(ctx context.Context, serviceVhostConfig *VirtualServiceConfig) error {
	// the service gateway dont require waited routing or traffic rate limiting all done via destination rules for secrutiy design

	defaultServiceRetryConfig := &VirtualServiceAddonConfig{
		Attempts:   10,
		TryTimeout: 3,
	}

	vService := &networkingv1alpha3.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      serviceVhostConfig.VServiceName,
			Namespace: serviceVhostConfig.Namespace,
		},
		Spec: v1alpha3.VirtualService{
			Hosts:    serviceVhostConfig.GatewayHostName,
			Gateways: getGatewaySelectorSvc(serviceVhostConfig.IstioGateways),
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: []*v1alpha3.HTTPRouteDestination{
						{
							Destination: &v1alpha3.Destination{
								Host: serviceVhostConfig.ServiceFqdn,
								Port: &v1alpha3.PortSelector{
									Number: uint32(serviceVhostConfig.ServicePort),
								},
							},
						},
					},
					Retries: &v1alpha3.HTTPRetry{
						Attempts: int32(defaultServiceRetryConfig.Attempts),
						PerTryTimeout: &durationpb.Duration{
							Seconds: int64(defaultServiceRetryConfig.TryTimeout),
						},
					},
				},
			},
		},
	}

	_, err := istio.ClientSet.NetworkingV1alpha3().VirtualServices(serviceVhostConfig.Namespace).Create(istio.Ctx, vService, v1.CreateOptions{})
	if err != nil {
		log.Printf("Error creating Istio Virtual Service for Service Name %s %+v", serviceVhostConfig.VServiceName, err)
		return err
	}

	return nil
}

func (istio *IstioClientSet) DeleteService(ctx context.Context, config *VirtualServiceConfig) error {
	err := istio.ClientSet.NetworkingV1().VirtualServices(config.Namespace).Delete(istio.Ctx, config.VServiceName,
		v1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting Istio Virtual Service fort Service Name %s with error %+v", config.VServiceName, err)
		return err
	}

	return nil
}

// create upstrem service lb/rate limit and outlier detection for traffic routing and load balancing to upstream DPAI service sidecar
func (istio *IstioClientSet) CreateDestinationRules(ctx context.Context, serviceVhostConfig *VirtualServiceConfig) error {

	dr := &networkingv1alpha3.DestinationRule{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-dr", serviceVhostConfig.VServiceName),
			Namespace: serviceVhostConfig.Namespace,
		},
		Spec: v1alpha3.DestinationRule{
			Host: serviceVhostConfig.ServiceFqdn,
			TrafficPolicy: &v1alpha3.TrafficPolicy{
				LoadBalancer: &v1alpha3.LoadBalancerSettings{
					LbPolicy: &v1alpha3.LoadBalancerSettings_Simple{
						Simple: v1alpha3.LoadBalancerSettings_LEAST_REQUEST, // not using consistent ring hashing because of less sclaed downstream services
					},
				},
			},
		},
	}

	_, err := istio.ClientSet.NetworkingV1alpha3().DestinationRules(serviceVhostConfig.Namespace).Create(istio.Ctx, dr, v1.CreateOptions{})

	if err != nil {
		log.Printf("Error creating Istio Destination Rule for Service Name %s %+v", serviceVhostConfig.VServiceName, err)
		return err
	}

	return nil
}

func (istio *IstioClientSet) DeleteDestinationRules(ctx context.Context, config *VirtualServiceConfig) error {
	err := istio.ClientSet.NetworkingV1alpha3().DestinationRules(config.Namespace).Delete(istio.Ctx, config.VServiceName,
		v1.DeleteOptions{})

	if err != nil {
		log.Printf("Error deleting Istio Destination Rule fort Service Name %s with error %+v", config.VServiceName, err)
		return err
	}

	return nil
}
