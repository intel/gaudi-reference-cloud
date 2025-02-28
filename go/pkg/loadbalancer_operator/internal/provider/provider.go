// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	mock_provider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider/mock"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/config"
)

type LoadbalancerProvider struct {
	Provider //contains an interface
}

type Config struct {
	BaseURL string
	Domain  string
	*config.Configuration
	Environment int
	UserGroup   int
}

type Provider interface {
	CreatePool(context.Context, string, string, loadbalancerv1alpha1.VPool, []cloudv1alpha1.Instance, int) (string, error)
	CreateVirtualServer(context.Context, string, string, loadbalancerv1alpha1.VServer, int, string, string) (string, error)
	LinkVSToPool(context.Context, int, int) (string, error)
	ProcessFinalizers(context.Context, *loadbalancerv1alpha1.Loadbalancer) (string, error)
	ObserveCurrentAndReconcile(context.Context, string, loadbalancerv1alpha1.LoadbalancerListener, int, []cloudv1alpha1.Instance) (string, []loadbalancerv1alpha1.PoolStatusMember, error)
	GetStatus(context.Context, *loadbalancerv1alpha1.Loadbalancer, []firewallv1alpha1.FirewallRule) error
	ReconcileListeners(ctx context.Context, name string, vip string, listeners []loadbalancerv1alpha1.LoadbalancerListener) ([]int, error)
}

func NewLoadbalancerProvider(provider string, c *Config) (*LoadbalancerProvider, error) {
		p, err := mock_provider.NewMockProvider(c.BaseURL, c.Domain, c.Configuration, c.Environment, c.UserGroup)
		if err != nil {
			return nil, err
		}
		return &LoadbalancerProvider{p}, nil
}
