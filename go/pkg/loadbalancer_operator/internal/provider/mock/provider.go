// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock_provider

import (
	context "context"

	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	v1alpha10 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	v1alpha11 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
)

// MockProvider is a mock of Provider interface.
type MockProvider struct {
}

// NewMockProvider creates a new mock instance.
func NewMockProvider() *MockProvider {
	return nil
}

// CreatePool mocks base method.
func (m *MockProvider) CreatePool(arg0 context.Context, arg1, arg2 string, arg3 v1alpha11.VPool, arg4 []v1alpha10.Instance, arg5 int) (string, error) {
	return "", nil
}

// CreateVirtualServer mocks base method.
func (m *MockProvider) CreateVirtualServer(arg0 context.Context, arg1, arg2 string, arg3 v1alpha11.VServer, arg4 int, arg6, arg7 string) (string, error) {
	return "", nil
}

// GetStatus mocks base method.
func (m *MockProvider) GetStatus(arg0 context.Context, arg1 *v1alpha11.Loadbalancer, arg2 []v1alpha1.FirewallRule) error {
	arg1.Status = v1alpha11.LoadbalancerStatus{
		State: v1alpha11.PENDING,
		Vip:   "1.2.3.4",
		Conditions: v1alpha11.ConditionsStatus{
			Listeners: []v1alpha11.ConditionsListenerStatus{{
				Port:          9090,
				PoolCreated:   false,
				VIPPoolLinked: false,
				VIPCreated:    false,
			}, {
				Port:          9091,
				PoolCreated:   false,
				VIPPoolLinked: false,
				VIPCreated:    false,
			}},
			FirewallRuleCreated: false,
		},
		Listeners: []v1alpha11.ListenerStatus{{
			Port:        9090,
			Name:        "listener1",
			State:       v1alpha11.PENDING,
			Message:     "provisioning",
			PoolMembers: nil,
			PoolID:      1,
			VipID:       2,
		}, {
			Port:        9091,
			Name:        "listener1",
			State:       v1alpha11.PENDING,
			Message:     "provisioning",
			PoolMembers: nil,
			PoolID:      1,
			VipID:       2,
		}},
	}
	return nil
}

// LinkVSToPool mocks base method.
func (m *MockProvider) LinkVSToPool(arg0 context.Context, arg1, arg2 int) (string, error) {
	return "", nil
}

// ObserveCurrentAndReconcile mocks base method.
func (m *MockProvider) ObserveCurrentAndReconcile(arg0 context.Context, arg1 string,
	arg2 v1alpha11.LoadbalancerListener, arg3 int, arg4 []v1alpha10.Instance) (string, []v1alpha11.PoolStatusMember, error) {
	return "", []v1alpha11.PoolStatusMember{}, nil
}

// ProcessFinalizers mocks base method.
func (m *MockProvider) ProcessFinalizers(arg0 context.Context, arg1 *v1alpha11.Loadbalancer) (string, error) {
	return "", nil
}

// ReconcileListeners mocks base method.
func (m *MockProvider) ReconcileListeners(ctx context.Context, name, vip string, listeners []v1alpha11.LoadbalancerListener) ([]int, error) {
	return []int{}, nil
}
