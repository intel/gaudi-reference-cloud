// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package lb_provider

import (
	"fmt"
	"time"

	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	highwire "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/internal/LB_Provider/highwire"
)

const (
	HighWireProvider = "highwire"
)

type LoadBalancerProvider struct {
	Provider //contains an interface
}

type Config struct {
	BaseURL         string
	Domain          string
	UserName        string
	Secret          string
	HighwireTimeout time.Duration
}

type Provider interface {
	//init provider initializes the load balancer provider.
	//InitProvider(context.Context, string) error
	CreatePool(*ilbv1alpha1.Ilb) error
	CreateVirtualServer(*ilbv1alpha1.Ilb) error
	LinkVSToPool(*ilbv1alpha1.Ilb) error
	ProcessFinalizers(*ilbv1alpha1.Ilb) error
	ObserveCurrentAndReconcile(*ilbv1alpha1.Ilb) error
	GetStatus(*ilbv1alpha1.Ilb) error
}

func NewLoadBalancerProvider(provider string, c *Config) (*LoadBalancerProvider, error) {
	if provider == HighWireProvider {
		p, err := highwire.NewHighWireProvider(c.BaseURL, c.Domain, c.UserName, c.Secret, c.HighwireTimeout)
		if err != nil {
			return nil, err
		}
		return &LoadBalancerProvider{p}, nil
	}
	return nil, fmt.Errorf("Provider not found")
}
